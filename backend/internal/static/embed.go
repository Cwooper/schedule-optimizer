package static

import "embed"

// Dist contains the compiled frontend assets.
// This will be empty during development - build with `make build-frontend` first.
//
//go:embed all:dist
var Dist embed.FS
