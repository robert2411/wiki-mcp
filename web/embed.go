// Package webtheme embeds the built-in default theme assets.
package webtheme

import "embed"

// FS holds all files under web/theme/default/.
//
//go:embed theme/default
var FS embed.FS
