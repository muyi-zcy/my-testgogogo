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
	PageSize int `yaml:"page_size"` // 默认分页大小，注入 vars 后供 ListParams 使用
}

// DefaultSeed 加载 Flow 测试种子数据并转为 Vars 可用的 map。
// 加载失败时回退到 pageSize=10，保证 Flow 测试在无配置文件时仍可运行。
func DefaultSeed() map[string]any {
	seed, err := LoadSeed()
	if err != nil {
		// 降级默认值：与 LoadSeed 内部 PageSize<=0 时的兜底一致
		return map[string]any{"pageSize": 10}
	}
	// 转为 map 供 flow.NewVars 消费，key 与测试代码中 vars.MustInt("pageSize") 对应
	return map[string]any{"pageSize": seed.PageSize}
}

// LoadSeed 从 configs/testdata/flow.yaml 读取 Flow 测试种子配置。
func LoadSeed() (*Seed, error) {
	// 定位模块根目录，确保无论从哪个包运行测试都能找到配置文件
	root, err := path.ModuleRoot()
	if err != nil {
		return nil, err
	}

	// 固定路径：模块根/configs/testdata/flow.yaml
	filePath := filepath.Join(root, "configs", "testdata", "flow.yaml")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read flow testdata: %w", err)
	}

	var seed Seed
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return nil, fmt.Errorf("parse flow testdata: %w", err)
	}
	// 非法或缺失时分页大小兜底为 10
	if seed.PageSize <= 0 {
		seed.PageSize = 10
	}
	return &seed, nil
}
