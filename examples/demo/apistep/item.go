// Package apistep 封装 demo 示例的 API 调用步骤，供单接口测试和 Flow 流程复用。
package apistep

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/muyi-zcy/my-testgogogo/client"
)

// Item 商品实体，对应后端 /api/items 响应中的 records 元素。
type Item struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// PageResult 分页查询结果。
type PageResult struct {
	Current int64  `json:"current"`
	Size    int64  `json:"size"`
	Total   int64  `json:"total"`
	Pages   int64  `json:"pages"`
	Records []Item `json:"records"`
}

// UserInfo 用户基本信息。
type UserInfo struct {
	Username string `json:"username"`
	NickName string `json:"nickName"`
}

// MeResult /api/auth/me 接口响应。
type MeResult struct {
	User        UserInfo `json:"user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// ListParams 商品列表查询参数。
type ListParams struct {
	PageNum  int
	PageSize int
	Code     string // 可选，按商品编码过滤
}

// defaults 填充分页参数默认值。
func (p ListParams) defaults() ListParams {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return p
}

// GetSystemInfo 获取系统信息，无需认证。
func GetSystemInfo(ctx context.Context, c *client.Client) (map[string]string, error) {
	var info map[string]string
	if err := c.DoJSON(ctx, "GET", "/api/system/info", &info, client.WithoutAuth()); err != nil {
		return nil, fmt.Errorf("get system info: %w", err)
	}
	return info, nil
}

// GetMe 获取当前登录用户信息。
func GetMe(ctx context.Context, c *client.Client) (*MeResult, error) {
	var result MeResult
	if err := c.DoJSON(ctx, "GET", "/api/auth/me", &result); err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}
	return &result, nil
}

// ListItems 分页查询商品列表。
func ListItems(ctx context.Context, c *client.Client, params ListParams) (*PageResult, error) {
	p := params.defaults()
	query := url.Values{}
	query.Set("pageNum", strconv.Itoa(p.PageNum))
	query.Set("pageSize", strconv.Itoa(p.PageSize))
	if p.Code != "" {
		query.Set("code", p.Code)
	}

	var page PageResult
	if err := c.DoJSON(ctx, "GET", "/api/items", &page, client.WithQuery(query)); err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return &page, nil
}

// Logout 登出并清除客户端 Token。
func Logout(ctx context.Context, c *client.Client) error {
	resp, err := c.Get(ctx, "/api/auth/logout")
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("logout failed: status=%d body=%s", resp.StatusCode, string(resp.Body))
	}
	c.SetToken("")
	return nil
}
