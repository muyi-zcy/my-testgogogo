package report

import "strings"

// Kind 报告类型：api / flow / load，对应独立目录与文件名前缀。
type Kind string

const (
	KindAPI  Kind = "api"
	KindFlow Kind = "flow"
	KindLoad Kind = "load"
)

// KindFromPackage 根据测试包路径推断报告类型。
func KindFromPackage(pkgPath string) Kind {
	if strings.Contains(pkgPath, "/flow/") {
		return KindFlow
	}
	return KindAPI
}

// ResolveKind 解析 Fragment 或 Meta 的报告类型。
func ResolveKind(kind Kind, pkgPath string) Kind {
	if kind != "" {
		return kind
	}
	return KindFromPackage(pkgPath)
}

// Title 返回该类型报告的 Markdown 主标题。
func (k Kind) Title() string {
	switch k {
	case KindFlow:
		return "Flow 流程测试报告"
	case KindLoad:
		return "压测报告"
	default:
		return "API 接口测试报告"
	}
}

// ReportPrefix 文件名前缀，如 api-report。
func (k Kind) ReportPrefix() string {
	return string(k) + "-report"
}
