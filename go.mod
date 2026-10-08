module github.com/system-inc/adamic

go 1.27

// Stage 0 reads the checker's proven types in process, through the shims cohere generates over
// typescript-go's internals. They live in the cohere submodule: one checker for both, pinned
// to cohere's TypeScript submodule commit (see internal/load/pin_test.go), with nothing copied.
replace (
	github.com/microsoft/TypeScript/tsc => ./cohere/TypeScript/tsc
	github.com/system-inc/cohere => ./cohere
	github.com/microsoft/TypeScript/tsc/shim/format => ./cohere/TypeScript-shim/format
	github.com/microsoft/TypeScript/tsc/shim/incremental => ./cohere/TypeScript-shim/incremental
	github.com/microsoft/TypeScript/tsc/shim/parser => ./cohere/TypeScript-shim/parser
	github.com/system-inc/cohere/static_single_assignment => ./cohere/static_single_assignment
	github.com/system-inc/cohere/mutation_aliasing => ./cohere/mutation_aliasing
	github.com/microsoft/TypeScript/tsc/shim/ast => ./cohere/TypeScript-shim/ast
	github.com/microsoft/TypeScript/tsc/shim/bundled => ./cohere/TypeScript-shim/bundled
	github.com/microsoft/TypeScript/tsc/shim/checker => ./cohere/TypeScript-shim/checker
	github.com/microsoft/TypeScript/tsc/shim/compiler => ./cohere/TypeScript-shim/compiler
	github.com/microsoft/TypeScript/tsc/shim/core => ./cohere/TypeScript-shim/core
	github.com/microsoft/TypeScript/tsc/shim/locale => ./cohere/TypeScript-shim/locale
	github.com/microsoft/TypeScript/tsc/shim/parser => ./cohere/TypeScript-shim/parser
	github.com/microsoft/TypeScript/tsc/shim/scanner => ./cohere/TypeScript-shim/scanner
	github.com/microsoft/TypeScript/tsc/shim/tsoptions => ./cohere/TypeScript-shim/tsoptions
	github.com/microsoft/TypeScript/tsc/shim/tspath => ./cohere/TypeScript-shim/tspath
	github.com/microsoft/TypeScript/tsc/shim/vfs => ./cohere/TypeScript-shim/vfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs => ./cohere/TypeScript-shim/vfs/cachedvfs
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs => ./cohere/TypeScript-shim/vfs/osvfs
)

require (
	github.com/dlclark/regexp2/v2 v2.5.2
	github.com/system-inc/cohere v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/ast v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/bundled v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/checker v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/compiler v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/core v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/locale v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/parser v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/scanner v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tsoptions v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/tspath v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs v0.0.0
	github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs v0.0.0
)
