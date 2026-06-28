package main

import (
	"os"

	"github.com/muyi-zcy/my-testgogogo/examples/demo/scenario"
	"github.com/muyi-zcy/my-testgogogo/loadkit"
)

func main() {
	os.Exit(loadkit.Main(scenario.Registry))
}
