// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/muyi-zcy/my-testgogogo/config"
	"github.com/muyi-zcy/my-testgogogo/path"
	"gopkg.in/yaml.v3"
)

// rootConfig 从 configs/config.yaml 中读取 report 相关配置。
type rootConfig struct {
	Report reportConfig `yaml:"report"`
}

// reportConfig YAML 中的报告配置段。
type reportConfig struct {
	Enabled    bool   `yaml:"enabled"`     // 是否全局启用报告
	OutputDir  string `yaml:"output_dir"`  // Markdown 输出目录
	StagingDir string `yaml:"staging_dir"` // Fragment JSON 暂存目录
}

// Config 报告模块的运行时配置（路径已解析为绝对路径）。
type Config struct {
	Enabled    bool
	OutputDir  string
	StagingDir string
	RootPath   string
	runID      string // 当前批次运行 ID，由环境变量注入
}

// LoadConfig 从项目 configs/config.yaml 加载报告配置。
func LoadConfig() (*Config, error) {
	root, err := path.ModuleRoot()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(root, "configs", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var rootCfg rootConfig
	if err := yaml.Unmarshal(data, &rootCfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg := &Config{
		Enabled:    rootCfg.Report.Enabled,
		OutputDir:  rootCfg.Report.OutputDir,
		StagingDir: rootCfg.Report.StagingDir,
		RootPath:   root,
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = "reports"
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

	return cfg, nil
}

// NewRunID 根据时间生成批次运行 ID，格式 YYYYMMDD-HHMMSS。
func NewRunID(now time.Time) string {
	return now.Format("20060102-150405")
}

// ReportFileName 生成批次报告的 Markdown 文件名。
func ReportFileName(runID string) string {
	return fmt.Sprintf("test-report-%s.md", runID)
}

// LoadEnvSummary 从应用配置中读取环境摘要信息，用于报告头部展示。
func LoadEnvSummary() (active, baseURL, username string) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", ""
	}
	return cfg.Active, cfg.BaseURL, cfg.User.Username
}
