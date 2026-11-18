// Package native_runtime holds Adamic's C runtime, the sources every native program's runtime archive is built from,
// under c/: the values, the heap and its regions, the library the emitted C calls, and the hosts it runs on. It builds
// the archive too: the options and flags clang compiles with, the request a product builds ahead, the runtime cache,
// and the jobserver the compiles run under.
//
// The C and its build live here, apart from internal/native's emitter, so an archive's key holds them and nothing of
// the compiler: this package imports none of it (compiler_free's guard). Package native compiles a program and links
// it with the archive this package builds.
package native_runtime

import "embed"

// Sources are the runtime's C files and headers, each at "c/<name>". Only the .c and .h files are embedded, so
// nothing else that sits beside them reaches an archive or its key.
//
//go:embed c/*.c c/*.h
var Sources embed.FS
