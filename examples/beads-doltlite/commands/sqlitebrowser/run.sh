#!/usr/bin/env bash
# Build/run DB Browser for SQLite against libdoltlite so it can open DoltLite DBs.
set -euo pipefail

usage() {
  cat <<'EOF'
usage: gc beads-doltlite sqlitebrowser [open|build|path] [options] [db-file]

Build or run DB Browser for SQLite linked against libdoltlite.

Subcommands:
  open    Open the active DoltLite beads database. Default.
  build   Clone/configure/build DB Browser against libdoltlite.
  path    Print the DoltLite database path that open would use.

Options:
  --city DIR        City workspace root. Default: GC_CITY_PATH or current dir.
  --db FILE         DoltLite database file to open instead of city metadata.
  --lib DIR         Directory containing libdoltlite.so. Default: doltlite-work/build.
  --source DIR      sqlitebrowser source checkout.
  --build-dir DIR   CMake build directory.
  --bin FILE        sqlitebrowser binary path to run.
  --repo URL        sqlitebrowser repository URL.
  --ref REF         sqlitebrowser branch/tag/commit. Default: master.
  --update          Fetch and checkout --ref before building.
  --jobs N          Parallel build jobs. Default: nproc or 4.
  --cmake-arg ARG   Extra argument passed to cmake configure. Repeatable.
  -h, --help        Show this help.

Examples:
  gc beads-doltlite sqlitebrowser build
  gc beads-doltlite sqlitebrowser open
  gc beads-doltlite sqlitebrowser open --city /path/to/city
  gc beads-doltlite sqlitebrowser open --db /path/to/.beads/doltlite/hq.db
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

abs_file() {
  local dir base
  dir="$(dirname "$1")"
  base="$(basename "$1")"
  cd "$dir" && printf '%s/%s\n' "$(pwd)" "$base"
}

has_doltlite_lib() {
  [ -r "$1/libdoltlite.so" ] || [ -r "$1/libdoltlite.so.0" ] || [ -r "$1/libdoltlite.dylib" ]
}

doltlite_lib_file() {
  for name in libdoltlite.so libdoltlite.so.0 libdoltlite.dylib; do
    if [ -r "$DOLTLITE_LIB/$name" ]; then
      printf '%s/%s\n' "$DOLTLITE_LIB" "$name"
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

default_jobs() {
  if command -v nproc >/dev/null 2>&1; then
    nproc
    return 0
  fi
  echo 4
}

json_db_name() {
  local metadata="$1"
  if command -v python3 >/dev/null 2>&1; then
    python3 - "$metadata" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    data = json.load(handle)
print(data.get("dolt_database") or data.get("database") or "")
PY
    return 0
  fi
  awk -F\" '
    /"dolt_database"[[:space:]]*:/ { print $4; found=1; exit }
    /"database"[[:space:]]*:/ && !found { value=$4 }
    END { if (!found && value != "") print value }
  ' "$metadata"
}

database_for_city() {
  local metadata db_name candidate
  metadata="$CITY_ROOT/.beads/metadata.json"
  if [ -r "$metadata" ]; then
    db_name="$(json_db_name "$metadata" || true)"
    if [ -n "$db_name" ]; then
      candidate="$CITY_ROOT/.beads/doltlite/$db_name.db"
      if [ -r "$candidate" ]; then
        abs_file "$candidate"
        return 0
      fi
    fi
  fi

  candidate="$(find "$CITY_ROOT/.beads/doltlite" -maxdepth 1 -type f -name '*.db' 2>/dev/null | sort | head -n 1 || true)"
  if [ -n "$candidate" ]; then
    abs_file "$candidate"
    return 0
  fi
  return 1
}

resolve_browser_bin() {
  if [ -n "$SQLITEBROWSER_BIN" ]; then
    printf '%s\n' "$SQLITEBROWSER_BIN"
    return 0
  fi
  for candidate in \
    "$BUILD_DIR/sqlitebrowser" \
    "$PACK_STATE_DIR/bin/sqlitebrowser-doltlite"; do
    if [ -x "$candidate" ]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done
  return 1
}

verify_browser_links_doltlite() {
  local bin="$1"
  if command -v ldd >/dev/null 2>&1; then
    if ! ldd "$bin" 2>/dev/null | grep -q 'libdoltlite'; then
      die "$bin is not linked to libdoltlite; run: gc beads-doltlite sqlitebrowser build"
    fi
  fi
}

prepare_source() {
  if [ ! -d "$SOURCE_DIR/.git" ]; then
    mkdir -p "$(dirname "$SOURCE_DIR")"
    echo "cloning DB Browser for SQLite into $SOURCE_DIR"
    git clone --depth 1 --branch "$SQLITEBROWSER_REF" "$SQLITEBROWSER_REPO" "$SOURCE_DIR" ||
      die "cloning sqlitebrowser failed"
    if [ -f "$SOURCE_DIR/.gitmodules" ]; then
      git -C "$SOURCE_DIR" submodule update --init --recursive ||
        die "initializing sqlitebrowser submodules failed"
    fi
    return 0
  fi

  if [ "$UPDATE_SOURCE" = "1" ]; then
    echo "updating DB Browser source in $SOURCE_DIR to $SQLITEBROWSER_REF"
    git -C "$SOURCE_DIR" fetch --depth 1 origin "$SQLITEBROWSER_REF" ||
      die "fetching sqlitebrowser ref failed"
    git -C "$SOURCE_DIR" checkout --detach FETCH_HEAD ||
      die "checking out sqlitebrowser ref failed"
    if [ -f "$SOURCE_DIR/.gitmodules" ]; then
      git -C "$SOURCE_DIR" submodule update --init --recursive ||
        die "updating sqlitebrowser submodules failed"
    fi
  fi
}

build_browser() {
  command -v git >/dev/null 2>&1 || die "git is required to clone sqlitebrowser"
  command -v cmake >/dev/null 2>&1 || die "cmake is required to build sqlitebrowser"

  local lib_file browser_bin
  lib_file="$(doltlite_lib_file)" || die "could not find libdoltlite shared library under $DOLTLITE_LIB"

  prepare_source
  mkdir -p "$BUILD_DIR"

  echo "configuring DB Browser for SQLite"
  echo "source: $SOURCE_DIR"
  echo "build:  $BUILD_DIR"
  echo "lib:    $lib_file"

  cmake \
    -S "$SOURCE_DIR" \
    -B "$BUILD_DIR" \
    -Dsqlcipher=0 \
    -DENABLE_TESTING=OFF \
    -DFORCE_INTERNAL_QSCINTILLA=ON \
    -DSQLite3_INCLUDE_DIR="$DOLTLITE_LIB" \
    -DSQLite3_LIBRARY="$lib_file" \
    -DCMAKE_BUILD_RPATH="$DOLTLITE_LIB" \
    -DCMAKE_INSTALL_RPATH="$DOLTLITE_LIB" \
    "${CMAKE_ARGS[@]}"

  cmake --build "$BUILD_DIR" --target sqlitebrowser --parallel "$JOBS"

  browser_bin="$BUILD_DIR/sqlitebrowser"
  if [ ! -x "$browser_bin" ]; then
    browser_bin="$(find "$BUILD_DIR" -type f -name sqlitebrowser -perm -111 | head -n 1 || true)"
  fi
  if [ -z "$browser_bin" ] || [ ! -x "$browser_bin" ]; then
    die "sqlitebrowser build completed but no executable was found under $BUILD_DIR"
  fi
  browser_bin="$(abs_file "$browser_bin")"
  verify_browser_links_doltlite "$browser_bin"

  mkdir -p "$PACK_STATE_DIR/bin"
  ln -sfn "$browser_bin" "$PACK_STATE_DIR/bin/sqlitebrowser-doltlite"
  echo "built DoltLite-linked sqlitebrowser: $browser_bin"
  echo "launcher symlink: $PACK_STATE_DIR/bin/sqlitebrowser-doltlite"
}

open_browser() {
  local db_file browser_bin
  if [ -n "$DB_FILE" ]; then
    [ -r "$DB_FILE" ] || die "database file does not exist or is not readable: $DB_FILE"
    db_file="$(abs_file "$DB_FILE")"
  else
    db_file="$(database_for_city || true)"
    [ -n "$db_file" ] || die "could not find a DoltLite database under $CITY_ROOT/.beads/doltlite; pass --db"
  fi

  browser_bin="$(resolve_browser_bin || true)"
  if [ -z "$browser_bin" ] || [ ! -x "$browser_bin" ]; then
    die "DoltLite-linked sqlitebrowser is not built; run: gc beads-doltlite sqlitebrowser build"
  fi
  browser_bin="$(abs_file "$browser_bin")"
  verify_browser_links_doltlite "$browser_bin"

  if [ -z "${DISPLAY:-}" ] && [ -z "${WAYLAND_DISPLAY:-}" ] && [ -z "${QT_QPA_PLATFORM:-}" ]; then
    die "DISPLAY, WAYLAND_DISPLAY, or QT_QPA_PLATFORM is required to launch sqlitebrowser"
  fi

  export LD_LIBRARY_PATH="$DOLTLITE_LIB${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
  echo "opening $db_file with $browser_bin"
  exec "$browser_bin" "$db_file" "${BROWSER_ARGS[@]}"
}

ACTION="open"
CITY_ROOT="${GC_CITY_PATH:-$(pwd)}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACK_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
PACK_STATE_DIR="${GC_PACK_STATE_DIR:-$CITY_ROOT/.gc/runtime/packs/beads-doltlite}"
DOLTLITE_LIB="${DOLTLITE_LIB:-${GC_DOLTLITE_LIB:-}}"
SOURCE_DIR="${GC_DOLTLITE_SQLITEBROWSER_SOURCE:-$PACK_STATE_DIR/sqlitebrowser-src}"
BUILD_DIR="${GC_DOLTLITE_SQLITEBROWSER_BUILD_DIR:-$PACK_STATE_DIR/sqlitebrowser-build}"
SOURCE_DIR_SET=0
BUILD_DIR_SET=0
SQLITEBROWSER_BIN="${GC_DOLTLITE_SQLITEBROWSER_BIN:-}"
SQLITEBROWSER_REPO="${GC_DOLTLITE_SQLITEBROWSER_REPO:-https://github.com/sqlitebrowser/sqlitebrowser.git}"
SQLITEBROWSER_REF="${GC_DOLTLITE_SQLITEBROWSER_REF:-master}"
UPDATE_SOURCE=0
JOBS="${GC_DOLTLITE_SQLITEBROWSER_JOBS:-$(default_jobs)}"
DB_FILE=""
CMAKE_ARGS=()
BROWSER_ARGS=()

if [ "$#" -gt 0 ]; then
  case "$1" in
    open|build|path)
      ACTION="$1"
      shift
      ;;
  esac
fi

while [ "$#" -gt 0 ]; do
  case "$1" in
    --city)
      require_value "$1" "${2:-}"
      CITY_ROOT="$2"
      shift 2
      ;;
    --city=*)
      CITY_ROOT="${1#*=}"
      shift
      ;;
    --db)
      require_value "$1" "${2:-}"
      DB_FILE="$2"
      shift 2
      ;;
    --db=*)
      DB_FILE="${1#*=}"
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
    --source)
      require_value "$1" "${2:-}"
      SOURCE_DIR="$2"
      SOURCE_DIR_SET=1
      shift 2
      ;;
    --source=*)
      SOURCE_DIR="${1#*=}"
      SOURCE_DIR_SET=1
      shift
      ;;
    --build-dir)
      require_value "$1" "${2:-}"
      BUILD_DIR="$2"
      BUILD_DIR_SET=1
      shift 2
      ;;
    --build-dir=*)
      BUILD_DIR="${1#*=}"
      BUILD_DIR_SET=1
      shift
      ;;
    --bin)
      require_value "$1" "${2:-}"
      SQLITEBROWSER_BIN="$2"
      shift 2
      ;;
    --bin=*)
      SQLITEBROWSER_BIN="${1#*=}"
      shift
      ;;
    --repo)
      require_value "$1" "${2:-}"
      SQLITEBROWSER_REPO="$2"
      shift 2
      ;;
    --repo=*)
      SQLITEBROWSER_REPO="${1#*=}"
      shift
      ;;
    --ref)
      require_value "$1" "${2:-}"
      SQLITEBROWSER_REF="$2"
      shift 2
      ;;
    --ref=*)
      SQLITEBROWSER_REF="${1#*=}"
      shift
      ;;
    --update)
      UPDATE_SOURCE=1
      shift
      ;;
    --jobs)
      require_value "$1" "${2:-}"
      JOBS="$2"
      shift 2
      ;;
    --jobs=*)
      JOBS="${1#*=}"
      shift
      ;;
    --cmake-arg)
      require_value "$1" "${2:-}"
      CMAKE_ARGS+=("$2")
      shift 2
      ;;
    --cmake-arg=*)
      CMAKE_ARGS+=("${1#*=}")
      shift
      ;;
    --)
      shift
      BROWSER_ARGS+=("$@")
      break
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --*)
      usage_error "unknown argument: $1"
      ;;
    *)
      if [ "$ACTION" = "open" ] && [ -z "$DB_FILE" ]; then
        DB_FILE="$1"
        shift
      else
        BROWSER_ARGS+=("$1")
        shift
      fi
      ;;
  esac
done

case "$ACTION" in
  open|build|path) ;;
  *) usage_error "unknown subcommand: $ACTION" ;;
esac

case "$JOBS" in
  ''|*[!0-9]*) usage_error "--jobs must be a positive integer" ;;
  0) usage_error "--jobs must be greater than zero" ;;
esac

CITY_ROOT="$(abs_dir "$CITY_ROOT")"
PACK_STATE_DIR="${GC_PACK_STATE_DIR:-$CITY_ROOT/.gc/runtime/packs/beads-doltlite}"
if [ "$SOURCE_DIR_SET" = "0" ] && [ -z "${GC_DOLTLITE_SQLITEBROWSER_SOURCE:-}" ]; then
  SOURCE_DIR="$PACK_STATE_DIR/sqlitebrowser-src"
fi
if [ "$BUILD_DIR_SET" = "0" ] && [ -z "${GC_DOLTLITE_SQLITEBROWSER_BUILD_DIR:-}" ]; then
  BUILD_DIR="$PACK_STATE_DIR/sqlitebrowser-build"
fi

if [ -z "$DOLTLITE_LIB" ]; then
  DOLTLITE_LIB="$(find_doltlite_lib || true)"
fi
if [ -z "$DOLTLITE_LIB" ] || ! has_doltlite_lib "$DOLTLITE_LIB"; then
  die "could not find libdoltlite; set DOLTLITE_LIB=/path/to/doltlite-work/build or pass --lib"
fi
DOLTLITE_LIB="$(abs_dir "$DOLTLITE_LIB")"

case "$ACTION" in
  build)
    build_browser
    ;;
  path)
    if [ -n "$DB_FILE" ]; then
      abs_file "$DB_FILE"
    else
      database_for_city || die "could not find a DoltLite database under $CITY_ROOT/.beads/doltlite; pass --db"
    fi
    ;;
  open)
    open_browser
    ;;
esac
