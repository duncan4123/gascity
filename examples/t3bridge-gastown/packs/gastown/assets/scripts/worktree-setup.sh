#!/bin/sh
set -eu

rig_root="${1:?rig root required}"
work_dir="${2:?work dir required}"
agent="${3:-agent}"
mode="${4:---sync}"

mkdir -p "$(dirname "$work_dir")"

acquire_worktree_lock() {
  command -v flock >/dev/null 2>&1 || return 0

  common_dir=$(git -C "$rig_root" rev-parse --git-common-dir 2>/dev/null) || return 0
  case "$common_dir" in
    /*) ;;
    *) common_dir="$rig_root/$common_dir" ;;
  esac

  lock_timeout="${GC_WORKTREE_SETUP_LOCK_TIMEOUT:-300}"
  exec 9>"$common_dir/gascity-worktree-setup.lock"
  if ! flock -w "$lock_timeout" 9; then
    echo "worktree-setup: timed out waiting for worktree setup lock for $rig_root" >&2
    exit 1
  fi
}

acquire_worktree_lock

if [ -d "$work_dir/.git" ] || git -C "$work_dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  if [ "$mode" = "--sync" ]; then
    git -C "$work_dir" fetch --all --prune >/dev/null 2>&1 || true
  fi
  exit 0
fi

branch="gc/${agent}"
git -C "$rig_root" worktree add -B "$branch" "$work_dir" HEAD
