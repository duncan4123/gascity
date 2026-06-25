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

func TestDoltliteBuildScriptBuildsGCWithNativeReadTag(t *testing.T) {
	script := filepath.Join(repoRootForTest(t), "commands", "build", "run.sh")
	text := string(mustReadFile(t, script))

	for _, required := range []string{
		`common_env_prefix "gascity_doltlite_lib,libsqlite3"`,
		`binary_has_go_build_tag "$output" "gascity_doltlite_lib"`,
		`built gc binary does not report -tags including gascity_doltlite_lib`,
		`built gc binary is missing native DoltLite read-store symbols`,
		`"tags": "%s"`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("build script missing %q", required)
		}
	}
}

func TestDoltliteBuildHelpExplainsTargetSelection(t *testing.T) {
	root := repoRootForTest(t)
	script := string(mustReadFile(t, filepath.Join(root, "commands", "build", "run.sh")))
	help := string(mustReadFile(t, filepath.Join(root, "commands", "build", "help.md")))
	skill := string(mustReadFile(t, filepath.Join(root, "skills", "doltlite", "SKILL.md")))

	for _, required := range []string{
		`gc      Normal iteration path`,
		`all     Bootstrap/coordinated rebuild`,
		`Builds bd, doltlite-client, then gc`,
		`It does not skip unchanged targets`,
		`gc beads-doltlite build gc --install --no-restart`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("build script help missing %q", required)
		}
	}

	for _, required := range []string{
		`The default target is ` + "`gc`",
		`normal Gas City iteration`,
		`all` + "` builds `" + `bd` + "`, `" + `doltlite-client` + "`, then `" + `gc`,
		`does not skip unchanged targets`,
		`does not build libdoltlite itself`,
	} {
		if !strings.Contains(help, required) {
			t.Fatalf("build long help missing %q", required)
		}
	}

	for _, required := range []string{
		`build only ` + "`gc`",
		`gc beads-doltlite build gc --install --no-restart`,
		`only for bootstrap`,
		`does not build libdoltlite itself or skip unchanged targets`,
	} {
		if !strings.Contains(skill, required) {
			t.Fatalf("doltlite skill missing %q", required)
		}
	}
}

func TestDoltliteSqlitebrowserCommandBuildsAgainstLibdoltlite(t *testing.T) {
	root := repoRootForTest(t)
	manifest := string(mustReadFile(t, filepath.Join(root, "commands", "sqlitebrowser", "command.toml")))
	script := string(mustReadFile(t, filepath.Join(root, "commands", "sqlitebrowser", "run.sh")))
	help := string(mustReadFile(t, filepath.Join(root, "commands", "sqlitebrowser", "help.md")))
	skill := string(mustReadFile(t, filepath.Join(root, "skills", "doltlite", "SKILL.md")))

	if !strings.Contains(manifest, "DB Browser for SQLite against libdoltlite") {
		t.Fatalf("sqlitebrowser manifest missing libdoltlite description: %s", manifest)
	}
	for _, required := range []string{
		`usage: gc beads-doltlite sqlitebrowser [open|build|path]`,
		`-Dsqlcipher=0`,
		`-DSQLite3_INCLUDE_DIR="$DOLTLITE_LIB"`,
		`-DSQLite3_LIBRARY="$lib_file"`,
		`libdoltlite`,
		`sqlitebrowser-doltlite`,
		`DISPLAY, WAYLAND_DISPLAY, or QT_QPA_PLATFORM`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("sqlitebrowser script missing %q", required)
		}
	}
	for _, required := range []string{
		`stock SQLite or SQLCipher`,
		`CMake's SQLite dependency at ` + "`libdoltlite`",
		`gc beads-doltlite sqlitebrowser build`,
		`gc beads-doltlite sqlitebrowser open --db`,
	} {
		if !strings.Contains(help, required) {
			t.Fatalf("sqlitebrowser help missing %q", required)
		}
	}
	for _, required := range []string{
		`gc beads-doltlite sqlitebrowser build/open`,
		`stock SQLite Browser builds cannot open DoltLite-format databases`,
	} {
		if !strings.Contains(skill, required) {
			t.Fatalf("doltlite skill missing %q", required)
		}
	}
}

func TestDoltliteSqlitebrowserPathUsesCityMetadata(t *testing.T) {
	root := repoRootForTest(t)
	city := t.TempDir()
	lib := filepath.Join(city, "doltlite-work", "build")
	dbDir := filepath.Join(city, ".beads", "doltlite")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lib, "libdoltlite.so"), []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(city, ".beads", "metadata.json"), []byte(`{"backend":"doltlite","database":"doltlite","dolt_database":"hq"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	wantDB := filepath.Join(dbDir, "hq.db")
	if err := os.WriteFile(wantDB, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", filepath.Join(root, "commands", "sqlitebrowser", "run.sh"), "path", "--city", city, "--lib", lib)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("sqlitebrowser path failed: %v\noutput=%s", err, out.String())
	}
	if got := strings.TrimSpace(out.String()); got != wantDB {
		t.Fatalf("sqlitebrowser path = %q, want %q", got, wantDB)
	}
}

func TestDoltliteGCLinkDoctorRequiresNativeReadBuildTag(t *testing.T) {
	script := filepath.Join(
		repoRootForTest(t),
		"doctor",
		"check-gc-doltlite-link",
		"run.sh",
	)
	text := string(mustReadFile(t, script))

	for _, required := range []string{
		`go version -m "$gc_bin"`,
		`gascity_doltlite_lib`,
		`gc binary was not built with -tags including gascity_doltlite_lib`,
		`gc beads-doltlite build gc --install`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("gc link doctor missing %q", required)
		}
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
	bin := mustReadlink(t, "bd")

	cmd := exec.Command("bash", script, "--json")
	cmd.Env = append(os.Environ(),
		"GC_CITY_PATH="+filepath.Dir(repoRootForTest(t)),
		"PATH="+filepath.Dir(bin)+":"+filepath.Dir(os.Getenv("HOME"))+"/bin:"+os.Getenv("PATH"),
	)

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

func mustReadlink(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("look up %q: %v", name, err)
	}
	return path
}
