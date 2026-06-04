package main

import (
	"os"
	"strings"
)

func gcDoltSkip() bool {
	if strings.TrimSpace(os.Getenv("GC_DOLT")) == "skip" {
		return true
	}
	// doltlite backend has no dolt server, no managed dolt, no dolt ops.
	return isDoltliteCity()
}
