//go:build gascity_doltlite_lib

package main

import "testing"

func TestDoltliteReadFastPathEnabled(t *testing.T) {
	tests := []struct {
		name    string
		native  string
		backend string
		want    bool
	}{
		{name: "unset with doltlite backend", backend: "doltlite", want: true},
		{name: "unset with dolt backend", backend: "dolt", want: false},
		{name: "true overrides dolt backend", native: "true", backend: "dolt", want: true},
		{name: "spaced true overrides dolt backend", native: " true ", backend: "dolt", want: true},
		{name: "false overrides doltlite backend", native: "false", backend: "doltlite", want: false},
		{name: "invalid disables doltlite backend", native: "enabled", backend: "doltlite", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cityPath := t.TempDir()
			t.Setenv(doltliteReadFastPathEnv, tt.native)
			t.Setenv("GC_BEADS_BACKEND", tt.backend)
			if got := doltliteReadFastPathEnabled(cityPath, cityPath); got != tt.want {
				t.Fatalf("doltliteReadFastPathEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
