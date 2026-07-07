---
schema: gc.build.decomposition.v1
workflow:
  id: ga-wzy
  formula: build-basic
methodology:
  pack: gascity
  name: build-basic
producer:
  formula: build-basic
  stage: decompose
  attempt: 1
status: approved
trace:
  upstream:
    - path: /data/projects/pg-latest/rigs/gascity/plans/version-control-plugin-backends/requirements.md
      hash: sha256:4ec1aae334583db89dad7499077d883898f8521b1e41a70d68dd6d5a62f2aa3e
      ids:
        - REQ-001
        - REQ-002
        - REQ-003
        - REQ-004
        - REQ-005
    - path: /data/projects/pg-latest/rigs/gascity/plans/version-control-plugin-backends/implementation-plan.md
      hash: sha256:7bd7d581d18d216dcc9c3b641ccce8994ec72cfccf5abecfc2ada2698cb25ab9
      ids:
        - PLAN-001
        - PLAN-002
        - PLAN-003
        - PLAN-004
        - PLAN-005
        - PLAN-006
        - PLAN-007
        - PLAN-008
  coverage:
    - id: REQ-001
      status: covered
    - id: REQ-002
      status: covered
    - id: REQ-003
      status: covered
    - id: REQ-004
      status: covered
    - id: REQ-005
      status: covered
    - id: PLAN-001
      status: covered
    - id: PLAN-002
      status: covered
    - id: PLAN-003
      status: covered
    - id: PLAN-004
      status: covered
    - id: PLAN-005
      status: covered
    - id: PLAN-006
      status: covered
    - id: PLAN-007
      status: covered
    - id: PLAN-008
      status: covered
---

## Summary

This decomposition creates one implementation convoy for adding pack-declared version-control and backend database provider capabilities. The work is split so config/provider registration, the typed VCS provider surface, exec-provider support, backend database provider alignment, call-site migration, and verification can proceed with clear traceability.

The claimed decomposition bead did not carry `gc.root_bead_id` or build artifact path metadata. This artifact uses `ga-wzy` as the workflow id because that bead owns the `artifact_root` and the approved requirements and plan artifacts for this build.

Coverage matrix:

| ID | Status |
| --- | --- |
| REQ-001 | covered |
| REQ-002 | covered |
| REQ-003 | covered |
| REQ-004 | covered |
| REQ-005 | covered |
| PLAN-001 | covered |
| PLAN-002 | covered |
| PLAN-003 | covered |
| PLAN-004 | covered |
| PLAN-005 | covered |
| PLAN-006 | covered |
| PLAN-007 | covered |
| PLAN-008 | covered |

Trace legend:

- REQ-001: City owner imports a version-control pack.
- REQ-002: Rig owner configures VCS semantics per rig.
- REQ-003: Core code asks for version-control facts through one interface.
- REQ-004: Plugin author implements a provider outside core.
- REQ-005: Backend DB integrations use the same packaging shape.
- PLAN-001: Define capability vocabulary.
- PLAN-002: Build a capability registry in config composition.
- PLAN-003: Introduce `internal/versioncontrol`.
- PLAN-004: Implement the built-in local provider.
- PLAN-005: Define the exec VCS provider contract.
- PLAN-006: Align backend DB provider declarations.
- PLAN-007: Migrate call sites incrementally.
- PLAN-008: Preserve pack-first behavior.

## Selected Downstream Formulas

- `implement`: drain implementation convoy `ga-cma`.
- `build-basic` starter review stages: review the implementation after the convoy is drained.

## Implementation Convoy

Convoy id: `ga-cma`

Convoy name: `build-basic-ga-wzy-implementation`

Work items:

- `ga-39u`
- `ga-in2`
- `ga-jkb`
- `ga-2dn`
- `ga-hsh`
- `ga-s68`

Creation and verification:

- `gc convoy create build-basic-ga-wzy-implementation ga-39u ga-in2 ga-jkb ga-2dn ga-hsh ga-s68 --json` returned convoy `ga-cma` with all six issue ids.
- `gc convoy list --json` returned convoy `ga-cma`.
- `bd dep list ga-cma ga-39u ga-in2 ga-jkb ga-2dn ga-hsh ga-s68 --json` showed `tracks` links from `ga-cma` to all six work-item beads.
- `gc convoy status ga-cma --json` reported an empty child summary even though the `tracks` links exist. Downstream implementation should treat the `tracks` links and create response as authoritative if the status summary remains stale.

## Work Items

| Work Item | Bead | Title | Trace |
| --- | --- | --- | --- |
| WI-001 | `ga-39u` | Add pack capability declarations and provider registry | REQ-001, REQ-002, REQ-005; PLAN-001, PLAN-002 |
| WI-002 | `ga-in2` | Introduce versioncontrol provider interface and built-in local provider | REQ-002, REQ-003; PLAN-003, PLAN-004 |
| WI-003 | `ga-jkb` | Add exec version-control provider contract and conformance tests | REQ-004; PLAN-005 |
| WI-004 | `ga-2dn` | Align backend DB provider declarations with provider catalog and doctor checks | REQ-005; PLAN-006 |
| WI-005 | `ga-hsh` | Migrate initial VCS call sites to the provider surface | REQ-002, REQ-003; PLAN-007, PLAN-008 |
| WI-006 | `ga-s68` | Add build-level verification and rollout documentation | REQ-001, REQ-002, REQ-003, REQ-004, REQ-005; PLAN-001, PLAN-002, PLAN-003, PLAN-004, PLAN-005, PLAN-006, PLAN-007, PLAN-008 |

### WI-001: Add pack capability declarations and provider registry

Add explicit `version_control` and `backend_db` capability families in pack declarations and config composition. The work should parse pack declarations, resolve city or rig provider selections, preserve default behavior when no provider is selected, and cover collision and diamond-import behavior with tests. It should update pack docs and schema artifacts alongside the config changes.

### WI-002: Introduce versioncontrol provider interface and built-in local provider

Add `internal/versioncontrol` with typed provider operations for repo facts and safe operations, then wrap current local Git and colocated Jujutsu behavior behind that interface. The default local path must preserve current Git-only behavior for existing rigs, while typed errors identify the root or rig, selected provider, operation, and failing command where applicable.

### WI-003: Add exec version-control provider contract and conformance tests

Implement the external VCS provider contract as direct exec with operation name as the first argument, JSON stdin/stdout, stderr diagnostics, per-operation timeout, a protocol handshake, optional capability reporting, and forward-compatible unsupported-operation behavior. Add reference documentation and fixture-backed tests for success and failure modes.

### WI-004: Align backend DB provider declarations with provider catalog and doctor checks

Allow packs to declare backend database providers by name and resolve those declarations into existing beads provider values without rewriting persistence. The provider catalog, doctor checks, health reporting, and collision rules should show backend database providers beside version-control providers while keeping existing `bd`, `file`, native Dolt, and `exec:<script>` open paths stable.

### WI-005: Migrate initial VCS call sites to the provider surface

Move the first repo-fact and safety-sensitive callers onto `internal/versioncontrol`: rig registration default-branch probing, status and doctor surfaces, prompt or formula context, and workspace safety checks. Preserve pack-first behavior by leaving commands, orders, formulas, doctor checks, and helper scripts in packs when they do not require structured core facts.

### WI-006: Add build-level verification and rollout documentation

Add tests and docs that make the provider-pack work shippable. This includes config and schema tests for capability declarations, local and exec provider tests, doctor/provider-catalog coverage for both provider families, integration tests with temporary Git and colocated Jujutsu repositories where available, user-facing docs for provider selection, and a reversible rollout path that preserves behavior for cities with no provider declarations.
