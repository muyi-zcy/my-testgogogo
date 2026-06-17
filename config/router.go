package config

import (
	"regexp"
	"strings"
)

type compiledRegexRule struct {
	pattern string
	re      *regexp.Regexp
	target  string
}

// Resolve 根据请求路径返回目标 base URL；无匹配时返回 default base_url。
func (r *Router) Resolve(path string) string {
	if r == nil {
		return ""
	}

	normalized := normalizeRequestPath(path)

	if match := r.matchPrefix(normalized); match != "" {
		return match
	}
	if match := r.matchRegex(normalized); match != "" {
		return match
	}
	return r.defaultBase
}

// Fingerprint 生成路由表指纹，用于 Token 缓存键区分不同路由配置。
func (r *Router) Fingerprint() string {
	if r == nil || (len(r.prefixRules) == 0 && len(r.regexRules) == 0) {
		return ""
	}

	var b strings.Builder
	for _, rule := range r.prefixRules {
		b.WriteString("p:")
		b.WriteString(rule.Pattern)
		b.WriteByte('=')
		b.WriteString(rule.Target)
		b.WriteByte(';')
	}
	for _, rule := range r.regexRules {
		b.WriteString("r:")
		b.WriteString(rule.pattern)
		b.WriteByte('=')
		b.WriteString(rule.target)
		b.WriteByte(';')
	}
	return b.String()
}

func (r *Router) matchPrefix(path string) string {
	var best string
	bestLen := -1
	for _, rule := range r.prefixRules {
		if !prefixMatches(path, rule.Pattern) {
			continue
		}
		if len(rule.Pattern) > bestLen {
			bestLen = len(rule.Pattern)
			best = rule.Target
		}
	}
	return best
}

func prefixMatches(path, prefix string) bool {
	if prefix == "/" {
		return true
	}
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func (r *Router) matchRegex(path string) string {
	for _, rule := range r.regexRules {
		if rule.re.MatchString(path) {
			return rule.target
		}
	}
	return ""
}

func normalizeRequestPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
