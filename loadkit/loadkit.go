// Package loadkit 提供压测便捷入口，类似 testkit。
package loadkit

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/flow"
	"github.com/muyi-zcy/my-testgogogo/load"
)

// NewEnv 从配置与客户端构造压测环境。
func NewEnv(cfg *config.Config, c *client.Client) *load.Env {
	vars := flow.NewVars(map[string]any{"pageSize": 10})
	if ps, err := cfg.VarInt("page_size"); err == nil {
		vars.Set("pageSize", ps)
	}
	return &load.Env{Client: c, Vars: vars, Config: cfg}
}

// Overrides CLI 可覆盖的压测参数；nil 表示使用 YAML 默认值。
type Overrides struct {
	Duration    *time.Duration
	Rate        *int
	Concurrency *int
	Warmup      *time.Duration
	Timeout     *time.Duration
}

// RunScenario 运行单个压测场景并写报告。
func RunScenario(ctx context.Context, registry map[string]load.ScenarioMeta, name string, overrides Overrides, runID, command string) error {
	meta, ok := registry[name]
	if !ok {
		return fmt.Errorf("scenario %q not found in registry", name)
	}

	loadCfg, err := load.LoadConfig()
	if err != nil {
		return err
	}

	scCfg := load.ScenarioConfig{Name: name, Enabled: true}
	if yamlSc, found := loadCfg.FindScenarioConfig(name); found {
		scCfg = yamlSc
	}

	opts, err := resolveOptions(loadCfg, scCfg, overrides)
	if err != nil {
		return err
	}

	out, err := load.Run(ctx, load.RunInput{Meta: meta, Options: opts}, NewEnv)
	if err != nil {
		return err
	}

	if runID == "" {
		runID = time.Now().Format("20060102-150405")
	}
	paths, err := load.WriteReport(loadCfg, load.ReportInput{
		RunID:   runID,
		Command: command,
		Meta:    meta,
		Options: opts,
		Metrics: out.Metrics,
		Report:  loadCfg.Report,
	})
	if err != nil {
		return err
	}

	fmt.Printf("load: %s (%d ok / %d total, %.1f qps)\n",
		name, out.Metrics.Success, out.Metrics.Total, out.Metrics.ActualQPS)
	fmt.Printf("report: %s\n", paths.Markdown)
	fmt.Printf("staging: %s\n", paths.JSON)
	return nil
}

// Main 解析 os.Args 并执行压测，返回进程退出码。
func Main(registry map[string]load.ScenarioMeta) int {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	scenario := fs.String("scenario", "", "scenario name")
	all := fs.Bool("all", false, "run all enabled scenarios from config")
	duration := fs.Duration("duration", 0, "load duration")
	rate := fs.Int("rate", 0, "target requests per second")
	concurrency := fs.Int("concurrency", 0, "max concurrent workers")
	warmup := fs.Duration("warmup", 0, "warmup duration")
	timeout := fs.Duration("timeout", 0, "per-scenario timeout")
	runID := fs.String("run-id", "", "report run id")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}

	// flag.Duration 无法区分「未设置」与「设为 0」；用 Visit 检测
	var overrides Overrides
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "duration":
			v := duration
			overrides.Duration = v
		case "rate":
			v := rate
			overrides.Rate = v
		case "concurrency":
			v := concurrency
			overrides.Concurrency = v
		case "warmup":
			v := warmup
			overrides.Warmup = v
		case "timeout":
			v := timeout
			overrides.Timeout = v
		}
	})

	command := "load " + strings.Join(os.Args[1:], " ")
	ctx := context.Background()

	if *all {
		loadCfg, err := load.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "load config: %v\n", err)
			return 1
		}
		names := loadCfg.EnabledScenarios()
		if len(names) == 0 {
			fmt.Fprintf(os.Stderr, "no enabled scenarios in config\n")
			return 1
		}
		rid := *runID
		for i, name := range names {
			if i > 0 && rid != "" {
				rid = *runID + "-" + name
			}
			if err := RunScenario(ctx, registry, name, overrides, rid, command); err != nil {
				fmt.Fprintf(os.Stderr, "load %s: %v\n", name, err)
				return 1
			}
		}
		return 0
	}

	if *scenario == "" {
		fmt.Fprintf(os.Stderr, "usage: load --scenario <name> | load --all\n")
		return 2
	}

	if err := RunScenario(ctx, registry, *scenario, overrides, *runID, command); err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		return 1
	}
	return 0
}

func resolveOptions(loadCfg *load.Config, sc load.ScenarioConfig, overrides Overrides) (load.Options, error) {
	base := loadCfg.Defaults

	if sc.Duration != "" {
		d, err := time.ParseDuration(sc.Duration)
		if err != nil {
			return base, err
		}
		base.Duration = d
	}
	if sc.Rate > 0 {
		base.Rate = sc.Rate
	}
	if sc.Concurrency > 0 {
		base.Concurrency = sc.Concurrency
	}
	if sc.Warmup != "" {
		w, err := time.ParseDuration(sc.Warmup)
		if err != nil {
			return base, err
		}
		base.Warmup = w
	}
	if sc.Timeout != "" {
		t, err := time.ParseDuration(sc.Timeout)
		if err != nil {
			return base, err
		}
		base.Timeout = t
	}

	if overrides.Duration != nil {
		base.Duration = *overrides.Duration
	}
	if overrides.Rate != nil && *overrides.Rate > 0 {
		base.Rate = *overrides.Rate
	}
	if overrides.Concurrency != nil && *overrides.Concurrency > 0 {
		base.Concurrency = *overrides.Concurrency
	}
	if overrides.Warmup != nil {
		base.Warmup = *overrides.Warmup
	}
	if overrides.Timeout != nil && *overrides.Timeout > 0 {
		base.Timeout = *overrides.Timeout
	}

	if base.Rate <= 0 {
		return base, fmt.Errorf("rate must be positive")
	}
	if base.Concurrency <= 0 {
		return base, fmt.Errorf("concurrency must be positive")
	}
	if base.Duration <= 0 {
		return base, fmt.Errorf("duration must be positive")
	}
	return base, nil
}
