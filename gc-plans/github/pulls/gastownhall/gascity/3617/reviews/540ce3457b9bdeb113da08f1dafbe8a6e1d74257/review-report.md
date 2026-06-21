---
schema: gc.verdict-report.v1
kind: review
verdict: fail
severity: major
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
---

# Review Report

PR: https://github.com/gastownhall/gascity/pull/3617
Repo: gastownhall/gascity
Head SHA: 540ce3457b9bdeb113da08f1dafbe8a6e1d74257
Subject: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/reviews/540ce3457b9bdeb113da08f1dafbe8a6e1d74257/subject.md
Snapshot: /data/projects/doltlite-gascity/gascity/gc-plans/github/pulls/gastownhall/gascity/3617/source.json

## Verdict

Request changes. The main feature direction is useful, and the PR adds relevant desired-state and API/schema test coverage, but the current head has two release-blocking risks. The repository-level Beads config deletion can break fresh checkouts and issue type handling, and pooled sessions can keep stale workspace metadata when moving from trigger-bound demand back to generic pool demand.

## Evidence Reviewed

- GitHub PR metadata for PR 3617, updated 2026-06-21T04:46:28Z.
- PR patch at head 540ce3457b9bdeb113da08f1dafbe8a6e1d74257.
- Final head content for cmd/gc/build_desired_state.go at 540ce3457b9bdeb113da08f1dafbe8a6e1d74257.

## Tests

I did not run the repository test suite in this report path. This review is based on the PR diff/head inspection and validator-backed report generation.
