// Package jjspr embeds the jj-spr workflow pack for bundling into the gc binary.
package jjspr

import "embed"

// PackFS contains the jj-spr pack files.
//
//go:embed pack.toml agents doctor formulas all:assets
var PackFS embed.FS
