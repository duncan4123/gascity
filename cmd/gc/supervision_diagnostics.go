package main

import (
	"fmt"
	"path/filepath"

	"github.com/gastownhall/gascity/internal/config"
	"github.com/gastownhall/gascity/internal/doctor"
	"github.com/gastownhall/gascity/internal/fsys"
	"github.com/gastownhall/gascity/internal/supervisor"
)

func supervisionStatusForCity(cityPath string) SupervisionJSON {
	status := SupervisionJSON{RegistryWarnings: citySupervisorRegistryDiagnostics(cityPath)}
	cfg, err := config.Load(fsys.OSFS{}, filepath.Join(cityPath, "city.toml"))
	if err == nil {
		status.Required = cfg.Supervision.RequiresSupervisor()
	}
	return status
}

// citySupervisorRegistryDiagnostics detects the ambiguous situation where a
// command is pointed at an isolated GC_HOME while the city is registered with
// the shared machine supervisor. It also recognizes the old per-city
// .gc-home registry left by earlier development workflows.
func citySupervisorRegistryDiagnostics(cityPath string) []string {
	return citySupervisorRegistryDiagnosticsForHomes(cityPath, supervisor.DefaultHome(), supervisor.BuiltinDefaultHome())
}

func citySupervisorRegistryDiagnosticsForHomes(cityPath, activeHome, defaultHome string) []string {
	activeHome = normalizePathForCompare(activeHome)
	defaultHome = normalizePathForCompare(defaultHome)
	activeRegistered, _ := cityRegisteredInSupervisorHome(activeHome, cityPath)
	defaultRegistered := activeRegistered
	if !samePath(activeHome, defaultHome) {
		defaultRegistered, _ = cityRegisteredInSupervisorHome(defaultHome, cityPath)
	}

	var diagnostics []string
	if !samePath(activeHome, defaultHome) && defaultRegistered {
		diagnostics = append(diagnostics, fmt.Sprintf(
			"GC_HOME=%s is isolated from machine supervisor home %s, which registers this city; run unset GC_HOME",
			activeHome, defaultHome,
		))
	}

	legacyHome := normalizePathForCompare(filepath.Join(cityPath, ".gc-home"))
	if !samePath(legacyHome, activeHome) && !samePath(legacyHome, defaultHome) {
		if legacyRegistered, _ := cityRegisteredInSupervisorHome(legacyHome, cityPath); legacyRegistered && (activeRegistered || defaultRegistered) {
			diagnostics = append(diagnostics, fmt.Sprintf(
				"legacy city-local registry %s also registers this city; remove it after confirming the shared supervisor is healthy",
				legacyHome,
			))
		}
	}
	return diagnostics
}

func cityRegisteredInSupervisorHome(home, cityPath string) (bool, error) {
	entries, err := supervisor.NewRegistry(filepath.Join(home, "cities.toml")).List()
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if samePath(entry.Path, cityPath) {
			return true, nil
		}
	}
	return false, nil
}

type supervisionRegistryDoctorCheck struct {
	cityPath string
	cfg      *config.City
}

func newSupervisionRegistryDoctorCheck(cityPath string, cfg *config.City) *supervisionRegistryDoctorCheck {
	return &supervisionRegistryDoctorCheck{cityPath: cityPath, cfg: cfg}
}

func (*supervisionRegistryDoctorCheck) Name() string                     { return "supervision-registry" }
func (*supervisionRegistryDoctorCheck) CanFix() bool                     { return false }
func (*supervisionRegistryDoctorCheck) Fix(_ *doctor.CheckContext) error { return nil }
func (*supervisionRegistryDoctorCheck) WarmupEligible() bool             { return false }

func (c *supervisionRegistryDoctorCheck) Run(_ *doctor.CheckContext) *doctor.CheckResult {
	r := &doctor.CheckResult{Name: c.Name()}
	diagnostics := citySupervisorRegistryDiagnostics(c.cityPath)
	registered, err := cityRegisteredInSupervisorHome(supervisor.DefaultHome(), c.cityPath)
	if err != nil {
		r.Status = doctor.StatusWarning
		r.Message = fmt.Sprintf("read active supervisor registry: %v", err)
		return r
	}
	if len(diagnostics) > 0 {
		r.Status = doctor.StatusWarning
		r.Message = "supervisor registry context is ambiguous"
		r.Details = diagnostics
		r.FixHint = "use the machine-wide GC_HOME (usually unset GC_HOME) and remove stale city-local registries"
		return r
	}
	if c.cfg != nil && c.cfg.Supervision.RequiresSupervisor() && !registered {
		r.Status = doctor.StatusError
		r.Message = "city requires machine-supervisor ownership but is not registered"
		r.FixHint = "run gc start from the machine-wide GC_HOME"
		return r
	}
	if c.cfg != nil && c.cfg.Supervision.RequiresSupervisor() {
		r.Status = doctor.StatusOK
		r.Message = "machine-supervisor ownership required and registered"
		return r
	}
	r.Status = doctor.StatusOK
	r.Message = "no explicit machine-supervisor ownership requirement"
	return r
}
