// Package flow 提供 Flow 流程测试的编排能力，包括步骤封装与跨步骤变量传递。
package flow

import "testing"

// Step 将流程中的一个步骤包装为子测试（t.Run），便于在报告中独立记录与统计。
func Step(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	t.Run(name, fn)
}
