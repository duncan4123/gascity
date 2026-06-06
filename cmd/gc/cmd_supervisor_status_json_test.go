package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSupervisorStatusJSON(t *testing.T) {
	clearGCEnv(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"supervisor", "status", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(supervisor status --json) = %d, stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var payload struct {
		SchemaVersion string   `json:"schema_version"`
		Running       bool     `json:"running"`
		PID           int      `json:"pid"`
		SocketPath    string   `json:"socket_path"`
		CheckedPaths  []string `json:"checked_paths"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload.SchemaVersion != "1" {
		t.Fatalf("schema_version = %q, want 1", payload.SchemaVersion)
	}
	if len(payload.CheckedPaths) == 0 {
		t.Fatalf("checked_paths empty: %+v", payload)
	}
	if !payload.Running && payload.PID != 0 {
		t.Fatalf("not running with pid = %d", payload.PID)
	}
}

func TestSupervisorStatusLineIncludesBinary(t *testing.T) {
	if got, want := supervisorStatusLine(4242, "/home/ubuntu/.local/bin/gc", nil), "Supervisor is running (PID 4242, binary /home/ubuntu/.local/bin/gc)"; got != want {
		t.Fatalf("supervisorStatusLine() = %q, want %q", got, want)
	}
}

func TestSupervisorStatusLineIncludesBuildCheck(t *testing.T) {
	build := &BinaryBuildJSON{Status: "matched"}
	if got, want := supervisorStatusLine(4242, "/home/ubuntu/.local/bin/gc", build), "Supervisor is running (PID 4242, binary /home/ubuntu/.local/bin/gc, build matches beads-doltlite stamp)"; got != want {
		t.Fatalf("supervisorStatusLine() = %q, want %q", got, want)
	}
}

func TestSupervisorStatusPayloadIncludesBinary(t *testing.T) {
	build := &BinaryBuildJSON{Status: "matched"}
	payload := supervisorStatusPayload("/tmp/gc.sock", 4242, "/usr/local/bin/gc", build)
	if got, ok := payload["binary"].(string); !ok || got != "/usr/local/bin/gc" {
		t.Fatalf("payload binary = %#v, want /usr/local/bin/gc", payload["binary"])
	}
	if got, ok := payload["build"].(*BinaryBuildJSON); !ok || got != build {
		t.Fatalf("payload build = %#v, want build pointer", payload["build"])
	}
	if got, ok := payload["running"].(bool); !ok || !got {
		t.Fatalf("payload running = %#v, want true", payload["running"])
	}
	if got, ok := payload["pid"].(int); !ok || got != 4242 {
		t.Fatalf("payload pid = %#v, want 4242", payload["pid"])
	}
}
