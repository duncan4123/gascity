//go:build gascity_doltlite_lib

package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/gastownhall/gascity/internal/beads"
)

const doltliteReadFastPathEnv = "GC_DOLTLITE_READ_FASTPATH"

func openOptimizedDoltliteStore(storePath, cityPath string, store *beads.BdStore) (beads.Store, bool) {
	if !doltliteReadFastPathEnabled(cityPath, storePath) {
		return nil, false
	}
	direct, err := beads.NewDoltliteReadStore(storePath, store)
	if err == nil {
		return direct, true
	}
	return nil, false
}

func doltliteReadFastPathEnabled(cityPath, storePath string) bool {
	raw := strings.TrimSpace(os.Getenv(doltliteReadFastPathEnv))
	if raw != "" {
		enabled, err := strconv.ParseBool(raw)
		return err == nil && enabled
	}
	return scopeBackendIsDoltlite(cityPath, resolveStoreScopeRoot(cityPath, storePath))
}
