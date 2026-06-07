#!/bin/sh
# workspace-setup.sh — idempotent jj workspace creation for Gas City agents.
#
# Usage: workspace-setup.sh <rig-root> <target-dir> <agent-name> [--sync]
#
# Creates a jj workspace at <target-dir> linked to the rig repo at <rig-root>.
# Workspaces share the rig's commit graph so agents work in isolated sandboxes
# that can rebase into each other's lineages without merge commits.
#
# Base revision: resolves to the rig's default bookmark (origin/trunk) via jj.
# Falls back to the current working-copy revision if no remote bookmarks exist.
#
# Called from pre_start in pack configs. Runs before the session is created
# so the agent starts IN the workspace directory.

set -eu

RIG_ROOT="${1:?usage: workspace-setup.sh <rig-root> <target-dir> <agent-name> [--sync]}"
WT="${2:?missing target-dir}"
AGENT="${3:?missing agent-name}"
SYNC="${4:-}"

write_workspace_runtime_files() {
    # Bead redirect: point .beads at the rig's bead store so bd commands work
    # from inside the workspace without a separate bead database.
    mkdir -p "$WT/.beads"
    echo "$RIG_ROOT/.beads" > "$WT/.beads/redirect"

    # Write .jjignore for Gas City runtime clutter.
    cat > "$WT/.jjignore" <<'JJIGNORE'
.beads/redirect
.beads/hooks/
.beads/formulas/
.logs/
.claude/
.codex/
.gemini/
.opencode/
.github/hooks/
.github/copilot-instructions.md
__pycache__/
state.json
JJIGNORE
}

# Idempotent: skip if workspace already exists.
if [ -d "$WT/.jj" ]; then
    write_workspace_runtime_files
    if [ "$SYNC" = "--sync" ]; then
        jj -R "$WT" git fetch 2>/dev/null || true
    fi
    exit 0
fi

mkdir -p "$(dirname "$WT")"

# Stage any pre-existing content in the target directory so it isn't lost
# when jj checks out the workspace revision.
STAGE=""
if [ -d "$WT" ] && [ "$(find "$WT" -mindepth 1 -maxdepth 1 | head -n 1)" ]; then
    STAGE=$(mktemp -d "$(dirname "$WT")/.gascity-workspace-stage.XXXXXX")
    find "$WT" -mindepth 1 -maxdepth 1 -exec mv {} "$STAGE"/ \;
    trap 'restore_stage' EXIT HUP INT TERM
fi
rmdir "$WT" 2>/dev/null || true

restore_stage() {
    if [ -n "${STAGE:-}" ] && [ -d "$STAGE" ]; then
        find "$STAGE" -mindepth 1 -maxdepth 1 | while read -r f; do
            mv "$f" "$WT/" 2>/dev/null || true
        done
        rm -rf "$STAGE"
    fi
}

# Fetch latest from remotes so the workspace starts from a fresh revision.
jj -R "$RIG_ROOT" git fetch 2>/dev/null || true

# Resolve the base revision for the new workspace.
# Prefer trunk()@origin (jj's built-in default-branch revset).
# Fall back through common bookmark names, then the rig's current @.
REVSET=""
if jj -R "$RIG_ROOT" log -r 'trunk()@origin' --no-graph -T '' >/dev/null 2>&1; then
    REVSET="trunk()@origin"
else
    for candidate in main master trunk; do
        if jj -R "$RIG_ROOT" log -r "$candidate@origin" --no-graph -T '' >/dev/null 2>&1; then
            REVSET="$candidate@origin"
            break
        fi
    done
fi
if [ -z "$REVSET" ]; then
    REVSET="@"
fi

SAFE_AGENT=$(printf '%s' "$AGENT" | tr -c 'A-Za-z0-9._-' '-')
WT_HASH=$(printf '%s' "$WT" | cksum | awk '{print $1}')
WORKSPACE_NAME="${SAFE_AGENT}-${WT_HASH}"

if ! jj -R "$RIG_ROOT" workspace add --name "$WORKSPACE_NAME" "$WT" -r "$REVSET" --sparse-patterns full; then
    echo "workspace-setup: failed to create jj workspace at $WT from $RIG_ROOT (revset $REVSET)" >&2
    restore_stage
    exit 1
fi

# Merge staged non-repo content back into the workspace.
if [ -n "$STAGE" ]; then
    find "$STAGE" -mindepth 1 -maxdepth 1 | while read -r f; do
        rel="${f#$STAGE/}"
        if [ -d "$f" ]; then
            mv "$f" "$WT/$rel" 2>/dev/null || true
        else
            if [ ! -e "$WT/$rel" ]; then
                mv "$f" "$WT/$rel" 2>/dev/null || true
            fi
        fi
    done
    rm -rf "$STAGE"
    STAGE=""
fi
trap - EXIT HUP INT TERM

write_workspace_runtime_files

if [ "$SYNC" = "--sync" ]; then
    jj -R "$WT" git fetch 2>/dev/null || true
fi

exit 0
