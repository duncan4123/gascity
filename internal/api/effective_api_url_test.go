package api

import (
	"errors"
	"testing"

	"github.com/gastownhall/gascity/internal/config"
)

type fakeEffectiveAPIConfigClient struct {
	err error
}

func (c fakeEffectiveAPIConfigClient) ListCities() ([]CityInfo, error) {
	if c.err != nil {
		return nil, c.err
	}
	return []CityInfo{{Name: "test-city"}}, nil
}

func TestConfigEffectiveAPIURLUsesReachableSupervisor(t *testing.T) {
	originalBaseURLHook := configEffectiveAPIBaseURLHook
	originalClientFactory := configEffectiveAPIClientFactory
	t.Cleanup(func() {
		configEffectiveAPIBaseURLHook = originalBaseURLHook
		configEffectiveAPIClientFactory = originalClientFactory
	})
	configEffectiveAPIBaseURLHook = func() (string, error) {
		return "http://127.0.0.1:9443/", nil
	}
	configEffectiveAPIClientFactory = func(baseURL string) effectiveAPIConfigClient {
		if baseURL != "http://127.0.0.1:9443" {
			t.Fatalf("baseURL = %q", baseURL)
		}
		return fakeEffectiveAPIConfigClient{}
	}

	state := newFakeState(t)
	state.cfg.API = config.APIConfig{Port: 7777}

	got := configEffectiveAPIURL(state)
	if got != "http://127.0.0.1:9443" {
		t.Fatalf("configEffectiveAPIURL() = %q", got)
	}
}

func TestConfigEffectiveAPIURLFallsBackToCityAPI(t *testing.T) {
	originalBaseURLHook := configEffectiveAPIBaseURLHook
	originalClientFactory := configEffectiveAPIClientFactory
	t.Cleanup(func() {
		configEffectiveAPIBaseURLHook = originalBaseURLHook
		configEffectiveAPIClientFactory = originalClientFactory
	})
	configEffectiveAPIBaseURLHook = func() (string, error) {
		return "http://127.0.0.1:9443", nil
	}
	configEffectiveAPIClientFactory = func(string) effectiveAPIConfigClient {
		return fakeEffectiveAPIConfigClient{err: errors.New("unreachable")}
	}

	state := newFakeState(t)
	state.cfg.API = config.APIConfig{Port: 7777, Bind: "0.0.0.0"}

	got := configEffectiveAPIURL(state)
	if got != "http://127.0.0.1:7777" {
		t.Fatalf("configEffectiveAPIURL() = %q", got)
	}
}
