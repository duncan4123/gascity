#!/usr/bin/env bash
# doltlite health — check doltlite database integrity and stats via bd.
set -euo pipefail

BEADS_DIR="${BEADS_DIR:-${GC_CITY_PATH:-.}/.beads}"
SCOPE_DIR="${BEADS_DIR%/.*}"

if command -v bd >/dev/null 2>&1; then
  cd "$SCOPE_DIR" || exit 1
  export BEADS_BACKEND=doltlite GC_BEADS_BACKEND=doltlite BD_NON_INTERACTIVE=1
  if command -v timeout >/dev/null 2>&1; then
    timeout "${GC_DOLTLITE_HEALTH_TIMEOUT:-15s}" bd status --json 2>&1
  else
    bd status --json 2>&1
  fi
else
  echo '{"ok":false,"error":"bd CLI not found"}'
  exit 1
fi
