package main

import (
	"testing"
	"time"
)

// TestWorkQueryTimeoutsAccommodateMultiRoundTripProbe guards the work-query
// timeout budget. The default work-probe performs multiple sequential bd/store
// round-trips before reaching the pool-demand tier that finds routed work. On a
// busy multi-rig DoltLite city, a too-small cap kills the probe before routed
// pool work can be claimed.
func TestWorkQueryTimeoutsAccommodateMultiRoundTripProbe(t *testing.T) {
	const minProbeBudget = 60 * time.Second

	if hookWorkQueryTimeout < minProbeBudget {
		t.Errorf("hookWorkQueryTimeout = %s, want >= %s (multi-round-trip probe budget)", hookWorkQueryTimeout, minProbeBudget)
	}
}
