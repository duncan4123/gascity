---
plan_slug: version-control-plugin-backends
phase: implementation-plan
rig: gascity
rig_root: /data/projects/pg-latest/rigs/gascity
artifact_root: /data/projects/pg-latest/rigs/gascity/plans
requirements_file: /data/projects/pg-latest/rigs/gascity/plans/version-control-plugin-backends/requirements.md
status: draft
created_at: 2026-07-07T06:31:19Z
updated_at: 2026-07-07T06:31:19Z
---

# Implementation Plan: Version Control and Backend DB Plugin Packs

## Summary

Add a pack-declared capability layer for version control and backend DB
providers. Version control gets a new internal provider interface with a
built-in local implementation and an optional exec plugin implementation.
Backend DBs keep their existing store APIs, but their pack declarations,
selection, doctor checks, and provider catalog entries move into the same
capability registry.

The first deliverable should be a narrow platform feature, not a full rewrite:
define the capability schema, register providers from packs, implement a VCS
provider facade over the current local Git/Jujutsu behavior, add an exec
contract for plugin authors, and migrate the highest-value call sites that
currently need repo facts.

## Current System

- `internal/git/git.go` contains a small Git wrapper used by command and
  runtime paths. It handles repo detection, branch detection, worktree lists,
  worktree removal, uncommitted-work checks, unpushed-commit checks, submodules,
  pruning, and `git fetch origin`.
- `cmd/gc/cmd_rig.go` handles `gc rig add`. It treats the positional argument
  as a local path, probes `.git`, uses `git.New(rigPath).ProbeDefaultBranch()`,
  writes rig config, and initializes rig infrastructure.
- `internal/beads/factory.go` selects bead stores from `[beads].provider`,
  including `file`, `exec:<script>`, bd-backed stores, and native Dolt
  fallbacks with diagnostics.
- `internal/beads/exec/` implements an exec-backed beads store. The reference
  docs define JSON stdin/stdout, direct exec without shell, operation names,
  exit-code semantics, and provider environment.
- `docs/reference/exec-session-provider.md` shows pack-declared runtime
  registration with `pack.toml`, collision rules, provider construction, doctor
  checks, and a protocol handshake.
- `docs/guides/shareable-packs.md` defines packs as the portable unit for
  providers, formulas, orders, commands, doctor checks, overlays, skills, and
  assets.
- `internal/config`, `internal/packman`, `internal/packregistry`, and doctor
  checks already load pack metadata, resolve imports, and validate some
  provider declarations.

## Proposed Implementation

### 1. Define capability vocabulary

Add a small capability vocabulary with two initial families:

- `version_control`
- `backend_db`

Use explicit family names in pack declarations and internal types. Avoid a
generic "plugin" catch-all in config; plugins are implementation details of a
capability.

Candidate pack shape:

```toml
[capabilities.version_control.local-jj]
provider = "exec:assets/scripts/gc-vcs-jj"
protocol = 0
description = "Jujutsu/Git provider for colocated repositories"

[capabilities.backend_db.br]
provider = "exec:assets/scripts/gc-beads-br"
protocol = 0
description = "beads_rust backend"
```

Candidate city or rig selection:

```toml
[version_control]
provider = "local-jj"
fetch_remotes = ["upstream", "origin"]
push_remote = "origin"
trunk = "main@upstream"

[beads]
provider = "br"
```

The exact key names can change during implementation, but the model should stay
clear: packs register providers, config selects providers, and core resolves a
typed provider instance.

Files likely touched:

- `internal/config/config.go`
- `internal/config/pack_runtimes.go` or a sibling provider-registration file
- `internal/config/validate_semantics_test.go`
- `docs/reference/specs/pack-spec.md`
- `docs/reference/schema/pack-schema.json`

### 2. Build a capability registry in config composition

Create an internal registry that records pack-declared providers by family and
name. Reuse the runtime registration rules where possible:

- pack-relative paths resolve relative to the declaring pack,
- bare commands resolve on PATH at provider start or doctor time,
- collisions with built-ins or other packs are composition errors,
- identical declarations through diamond imports dedupe,
- selected providers are bound at provider construction time.

Backend DB declarations should initially register aliases for existing
`exec:<script>` providers without changing the beads store interface. This
lets a pack expose a backend by name while `internal/beads/factory.go` continues
to open the selected store.

Files likely touched:

- `internal/config/pack_runtimes.go`
- `internal/config/compose_test.go`
- `internal/config/pack_test.go`
- `internal/doctor/pack_checks.go`
- `cmd/gc/doctor_provider_catalog.go`

### 3. Introduce `internal/versioncontrol`

Add a focused provider interface for repo facts and safe operations:

```go
type Provider interface {
    Info(ctx context.Context, root string) (Info, error)
    Status(ctx context.Context, root string) (Status, error)
    Remotes(ctx context.Context, root string) ([]Remote, error)
    DefaultBranch(ctx context.Context, root string) (string, error)
    Trunk(ctx context.Context, root string) (string, error)
    Fetch(ctx context.Context, root string, opts FetchOptions) (FetchResult, error)
    Workspaces(ctx context.Context, root string) ([]Workspace, error)
    Safety(ctx context.Context, root string) (Safety, error)
}
```

Keep mutating operations explicit and capability-gated. The first pass should
not expose arbitrary commit, push, or PR creation unless there is a concrete
caller and a clear safety contract.

Provider response structs should be JSON-friendly because the same structs can
serve dashboards, doctors, prompt context, and exec plugin fixtures.

Files likely added:

- `internal/versioncontrol/provider.go`
- `internal/versioncontrol/local.go`
- `internal/versioncontrol/exec/exec.go`
- `internal/versioncontrol/versioncontroltest/`

### 4. Implement the built-in local provider

Wrap today's local behavior behind the new interface:

- use `internal/git` for Git repos,
- detect `.jj/` and use `jj` only through bounded, non-interactive commands,
- support colocated Jujutsu repositories without raw Git assumptions where jj
  state is authoritative,
- keep current Git-only behavior as the default for existing rigs.

The built-in provider should understand the common configurations this rig now
uses:

- fetch remotes can include `upstream` and `origin`,
- push remote can differ from fetch remotes,
- trunk can be a branch name or jj revset such as `main@upstream`,
- default branch remains available for older Git-oriented call sites.

Files likely touched:

- `internal/git/git.go`
- `internal/git/git_test.go`
- new `internal/versioncontrol/local.go`

### 5. Define the exec VCS provider contract

Model the contract after exec beads and exec sessions:

- direct exec, no shell,
- operation name as first argument,
- JSON stdin for structured input,
- JSON stdout for structured output,
- stderr for diagnostic messages,
- per-operation timeout,
- `protocol` handshake with version and optional capabilities,
- forward-compatible unknown operation handling.

Candidate operations:

| Operation | Purpose |
| --- | --- |
| `protocol` | returns provider protocol and capabilities |
| `info` | repo kind, root, current commit/change, active workspace |
| `status` | dirty state, conflicts, untracked summary |
| `remotes` | remote names, URLs, fetch/push roles |
| `default-branch` | legacy branch answer for Git-oriented consumers |
| `trunk` | provider-native trunk target, branch or revset |
| `fetch` | fetch configured or requested remotes |
| `workspaces` | worktrees, jj workspaces, stale markers |
| `safety` | uncommitted work, unpushed work, stashes/conflicts |

Initial capabilities:

- `vcs.git`
- `vcs.jj`
- `vcs.fetch`
- `vcs.workspaces`
- `vcs.safety`
- `vcs.hosting.github`

Files likely added:

- `docs/reference/exec-version-control-provider.md`
- `internal/versioncontrol/exec/exec.go`
- `internal/versioncontrol/exec/exec_test.go`
- `contrib/version-control-scripts/` if a maintained script is included

### 6. Align backend DB provider declarations

Do not rewrite bead persistence in this pass. Instead:

- let packs declare backend DB providers by name,
- resolve those declarations into existing `[beads].provider` values,
- make `gc doctor` and provider catalog show declared backend DB providers,
- keep the current `bd`, `file`, native Dolt, and `exec:<script>` open paths,
- document how a backend DB pack ships scripts, checks, and formulas.

This keeps existing storage stable while moving the user-facing model toward
pack-backed provider selection.

Files likely touched:

- `internal/beads/factory.go`
- `cmd/gc/providers.go`
- `cmd/gc/provider_health_gate.go`
- `cmd/gc/doctor_provider_catalog.go`
- `docs/reference/exec-beads-provider.md`
- `docs/guides/shareable-packs.md`

### 7. Migrate call sites incrementally

Move callers onto `internal/versioncontrol` in small groups:

1. Rig registration:
   - keep `gc rig add <path>` as path registration,
   - add a separate clone/register flow later if needed,
   - replace default-branch probing with the selected VCS provider.
2. Worktree and workspace safety:
   - route cleanup and reaper checks through provider safety/workspace calls.
3. Prompt and formula context:
   - source default branch, trunk, remotes, and dirty state from the provider.
4. Dashboard and API helper surfaces:
   - use typed provider responses instead of ad hoc `git log` or status parsing.
5. Doctor and status:
   - report selected VCS provider, capabilities, and failed health checks.

Files likely touched:

- `cmd/gc/cmd_rig.go`
- `cmd/gc/bead_worktree_reaper.go`
- `cmd/gc/session_worktree_prune.go`
- `cmd/gc/session_work_guard.go`
- `cmd/gc/status_provider.go`
- formula or prompt-context code that currently reads Git metadata

### 8. Preserve pack-first behavior

Use a plugin only when core needs structured facts or safe operations. If a
workflow can be expressed as a command, order, formula, doctor check, or helper
script inside a pack, leave it as pack behavior.

Examples:

- A nightly fetch order can remain an order.
- A repo hygiene report can be a command or formula.
- Default branch detection used by `gc rig add` should use a provider.
- Dashboard status data should use a provider.
- Backend DB conformance should use provider catalog and doctor checks.

This boundary prevents the plugin system from becoming a second orchestration
language.

## Testing

- Unit-test config composition for pack-declared `version_control` and
  `backend_db` providers, including path resolution, collision errors, diamond
  import dedupe, and selected-provider lookup.
- Add schema tests for the new pack declaration fields.
- Add provider conformance tests for `internal/versioncontrol.Provider`.
- Add exec provider tests with fixture scripts covering success, malformed
  JSON, timeout, unknown operation, unsupported capability, and stderr
  diagnostics.
- Add integration tests with temporary Git and colocated Jujutsu repos where
  available.
- Preserve existing `cmd/gc/cmd_rig_test.go`, `internal/git/git_test.go`, and
  beads factory tests while migrating behavior.
- Add doctor/provider-catalog tests showing VCS and backend DB providers side
  by side.

## Rollout

1. Add capability registry and docs while leaving existing behavior untouched.
2. Register built-in `local-git` and `local-jj` or a single `local` provider
   with detected capabilities.
3. Add backend DB provider aliases for existing beads providers.
4. Add VCS provider interface and local implementation.
5. Add exec VCS provider contract and conformance tests.
6. Migrate `gc rig add` default-branch probing and status/doctor surfaces.
7. Migrate workspace safety and dashboard helpers.
8. Publish or bundle a first VCS pack that demonstrates commands, orders,
   doctor checks, and optional provider plugin binding.

Each step should be reversible and should preserve default behavior for cities
that do not opt into new declarations.

## Open Questions

- Should the public config key be `[version_control]`, `[vcs]`, or a more
  general `[capabilities.version_control]` binding?
- Should the first built-in provider be one auto-detecting `local` provider, or
  separate `local-git` and `local-jj` providers?
- Should backend DB provider aliases live under `[beads] provider = "name"` or
  a new capability selection key that maps onto beads internally?
- Should exec VCS plugins be one process per operation, matching beads, or may
  future providers opt into a long-running daemon?
- Which mutating VCS operations should be in protocol version 0, if any?
- How should credentials be scoped between VCS providers, git credential
  helpers, and hosting-service integrations?
