#!/usr/bin/env bash
# Build gc with CGO/libsqlite3 so it links against libdoltlite.
set -euo pipefail

usage() {
  cat <<'EOF'
usage: gc beads-doltlite build [--source DIR] [--lib DIR] [--output FILE] [--version VERSION]

Builds the Gas City gc binary against libdoltlite.

Options:
  --source DIR      Gas City source checkout. Default: ./gascity, current dir, or script checkout.
  --lib DIR         Directory containing libdoltlite.so. Default: ./doltlite-work/build or ./doltlite/build.
  --output FILE     Output/install path for the binary. Default: <source>/bin/gc.
  --version VALUE   Version string embedded in the binary. Default: dev.

Environment overrides:
  GASCITY_SRC, GC_GASCITY_SRC
  DOLTLITE_LIB, GC_DOLTLITE_LIB
  OUTPUT, GC_DOLTLITE_GC_OUTPUT
  GC_VERSION, GC_COMMIT, GC_BUILD_DATE
EOF
}

require_value() {
  if [ -z "${2:-}" ] || [[ "${2:-}" == --* ]]; then
    echo "$1 requires a value" >&2
    usage >&2
    exit 2
  fi
}

has_gascity_source() {
  [ -f "$1/go.mod" ] && [ -d "$1/cmd/gc" ]
}

has_doltlite_lib() {
  [ -r "$1/libdoltlite.so" ] || [ -r "$1/libdoltlite.so.0" ] || [ -r "$1/libdoltlite.dylib" ]
}

CITY_ROOT="${GC_CITY_PATH:-$(pwd)}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_CHECKOUT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

GASCITY_SRC="${GASCITY_SRC:-${GC_GASCITY_SRC:-}}"
DOLTLITE_LIB="${DOLTLITE_LIB:-${GC_DOLTLITE_LIB:-}}"
OUTPUT="${OUTPUT:-${GC_DOLTLITE_GC_OUTPUT:-}}"
VERSION="${GC_VERSION:-dev}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --source)
      require_value "$1" "${2:-}"
      GASCITY_SRC="$2"
      shift 2
      ;;
    --source=*)
      GASCITY_SRC="${1#*=}"
      shift
      ;;
    --lib)
      require_value "$1" "${2:-}"
      DOLTLITE_LIB="$2"
      shift 2
      ;;
    --lib=*)
      DOLTLITE_LIB="${1#*=}"
      shift
      ;;
    --output)
      require_value "$1" "${2:-}"
      OUTPUT="$2"
      shift 2
      ;;
    --output=*)
      OUTPUT="${1#*=}"
      shift
      ;;
    --version)
      require_value "$1" "${2:-}"
      VERSION="$2"
      shift 2
      ;;
    --version=*)
      VERSION="${1#*=}"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [ -z "$GASCITY_SRC" ]; then
  for candidate in \
    "$CITY_ROOT/gascity" \
    "$CITY_ROOT/../gascity" \
    "$(pwd)" \
    "$SCRIPT_CHECKOUT"; do
    if has_gascity_source "$candidate"; then
      GASCITY_SRC="$(cd "$candidate" && pwd)"
      break
    fi
  done
fi

if [ -z "$GASCITY_SRC" ] || ! has_gascity_source "$GASCITY_SRC"; then
  echo "could not find Gas City source; set GASCITY_SRC=/path/to/gascity or pass --source" >&2
  exit 1
fi

if [ -z "$DOLTLITE_LIB" ]; then
  for candidate in \
    "$CITY_ROOT/doltlite-work/build" \
    "$CITY_ROOT/doltlite/build" \
    "$CITY_ROOT/../doltlite-work/build" \
    "$CITY_ROOT/../doltlite/build"; do
    if has_doltlite_lib "$candidate"; then
      DOLTLITE_LIB="$(cd "$candidate" && pwd)"
      break
    fi
  done
fi

if [ -z "$DOLTLITE_LIB" ] || ! has_doltlite_lib "$DOLTLITE_LIB"; then
  echo "could not find libdoltlite; set DOLTLITE_LIB=/path/to/doltlite-work/build or pass --lib" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go is required to build gc" >&2
  exit 1
fi

if [ -z "$OUTPUT" ]; then
  OUTPUT="$GASCITY_SRC/bin/gc"
fi

COMMIT="${GC_COMMIT:-}"
if [ -z "$COMMIT" ] && command -v jj >/dev/null 2>&1 && [ -d "$GASCITY_SRC/.jj" ]; then
  COMMIT="$(cd "$GASCITY_SRC" && jj log --no-graph -r @ -T 'commit_id.short()' 2>/dev/null | tr -d '\n' || true)"
fi
if [ -z "$COMMIT" ] && command -v git >/dev/null 2>&1; then
  COMMIT="$(git -C "$GASCITY_SRC" rev-parse --short HEAD 2>/dev/null || true)"
fi
COMMIT="${COMMIT:-unknown}"
DATE="${GC_BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

mkdir -p "$(dirname "$OUTPUT")"

export CGO_ENABLED=1
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-tags=gascity_doltlite_lib,libsqlite3"
export CGO_LDFLAGS="${CGO_LDFLAGS:+$CGO_LDFLAGS }-L${DOLTLITE_LIB} -Wl,-rpath,${DOLTLITE_LIB} -ldoltlite"
export LD_LIBRARY_PATH="${DOLTLITE_LIB}${LD_LIBRARY_PATH:+:${LD_LIBRARY_PATH}}"

echo "building gc from $GASCITY_SRC"
echo "linking libdoltlite from $DOLTLITE_LIB"
echo "writing $OUTPUT"

(
  cd "$GASCITY_SRC"
  go build \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o "$OUTPUT" \
    ./cmd/gc
)

if ! go version -m "$OUTPUT" 2>/dev/null | grep -q 'CGO_ENABLED=1'; then
  echo "built binary does not report CGO_ENABLED=1" >&2
  exit 1
fi

if command -v ldd >/dev/null 2>&1; then
  if ! ldd "$OUTPUT" 2>/dev/null | grep -q 'libdoltlite'; then
    echo "built binary does not appear to link libdoltlite" >&2
    exit 1
  fi
fi

echo "built libdoltlite-linked gc: $OUTPUT"
