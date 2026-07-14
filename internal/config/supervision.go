package config

// SupervisionConfig expresses a city's lifecycle-ownership requirement.
//
// The supervisor itself is machine-wide infrastructure. This city-level
// contract only says that a standalone controller is not an acceptable owner
// for the city. New cities write Required=true; an omitted table remains a
// compatibility mode for older cities that may still use gc start --foreground.
type SupervisionConfig struct {
	Required bool `toml:"required,omitempty"`
}

// RequiresSupervisor reports whether the city must be run by the shared
// supervisor rather than a standalone controller.
func (s SupervisionConfig) RequiresSupervisor() bool {
	return s.Required
}
