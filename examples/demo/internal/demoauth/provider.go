// Package demoauth 是 demo 示例的自定义认证 Provider，演示如何扩展框架认证能力。
//
// 特点：
//   - 调用 /api/auth/v2/login 获取嵌套 JSON 中的 accessToken
//   - 自动拼接 Bearer 前缀
//   - 实现 CredentialRehydrator 以正确恢复缓存 Token
package demoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/muyi-zcy/my-testgogogo/auth"
	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

const providerName = "demoauth"

func init() {
	auth.RegisterProvider(Provider{})
}

// Provider 自定义认证：调用 /api/auth/v2/login，解析嵌套 JSON 中的 accessToken。
type Provider struct{}

func (Provider) Name() string { return providerName }

// Authenticate 优先使用配置中的静态 Token；否则调用 v2 登录接口获取 accessToken。
func (Provider) Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (auth.Credential, error) {
	if cfg.Token != "" {
		cred := auth.Credential{
			Token:  cfg.Token,
			Header: "Authorization",
			Prefix: "Bearer ",
		}
		c.SetToken(cred.HeaderValue())
		return cred, nil
	}

	var resp loginResponse
	err := c.DoJSON(ctx, http.MethodPost, "/api/auth/v2/login", &resp,
		client.WithoutAuth(),
		client.WithBody(map[string]string{
			"username": cfg.User.Username,
			"password": cfg.User.Password,
		}),
	)
	if err != nil {
		return auth.Credential{}, fmt.Errorf("demoauth login: %w", err)
	}
	if resp.Code != 0 {
		return auth.Credential{}, fmt.Errorf("demoauth login: code=%d message=%s", resp.Code, resp.Message)
	}
	if resp.Data.AccessToken == "" {
		return auth.Credential{}, fmt.Errorf("demoauth login: accessToken is empty")
	}

	prefix := resp.Data.TokenType
	if prefix != "" {
		prefix += " "
	}

	cred := auth.Credential{
		Token:  resp.Data.AccessToken,
		Header: "Authorization",
		Prefix: prefix,
	}
	c.SetToken(cred.HeaderValue())
	return cred, nil
}

// Validate 调用 /api/auth/me 远程校验 Token 有效性。
func (Provider) Validate(ctx context.Context, c *client.Client, cred auth.Credential) (bool, error) {
	resp, err := c.Get(ctx, "/api/auth/me")
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("validate token: status=%d", resp.StatusCode)
	}
	return cred.Token != "", nil
}

// Rehydrate 从缓存 raw token 恢复带 Bearer 前缀的 Credential。
func (Provider) Rehydrate(rawToken string) auth.Credential {
	return auth.Credential{
		Token:  rawToken,
		Header: "Authorization",
		Prefix: "Bearer ",
	}
}

// loginResponse v2 登录接口的响应结构。
type loginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
	} `json:"data"`
}

// DecodeLoginResponse 供测试或调试使用的响应解析辅助函数。
func DecodeLoginResponse(body []byte) (accessToken, tokenType string, err error) {
	var resp loginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", "", err
	}
	if resp.Code != 0 {
		return "", "", fmt.Errorf("code=%d message=%s", resp.Code, resp.Message)
	}
	return resp.Data.AccessToken, resp.Data.TokenType, nil
}
