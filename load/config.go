package load

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/muyi-zcy/my-testgogogo/path"
	"gopkg.in/yaml.v3"
)

// Config 压测模块运行时配置（路径已解析为绝对路径）。
type Config struct {
	Enabled    bool
	OutputDir  string
	StagingDir string
	Defaults   Options
	Report     ReportOptions
	Scenarios  []ScenarioConfig
	RootPath   string
}

// ScenarioConfig YAML 中单个场景配置。
type ScenarioConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Enabled     bool   `yaml:"enabled"`
	Duration    string `yaml:"duration"`
	Rate        int    `yaml:"rate"`
	Concurrency int    `yaml:"concurrency"`
	Warmup      string `yaml:"warmup"`
	Timeout     string `yaml:"timeout"`
}

// Options 单次压测运行参数。
type Options struct {
	Duration       time.Duration
	Rate           int
	Concurrency    int
	Warmup         time.Duration
	Timeout        time.Duration
	BucketInterval time.Duration
}

// ReportOptions 压测报告展示选项（从 YAML load.report 读取）。
type ReportOptions struct {
	Charts       bool
	ChartMetrics []string
}

type rootConfig struct {
	Load loadYAML `yaml:"load"`
}

type loadYAML struct {
	Enabled    bool             `yaml:"enabled"`
	OutputDir  string           `yaml:"output_dir"`
	StagingDir string           `yaml:"staging_dir"`
	Defaults   defaultsYAML     `yaml:"defaults"`
	Report     reportYAML       `yaml:"report"`
	Scenarios  []ScenarioConfig `yaml:"scenarios"`
}

type reportYAML struct {
	Charts       *bool    `yaml:"charts"`
	ChartMetrics []string `yaml:"chart_metrics"`
}

type defaultsYAML struct {
	Duration       string `yaml:"duration"`
	Rate           int    `yaml:"rate"`
	Concurrency    int    `yaml:"concurrency"`
	Warmup         string `yaml:"warmup"`
	Timeout        string `yaml:"timeout"`
	BucketInterval string `yaml:"bucket_interval"`
}

// LoadConfig 从 configs/config.yaml 读取 load 段。
func LoadConfig() (*Config, error) {
	root, err := path.ModuleRoot()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(root, "configs", "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read load config: %w", err)
	}

	var rootCfg rootConfig
	if err := yaml.Unmarshal(data, &rootCfg); err != nil {
		return nil, fmt.Errorf("parse load config: %w", err)
	}

	cfg := &Config{
		Enabled:    rootCfg.Load.Enabled,
		OutputDir:  rootCfg.Load.OutputDir,
		StagingDir: rootCfg.Load.StagingDir,
		RootPath:   root,
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = "reports/load"
	}
	if cfg.StagingDir == "" {
		cfg.StagingDir = filepath.Join(cfg.OutputDir, "staging")
	}
	if !filepath.IsAbs(cfg.OutputDir) {
		cfg.OutputDir = filepath.Join(root, cfg.OutputDir)
	}
	if !filepath.IsAbs(cfg.StagingDir) {
		cfg.StagingDir = filepath.Join(root, cfg.StagingDir)
	}

	defaults, err := parseDefaults(rootCfg.Load.Defaults)
	if err != nil {
		return nil, err
	}
	cfg.Defaults = defaults
	cfg.Report = ReportOptions{
		ChartMetrics: rootCfg.Load.Report.ChartMetrics,
		Charts:       true,
	}
	if rootCfg.Load.Report.Charts != nil {
		cfg.Report.Charts = *rootCfg.Load.Report.Charts
	}
	cfg.Scenarios = rootCfg.Load.Scenarios
	return cfg, nil
}

func parseDefaults(d defaultsYAML) (Options, error) {
	opts := Options{
		Duration:       30 * time.Second,
		Rate:           20,
		Concurrency:    10,
		Warmup:         5 * time.Second,
		Timeout:        30 * time.Second,
		BucketInterval: time.Second,
	}

	if d.Rate > 0 {
		opts.Rate = d.Rate
	}
	if d.Concurrency > 0 {
		opts.Concurrency = d.Concurrency
	}

	var err error
	if d.Duration != "" {
		opts.Duration, err = time.ParseDuration(d.Duration)
		if err != nil {
			return opts, fmt.Errorf("load.defaults.duration: %w", err)
		}
	}
	if d.Warmup != "" {
		opts.Warmup, err = time.ParseDuration(d.Warmup)
		if err != nil {
			return opts, fmt.Errorf("load.defaults.warmup: %w", err)
		}
	}
	if d.Timeout != "" {
		opts.Timeout, err = time.ParseDuration(d.Timeout)
		if err != nil {
			return opts, fmt.Errorf("load.defaults.timeout: %w", err)
		}
	}
	if d.BucketInterval != "" {
		opts.BucketInterval, err = time.ParseDuration(d.BucketInterval)
		if err != nil {
			return opts, fmt.Errorf("load.defaults.bucket_interval: %w", err)
		}
	}
	return opts, nil
}

// EnabledScenarios 返回 YAML 中 enabled 的场景名列表。
func (c *Config) EnabledScenarios() []string {
	var names []string
	for _, sc := range c.Scenarios {
		if sc.Enabled {
			names = append(names, sc.Name)
		}
	}
	return names
}

// FindScenarioConfig 按名称查找场景配置。
func (c *Config) FindScenarioConfig(name string) (ScenarioConfig, bool) {
	for _, sc := range c.Scenarios {
		if sc.Name == name {
			return sc, true
		}
	}
	return ScenarioConfig{}, false
}
