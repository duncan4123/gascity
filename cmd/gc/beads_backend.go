package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gastownhall/gascity/internal/fsys"
)

// BeadsBackend abstracts bead storage backend behavior so callers
// dispatch on backend identity in one place instead of branching on
// cityUsesDoltliteBeadsBackend() across the codebase.
type BeadsBackend interface {
	Name() string
	NeedsManagedServer() bool
	NeedsDoltBinary() bool
	MinBDVersion() string
	NeedsBeadHooks() bool
	MetadataInit(fs fsys.FS, scopeRoot, doltDatabase string, preserveExisting bool) error
	MetadataEnforce(fs fsys.FS, scopeRoot, doltDatabase string) error
	ProviderEnv() []string
	RequiredBuiltinPacks() []string
}

type doltBackend struct{}

func (d *doltBackend) Name() string                   { return "dolt" }
func (d *doltBackend) NeedsManagedServer() bool       { return true }
func (d *doltBackend) NeedsDoltBinary() bool          { return true }
func (d *doltBackend) MinBDVersion() string           { return "1.0.4" }
func (d *doltBackend) NeedsBeadHooks() bool           { return true }
func (d *doltBackend) RequiredBuiltinPacks() []string { return []string{"dolt"} }

func (d *doltBackend) MetadataInit(fs fsys.FS, scopeRoot, doltDatabase string, preserveExisting bool) error {
	return ensureCanonicalScopeMetadata(fs, scopeRoot, doltDatabase, preserveExisting)
}

func (d *doltBackend) MetadataEnforce(fs fsys.FS, scopeRoot, doltDatabase string) error {
	return enforceCanonicalScopeMetadataForInit(fs, scopeRoot, doltDatabase)
}

func (d *doltBackend) ProviderEnv() []string { return nil }

type doltliteBackend struct{}

func (dl *doltliteBackend) Name() string                   { return "doltlite" }
func (dl *doltliteBackend) NeedsManagedServer() bool       { return false }
func (dl *doltliteBackend) NeedsDoltBinary() bool          { return false }
func (dl *doltliteBackend) MinBDVersion() string           { return "1.0.3" }
func (dl *doltliteBackend) NeedsBeadHooks() bool           { return false }
func (dl *doltliteBackend) RequiredBuiltinPacks() []string { return []string{"beads-doltlite"} }

func (dl *doltliteBackend) MetadataInit(fs fsys.FS, scopeRoot, doltDatabase string, preserveExisting bool) error {
	return ensureCanonicalDoltliteScopeMetadata(fs, scopeRoot, doltDatabase, preserveExisting)
}

func (dl *doltliteBackend) MetadataEnforce(fs fsys.FS, scopeRoot, doltDatabase string) error {
	return enforceCanonicalDoltliteScopeMetadataForInit(fs, scopeRoot, doltDatabase)
}

func (dl *doltliteBackend) ProviderEnv() []string {
	return []string{"GC_BEADS_BACKEND=doltlite", "BEADS_BACKEND=doltlite"}
}

// resolveBeadsBackend returns the active backend for a city path.
func resolveBeadsBackend(cityPath string) BeadsBackend {
	backend := strings.ToLower(resolveBeadsBackendString(cityPath))
	if backend == "doltlite" {
		return &doltliteBackend{}
	}
	return &doltBackend{}
}

func resolveBeadsBackendString(cityPath string) string {
	if v := strings.TrimSpace(os.Getenv("GC_BEADS_BACKEND")); v != "" {
		return v
	}
	return strings.TrimSpace(peekBeadsBackend(filepath.Join(cityPath, "city.toml")))
}
