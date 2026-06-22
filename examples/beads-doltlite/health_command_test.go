package beadsdoltlite_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDoltliteHealthScriptDoesNotForceDefaultShellTimeout(t *testing.T) {
	script := filepath.Join(repoRootForTest(t), "commands", "health", "run.sh")
	text := mustReadFile(t, script)
	if strings.Contains(string(text), "\"${GC_DOLTLITE_HEALTH_TIMEOUT:-15s}\"") {
		t.Fatalf("health run script still hardcodes a 15s timeout default")
	}
}

func TestDoltliteHealthJSONSchemaIsValidObject(t *testing.T) {
	schemaPath := filepath.Join(repoRootForTest(t), "commands", "health", "schemas", "result.schema.json")
	raw := mustReadFile(t, schemaPath)

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("result schema is not valid JSON: %v", err)
	}
	if parsed["type"] != "object" {
		t.Fatalf("result schema should declare type=object; got %v", parsed["type"])
	}
}

func TestDoltliteHealthScriptOutputsJSONOKWithoutJq(t *testing.T) {
	script := filepath.Join(repoRootForTest(t), "commands", "health", "run.sh")
	cityRoot := filepath.Clean(filepath.Join(repoRootForTest(t), "..", ".."))

	cmd := exec.Command("bash", script, "--json")
	cmd.Env = append(os.Environ(), "GC_CITY_PATH="+cityRoot)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("script failed: %v\noutput=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "\"ok\":true") && !strings.Contains(out.String(), "\"ok\": true") {
		t.Fatalf("script output must include ok=true: %s", out.String())
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	dir := filepath.Dir(filename)
	for {
		if filepath.Base(dir) == "beads-doltlite" {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate beads-doltlite root from %s", filename)
		}
		dir = parent
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}
	return raw
}
