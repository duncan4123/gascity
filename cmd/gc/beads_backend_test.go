package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gastownhall/gascity/internal/beads/contract"
	"github.com/gastownhall/gascity/internal/fsys"
)

func TestScopeNeedsDoltDoctorChecksFollowsBackendCapability(t *testing.T) {
	clearInheritedBeadsEnv(t)

	doltCity := writeBackendCity(t, "dolt")
	if !scopeNeedsDoltDoctorChecks(doltCity, doltCity) {
		t.Fatal("dolt city should need built-in Dolt doctor checks")
	}

	doltliteCity := writeBackendCity(t, "doltlite")
	if scopeNeedsDoltDoctorChecks(doltliteCity, doltliteCity) {
		t.Fatal("doltlite city should use pack-managed doctor checks")
	}

	inheritedRig := filepath.Join(doltliteCity, "rigs", "dog")
	if err := os.MkdirAll(filepath.Join(inheritedRig, ".beads"), 0o755); err != nil {
		t.Fatalf("mkdir inherited rig: %v", err)
	}
	if scopeNeedsDoltDoctorChecks(doltliteCity, inheritedRig) {
		t.Fatal("rig without backend override should inherit doltlite doctor behavior")
	}

	explicitRig := filepath.Join(doltliteCity, "rigs", "ops")
	if err := os.MkdirAll(filepath.Join(explicitRig, ".beads"), 0o755); err != nil {
		t.Fatalf("mkdir explicit rig: %v", err)
	}
	if _, err := contract.EnsureCanonicalConfig(fsys.OSFS{}, filepath.Join(explicitRig, ".beads", "config.yaml"), contract.ConfigState{
		IssuePrefix:    "ops",
		EndpointOrigin: contract.EndpointOriginExplicit,
		EndpointStatus: contract.EndpointStatusVerified,
		DoltHost:       "db.example.test",
		DoltPort:       "3306",
		DoltUser:       "bd",
	}); err != nil {
		t.Fatalf("write explicit rig config: %v", err)
	}
	if !scopeNeedsDoltDoctorChecks(doltliteCity, explicitRig) {
		t.Fatal("explicit Dolt rig under doltlite city should need built-in Dolt doctor checks")
	}
}

func writeBackendCity(t *testing.T, backend string) string {
	t.Helper()

	cityDir := t.TempDir()
	data := []byte("[workspace]\nname = \"demo\"\n\n[beads]\nprovider = \"bd\"\nbackend = \"" + backend + "\"\n")
	if err := os.WriteFile(filepath.Join(cityDir, "city.toml"), data, 0o644); err != nil {
		t.Fatalf("write city.toml: %v", err)
	}
	return cityDir
}
