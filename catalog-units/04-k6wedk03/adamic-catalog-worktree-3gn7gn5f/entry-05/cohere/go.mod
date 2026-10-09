module github.com/system-inc/cohere

go 1.27

replace (
	github.com/microsoft/TypeScript/tsc/shim/ast => ./TypeScript-shim/ast
	github.com/microsoft/TypeScript/tsc/shim/bundled => ./TypeScript-shim/bundled
	github.com/microsoft/TypeScript/tsc/shim/checker => ./TypeScript-shim/checker
	github.com/microsoft/TypeScript/tsc/shim/compiler => ./TypeScript-shim/compiler
	github.com/microsoft/TypeScript/tsc/shim/core => ./TypeScript-shim/core
	github.com/microsoft/TypeScript/tsc/shim/format => ./TypeScript-shim/format
	github.com/microsoft/TypeScript/tsc/shim/incremental => ./TypeScript-shim/incremental
	github.com/microsoft/TypeScript/tsc/shim/locale => ./TypeScript-shim/locale
	github.com/microsoft/TypeScript/tsc/shim/parser => ./TypeScript-shim/parser
	github.com/microsoft/TypeScript/tsc/shim/scanner => ./TypeScript-shim/scanner
	github.com/microsoft/TypeScript/tsc/shim/tsoptions => ./TypeScript-shim/tsoptions
	github.com/microsoft/TypeScript/tsc/shim/tspath => ./TypeScript-shim/tspath
	github.com/microsoft/TypeScript/tsc/shim/vfs => ./TypeScript-shim/vfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs => ./TypeScript-shim/vfs/cachedvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs => ./TypeScript-shim/vfs/osvfs

	// The single static assignment form cohere's high-level IR and Adamic's flow graph share, a
	// module of its own so Adamic can import it through its cohere submodule (#ejcnkja).
	github.com/system-inc/cohere/static_single_assignment => ./static_single_assignment

	// React's mutation and aliasing model, the effect vocabulary and the mutable ranges inferred from
	// it, shared with Adamic the same way (#kr2yp54).
	github.com/system-inc/cohere/mutation_aliasing => ./mutation_aliasing
)

require (
	github.com/microsoft/TypeScript/tsc/shim/ast v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/bundled v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/checker v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/compiler v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/core v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/format v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/incremental v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/locale v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/parser v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/scanner v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tsoptions v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tspath v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs v0.0.0
	github.com/system-inc/cohere/mutation_aliasing v0.0.0
	github.com/system-inc/cohere/static_single_assignment v0.0.0

	// A JavaScript interpreter, in a Go linter, on purpose. cohere replaces Prettier, and replacing
	// Prettier means matching it byte for byte on an already-Prettier-formatted tree. typescript-go's
	// own formatter cannot: FormatCodeSettings has no printWidth and no line-breaking engine, and
	// 21.8% of tracked files diverge from our fork after every settings-reachable fix. The only thing
	// that reproduces our formatter is our formatter, and goja is what runs it in-process.
	github.com/dop251/goja v0.0.0-20260822123354-58e940e0d230

	// cohere-dev reads go.work and each module's go.mod with the parser go itself uses, to gate every
	// module the repository builds (command/cohere-dev/workspace.go).
	golang.org/x/mod v0.39.0
	golang.org/x/sys v0.47.0

	// Tests only: the registry's option-decoding guard loads the rules tree with type information,
	// because whether a decode is strict depends on the type it decodes into, which a regex over the
	// source cannot see. It was already in the graph through stringer.
	golang.org/x/tools v0.49.0
)

require (
	github.com/dlclark/regexp2/v2 v2.5.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
