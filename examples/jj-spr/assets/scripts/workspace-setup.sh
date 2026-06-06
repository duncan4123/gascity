#!/bin/sh
set -eu

RIG_ROOT="${1:?usage: workspace-setup.sh <rig-root> <target-dir> <agent-name> [--sync]}"
WT="${2:?missing target-dir}"
AGENT="${3:?missing agent-name}"
SYNC="${4:-}"

write_runtime_files() {
    mkdir -p "$WT/.beads"
    echo "$RIG_ROOT/.beads" > "$WT/.beads/redirect"
    cat > "$WT/.jjignore" <<'JJIGNORE'
.beads/redirect
.beads/hooks/
.beads/formulas/
.logs/
.claude/
.codex/
.gemini/
.opencode/
__pycache__/
state.json
JJIGNORE
}

if [ -d "$WT/.jj" ]; then
    write_runtime_files
    if [ "$SYNC" = "--sync" ]; then
        jj -R "$WT" git fetch 2>/dev/null || true
    fi
    exit 0
fi

mkdir -p "$(dirname "$WT")"
if [ -d "$WT" ] && [ "$(find "$WT" -mindepth 1 -maxdepth 1 | head -n 1)" ]; then
    echo "workspace-setup: target exists and is non-empty: $WT" >&2
    exit 1
fi

jj -R "$RIG_ROOT" git fetch 2>/dev/null || true

REVSET=""
if jj -R "$RIG_ROOT" log -r 'trunk()@origin' --no-graph -T '' >/dev/null 2>&1; then
    REVSET="trunk()@origin"
else
    for candidate in main master trunk beads-doltlite; do
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

jj -R "$RIG_ROOT" workspace add --name "$WORKSPACE_NAME" "$WT" -r "$REVSET" --sparse-patterns full
write_runtime_files

if [ "$SYNC" = "--sync" ]; then
    jj -R "$WT" git fetch 2>/dev/null || true
fi

