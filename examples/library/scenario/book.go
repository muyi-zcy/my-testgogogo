// Package scenario 封装 library 示例的业务编排，供 api / flow / 压测共用。
package scenario

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/examples/library/apistep"
	"github.com/muyi-zcy/my-testgogogo/runtime"
)

// ListBooksInput 图书列表查询入参。
type ListBooksInput struct {
	PageNum  int
	PageSize int
	ISBN     string
}

// GetSystemInfo 获取系统信息（无需登录）。
func GetSystemInfo(ctx context.Context, env *runtime.Env) (map[string]string, error) {
	return apistep.GetSystemInfo(ctx, env.Client)
}

// GetMe 获取当前管理员信息。
func GetMe(ctx context.Context, env *runtime.Env) (*apistep.MeResult, error) {
	return apistep.GetMe(ctx, env.AuthClient)
}

// ListBooks 分页查询图书列表，并将结果写入 Vars。
func ListBooks(ctx context.Context, env *runtime.Env, in ListBooksInput) (*apistep.PageResult, error) {
	page, err := apistep.ListBooks(ctx, env.AuthClient, apistep.ListParams{
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
		ISBN:     in.ISBN,
	})
	if err != nil {
		return nil, err
	}

	env.Vars.Set("bookCount", int(page.Total))
	if len(page.Records) > 0 {
		first := page.Records[0]
		env.Vars.Set("firstISBN", first.ISBN)
		env.Vars.Set("firstTitle", first.Title)
	}
	return page, nil
}
