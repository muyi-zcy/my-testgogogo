// Package system 测试 library 示例的系统信息接口。
package system

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/examples/library/apistep"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/require"
)

// TestSystemInfo 验证系统信息接口返回预期的服务名称。
func TestSystemInfo(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "系统信息", "GET /api/system/info 无需登录")

	cfg := testkit.LoadConfig(t)
	c := testkit.NewClient(t, cfg)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("get system info", func(t *testing.T) {
		info, err := apistep.GetSystemInfo(ctx, c)
		require.NoError(t, err)
		require.Equal(t, "my-testgogogo-library", info["name"])
		r.SetResult(info)
	})
}
