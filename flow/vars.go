// Package flow 提供 Flow 流程测试的编排能力，包括步骤封装与跨步骤变量传递。
package flow

import (
	"fmt"
)

// Vars 是 Flow 测试中的跨步骤变量容器，前序步骤写入、后续步骤读取。
// 典型用法：Step 1 写入 branch/firstItemCode，Step 2 或 switch 分支读取并决定后续路径。
type Vars struct {
	data map[string]any // 键值存储，key 为变量名，value 为任意类型
}

// NewVars 用初始种子数据创建 Vars 实例。
// seed 通常来自 DefaultSeed()，可预置 pageSize 等全局默认值。
func NewVars(seed map[string]any) *Vars {
	// 拷贝 seed，避免外部修改影响 Vars 内部状态
	data := make(map[string]any, len(seed))
	for key, value := range seed {
		data[key] = value
	}
	return &Vars{data: data}
}

// Set 写入变量。同名 key 会覆盖旧值，供后续 Step 或 switch 分支读取。
func (v *Vars) Set(key string, val any) {
	v.data[key] = val
}

// Get 读取变量，不存在时返回 nil。
// 常用于 switch 分支判断，如 switch vars.Get("branch") { case "has_data": ... }
func (v *Vars) Get(key string) any {
	return v.data[key]
}

// MustString 读取字符串变量，不存在、类型错误或空字符串时 panic。
// 适用于必须存在且非空的标识字段，如 firstItemCode、firstISBN。
func (v *Vars) MustString(key string) string {
	val, ok := v.data[key]
	if !ok {
		panic(fmt.Sprintf("flow var %q not found", key))
	}
	s, ok := val.(string)
	if !ok || s == "" {
		panic(fmt.Sprintf("flow var %q is not a non-empty string", key))
	}
	return s
}

// MustInt 读取整型变量，支持 int、int64、float64 类型转换，否则 panic。
// 适用于计数、分页大小等数值字段，如 pageSize、roleCount。
func (v *Vars) MustInt(key string) int {
	val, ok := v.data[key]
	if !ok {
		panic(fmt.Sprintf("flow var %q not found", key))
	}
	switch n := val.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		panic(fmt.Sprintf("flow var %q is not an int", key))
	}
}

// Has 判断变量是否存在。适用于可选分支或条件写入前的探测。
func (v *Vars) Has(key string) bool {
	_, ok := v.data[key]
	return ok
}
