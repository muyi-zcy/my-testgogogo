// Package example 演示 library 示例的 Flow 流程测试：多步骤串联、Vars 变量传递与条件分支。
package example

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/examples/library/apistep"
	"github.com/muyi-zcy/my-testgogogo/flow"
	"github.com/muyi-zcy/my-testgogogo/report"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const nonexistentISBN = "__NONEXISTENT__"

// TestFlowBookQueryChain 图书查询流程：系统信息 → 管理员信息 → 图书列表 → 条件分支。
// 通过 t.Run 分别演示「有数据」与「空列表」两条互斥分支。
func TestFlowBookQueryChain(t *testing.T) {
	testkit.SkipIfDisabled(t)

	cases := []struct {
		name            string
		simulateEmpty   bool
		reportTitle     string
		reportDesc      string
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

	cfg := testkit.LoadConfig(t)
	vars := flow.NewVars(flow.DefaultSeed())

	c := testkit.NewClient(t, cfg)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("get system info", func(t *testing.T) {
		info, err := apistep.GetSystemInfo(ctx, c)
		require.NoError(t, err)
		require.NotEmpty(t, info["name"])
		vars.Set("systemName", info["name"])
		r.SetResult(map[string]any{"system": info})
	})

	authClient := testkit.NewAuthenticatedClient(t)

	r.Step("get librarian info", func(t *testing.T) {
		userInfo, err := apistep.GetMe(ctx, authClient)
		require.NoError(t, err)
		require.NotEmpty(t, userInfo.User.Username)

		vars.Set("username", userInfo.User.Username)
		vars.Set("roleCount", len(userInfo.Roles))
		r.SetResult(map[string]any{
			"username":    userInfo.User.Username,
			"nickName":    userInfo.User.NickName,
			"roleCount":   len(userInfo.Roles),
			"permissions": len(userInfo.Permissions),
		})
	})

	r.Step("list books", func(t *testing.T) {
		params := apistep.ListParams{
			PageNum:  1,
			PageSize: vars.MustInt("pageSize"),
		}
		if simulateEmptyList {
			params.ISBN = nonexistentISBN
		}

		page, err := apistep.ListBooks(ctx, authClient, params)
		require.NoError(t, err)

		vars.Set("bookCount", int(page.Total))
		result := map[string]any{
			"total":   page.Total,
			"current": page.Current,
			"size":    page.Size,
		}
		if len(page.Records) == 0 {
			vars.Set("branch", "empty")
			r.Note("图书列表为空，走空列表分支")
			r.SetResult(result)
			return
		}

		first := page.Records[0]
		vars.Set("branch", "has_data")
		vars.Set("firstISBN", first.ISBN)
		vars.Set("firstTitle", first.Title)
		result["firstBook"] = map[string]any{
			"isbn":   first.ISBN,
			"title":  first.Title,
			"author": first.Author,
		}
		r.SetResult(result)
	})

	switch vars.Get("branch") {
	case "has_data":
		r.Note("走有数据分支：按 ISBN 精确过滤")
		r.Step("query book by isbn from vars", func(t *testing.T) {
			isbn := vars.MustString("firstISBN")

			page, err := apistep.ListBooks(ctx, authClient, apistep.ListParams{
				PageNum:  1,
				PageSize: vars.MustInt("pageSize"),
				ISBN:     isbn,
			})
			require.NoError(t, err)
			require.NotEmpty(t, page.Records)
			assert.Equal(t, isbn, page.Records[0].ISBN)
			vars.Set("filteredCount", int(page.Total))
			r.SetResult(map[string]any{
				"queryISBN": isbn,
				"matched":   page.Records[0],
				"total":     page.Total,
			})
		})

	case "empty":
		r.Note("走空列表分支：验证空结果分页结构")
		r.Step("verify empty list pagination", func(t *testing.T) {
			page, err := apistep.ListBooks(ctx, authClient, apistep.ListParams{
				PageNum:  1,
				PageSize: vars.MustInt("pageSize"),
				ISBN:     nonexistentISBN,
			})
			require.NoError(t, err)
			assert.Empty(t, page.Records)
			assert.Zero(t, page.Total)
			r.SetResult(map[string]any{
				"total":   page.Total,
				"records": len(page.Records),
			})
		})

	default:
		t.Fatal("unknown branch:", vars.Get("branch"))
	}

	if vars.MustInt("roleCount") > 0 {
		r.Note("管理员具备角色权限，流程完整执行")
	}
}
