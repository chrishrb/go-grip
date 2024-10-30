package defaults

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed static
var StaticFiles embed.FS
