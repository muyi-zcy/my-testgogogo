// Package report 提供测试报告采集、片段暂存与 Markdown 报告生成能力。
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LoadFragments 从指定类型的 staging 目录加载 runID 下的 Fragment。
func LoadFragments(cfg *Config, runID string, kind Kind) ([]Fragment, error) {
	dir := filepath.Join(cfg.StagingDir(kind), runID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	fragments := make([]Fragment, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var fragment Fragment
		if err := json.Unmarshal(data, &fragment); err != nil {
			return nil, fmt.Errorf("parse fragment %s: %w", entry.Name(), err)
		}
		if fragment.Kind == "" {
			fragment.Kind = ResolveKind("", fragment.Package)
		}
		fragments = append(fragments, fragment)
	}

	sort.Slice(fragments, func(i, j int) bool {
		if fragments[i].Package == fragments[j].Package {
			return fragments[i].TestName < fragments[j].TestName
		}
		return fragments[i].Package < fragments[j].Package
	})
	return fragments, nil
}

// ParseGoTestJSONLines 解析 go test -json 输出的逐行 JSON，提取事件列表与总耗时。
func ParseGoTestJSONLines(lines []string) ([]GoTestEvent, time.Duration, error) {
	events := make([]GoTestEvent, 0, len(lines))
	var totalElapsed float64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw struct {
			Time    string  `json:"Time"`
			Package string  `json:"Package"`
			Test    string  `json:"Test"`
			Action  string  `json:"Action"`
			Elapsed float64 `json:"Elapsed"`
			Output  string  `json:"Output"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue // 跳过非 JSON 行
		}
		parsedTime, _ := parseTime(raw.Time)
		events = append(events, GoTestEvent{
			Time:    parsedTime,
			Package: raw.Package,
			Test:    raw.Test,
			Action:  raw.Action,
			Elapsed: raw.Elapsed,
			Output:  strings.TrimSpace(raw.Output),
		})
		// 包级 pass 事件的 Elapsed 累加为总耗时
		if raw.Action == "pass" && raw.Test == "" {
			totalElapsed += raw.Elapsed
		}
	}

	return events, durationFromSeconds(totalElapsed), nil
}

// parseTime 尝试多种时间格式解析 go test -json 的 Time 字段。
func parseTime(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time: %s", value)
}

// durationFromSeconds 将秒数转换为 time.Duration。
func durationFromSeconds(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
