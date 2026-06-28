package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/runtime"
)

// NonexistentISBN 用于模拟空列表分支。
const NonexistentISBN = "__NONEXISTENT__"

// BookQueryOptions 图书查询流程选项。
type BookQueryOptions struct {
	SimulateEmpty bool
}

// BookQueryFlow 图书查询流程：系统信息 → 管理员信息 → 列表 → 按 ISBN 过滤（或空列表验证）。
func BookQueryFlow(ctx context.Context, env *runtime.Env, opts BookQueryOptions) error {
	if err := env.RunStep("system", func() error {
		info, err := GetSystemInfo(ctx, env)
		if err != nil {
			return err
		}
		env.Vars.Set("systemName", info["name"])
		return nil
	}); err != nil {
		return err
	}

	if err := env.RunStep("me", func() error {
		me, err := GetMe(ctx, env)
		if err != nil {
			return err
		}
		env.Vars.Set("username", me.User.Username)
		env.Vars.Set("roleCount", len(me.Roles))
		return nil
	}); err != nil {
		return err
	}

	in := ListBooksInput{
		PageNum:  1,
		PageSize: env.Vars.MustInt("pageSize"),
	}
	if opts.SimulateEmpty {
		in.ISBN = NonexistentISBN
	}

	var pageCount int
	if err := env.RunStep("list", func() error {
		page, err := ListBooks(ctx, env, in)
		if err != nil {
			return err
		}
		pageCount = len(page.Records)
		return nil
	}); err != nil {
		return err
	}

	if pageCount == 0 {
		env.Vars.Set("branch", "empty")
		return nil
	}

	env.Vars.Set("branch", "has_data")
	return env.RunStep("filter", func() error {
		_, err := ListBooks(ctx, env, ListBooksInput{
			PageNum:  1,
			PageSize: env.Vars.MustInt("pageSize"),
			ISBN:     env.Vars.MustString("firstISBN"),
		})
		return err
	})
}
