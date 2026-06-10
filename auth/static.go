// Package auth 提供认证 Provider 注册、Token 缓存与统一认证入口。
package auth

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

// StaticTokenProvider 静态 Token 认证 Provider，直接使用环境配置中的 token 字段。
// 适用于已有长期有效 Token 或无需登录接口的场景；不走 Token 缓存。
type StaticTokenProvider struct{}

// Name 返回 Provider 标识，对应 configs/config.yaml 中 auth.provider: static_token。
func (StaticTokenProvider) Name() string { return "static_token" }

// Authenticate 将配置中的 Token 写入 Client 的 Authorization 头并返回 Credential。
func (StaticTokenProvider) Authenticate(_ context.Context, c *client.Client, cfg *config.Config) (Credential, error) {
	cred := Credential{
		Token:  cfg.Token,
		Header: "Authorization",
	}
	c.SetToken(cred.HeaderValue())
	return cred, nil
}

// Validate 校验 Token 非空即可。
func (StaticTokenProvider) Validate(_ context.Context, _ *client.Client, cred Credential) (bool, error) {
	return cred.Token != "", nil
}
