package load

import _ "embed"

//go:embed node_buffer.d.ts
var nodeBufferDeclarations string

//go:embed node_crypto.d.ts
var nodeCryptoDeclarations string

func nodeBufferPrelude() string { return nodeBufferDeclarations + "\n" + nodeCryptoDeclarations }
