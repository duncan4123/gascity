---
plan_slug: version-control-plugin-backends
phase: plan-review
rig: gascity
reviewed_plan: /data/projects/pg-latest/rigs/gascity/plans/version-control-plugin-backends/implementation-plan.md
requirements_file: /data/projects/pg-latest/rigs/gascity/plans/version-control-plugin-backends/requirements.md
status: iterate
created_at: 2026-07-07T06:50:00Z
---

# Plan Review: Version Control and Backend DB Plugin Packs

## Verdict

Iterate before decomposition.

The plan is directionally sound and maps to the requirements, but several open
questions are not implementation details. They define the public config shape,
provider identity, and backend DB binding semantics for the first tasks. Those
choices need to be settled in the plan before creating implementation beads.

## Required Changes Before Decomposition

1. Resolve the public config shape.

   The plan currently leaves `[version_control]`, `[vcs]`, and capability-level
   binding keys open. This affects `internal/config`, generated schemas,
   docs, tests, and the acceptance criteria for city and rig configuration.
   Pick the canonical city and rig TOML shape, including how rig overrides
   interact with city defaults.

2. Resolve built-in VCS provider identity.

   The plan leaves `local`, `local-git`, and `local-jj` unresolved. This is
   required before implementing defaults, collision rules, provider catalog
   output, and backwards-compatible behavior for existing rigs. Choose the
   default provider names and state how capability detection is reported.

3. Resolve backend DB alias binding.

   Task 6 depends on whether pack-declared backend DB providers are selected by
   `[beads].provider = "name"` or by a new capability binding that maps onto
   beads internally. This choice controls parser changes, validation, doctor
   output, migration behavior, and tests.

4. Add generated-docs and schema source-of-truth steps.

   The plan names `docs/reference/schema/pack-schema.json`, which is generated.
   Implementation beads must edit the Go config/schema source and regenerate
   docs with `go run ./cmd/genschema`; they should not hand-edit generated
   schema files. Add `make check-docs` for docs changes.

5. Add focused proof commands.

   The Testing section gives a good strategy, but decomposition needs commands
   implementers can run for each slice. Add at least:

   - `go test ./internal/config -run 'Capability|Runtime|Provider|Compose|Schema'`
   - `go test ./internal/versioncontrol/...`
   - `go test ./cmd/gc -run 'Rig|ProviderCatalog|Doctor|Status'`
   - `go test ./internal/beads/... ./cmd/gc -run 'Beads|Provider|Catalog'`
   - `go run ./cmd/genschema`
   - `make check-docs`
   - `make dashboard-check` for any `internal/api/`, OpenAPI, dashboard, or
     generated dashboard type change

## Readiness Pass

### Requirements Traceability

- City owner imports a VCS pack: covered by tasks 1, 2, 5, and 8, but blocked
  on the public config shape and provider-selection semantics.
- Rig owner configures VCS semantics per rig: covered by tasks 1, 4, and 7,
  but needs explicit city-vs-rig override rules.
- Core code asks for version-control facts through one interface: covered by
  tasks 3, 4, and 7.
- Plugin author implements a provider outside core: covered by task 5.
- Backend DB integrations use the same packaging shape: covered by task 6, but
  blocked on backend DB alias binding.

### Task Boundaries

The eight tasks are plausible top-level slices, but tasks 2, 6, and 7 are too
large for single implementation beads.

Suggested decomposition:

- Config schema and typed structs for capability declarations.
- Registry merge, path resolution, collision errors, and diamond-import dedupe.
- Selection binding and validation for city and rig config.
- Provider catalog and doctor display for VCS providers.
- Backend DB alias resolution, separate from VCS registry work.
- VCS provider interface and local implementation.
- Exec VCS provider contract and conformance tests.
- One call-site migration bead per call-site group in task 7.

### Test Strategy

The plan identifies the right test families: config composition, schema,
provider conformance, exec fixtures, Git/Jujutsu integration, existing rig and
beads regression tests, and provider-catalog tests. It needs the concrete proof
commands listed above so each implementation bead has an exit criterion.

### Risk

- Public config keys are user-facing and hard to rename after release.
- Provider names participate in collision rules and doctor output, so they need
  stable semantics before registry work starts.
- Generated docs and schemas must be regenerated from source.
- `gc rig add`, workspace safety, status, and doctor call sites are operational
  paths; migrate them in small, reversible slices.
- API or dashboard helper changes trigger typed-wire and generated-client gates.
- Fetch is a mutating VCS operation. Protocol v0 should explicitly say whether
  `fetch` is the only mutating operation and how unsupported mutating operations
  fail closed.
