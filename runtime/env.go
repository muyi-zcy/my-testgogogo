// Package runtime 提供功能测试与压测共用的运行时环境。
package runtime

import (
	"context"
	"time"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/flow"
)

type stepRecorder func(name string, latency time.Duration, err error)

// Env 是 scenario 编排层的运行时环境，功能测试与压测共用。
type Env struct {
	CTX        context.Context
	Config     *config.Config
	Client     *client.Client // 匿名客户端（无需登录的接口）
	AuthClient *client.Client // 已认证客户端
	Vars       *flow.Vars
	onStep     stepRecorder
}

// NewVars 创建跨步骤变量容器，合并 flow 种子数据与配置 vars。
func NewVars(cfg *config.Config) *flow.Vars {
	vars := flow.NewVars(flow.DefaultSeed())
	if cfg != nil {
		if ps, err := cfg.VarInt("page_size"); err == nil {
			vars.Set("pageSize", ps)
		}
	}
	return vars
}

// New 构造运行时环境。anon/auth 可只传其一：仅 auth 时 Client 与 AuthClient 相同。
func New(cfg *config.Config, anon, auth *client.Client, ctx context.Context) *Env {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg != nil && anon == nil && auth == nil {
		anon = client.NewWithRouter(cfg.BaseURL, cfg.Timeout, cfg.Router)
	}
	if auth == nil {
		auth = anon
	}
	if anon == nil {
		anon = auth
	}
	return &Env{
		CTX:        ctx,
		Config:     cfg,
		Client:     anon,
		AuthClient: auth,
		Vars:       NewVars(cfg),
	}
}

// SetStepRecorder 注册步骤耗时回调（压测 BindMetrics 时注入）。
func (e *Env) SetStepRecorder(fn stepRecorder) {
	if e == nil {
		return
	}
	e.onStep = fn
}

// RunStep 执行 Flow 中的一个命名步骤，并在配置了 recorder 时上报耗时。
func (e *Env) RunStep(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	if e != nil && e.onStep != nil {
		e.onStep(name, time.Since(start), err)
	}
	return err
}

// CloneForWorker 为压测 worker 创建隔离 Env（独立 Vars，共享 Client 与 step recorder）。
func (e *Env) CloneForWorker() *Env {
	if e == nil {
		return nil
	}
	return &Env{
		CTX:        e.CTX,
		Config:     e.Config,
		Client:     e.Client,
		AuthClient: e.AuthClient,
		Vars:       NewVars(e.Config),
		onStep:     e.onStep,
	}
}
