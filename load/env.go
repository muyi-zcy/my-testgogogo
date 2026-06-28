// Package load 提供压测 Runner、指标采集与 Markdown 报告生成。
package load

import (
	"context"
	"time"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/runtime"
)

// ScenarioType 区分单接口与多步 Flow 压测场景。
type ScenarioType string

const (
	TypeAPI  ScenarioType = "api"
	TypeFlow ScenarioType = "flow"
)

// Env 是压测 scenario 的运行时环境，嵌入 runtime.Env 并与功能测试共用编排层。
type Env struct {
	runtime.Env
	metrics *Metrics
}

// NewEnv 从配置与已认证客户端构造压测环境（Client 与 AuthClient 相同，兼容旧用法）。
func NewEnv(cfg *config.Config, c *client.Client) *Env {
	return &Env{Env: *runtime.New(cfg, c, c, context.Background())}
}

// CloneForWorker 为压测 worker 创建隔离 Env，共享 metrics 采集器。
func (e *Env) CloneForWorker() *Env {
	if e == nil {
		return nil
	}
	return &Env{
		Env:     *e.Env.CloneForWorker(),
		metrics: e.metrics,
	}
}

// BindMetrics 绑定压测指标采集器（Runner 内部调用）。
func (e *Env) BindMetrics(m *Metrics) {
	e.metrics = m
	e.Env.SetStepRecorder(func(name string, d time.Duration, err error) {
		m.RecordStep(name, d, err)
	})
}

// Record 上报业务指标到当前时间桶；非压测场景（未绑定 metrics）时为 no-op。
func (e *Env) Record(name string, value float64) {
	if e.metrics != nil {
		e.metrics.RecordCustom(name, value)
	}
}

// Scenario 可被功能测试与压测共用的纯函数。
type Scenario func(ctx context.Context, env *Env) error

// ScenarioMeta 描述一个可注册压测场景。
type ScenarioMeta struct {
	Name  string
	Type  ScenarioType
	Title string
	Fn    Scenario
	Steps []string // flow 类型步骤名，预留 per-step 指标
}
