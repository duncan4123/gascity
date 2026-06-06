package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// DoltliteBackendCheck
// ---------------------------------------------------------------------------

type DoltliteBackendCheck struct{ cityPath string }

func NewDoltliteBackendCheck(cityPath string) *DoltliteBackendCheck {
	return &DoltliteBackendCheck{cityPath: cityPath}
}
func (c *DoltliteBackendCheck) Name() string           { return "doltlite-backend" }
func (c *DoltliteBackendCheck) CanFix() bool           { return false }
func (c *DoltliteBackendCheck) Fix(_ *CheckContext) error  { return nil }
func (c *DoltliteBackendCheck) WarmupEligible() bool    { return false }

func (c *DoltliteBackendCheck) Run(_ *CheckContext) *CheckResult {
	r := &CheckResult{Name: c.Name()}
	if !scopeUsesBDDoltliteStore(c.cityPath, c.cityPath) {
		r.Status = StatusOK
		r.Message = "not applicable (non-doltlite backend)"
		return r
	}
	beadsDir := resolveBeadsDir(c.cityPath)
	data, err := os.ReadFile(filepath.Join(beadsDir, "metadata.json"))
	if err != nil {
		r.Status = StatusError
		r.Message = fmt.Sprintf("cannot read beads metadata: %v", err)
		r.FixHint = "run bd init --backend doltlite to repair"
		return r
	}
	var meta struct {
		Backend  string `json:"backend"`
		Database string `json:"database"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		r.Status = StatusError
		r.Message = fmt.Sprintf("cannot parse beads metadata: %v", err)
		r.FixHint = "run bd init --backend doltlite to repair"
		return r
	}
	if !strings.EqualFold(meta.Backend, "doltlite") {
		r.Status = StatusError
		r.Message = fmt.Sprintf("metadata.json backend is %q, expected doltlite", meta.Backend)
		r.FixHint = "run bd init --backend doltlite to repair"
		return r
	}
	r.Status = StatusOK
	r.Message = fmt.Sprintf("doltlite backend configured (database: %s)", meta.Database)
	return r
}

// ---------------------------------------------------------------------------
// DoltliteLibraryCheck
// ---------------------------------------------------------------------------

type DoltliteLibraryCheck struct{ cityPath string }

func NewDoltliteLibraryCheck(cityPath string) *DoltliteLibraryCheck {
	return &DoltliteLibraryCheck{cityPath: cityPath}
}
func (c *DoltliteLibraryCheck) Name() string          { return "doltlite-library" }
func (c *DoltliteLibraryCheck) CanFix() bool          { return false }
func (c *DoltliteLibraryCheck) Fix(_ *CheckContext) error { return nil }
func (c *DoltliteLibraryCheck) WarmupEligible() bool   { return false }

func (c *DoltliteLibraryCheck) Run(_ *CheckContext) *CheckResult {
	r := &CheckResult{Name: c.Name()}
	if !scopeUsesBDDoltliteStore(c.cityPath, c.cityPath) {
		r.Status = StatusOK
		r.Message = "not applicable (non-doltlite backend)"
		return r
	}
	bdPath, err := exec.LookPath("bd")
	if err != nil {
		r.Status = StatusError
		r.Message = "bd binary not found on PATH"
		r.FixHint = "install beads-doltlite binary"
		return r
	}
	out, err := exec.Command("ldd", bdPath).CombinedOutput()
	if err != nil {
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("cannot check library linkage: %v", err)
		return r
	}
	outStr := string(out)
	if !strings.Contains(outStr, "libdoltlite") {
		r.Status = StatusError
		r.Message = "bd binary does not link to libdoltlite.so (built without -tags=libsqlite3?)"
		r.FixHint = "rebuild with: GOFLAGS=-tags=libsqlite3 go build ./cmd/bd"
		return r
	}
	re := regexp.MustCompile(`libdoltlite\.so(?:\.\d+)?\s+=>\s+(.+)`)
	if matches := re.FindStringSubmatch(outStr); len(matches) > 1 {
		r.Status = StatusOK
		r.Message = fmt.Sprintf("linked to %s", matches[1])
	} else {
		r.Status = StatusOK
		r.Message = "linked to libdoltlite.so"
	}
	return r
}

// ---------------------------------------------------------------------------
// DoltliteStoreSizeCheck
// ---------------------------------------------------------------------------

type DoltliteStoreSizeCheck struct {
	cityPath string
	skip     bool
}

func NewDoltliteStoreSizeCheck(cityPath string, skip bool) *DoltliteStoreSizeCheck {
	return &DoltliteStoreSizeCheck{cityPath: cityPath, skip: skip}
}
func (c *DoltliteStoreSizeCheck) Name() string          { return "doltlite-store-size" }
func (c *DoltliteStoreSizeCheck) CanFix() bool          { return false }
func (c *DoltliteStoreSizeCheck) Fix(_ *CheckContext) error { return nil }
func (c *DoltliteStoreSizeCheck) WarmupEligible() bool   { return false }

func (c *DoltliteStoreSizeCheck) Run(_ *CheckContext) *CheckResult {
	r := &CheckResult{Name: c.Name()}
	if c.skip || !scopeUsesBDDoltliteStore(c.cityPath, c.cityPath) {
		r.Status = StatusOK
		r.Message = "not applicable (non-doltlite backend or skip)"
		return r
	}
	doltliteDir := filepath.Join(resolveBeadsDir(c.cityPath), "doltlite")
	totalSize, found, err := sumDirBytes(doltliteDir)
	if err != nil {
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("cannot measure doltlite store: %v", err)
		return r
	}
	if !found {
		r.Status = StatusOK
		r.Message = "no doltlite store directory"
		return r
	}
	switch {
	case totalSize > 10*1024*1024*1024:
		r.Status = StatusError
		r.Message = fmt.Sprintf("doltlite store is large (%s)", formatGB(totalSize))
		r.FixHint = "run `bd flatten --force` and `bd gc --skip-decay --force` to compact"
	case totalSize > 2*1024*1024*1024:
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("doltlite store growing (%s)", formatGB(totalSize))
		r.FixHint = "consider running `bd flatten` to compact"
	default:
		r.Status = StatusOK
		r.Message = fmt.Sprintf("doltlite store size: %s", formatGB(totalSize))
	}
	return r
}

// ---------------------------------------------------------------------------
// DoltliteStaleLockCheck
// ---------------------------------------------------------------------------

type DoltliteStaleLockCheck struct{ cityPath string }

func NewDoltliteStaleLockCheck(cityPath string) *DoltliteStaleLockCheck {
	return &DoltliteStaleLockCheck{cityPath: cityPath}
}
func (c *DoltliteStaleLockCheck) Name() string           { return "doltlite-locks" }
func (c *DoltliteStaleLockCheck) CanFix() bool           { return true }
func (c *DoltliteStaleLockCheck) WarmupEligible() bool    { return false }

func (c *DoltliteStaleLockCheck) Run(_ *CheckContext) *CheckResult {
	r := &CheckResult{Name: c.Name()}
	if !scopeUsesBDDoltliteStore(c.cityPath, c.cityPath) {
		r.Status = StatusOK
		r.Message = "not applicable (non-doltlite backend)"
		return r
	}
	doltliteDir := filepath.Join(resolveBeadsDir(c.cityPath), "doltlite")
	entries, err := os.ReadDir(doltliteDir)
	if err != nil {
		if os.IsNotExist(err) {
			r.Status = StatusOK
			r.Message = "no doltlite store directory"
			return r
		}
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("cannot read doltlite store: %v", err)
		return r
	}
	var lockFiles []string
	for _, entry := range entries {
		if entry.Name() == ".lock" || strings.HasSuffix(entry.Name(), ".lock") {
			lockFiles = append(lockFiles, filepath.Join(doltliteDir, entry.Name()))
		}
	}
	if len(lockFiles) == 0 {
		r.Status = StatusOK
		r.Message = "no lock files"
		return r
	}
	var stale []string
	for _, lf := range lockFiles {
		info, err := os.Stat(lf)
		if err != nil {
			continue
		}
		if info.Size() == 0 {
			stale = append(stale, lf)
		}
	}
	if len(stale) > 0 {
		r.Status = StatusWarning
		r.Message = fmt.Sprintf("%d stale 0-byte lock file(s)", len(stale))
		r.FixHint = "remove stale locks: rm -f " + strings.Join(stale, " ")
	} else {
		r.Status = StatusOK
		r.Message = fmt.Sprintf("%d active lock file(s)", len(lockFiles))
	}
	return r
}

func (c *DoltliteStaleLockCheck) Fix(_ *CheckContext) error {
	doltliteDir := filepath.Join(resolveBeadsDir(c.cityPath), "doltlite")
	entries, err := os.ReadDir(doltliteDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".lock" || strings.HasSuffix(entry.Name(), ".lock") {
			lf := filepath.Join(doltliteDir, entry.Name())
			info, err := os.Stat(lf)
			if err != nil {
				continue
			}
			if info.Size() == 0 {
				os.Remove(lf) //nolint:errcheck
			}
		}
	}
	return nil
}

// resolveBeadsDir returns the .beads directory for a scope.
func resolveBeadsDir(scopePath string) string {
	return filepath.Join(scopePath, ".beads")
}
