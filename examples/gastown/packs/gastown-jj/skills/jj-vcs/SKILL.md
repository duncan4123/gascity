---
name: jj-vcs
description: >
  Enforce Jujutsu (jj) workflow and commit message policy.
  Required before writing any commit message. Covers jj commands,
  commit message format rules, workflow patterns, and error handling.
  Triggers on: version control, commits, history, jj commands,
  .jj directory present, git operations, git commands, making changes to code,
  describe, commit message, rename change, rename commit, reword.
category: development
allowed-tools: Bash
---

# jj — Jujutsu Version Control

## Intention

**You must track your work in version control.** The user needs a log of changes to review your progress, understand your decisions, and evaluate different approaches. Each logical step should be a separate commit. This is not optional — it's how you demonstrate your work.

## Overview

jj (Jujutsu) is a Git-compatible version control system. Works on top of Git repositories with simpler mental model.

**In this project, use jj instead of git.** If you think about using git commands — use jj equivalents instead.

**Key principles:**
- Working copy is always a commit — no staging area
- Each step starts with `jj new`, describe when done with `jj describe`
- Never destroy work — branch instead of undo
- Never use git commands in jj repo
- Suppress pagers and ANSI colors: prefix jj commands with `--no-pager --color=never` to keep output clean for agent consumption

## Initialization

If project has `.git` but no `.jj`:
```bash
jj git init --colocate
```

## Core Workflow

### 0. Check the current change
```bash
jj st
```
If `@` is already `(empty) (no description set)` then start editing directly (step 2). Otherwise create a new change with `jj new` (step 1).

### 1. Start new work
```bash
jj new                      # create empty commit on top of current
jj new -m "brief intent"    # with initial description (optional)
```

### 2. Make changes
Edit files normally. Changes automatically tracked in current commit.

### 3. Describe when done
```bash
jj describe -m "🦺 Add input validation for email field"
```

### 4. Repeat
```bash
jj new                      # start next piece of work
```

## Commit Message Format

Based on [cbea.ms/git-commit](https://cbea.ms/git-commit/). The diff shows *what* changed; the message explains *why*.

### Subject line (required)

- ≤ 50 characters (72 absolute max — GitHub truncates beyond that)
- Capitalize first word, no trailing period
- **Imperative mood**: "Add feature" not "Added feature" or "Adding feature"
- Test: the sentence "If applied, this commit will **[subject]**" must read naturally
- If summarizing is hard, the commit probably contains too many changes — split it
- Subject describes the *outcome*, not the implementation — no flags, file names, or CLI arguments; save those for the body
- A reader unfamiliar with the codebase should understand the intent from the subject alone — no jargon, tool names, or acronyms without context

### Body (usually unnecessary)

Most commits need **only a subject line**. Add a body only when:
- The *reason* for the change isn't obvious from the diff
- You chose one approach over another and want to record *why*
- The commit marks a failed attempt (⚗️) and you need to explain what went wrong
- The subject references a tool, library, or concept that an unfamiliar reader wouldn't recognize — add a link or brief explanation in the body

When you do write a body:
- Separate from subject with a blank line
- Wrap at 72 characters
- Explain **what problem existed** and **why this solution** — not how the code works
- Bullet points (with `-`) are fine

### Gitmoji

Prefix the subject line with a [gitmoji](https://gitmoji.dev/) that describes the nature of the change. The emoji goes before the text, separated by a space. The 50-char limit applies to the full subject including the emoji.

```bash
# Examples:
jj describe -m "✨ Add email format validation"
jj describe -m "🐛 Fix login timeout on slow connections"
jj describe -m "♻️ Extract helper function for validation"
```

**Full reference**: see [gitmoji-reference.md](gitmoji-reference.md) in this skill directory. Consult it when choosing the emoji — don't guess.

### Anti-patterns

```bash
# WRONG — restating the diff as a changelog:
jj describe -m "Add French locale translation

Add full FTL translation for French (fr, config ID 4).
Validated: Fluent syntax clean, correct encoding, no missing keys.
Config: added to default.json and checkFluentFiles.ts.
Switcher keys added to all existing locale files."

# RIGHT — subject says it all:
jj describe -m "🌐 Add French locale translation"
```

```bash
# WRONG — body explains How instead of Why:
jj describe -m "Fix login timeout

Set socket timeout to 30s in HttpClient config.
Added retry logic with exponential backoff."

# RIGHT — body explains Why:
jj describe -m "🐛 Fix login timeout

Users on slow connections were getting dropped after the
default 5s timeout. 30s matches the load balancer ceiling."
```

```bash
# WRONG — non-imperative, has trailing period:
jj describe -m "Fixed the bug with validation."

# WRONG — too vague:
jj describe -m "Update code"

# RIGHT:
jj describe -m "🦺 Validate email format on signup form"
```

```bash
# WRONG — comma joins two changes, unfamiliar name unexplained:
jj describe -m "✨ Add Roogle search, update config parser"

# RIGHT — single intent, unfamiliar tool explained in body:
jj describe -m "✨ Add full-text search via Roogle

Roogle (https://example.com/roogle) indexes Rust docs
for offline search. Replaces the grep-based approach."
```

## Checking Status and History

```bash
jj st                       # current status
jj log                      # view history
jj diff                     # see current changes
```

### Reading `jj log` output
```
@  qvlmxkoz user 2025-12-06 15:30
│  (empty) (no description set)
○  pzlwrkmv user 2025-12-06 15:25
│  Add input validation
○  rstqwxyz user 2025-12-06 15:20
│  Extract helper function
◆  zzzzzzzz root()
```

- `@` — current position (working copy)
- `qvlmxkoz` — change ID (stable, use for references)
- `(empty)` — commit has no file changes
- `(no description set)` — needs `jj describe`

## Finding Where to Continue

1. Run `jj log` to see history
2. Find last meaningful commit (has description, no ⚗️ marker)
3. Start from there: `jj new <change-id>`

## Granularity

- Each change should be **atomic** — one logical unit
- **Separate refactoring from functional changes**
- Not too small, not too large

Good:
```
1. "♻️ Extract helper function for validation"  ← refactoring
2. "🦺 Add email format validation"              ← validation
```

## Preserving Work History

**The core rule: never overwrite your own failed attempt with a new approach.**

When you try something and it doesn't work, the previous attempt has diagnostic value — the user may find insights, partial solutions, or useful context in what you tried. Overwriting it destroys that evidence.

### Who initiated the change?

| Initiator | Scenario | What to do |
|-----------|----------|------------|
| **User** asked to modify a change | Feedback, correction, "change X to Y" | `jj edit` — user sees what's happening and directed it |
| **Agent** fixes a trivial mistake | Forgot an import, typo, obvious omission | `jj edit` — no diagnostic value in the broken version |
| **Agent** changes strategy | Approach A failed, switching to B | **Branch.** Describe A with ⚗️, `jj new` from before A |

**The litmus test**: before rewriting a change you made, ask — *would comparing the old version with the new one be useful to the user?* If possibly yes — branch. Extra history is cheap; lost work is not.

**Don't rationalize.** "I'm just improving my previous approach" when you're actually replacing the strategy is the exact failure mode this rule prevents. The words "try a different approach", "start over", "rethink" in your own reasoning are the signal. When in doubt, branch.

### When an approach fails

1. Describe what went wrong:
```bash
jj describe -m "⚗️ Add Redis caching

Connection failed in tests — config not available
in CI environment. Need different approach."
```

2. Go back and start fresh branch:
```bash
jj new <change-id-before-attempt>
```

Result — all attempts preserved:
```
A → B → C (failed attempt with ⚗️, preserved)
      ↘ D → E (new attempt from B)
```

## Splitting and Moving Changes

Sometimes a commit needs to be reorganized after the fact:
- It mixes unrelated changes (refactoring + feature)
- It's too large and should be broken into atomic units
- Some changes belong in a different commit

### File-Level Operations

**Split a commit by files** — changes to listed files go into the first commit, everything else into the second:
```bash
jj split -r <change-id> path/to/file1 path/to/file2 -m "First part description"
# Run jj log to find the second commit's change ID
jj log
# The second commit keeps the original description; update if needed
jj describe -r <second-change-id> -m "Second part description"
```

**Move files between commits**:
```bash
jj squash --from <source> --into <destination> path/to/file -u
# -u keeps destination's description unchanged
```

When both source and destination already have descriptions, omitting `-u` opens an editor to merge the two messages — which **hangs a headless/no-TTY agent session**. Always pass `-u` unless you deliberately want to set a new combined message with `-m`.

### Squash and `-m`

`jj squash` without `-m` preserves the parent's existing description. **Only use `-m` when setting a new description** (e.g., parent had no description). When squashing a fix into an already-described change, always omit `-m` — otherwise you silently overwrite the user's message.

### Hunk-Level Operations (jjc)

For splitting, dropping, or moving individual hunks between commits, use `jjc` — a non-interactive, scriptable tool for hunk-level operations.

```bash
# 1. List hunks in a revision
jjc hunks                          # hunks in @
jjc hunks -r <change-id>           # hunks in a specific revision
jjc hunks --full                   # with context lines and line numbers

# 2. Pick hunks into a new commit (like jj split, but non-interactive)
jjc pick a1b -m "♻️ Extract helper"             # whole hunk by ID prefix
jjc pick a1b#2 -m "🐛 Fix off-by-one"           # single change atom
jjc pick a1b@15-20 -m "♻️ Refactor validation"  # by target line range
jjc pick a1b c3d -m "msg"                        # multiple hunks

# 3. Drop hunks (revert to parent content)
jjc drop a1b

# 4. Fold hunks into another revision
jjc fold a1b --into <target-rev>
```

If `jjc` is not available, check [line-level-split-fallback.md](line-level-split-fallback.md) — it has installation instructions (ask the user first) and a manual `jj split --tool` fallback.

### Moving Hunk-Level Changes Between Commits

To move specific hunks from commit X into commit Y:
```bash
# With jjc — one step:
jjc fold a1b --into <target-rev>

# Without jjc — two steps:
# 1. Split to isolate the hunks (see fallback doc)
# 2. jj squash --from <split-result> --into <target> -u
```

**New work that belongs in an earlier commit**: make the edit at the top of the stack (`@`), then move the hunk down with `jjc fold <hunk> --into <target>`. Don't `jj edit <earlier-commit>` to write it there directly — early commits often have an incomplete `.gitignore` (build dirs not yet ignored), so editing there can pull build artifacts into the commit. The top of the stack has the full `.gitignore` and avoids the trap.

### Reordering Changes

To move a change to a different position in history:
```bash
# Move change X to come right before change Y (Y becomes child of X)
jj rebase -r <change-id> --before <target-id>

# Move change X to come right after change Y (X becomes child of Y)
jj rebase -r <change-id> --after <target-id>
```

With `-r`, only the specified change moves — its descendants are rebased onto its former parent, filling the "hole".

### Linter and Formatter Results

Formatter/lint fixes are **not separate logical work** — they belong in the change that introduced the code, not in a commit of their own.

**Single change** — fixes are in `@`, original work in parent:
```bash
jj squash    # folds @ into parent, keeps parent's description
```

**Across a stack of changes** — use `jj fix`, not a manual per-commit pass. It runs the configured tool over a revision and all its descendants parent-first, recreating each commit with the fix applied. It **never creates new conflicts**: each descendant gets the same tool applied to its own version of the file.

```bash
# Configure the tool once (repo or user config). The tool reads stdin,
# writes stdout; $path expands to the file's repo path:
#   [fix.tools.rustfmt]
#   command  = ["rustfmt", "--emit=stdout"]
#   patterns = ["glob:'**/*.rs'"]

jj fix -s <earliest-change-id>   # fix this rev + all its descendants
jj fix                           # default revset: reachable(@, mutable())
jj op show -p                    # review what fix changed
```

**Gotchas:**
- Tools must read **stdin → stdout**. Many formatters default to in-place editing — pass the stdout flag explicitly (`rustfmt --emit=stdout`, `black -`, `prettier --stdin-filepath $path`). The real file path is *not* on disk for the tool.
- Pass the config path explicitly if the tool needs it — it can't infer project config from a stdin stream.
- Output must be **deterministic**: jj dedupes identical file content across commits and runs the tool only once.

Editing deep history is cheaper than it looks — cost scales with *line overlap*, not the number of descendants. jj rebases descendants automatically; you only do manual work where a descendant actually touches the same lines `jj fix` rewrote.

If no formatter is configured for `jj fix`, fall back to a manual change **before** your work stack (`jj new <parent-of-A>`, run the tool, describe), then resolve any conflicts in the rebased A'/B'/C' bottom-up.

### Safety

**Before** any split or move:
```bash
jj op log -n 1    # record current operation state
```

**After** any split or move — verify both halves:
```bash
jj log                          # check commit graph looks correct
jj diff -r <first-change-id>    # verify first commit has intended changes
jj diff -r <second-change-id>   # verify second commit has the rest
```

**Recovery** — if something went wrong:
```bash
jj undo    # reverts the last jj operation cleanly
```

This is the **one case** where `jj undo` is permitted. See Anti-patterns below.

## Conflicts

In jj, conflicts are **not blocking**. Operations like rebase, squash, and split always succeed — if conflicts arise, they are recorded inside the commit. You can defer resolution and keep working.

For detecting, reading markers, and resolving conflicts (including stacks), see [conflicts.md](conflicts.md).

## Anti-patterns

### Never use git commands
```bash
# WRONG — don't use in jj repo:
git add, git commit, git status
git checkout, git branch
git log

# RIGHT — use jj equivalents:
jj st, jj log, jj new, jj describe
```

### Never destroy work
```bash
# WRONG — don't use for regular workflow:
jj undo        # reverses last operation, loses recent work
jj abandon     # discards a change

# WRONG — agent rewrites its own failed approach:
jj edit <change-with-approach-A>
# ... replaces with approach B — A is now lost

# RIGHT — branch and preserve:
jj describe -m "⚗️ ..."    # mark the failed attempt
jj new <change-id-before>   # branch from earlier point
```

See **Preserving Work History** for the full rule — `jj edit` is fine when the user requested the change or the agent fixes a trivial mistake.

**Exceptions**:
- `jj undo` is permitted to recover from a failed `jj split` or `jj squash --from/--into` operation. See Safety under "Splitting and Moving Changes".
- `jj abandon` is the correct tool for removing an orphan empty commit — a change that is both `(empty)` and `(no description set)` with no descendants.

### Never leave commits undescribed
```bash
# WRONG:
jj new
# ... work ...
jj new          # forgot to describe!

# RIGHT:
jj new
# ... work ...
jj describe -m "What was done"
jj new          # now start next
```

## Quick Reference

| Task | Command |
|------|---------|
| Initialize | `jj git init --colocate` |
| Status | `jj st` |
| History | `jj log` |
| Diff | `jj diff` |
| Start new work | `jj new` |
| Describe commit | `jj describe -m "message"` |
| Continue from point | `jj new <change-id>` |
| Modify existing change | `jj edit <change-id>` (see Preserving Work History) |
| Split by files | `jj split -r X <paths> -m "msg"` |
| Pick hunks | `jjc pick <hunk> -m "msg"` |
| Drop hunks | `jjc drop <hunk>` |
| Fold hunks | `jjc fold <hunk> --into <rev>` |
| Squash into parent | `jj squash` (keeps parent's description) |
| Move between commits | `jj squash --from X --into Y <paths> -u` |
| Reorder change | `jj rebase -r X --before Y` / `--after Y` |
| Find conflicts | `jj log -r 'conflicts()'` |
| List conflicted files | `jj resolve --list` |
| Keep one side | `jj resolve --tool :ours <path>` |
| Operation history | `jj op log -n 5` |
| Undo last operation | `jj undo` (split/move recovery only) |

## Sources

- [jj Documentation](https://docs.jj-vcs.dev/latest/)
- [Git command table](https://docs.jj-vcs.dev/latest/git-command-table/)
- [Commit message guide](https://cbea.ms/git-commit/)

## License

EUPL 1.2 — see [LICENSE](LICENSE).
