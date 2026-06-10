// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
//
// 工作流程：
//  1. 测试用例通过 testkit.EnableReport 启用 Collector
//  2. Collector 在测试结束时将 Fragment 写入 staging 目录
//  3. CLI 工具 my-testgogogo report 合并 go test -json 输出与 Fragment 生成 Markdown
package report

import "strings"

const (
	// CategoryAPI 单接口契约测试分类。
	CategoryAPI = "单接口测试"
	// CategoryFlow 多步骤流程编排测试分类。
	CategoryFlow = "功能测试"
)

// Meta 控制单个测试用例是否参与报告生成及其元信息。
type Meta struct {
	// Generate 为 true 时该用例参与报告生成；默认 false 不生成。
	Generate bool
	Title       string // 报告中的用例标题
	Category    string // 分类，空时根据包路径自动推断
	Description string // 用例说明
	// Standalone 为 true 时，批量运行时也额外生成单用例报告文件。
	Standalone bool
}

// categoryOrDefault 返回分类：优先使用 Meta.Category，否则根据包路径推断。
func (m Meta) categoryOrDefault(pkgPath string) string {
	if m.Category != "" {
		return m.Category
	}
	if strings.Contains(pkgPath, "/flow/") {
		return CategoryFlow
	}
	return CategoryAPI
}
