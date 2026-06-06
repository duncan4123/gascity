{{ define "city-understanding" }}
## City: doltlite-gascity

### Purpose

This city exists to develop, test, and maintain the **doltlite backend** for Gas Town and beads. It runs Gas Town's full machinery (mayor, polecats, witnesses, refinery, dogs) on a doltlite storage backend — an embedded prolly-tree SQLite engine with zero server process.

The city is the **primary development and testing ground** for running Gas Town cities without a Dolt MySQL-compatible server.

### The Ship-It Principle

A fix is only real when another user on another machine gets the same result by running the same commands — no manual steps, no ad-hoc edits, no "we already fixed that on this machine." Every change must be in code or config that ships:

- **gascity source** — patches to `cmd/gc/`, `internal/`, or `examples/*/pack/` that compile into the `gc` binary.
- **beads-doltlite source** — patches to `internal/storage/doltlite/` or `cmd/bd/` that compile into the `bd` binary.
- **doltlite C source** — patches to the prolly-tree engine that produce `libdoltlite.so`.
- **city.toml** — declarative config (`backend = "doltlite"`, pack includes, order overrides) that any `gc init` can reproduce.

Manual edits to runtime files (`.gc/system/packs/`, wrapper scripts, installed binaries) are scaffolding. They prove the fix works but are not the fix. Before declaring done, port every manual edit into its upstream source and rebuild.

### Codebases

| Repo | Location | Purpose |
|------|----------|---------|
| `gastownhall/gascity` | `./gascity/` | Gas Town controller, CLI, packs, molecules |
| `dolthub/doltlite` | `./doltlite/` | C library: prolly-tree SQLite fork (libdoltlite.so) |
| `duncan4123/beads-doltlite` | `./beads-doltlite/` | `bd` CLI: beads issue tracker with doltlite storage backend |

### Build Pipeline

```
doltlite C source              beads-doltlite Go source
  ../configure && make            GOFLAGS=-tags=libsqlite3
  doltlite-lib                    CGO_LDFLAGS=-ldoltlite
  → libdoltlite.so ──────────→  go build ./cmd/bd → bd binary
```

1. Build `libdoltlite.so` from `dolthub/doltlite` with `make doltlite-lib`
2. Build `bd` from `duncan4123/beads-doltlite` with `GOFLAGS=-tags=libsqlite3` linking to `libdoltlite.so`
3. `bd` binary provides beads CLI; Gas Town's `gc bd` commands shell out to it
4. Gas Town's `gc` binary embeds pack definitions (including the bd pack with `gc-beads-bd.sh` wrapper)

### Backend Architecture

- **Storage**: `libdoltlite.so` — embedded prolly-tree engine. Single `.db` file per database, no server process.
- **Beads CLI**: `bd` binary — links dynamically to `libdoltlite.so` via `go-sqlite3` with `libsqlite3` build tag.
- **Gas Town integration**: `gc bd` commands delegate to `bd` via `gc-beads-bd.sh` wrapper script. The wrapper detects `BEADS_BACKEND=doltlite` and routes init/operations through doltlite-specific code paths.
- **No Dolt server**: No MySQL protocol, no port, no `dolt sql-server` process. The dolt pack is conditionally skipped when backend is doltlite (see `embed_builtin_packs.go`).

### Key Differences from Dolt-Backed Cities

| Aspect | Dolt (default) | Doltlite |
|--------|---------------|----------|
| Storage | Dolt SQL server (port 37282) | Embedded prolly-tree `.db` file |
| Beads init | `DOLT_COMMIT()` via MySQL | `dolt_commit()` via SQLite built-in |
| Pack auto-install | `dolt` pack included | `dolt` pack skipped |
| Formulas | `mol-dolt-health`, `mol-dolt-remotes-patrol` | `mol-doltlite-maintenance` |
| `--dolt-auto-commit` | Controls VCS commit timing | Same flag passed but doltlite ignores VCS semantics |

### Database Layout

```
.beads/
  metadata.json        → {"backend":"doltlite","database":"doltlite","dolt_database":"hq"}
  doltlite/
    hq.db              → Single-file prolly-tree database (city-level beads)
    .lock              → flock() sentinel for exclusive write access
  routes.jsonl         → Rig prefix routing

gascity/.beads/        → Rig-level beads store (same layout, database: gc.db)
beads-doltlite/.beads/ → Rig-level beads store
```

### Rig Status

| Rig | Prefix | Repo | Role |
|-----|--------|------|------|
| gascity | `gc-` | `gastownhall/gascity` | Gas Town source, packs, CLI |
| beads-doltlite | `bd-` | `duncan4123/beads-doltlite` | Beads CLI with doltlite backend |
{{ end }}