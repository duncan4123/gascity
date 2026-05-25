package main

import (
	"testing"
	"time"
)

func TestBoundedStatusCallDegradesAfterTimeout(t *testing.T) {
	originalTimeout := statusProviderCallTimeout
	originalWindow := statusProviderDegradeWindow
	t.Cleanup(func() {
		statusProviderCallTimeout = originalTimeout
		statusProviderDegradeWindow = originalWindow
	})
	statusProviderCallTimeout = 5 * time.Millisecond
	statusProviderDegradeWindow = time.Minute

	provider := &statusProvider{}
	calls := 0
	got := boundedStatusCall(provider, "fallback", func() string {
		calls++
		time.Sleep(50 * time.Millisecond)
		return "slow"
	})
	if got != "fallback" {
		t.Fatalf("first boundedStatusCall() = %q", got)
	}

	got = boundedStatusCall(provider, "degraded", func() string {
		calls++
		return "unexpected"
	})
	if got != "degraded" {
		t.Fatalf("degraded boundedStatusCall() = %q", got)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}
