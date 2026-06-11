package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/muyi-zcy/my-testgogogo/path"
)

// runLoad 若项目存在 cmd/load/main.go 则委托执行，否则提示创建入口。
func runLoad(args []string) {
	root, err := path.ModuleRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "module root: %v\n", err)
		os.Exit(1)
	}

	loadMain := filepath.Join(root, "cmd", "load", "main.go")
	if _, err := os.Stat(loadMain); err != nil {
		fmt.Fprintf(os.Stderr, "load entrypoint not found: %s\n", loadMain)
		fmt.Fprintf(os.Stderr, "create cmd/load/main.go that calls loadkit.Main(scenario.Registry)\n")
		os.Exit(1)
	}

	cmdArgs := append([]string{"run", "./cmd/load"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
