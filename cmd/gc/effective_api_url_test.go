package main

import (
	"errors"
	"testing"

	"github.com/gastownhall/gascity/internal/api"
	"github.com/gastownhall/gascity/internal/config"
)

type fakeEffectiveAPIClient struct {
	err error
}

func (c fakeEffectiveAPIClient) ListCities() ([]api.CityInfo, error) {
	if c.err != nil {
		return nil, c.err
	}
	return []api.CityInfo{{Name: "test-city"}}, nil
}

func TestResolveEffectiveAPIURLUsesReachableSupervisor(t *testing.T) {
	t.Cleanup(func() {
		effectiveAPIBaseURLHook = supervisorAPIBaseURL
		effectiveAPIClientFactory = func(baseURL string) effectiveAPIClient {
			return api.NewClient(baseURL)
		}
	})

	effectiveAPIBaseURLHook = func() (string, error) {
		return "http://127.0.0.1:9443/", nil
	}
	effectiveAPIClientFactory = func(baseURL string) effectiveAPIClient {
		if baseURL != "http://127.0.0.1:9443" {
			t.Fatalf("baseURL = %q", baseURL)
		}
		return fakeEffectiveAPIClient{}
	}

	got := resolveEffectiveAPIURL(t.TempDir(), &config.City{
		API: config.APIConfig{Port: 7777},
	})
	if got != "http://127.0.0.1:9443" {
		t.Fatalf("resolveEffectiveAPIURL() = %q", got)
	}
}

func TestResolveEffectiveAPIURLFallsBackToStandalone(t *testing.T) {
	t.Cleanup(func() {
		effectiveAPIBaseURLHook = supervisorAPIBaseURL
		effectiveAPIClientFactory = func(baseURL string) effectiveAPIClient {
			return api.NewClient(baseURL)
		}
	})

	effectiveAPIBaseURLHook = func() (string, error) {
		return "http://127.0.0.1:9443", nil
	}
	effectiveAPIClientFactory = func(string) effectiveAPIClient {
		return fakeEffectiveAPIClient{err: errors.New("unreachable")}
	}

	got := resolveEffectiveAPIURL(t.TempDir(), &config.City{
		API: config.APIConfig{Port: 7777, Bind: "0.0.0.0"},
	})
	if got != "http://127.0.0.1:7777" {
		t.Fatalf("resolveEffectiveAPIURL() = %q", got)
	}
}
