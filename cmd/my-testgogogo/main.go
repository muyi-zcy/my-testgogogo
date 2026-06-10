// Command my-testgogogo 是测试框架的 CLI 工具，当前支持 report 子命令。
//
// report 子命令从 stdin 读取 go test -json 输出，合并 staging 中的 Fragment，
// 生成 Markdown 格式的批次测试报告。
//
// 用法：
//
//	go test -json ./... | my-testgogogo report --run-id <id> --command "go test ./..."
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/muyi-zcy/my-testgogogo/report"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "report":
		runReport(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

// runReport 执行 report 子命令：解析 go test JSON、加载 Fragment、生成 Markdown。
func runReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	runID := fs.String("run-id", "", "report run id, format YYYYMMDD-HHMMSS")
	command := fs.String("command", "", "command shown in report")
	_ = fs.Parse(args)

	cfg, err := report.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load report config: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()
	if *runID == "" {
		*runID = report.NewRunID(now)
	}

	// 从 stdin 读取 go test -json 的逐行输出
	lines := readLines(os.Stdin)
	events, duration, err := report.ParseGoTestJSONLines(lines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse go test json: %v\n", err)
		os.Exit(1)
	}

	// 加载同 runID 下各用例采集的 Fragment
	fragments, err := report.LoadFragments(cfg, *runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load report fragments: %v\n", err)
		os.Exit(1)
	}

	summary := report.Summary{
		RunID:     *runID,
		Generated: now,
		GoVersion: runtime.Version(),
		Command:   *command,
		Duration:  duration,
		Events:    events,
		Fragments: fragments,
	}

	path, latest, err := report.WriteBatchMarkdown(cfg, summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write markdown report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("report: %s\n", path)
	fmt.Printf("latest: %s\n", latest)
}

// readLines 从 Reader 逐行读取，跳过空行；缓冲区上限 1MB 以容纳大段输出。
func readLines(r *os.File) []string {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  my-testgogogo report merge --run-id <id> [--command <cmd>]\n")
}
