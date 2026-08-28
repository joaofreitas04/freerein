// Package content embeds the official core components and host
// adapters into the engine binary, so every consume-side command
// works offline (spec guarantee 4).
package content

import "embed"

//go:embed all:core all:adapters
var FS embed.FS
