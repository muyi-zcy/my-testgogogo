// Package example 演示 demo 示例的 Flow 流程测试：多步骤串联、Vars 变量传递与条件分支。
package example

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/examples/demo/scenario"
	"github.com/muyi-zcy/my-testgogogo/flow"
	"github.com/muyi-zcy/my-testgogogo/report"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlowItemQueryChain 商品查询流程：系统信息 → 用户信息 → 商品列表 → 条件分支。
func TestFlowItemQueryChain(t *testing.T) {
	testkit.SkipIfDisabled(t)

	cases := []struct {
		name          string
		simulateEmpty bool
		reportTitle   string
		reportDesc    string
	}{
		{
			name:          "has_data",
			simulateEmpty: false,
			reportTitle:   "商品查询流程验证（有数据分支）",
			reportDesc:    "系统信息 → 用户信息 → 商品列表 → 按 vars 编码过滤",
		},
		{
			name:          "empty_list",
			simulateEmpty: true,
			reportTitle:   "商品查询流程验证（空列表分支）",
			reportDesc:    "系统信息 → 用户信息 → 空列表 → 验证空结果分页",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runItemQueryFlow(t, tc.simulateEmpty, tc.reportTitle, tc.reportDesc)
		})
	}
}

func runItemQueryFlow(t *testing.T, simulateEmptyList bool, title, description string) {
	t.Helper()

	r := testkit.EnableReport(t, testkit.ReportMeta{
		Generate:    true,
		Title:       title,
		Category:    report.CategoryFlow,
		Description: description,
	})

	env := testkit.NewScenarioEnv(t)
	env.Vars = flow.NewVars(flow.DefaultSeed())

	r.Step("run item query flow", func(t *testing.T) {
		err := scenario.ItemQueryFlow(env.CTX, env, scenario.ItemQueryOptions{
			SimulateEmpty: simulateEmptyList,
		})
		require.NoError(t, err)

		switch env.Vars.Get("branch") {
		case "has_data":
			assert.NotZero(t, env.Vars.MustInt("itemCount"))
			assert.NotEmpty(t, env.Vars.MustString("firstItemCode"))
			r.SetResult(map[string]any{
				"branch":        "has_data",
				"itemCount":     env.Vars.MustInt("itemCount"),
				"firstItemCode": env.Vars.MustString("firstItemCode"),
			})
		case "empty":
			assert.Zero(t, env.Vars.MustInt("itemCount"))
			r.SetResult(map[string]any{
				"branch":    "empty",
				"itemCount": 0,
			})
		default:
			t.Fatal("unknown branch:", env.Vars.Get("branch"))
		}
	})

	if env.Vars.MustInt("roleCount") > 0 {
		r.Note("用户具备角色权限，流程完整执行")
	} else {
		r.Note("用户无角色，流程以有限断言结束")
	}
}
