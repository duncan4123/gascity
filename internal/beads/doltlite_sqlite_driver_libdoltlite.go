//go:build gascity_native_beads || gascity_doltlite_lib

package beads

import _ "github.com/mattn/go-sqlite3"

// doltliteSQLDriverName names the C-backed sqlite driver used by native DoltLite builds.
const doltliteSQLDriverName = "sqlite3"
