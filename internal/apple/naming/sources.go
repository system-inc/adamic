package naming

import "embed"

// Sources are the naming layer's own sources, so the generator's Version changes when any rule
// here does.
//
//go:embed *.go
var Sources embed.FS
