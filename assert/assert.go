// Package assert 提供 API 集成测试中常用的断言辅助函数。
//
// 基于 testify/assert 和 testify/require，封装 HTTP 状态码、JSON 字段及分页结果断言。
package assert

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/muyi-zcy/my-testgogogo/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPStatus 断言 HTTP 响应状态码等于期望值，失败时附带响应体便于排查。
func HTTPStatus(t *testing.T, resp *client.Response, expected int) {
	t.Helper()
	require.NotNil(t, resp)
	assert.Equal(t, expected, resp.StatusCode, "body=%s", string(resp.Body))
}

// JSONFieldEqual 断言 JSON 响应体中指定点分路径的字段值等于期望值。
// path 示例："code" 或 "data.accessToken"。
func JSONFieldEqual(t *testing.T, body []byte, path string, expected any) {
	t.Helper()
	value, err := jsonPathValue(body, path)
	require.NoError(t, err)
	assert.Equal(t, expected, value)
}

// JSONPathExists 断言 JSON 响应体中存在指定点分路径的字段。
func JSONPathExists(t *testing.T, body []byte, path string) {
	t.Helper()
	_, err := jsonPathValue(body, path)
	require.NoError(t, err)
}

// PageNotEmpty 断言分页查询结果合法：total 非负且 records 非 nil。
func PageNotEmpty[T any](t *testing.T, total int64, records []T) {
	t.Helper()
	assert.GreaterOrEqual(t, total, int64(0))
	assert.NotNil(t, records)
}

// jsonPathValue 从 JSON 字节中按点分路径提取字段值。
func jsonPathValue(body []byte, path string) (any, error) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	current := data
	for _, part := range strings.Split(path, ".") {
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("path %q: not an object at %q", path, part)
		}
		val, ok := obj[part]
		if !ok {
			return nil, fmt.Errorf("path %q: key %q not found", path, part)
		}
		current = val
	}
	return current, nil
}
