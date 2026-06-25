// Package beadsdoltlite embeds the built-in DoltLite beads provider pack.
package beadsdoltlite

import "embed"

// PackFS contains the DoltLite beads provider pack files.
//
//go:embed pack.toml doctor commands examples formulas orders skills all:assets
var PackFS embed.FS
