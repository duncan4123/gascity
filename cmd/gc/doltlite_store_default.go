//go:build !cgo

package main

import "github.com/gastownhall/gascity/internal/beads"

func openOptimizedDoltliteStore(_ string, _ *beads.BdStore) (beads.Store, bool) {
	return nil, false
}
