// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import "time"

// StepRecord 记录 Flow/API 测试中单个步骤的执行详情。
type StepRecord struct {
	Index      int            `json:"index"`                 // 步骤序号，从 1 开始
	Name       string         `json:"name"`                  // 步骤名称
	Status     string         `json:"status"`                // PASS / FAIL
	StartedAt  time.Time      `json:"started_at"`            // 开始时间
	FinishedAt time.Time      `json:"finished_at"`           // 结束时间
	Duration   string         `json:"duration,omitempty"`    // 耗时（格式化字符串）
	DurationMs int64          `json:"duration_ms,omitempty"` // 耗时（毫秒）
	Detail     string         `json:"detail,omitempty"`    // 失败说明
	Result     map[string]any `json:"result,omitempty"`    // 结构化结果
}

// VarRecord 记录用例级关键变量（在报告汇总区展示）。
type VarRecord struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Fragment 是单个测试用例的报告片段，序列化为 JSON 暂存于 staging 目录。
type Fragment struct {
	RunID       string       `json:"run_id"`                // 批次运行 ID
	Package     string       `json:"package"`               // 测试包路径
	TestName    string       `json:"test_name"`             // 测试函数名
	Title       string       `json:"title"`                 // 报告标题
	Category    string       `json:"category"`              // 分类
	Description string       `json:"description"`           // 说明
	Status      string       `json:"status"`                // PASS / FAIL / SKIP
	Duration    string       `json:"duration"`              // 总耗时
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Steps       []StepRecord `json:"steps,omitempty"`       // 步骤列表
	Vars        []VarRecord  `json:"vars,omitempty"`        // 关键变量
	Logs        []string     `json:"logs,omitempty"`        // 备注日志
	Error       string       `json:"error,omitempty"`       // 错误信息
}
