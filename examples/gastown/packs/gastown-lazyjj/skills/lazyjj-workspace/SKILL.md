---
name: lazyjj-workspace
description: Use LazyJJ workspace and stack conventions for Gas Town jj polecat work.
category: development
allowed-tools: Bash
---

# LazyJJ Workspace Workflow

Use this skill when a Gas Town jj polecat is assigned work through
`mol-polecat-lazyjj-work`.

## Requirements

The session launcher must already place the polecat in its assigned jj
workspace. This skill does not create workspaces.

The host jj config must provide the LazyJJ revset aliases:

```bash
jj log -r branch_off
jj log -r stack_base
jj log -r stack
jj log -r no_description
```

The preferred checkpoint command is:

```bash
jj claude-checkpoint "short description"
```

If that alias is unavailable, use:

```bash
jj new -m "next"
jj describe -r @- -m "short description"
```

## Model

- The polecat's current jj workspace owns the work.
- The workspace's `@` is the polecat working head.
- A stack is inferred from jj graph ancestry, not from a separate database.
- Bead metadata only links the work bead to the workspace, stack revset, and
  review bookmark.
- Bookmarks are review/export handles. Do not create a bookmark before useful
  work exists.

## Required Bead Metadata

Record these on the work bead as soon as the formula starts:

```text
lazyjj_workspace
lazyjj_workspace_dir
```

Record these at submit time:

```text
lazyjj_review_bookmark
lazyjj_stack_revset
```

## Stack Commands

Use these for normal inspection:

```bash
jj status
jj log -r 'trunk() | branch_off | stack'
jj log -r 'stack & no_description'
jj diff --from branch_off
```

If a fix belongs in an earlier stack commit, prefer:

```bash
jj absorb
```

If exact hunk movement is required, use `jj-hunk` rather than an interactive
jj split/squash UI.

## Publish Policy

This pack is allowed to publish the formula's review bookmark when the formula
sets `publish_mode = "push"`. That is a formula-specific exception to generic
jj-vcs no-push guidance.

If `publish_mode = "handoff"`, do not push. Set the review bookmark locally,
record metadata on the bead, and mail the refinery.
