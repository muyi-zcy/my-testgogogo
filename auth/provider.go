// Package auth 提供认证 Provider 注册、Token 缓存与统一认证入口。
//
// 内置 Provider：
//   - login：通过 YAML 配置的登录接口获取 Token
//   - static_token：直接使用配置中的静态 Token
//
// 自定义 Provider 可通过 RegisterProvider 在 init() 中注册。
package auth

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

// Credential 表示一次认证成功后获得的凭证，包含 Token 及其 HTTP 头信息。
type Credential struct {
	Token  string // 原始 Token 值（不含前缀）
	Header string // HTTP 头名称，默认 "Authorization"
	Prefix string // Token 前缀，如 "Bearer "
}

// HeaderValue 返回写入 HTTP 请求头的完整值（Prefix + Token）。
func (c Credential) HeaderValue() string {
	if c.Prefix != "" {
		return c.Prefix + c.Token
	}
	return c.Token
}

// Provider 认证提供者接口，每种认证方式实现此接口并注册到全局表。
type Provider interface {
	// Name 返回 Provider 唯一标识，与 configs/config.yaml 中 auth.provider 对应。
	Name() string
	// Authenticate 执行认证逻辑，成功后返回 Credential 并将 Token 写入 Client。
	Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (Credential, error)
	// Validate 校验已有 Credential 是否仍然有效。
	Validate(ctx context.Context, c *client.Client, cred Credential) (bool, error)
}

// CredentialRehydrator 可选接口：从缓存的 raw token 恢复完整 Credential（如补全 Bearer 前缀）。
// 实现此接口的 Provider 在 Token 缓存命中时能够正确还原请求头格式。
type CredentialRehydrator interface {
	Rehydrate(rawToken string) Credential
}

// providers 全局 Provider 注册表，key 为 Provider.Name()。
var providers = map[string]Provider{}

// RegisterProvider 将 Provider 注册到全局表，通常在 init() 中调用。
func RegisterProvider(p Provider) {
	providers[p.Name()] = p
}

// GetProvider 按名称查找已注册的 Provider。
func GetProvider(name string) (Provider, bool) {
	p, ok := providers[name]
	return p, ok
}

func init() {
	RegisterProvider(StaticTokenProvider{})
	RegisterProvider(LoginProvider{})
}

// ResolveProvider 根据配置中的 auth.provider 解析对应的 Provider 实例。
func ResolveProvider(cfg *config.Config) (Provider, error) {
	name := cfg.Auth.Provider
	if p, ok := providers[name]; ok {
		return p, nil
	}
	return nil, &UnknownProviderError{Name: name}
}

// UnknownProviderError 当配置的 auth.provider 未注册时返回此错误。
type UnknownProviderError struct {
	Name string
}

func (e *UnknownProviderError) Error() string {
	return "unknown auth provider: " + e.Name
}
