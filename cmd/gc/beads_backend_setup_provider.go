package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gastownhall/gascity/internal/config"
)

type beadsBackendSetupContext struct {
	CityPath string
	Provider string
	Backend  string
}

type beadsBackendPluginCapabilities struct {
	// SetupHook means the plugin can initialize scope files and owns
	// .beads/metadata.json creation/normalization for this backend.
	SetupHook bool
	// ProviderLifecycle means the plugin can provide the bd-compatible command
	// surface GC uses for normal bead operations.
	ProviderLifecycle bool
	// BackendPluginMetadata means metadata.json may contain bd backend plugin
	// fields such as backend_plugin_command/backend_plugin_args.
	BackendPluginMetadata bool
	// GascityFastpathMetadata means metadata.json may contain GC fastpath fields
	// such as gascity_backend_command/gascity_backend_args.
	GascityFastpathMetadata bool
	// NativeReadStore means GC may use an optimized in-process read store for
	// hot paths when explicitly enabled and compatible with the metadata shape.
	NativeReadStore bool
	// StoreHealthPath means this backend owns its own on-disk health/size path.
	StoreHealthPath bool
	// BDCompatibility is the bd CLI contract the plugin expects by default.
	BDCompatibility string
}

type beadsBackendPluginEndpoint struct {
	Command  string
	Args     []string
	Protocol string
}

type beadsBackendPlugin interface {
	Name() string
	Capabilities(beadsBackendSetupContext) beadsBackendPluginCapabilities
	SetupHook(beadsBackendSetupContext) (string, bool)
	StorePath(beadsBackendSetupContext) (string, bool)
	BeadsEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool)
	GascityEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool)
}

var beadsBackendSetupRegistry = struct {
	sync.RWMutex
	providers map[string]beadsBackendPlugin
}{providers: map[string]beadsBackendPlugin{}}

func registerBeadsBackendPlugin(provider beadsBackendPlugin) {
	if provider == nil {
		panic("beads backend plugin: nil provider")
	}
	name := strings.TrimSpace(provider.Name())
	if name == "" {
		panic("beads backend plugin: provider with empty name")
	}
	beadsBackendSetupRegistry.Lock()
	defer beadsBackendSetupRegistry.Unlock()
	if _, exists := beadsBackendSetupRegistry.providers[name]; exists {
		panic("beads backend plugin: duplicate provider " + name)
	}
	beadsBackendSetupRegistry.providers[name] = provider
}

func lookupBeadsBackendPlugin(name string) (beadsBackendPlugin, bool) {
	beadsBackendSetupRegistry.RLock()
	defer beadsBackendSetupRegistry.RUnlock()
	provider, ok := beadsBackendSetupRegistry.providers[strings.TrimSpace(name)]
	return provider, ok
}

func beadsProviderSetupHook(cityPath string) (string, bool) {
	if script := strings.TrimSpace(os.Getenv("GC_BEADS_SETUP")); script != "" {
		return script, true
	}
	provider, ctx, ok := beadsBackendPluginForCity(cityPath)
	if !ok {
		return "", false
	}
	return provider.SetupHook(ctx)
}

func beadsBackendPluginForCity(cityPath string) (beadsBackendPlugin, beadsBackendSetupContext, bool) {
	ctx := beadsBackendSetupContext{
		CityPath: cityPath,
		Provider: rawBeadsProvider(cityPath),
		Backend:  beadsBackend(cityPath),
	}
	if ctx.Provider != "plugin" {
		return nil, ctx, false
	}
	if provider, ok := discoveredBeadsBackendPluginForCity(ctx.CityPath, ctx.Backend); ok {
		return provider, ctx, true
	}
	provider, ok := lookupBeadsBackendPlugin(ctx.Backend)
	return provider, ctx, ok
}

func beadsBackendPluginCapabilitiesForCity(cityPath string) (beadsBackendPluginCapabilities, bool) {
	provider, ctx, ok := beadsBackendPluginForCity(cityPath)
	if !ok {
		return beadsBackendPluginCapabilities{}, false
	}
	return provider.Capabilities(ctx), true
}

func beadsBackendPluginStorePath(cityPath string) (string, bool) {
	provider, ctx, ok := beadsBackendPluginForCity(cityPath)
	if !ok {
		return "", false
	}
	return provider.StorePath(ctx)
}

type discoveredBeadsBackendPlugin struct {
	plugin config.DiscoveredBackendPlugin
}

func discoveredBeadsBackendPluginForCity(cityPath, backend string) (beadsBackendPlugin, bool) {
	cfg, err := loadCityConfig(cityPath, io.Discard)
	if err != nil || cfg == nil {
		return nil, false
	}
	backend = strings.ToLower(strings.TrimSpace(backend))
	for _, plugin := range cfg.BackendPlugins {
		if strings.ToLower(strings.TrimSpace(plugin.Backend)) == backend {
			return discoveredBeadsBackendPlugin{plugin: plugin}, true
		}
	}
	return nil, false
}

func (p discoveredBeadsBackendPlugin) Name() string {
	return p.plugin.Backend
}

func (p discoveredBeadsBackendPlugin) Capabilities(beadsBackendSetupContext) beadsBackendPluginCapabilities {
	capSet := make(map[string]bool, len(p.plugin.Capabilities))
	for _, cap := range p.plugin.Capabilities {
		capSet[strings.ToLower(strings.TrimSpace(cap))] = true
	}
	return beadsBackendPluginCapabilities{
		SetupHook:               p.plugin.SetupHook != "" || p.plugin.ProviderCommand != "" || capSet["setup"],
		ProviderLifecycle:       p.plugin.ProviderCommand != "" || p.plugin.BeadsEndpoint.Command != "" || capSet["provider"],
		BackendPluginMetadata:   p.plugin.BeadsEndpoint.Command != "" || capSet["metadata"] || capSet["backend-metadata"],
		GascityFastpathMetadata: p.plugin.GascityEndpoint.Command != "" || capSet["fastpath"] || capSet["gascity-fastpath"],
		NativeReadStore:         capSet["fastpath"] || capSet["native-read"] || capSet["native-read-store"],
		StoreHealthPath:         p.plugin.StorePath != "" || capSet["store-health"],
		BDCompatibility:         p.plugin.BDCompatibility,
	}
}

func (p discoveredBeadsBackendPlugin) SetupHook(beadsBackendSetupContext) (string, bool) {
	for _, candidate := range []string{p.plugin.SetupHook, p.plugin.ProviderCommand} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
			return candidate, true
		}
	}
	return "", false
}

func (p discoveredBeadsBackendPlugin) StorePath(ctx beadsBackendSetupContext) (string, bool) {
	storePath := strings.TrimSpace(p.plugin.StorePath)
	if storePath == "" {
		return "", false
	}
	if filepath.IsAbs(storePath) {
		return storePath, true
	}
	if strings.TrimSpace(ctx.CityPath) == "" {
		return "", false
	}
	return filepath.Join(ctx.CityPath, storePath), true
}

func (p discoveredBeadsBackendPlugin) BeadsEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool) {
	return backendPluginEndpointFromConfig(p.plugin.BeadsEndpoint)
}

func (p discoveredBeadsBackendPlugin) GascityEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool) {
	return backendPluginEndpointFromConfig(p.plugin.GascityEndpoint)
}

func backendPluginEndpointFromConfig(in config.DiscoveredBackendPluginEndpoint) (beadsBackendPluginEndpoint, bool) {
	command := strings.TrimSpace(in.Command)
	if command == "" {
		return beadsBackendPluginEndpoint{}, false
	}
	return beadsBackendPluginEndpoint{
		Command:  command,
		Args:     append([]string(nil), in.Args...),
		Protocol: strings.TrimSpace(in.Protocol),
	}, true
}

type scriptBeadsBackendPlugin struct {
	name       string
	scriptBase string
	storeDir   string
	compat     string
}

func (p scriptBeadsBackendPlugin) Name() string {
	return p.name
}

func (p scriptBeadsBackendPlugin) Capabilities(beadsBackendSetupContext) beadsBackendPluginCapabilities {
	return beadsBackendPluginCapabilities{
		SetupHook:               true,
		ProviderLifecycle:       true,
		BackendPluginMetadata:   true,
		GascityFastpathMetadata: true,
		NativeReadStore:         true,
		StoreHealthPath:         true,
		BDCompatibility:         p.compat,
	}
}

func (p scriptBeadsBackendPlugin) SetupHook(ctx beadsBackendSetupContext) (string, bool) {
	if strings.TrimSpace(ctx.CityPath) == "" {
		return "", false
	}
	script := filepath.Join(ctx.CityPath, ".gc", "scripts", p.scriptBase)
	if st, err := os.Stat(script); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
		return script, true
	}
	return "", false
}

func (p scriptBeadsBackendPlugin) StorePath(ctx beadsBackendSetupContext) (string, bool) {
	if strings.TrimSpace(ctx.CityPath) == "" || strings.TrimSpace(p.storeDir) == "" {
		return "", false
	}
	return filepath.Join(ctx.CityPath, ".beads", p.storeDir), true
}

func (p scriptBeadsBackendPlugin) BeadsEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool) {
	return beadsBackendPluginEndpoint{}, false
}

func (p scriptBeadsBackendPlugin) GascityEndpoint(beadsBackendSetupContext) (beadsBackendPluginEndpoint, bool) {
	return beadsBackendPluginEndpoint{}, false
}

func init() {
	registerBeadsBackendPlugin(scriptBeadsBackendPlugin{
		name:       "doltlite",
		scriptBase: "gc-beads-doltlite-bd.sh",
		storeDir:   "doltlite",
		compat:     "bd-1.0.5",
	})
}

func mustLookupBeadsBackendPlugin(name string) (beadsBackendPlugin, error) {
	provider, ok := lookupBeadsBackendPlugin(name)
	if !ok {
		return nil, fmt.Errorf("unknown beads backend plugin %q", strings.TrimSpace(name))
	}
	return provider, nil
}
