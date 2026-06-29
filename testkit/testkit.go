// Package testkit 是测试用例的便捷入口，封装配置加载、客户端创建、认证与报告启用。
//
// 典型用法：
//
//	func TestExample(t *testing.T) {
//	    testkit.SkipIfDisabled(t)
//	    c := testkit.NewAuthenticatedClient(t)
//	    ctx, cancel := testkit.TestContext(t)
//	    defer cancel()
//	    // ... 调用 API 并断言
//	}
package testkit

import (
	"context"
	"testing"
	"time"

	"github.com/muyi-zcy/my-testgogogo/auth"
	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/report"
	"github.com/muyi-zcy/my-testgogogo/runtime"
	"github.com/stretchr/testify/require"
)

const defaultTestTimeout = 30 * time.Second

// SkipIfDisabled 当 configs/config.yaml 中 test.skip_integration 为 true 时跳过测试。
func SkipIfDisabled(t *testing.T) {
	t.Helper()
	if config.SkipIntegration() {
		t.Skip("integration tests disabled via configs/config.yaml test.skip_integration")
	}
}

// LoadConfig 加载项目配置，失败时通过 require 终止测试。
func LoadConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err, "load test config")
	return cfg
}

// MustVarString 读取 configs 中定义的全局字符串变量，缺失时终止测试。
func MustVarString(t *testing.T, cfg *config.Config, key string) string {
	t.Helper()
	val, err := cfg.VarString(key)
	require.NoError(t, err, "config var %q", key)
	return val
}

// MustVarInt 读取 configs 中定义的全局整型变量，缺失或类型不符时终止测试。
func MustVarInt(t *testing.T, cfg *config.Config, key string) int {
	t.Helper()
	val, err := cfg.VarInt(key)
	require.NoError(t, err, "config var %q", key)
	return val
}

// MustVars 将 configs 中的 vars 解码为自定义结构体，失败时终止测试。
func MustVars[T any](t *testing.T, cfg *config.Config) T {
	t.Helper()
	var vars T
	require.NoError(t, cfg.VarsInto(&vars), "decode config vars")
	return vars
}

// NewClient 根据配置创建未认证的 HTTP 客户端。
func NewClient(t *testing.T, cfg *config.Config) *client.Client {
	t.Helper()
	return client.NewWithRouter(cfg.BaseURL, cfg.Timeout, cfg.Router)
}

// NewAuthenticatedClient 创建已完成认证的 HTTP 客户端，可直接调用需登录的接口。
// 内部自动执行 SkipIfDisabled、LoadConfig 和 auth.Authenticate。
func NewAuthenticatedClient(t *testing.T) *client.Client {
	t.Helper()
	SkipIfDisabled(t)
	return NewAuthenticatedClientWithConfig(t, LoadConfig(t))
}

// NewAuthenticatedClientWithConfig 使用已加载的配置创建并完成认证的 HTTP 客户端。
func NewAuthenticatedClientWithConfig(t *testing.T, cfg *config.Config) *client.Client {
	t.Helper()
	c := NewClient(t, cfg)

	ctx, _ := TestContext(t)
	_, err := auth.Authenticate(ctx, c, cfg)
	require.NoError(t, err, "authenticate before test")
	return c
}

// TestContext 返回带超时的 context，默认 30 秒，通过 t.Cleanup 自动取消。
func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout)
	t.Cleanup(cancel)
	return ctx, cancel
}

// ClearAuthCache 清除 Token 本地缓存，用于登录相关测试的前置清理。
func ClearAuthCache(t *testing.T, cfg *config.Config) {
	t.Helper()
	require.NoError(t, auth.ClearCache(cfg))
}

// ReportMeta 是 report.Meta 的类型别名，便于测试代码引用。
type ReportMeta = report.Meta

// EnableReport 启用测试报告采集，返回 Reporter 接口用于记录步骤与结果。
func EnableReport(t *testing.T, meta ReportMeta) report.Reporter {
	t.Helper()
	return report.Enable(t, meta)
}

// NewScenarioEnv 构造 scenario 编排层运行时环境（匿名 + 已认证客户端、Vars、Context）。
func NewScenarioEnv(t *testing.T) *runtime.Env {
	t.Helper()
	SkipIfDisabled(t)
	cfg := LoadConfig(t)
	ctx, _ := TestContext(t)
	anon := NewClient(t, cfg)
	authClient := NewAuthenticatedClientWithConfig(t, cfg)
	return runtime.New(cfg, anon, authClient, ctx)
}

// EnableAPIReport 快捷启用单接口测试报告，分类默认为 CategoryAPI。
func EnableAPIReport(t *testing.T, title, description string) report.Reporter {
	t.Helper()
	return EnableReport(t, ReportMeta{
		Generate:    true,
		Title:       title,
		Category:    report.CategoryAPI,
		Description: description,
	})
}
