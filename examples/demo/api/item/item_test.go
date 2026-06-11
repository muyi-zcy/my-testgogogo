// Package item 测试 demo 示例的商品列表接口。
package item

import (
	"testing"

	"github.com/muyi-zcy/my-testgogogo/assert"
	"github.com/muyi-zcy/my-testgogogo/examples/demo/apistep"
	"github.com/muyi-zcy/my-testgogogo/testkit"
	"github.com/stretchr/testify/require"
)

// TestItemList 验证已认证用户能分页查询商品列表。
func TestItemList(t *testing.T) {
	testkit.SkipIfDisabled(t)
	r := testkit.EnableAPIReport(t, "商品列表查询", "GET /api/items 分页查询")

	c := testkit.NewAuthenticatedClient(t)
	ctx, cancel := testkit.TestContext(t)
	defer cancel()

	r.Step("list items", func(t *testing.T) {
		params := apistep.ListParams{
			PageNum:  1,
			PageSize: 10,
		}
		r.SetInput(params)
		page, err := apistep.ListItems(ctx, c, params)
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
				"code": page.Records[0].Code,
				"name": page.Records[0].Name,
			}
		}
		r.SetResult(result)
	})
}
