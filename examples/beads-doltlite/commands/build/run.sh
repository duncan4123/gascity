#!/usr/bin/env bash
# Build DoltLite-linked binaries with CGO/libsqlite3.
set -euo pipefail

usage() {
  cat <<'EOF'
usage: gc beads-doltlite build [gc|bd|all] [options]

Builds DoltLite-linked binaries from the Gas City and beads-doltlite source trees.
The default target is gc.

Options:
  --source DIR       Source checkout for the selected single target.
  --gc-source DIR    Gas City source checkout. Default: ./gascity, current dir, or script checkout.
  --bd-source DIR    beads-doltlite source checkout. Default: ./beads-doltlite or adjacent checkout.
  --lib DIR          Directory containing libdoltlite.so. Default: ./doltlite-work/build or ./doltlite/build.
  --output FILE      Output/install path for the selected single target.
  --gc-output FILE   Output/install path for gc. Default: <gc-source>/bin/gc.
  --bd-output FILE   Output/install path for bd. Default: <bd-source>/bin/bd.
  --version VALUE    Version string embedded in gc. Default: dev.
  --bd-build VALUE   Build string embedded in bd. Default: dev.

Environment overrides:
  GASCITY_SRC, GC_GASCITY_SRC
  BD_SRC, BEADS_DOLTLITE_SRC, GC_BEADS_DOLTLITE_SRC
  DOLTLITE_LIB, GC_DOLTLITE_LIB
  OUTPUT, GC_DOLTLITE_GC_OUTPUT, BD_OUTPUT, GC_DOLTLITE_BD_OUTPUT
  GC_VERSION, GC_COMMIT, GC_BUILD_DATE
  BD_BUILD, BD_COMMIT, BD_BRANCH
EOF
}

die() {
  echo "$*" >&2
  exit 1
}

usage_error() {
  echo "$*" >&2
  usage >&2
  exit 2
}

require_value() {
  if [ -z "${2:-}" ] || [[ "${2:-}" == --* ]]; then
    usage_error "$1 requires a value"
  fi
}

abs_dir() {
  cd "$1" && pwd
}

has_gascity_source() {
  [ -f "$1/go.mod" ] && [ -d "$1/cmd/gc" ]
}

has_bd_source() {
  [ -f "$1/go.mod" ] && [ -d "$1/cmd/bd" ]
}

has_doltlite_lib() {
  [ -r "$1/libdoltlite.so" ] || [ -r "$1/libdoltlite.so.0" ] || [ -r "$1/libdoltlite.dylib" ]
}

find_gascity_source() {
  for candidate in \
    "$CITY_ROOT/gascity" \
    "$CITY_ROOT/../gascity" \
    "$(pwd)" \
    "$SCRIPT_CHECKOUT"; do
    if has_gascity_source "$candidate"; then
      abs_dir "$candidate"
      return 0
    fi
  done
  return 1
}

find_bd_source() {
  for candidate in \
    "$CITY_ROOT/beads-doltlite" \
    "$CITY_ROOT/../beads-doltlite" \
    "$SCRIPT_CHECKOUT/../beads-doltlite" \
    "$(pwd)"; do
    if has_bd_source "$candidate"; then
      abs_dir "$candidate"
      return 0
    fi
  done
  return 1
}

find_doltlite_lib() {
  for candidate in \
    "$CITY_ROOT/doltlite-work/build" \
    "$CITY_ROOT/doltlite/build" \
    "$CITY_ROOT/../doltlite-work/build" \
    "$CITY_ROOT/../doltlite/build"; do
    if has_doltlite_lib "$candidate"; then
      abs_dir "$candidate"
      return 0
    fi
  done
  return 1
}

revision_for() {
  local source_dir="$1"
  if command -v jj >/dev/null 2>&1 && [ -d "$source_dir/.jj" ]; then
    (cd "$source_dir" && jj log --no-graph -r @ -T 'commit_id.short()' 2>/dev/null | tr -d '\n' || true)
    return 0
  fi
  if command -v git >/dev/null 2>&1; then
    git -C "$source_dir" rev-parse --short HEAD 2>/dev/null || true
  fi
}

branch_for() {
  local source_dir="$1"
  if command -v git >/dev/null 2>&1; then
    git -C "$source_dir" symbolic-ref --short HEAD 2>/dev/null || true
  fi
}

common_env_prefix() {
  local tags="$1"
  export CGO_ENABLED=1
  export GOFLAGS="${BASE_GOFLAGS:+$BASE_GOFLAGS }-tags=${tags}"
  export CGO_LDFLAGS="${BASE_CGO_LDFLAGS:+$BASE_CGO_LDFLAGS }-L${DOLTLITE_LIB} -Wl,-rpath,${DOLTLITE_LIB} -ldoltlite"
  export LD_LIBRARY_PATH="${DOLTLITE_LIB}${BASE_LD_LIBRARY_PATH:+:${BASE_LD_LIBRARY_PATH}}"
}

verify_linked_binary() {
  local output="$1"
  local name="$2"
  if ! go version -m "$output" 2>/dev/null | grep -q 'CGO_ENABLED=1'; then
    die "built $name binary does not report CGO_ENABLED=1"
  fi
  if command -v ldd >/dev/null 2>&1; then
    if ! ldd "$output" 2>/dev/null | grep -q 'libdoltlite'; then
      die "built $name binary does not appear to link libdoltlite"
    fi
  fi
}

build_gc() {
  if [ -z "$GASCITY_SRC" ]; then
    GASCITY_SRC="$(find_gascity_source || true)"
  fi
  if [ -z "$GASCITY_SRC" ] || ! has_gascity_source "$GASCITY_SRC"; then
    die "could not find Gas City source; set GASCITY_SRC=/path/to/gascity or pass --gc-source"
  fi
  GASCITY_SRC="$(abs_dir "$GASCITY_SRC")"

  if [ -z "$GC_OUTPUT" ]; then
    GC_OUTPUT="$GASCITY_SRC/bin/gc"
  fi

  local commit="${GC_COMMIT:-}"
  if [ -z "$commit" ]; then
    commit="$(revision_for "$GASCITY_SRC")"
  fi
  commit="${commit:-unknown}"
  local date="${GC_BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

  mkdir -p "$(dirname "$GC_OUTPUT")"
  common_env_prefix "gascity_doltlite_lib,libsqlite3"

  echo "building gc from $GASCITY_SRC"
  echo "linking libdoltlite from $DOLTLITE_LIB"
  echo "writing $GC_OUTPUT"

  (
    cd "$GASCITY_SRC"
    go build \
      -ldflags "-X main.version=${VERSION} -X main.commit=${commit} -X main.date=${date}" \
      -o "$GC_OUTPUT" \
      ./cmd/gc
  )

  verify_linked_binary "$GC_OUTPUT" "gc"
  echo "built libdoltlite-linked gc: $GC_OUTPUT"
}

build_bd() {
  if [ -z "$BD_SRC" ]; then
    BD_SRC="$(find_bd_source || true)"
  fi
  if [ -z "$BD_SRC" ] || ! has_bd_source "$BD_SRC"; then
    die "could not find beads-doltlite source; set BD_SRC=/path/to/beads-doltlite or pass --bd-source"
  fi
  BD_SRC="$(abs_dir "$BD_SRC")"

  if [ -z "$BD_OUTPUT" ]; then
    BD_OUTPUT="$BD_SRC/bin/bd"
  fi

  local commit="${BD_COMMIT:-}"
  if [ -z "$commit" ]; then
    commit="$(revision_for "$BD_SRC")"
  fi
  commit="${commit:-unknown}"
  local branch="${BD_BRANCH:-}"
  if [ -z "$branch" ]; then
    branch="$(branch_for "$BD_SRC")"
  fi
  local ldflags="-X main.Build=${BD_BUILD_VALUE} -X main.Commit=${commit}"
  if [ -n "$branch" ]; then
    ldflags="${ldflags} -X main.Branch=${branch}"
  fi

  mkdir -p "$(dirname "$BD_OUTPUT")"
  common_env_prefix "libsqlite3"

  echo "building bd from $BD_SRC"
  echo "linking libdoltlite from $DOLTLITE_LIB"
  echo "writing $BD_OUTPUT"

  (
    cd "$BD_SRC"
    go build \
      -ldflags "$ldflags" \
      -o "$BD_OUTPUT" \
      ./cmd/bd
  )

  verify_linked_binary "$BD_OUTPUT" "bd"
  echo "built libdoltlite-linked bd: $BD_OUTPUT"
}

CITY_ROOT="${GC_CITY_PATH:-$(pwd)}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT_CHECKOUT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BASE_GOFLAGS="${GOFLAGS:-}"
BASE_CGO_LDFLAGS="${CGO_LDFLAGS:-}"
BASE_LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-}"

TARGET="gc"
COMMON_SOURCE=""
COMMON_OUTPUT="${OUTPUT:-}"
GASCITY_SRC="${GASCITY_SRC:-${GC_GASCITY_SRC:-}}"
BD_SRC="${BD_SRC:-${BEADS_DOLTLITE_SRC:-${GC_BEADS_DOLTLITE_SRC:-}}}"
DOLTLITE_LIB="${DOLTLITE_LIB:-${GC_DOLTLITE_LIB:-}}"
GC_OUTPUT="${GC_DOLTLITE_GC_OUTPUT:-}"
BD_OUTPUT="${BD_OUTPUT:-${GC_DOLTLITE_BD_OUTPUT:-}}"
VERSION="${GC_VERSION:-dev}"
BD_BUILD_VALUE="${BD_BUILD:-dev}"

while [ "$#" -gt 0 ]; do
  case "$1" in
    gc|bd|all)
      TARGET="$1"
      shift
      ;;
    --target)
      require_value "$1" "${2:-}"
      TARGET="$2"
      shift 2
      ;;
    --target=*)
      TARGET="${1#*=}"
      shift
      ;;
    --source)
      require_value "$1" "${2:-}"
      COMMON_SOURCE="$2"
      shift 2
      ;;
    --source=*)
      COMMON_SOURCE="${1#*=}"
      shift
      ;;
    --gc-source)
      require_value "$1" "${2:-}"
      GASCITY_SRC="$2"
      shift 2
      ;;
    --gc-source=*)
      GASCITY_SRC="${1#*=}"
      shift
      ;;
    --bd-source)
      require_value "$1" "${2:-}"
      BD_SRC="$2"
      shift 2
      ;;
    --bd-source=*)
      BD_SRC="${1#*=}"
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
      COMMON_OUTPUT="$2"
      shift 2
      ;;
    --output=*)
      COMMON_OUTPUT="${1#*=}"
      shift
      ;;
    --gc-output)
      require_value "$1" "${2:-}"
      GC_OUTPUT="$2"
      shift 2
      ;;
    --gc-output=*)
      GC_OUTPUT="${1#*=}"
      shift
      ;;
    --bd-output)
      require_value "$1" "${2:-}"
      BD_OUTPUT="$2"
      shift 2
      ;;
    --bd-output=*)
      BD_OUTPUT="${1#*=}"
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
    --bd-build)
      require_value "$1" "${2:-}"
      BD_BUILD_VALUE="$2"
      shift 2
      ;;
    --bd-build=*)
      BD_BUILD_VALUE="${1#*=}"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage_error "unknown argument: $1"
      ;;
  esac
done

case "$TARGET" in
  gc|bd|all) ;;
  *) usage_error "unknown target: $TARGET" ;;
esac

if [ -n "$COMMON_SOURCE" ]; then
  case "$TARGET" in
    gc) GASCITY_SRC="$COMMON_SOURCE" ;;
    bd) BD_SRC="$COMMON_SOURCE" ;;
    all) usage_error "--source is ambiguous with target all; use --gc-source and --bd-source" ;;
  esac
fi

if [ -n "$COMMON_OUTPUT" ]; then
  case "$TARGET" in
    gc) GC_OUTPUT="$COMMON_OUTPUT" ;;
    bd) BD_OUTPUT="$COMMON_OUTPUT" ;;
    all) usage_error "--output is ambiguous with target all; use --gc-output and --bd-output" ;;
  esac
fi

if [ -z "$DOLTLITE_LIB" ]; then
  DOLTLITE_LIB="$(find_doltlite_lib || true)"
fi
if [ -z "$DOLTLITE_LIB" ] || ! has_doltlite_lib "$DOLTLITE_LIB"; then
  die "could not find libdoltlite; set DOLTLITE_LIB=/path/to/doltlite-work/build or pass --lib"
fi
DOLTLITE_LIB="$(abs_dir "$DOLTLITE_LIB")"

if ! command -v go >/dev/null 2>&1; then
  die "go is required to build DoltLite-linked binaries"
fi

case "$TARGET" in
  gc)
    build_gc
    ;;
  bd)
    build_bd
    ;;
  all)
    build_bd
    build_gc
    ;;
esac
