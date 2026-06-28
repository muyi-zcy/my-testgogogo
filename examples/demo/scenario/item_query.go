package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/runtime"
)

// NonexistentItemCode 用于模拟空列表分支。
const NonexistentItemCode = "__NONEXISTENT__"

// ItemQueryOptions 商品查询流程选项。
type ItemQueryOptions struct {
	SimulateEmpty bool
}

// ItemQueryFlow 商品查询流程：系统信息 → 用户信息 → 列表 → 按编码过滤（或空列表验证）。
func ItemQueryFlow(ctx context.Context, env *runtime.Env, opts ItemQueryOptions) error {
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

	in := ListItemsInput{
		PageNum:  1,
		PageSize: env.Vars.MustInt("pageSize"),
	}
	if opts.SimulateEmpty {
		in.Code = NonexistentItemCode
	}

	var pageCount int
	if err := env.RunStep("list", func() error {
		page, err := ListItems(ctx, env, in)
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
		_, err := ListItems(ctx, env, ListItemsInput{
			PageNum:  1,
			PageSize: env.Vars.MustInt("pageSize"),
			Code:     env.Vars.MustString("firstItemCode"),
		})
		return err
	})
}
