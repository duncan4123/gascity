package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gastownhall/gascity/internal/beads"
	"github.com/gastownhall/gascity/internal/beads/contract"
	"github.com/gastownhall/gascity/internal/fsys"
	"github.com/gastownhall/gascity/internal/pgauth"
)

// initBeadsViaNativeBackend initializes a scope through the backend seam built
// into stock upstream bd. It deliberately bypasses the external-plugin setup
// hook: native backends have no endpoint process for Gas City to supervise.
func initBeadsViaNativeBackend(cityPath, dir, prefix string) (bool, error) {
	driver, packSQLitePath, ok := nativeBeadsBackendForCity(cityPath)
	if !ok {
		return false, nil
	}
	cfg, err := loadCityConfig(cityPath, io.Discard)
	if err != nil {
		return true, err
	}
	scope := beadsBackendScopeContextForCityScope(cityPath, dir, prefix, "")
	args := []string{
		"init",
		"--backend=" + driver,
		"--prefix=" + prefix,
		"--skip-hooks",
		"--skip-agents",
		"--init-if-missing",
		"--quiet",
	}
	env := nativeBackendInitEnv(cityPath, dir)
	metadataState := contract.MetadataState{Backend: driver}
	switch driver {
	case "sqlite":
		sqlitePath := strings.TrimSpace(cfg.Beads.SQLitePath)
		if sqlitePath == "" {
			sqlitePath = packSQLitePath
		}
		if sqlitePath != "" {
			args = append(args, "--sqlite-path="+sqlitePath)
		}
		metadataState.SQLitePath = sqlitePath
		if metadataState.SQLitePath == "" {
			metadataState.SQLitePath = "beads.db"
		}
	case "postgres":
		postgresURL := strings.TrimSpace(cfg.Beads.PostgresURL)
		if postgresURL == "" {
			postgresURL = pgauth.PostgresURL()
		}
		if postgresURL == "" {
			return true, fmt.Errorf("native postgres backend requires [beads].postgres_url or BEADS_POSTGRES_URL")
		}
		postgresURL, resolved, err := postgresInitURL(cityPath, dir, postgresURL)
		if err != nil {
			return true, err
		}
		if resolved.Password != "" {
			env["BEADS_PG_PASSWORD"] = resolved.Password
			env["BEADS_POSTGRES_PASSWORD"] = resolved.Password // compatibility with pre-1.1 bd builds
		}
		schema := strings.TrimSpace(cfg.Beads.PostgresSchema)
		if schema == "" || !samePath(cityPath, dir) {
			schema = strings.TrimSpace(scope.Namespace)
		}
		if schema == "" {
			schema = prefix
		}
		args = append(args, "--pg-url="+postgresURL, "--pg-schema="+schema)
	default:
		return true, fmt.Errorf("unsupported native beads backend %q", driver)
	}

	runner := beads.ExecCommandRunnerWithEnv(env)
	if _, err := runner(dir, "bd", args...); err != nil {
		if isBdAlreadyInitializedError(err) {
			return true, nil
		}
		return true, fmt.Errorf("bd native %s init for %s: %w", driver, dir, err)
	}
	// gc may have pre-seeded Dolt connection fields before pack composition
	// selected the native backend. Upstream bd owns and writes the native
	// fields; this pass only removes stale fields belonging to other backends.
	if _, err := contract.EnsureCanonicalMetadata(
		fsys.OSFS{}, filepath.Join(dir, ".beads", "metadata.json"), metadataState,
	); err != nil {
		return true, fmt.Errorf("normalize native %s metadata for %s: %w", driver, dir, err)
	}
	return true, nil
}

func nativeBackendInitEnv(cityPath, scopeRoot string) map[string]string {
	env := cityRuntimeEnvMapForCity(cityPath)
	env["BEADS_DIR"] = filepath.Join(scopeRoot, ".beads")
	env["BD_EXPORT_AUTO"] = "false"
	applyBdContributorRoutingOptOut(env)
	applyBdCLIRemoteSyncOptOut(env)
	applyBdAutoBackupOptOut(env)
	return env
}

func postgresInitURL(cityPath, scopeRoot, raw string) (string, pgauth.Resolved, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return "", pgauth.Resolved{}, fmt.Errorf("invalid native postgres URL: %q", raw)
	}
	if parsed.User == nil {
		return "", pgauth.Resolved{}, fmt.Errorf("native postgres URL must include a username")
	}
	if _, hasPassword := parsed.User.Password(); hasPassword {
		return raw, pgauth.Resolved{User: parsed.User.Username()}, nil
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "5432"
	}
	resolved, resolveErr := pgauth.ResolveFromEnv(nil, scopeRoot, pgauth.Endpoint{
		Host: host,
		Port: port,
		User: parsed.User.Username(),
	})
	if resolveErr != nil {
		// Passwordless/trust and driver-native authentication remain valid for
		// upstream bd. Only fail when a configured credential source is invalid.
		if errors.Is(resolveErr, pgauth.ErrNoPasswordResolvable) {
			return raw, pgauth.Resolved{User: parsed.User.Username()}, nil
		}
		return "", pgauth.Resolved{}, fmt.Errorf("resolving native postgres credentials: %w", resolveErr)
	}
	parsed.User = url.UserPassword(parsed.User.Username(), resolved.Password)
	if resolved.Password != "" {
		emitPostgresCredentialResolved(cityPath, scopeRoot, contract.MetadataState{
			Backend:          "postgres",
			PostgresHost:     host,
			PostgresPort:     port,
			PostgresUser:     parsed.User.Username(),
			PostgresDatabase: strings.TrimPrefix(parsed.Path, "/"),
		}, resolved.Source)
	}
	return parsed.String(), resolved, nil
}
