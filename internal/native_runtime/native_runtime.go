// Package native_runtime holds Adamic's C runtime, the sources every native program's runtime archive is built from,
// under c/: the values, the heap and its regions, the library the emitted C calls, and the hosts it runs on.
//
// The C lives here, apart from internal/native's emitter, so the archive's sources have one home a build can key by
// itself. Package native compiles them; this package only holds them.
package native_runtime

import "embed"

// Sources are the runtime's C files and headers, each at "c/<name>". Only the .c and .h files are embedded, so
// nothing else that sits beside them reaches an archive or its key.
//
//go:embed c/*.c c/*.h
var Sources embed.FS
