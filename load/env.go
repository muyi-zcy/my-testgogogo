// Package load 提供压测 Runner、指标采集与 Markdown 报告生成。
package load

import (
	"context"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/flow"
)

// ScenarioType 区分单接口与多步 Flow 压测场景。
type ScenarioType string

const (
	TypeAPI  ScenarioType = "api"
	TypeFlow ScenarioType = "flow"
)

// Env 是压测 scenario 的运行时环境，与功能测试共用 apistep 与 config。
type Env struct {
	Client  *client.Client
	Vars    *flow.Vars
	Config  *config.Config
	metrics *Metrics
}

// BindMetrics 绑定压测指标采集器（Runner 内部调用）。
func (e *Env) BindMetrics(m *Metrics) {
	e.metrics = m
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
