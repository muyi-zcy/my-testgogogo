// Package auth 提供认证 Provider 注册、Token 缓存与统一认证入口。
package auth

import (
	"context"
	"fmt"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

// Authenticate 是测试框架的统一认证入口，流程如下：
//  1. 根据配置解析 Provider
//  2. static_token 模式直接返回配置中的 Token
//  3. 其他模式先尝试从本地缓存加载 Token，校验通过后复用
//  4. 缓存未命中或校验失败时，调用 Provider.Authenticate 重新登录并写入缓存
func Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (Credential, error) {
	provider, err := ResolveProvider(cfg)
	if err != nil {
		return Credential{}, err
	}

	store := NewTokenStore(cfg.AuthCache)

	// 静态 Token 不走缓存，直接使用配置值
	if cfg.Auth.Provider == "static_token" {
		return provider.Authenticate(ctx, c, cfg)
	}

	// 尝试从本地文件缓存复用 Token
	if token, ok := store.Load(cfg); ok {
		cred := rehydrateCredential(provider, token)
		c.SetToken(cred.HeaderValue())
		if valid, err := validateCredential(ctx, c, cfg, provider, cred); err == nil && valid {
			return cred, nil
		}
		// 缓存 Token 已失效，清除后重新登录
		_ = store.Clear(cfg)
		c.SetToken("")
	}

	cred, err := provider.Authenticate(ctx, c, cfg)
	if err != nil {
		return Credential{}, err
	}

	if err := store.Save(cfg, cred.Token); err != nil {
		return Credential{}, fmt.Errorf("save token cache: %w", err)
	}
	return cred, nil
}

// ClearCache 清除当前配置对应的 Token 本地缓存文件。
func ClearCache(cfg *config.Config) error {
	return NewTokenStore(cfg.AuthCache).Clear(cfg)
}

// validateCredential 校验凭证有效性：优先使用配置的 validate 端点，否则调用 Provider.Validate。
func validateCredential(ctx context.Context, c *client.Client, cfg *config.Config, provider Provider, cred Credential) (bool, error) {
	if cfg.Auth.Login.Validate.URL != "" {
		return ValidateWithConfig(ctx, c, cfg)
	}
	return provider.Validate(ctx, c, cred)
}

// rehydrateCredential 从缓存的原始 Token 字符串恢复完整 Credential。
// 若 Provider 实现了 CredentialRehydrator（如 demoauth 需补全 Bearer 前缀），则使用其逻辑。
func rehydrateCredential(provider Provider, rawToken string) Credential {
	if r, ok := provider.(CredentialRehydrator); ok {
		return r.Rehydrate(rawToken)
	}
	return Credential{Token: rawToken, Header: "Authorization"}
}
