---
schema: gc.build.review.v1
workflow:
  id: gc-dt5
  formula: github-pr-review
methodology:
  pack: gstack.review
  name: implementation-reviewer report
producer:
  formula: github-pr-review
  stage: write-report
  attempt: 1
status: changes_required
review:
  kind: github_pr
  repository: gastownhall/gascity
  pr_number: 3617
  head_sha: 540ce3457b9bdeb113da08f1dafbe8a6e1d74257
  verdict: fail
  severity: major
  outcome: request_changes
findings:
  - id: GC-PR3617-001
    severity: major
    title: Deleting the committed Beads config breaks fresh checkouts
    evidence: "PR head 540ce3457b9bdeb113da08f1dafbe8a6e1d74257 deletes .beads/config.yaml, removing issue_prefix/issue-prefix, dolt.local-only/no-push/export settings, and the types.custom registry for molecule, convoy, message, event, gate, merge-request, agent, role, rig, session, spec, convergence, and step. A fresh checkout would lose the repo's Beads type/config defaults."
    required_fix: Restore .beads/config.yaml or move the same required Beads/DoltLite defaults to another committed configuration path that bd/gc loads in fresh clones, then add a regression check that a clean checkout can create/read the custom issue types used by Gas City.
  - id: GC-PR3617-002
    severity: major
    title: Pool session reuse can retain stale trigger workspace metadata
    evidence: "In the final head file cmd/gc/build_desired_state.go, bindPoolSessionTriggerBead returns early when request.WorkBeadID is empty after clearing only gc.trigger_bead_id and gc.trigger_bead_store_ref. That path does not clear gc.pack, gc.pack_workspace, gc.work_dir, or legacy work_dir. A pooled session previously bound to a pack trigger can therefore be reused for generic pool demand while still carrying the old workspace metadata."
    required_fix: When WorkBeadID is empty, clear all trigger-derived metadata keys, including pack, pack workspace, canonical work dir, and legacy work_dir. Add a regression test that reuses a session bead first bound to a pack/workspace trigger and then selected for demand without a trigger bead.
trace:
  upstream:
    - path: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/reviews/540ce3457b9bdeb113da08f1dafbe8a6e1d74257/subject.md
      hash: sha256:e7d0d0f235a5d68cdb0eb63a2d9be8e0e0fc421801ef9429115daaefcea64784
      ids:
        - subject
    - path: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/source.json
      hash: sha256:a2c451f07a296128b926feefb1a49154aa7f4d24e01cc190e291e3becdb9a06d
      ids:
        - source
    - path: git:gastownhall/gascity@540ce3457b9bdeb113da08f1dafbe8a6e1d74257
      hash: git:540ce3457b9bdeb113da08f1dafbe8a6e1d74257
      ids:
        - pr-diff
  coverage:
    - id: subject
      status: covered
    - id: source
      status: covered
    - id: pr-diff
      status: covered
---

# Review Report

PR: https://github.com/gastownhall/gascity/pull/3617
Repository: gastownhall/gascity
Head SHA: 540ce3457b9bdeb113da08f1dafbe8a6e1d74257
Subject: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/reviews/540ce3457b9bdeb113da08f1dafbe8a6e1d74257/subject.md
Snapshot: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/source.json

## Verdict

Fail, major severity. Request changes before merge.

The PR adds useful pool trigger workspace plumbing and related API/schema test coverage, but the reviewed head has two release-blocking defects. First, it deletes the committed Beads/DoltLite configuration without moving those defaults elsewhere. Second, pooled sessions can retain stale trigger workspace metadata when reused for non-trigger demand.

## Findings

### GC-PR3617-001: Deleting committed Beads config breaks fresh checkouts

Severity: major

Evidence:

- The diff deletes `.beads/config.yaml`.
- The deleted file carried repository defaults including `issue_prefix: ga`, `issue-prefix: ga`, `dolt.local-only: true`, `no-push: true`, `export.auto: false`, `dolt.disable-event-flush: true`, `backup.enabled: false`, and the `types.custom` registry for `molecule`, `convoy`, `message`, `event`, `gate`, `merge-request`, `agent`, `role`, `rig`, `session`, `spec`, `convergence`, and `step`.
- Searching the target tree at `540ce3457b9bdeb113da08f1dafbe8a6e1d74257` found no replacement occurrences for `types.custom`, `issue_prefix`, `issue-prefix`, `dolt.local-only`, `disable-event-flush`, or `backup.enabled`.

Impact:

Fresh clones lose the repo-specific Beads type/config defaults. That can break `bd`/Gas City workflows that expect those custom bead types and local-only Dolt behavior to exist without per-developer manual setup.

Required fix:

Restore `.beads/config.yaml`, or move the same required Beads/DoltLite defaults to another committed configuration path that `bd` and `gc` load in fresh checkouts. Add a regression check that a clean checkout can create and read the custom issue types Gas City uses.

### GC-PR3617-002: Pool session reuse can retain stale trigger workspace metadata

Severity: major

Evidence:

- In `cmd/gc/build_desired_state.go:2375`, `bindPoolSessionTriggerBead` handles reused pool sessions.
- Lines 2381-2400 return early when `request.WorkBeadID` is empty after clearing only `gc.trigger_bead_id` and `gc.trigger_bead_store_ref`.
- The keys added for this feature live in `internal/beadmeta/keys.go`: `gc.pack`, `gc.pack_workspace`, `gc.work_dir`, and legacy `work_dir`.
- The clearing path returns before the later metadata reconciliation for pack, pack workspace, canonical work dir, and legacy work dir can run.
- The target tests cover binding and rebinding trigger metadata, including `TestRealizePoolDesiredSessionsRebindUpdatesPackWorkspaceMetadata`, but I did not find a regression test for reusing a session that was previously trigger-bound and then selected for demand without a trigger bead.

Impact:

A pooled session previously bound to a pack trigger can later be reused for generic pool demand while still carrying the old `gc.pack`, `gc.pack_workspace`, `gc.work_dir`, or `work_dir` metadata. That can send subsequent work to the wrong workspace or preserve stale workspace hints in the session bead.

Required fix:

When `WorkBeadID` is empty, clear all trigger-derived metadata, not only the trigger bead id and store ref. The clear set should include `gc.pack`, `gc.pack_workspace`, `gc.work_dir`, and legacy `work_dir`. Add a regression test that reuses a session bead first bound to a pack/workspace trigger and then selected for demand without a trigger bead.

## Verification

Coverage:

| ID | Status |
| --- | --- |
| subject | covered |
| source | covered |
| pr-diff | covered |

Evidence reviewed:

- GitHub PR subject artifact for PR 3617.
- Snapshot metadata at `source.json`.
- Git diff from merge base `0d328a8b0413bb62c77fe7053eb27d18bd52925d` to head `540ce3457b9bdeb113da08f1dafbe8a6e1d74257`.
- Target content for `.beads/config.yaml`, `cmd/gc/build_desired_state.go`, `cmd/gc/build_desired_state_test.go`, and `internal/beadmeta/keys.go`.

Validation performed:

- Confirmed target commit exists locally.
- Confirmed `.beads/config.yaml` is absent from the target tree.
- Confirmed the deleted Beads/DoltLite defaults do not appear elsewhere in the target tree.
- Confirmed `bindPoolSessionTriggerBead` clears only two trigger keys on the empty `WorkBeadID` path.

Tests:

I did not run the repository test suite for this report. This review is based on PR diff and target-head inspection.

## Missing Evidence

- No evidence that the deleted Beads/DoltLite config is recreated by another committed bootstrap path.
- No test evidence for the stale-metadata case where a trigger-bound pooled session is later reused without `WorkBeadID`.
- No local test run evidence was produced as part of this report-only review stage.

## Recommended Fixes

1. Restore or relocate the committed Beads/DoltLite defaults, then add a clean-checkout regression check for custom bead types and local-only Dolt settings.
2. Clear all trigger-derived metadata on the no-trigger reuse path in `bindPoolSessionTriggerBead`.
3. Add a regression test that starts from a session bead with `gc.trigger_bead_id`, `gc.trigger_bead_store_ref`, `gc.pack`, `gc.pack_workspace`, `gc.work_dir`, and `work_dir`, then realizes demand without a trigger and asserts all trigger-derived values are cleared.
