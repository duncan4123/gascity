#!/bin/sh
# worktree-setup.sh — idempotent jj workspace creation for Gas City agents.
#
# Usage: worktree-setup.sh <rig-root> <target-dir> <agent-name> [--sync]
#
# Ensures the target directory is a jj workspace of the rig repo. For
# backward compatibility, the older <repo-dir> <agent-name> <city-root>
# signature still works and resolves the target under
# <city-root>/.gc/worktrees/<rig>/<agent-name>.
#
# Called from pre_start in pack configs. Runs before the session is created
# so the agent starts IN the workspace directory.

set -eu

RIG_ROOT="${1:?usage: worktree-setup.sh <rig-root> <target-dir> <agent-name> [--sync]}"
ARG2="${2:?missing target-dir}"
ARG3="${3:?missing agent-name}"

is_path_like() {
    # Legacy mode passes the city path as arg 3. Agent names are validated
    # elsewhere and are not expected to look like filesystem paths.
    case "$1" in
        */*|.*|*:*|*\\*) return 0 ;;
        *) return 1 ;;
    esac
}

if is_path_like "$ARG3"; then
    # Legacy signature: <rig-root> <agent-name> <city-root> [--sync]
    AGENT="$ARG2"
    CITY="$ARG3"
    RIG=$(basename "$RIG_ROOT")
    WT="$CITY/.gc/worktrees/$RIG/$AGENT"
    SYNC="${4:-}"
else
    WT="$ARG2"
    AGENT="$ARG3"
    SYNC="${4:-}"
fi

sync_workspace() {
    [ "$SYNC" = "--sync" ] || return 0
    if ! git -C "$WT" remote get-url origin >/dev/null 2>&1; then
        return 0
    fi
    git -C "$WT" fetch origin --prune --prune-tags --quiet 2>/dev/null || true
    # jj needs to see the latest git refs; fetch updates both.
    jj -R "$WT" git fetch 2>/dev/null || true
}

# Idempotent: skip if workspace already exists.
if [ -d "$WT/.jj" ]; then
    sync_workspace
    exit 0
fi

mkdir -p "$(dirname "$WT")"

STAGE=""

merge_stage_entry() (
    # Merge stashed non-repo content (from a pre-existing dir) back into the
    # workspace after workspace creation. Entries are emitted from stage_dir
    # as "TYPE:REL_PATH" where TYPE is "dir" or "file".
    set -eu
    ENTRY="$1"
    WT="$2"
    STAGE="$3"
    case "$ENTRY" in
        dir:*) mv "$STAGE/${ENTRY#dir:}" "$WT/${ENTRY#dir:}" 2>/dev/null || true ;;
        file:*)
            # Don't overwrite files that exist in the repo checkout.
            if [ ! -e "$WT/${ENTRY#file:}" ]; then
                mv "$STAGE/${ENTRY#file:}" "$WT/${ENTRY#file:}" 2>/dev/null || true
            fi
            ;;
    esac
)

restore_stage() {
    if [ -n "$STAGE" ] && [ -d "$STAGE" ]; then
        find "$STAGE" -mindepth 1 -maxdepth 1 | while read -r f; do
            mv "$f" "$WT/" 2>/dev/null || true
        done
        rm -rf "$STAGE"
    fi
}

if [ -d "$WT" ] && [ "$(find "$WT" -mindepth 1 -maxdepth 1 | head -n 1)" ]; then
    STAGE=$(mktemp -d "$(dirname "$WT")/.gascity-worktree-stage.XXXXXX")
    find "$WT" -mindepth 1 -maxdepth 1 -exec mv {} "$STAGE"/ \;
    trap 'restore_stage' EXIT HUP INT TERM
fi

rmdir "$WT" 2>/dev/null || true

# Determine the upstream default branch ref and fetch it so the workspace
# is created from the remote tip. Without this, the workspace inherits a
# stale local revision and feature branches cut from it carry already-merged
# commits that the refinery rebase rejects.
jj -R "$RIG_ROOT" git fetch 2>/dev/null || true
git -C "$RIG_ROOT" fetch origin --prune --prune-tags --quiet 2>/dev/null || true

# Resolve the default branch revset for the new workspace.
DEFAULT_REF=$(git -C "$RIG_ROOT" symbolic-ref refs/remotes/origin/HEAD 2>/dev/null || true)
if [ -z "$DEFAULT_REF" ]; then
    # Fallback: try common default branch names.
    for candidate in main master; do
        if git -C "$RIG_ROOT" show-ref --verify --quiet "refs/remotes/origin/$candidate"; then
            DEFAULT_REF="refs/remotes/origin/$candidate"
            break
        fi
    done
fi

if [ -n "$DEFAULT_REF" ]; then
    # Convert git ref (refs/remotes/origin/main) to jj revset.
    REVSET=$(echo "$DEFAULT_REF" | sed 's|refs/remotes/origin/||')
else
    # Last resort: use whatever HEAD is.
    REVSET="@"
fi

if ! jj -R "$RIG_ROOT" workspace add "$WT" -r "$REVSET" --sparse-patterns full; then
    echo "worktree-setup: failed to create jj workspace at $WT from $RIG_ROOT (revset $REVSET)" >&2
    restore_stage
    exit 1
fi

# Merge staged non-repo content back into the workspace.
if [ -n "$STAGE" ]; then
    find "$STAGE" -mindepth 1 -maxdepth 1 | while read -r f; do
        rel="${f#$STAGE/}"
        if [ -d "$f" ]; then
            echo "dir:$rel"
        else
            echo "file:$rel"
        fi
    done | while read -r entry; do
        merge_stage_entry "$entry" "$WT" "$STAGE"
    done
    rm -rf "$STAGE"
    STAGE=""
fi
trap - EXIT HUP INT TERM

# Bead redirect for filesystem beads.
mkdir -p "$WT/.beads"
echo "$RIG_ROOT/.beads" > "$WT/.beads/redirect"

# Submodule init (best-effort).
git -C "$WT" submodule init 2>/dev/null || true

# Per-workspace git excludes (colocated repos still use git for branch ops).
EXCLUDE=$(git -C "$WT" rev-parse --git-path info/exclude 2>/dev/null) || EXCLUDE=""
if [ -n "$EXCLUDE" ]; then
    case "$EXCLUDE" in
        /*) ;;
        *) EXCLUDE="$WT/$EXCLUDE" ;;
    esac
    mkdir -p "$(dirname "$EXCLUDE")"
    touch "$EXCLUDE"

    MARKER="# Gas City workspace infrastructure (local excludes)"
    if ! grep -qF "$MARKER" "$EXCLUDE" 2>/dev/null; then
        printf '\n%s\n' "$MARKER" >> "$EXCLUDE"
    fi

    append_exclude() {
        if ! grep -qxF "$1" "$EXCLUDE" 2>/dev/null; then
            printf '%s\n' "$1" >> "$EXCLUDE"
        fi
    }

    append_exclude ".beads/redirect"
    append_exclude ".beads/hooks/"
    append_exclude ".beads/formulas/"
    append_exclude ".logs/"
    append_exclude "worktrees/"
    append_exclude "__pycache__/"
    append_exclude ".claude/"
    append_exclude ".codex/"
    append_exclude ".gemini/"
    append_exclude ".opencode/"
    append_exclude ".github/hooks/"
    append_exclude ".github/copilot-instructions.md"
    append_exclude "state.json"
fi

# Optional sync.
sync_workspace

exit 0
