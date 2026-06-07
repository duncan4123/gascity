//go:build gascity_native_beads && !gascity_doltlite_lib

package beads

import _ "modernc.org/sqlite"

const doltliteSQLDriverName = "sqlite"
