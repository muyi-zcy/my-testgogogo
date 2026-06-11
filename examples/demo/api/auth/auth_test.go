// Package auth 测试 demo 示例的认证相关接口。
package auth

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/auth"
	"github.com/muyi-zcy/my-testgogogo/examples/demo/apistep"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogin 验证 demoauth 自定义 Provider 能成功登录并获取 Bearer Token。
func TestLogin(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "用户登录", "POST /api/auth/v2/login 自定义 demoauth Provider")

	cfg := testkit.LoadConfig(t)
	testkit.ClearAuthCache(t, cfg) // 清除缓存，确保走完整登录流程

	c := testkit.NewClient(t, cfg)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("authenticate", func(t *testing.T) {
		r.SetInput(map[string]any{
			"username": cfg.User.Username,
			"provider": cfg.Auth.Provider,
		})
		cred, err := auth.Authenticate(ctx, c, cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, cred.Token)
		assert.NotEmpty(t, c.Token())
		assert.Contains(t, c.Token(), "Bearer ")
		r.SetResult(map[string]any{
			"tokenLength": len(cred.Token),
			"hasToken":    c.Token() != "",
		})
	})
}

// TestGetMe 验证已认证客户端能获取当前用户信息。
func TestGetMe(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "获取用户信息", "GET /api/auth/me")

	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("get me", func(t *testing.T) {
		r.SetInput(map[string]any{
			"method": "GET",
			"path":   "/api/auth/me",
		})
		info, err := apistep.GetMe(ctx, c)
		require.NoError(t, err)
		assert.NotEmpty(t, info.User.Username)
		assert.NotEmpty(t, info.Roles)
		r.SetResult(map[string]any{
			"username":    info.User.Username,
			"nickName":    info.User.NickName,
			"roleCount":   len(info.Roles),
			"permissions": len(info.Permissions),
		})
	})
}

// TestLoginWithWrongPassword 验证错误密码时登录失败。
func TestLoginWithWrongPassword(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "错误密码登录", "验证错误密码时登录失败")

	cfg := testkit.LoadConfig(t)
	c := testkit.NewClient(t, cfg)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("login with bad password", func(t *testing.T) {
		badCfg := *cfg
		badCfg.User.Password = "invalid-password"
		r.SetInput(map[string]any{
			"username": badCfg.User.Username,
			"password": badCfg.User.Password,
			"provider": badCfg.Auth.Provider,
		})
		_, err := auth.Authenticate(ctx, c, &badCfg)
		require.Error(t, err)
		r.SetResult(map[string]any{
			"expected": "login failed",
			"actual":   err.Error(),
		})
	})
}

// TestLogout 验证登出后 Token 被清除且后续请求失败。
func TestLogout(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "用户登出", "GET /api/auth/logout 后 token 失效")

	_ = testkit.LoadConfig(t)
	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("logout", func(t *testing.T) {
		r.SetInput(map[string]any{
			"method": "GET",
			"path":   "/api/auth/logout",
		})
		err := apistep.Logout(ctx, c)
		require.NoError(t, err)
		assert.Empty(t, c.Token())
		r.Record("tokenCleared", true)
	})

	r.Step("get me after logout", func(t *testing.T) {
		r.SetInput(map[string]any{
			"method": "GET",
			"path":   "/api/auth/me",
		})
		_, err := apistep.GetMe(ctx, c)
		require.Error(t, err)
		r.SetResult(map[string]any{
			"expected": "unauthorized",
			"actual":   err.Error(),
		})
	})
}
