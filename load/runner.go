package load

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/muyi-zcy/my-testgogogo/auth"
	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
)

// RunInput 单次压测运行输入。
type RunInput struct {
	Meta    ScenarioMeta
	Options Options
}

// RunOutput 单次压测运行输出。
type RunOutput struct {
	Meta    ScenarioMeta
	Options Options
	Metrics Snapshot
}

// Run 执行压测：预热 → 按 rate 发压 duration 时长 → 返回指标。
func Run(ctx context.Context, input RunInput, newEnv func(*config.Config, *client.Client) *Env) (*RunOutput, error) {
	if input.Meta.Fn == nil {
		return nil, fmt.Errorf("scenario %q has no function", input.Meta.Name)
	}
	if newEnv == nil {
		newEnv = defaultNewEnv
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	baseClient := client.NewForLoadWithRouter(cfg.BaseURL, cfg.Timeout, cfg.Router)
	authCtx, authCancel := context.WithTimeout(ctx, 15*time.Second)
	if _, err := auth.Authenticate(authCtx, baseClient, cfg); err != nil {
		authCancel()
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	authCancel()

	opts := input.Options
	metrics := NewMetrics(opts.BucketInterval)

	interval := time.Second / time.Duration(opts.Rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	warmupEnd := time.Now().Add(opts.Warmup)
	deadline := warmupEnd.Add(opts.Duration)

	sem := make(chan struct{}, opts.Concurrency)
	var wg sync.WaitGroup

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			if time.Now().After(deadline) {
				break loop
			}

			warming := time.Now().Before(warmupEnd)

			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				scenarioCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
				defer cancel()

				workerClient := baseClient.Clone()
				workerEnv := newEnv(cfg, workerClient)
				workerEnv.BindMetrics(metrics)

				start := time.Now()
				err := input.Meta.Fn(scenarioCtx, workerEnv)
				if !warming {
					metrics.MarkStarted()
					metrics.Record(time.Since(start), err)
				}
			}()
		}
	}

	wg.Wait()
	metrics.MarkEnded()

	out := &RunOutput{
		Meta:    input.Meta,
		Options: opts,
		Metrics: metrics.Snapshot(),
	}
	return out, nil
}

func defaultNewEnv(cfg *config.Config, c *client.Client) *Env {
	return NewEnv(cfg, c)
}
