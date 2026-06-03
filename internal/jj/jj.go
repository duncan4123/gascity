// Package jj provides Jujutsu workspace operations for agent sandboxing.
// Mirrors internal/git for jj workspace equivalents.
//
// Command construction follows lightjj/internal/jj/commands.go patterns:
// pure functions returning []string arg slices, -- separator for forget,
// template-based structured output for workspace listing.
package jj

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Workspace represents a jj workspace as returned by WorkspaceList.
// Mirrors lightjj/internal/jj/commands.go Workspace struct.
type Workspace struct {
	Name     string
	ChangeId string
	CommitId string
	Current  bool
	Path     string
}

// Jj wraps jj operations scoped to a working directory.
type Jj struct {
	workDir string
	runner  *runner
}

// runner abstracts subprocess execution (mirrors lightjj runner.CommandRunner).
type runner struct {
	workDir string
}

func (r *runner) run(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "jj", args...)
	cmd.Dir = r.workDir
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

// New returns a Jj instance scoped to the given directory.
func New(workDir string) *Jj {
	return &Jj{workDir: workDir, runner: &runner{workDir: workDir}}
}

// IsRepo reports whether workDir is inside a jj repository.
func (j *Jj) IsRepo() bool {
	return j.IsRepoCtx(context.Background())
}

// IsRepoCtx is like IsRepo but accepts a context.
func (j *Jj) IsRepoCtx(ctx context.Context) bool {
	_, err := j.runner.run(ctx, []string{"root"})
	return err == nil
}

// WorkspaceRemove removes a jj workspace: forgets it from jj tracking,
// then deletes the directory. force is accepted for API compatibility
// with git.WorktreeRemove but is ignored.
func (j *Jj) WorkspaceRemove(path string, force bool) error {
	absPath, err := abs(path)
	if err != nil {
		return fmt.Errorf("resolving workspace path %q: %w", path, err)
	}
	name := basename(absPath)

	// Forget via `jj workspace forget -- <name>`.
	// The -- separator prevents flag-like names from being parsed as flags
	// (lightjj pattern: WorkspaceForget uses ["workspace", "forget", "--", name]).
	_, err = j.runner.run(context.Background(), []string{"workspace", "forget", "--", name})
	if err != nil {
		// Non-fatal: workspace may already be forgotten or not tracked.
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("removing jj workspace directory %q: %w", absPath, err)
	}
	return nil
}

// WorktreeRemove delegates to WorkspaceRemove for gitProbe interface compatibility.
func (j *Jj) WorktreeRemove(path string, force bool) error {
	return j.WorkspaceRemove(path, force)
}

// WorkspaceList returns all jj workspaces using a structured template.
// Uses the lightjj WorkspaceList template pattern — delimited fields
// instead of free-text parsing.
func (j *Jj) WorkspaceList() ([]Workspace, error) {
	out, err := j.runner.run(context.Background(), workspaceListArgs())
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	return parseWorkspaceList(string(out)), nil
}

// workspaceListArgs mirrors lightjj WorkspaceList(withRoot bool) template.
// Output per workspace: name\x1Fchange_id\x1Fcommit_id\x1Fcurrent[\x1Fpath]
func workspaceListArgs() []string {
	tmpl := `name ++ "\x1F" ++ change_id.short() ++ "\x1F" ++ commit_id.short() ++ "\x1F" ++ stringify(current_working_copy()) ++ "\x1F" ++ root ++ "\n"`
	return []string{"workspace", "list", "--template", tmpl, "--color", "never", "--ignore-working-copy"}
}

// parseWorkspaceList mirrors lightjj ParseWorkspaceList.
func parseWorkspaceList(output string) []Workspace {
	var workspaces []Workspace
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\x1F", 5)
		if len(fields) < 4 {
			continue
		}
		ws := Workspace{
			Name:     fields[0],
			ChangeId: fields[1],
			CommitId: fields[2],
			Current:  fields[3] == "true",
		}
		if len(fields) >= 5 {
			ws.Path = fields[4]
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces
}

// HasUncommittedWork reports whether the working directory has uncommitted
// changes. Checks if @ differs from its parent via diff --stat.
func (j *Jj) HasUncommittedWork() bool {
	out, err := j.runner.run(context.Background(), []string{"diff", "--stat", "-r", "@-"})
	if err != nil {
		return true // assume dirty on error (safe default)
	}
	return strings.TrimSpace(string(out)) != ""
}

// HasUnpushedCommits reports whether the workspace has unpushed work.
func (j *Jj) HasUnpushedCommits() bool {
	has, err := j.HasUnpushedCommitsResult()
	if err != nil {
		return true
	}
	return has
}

// HasUnpushedCommitsResult checks for unpushed commits via git
// (colocated repos share git refs for remote push status).
func (j *Jj) HasUnpushedCommitsResult() (bool, error) {
	cmd := exec.Command("git", "log", "HEAD", "--oneline", "--not", "--remotes")
	cmd.Dir = j.workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("checking unpushed commits: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// HasStashes always returns false — jj has no stash concept.
func (j *Jj) HasStashes() bool { return false }

// HasStashesResult always returns (false, nil).
func (j *Jj) HasStashesResult() (bool, error) { return false, nil }

// WorkspacePrune is a no-op for jj — workspaces are auto-managed.
func (j *Jj) WorkspacePrune() error { return nil }

// Ensure checks that the jj binary is available.
func Ensure() error {
	if _, err := exec.LookPath("jj"); err != nil {
		return fmt.Errorf("jj binary not found in PATH: %w", err)
	}
	cmd := exec.Command("jj", "--version")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("jj --version failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// helpers
func abs(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		return path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd + "/" + path, nil
}

func basename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
