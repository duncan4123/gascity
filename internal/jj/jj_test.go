package jj

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a jj repo with one commit in a temp directory.
// Uses jj git init --colocate to create a colocated git+jj repo.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init git repo first, then init jj colocated.
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	// Need an initial commit for jj to work with.
	runGit(t, dir, "commit", "--allow-empty", "-m", "init")

	// Init jj colocated.
	runJJ(t, dir, "git", "init", "--colocate")
	return dir
}

// runJJ runs a jj command in dir and fails the test on error.
func runJJ(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("jj", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj %s: %s: %v", strings.Join(args, " "), out, err)
	}
}

// runGit runs a git command in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %s: %v", strings.Join(args, " "), out, err)
	}
}

func TestEnsure(t *testing.T) {
	if err := Ensure(); err != nil {
		t.Fatalf("Ensure() = %v, want nil (jj binary not found)", err)
	}
}

func TestIsRepo(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)
	if !j.IsRepo() {
		t.Error("IsRepo() = false for jj repo, want true")
	}

	notRepo := t.TempDir()
	j2 := New(notRepo)
	if j2.IsRepo() {
		t.Error("IsRepo() = true for non-repo, want false")
	}
}

func TestIsRepoCtx(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)
	ctx := t.Context()
	if !j.IsRepoCtx(ctx) {
		t.Error("IsRepoCtx() = false for jj repo, want true")
	}
}

func TestWorkspaceList(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	workspaces, err := j.WorkspaceList()
	if err != nil {
		t.Fatalf("WorkspaceList: %v", err)
	}

	// Should have at least 1 workspace (the default).
	if len(workspaces) < 1 {
		t.Fatalf("len(workspaces) = %d, want >= 1", len(workspaces))
	}

	// The default workspace should be current.
	foundDefault := false
	for _, ws := range workspaces {
		if ws.Name == "default" {
			foundDefault = true
			if !ws.Current {
				t.Error("default workspace should be current")
			}
		}
	}
	if !foundDefault {
		t.Error("default workspace not found in list")
	}
}

func TestWorkspaceAddAndRemove(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// Create a new jj workspace (simulating the formula step).
	wsPath := filepath.Join(t.TempDir(), "test-workspace")
	runJJ(t, repo, "workspace", "add", wsPath, "--sparse-patterns", "full")

	// Verify it appears in the list.
	workspaces, err := j.WorkspaceList()
	if err != nil {
		t.Fatalf("WorkspaceList after add: %v", err)
	}

	var found bool
	for _, ws := range workspaces {
		if ws.Name == filepath.Base(wsPath) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new workspace %q not found in list; got %d workspaces", wsPath, len(workspaces))
	}

	// Remove the workspace.
	if err := j.WorkspaceRemove(wsPath, false); err != nil {
		t.Fatalf("WorkspaceRemove: %v", err)
	}

	// Directory should be gone.
	if _, err := os.Stat(wsPath); !os.IsNotExist(err) {
		t.Errorf("workspace dir %q still exists after remove", wsPath)
	}
}

func TestWorkspaceRemove_Force(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	wsPath := filepath.Join(t.TempDir(), "dirty-workspace")
	runJJ(t, repo, "workspace", "add", wsPath, "--sparse-patterns", "full")

	// Create an uncommitted file to simulate dirty workspace.
	if err := os.WriteFile(filepath.Join(wsPath, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Force remove should succeed even with dirty workspace.
	if err := j.WorkspaceRemove(wsPath, true); err != nil {
		t.Fatalf("WorkspaceRemove(force): %v", err)
	}
}

func TestWorktreeRemove_DelegatesToWorkspaceRemove(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	wsPath := filepath.Join(t.TempDir(), "delegate-ws")
	runJJ(t, repo, "workspace", "add", wsPath, "--sparse-patterns", "full")

	// WorktreeRemove should delegate to WorkspaceRemove.
	if err := j.WorktreeRemove(wsPath, false); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}

	if _, err := os.Stat(wsPath); !os.IsNotExist(err) {
		t.Errorf("workspace dir %q still exists after WorktreeRemove", wsPath)
	}
}

func TestHasUncommittedWork_Clean(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	if j.HasUncommittedWork() {
		t.Error("HasUncommittedWork() = true for clean repo, want false")
	}
}

func TestHasUncommittedWork_Dirty(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// Create a file to make the working copy dirty.
	dirtyFile := filepath.Join(repo, "dirty.txt")
	if err := os.WriteFile(dirtyFile, []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !j.HasUncommittedWork() {
		t.Error("HasUncommittedWork() = false for dirty repo, want true")
	}
}

func TestHasUnpushedCommits_None(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// Fresh repo with no remote should report unpushed (safe default).
	// Actually, "no remote" means all commits are unpushed by definition.
	// This test just verifies the method doesn't panic or error unexpectedly.
	has, err := j.HasUnpushedCommitsResult()
	if err != nil {
		t.Logf("HasUnpushedCommitsResult error (expected with no remote): %v", err)
		// HasUnpushedCommits() returns true on error (safe default).
		if !j.HasUnpushedCommits() {
			t.Error("HasUnpushedCommits() = false on error, want true (safe default)")
		}
		return
	}
	// No remote: all commits are unpushed.
	if !has {
		t.Log("no remote: HasUnpushedCommits = false (may be expected)")
	}
}

func TestHasStashes(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// jj has no stash concept, always returns false.
	if j.HasStashes() {
		t.Error("HasStashes() = true, want false (jj has no stashes)")
	}

	has, err := j.HasStashesResult()
	if err != nil {
		t.Fatalf("HasStashesResult: %v", err)
	}
	if has {
		t.Error("HasStashesResult() = true, want false")
	}
}

func TestWorkspacePrune(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// WorkspacePrune is a no-op for jj.
	if err := j.WorkspacePrune(); err != nil {
		t.Fatalf("WorkspacePrune: %v", err)
	}
}

func TestParseWorkspaceList(t *testing.T) {
	// Simulated output from the jj template.
	output := "default\x1Fabc123\x1Fdef456\x1Ftrue\x1F/path/to/repo\n"
	workspaces := parseWorkspaceList(output)

	if len(workspaces) != 1 {
		t.Fatalf("len(workspaces) = %d, want 1", len(workspaces))
	}

	ws := workspaces[0]
	if ws.Name != "default" {
		t.Errorf("Name = %q, want %q", ws.Name, "default")
	}
	if ws.ChangeId != "abc123" {
		t.Errorf("ChangeId = %q, want %q", ws.ChangeId, "abc123")
	}
	if ws.CommitId != "def456" {
		t.Errorf("CommitId = %q, want %q", ws.CommitId, "def456")
	}
	if !ws.Current {
		t.Error("Current = false, want true")
	}
	if ws.Path != "/path/to/repo" {
		t.Errorf("Path = %q, want %q", ws.Path, "/path/to/repo")
	}
}

func TestParseWorkspaceList_Empty(t *testing.T) {
	workspaces := parseWorkspaceList("")
	if len(workspaces) != 0 {
		t.Errorf("len(workspaces) = %d, want 0", len(workspaces))
	}
}

func TestParseWorkspaceList_Multiple(t *testing.T) {
	output := "default\x1Fabc\x1Fdef\x1Ftrue\x1F/repo\n" +
		"other\x1Fghi\x1Fjkl\x1Ffalse\x1F/other\n"
	workspaces := parseWorkspaceList(output)

	if len(workspaces) != 2 {
		t.Fatalf("len(workspaces) = %d, want 2", len(workspaces))
	}

	if workspaces[0].Name != "default" || !workspaces[0].Current {
		t.Errorf("first workspace = %+v, want default/current", workspaces[0])
	}
	if workspaces[1].Name != "other" || workspaces[1].Current {
		t.Errorf("second workspace = %+v, want other/not-current", workspaces[1])
	}
}

func TestAbs(t *testing.T) {
	// Absolute path passes through.
	got, err := abs("/absolute/path")
	if err != nil {
		t.Fatalf("abs(/absolute/path): %v", err)
	}
	if got != "/absolute/path" {
		t.Errorf("abs(/absolute/path) = %q, want /absolute/path", got)
	}

	// Relative path is made absolute.
	got, err = abs("relative/path")
	if err != nil {
		t.Fatalf("abs(relative/path): %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("abs(relative/path) = %q, want absolute path", got)
	}
	if !strings.HasSuffix(got, "/relative/path") {
		t.Errorf("abs(relative/path) = %q, want suffix /relative/path", got)
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/foo/bar/baz", "baz"},
		{"foo/bar", "bar"},
		{"foo", "foo"},
		{"/foo", "foo"},
		{"foo/", ""},
	}

	for _, tt := range tests {
		got := basename(tt.path)
		if got != tt.want {
			t.Errorf("basename(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestWorkspaceList_MultipleWorkspaces(t *testing.T) {
	repo := initTestRepo(t)
	j := New(repo)

	// Create a second workspace.
	ws2Path := filepath.Join(t.TempDir(), "ws2")
	runJJ(t, repo, "workspace", "add", ws2Path, "--sparse-patterns", "full")

	workspaces, err := j.WorkspaceList()
	if err != nil {
		t.Fatalf("WorkspaceList: %v", err)
	}

	// Should have at least 2 workspaces.
	if len(workspaces) < 2 {
		t.Fatalf("len(workspaces) = %d, want >= 2", len(workspaces))
	}

	// Cleanup.
	_ = j.WorkspaceRemove(ws2Path, true)
}
