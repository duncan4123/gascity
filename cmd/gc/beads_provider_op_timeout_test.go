package main

import (
	"testing"
	"time"
)

func TestProviderOpTimeoutInitGetsLongWindow(t *testing.T) {
	const long = 120 * time.Second
	const short = 30 * time.Second
	for _, op := range []string{"start", "recover", "init"} {
		if got := providerOpTimeout(op); got != long {
			t.Errorf("providerOpTimeout(%q) = %v, want %v", op, got, long)
		}
	}
	for _, op := range []string{"health", "stop", "probe", ""} {
		if got := providerOpTimeout(op); got != short {
			t.Errorf("providerOpTimeout(%q) = %v, want %v", op, got, short)
		}
	}
}
