// Package path 提供项目路径解析工具，用于定位 go.mod 所在的项目根目录。
package path

import (
	"fmt"
	"os"
	"path/filepath"
)

// ModuleRoot 从当前工作目录向上逐级查找，返回包含 go.mod 的目录路径。
// 测试运行时工作目录通常为 examples/<name>，因此需要向上回溯。
func ModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
