// Package jsonutil 提供兼容 WMS 等后端 JSON 序列化差异的工具类型。
package jsonutil

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt 兼容 JSON 中 number 或 string 形式的整数值。
type FlexInt int

// Int 返回 int 值。
func (v FlexInt) Int() int { return int(v) }

// Int64 返回 int64 值。
func (v FlexInt) Int64() int64 { return int64(v) }

// UnmarshalJSON 解析 JSON 整数值（支持 number、string、null）。
func (v *FlexInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = 0
		return nil
	}

	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*v = FlexInt(n)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("flex int: %w", err)
	}
	if s == "" {
		*v = 0
		return nil
	}

	parsed, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("flex int: parse %q: %w", s, err)
	}
	*v = FlexInt(parsed)
	return nil
}

// FlexString 兼容 JSON 中 string 或 number 形式的字符串值。
type FlexString string

// String 返回字符串值。
func (v FlexString) String() string { return string(v) }

// UnmarshalJSON 解析 JSON 字符串值（支持 string、number、null）。
func (v *FlexString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*v = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*v = FlexString(s)
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("flex string: %w", err)
	}
	*v = FlexString(n.String())
	return nil
}
