package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const routeRegexPrefix = "regex:"

// RouteRule 单条路由规则，支持前缀或正则匹配。
type RouteRule struct {
	Type    string // prefix | regex
	Pattern string
	Target  string // 解析后的目标 base URL
}

// Router 按请求路径解析目标 base URL，类似 nginx location。
type Router struct {
	defaultBase string
	prefixRules []RouteRule
	regexRules  []compiledRegexRule
}

// RoutesYAML 支持 map 与 list 两种 routes 配置写法。
type RoutesYAML struct {
	Rules []RouteRule
}

// UnmarshalYAML 解析 routes 配置。map 与 list 二选一，不可混用。
//
// Map 形式（前缀 key 直接写路径；正则 key 以 regex: 开头）：
//
//	routes:
//	  /wms: http://host:20002
//	  "regex:^/wcs/": http://host:20003
//
// List 形式（prefix / regex 字段二选一）：
//
//	routes:
//	  - prefix: /wms
//	    target: http://host:20002
//	  - regex: '^/wcs/'
//	    target: http://host:20003
func (r *RoutesYAML) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}

	switch value.Kind {
	case yaml.MappingNode:
		rules, err := unmarshalRoutesMap(value)
		if err != nil {
			return err
		}
		r.Rules = rules
	case yaml.SequenceNode:
		rules, err := unmarshalRoutesList(value)
		if err != nil {
			return err
		}
		r.Rules = rules
	default:
		return fmt.Errorf("routes must be a map or list")
	}
	return nil
}

type routeListEntry struct {
	Prefix string `yaml:"prefix"`
	Regex  string `yaml:"regex"`
	Target string `yaml:"target"`
}

func unmarshalRoutesMap(node *yaml.Node) ([]RouteRule, error) {
	if len(node.Content)%2 != 0 {
		return nil, fmt.Errorf("routes map has invalid number of entries")
	}

	rules := make([]RouteRule, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		pattern := strings.TrimSpace(keyNode.Value)
		target := strings.TrimSpace(valNode.Value)
		if pattern == "" {
			return nil, fmt.Errorf("routes map key cannot be empty")
		}
		if target == "" {
			return nil, fmt.Errorf("routes[%q] target cannot be empty", pattern)
		}

		ruleType := "prefix"
		if strings.HasPrefix(pattern, routeRegexPrefix) {
			ruleType = "regex"
			pattern = strings.TrimPrefix(pattern, routeRegexPrefix)
		}

		rules = append(rules, RouteRule{
			Type:    ruleType,
			Pattern: pattern,
			Target:  target,
		})
	}
	return rules, nil
}

func unmarshalRoutesList(node *yaml.Node) ([]RouteRule, error) {
	rules := make([]RouteRule, 0, len(node.Content))
	for i, item := range node.Content {
		var entry routeListEntry
		if err := item.Decode(&entry); err != nil {
			return nil, fmt.Errorf("routes[%d]: %w", i, err)
		}

		switch {
		case entry.Prefix != "" && entry.Regex != "":
			return nil, fmt.Errorf("routes[%d]: prefix and regex are mutually exclusive", i)
		case entry.Prefix != "":
			rules = append(rules, RouteRule{
				Type:    "prefix",
				Pattern: entry.Prefix,
				Target:  entry.Target,
			})
		case entry.Regex != "":
			rules = append(rules, RouteRule{
				Type:    "regex",
				Pattern: entry.Regex,
				Target:  entry.Target,
			})
		default:
			return nil, fmt.Errorf("routes[%d]: prefix or regex is required", i)
		}
	}
	return rules, nil
}

func buildRouter(defaultBase string, services map[string]string, routes RoutesYAML) (*Router, error) {
	defaultBase = strings.TrimRight(defaultBase, "/")
	if len(routes.Rules) == 0 {
		return nil, nil
	}

	prefixRules := make([]RouteRule, 0)
	regexRules := make([]compiledRegexRule, 0)

	for i, rule := range routes.Rules {
		target, err := resolveRouteTarget(rule.Target, services)
		if err != nil {
			return nil, fmt.Errorf("routes[%d]: %w", i, err)
		}

		switch rule.Type {
		case "prefix":
			pattern, err := normalizePrefixPattern(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("routes[%d]: %w", i, err)
			}
			prefixRules = append(prefixRules, RouteRule{
				Type:    "prefix",
				Pattern: pattern,
				Target:  target,
			})
		case "regex":
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("routes[%d]: invalid regex %q: %w", i, rule.Pattern, err)
			}
			regexRules = append(regexRules, compiledRegexRule{
				pattern: rule.Pattern,
				re:      re,
				target:  target,
			})
		default:
			return nil, fmt.Errorf("routes[%d]: unsupported match type %q", i, rule.Type)
		}
	}

	sortPrefixRules(prefixRules)

	return &Router{
		defaultBase: defaultBase,
		prefixRules: prefixRules,
		regexRules:  regexRules,
	}, nil
}

func resolveRouteTarget(target string, services map[string]string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("target is required")
	}

	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return normalizeBaseURL(target)
	}

	if services == nil {
		return "", fmt.Errorf("target %q is not a URL and services is not configured", target)
	}

	serviceURL, ok := services[target]
	if !ok {
		return "", fmt.Errorf("unknown service %q", target)
	}
	return normalizeBaseURL(serviceURL)
}

func normalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid base url %q: %w", raw, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid base url %q: scheme must be http or https", raw)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid base url %q: host is required", raw)
	}
	return strings.TrimRight(raw, "/"), nil
}

func normalizePrefixPattern(pattern string) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", fmt.Errorf("prefix cannot be empty")
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		pattern = "/"
	}
	return pattern, nil
}

func sortPrefixRules(rules []RouteRule) {
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if len(rules[j].Pattern) > len(rules[i].Pattern) {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}
