package main

import (
    "fmt"
    "github.com/gastownhall/gascity/internal/config"
    "github.com/gastownhall/gascity/internal/fsys"
)

func main() {
    cmds, err := config.DiscoverPackCommands(fsys.OSFS{}, "/data/projects/doltlite-gascity/gascity/examples/beads-doltlite", "beads-doltlite")
    if err != nil {
        panic(err)
    }
    for _, c := range cmds {
        if len(c.Command) > 0 && c.Command[0] == "health" {
            fmt.Printf("command=%q source=%q\n", c.Command, c.SourceDir)
        }
    }
}
