// Package book 测试 library 示例的图书列表接口。
package book

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/assert"
	"github.com/muyi-zcy/my-testgogogo/examples/library/apistep"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/require"
)

// TestBookList 验证已认证管理员能分页查询图书列表。
func TestBookList(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "图书列表查询", "GET /api/books 分页查询")

	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("list books", func(t *testing.T) {
		page, err := apistep.ListBooks(ctx, c, apistep.ListParams{
			PageNum:  1,
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.PageNotEmpty(t, page.Total, page.Records)

		result := map[string]any{
			"total":   page.Total,
			"current": page.Current,
			"size":    page.Size,
			"count":   len(page.Records),
		}
		if len(page.Records) > 0 {
			result["first"] = map[string]any{
				"isbn":   page.Records[0].ISBN,
				"title":  page.Records[0].Title,
				"author": page.Records[0].Author,
			}
		}
		r.SetResult(result)
	})
}
