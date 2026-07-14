package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gastownhall/gascity/internal/config"
	"github.com/gastownhall/gascity/internal/doctor"
	"github.com/gastownhall/gascity/internal/supervisor"
)

func TestStandaloneOwnershipPolicyError(t *testing.T) {
	cityPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(cityPath, "city.toml"), []byte("[supervision]\nrequired = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := standaloneOwnershipPolicyError(cityPath); err == nil || !strings.Contains(err.Error(), "machine-supervisor ownership") {
		t.Fatalf("standaloneOwnershipPolicyError() = %v, want supervisor requirement", err)
	}
	if err := os.WriteFile(filepath.Join(cityPath, "city.toml"), []byte("[workspace]\nname = \"legacy\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := standaloneOwnershipPolicyError(cityPath); err != nil {
		t.Fatalf("standaloneOwnershipPolicyError legacy = %v, want nil", err)
	}
}

func TestCitySupervisorRegistryDiagnosticsReportsIsolatedHome(t *testing.T) {
	cityPath := t.TempDir()
	activeHome := filepath.Join(t.TempDir(), "isolated")
	defaultHome := filepath.Join(t.TempDir(), "shared")
	if err := supervisor.NewRegistry(filepath.Join(defaultHome, "cities.toml")).Register(cityPath, "city"); err != nil {
		t.Fatal(err)
	}
	diagnostics := citySupervisorRegistryDiagnosticsForHomes(cityPath, activeHome, defaultHome)
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "unset GC_HOME") {
		t.Fatalf("diagnostics = %v, want isolated GC_HOME warning", diagnostics)
	}
}

func TestCitySupervisorRegistryDiagnosticsReportsLegacyCityHome(t *testing.T) {
	cityPath := t.TempDir()
	defaultHome := filepath.Join(t.TempDir(), "shared")
	legacyHome := filepath.Join(cityPath, ".gc-home")
	if err := supervisor.NewRegistry(filepath.Join(defaultHome, "cities.toml")).Register(cityPath, "city"); err != nil {
		t.Fatal(err)
	}
	if err := supervisor.NewRegistry(filepath.Join(legacyHome, "cities.toml")).Register(cityPath, "city"); err != nil {
		t.Fatal(err)
	}
	diagnostics := citySupervisorRegistryDiagnosticsForHomes(cityPath, defaultHome, defaultHome)
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "legacy city-local registry") {
		t.Fatalf("diagnostics = %v, want legacy registry warning", diagnostics)
	}
}

func TestSupervisionRegistryDoctorCheckRequiresRegistration(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GC_HOME", home)
	cityPath := t.TempDir()
	check := newSupervisionRegistryDoctorCheck(cityPath, &config.City{Supervision: config.SupervisionConfig{Required: true}})
	result := check.Run(&doctor.CheckContext{CityPath: cityPath})
	if result.Status != doctor.StatusError || !strings.Contains(result.Message, "not registered") {
		t.Fatalf("unregistered result = %#v, want registration error", result)
	}
	if err := supervisor.NewRegistry(filepath.Join(home, "cities.toml")).Register(cityPath, "city"); err != nil {
		t.Fatal(err)
	}
	result = check.Run(&doctor.CheckContext{CityPath: cityPath})
	if result.Status != doctor.StatusOK {
		t.Fatalf("registered result = %#v, want OK", result)
	}
}
