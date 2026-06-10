// Package flow 提供 Flow 流程测试的编排能力，包括步骤封装与跨步骤变量传递。
package flow

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muyi-zcy/my-testgogogo/path"
	"gopkg.in/yaml.v3"
)

// Seed 对应 configs/testdata/flow.yaml 中的 Flow 测试种子数据。
type Seed struct {
	PageSize int `yaml:"page_size"` // 默认分页大小
}

// DefaultSeed 加载 Flow 测试种子数据并转为 Vars 可用的 map。
// 加载失败时回退到 pageSize=10。
func DefaultSeed() map[string]any {
	seed, err := LoadSeed()
	if err != nil {
		return map[string]any{"pageSize": 10}
	}
	return map[string]any{"pageSize": seed.PageSize}
}

// LoadSeed 从 configs/testdata/flow.yaml 读取 Flow 测试种子配置。
func LoadSeed() (*Seed, error) {
	root, err := path.ModuleRoot()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(root, "configs", "testdata", "flow.yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read flow testdata: %w", err)
	}

	var seed Seed
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return nil, fmt.Errorf("parse flow testdata: %w", err)
	}
	if seed.PageSize <= 0 {
		seed.PageSize = 10
	}
	return &seed, nil
}
