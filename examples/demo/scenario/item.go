// Package scenario 封装 demo 示例的业务编排，供 api / flow / 压测共用。
package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/examples/demo/apistep"
	"github.com/muyi-zcy/my-testgogogo/runtime"
)

// ListItemsInput 商品列表查询入参。
type ListItemsInput struct {
	PageNum  int
	PageSize int
	Code     string
}

// GetSystemInfo 获取系统信息（无需登录）。
func GetSystemInfo(ctx context.Context, env *runtime.Env) (map[string]string, error) {
	return apistep.GetSystemInfo(ctx, env.Client)
}

// GetMe 获取当前登录用户信息。
func GetMe(ctx context.Context, env *runtime.Env) (*apistep.MeResult, error) {
	return apistep.GetMe(ctx, env.AuthClient)
}

// ListItems 分页查询商品列表，并将结果写入 Vars。
func ListItems(ctx context.Context, env *runtime.Env, in ListItemsInput) (*apistep.PageResult, error) {
	page, err := apistep.ListItems(ctx, env.AuthClient, apistep.ListParams{
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
		Code:     in.Code,
	})
	if err != nil {
		return nil, err
	}

	env.Vars.Set("itemCount", int(page.Total))
	if len(page.Records) > 0 {
		first := page.Records[0]
		env.Vars.Set("firstItemCode", first.Code)
		env.Vars.Set("firstItemName", first.Name)
	}
	return page, nil
}
