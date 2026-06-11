package load

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muyi-zcy/my-testgogogo/config"
)

// ReportInput 压测报告输入。
type ReportInput struct {
	RunID   string
	Command string
	Meta    ScenarioMeta
	Options Options
	Metrics Snapshot
	Report  ReportOptions
}

// ReportPaths 报告输出路径。
type ReportPaths struct {
	Markdown string
	JSON     string
}

// WriteReport 写入 Markdown 压测报告与 JSON staging。
func WriteReport(cfg *Config, in ReportInput) (*ReportPaths, error) {
	if in.RunID == "" {
		in.RunID = time.Now().Format("20060102-150405")
	}

	appCfg, _ := config.Load()
	active, baseURL, username := "", "", ""
	if appCfg != nil {
		active = appCfg.Active
		baseURL = appCfg.BaseURL
		username = appCfg.User.Username
	}

	if err := os.MkdirAll(cfg.StagingDir, 0o755); err != nil {
		return nil, err
	}

	jsonPath := filepath.Join(cfg.StagingDir, in.RunID, in.Meta.Name+".json")
	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		return nil, err
	}

	payload := map[string]any{
		"run_id":   in.RunID,
		"scenario": in.Meta.Name,
		"type":     in.Meta.Type,
		"title":    in.Meta.Title,
		"options":  in.Options,
		"metrics":  in.Metrics,
		"env": map[string]string{
			"active":   active,
			"base_url": baseURL,
			"username": username,
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return nil, err
	}

	dateDir := time.Now().Format("2006-01-02")
	mdDir := filepath.Join(cfg.OutputDir, dateDir)
	if err := os.MkdirAll(mdDir, 0o755); err != nil {
		return nil, err
	}

	mdName := fmt.Sprintf("load-report-%s-%s.md", in.RunID, sanitizeFileName(in.Meta.Name))
	mdPath := filepath.Join(mdDir, mdName)
	md := renderMarkdown(in, active, baseURL, username)
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return nil, err
	}

	return &ReportPaths{Markdown: mdPath, JSON: jsonPath}, nil
}

func renderMarkdown(in ReportInput, active, baseURL, username string) string {
	m := in.Metrics
	var b strings.Builder

	b.WriteString("# 压测报告：")
	b.WriteString(in.Meta.Title)
	b.WriteString("\n\n")
	b.WriteString("> 类型：load\n\n")

	b.WriteString("## 环境\n\n")
	b.WriteString(fmt.Sprintf("- 环境: %s\n", active))
	b.WriteString(fmt.Sprintf("- 地址: %s\n", baseURL))
	b.WriteString(fmt.Sprintf("- 账号: %s\n", username))
	b.WriteString(fmt.Sprintf("- 场景: %s (%s)\n", in.Meta.Name, in.Meta.Type))
	if in.Command != "" {
		b.WriteString(fmt.Sprintf("- 命令: `%s`\n", in.Command))
	}
	b.WriteString("\n")

	b.WriteString("## 参数\n\n")
	b.WriteString("| 项 | 值 |\n")
	b.WriteString("|----|-----|\n")
	b.WriteString(fmt.Sprintf("| duration | %s |\n", in.Options.Duration))
	b.WriteString(fmt.Sprintf("| rate | %d/s |\n", in.Options.Rate))
	b.WriteString(fmt.Sprintf("| concurrency | %d |\n", in.Options.Concurrency))
	b.WriteString(fmt.Sprintf("| warmup | %s |\n", in.Options.Warmup))
	b.WriteString(fmt.Sprintf("| timeout | %s |\n", in.Options.Timeout))
	b.WriteString("\n")

	b.WriteString("## 汇总\n\n")
	b.WriteString("| 指标 | 值 |\n")
	b.WriteString("|------|-----|\n")
	b.WriteString(fmt.Sprintf("| 总请求 | %d |\n", m.Total))
	b.WriteString(fmt.Sprintf("| 成功 | %d (%.1f%%) |\n", m.Success, m.SuccessPct))
	b.WriteString(fmt.Sprintf("| 失败 | %d |\n", m.Failed))
	b.WriteString(fmt.Sprintf("| 实际 QPS | %.1f |\n", m.ActualQPS))
	if m.Latency.P50 > 0 {
		b.WriteString(fmt.Sprintf("| 延迟 p50 | %s |\n", m.Latency.P50.Round(time.Millisecond)))
		b.WriteString(fmt.Sprintf("| 延迟 p95 | %s |\n", m.Latency.P95.Round(time.Millisecond)))
		b.WriteString(fmt.Sprintf("| 延迟 p99 | %s |\n", m.Latency.P99.Round(time.Millisecond)))
		b.WriteString(fmt.Sprintf("| 延迟 max | %s |\n", m.Latency.Max.Round(time.Millisecond)))
	}
	b.WriteString("\n")

	if len(m.Errors) > 0 {
		b.WriteString("## 错误分布\n\n")
		b.WriteString("| 类型 | 次数 |\n")
		b.WriteString("|------|------|\n")
		keys := sortedKeys(m.Errors)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("| %s | %d |\n", k, m.Errors[k]))
		}
		b.WriteString("\n")
	}

	b.WriteString(renderTimeSeriesSection(m, in.Report))

	return b.String()
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", ":", "-")
	return replacer.Replace(name)
}
