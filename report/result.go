// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// normalizeResult 将任意类型的结构化结果统一转换为 map[string]any。
// 支持 map、reflect.Map、struct（经 JSON 序列化）及标量值。
func normalizeResult(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}

	if m, ok := value.(map[string]any); ok {
		return cloneMap(m), nil
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map {
		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			result[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
		}
		return result, nil
	}

	// struct 等类型通过 JSON 往返转换
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]any{"value": fmt.Sprint(value)}, nil
	}

	var object map[string]any
	if err := json.Unmarshal(data, &object); err == nil && len(object) > 0 {
		return object, nil
	}

	var scalar any
	if err := json.Unmarshal(data, &scalar); err == nil {
		return map[string]any{"value": scalar}, nil
	}

	return map[string]any{"value": fmt.Sprint(value)}, nil
}

// cloneMap 浅拷贝 map，避免步骤间共享引用。
func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// formatRecordValue 格式化单条 Record 的值，标量包装为 {"value": ...} 后取 value 字段。
func formatRecordValue(value any) any {
	normalized, err := normalizeResult(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	if len(normalized) == 1 {
		if v, ok := normalized["value"]; ok {
			return v
		}
	}
	return normalized
}
