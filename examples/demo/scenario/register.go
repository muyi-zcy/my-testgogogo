package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/load"
)

// Registry 注册 demo 示例压测场景。
var Registry = map[string]load.ScenarioMeta{
	"item-list": {
		Name:  "item-list",
		Type:  load.TypeAPI,
		Title: "商品列表",
		Fn: func(ctx context.Context, env *load.Env) error {
			_, err := ListItems(ctx, &env.Env, ListItemsInput{
				PageNum:  1,
				PageSize: env.Vars.MustInt("pageSize"),
			})
			return err
		},
	},
	"item-query-has-data": {
		Name:  "item-query-has-data",
		Type:  load.TypeFlow,
		Title: "商品查询流程（有数据）",
		Steps: []string{"system", "me", "list", "filter"},
		Fn: func(ctx context.Context, env *load.Env) error {
			return ItemQueryFlow(ctx, &env.Env, ItemQueryOptions{SimulateEmpty: false})
		},
	},
	"item-query-empty": {
		Name:  "item-query-empty",
		Type:  load.TypeFlow,
		Title: "商品查询流程（空列表）",
		Steps: []string{"system", "me", "list"},
		Fn: func(ctx context.Context, env *load.Env) error {
			return ItemQueryFlow(ctx, &env.Env, ItemQueryOptions{SimulateEmpty: true})
		},
	},
}
