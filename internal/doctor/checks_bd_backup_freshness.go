package doctor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gastownhall/gascity/internal/config"
)

const defaultBackupFreshnessMaxAge = 24 * time.Hour

// BdBackupFreshnessCheck warns when bd's backup state exists but has not
// advanced recently. Backup presence alone is not enough: a disabled or broken
// sync pipeline can leave an old recovery point that still looks configured.
type BdBackupFreshnessCheck struct {
	cityPath   string
	scopeRoots []string
	maxAge     time.Duration
	now        func() time.Time
}

func NewBdBackupFreshnessCheckForConfig(cityPath string, cfg *config.City, cfgErr error) *BdBackupFreshnessCheck {
	return &BdBackupFreshnessCheck{
		cityPath:   cityPath,
		scopeRoots: managedDoltScopeRootsForConfig(cityPath, cfg, cfgErr),
		maxAge:     defaultBackupFreshnessMaxAge,
		now:        time.Now,
	}
}

func NewBdBackupFreshnessCheckForScopeRoots(cityPath string, scopeRoots []string, maxAge time.Duration, now func() time.Time) *BdBackupFreshnessCheck {
	if maxAge <= 0 {
		maxAge = defaultBackupFreshnessMaxAge
	}
	if now == nil {
		now = time.Now
	}
	return &BdBackupFreshnessCheck{cityPath: cityPath, scopeRoots: scopeRoots, maxAge: maxAge, now: now}
}

func (c *BdBackupFreshnessCheck) Name() string { return "bd-backup-freshness" }

func (c *BdBackupFreshnessCheck) WarmupEligible() bool { return false }

func (c *BdBackupFreshnessCheck) CanFix() bool { return false }

func (c *BdBackupFreshnessCheck) Fix(_ *CheckContext) error { return nil }

func (c *BdBackupFreshnessCheck) Run(_ *CheckContext) *CheckResult {
	r := &CheckResult{Name: c.Name()}
	now := c.now()

	var findings []string
	for _, target := range c.freshnessScanTargets() {
		if finding, ok := scanBackupFreshness(target.Label, target.BeadsDir, now, c.maxAge); ok {
			findings = append(findings, finding)
		}
	}

	if len(findings) == 0 {
		r.Status = StatusOK
		r.Message = "all configured bd backups synced within " + c.maxAge.String()
		return r
	}
	sort.Strings(findings)
	r.Status = StatusWarning
	r.Severity = SeverityAdvisory
	r.Message = strings.Join(findings, "; ")
	r.FixHint = "re-enable or repair the bd backup pipeline for the listed scopes " +
		"(bd backup sync; verify backup.enabled and BD_BACKUP_ENABLED), then confirm " +
		"bd backup status shows a recent sync"
	return r
}

type bdBackupFreshnessTarget struct {
	Label    string
	BeadsDir string
}

func (c *BdBackupFreshnessCheck) freshnessScanTargets() []bdBackupFreshnessTarget {
	scopeRoots := c.scopeRoots
	if len(scopeRoots) == 0 {
		scopeRoots = managedDoltScopeRoots(c.cityPath)
	}
	if len(scopeRoots) == 0 {
		scopeRoots = []string{c.cityPath}
	}

	seen := make(map[string]struct{}, len(scopeRoots))
	targets := make([]bdBackupFreshnessTarget, 0, len(scopeRoots))
	for _, scopeRoot := range scopeRoots {
		scopeRoot = strings.TrimSpace(scopeRoot)
		if scopeRoot == "" {
			continue
		}
		scopeRoot = filepath.Clean(scopeRoot)
		if _, ok := seen[scopeRoot]; ok {
			continue
		}
		seen[scopeRoot] = struct{}{}
		targets = append(targets, bdBackupFreshnessTarget{
			Label:    bdBackupScopeLabel(c.cityPath, scopeRoot),
			BeadsDir: filepath.Join(scopeRoot, ".beads"),
		})
	}
	return targets
}

func scanBackupFreshness(label, beadsDir string, now time.Time, maxAge time.Duration) (string, bool) {
	path := filepath.Join(beadsDir, "backup", "backup_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false
		}
		return fmt.Sprintf("%s: read backup_state.json: %v", label, err), true
	}
	var state struct {
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Sprintf("%s: backup_state.json is unparseable: %v", label, err), true
	}
	ts := strings.TrimSpace(state.Timestamp)
	if ts == "" {
		return fmt.Sprintf("%s: backup_state.json has no timestamp", label), true
	}
	synced, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return fmt.Sprintf("%s: backup_state.json timestamp %q is unparseable: %v", label, ts, err), true
	}
	if age := now.Sub(synced); age > maxAge {
		return fmt.Sprintf("%s: last bd backup sync was %s ago (> %s) - backup pipeline may be disabled or broken",
			label, age.Round(time.Minute), maxAge), true
	}
	return "", false
}
