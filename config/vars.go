package config

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Var 返回名为 key 的全局变量值。
func (c *Config) Var(key string) (any, bool) {
	if c == nil || len(c.Vars) == 0 {
		return nil, false
	}
	val, ok := c.Vars[key]
	return val, ok
}

// HasVar 判断全局变量 key 是否存在。
func (c *Config) HasVar(key string) bool {
	_, ok := c.Var(key)
	return ok
}

// VarString 返回名为 key 的全局变量字符串值。
func (c *Config) VarString(key string) (string, error) {
	val, ok := c.Var(key)
	if !ok {
		return "", fmt.Errorf("config var %q not found", key)
	}
	return stringifyVar(val), nil
}

// VarInt 返回名为 key 的全局变量整数值。
func (c *Config) VarInt(key string) (int, error) {
	val, ok := c.Var(key)
	if !ok {
		return 0, fmt.Errorf("config var %q not found", key)
	}
	return parseVarInt(val, key)
}

// VarsInto 将合并后的 vars 解码到 dst 结构体。
// dst 需为指针；字段通过 yaml tag 与 configs 中 vars 键名对应。
func (c *Config) VarsInto(dst any) error {
	if dst == nil {
		return fmt.Errorf("config vars decode target is nil")
	}

	vars := c.Vars
	if vars == nil {
		vars = map[string]any{}
	}

	data, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("marshal config vars: %w", err)
	}
	if err := yaml.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("decode config vars: %w", err)
	}
	return nil
}

// Expand 将字符串中的 {{user.*}} 与 {{vars.*}} 模板替换为配置值。
func (c *Config) Expand(s string) string {
	if c == nil {
		return s
	}

	replacer := strings.NewReplacer(
		"{{user.username}}", c.User.Username,
		"{{user.password}}", c.User.Password,
		"{{user.type}}", c.User.Type,
	)
	s = replacer.Replace(s)

	for key, val := range c.Vars {
		s = strings.ReplaceAll(s, "{{vars."+key+"}}", stringifyVar(val))
	}
	return s
}

func mergeVars(base, override map[string]any) map[string]any {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}

	merged := make(map[string]any, len(base)+len(override))
	for key, val := range base {
		merged[key] = val
	}
	for key, val := range override {
		merged[key] = val
	}
	return merged
}

func stringifyVar(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprint(v)
	}
}

func parseVarInt(val any, key string) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if v != float64(int64(v)) {
			return 0, fmt.Errorf("config var %q is not an integer: %v", key, val)
		}
		return int(v), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("config var %q is not an integer: %w", key, err)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("config var %q has unsupported integer type %T", key, val)
	}
}
