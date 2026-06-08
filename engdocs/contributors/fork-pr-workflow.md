---
title: Fork PR Workflow
description: How this integration fork stays clean while preparing upstream pull requests.
---

# Fork PR Workflow

This runbook is for the DoltLite/Gas City integration fork. Its purpose is to
keep local jj work, fork staging PRs, and upstream PRs separate enough that the
fork can keep tracking `origin/main` without losing useful integration work.

## Terms

- `origin` is the upstream repository: `gastownhall/gascity`.
- `fork` is the working fork used for publishing review branches:
  `duncan4123/gascity`.
- `main@origin` is the upstream base. In this checkout, `trunk()` resolves to
  `main@origin`.
- Local `main` should mirror `main@origin`. Do not use local `main` for feature
  work.
- A `jj` change ID is the durable handle for local work. A bookmark is only the
  Git/GitHub publication handle.

## Policy

1. Keep the default workspace on top of `main@origin`.
2. Keep each change focused on one PR-sized behavior.
3. Prefer independent PRs. Use a stacked PR only when the upper change cannot
   build or make sense without the lower change.
4. Publish from the fork; target upstream only when the change is ready for
   upstream review.
5. Track every open review or cleanup item with a bead.
6. Keep `conflicts()` empty for active work. If a conflicted side revision is
   intentionally retained, it needs a bead that says why.

## Current Checkout Shape

The repo config currently fetches both upstream and the fork:

```bash
jj config list | grep '^git\.'
```

Expected values:

```text
git.fetch = ["origin", "fork"]
git.push = "fork"
```

The SPR configuration currently publishes to the fork repository:

```text
spr.githubRepository = "duncan4123/gascity"
spr.githubRemoteName = "fork"
spr.githubMasterBranch = "main"
spr.branchPrefix = "jj-spr/duncan4123/"
```

That means `jj spr diff` creates fork-local review PRs, not upstream PRs
against `gastownhall/gascity`. Treat those PRs as a staging queue unless the
configuration is deliberately changed.

## Daily Start

Run:

```bash
jj git fetch
jj log -r 'trunk() | main@origin | main'
jj status
jj log -r 'conflicts()'
```

Expected state:

- `trunk()` and `main@origin` are the same change.
- local `main` is the same change as `main@origin`.
- the default workspace `@` is either empty or a single focused change on top
  of `main@origin`.
- `conflicts()` is empty for active work.

If `conflicts()` is not empty, classify those changes before doing feature
work:

- active PR work: resolve it now
- obsolete experiment: abandon it
- backup/archive state: attach a bead and leave it out of active revsets

## Starting Work

Start a new independent change from upstream main:

```bash
jj new main@origin -m "wip: short topic"
```

Edit files, then give the change its final message:

```bash
jj describe -m "fix(scope): short imperative summary"
```

If the working copy now contains mixed concerns, split before publishing. Use
`jj-hunk` for non-interactive hunk selection:

```bash
jj-hunk list
jj-hunk split '{"files": {"path/to/file": {"hunks": [0]}}, "default": "reset"}' \
  "fix(scope): first logical change"
```

After a review change is ready, create an empty workspace on top of it:

```bash
jj new -m "wip: next"
```

The common review shape is then:

- review change at `@-`
- empty working change at `@`

## Fork-Local Review Queue

Use fork-local PRs when the change is still part of integration staging,
DoltLite validation, or local CI sequencing.

Publish an independent fork-local review:

```bash
jj spr diff --cherry-pick
```

List the queue:

```bash
jj spr list
```

Land independent fork reviews only when that is the intended target:

```bash
jj spr land --cherry-pick -r <change-id>
```

The fork-local queue must not be mistaken for upstream readiness. A clean
fork-local PR is evidence, not the upstream PR itself.

## Upstream PR Lane

Use upstream PRs when a change is ready for `gastownhall/gascity`.

1. Confirm the change is based on current upstream main:

   ```bash
   jj git fetch
   jj log -r 'main@origin | @-'
   ```

2. Give the change an upstream PR bookmark:

   ```bash
   jj bookmark set upstream/<short-topic> -r <change-id>
   jj git push --remote fork --allow-new -b upstream/<short-topic>
   ```

3. Open the upstream PR explicitly:

   ```bash
   gh pr create \
     --repo gastownhall/gascity \
     --base main \
     --head duncan4123:upstream/<short-topic> \
     --title "<title>" \
     --body-file <body-file>
   ```

Do not rely on detached-HEAD branch detection. Always pass `--repo`,
`--base`, and `--head`.

## Queue Organization

Each review gets a bead with:

- target: `fork` or `origin`
- change ID
- bookmark
- PR URL, once published
- dependency relationship, if any
- required quality gates

Use this triage order:

1. Upstream-blocking PRs
2. Fork-local PRs that unblock upstream PRs
3. Active local changes on `main@origin`
4. Conflicted side revisions
5. Archived backup bookmarks

For dependent stacks, land bottom-up. For independent changes, use
`--cherry-pick` and keep them reorderable.

## Cleanup Rules

- Do not fix conflicts by mass-rebasing every bookmark. Classify first.
- Do not keep unnamed conflicted revisions. Describe, bead, resolve, or
  abandon them.
- Do not delete backup bookmarks until their useful content has been mapped to
  a PR, bead, or archive note.
- Do not mix workflow cleanup into feature changes. Use a separate jj change
  from `main@origin`.
- Keep fork-specific behavior behind adapter, config, pack, or provider
  boundaries so upstream patches stay reviewable.

## Quality Gates

For code changes:

```bash
make test
go vet ./...
```

Use the sharded targets in `TESTING.md` for broader sweeps instead of a
monolithic `go test ./...`.

For docs-only workflow changes:

```bash
make check-docs
```

For dashboard or API schema changes, also run `make dashboard-check` and smoke
the built dashboard locally as documented in `AGENTS.md`.

## End-of-Session Checklist

1. `jj status` shows the expected single-purpose change.
2. `jj log -r 'trunk()..@'` shows only the intended active stack.
3. `jj log -r 'conflicts()'` has no active-work conflicts.
4. Every remaining cleanup item has a bead.
5. Published reviews have PR URLs recorded on their beads.
6. Changes are pushed to the correct remote.
