package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/load"
)

// Registry 注册 library 示例压测场景。
var Registry = map[string]load.ScenarioMeta{
	"book-list": {
		Name:  "book-list",
		Type:  load.TypeAPI,
		Title: "图书列表",
		Fn: func(ctx context.Context, env *load.Env) error {
			_, err := ListBooks(ctx, &env.Env, ListBooksInput{
				PageNum:  1,
				PageSize: env.Vars.MustInt("pageSize"),
			})
			return err
		},
	},
	"book-query-has-data": {
		Name:  "book-query-has-data",
		Type:  load.TypeFlow,
		Title: "图书查询流程（有数据）",
		Steps: []string{"system", "me", "list", "filter"},
		Fn: func(ctx context.Context, env *load.Env) error {
			return BookQueryFlow(ctx, &env.Env, BookQueryOptions{SimulateEmpty: false})
		},
	},
	"book-query-empty": {
		Name:  "book-query-empty",
		Type:  load.TypeFlow,
		Title: "图书查询流程（空列表）",
		Steps: []string{"system", "me", "list"},
		Fn: func(ctx context.Context, env *load.Env) error {
			return BookQueryFlow(ctx, &env.Env, BookQueryOptions{SimulateEmpty: true})
		},
	},
}
