Build or run DB Browser for SQLite linked against DoltLite.

The system `sqlitebrowser` package cannot open DoltLite databases because it is
linked against stock SQLite or SQLCipher. This command builds DB Browser in
non-SQLCipher mode and points CMake's SQLite dependency at `libdoltlite`.

Examples:

```bash
gc beads-doltlite sqlitebrowser build
gc beads-doltlite sqlitebrowser open
gc beads-doltlite sqlitebrowser open --city /path/to/city
gc beads-doltlite sqlitebrowser open --db /path/to/.beads/doltlite/hq.db
gc beads-doltlite sqlitebrowser path
```

Useful options:

```bash
--lib DIR          Directory containing libdoltlite.so.
--source DIR       sqlitebrowser source checkout.
--build-dir DIR    CMake build directory.
--bin FILE         Built sqlitebrowser binary to launch.
--repo URL         sqlitebrowser repository URL.
--ref REF          sqlitebrowser branch/tag/commit to build.
--update           Fetch and checkout the configured ref in the source dir.
--jobs N           Parallel build jobs.
```

The default source and build directories are stored under the pack runtime state
directory, usually `.gc/runtime/packs/beads-doltlite`.
