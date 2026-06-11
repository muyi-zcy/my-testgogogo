// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// GoTestEvent 表示 go test -json 输出中的一条事件。
type GoTestEvent struct {
	Time    time.Time
	Package string
	Test    string  // 空表示包级事件
	Action  string  // run / pass / fail / skip / output 等
	Elapsed float64 // 耗时（秒）
	Output  string
}

// Summary 是批次报告的汇总数据，包含 go test 事件与各用例 Fragment。
type Summary struct {
	RunID     string
	Kind      Kind // 报告类型：api / flow
	Generated time.Time
	GoVersion string
	Active    string // 运行环境
	BaseURL   string // 接口地址
	Username  string // 测试账号
	Command   string // 执行命令
	Total     int    // 用例总数
	Passed    int
	Failed    int
	Skipped   int
	Duration  time.Duration
	Events    []GoTestEvent
	Fragments []Fragment
}

// WriteSingleMarkdown 为单个 Fragment 生成独立的 Markdown 报告文件。
func WriteSingleMarkdown(cfg *Config, fragment Fragment) (string, error) {
	kind := ResolveKind(fragment.Kind, fragment.Package)

	summary := Summary{
		RunID:     fragment.RunID,
		Kind:      kind,
		Generated: time.Now(),
		GoVersion: runtime.Version(),
		Fragments: []Fragment{fragment},
	}
	summary.applyEnv()

	if fragment.RunID == "" {
		fragment.RunID = fragment.FinishedAt.Format("20060102-150405")
	}
	dateDir := fragment.FinishedAt.Format("2006-01-02")
	dir := filepath.Join(cfg.OutputDir(kind), dateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s-%s-%s.md",
		kind.ReportPrefix(),
		fragment.FinishedAt.Format("20060102-150405"),
		sanitizeFileName(fragment.TestName),
	)
	path := filepath.Join(dir, fileName)
	return path, os.WriteFile(path, []byte(renderMarkdown(summary, true)), 0o644)
}

// WriteBatchMarkdown 生成指定类型的批次 Markdown 报告，并更新 reports/<kind>/latest.md。
func WriteBatchMarkdown(cfg *Config, summary Summary) (string, string, error) {
	summary.applyEnv()
	if summary.RunID == "" {
		summary.RunID = NewRunID(summary.Generated)
	}
	if summary.Kind == "" {
		summary.Kind = KindAPI
	}

	dateDir := summary.Generated.Format("2006-01-02")
	dir := filepath.Join(cfg.OutputDir(summary.Kind), dateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	fileName := ReportFileName(summary.Kind, summary.RunID)
	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, []byte(renderMarkdown(summary, false)), 0o644); err != nil {
		return "", "", err
	}

	latest := filepath.Join(cfg.OutputDir(summary.Kind), "latest.md")
	if err := os.WriteFile(latest, []byte(renderMarkdown(summary, false)), 0o644); err != nil {
		return path, "", err
	}
	return path, latest, nil
}

// applyEnv 填充环境摘要信息并计算统计数据。
func (s *Summary) applyEnv() {
	if s.Active == "" || s.BaseURL == "" || s.Username == "" {
		active, baseURL, username := LoadEnvSummary()
		if s.Active == "" {
			s.Active = active
		}
		if s.BaseURL == "" {
			s.BaseURL = baseURL
		}
		if s.Username == "" {
			s.Username = username
		}
	}
	if s.GoVersion == "" {
		s.GoVersion = runtime.Version()
	}
	s.computeStats()
}

// computeStats 从 Fragment 或 go test 事件统计用例数量。
func (s *Summary) computeStats() {
	if len(s.Fragments) > 0 {
		s.Total = len(s.Fragments)
		for _, f := range s.Fragments {
			switch f.Status {
			case "PASS":
				s.Passed++
			case "FAIL":
				s.Failed++
			case "SKIP":
				s.Skipped++
			}
		}
		return
	}

	topLevel := collectTopLevelResults(s.Events)
	s.Total = len(topLevel)
	for _, action := range topLevel {
		switch action {
		case "pass":
			s.Passed++
		case "fail":
			s.Failed++
		case "skip":
			s.Skipped++
		}
	}
}

// collectTopLevelResults 收集顶层测试（不含子测试）的最终结果。
func collectTopLevelResults(events []GoTestEvent) map[string]string {
	results := map[string]string{}
	for _, ev := range events {
		// 跳过包级事件和子测试（名称含 /）
		if ev.Test == "" || strings.Contains(ev.Test, "/") {
			continue
		}
		key := ev.Package + "\x00" + ev.Test
		switch ev.Action {
		case "pass", "fail", "skip":
			results[key] = ev.Action
		}
	}
	return results
}

// renderMarkdown 将 Summary 渲染为 Markdown 文本。single=true 时不输出分类统计。
func renderMarkdown(s Summary, single bool) string {
	var b strings.Builder

	title := KindAPI.Title()
	if s.Kind != "" {
		title = s.Kind.Title()
	}
	if single && len(s.Fragments) == 1 && s.Fragments[0].Title != "" {
		b.WriteString("# ")
		b.WriteString(s.Fragments[0].Title)
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("> 类型：%s · %s\n\n", s.Fragments[0].Kind, s.Fragments[0].Category))
	} else {
		b.WriteString("# my-testgogogo ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	b.WriteString("| 项目 | 值 |\n|------|----|\n")
	b.WriteString(fmt.Sprintf("| 生成时间 | %s |\n", s.Generated.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("| 报告编号 | %s |\n", s.RunID))
	b.WriteString(fmt.Sprintf("| Go 版本 | %s |\n", s.GoVersion))
	b.WriteString(fmt.Sprintf("| 运行环境 | %s |\n", valueOrDash(s.Active)))
	b.WriteString(fmt.Sprintf("| 接口地址 | %s |\n", valueOrDash(s.BaseURL)))
	b.WriteString(fmt.Sprintf("| 测试账号 | %s |\n", valueOrDash(s.Username)))
	if s.Command != "" {
		b.WriteString(fmt.Sprintf("| 执行命令 | `%s` |\n", s.Command))
	}

	b.WriteString("\n## 总览\n\n")
	overall := "通过"
	if s.Failed > 0 {
		overall = "失败"
	} else if s.Total == 0 && len(s.Fragments) > 0 {
		overall = fragmentOverall(s.Fragments)
	}

	passRate := 0.0
	if s.Total > 0 {
		passRate = float64(s.Passed) / float64(s.Total) * 100
	}
	b.WriteString(fmt.Sprintf("**结果：%s**", overall))
	if s.Total > 0 {
		b.WriteString(fmt.Sprintf(" · 通过率 **%.1f%%**", passRate))
	}
	if s.Duration > 0 {
		b.WriteString(fmt.Sprintf(" · 耗时 **%s**", formatDuration(s.Duration)))
	}
	b.WriteString("\n\n")

	if s.Total > 0 {
		b.WriteString("| 指标 | 数量 |\n|------|------|\n")
		b.WriteString(fmt.Sprintf("| 用例总数 | %d |\n", s.Total))
		b.WriteString(fmt.Sprintf("| 通过 | %d |\n", s.Passed))
		b.WriteString(fmt.Sprintf("| 失败 | %d |\n", s.Failed))
		b.WriteString(fmt.Sprintf("| 跳过 | %d |\n\n", s.Skipped))
	}

	if !single && s.Kind == "" {
		b.WriteString(renderCategoryStats(s))
	}
	b.WriteString(renderAllCases(s))
	b.WriteString(renderDetailedFragments(s.Fragments))

	return b.String()
}

// renderCategoryStats 按分类（单接口/功能测试）统计通过/失败/跳过数量。
func renderCategoryStats(s Summary) string {
	type stat struct{ pass, fail, skip, total int }
	stats := map[string]*stat{}

	for key, action := range collectTopLevelResults(s.Events) {
		pkg := key
		if idx := strings.Index(key, "\x00"); idx >= 0 {
			pkg = key[:idx]
		}
		category := CategoryAPI
		if strings.Contains(pkg, "/flow/") || strings.Contains(pkg, "/tests/flow/") {
			category = CategoryFlow
		}
		if stats[category] == nil {
			stats[category] = &stat{}
		}
		stats[category].total++
		switch action {
		case "pass":
			stats[category].pass++
		case "fail":
			stats[category].fail++
		case "skip":
			stats[category].skip++
		}
	}

	if len(stats) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("### 分类统计\n\n")
	b.WriteString("| 分类 | 通过 | 失败 | 跳过 | 合计 |\n|------|------|------|------|------|\n")
	for _, category := range []string{CategoryFlow, CategoryAPI} {
		st, ok := stats[category]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n", category, st.pass, st.fail, st.skip, st.total))
	}
	b.WriteString("\n")
	return b.String()
}

// renderAllCases 渲染用例明细表，按分类和包路径分组。
func renderAllCases(s Summary) string {
	type caseResult struct {
		Package  string
		Test     string
		Status   string
		Duration string
	}
	byTest := map[string]caseResult{}

	for _, ev := range s.Events {
		if ev.Test == "" {
			continue
		}
		key := ev.Package + "\x00" + ev.Test
		item := byTest[key]
		item.Package = ev.Package
		item.Test = ev.Test
		if ev.Action == "pass" || ev.Action == "fail" || ev.Action == "skip" {
			item.Status = strings.ToUpper(ev.Action)
			if ev.Elapsed > 0 {
				item.Duration = formatDuration(time.Duration(ev.Elapsed * float64(time.Second)))
			}
		}
		byTest[key] = item
	}

	if len(byTest) == 0 {
		return ""
	}

	items := make([]caseResult, 0, len(byTest))
	for _, item := range byTest {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Package == items[j].Package {
			return items[i].Test < items[j].Test
		}
		return items[i].Package < items[j].Package
	})

	var b strings.Builder
	b.WriteString("## 用例明细\n\n")

	currentPkg := ""
	currentCategory := ""
	for _, item := range items {
		if strings.Contains(item.Test, "/") {
			continue // 跳过子测试行
		}
		category := CategoryAPI
		if strings.Contains(item.Package, "/tests/flow/") {
			category = CategoryFlow
		}
		if category != currentCategory {
			currentCategory = category
			b.WriteString(fmt.Sprintf("### %s\n\n", category))
			currentPkg = ""
		}
		pkgShort := shortPackage(item.Package)
		if pkgShort != currentPkg {
			currentPkg = pkgShort
			b.WriteString(fmt.Sprintf("#### `%s`\n\n", pkgShort))
			b.WriteString("| 用例 | 结果 | 耗时 |\n|------|------|------|\n")
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", item.Test, item.Status, valueOrDash(item.Duration)))
	}
	b.WriteString("\n")
	return b.String()
}

// renderDetailedFragments 渲染各用例的详细报告（步骤、变量、备注）。
func renderDetailedFragments(fragments []Fragment) string {
	if len(fragments) == 0 {
		return ""
	}

	sorted := append([]Fragment(nil), fragments...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Category == sorted[j].Category {
			if sorted[i].Package == sorted[j].Package {
				return sorted[i].TestName < sorted[j].TestName
			}
			return sorted[i].Package < sorted[j].Package
		}
		return sorted[i].Category < sorted[j].Category
	})

	var b strings.Builder
	b.WriteString("## 详细报告\n\n")
	for _, f := range sorted {
		b.WriteString(fmt.Sprintf("### %s\n\n", f.Title))
		b.WriteString(fmt.Sprintf("- 用例：`%s`\n", f.TestName))
		b.WriteString(fmt.Sprintf("- 分类：%s\n", f.Category))
		b.WriteString(fmt.Sprintf("- 包路径：`%s`\n", shortPackage(f.Package)))
		b.WriteString(fmt.Sprintf("- 结果：**%s**\n", f.Status))
		b.WriteString(fmt.Sprintf("- 耗时：%s\n", valueOrDash(f.Duration)))
		if f.Description != "" {
			b.WriteString(fmt.Sprintf("- 说明：%s\n", f.Description))
		}
		if f.Error != "" {
			b.WriteString(fmt.Sprintf("- 错误：%s\n", f.Error))
		}
		b.WriteString("\n")

		if len(f.Steps) > 0 {
			b.WriteString("#### 步骤\n\n")
			b.WriteString(renderStepSummaryTable(f.Steps))
			b.WriteString(renderStepDetails(f.Steps))
		}

		if len(f.Vars) > 0 {
			b.WriteString("#### 关键变量\n\n")
			b.WriteString("| 变量 | 值 |\n|------|----|\n")
			for _, v := range f.Vars {
				b.WriteString(fmt.Sprintf("| %s | %s |\n", v.Key, v.Value))
			}
			b.WriteString("\n")
		}

		if len(f.Logs) > 0 {
			b.WriteString("#### 备注\n\n")
			for _, log := range f.Logs {
				b.WriteString(fmt.Sprintf("- %s\n", log))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

// renderStepSummaryTable 渲染步骤汇总表。
func renderStepSummaryTable(steps []StepRecord) string {
	var b strings.Builder
	b.WriteString("| # | 步骤 | 结果 | 开始时间 | 结束时间 | 耗时 |\n")
	b.WriteString("|---|------|------|----------|----------|------|\n")
	for _, step := range steps {
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			step.Index,
			step.Name,
			step.Status,
			formatTimestamp(step.StartedAt),
			formatTimestamp(step.FinishedAt),
			valueOrDash(step.Duration),
		))
	}
	b.WriteString("\n")
	return b.String()
}

// renderStepDetails 渲染每个步骤的详细信息和结构化结果。
func renderStepDetails(steps []StepRecord) string {
	var b strings.Builder
	for _, step := range steps {
		b.WriteString(fmt.Sprintf("##### %d. %s\n\n", step.Index, step.Name))
		b.WriteString("| 项目 | 值 |\n|------|----|\n")
		b.WriteString(fmt.Sprintf("| 结果 | %s |\n", step.Status))
		b.WriteString(fmt.Sprintf("| 开始时间 | %s |\n", formatTimestamp(step.StartedAt)))
		b.WriteString(fmt.Sprintf("| 结束时间 | %s |\n", formatTimestamp(step.FinishedAt)))
		b.WriteString(fmt.Sprintf("| 耗时 | %s |\n", valueOrDash(step.Duration)))
		if step.DurationMs > 0 {
			b.WriteString(fmt.Sprintf("| 耗时(ms) | %d |\n", step.DurationMs))
		}
		if step.Detail != "" {
			b.WriteString(fmt.Sprintf("| 说明 | %s |\n", step.Detail))
		}
		b.WriteString("\n")

		if len(step.Input) > 0 {
			b.WriteString("**入参：**\n\n")
			b.WriteString(renderStructuredResult(step.Input))
			b.WriteString("\n")
		}
		if len(step.Result) > 0 {
			b.WriteString("**结构化结果：**\n\n")
			b.WriteString(renderStructuredResult(step.Result))
			b.WriteString("\n")
		}
		if len(step.Response) > 0 {
			b.WriteString("**接口响应：**\n\n")
			b.WriteString(renderStructuredResult(step.Response))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// renderStructuredResult 将结构化结果渲染为表格和 JSON 代码块。
func renderStructuredResult(result map[string]any) string {
	flat := flattenResult("", result)
	if len(flat) == 0 {
		return "_无_\n"
	}

	var b strings.Builder
	b.WriteString("| 字段 | 值 |\n|------|----|\n")
	for _, item := range flat {
		b.WriteString(fmt.Sprintf("| %s | %s |\n", item.Key, item.Value))
	}
	b.WriteString("\n```json\n")
	if payload, err := json.MarshalIndent(result, "", "  "); err == nil {
		b.WriteString(string(payload))
	}
	b.WriteString("\n```\n\n")
	return b.String()
}

// flatField 表示扁平化后的结构化结果字段。
type flatField struct {
	Key   string
	Value string
}

// flattenResult 递归扁平化嵌套 map，生成点分键名。
func flattenResult(prefix string, value map[string]any) []flatField {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]flatField, 0, len(value))
	for _, key := range keys {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch typed := value[key].(type) {
		case map[string]any:
			fields = append(fields, flattenResult(fullKey, typed)...)
		default:
			fields = append(fields, flatField{
				Key:   fullKey,
				Value: fmt.Sprint(typed),
			})
		}
	}
	return fields
}

// formatTimestamp 格式化时间戳，零值返回 "-"。
func formatTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return "-"
	}
	return ts.Format("2006-01-02 15:04:05.000")
}

// fragmentOverall 根据 Fragment 列表推断整体结果。
func fragmentOverall(fragments []Fragment) string {
	for _, f := range fragments {
		if f.Status == "FAIL" {
			return "失败"
		}
	}
	return "通过"
}

// shortPackage 截取包路径的简短形式（从 /tests/ 或完整路径开始）。
func shortPackage(pkg string) string {
	if idx := strings.Index(pkg, "/tests/"); idx >= 0 {
		return pkg[idx+1:]
	}
	return pkg
}

// valueOrDash 空字符串返回 "-"，否则返回原值。
func valueOrDash(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
