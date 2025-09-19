package web

import "embed"

// Templates holds the embedded HTML templates.
//
//go:embed templates/*.html
var Templates embed.FS
