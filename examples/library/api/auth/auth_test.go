// Package auth 测试 library 示例的认证相关接口（内置 login Provider）。
package auth

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/auth"
	"github.com/muyi-zcy/my-testgogogo/examples/library/apistep"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogin 验证内置 login Provider 能成功登录并获取 Token。
func TestLogin(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "管理员登录", "POST /api/auth/login 内置 login Provider")

	cfg := testkit.LoadConfig(t)
	testkit.ClearAuthCache(t, cfg)

	c := testkit.NewClient(t, cfg)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("authenticate", func(t *testing.T) {
		cred, err := auth.Authenticate(ctx, c, cfg)
		require.NoError(t, err)
		assert.NotEmpty(t, cred.Token)
		assert.NotEmpty(t, c.Token())
		r.SetResult(map[string]any{
			"tokenLength": len(cred.Token),
			"hasToken":    c.Token() != "",
		})
	})
}

// TestGetMe 验证管理员信息接口返回预期用户名。
func TestGetMe(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "获取管理员信息", "GET /api/auth/me")

	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("get me", func(t *testing.T) {
		info, err := apistep.GetMe(ctx, c)
		require.NoError(t, err)
		assert.Equal(t, "librarian", info.User.Username)
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
		badCfg.User.Password = "wrong-password"
		_, err := auth.Authenticate(ctx, c, &badCfg)
		require.Error(t, err)
		r.SetResult(map[string]any{
			"expected": "login failed",
			"actual":   err.Error(),
		})
	})
}

// TestLogout 验证登出后 Token 失效。
func TestLogout(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "管理员登出", "GET /api/auth/logout 后 token 失效")

	cfg := testkit.LoadConfig(t)
	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("logout", func(t *testing.T) {
		err := apistep.Logout(ctx, c)
		require.NoError(t, err)
		assert.Empty(t, c.Token())
		testkit.ClearAuthCache(t, cfg)
		r.Record("tokenCleared", true)
	})

	r.Step("get me after logout", func(t *testing.T) {
		_, err := apistep.GetMe(ctx, c)
		require.Error(t, err)
		r.SetResult(map[string]any{
			"expected": "unauthorized",
			"actual":   err.Error(),
		})
	})
}
