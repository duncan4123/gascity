# Conflict Resolution in jj

## Detecting Conflicts

```bash
jj log -r 'conflicts()'       # find all conflicted changes
jj resolve --list              # list conflicted files in @
jj resolve --list -r <id>      # list conflicted files in a specific change
```

`jj log` marks conflicted changes with `(conflict)`. `jj status` shows a warning with the list of conflicted files.

## Conflict Markers

jj's default marker style (`"diff"`) shows diffs to apply — this is hard to parse. For easier resolution, use `"snapshot"` style which shows full content of each side:

```
<<<<<<< conflict 1 of 1
+++++++ side #1
apple
grapefruit
orange
------- base
apple
grape
orange
+++++++ side #2
APPLE
GRAPE
ORANGE
>>>>>>> conflict 1 of 1 ends
```

- `+++++++` — full content of a side
- `-------` — full content of the common base
- Choose the correct content or combine sides, then remove all markers

To set snapshot style: `jj config set --user ui.conflict-marker-style "snapshot"`

If you encounter markers with diff-like `+`/`-` lines instead of full content blocks, the repo uses the default `"diff"` style. Set `"snapshot"` first for easier resolution.

## Resolving Conflicts

**Edit the file directly** — remove all conflict markers, write the correct content. jj auto-detects the resolution when it snapshots the working copy (no special command needed).

**Workflow** (either `jj edit` directly, or `jj new` + edit + `jj squash`):
```bash
# 1. Create a resolution change on top of the conflicted one
jj new <conflicted-change-id>

# 2. Read the conflicted file, edit it to remove markers and write correct content

# 3. Verify
jj diff

# 4. Squash the resolution into the conflicted change
#    Do NOT describe the resolution change — leave it empty so jj
#    preserves the original change's description automatically
jj squash
```

**Quick resolution** — keep one side entirely, applied directly to the conflicted change:
```bash
jj resolve -r <conflicted-change-id> --tool :ours <path>      # keep side #1
jj resolve -r <conflicted-change-id> --tool :theirs <path>    # keep side #2
```

In a rebase conflict: side #1 (`:ours`) is the content from the rebased commit, side #2 (`:theirs`) is the content from the destination. Examine the conflict markers to confirm which side is which before choosing.

## Resolving Conflicts in a Stack

When a rebase or insertion creates conflicts in multiple descendants, **resolve from the bottom up**. Resolving a parent often auto-resolves its descendants.

```bash
# 1. Find the earliest conflicted change
jj log -r 'conflicts()'

# 2. Resolve the earliest one (see workflow above)
jj new <earliest-conflicted>
# ... edit files ...
jj squash

# 3. Check if descendants are now resolved — squashing may auto-resolve them,
#    but can also create new conflicts in descendants. Either way, check:
jj log -r 'conflicts()'

# 4. If conflicts remain, repeat from step 2 for the next one up the stack
```

jj auto-rebases all descendants after each resolution. A single resolution at the root of the conflict often clears the entire stack.
