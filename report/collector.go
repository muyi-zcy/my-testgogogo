// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Collector 是 Reporter 的具体实现，在测试执行期间采集步骤、变量与结构化结果。
type Collector struct {
	meta          Meta
	t             *testing.T
	packagePath   string
	startedAt     time.Time
	steps         []StepRecord
	vars          []VarRecord
	logs          []string
	finished      bool
	currentResult map[string]any // 当前步骤的结构化结果缓冲区
}

// noopCollector 是未启用报告时的空实现，所有方法均为 no-op。
type noopCollector struct{}

func (n *noopCollector) Step(string, func(*testing.T)) {}
func (n *noopCollector) Note(string)                   {}
func (n *noopCollector) Var(string, any)               {}
func (n *noopCollector) Record(string, any)            {}
func (n *noopCollector) SetResult(any)                 {}

// Reporter 报告采集接口；未启用时返回 no-op 实现。
type Reporter interface {
	// Step 执行一个子步骤并记录其状态、耗时与结构化结果。
	Step(name string, fn func(t *testing.T))
	// Note 添加备注信息，展示在报告日志区。
	Note(msg string)
	// Var 记录用例级变量（汇总区展示）。
	Var(key string, value any)
	// Record 记录当前步骤的结构化字段（仅 Step 回调内有效）。
	Record(key string, value any)
	// SetResult 一次性设置当前步骤的结构化结果（map/struct）。
	SetResult(value any)
}

// Enable 根据 Meta 配置启用报告采集。Generate=false 或全局禁用时返回 no-op。
func Enable(t *testing.T, meta Meta) Reporter {
	t.Helper()
	if !meta.Generate {
		return &noopCollector{}
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Logf("report: load config failed: %v", err)
		return &noopCollector{}
	}
	if !cfg.Enabled && !meta.Standalone {
		return &noopCollector{}
	}

	c := &Collector{
		meta:        meta,
		t:           t,
		packagePath: resolveTestPackage(3),
		startedAt:   time.Now(),
	}
	t.Cleanup(c.finish)
	return c
}

// Step 将 fn 包装为子测试执行，并记录步骤状态与耗时。
func (c *Collector) Step(name string, fn func(t *testing.T)) {
	c.t.Helper()
	stepStart := time.Now()
	stepResult := make(map[string]any)
	passed := true
	detail := ""

	c.currentResult = stepResult
	defer func() {
		c.currentResult = nil
		stepEnd := time.Now()
		elapsed := stepEnd.Sub(stepStart)

		status := "PASS"
		if !passed {
			status = "FAIL"
		}

		c.steps = append(c.steps, StepRecord{
			Index:      len(c.steps) + 1,
			Name:       name,
			Status:     status,
			StartedAt:  stepStart,
			FinishedAt: stepEnd,
			Duration:   formatDuration(elapsed),
			DurationMs: elapsed.Milliseconds(),
			Detail:     detail,
			Result:     cloneMap(stepResult),
		})
	}()

	c.t.Run(name, func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				passed = false
				detail = fmt.Sprintf("panic: %v", r)
				panic(r) // 重新抛出，不吞掉 panic
			}
		}()
		fn(t)
		if t.Failed() {
			passed = false
			if detail == "" {
				detail = "step failed"
			}
		}
	})
}

// Note 追加备注日志。
func (c *Collector) Note(msg string) {
	if msg == "" {
		return
	}
	c.logs = append(c.logs, msg)
}

// Var 记录用例级变量。
func (c *Collector) Var(key string, value any) {
	c.vars = append(c.vars, VarRecord{Key: key, Value: fmt.Sprint(value)})
}

// Record 记录当前步骤的结构化字段；若不在 Step 内则降级为 Var。
func (c *Collector) Record(key string, value any) {
	if c.currentResult == nil {
		c.Var(key, value)
		return
	}
	c.currentResult[key] = formatRecordValue(value)
}

// SetResult 将结构化结果合并到当前步骤的结果 map 中。
func (c *Collector) SetResult(value any) {
	if c.currentResult == nil {
		return
	}
	normalized, err := normalizeResult(value)
	if err != nil {
		c.currentResult["_error"] = err.Error()
		return
	}
	for key, val := range normalized {
		c.currentResult[key] = val
	}
}

// finish 在测试结束时由 t.Cleanup 调用：保存 Fragment，非批量模式时生成单用例 Markdown。
func (c *Collector) finish() {
	if c.finished {
		return
	}
	c.finished = true

	cfg, err := LoadConfig()
	if err != nil {
		c.t.Logf("report: load config failed: %v", err)
		return
	}

	fragment := c.buildFragment(cfg)
	if err := saveFragment(cfg, fragment); err != nil {
		c.t.Logf("report: save fragment failed: %v", err)
	}

	runID := currentRunID(cfg)
	// 批量模式（有 runID）时仅写 Fragment，由 CLI 合并生成报告
	if runID != "" && !c.meta.Standalone {
		return
	}

	path, err := WriteSingleMarkdown(cfg, fragment)
	if err != nil {
		c.t.Logf("report: write single markdown failed: %v", err)
		return
	}
	c.t.Logf("report: %s", path)
}

// buildFragment 将采集的数据组装为 Fragment。
func (c *Collector) buildFragment(cfg *Config) Fragment {
	status := "PASS"
	if c.t.Failed() {
		status = "FAIL"
	} else if c.t.Skipped() {
		status = "SKIP"
	}

	finishedAt := time.Now()
	pkgPath := c.packagePath
	if pkgPath == "" {
		pkgPath = c.packagePathFromRuntime()
	}

	return Fragment{
		RunID:       currentRunID(cfg),
		Package:     pkgPath,
		TestName:    c.t.Name(),
		Title:       titleOrDefault(c.meta, c.t.Name()),
		Category:    c.meta.categoryOrDefault(pkgPath),
		Description: c.meta.Description,
		Status:      status,
		Duration:    formatDuration(finishedAt.Sub(c.startedAt)),
		StartedAt:   c.startedAt,
		FinishedAt:  finishedAt,
		Steps:       append([]StepRecord(nil), c.steps...),
		Vars:        append([]VarRecord(nil), c.vars...),
		Logs:        append([]string(nil), c.logs...),
	}
}

// packagePathFromRuntime 从调用栈推断测试包路径（回退方案）。
func (c *Collector) packagePathFromRuntime() string {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	name := fn.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[:idx]
	}
	if idx := strings.Index(name, "/examples/"); idx >= 0 {
		return name[idx+1:]
	}
	if idx := strings.Index(name, "/tests/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// resolveTestPackage 从调用栈的文件路径解析测试包相对路径。
func resolveTestPackage(skip int) string {
	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	for _, marker := range []string{"/examples/", "/tests/"} {
		if idx := strings.Index(file, marker); idx >= 0 {
			return filepath.Dir(file[idx+1:])
		}
	}
	return ""
}

// titleOrDefault 返回报告标题，空时使用测试函数名。
func titleOrDefault(meta Meta, testName string) string {
	if meta.Title != "" {
		return meta.Title
	}
	return testName
}

// saveFragment 将 Fragment 序列化为 JSON 写入 staging 目录。
func saveFragment(cfg *Config, fragment Fragment) error {
	runID := fragment.RunID
	if runID == "" {
		runID = fragment.FinishedAt.Format("20060102-150405")
		fragment.RunID = runID
	}

	dir := filepath.Join(cfg.StagingDir, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fileName := sanitizeFileName(fmt.Sprintf("%s__%s.json", fragment.Package, fragment.TestName))
	path := filepath.Join(dir, fileName)

	payload, err := json.MarshalIndent(fragment, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

var runIDMu sync.Mutex

// currentRunID 获取当前批次运行 ID，优先从环境变量 MY_TESTGOGOGO_REPORT_RUN_ID 读取。
func currentRunID(cfg *Config) string {
	if v := os.Getenv("MY_TESTGOGOGO_REPORT_RUN_ID"); v != "" {
		return v
	}
	runIDMu.Lock()
	defer runIDMu.Unlock()
	if cfg.runID != "" {
		return cfg.runID
	}
	return ""
}

// sanitizeFileName 将字符串中的非法文件名字符替换为下划线。
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	return replacer.Replace(name)
}

// formatDuration 将 Duration 格式化为人类可读的字符串。
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
