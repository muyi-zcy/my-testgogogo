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
	Enabled    bool   `yaml:"enabled"`
	BaseDir    string `yaml:"base_dir"`
	OutputDir  string `yaml:"output_dir"`  // 兼容旧配置，等同 base_dir
	StagingDir string `yaml:"staging_dir"` // 已废弃，忽略
}

// Config 报告模块的运行时配置（路径已解析为绝对路径）。
type Config struct {
	Enabled  bool
	BaseDir  string // 报告根目录，各类型输出至 base_dir/{api,flow,load}/
	RootPath string
	runID    string
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

	baseDir := rootCfg.Report.BaseDir
	if baseDir == "" {
		baseDir = rootCfg.Report.OutputDir
	}
	if baseDir == "" {
		baseDir = "reports"
	}
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(root, baseDir)
	}

	return &Config{
		Enabled:  rootCfg.Report.Enabled,
		BaseDir:  baseDir,
		RootPath: root,
	}, nil
}

// OutputDir 返回指定类型的 Markdown 输出根目录，如 reports/api。
func (c *Config) OutputDir(kind Kind) string {
	return filepath.Join(c.BaseDir, string(kind))
}

// StagingDir 返回指定类型的 Fragment 暂存目录，如 reports/api/staging。
func (c *Config) StagingDir(kind Kind) string {
	return filepath.Join(c.OutputDir(kind), "staging")
}

// NewRunID 根据时间生成批次运行 ID，格式 YYYYMMDD-HHMMSS。
func NewRunID(now time.Time) string {
	return now.Format("20060102-150405")
}

// ReportFileName 生成批次报告的 Markdown 文件名。
func ReportFileName(kind Kind, runID string) string {
	return fmt.Sprintf("%s-report-%s.md", kind, runID)
}

// LoadEnvSummary 从应用配置中读取环境摘要信息，用于报告头部展示。
func LoadEnvSummary() (active, baseURL, username string) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", ""
	}
	return cfg.Active, cfg.BaseURL, cfg.User.Username
}
