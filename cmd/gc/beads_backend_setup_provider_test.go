package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDoltliteBeadsBackendPluginCapabilities(t *testing.T) {
	t.Setenv("GC_BEADS", "")
	t.Setenv("GC_BEADS_BACKEND", "")
	cityPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(cityPath, "city.toml"), []byte(`[beads]
provider = "plugin"
backend = "doltlite"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	caps, ok := beadsBackendPluginCapabilitiesForCity(cityPath)
	if !ok {
		t.Fatal("beadsBackendPluginCapabilitiesForCity ok = false, want true")
	}
	if !caps.SetupHook || !caps.ProviderLifecycle || !caps.BackendPluginMetadata || !caps.GascityFastpathMetadata || !caps.NativeReadStore || !caps.StoreHealthPath {
		t.Fatalf("capabilities = %+v, want all DoltLite plugin integration points enabled", caps)
	}
	if caps.BDCompatibility != "bd-1.0.5" {
		t.Fatalf("BDCompatibility = %q, want bd-1.0.5", caps.BDCompatibility)
	}
}

func TestDoltliteBeadsBackendPluginStorePath(t *testing.T) {
	t.Setenv("GC_BEADS", "")
	t.Setenv("GC_BEADS_BACKEND", "")
	cityPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(cityPath, "city.toml"), []byte(`[beads]
provider = "plugin"
backend = "doltlite"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, ok := beadsBackendPluginStorePath(cityPath)
	if !ok {
		t.Fatal("beadsBackendPluginStorePath ok = false, want true")
	}
	want := filepath.Join(cityPath, ".beads", "doltlite")
	if got != want {
		t.Fatalf("beadsBackendPluginStorePath = %q, want %q", got, want)
	}
}

func TestPackDeclaredBeadsBackendPluginWinsOverFallback(t *testing.T) {
	t.Setenv("GC_BEADS", "")
	t.Setenv("GC_BEADS_BACKEND", "")
	cityPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(cityPath, "city.toml"), []byte(`[beads]
provider = "plugin"
backend = "doltlite"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cityPath, "assets", "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	setupHook := filepath.Join(cityPath, "assets", "scripts", "setup-doltlite.sh")
	if err := os.WriteFile(setupHook, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cityPath, "pack.toml"), []byte(`
[pack]
name = "city"
schema = 2

[[backend_plugins]]
backend = "doltlite"
setup_hook = "assets/scripts/setup-doltlite.sh"
store_path = ".beads/doltlite"
bd_compatibility = "bd-test"
capabilities = ["setup", "provider", "metadata", "fastpath", "store-health"]

[backend_plugins.beads_endpoint]
command = ".gc/runtime/packs/bd-gc-dl/bin/bd-backend-doltlite"
args = ["serve"]
protocol = "beads.backend.v1alpha1"

[backend_plugins.gascity_endpoint]
command = ".gc/runtime/packs/bd-gc-dl/bin/gc-doltlite-fastpath"
args = ["serve"]
protocol = "gascity.backend.v1alpha1"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	gotHook, ok := beadsProviderSetupHook(cityPath)
	if !ok {
		t.Fatal("beadsProviderSetupHook ok = false, want true")
	}
	if gotHook != setupHook {
		t.Fatalf("beadsProviderSetupHook = %q, want pack-declared %q", gotHook, setupHook)
	}

	gotStore, ok := beadsBackendPluginStorePath(cityPath)
	if !ok {
		t.Fatal("beadsBackendPluginStorePath ok = false, want true")
	}
	wantStore := filepath.Join(cityPath, ".beads", "doltlite")
	if gotStore != wantStore {
		t.Fatalf("beadsBackendPluginStorePath = %q, want %q", gotStore, wantStore)
	}

	caps, ok := beadsBackendPluginCapabilitiesForCity(cityPath)
	if !ok {
		t.Fatal("beadsBackendPluginCapabilitiesForCity ok = false, want true")
	}
	if caps.BDCompatibility != "bd-test" {
		t.Fatalf("BDCompatibility = %q, want bd-test", caps.BDCompatibility)
	}
	if !caps.SetupHook || !caps.ProviderLifecycle || !caps.BackendPluginMetadata || !caps.GascityFastpathMetadata || !caps.NativeReadStore || !caps.StoreHealthPath {
		t.Fatalf("capabilities = %+v, want pack-declared plugin integration points enabled", caps)
	}
}
