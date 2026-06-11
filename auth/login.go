// Package auth 提供认证 Provider 注册、Token 缓存与统一认证入口。
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

// LoginProvider 内置登录认证 Provider，通过 YAML 配置的登录接口获取 Token。
// 适用于标准 REST 登录接口，响应体为 JSON 且 Token 位于可配置的 JSON 路径。
type LoginProvider struct{}

// Name 返回 Provider 标识，对应 configs/config.yaml 中 auth.provider: login。
func (LoginProvider) Name() string { return "login" }

// Authenticate 调用配置的登录接口，从响应 JSON 中提取 Token 并写入 HTTP 客户端。
func (LoginProvider) Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (Credential, error) {
	loginCfg := cfg.Auth.Login
	method := strings.ToUpper(loginCfg.Method)
	if method == "" {
		method = http.MethodPost
	}

	// 渲染请求体：支持模板变量 {{user.username}} 等
	body := renderBody(loginCfg.Body, cfg)
	opts := []client.RequestOption{client.WithBody(body)}
	// 登录请求通常不需要携带已有 Token
	if loginCfg.WithoutAuth {
		opts = append(opts, client.WithoutAuth())
	}

	resp, err := c.Do(ctx, method, loginCfg.URL, opts...)
	if err != nil {
		return Credential{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Credential{}, fmt.Errorf("login failed: status=%d body=%s", resp.StatusCode, string(resp.Body))
	}

	// 按点分路径（如 data.accessToken）从 JSON 响应中提取 Token
	token, err := ExtractJSONPath(resp.Body, loginCfg.TokenPath)
	if err != nil {
		return Credential{}, fmt.Errorf("extract token: %w", err)
	}
	if token == "" {
		return Credential{}, fmt.Errorf("login token is empty at path %q", loginCfg.TokenPath)
	}

	cred := Credential{Token: token, Header: "Authorization"}
	c.SetToken(cred.HeaderValue())
	return cred, nil
}

// Validate 校验凭证是否有效；LoginProvider 默认仅检查 Token 非空。
// 若配置了 validate.url，则由 Authenticator 调用 ValidateWithConfig 做远程校验。
func (LoginProvider) Validate(ctx context.Context, c *client.Client, cred Credential) (bool, error) {
	return cred.Token != "", nil
}

// ValidateWithConfig 使用配置中的 validate 端点远程校验 Token 是否仍有效。
// 若未配置 validate.url，则直接返回 true（跳过远程校验）。
func ValidateWithConfig(ctx context.Context, c *client.Client, cfg *config.Config) (bool, error) {
	validate := cfg.Auth.Login.Validate
	if validate.URL == "" {
		return true, nil
	}
	method := strings.ToUpper(validate.Method)
	if method == "" {
		method = http.MethodGet
	}
	resp, err := c.Do(ctx, method, validate.URL)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil
}

// renderBody 将登录请求体模板渲染为实际键值对。
// 若未配置 body 模板，则默认使用 username/password 字段。
func renderBody(templates map[string]string, cfg *config.Config) map[string]string {
	if len(templates) == 0 {
		return map[string]string{
			"username": cfg.User.Username,
			"password": cfg.User.Password,
		}
	}
	out := make(map[string]string, len(templates))
	for key, value := range templates {
		out[key] = cfg.Expand(value)
	}
	return out
}

// ExtractJSONPath 从 JSON 响应体中按点分路径提取字符串值。
// 支持 string、json.Number、float64、bool 类型；路径示例："token" 或 "data.accessToken"。
func ExtractJSONPath(body []byte, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty json path")
	}

	var data any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // 避免大整数被解析为 float64 丢失精度
	if err := dec.Decode(&data); err != nil {
		return "", fmt.Errorf("decode json: %w", err)
	}

	// 逐级遍历 JSON 对象
	current := data
	for _, part := range strings.Split(path, ".") {
		obj, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("path %q: not an object at %q", path, part)
		}
		val, ok := obj[part]
		if !ok {
			return "", fmt.Errorf("path %q: key %q not found", path, part)
		}
		current = val
	}
	return stringifyJSONValue(current)
}

// stringifyJSONValue 将 JSON 叶子节点值转换为字符串形式的 Token。
func stringifyJSONValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case float64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("unsupported token value type %T", value)
	}
}
