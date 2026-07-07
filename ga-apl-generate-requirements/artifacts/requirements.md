---
schema: gc.build.requirements.v1
workflow:
  id: ga-apl
  formula: build-basic
methodology:
  pack: gascity
  name: build-basic
producer:
  formula: build-basic
  stage: requirements
  attempt: 1
status: approved
trace:
  upstream:
    - path: beads/ga-apl
      hash: bead:ga-apl
      ids:
        - REQ-001
        - REQ-002
        - REQ-003
        - REQ-004
        - REQ-005
        - REQ-006
        - REQ-007
        - REQ-008
        - REQ-009
        - REQ-010
        - REQ-011
        - REQ-012
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
    - id: REQ-006
      status: covered
    - id: REQ-007
      status: covered
    - id: REQ-008
      status: covered
    - id: REQ-009
      status: covered
    - id: REQ-010
      status: covered
    - id: REQ-011
      status: covered
    - id: REQ-012
      status: covered
---

# Requirements

## Problem Statement

The first Gas City guided starter factory run needs a clear requirements artifact before implementation planning begins. The run was claimed through bead `ga-apl`, but the workflow root id, build target, and preselected requirements path were not present in bead metadata. This artifact therefore preserves the available input target as the claimed workflow bead and records the missing product target as an open question rather than blocking the autonomous run.

The factory must produce requirements that are concrete enough for plan review, validation, and downstream implementation while staying approachable for a first run.

## W6H

| Question | Answer |
| --- | --- |
| Who | The Gas City starter factory user, the requirements planner, plan reviewer, and downstream implementation worker. |
| What | Produce a schema-valid Markdown requirements artifact for the `build-basic` workflow. |
| When | During the requirements stage before implementation planning starts. |
| Where | In the claimed work directory for bead `ga-apl`. |
| Why | Downstream planning needs a stable, reviewable statement of scope, constraints, and acceptance criteria. |
| Which | Use the built-in `gascity/build-basic` methodology and preserve the claimed bead as the traceable input. |
| How | Record front matter, coverage traceability, behavior requirements, examples, acceptance criteria, non-goals, and open questions in one Markdown artifact. |

## User Stories

- REQ-001: As a starter factory user, I need the workflow target to be stated clearly so downstream agents can plan against the same goal.
- REQ-002: As a plan reviewer, I need explicit constraints and non-goals so I can identify scope drift before implementation begins.
- REQ-003: As an implementation worker, I need behavior requirements and examples that describe observable outcomes rather than hidden intent.
- REQ-004: As the workflow validator, I need schema-valid front matter and matching coverage entries so artifact validation can proceed deterministically.

## Technical Stories

- REQ-005: The artifact must be Markdown with YAML front matter using schema `gc.build.requirements.v1`.
- REQ-006: The front matter must use mapping objects for workflow, methodology, producer, and trace data.
- REQ-007: The artifact must include a trace upstream entry for bead `ga-apl` with `path: beads/ga-apl` and `hash: bead:ga-apl`.
- REQ-008: Every traced requirement id must appear exactly once in `trace.coverage` and exactly once in the Markdown coverage table with status `covered`.
- REQ-009: The claimed bead, or the workflow root if later supplied, must record the absolute requirements artifact path in `gc.build.requirements_path`.

## Behavior Requirements

- REQ-010: The requirements must support the next planning stage by making the goal, assumptions, constraints, acceptance criteria, and open questions explicit.
- REQ-011: The artifact must not invent a hidden product target when the routed bead provides none; unresolved target ambiguity must be visible in Open Questions.
- REQ-012: The artifact must remain suitable for a first guided factory run, avoiding advanced workflow details that are not required for planning the initial implementation.

## Example Mapping

| Example | Given | When | Then | Covers |
| --- | --- | --- | --- | --- |
| Valid artifact shape | A `build-basic` requirements stage is claimed | The requirements planner writes the artifact | The file has YAML front matter, required sections, and matching coverage data | REQ-004, REQ-005, REQ-006, REQ-008 |
| Traceable input | The only available input is bead `ga-apl` | The planner records upstream trace metadata | The upstream entry uses `beads/ga-apl` and `bead:ga-apl` | REQ-007 |
| Missing product target | No target metadata is available | The planner completes the autonomous run | The ambiguity is captured as an open question instead of blocking | REQ-001, REQ-011 |
| Plan-ready scope | A downstream planner reviews the artifact | They inspect requirements, non-goals, and acceptance criteria | They can decide whether to proceed or request target clarification | REQ-002, REQ-003, REQ-010, REQ-012 |

## Acceptance Criteria

- The requirements artifact is saved as Markdown with YAML front matter and no JSON wrapper.
- The front matter declares `schema: gc.build.requirements.v1`.
- `workflow`, `methodology`, `producer`, and `trace` are mapping objects, not scalar shortcuts.
- `producer.stage` is `requirements` and `producer.attempt` is a positive integer.
- `trace.upstream` contains an entry for `beads/ga-apl` with `hash: bead:ga-apl`.
- Every id listed in `trace.upstream[].ids` appears exactly once in `trace.coverage`.
- The Markdown coverage table below contains the same ID/status pairs as `trace.coverage`.
- Coverage statuses use `covered`; they do not use the artifact status `approved`.
- The claimed bead records the absolute requirements path in `gc.build.requirements_path`.
- Missing workflow-root and product-target metadata remain visible in Open Questions for plan review.

## Coverage

| ID | Status |
| --- | --- |
| REQ-001 | covered |
| REQ-002 | covered |
| REQ-003 | covered |
| REQ-004 | covered |
| REQ-005 | covered |
| REQ-006 | covered |
| REQ-007 | covered |
| REQ-008 | covered |
| REQ-009 | covered |
| REQ-010 | covered |
| REQ-011 | covered |
| REQ-012 | covered |

## Out Of Scope

- Implementing the downstream plan or modifying application code.
- Selecting a product feature target that was not present in routed workflow metadata.
- Creating additional workflow beads or changing graph routing.
- Defining a permanent schema beyond `gc.build.requirements.v1`.
- Replacing the factory validator or adding a new validation tool.

## Open Questions

- What product or code change should the starter factory build after this requirements stage?
- Should a later repair step replace workflow id `ga-apl` with a distinct workflow root id if one is attached after this artifact is created?
- Should future factory launches reject missing `gc.build.requirements_path` metadata before routing the requirements planner?
