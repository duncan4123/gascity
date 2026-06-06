//go:build cgo

package main

import (
	"github.com/gastownhall/gascity/internal/beads"
)

func openOptimizedDoltliteStore(storePath string, store *beads.BdStore) (beads.Store, bool) {
	direct, err := beads.NewDoltliteReadStore(storePath, store)
	if err == nil {
		return direct, true
	}
	return nil, false
}
