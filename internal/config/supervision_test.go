package config

import (
	"strings"
	"testing"
)

func TestSupervisionConfigRequiredRoundTrip(t *testing.T) {
	cfg, err := Parse([]byte("[supervision]\nrequired = true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Supervision.RequiresSupervisor() {
		t.Fatal("Supervision.RequiresSupervisor() = false, want true")
	}
	data, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[supervision]\nrequired = true") {
		t.Fatalf("Marshal() = %q, want supervision requirement", data)
	}
}

func TestSupervisionConfigLegacyDefaultAllowsStandalone(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"legacy\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Supervision.RequiresSupervisor() {
		t.Fatal("legacy config unexpectedly requires supervisor")
	}
}
