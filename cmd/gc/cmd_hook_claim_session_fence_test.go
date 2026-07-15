package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/session"
)

func TestHookCommandClaimRejectsFailedSessionBeforeWorkQuery(t *testing.T) {
	clearGCEnv(t)
	disableManagedDoltRecoveryForTest(t)
	t.Setenv("GC_BEADS", "file")
	cityDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cityDir, ".gc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cityDir, "city.toml"), []byte(`[workspace]
name = "test-city"

[[agent]]
name = "worker"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := openCityStoreAt(cityDir)
	if err != nil {
		t.Fatalf("openCityStoreAt: %v", err)
	}
	sessionBead, err := store.Create(beads.Bead{
		Title:  "worker-1",
		Type:   session.BeadType,
		Labels: []string{"gc:session", "agent:worker-1"},
		Metadata: map[string]string{
			"session_name":   "worker-1",
			"template":       "worker",
			"state":          string(session.StateFailedCreate),
			"instance_token": "failed-token",
		},
	})
	if err != nil {
		t.Fatalf("create session bead: %v", err)
	}

	fakeBin := t.TempDir()
	queryMarker := filepath.Join(t.TempDir(), "query-ran")
	fakeBD := filepath.Join(fakeBin, "bd")
	if err := os.WriteFile(fakeBD, []byte("#!/bin/sh\ntouch \"$QUERY_MARKER\"\nprintf '[]'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("QUERY_MARKER", queryMarker)
	t.Setenv("GC_CITY", cityDir)
	t.Setenv("GC_TEMPLATE", "worker")
	t.Setenv("GC_ALIAS", "worker-1")
	t.Setenv("GC_SESSION_ID", sessionBead.ID)
	t.Setenv("GC_SESSION_NAME", "worker-1")
	t.Setenv("GC_SESSION_ORIGIN", "ephemeral")
	t.Setenv("GC_INSTANCE_TOKEN", "failed-token")

	var stdout, stderr bytes.Buffer
	code := cmdHookWithOptions(nil, hookCommandOptions{Claim: true, JSON: true}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("cmdHookWithOptions(failed session) = %d, want 1; stdout=%q stderr=%s", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "refusing stale session") ||
		!strings.Contains(stderr.String(), "failed-create") {
		t.Fatalf("stderr = %q, want failed-session refusal", stderr.String())
	}
	if _, err := os.Stat(queryMarker); !os.IsNotExist(err) {
		t.Fatalf("work query ran for failed session; stat error = %v", err)
	}
}

func TestHookClaimSessionEligibleRejectsFailedCreateRuntime(t *testing.T) {
	info := session.Info{
		ID:            "sqlite-gc-vdq6",
		MetadataState: string(session.StateFailedCreate),
		InstanceToken: "failed-token",
	}

	err := hookClaimSessionEligible(info, "failed-token")
	if err == nil || !strings.Contains(err.Error(), "failed-create") {
		t.Fatalf("hookClaimSessionEligible(failed-create) error = %v, want state rejection", err)
	}
}

func TestHookClaimSessionEligibleRejectsClosedRuntime(t *testing.T) {
	info := session.Info{
		ID:            "sqlite-gc-vdq6",
		Closed:        true,
		InstanceToken: "failed-token",
	}

	err := hookClaimSessionEligible(info, "failed-token")
	if err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("hookClaimSessionEligible(closed) error = %v, want closed rejection", err)
	}
}

func TestHookClaimSessionEligibleRejectsSupersededRuntimeToken(t *testing.T) {
	info := session.Info{
		ID:            "sqlite-gc-worker",
		MetadataState: string(session.StateActive),
		InstanceToken: "replacement-token",
	}

	err := hookClaimSessionEligible(info, "stale-runtime-token")
	if err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("hookClaimSessionEligible(token mismatch) error = %v, want token rejection", err)
	}
}

func TestHookClaimSessionEligibleAllowsCurrentLiveRuntime(t *testing.T) {
	for _, state := range []session.State{session.StateActive, session.StateAwake} {
		t.Run(string(state), func(t *testing.T) {
			info := session.Info{
				ID:            "sqlite-gc-worker",
				MetadataState: string(state),
				InstanceToken: "current-token",
			}
			if err := hookClaimSessionEligible(info, "current-token"); err != nil {
				t.Fatalf("hookClaimSessionEligible(%s) error = %v, want nil", state, err)
			}
		})
	}
}
