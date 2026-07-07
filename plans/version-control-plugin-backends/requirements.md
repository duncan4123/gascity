---
plan_slug: version-control-plugin-backends
phase: requirements
rig: gascity
rig_root: /data/projects/pg-latest/rigs/gascity
artifact_root: /data/projects/pg-latest/rigs/gascity/plans
status: draft
created_at: 2026-07-07T06:31:19Z
updated_at: 2026-07-07T06:31:19Z
---

# Requirements: Version Control and Backend DB Plugin Packs

## Problem Statement

Gas City already treats several runtime surfaces as pluggable: beads can use
`bd`, file, native Dolt, or `exec:<script>` stores; sessions, mail, and events
also have provider seams. Version-control behavior is not modeled the same way.
Repo inspection, default branch detection, fetches, worktree checks, prompt
metadata, and dashboard helpers are spread across direct Git helpers and command
code, with no capability boundary for Git, Jujutsu, GitHub, or future providers.

The same mismatch exists at the packaging layer. Packs can ship agents,
formulas, orders, commands, doctor checks, runtimes, and helper scripts, but
there is no first-class way for a pack to declare "this is the version-control
integration" or "this is the backend database integration" and have Gas City
bind, validate, and expose it consistently. Users end up debugging local config
and command behavior instead of selecting a pack-backed capability.

## Solution

Introduce a pack-declared capability model for version control and backend DBs.
A pack should be able to provide:

- declarative capability metadata in `pack.toml`,
- commands, orders, formulas, doctor checks, and scripts that implement the
  capability,
- an optional provider plugin executable when core code needs structured
  operations instead of just running pack orders.

Version control should become a provider-backed capability with a built-in local
provider for today's Git/Jujutsu behavior and an exec-style plugin contract for
non-built-in implementations. Backend DBs should keep their existing store
interfaces, but their provider declarations, health checks, conformance checks,
and pack registration should move toward the same capability registry instead
of one-off config paths.

The intended user model is: import a pack that handles the capability, select it
in city or rig config, and let Gas City route core operations through the
selected provider only when it needs a structured API.

## User Stories

### City owner imports a version-control pack

Acceptance criteria:

- A city can import a VCS pack through normal pack imports.
- The pack can register commands, orders, formulas, and doctor checks related to
  fetch, sync, workspace hygiene, review setup, and remote policy.
- City or rig config can select the pack-declared VCS provider without relying
  on PATH-only conventions.
- `gc doctor` reports whether the selected VCS capability is installed,
  reachable, and compatible.

### Rig owner configures VCS semantics per rig

Acceptance criteria:

- A rig can declare provider choice and repository semantics such as default
  branch, trunk revset, fetch remotes, push remote, and hosting service.
- Git-only and colocated Jujutsu repositories both have a supported default
  path.
- Existing rigs without explicit VCS config continue to use the built-in local
  provider and current default-branch behavior.
- Error messages identify the rig, selected provider, operation, and failing
  command or plugin response.

### Core code asks for version-control facts through one interface

Acceptance criteria:

- Core operations that need repo facts do not shell directly to `git` or `jj`
  when a provider interface is available.
- Initial operations include repo detection, status, default branch/trunk,
  remotes, fetch, unpushed-work checks, and worktree or workspace health.
- Provider responses are typed enough for prompts, dashboards, doctors, and
  formulas to consume without parsing human CLI output.
- Unsupported optional operations fail closed where safety matters and degrade
  cleanly where read-only features are optional.

### Plugin author implements a provider outside core

Acceptance criteria:

- An external executable can implement the VCS provider contract with JSON
  stdin/stdout and a protocol handshake.
- The handshake declares protocol version and optional capabilities.
- Unknown future operations have a forward-compatible response path.
- A conformance command or doctor check can validate the executable without
  requiring a full city.

### Backend DB integrations use the same packaging shape

Acceptance criteria:

- Existing beads backend selections keep working.
- A pack can declare a backend DB provider binding and the executable or
  connection requirements needed to run it.
- Backend DB providers participate in the same provider catalog, doctor checks,
  health reporting, and collision rules as VCS providers.
- The initial implementation does not require rewriting all existing store code
  before users can benefit from pack-declared backends.

## Out Of Scope

- Replacing Git, Jujutsu, Dolt, bd, SQLite, or Postgres implementations.
- Building a full GitHub or forge workflow engine in the first pass.
- Creating beads or running implementation workflows from this plan.
- Migrating every existing direct Git call before the first provider contract
  lands.
- Making pack-only orders into plugins when a pack command or order is enough.

## Other Notes

- The existing exec provider style for beads and sessions is the closest
  precedent: direct exec, JSON payloads, protocol handshake where needed, and
  doctor/conformance coverage.
- The existing `gc rig add` path confusion came from treating every argument as
  a path. This plan should not special-case one URL shape there; repository
  creation and registration should be expressed through the selected VCS
  capability.
- Pack docs already describe packs as the portable unit for providers,
  formulas, orders, commands, doctor checks, overlays, skills, and assets. This
  work should extend that model, not create a parallel plugin format.
- Provider names declared by packs must use the existing collision discipline:
  no silent shadowing of built-ins or other packs.
