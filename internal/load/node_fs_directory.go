package load

import _ "embed"

//go:embed node_fs_directory.d.ts
var nodeFSDirectoryDeclarations string

//go:embed node_path.d.ts
var nodePathDeclarations string
