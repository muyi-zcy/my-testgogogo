// Package example 演示 library 示例的 Flow 流程测试：多步骤串联、Vars 变量传递与条件分支。
package example

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/examples/library/scenario"
	"github.com/muyi-zcy/my-testgogogo/flow"
	"github.com/muyi-zcy/my-testgogogo/report"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFlowBookQueryChain 图书查询流程：系统信息 → 管理员信息 → 图书列表 → 条件分支。
func TestFlowBookQueryChain(t *testing.T) {
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
			reportTitle:   "图书查询流程验证（有数据分支）",
			reportDesc:    "系统信息 → 管理员信息 → 图书列表 → 按 ISBN 过滤",
		},
		{
			name:          "empty_list",
			simulateEmpty: true,
			reportTitle:   "图书查询流程验证（空列表分支）",
			reportDesc:    "系统信息 → 管理员信息 → 空列表 → 验证空结果分页",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runBookQueryFlow(t, tc.simulateEmpty, tc.reportTitle, tc.reportDesc)
		})
	}
}

func runBookQueryFlow(t *testing.T, simulateEmptyList bool, title, description string) {
	t.Helper()

	r := testkit.EnableReport(t, testkit.ReportMeta{
		Generate:    true,
		Title:       title,
		Category:    report.CategoryFlow,
		Description: description,
	})

	env := testkit.NewScenarioEnv(t)
	env.Vars = flow.NewVars(flow.DefaultSeed())

	r.Step("run book query flow", func(t *testing.T) {
		err := scenario.BookQueryFlow(env.CTX, env, scenario.BookQueryOptions{
			SimulateEmpty: simulateEmptyList,
		})
		require.NoError(t, err)

		switch env.Vars.Get("branch") {
		case "has_data":
			assert.NotZero(t, env.Vars.MustInt("bookCount"))
			assert.NotEmpty(t, env.Vars.MustString("firstISBN"))
			r.SetResult(map[string]any{
				"branch":    "has_data",
				"bookCount": env.Vars.MustInt("bookCount"),
				"firstISBN": env.Vars.MustString("firstISBN"),
			})
		case "empty":
			assert.Zero(t, env.Vars.MustInt("bookCount"))
			r.SetResult(map[string]any{
				"branch":    "empty",
				"bookCount": 0,
			})
		default:
			t.Fatal("unknown branch:", env.Vars.Get("branch"))
		}
	})

	if env.Vars.MustInt("roleCount") > 0 {
		r.Note("管理员具备角色权限，流程完整执行")
	}
}
