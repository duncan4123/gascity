# Line-Level Split Fallback

This document covers two things:
1. Installing `jjc` (the preferred tool for hunk-level operations)
2. Manual `jj split --tool` approach (when jjc is unavailable)

## Installing jjc

**Before installing, ask the user.** Don't install tools without consent. Explain what jjc gives you:

> jjc is a non-interactive hunk-level tool for jj. It lets me split, drop, and move individual hunks by ID — without writing temporary scripts. Without it I can still do line-level splits, but the process is more manual and error-prone (see below). Want me to install it?

If the user declines, skip to the manual approach below.

### Prerequisites

- **jj v0.41.0+** — check with `jj --version`
- **Rust toolchain** — check with `cargo --version`

If `cargo` is not found, the user needs to install the Rust toolchain first. **This is a separate decision — ask the user how they prefer to install Rust** (rustup, Homebrew, Nix, system package manager, etc.). Point them to https://rustup.rs if they have no preference.

### Install

```bash
cargo install --git https://tangled.sh/akashina.tngl.sh/jjc
```

Compiles from source — may take a few minutes on first run.

After installation, verify: `jjc --help`

Project page: https://tangled.org/akashina.tngl.sh/jjc

## Manual Line-Level Split (jj split --tool)

For splitting changes **within a single file** — keeping some hunks in one commit and moving the rest to another — use a custom diff editor script.

### Protocol

jj passes two directories to the script:
- `$1` (`left`) — parent commit state (read-only)
- `$2` (`right`) — current commit state (writable)

Modify files in `$2` to contain only the changes wanted in the **first** commit. Whatever you revert toward `$1` goes into the **second** commit.

### Step-by-step

1. Inspect the changes:
```bash
jj diff -r <change-id>
```

2. Decide which hunks belong in commit 1 vs commit 2.

3. Write a temporary tool script:
```bash
#!/bin/bash
left="$1"    # parent state (read-only)
right="$2"   # current state (writable) — edit this

# To REVERT a file entirely (move ALL its changes to commit 2):
cp "$left/path/to/file.rs" "$right/path/to/file.rs"

# To KEEP a file entirely in commit 1:
# (do nothing — leave $right as-is)

# To keep SOME lines in commit 1 (line-level split):
# Write the desired partial state for commit 1:
cat > "$right/path/to/file.rs" << 'FILECONTENT'
... file content with only the changes wanted in commit 1 ...
FILECONTENT
```

4. Run the split (the script must exit 0 — non-zero aborts the split):
```bash
chmod +x /tmp/jj-split-tool.sh
jj split -r <change-id> --tool /tmp/jj-split-tool.sh -m "First part"
```

5. Clean up, find the second commit, and describe it:
```bash
rm /tmp/jj-split-tool.sh
jj log    # find the second commit's change ID
jj describe -r <second-change-id> -m "Second part"
```

6. Verify (see Safety in SKILL.md).

**Note**: `-m` sets the description for the first commit. The second commit keeps the original description from before the split. If the original had no description, neither will the second commit.

### Example

A commit changes both `src/utils.ts` (refactors a helper) and `src/validation.ts` (adds email validation). Split into two atomic commits:

```bash
# 1. Check what changed
jj diff -r qvlmxkoz

# 2. Write tool script that keeps only the utils.ts refactoring
#    by reverting validation.ts back to parent state
cat > /tmp/jj-split-tool.sh << 'EOF'
#!/bin/bash
cp "$1/src/validation.ts" "$2/src/validation.ts"
EOF

# 3. Split
chmod +x /tmp/jj-split-tool.sh
jj split -r qvlmxkoz --tool /tmp/jj-split-tool.sh -m "♻️ Extract helper function for validation"
rm /tmp/jj-split-tool.sh

# 4. Describe the second commit (has the validation changes)
jj describe -r <new-change-id> -m "🦺 Add email format validation"
```

### Moving line-level changes between commits (manual)

1. Line-level split commit X (see above) to isolate the target changes
2. Move the isolated commit into Y:
```bash
jj squash --from <split-result> --into <target> -u
```
