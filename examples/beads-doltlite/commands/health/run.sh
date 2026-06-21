#!/usr/bin/env bash
# doltlite health — check doltlite database integrity and stats via bd.
set -euo pipefail

BEADS_DIR="${BEADS_DIR:-${GC_CITY_PATH:-.}/.beads}"
SCOPE_DIR="${BEADS_DIR%/.*}"
OUTPUT_FILE="${TMPDIR:-/tmp}/beads-doltlite-health-$$.json"
ERROR_FILE="${TMPDIR:-/tmp}/beads-doltlite-health-err-$$.txt"

if command -v bd >/dev/null 2>&1; then
  cd "$SCOPE_DIR" || exit 1
  export BEADS_BACKEND=doltlite GC_BEADS_BACKEND=doltlite BD_NON_INTERACTIVE=1

  # Let order-level context deadlines control total runtime.
  # A fixed shell timeout here can fire first and produce misleading
  # "context canceled" from the dispatcher while the store is still
  # making progress.
  status=0
  if [ -n "${GC_DOLTLITE_HEALTH_TIMEOUT:-}" ]; then
    if command -v timeout >/dev/null 2>&1; then
      timeout "$GC_DOLTLITE_HEALTH_TIMEOUT" bd status --json >"$OUTPUT_FILE" 2>&1 || status=$?
    else
      bd status --json >"$OUTPUT_FILE" 2>&1 || status=$?
    fi
  else
    bd status --json >"$OUTPUT_FILE" 2>&1 || status=$?
  fi

  if [ "$status" -ne 0 ]; then
    cat "$OUTPUT_FILE"
    rm -f "$OUTPUT_FILE" "$ERROR_FILE"
    exit "$status"
  fi

  if [ -f "$OUTPUT_FILE" ]; then
    if command -v jq >/dev/null 2>&1; then
      jq '.ok = (if has("error") then false else true end) | .schema_version = ( .schema_version // 1 )' "$OUTPUT_FILE"
    elif command -v python3 >/dev/null 2>&1; then
      python3 - "$OUTPUT_FILE" <<'PY'
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as handle:
    payload = json.load(handle)
if "ok" not in payload:
    payload["ok"] = not bool(payload.get("error"))
if "schema_version" not in payload:
    payload["schema_version"] = 1
print(json.dumps(payload, separators=(",", ":")))
PY
    else
      cat "$OUTPUT_FILE"
    fi
  fi

  rm -f "$OUTPUT_FILE" "$ERROR_FILE"
  exit 0
else
  echo '{"ok":false,"error":"bd CLI not found","schema_version":1}'
  rm -f "$OUTPUT_FILE" "$ERROR_FILE"
  exit 1
fi
