// Package flow 提供 Flow 流程测试的编排能力，包括步骤封装与跨步骤变量传递。
package flow

import "testing"

// Step 将流程中的一个步骤包装为子测试（t.Run），便于在报告中独立记录与统计。
// 每个 Step 拥有独立的 *testing.T，失败时只标记该步骤失败，不影响已通过的步骤记录。
func Step(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	// 底层调用 t.Run：Go 原生子测试机制，支持独立命名、并行控制与失败隔离
	t.Run(name, fn)
}
