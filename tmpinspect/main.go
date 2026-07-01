package main

import (
	"fmt"

	"github.com/gastownhall/gascity/internal/config"
)

func main() {
	cfg, err := config.LoadCityConfig("/data/projects/doltlite-gascity", nil)
	if err != nil {
		panic(err)
	}
	for _, c := range cfg.PackCommands {
		if c.BindingName == "beads-doltlite" {
			fmt.Printf("cmd=%q\nsource=%q\npack=%q\npackdir=%q\nrun=%q\n", c.Command, c.SourceDir, c.PackName, c.PackDir, c.RunScript)
		}
	}
}
