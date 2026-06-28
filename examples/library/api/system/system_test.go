// Package system 测试 library 示例的系统信息接口。
package system

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/examples/library/scenario"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/require"
)

// TestSystemInfo 验证无需登录即可获取系统信息。
func TestSystemInfo(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "系统信息", "GET /api/system/info 无需登录")

	env := testkit.NewScenarioEnv(t)

	r.Step("get system info", func(t *testing.T) {
		info, err := scenario.GetSystemInfo(env.CTX, env)
		require.NoError(t, err)
		require.NotEmpty(t, info["name"])
		r.SetResult(info)
	})
}
