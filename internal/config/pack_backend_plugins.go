package config

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

func packLocalBackendPlugins(tc *PackConfig, packDir, cityRoot string) ([]DiscoveredBackendPlugin, error) {
	if len(tc.BackendPlugins) == 0 {
		return nil, nil
	}
	out := make([]DiscoveredBackendPlugin, 0, len(tc.BackendPlugins))
	seen := map[string]bool{}
	for i, entry := range tc.BackendPlugins {
		backend := strings.ToLower(strings.TrimSpace(entry.Backend))
		if backend == "" {
			return nil, fmt.Errorf("pack %q backend_plugins[%d]: backend is required", tc.Pack.Name, i)
		}
		if seen[backend] {
			return nil, fmt.Errorf("pack %q backend plugin %q: duplicate declaration", tc.Pack.Name, backend)
		}
		seen[backend] = true

		setupHook := resolveBackendPluginPath(strings.TrimSpace(entry.SetupHook), packDir, cityRoot)
		providerCommand := resolveBackendPluginPath(strings.TrimSpace(entry.ProviderCommand), packDir, cityRoot)
		if providerCommand == "" {
			providerCommand = setupHook
		}
		if setupHook == "" && providerCommand == "" {
			return nil, fmt.Errorf("pack %q backend plugin %q: setup_hook or provider_command is required", tc.Pack.Name, backend)
		}
		out = append(out, DiscoveredBackendPlugin{
			Backend:         backend,
			SetupHook:       setupHook,
			ProviderCommand: providerCommand,
			PrepareCommand:  trimStringList(entry.PrepareCommand),
			StorePath:       strings.TrimSpace(entry.StorePath),
			BDCompatibility: strings.TrimSpace(entry.BDCompatibility),
			BeadsEndpoint:   resolveBackendPluginEndpoint(entry.BeadsEndpoint, packDir, cityRoot),
			GascityEndpoint: resolveBackendPluginEndpoint(entry.GascityEndpoint, packDir, cityRoot),
			Capabilities:    trimStringList(entry.Capabilities),
			PackName:        tc.Pack.Name,
			PackDir:         packDir,
		})
	}
	return out, nil
}

func resolveBackendPluginEndpoint(entry PackBackendPluginEndpoint, packDir, cityRoot string) DiscoveredBackendPluginEndpoint {
	return DiscoveredBackendPluginEndpoint{
		Command:  resolveBackendPluginPath(strings.TrimSpace(entry.Command), packDir, cityRoot),
		Args:     append([]string(nil), entry.Args...),
		Protocol: strings.TrimSpace(entry.Protocol),
	}
}

func resolveBackendPluginPath(path, packDir, cityRoot string) string {
	if path == "" || filepath.IsAbs(path) || !strings.Contains(path, "/") {
		return path
	}
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, ".gc"+string(filepath.Separator)) || clean == ".gc" {
		return filepath.Join(cityRoot, clean)
	}
	return filepath.Join(packDir, path)
}

func trimStringList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func mergeCityBackendPlugins(cfg *City, plugins []DiscoveredBackendPlugin) error {
	for _, plugin := range plugins {
		for _, existing := range cfg.BackendPlugins {
			if existing.Backend != plugin.Backend {
				continue
			}
			if sameBackendPluginDeclaration(existing, plugin) {
				continue
			}
			return fmt.Errorf("backend plugin %q: pack %q (%s) conflicts with declaration from pack %q (%s)",
				plugin.Backend, plugin.PackName, plugin.PackDir, existing.PackName, existing.PackDir)
		}
		cfg.BackendPlugins = append(cfg.BackendPlugins, plugin)
	}
	return nil
}

func sameBackendPluginDeclaration(a, b DiscoveredBackendPlugin) bool {
	return samePackDir(a.PackDir, b.PackDir) &&
		a.SetupHook == b.SetupHook &&
		a.ProviderCommand == b.ProviderCommand &&
		reflect.DeepEqual(a.PrepareCommand, b.PrepareCommand) &&
		a.StorePath == b.StorePath &&
		a.BDCompatibility == b.BDCompatibility &&
		a.BeadsEndpoint.Command == b.BeadsEndpoint.Command &&
		reflect.DeepEqual(a.BeadsEndpoint.Args, b.BeadsEndpoint.Args) &&
		a.BeadsEndpoint.Protocol == b.BeadsEndpoint.Protocol &&
		a.GascityEndpoint.Command == b.GascityEndpoint.Command &&
		reflect.DeepEqual(a.GascityEndpoint.Args, b.GascityEndpoint.Args) &&
		a.GascityEndpoint.Protocol == b.GascityEndpoint.Protocol &&
		reflect.DeepEqual(a.Capabilities, b.Capabilities)
}

func filterBackendPluginsByPackDir(plugins []DiscoveredBackendPlugin, packDir string) []DiscoveredBackendPlugin {
	absPackDir, _ := filepath.Abs(packDir)
	var out []DiscoveredBackendPlugin
	for _, plugin := range plugins {
		absDir, _ := filepath.Abs(plugin.PackDir)
		if absDir == absPackDir {
			out = append(out, plugin)
		}
	}
	return out
}

func cachedPackBackendPlugins(cache *packLoadCache, topoDir string) []DiscoveredBackendPlugin {
	if cache == nil {
		return nil
	}
	absDir, err := filepath.Abs(topoDir)
	if err != nil {
		absDir = topoDir
	}
	result, ok := cache.results[absDir]
	if !ok {
		return nil
	}
	return deepCopyBackendPlugins(result.backendPlugins)
}

func deepCopyBackendPlugins(in []DiscoveredBackendPlugin) []DiscoveredBackendPlugin {
	out := append([]DiscoveredBackendPlugin(nil), in...)
	for i := range out {
		out[i].BeadsEndpoint.Args = append([]string(nil), out[i].BeadsEndpoint.Args...)
		out[i].GascityEndpoint.Args = append([]string(nil), out[i].GascityEndpoint.Args...)
		out[i].Capabilities = append([]string(nil), out[i].Capabilities...)
		out[i].PrepareCommand = append([]string(nil), out[i].PrepareCommand...)
	}
	return out
}
