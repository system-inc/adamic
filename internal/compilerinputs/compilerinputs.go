// Package compilerinputs names the repository paths the compiler a port is lowered and built by is made of: every
// in-repository package internal/load, internal/lower, internal/native and internal/javascript link, and each module
// file that pins one. A test that lowers a port links that compiler into its own binary, so no trace of a build can see
// it read: a product the compiler makes has to name it in its key, and Lowering is the one list every such key names.
//
// From cohere/TypeScript/tsc it names go.mod, go.sum and the package directories linked, never tsc whole: tsc is 66,741
// of the submodule's 67,144 files, baselines and fourslash tests no compiler links, which a populated tree hashed for
// every lookup (#4ycgm2n). A nested package sits under its parent's entry. TestLoweringIsWhatTheCompilerLinks keeps the
// list complete and every entry needed.
package compilerinputs

import "slices"

var lowering = []string{
	"internal", "go.mod", "go.work", "bridge/tsgo", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript-shim",
	"cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum", "cohere/TypeScript/tsc/internal/ast",
	"cohere/TypeScript/tsc/internal/astnav", "cohere/TypeScript/tsc/internal/binder",
	"cohere/TypeScript/tsc/internal/bundled", "cohere/TypeScript/tsc/internal/checker",
	"cohere/TypeScript/tsc/internal/collections", "cohere/TypeScript/tsc/internal/compiler",
	"cohere/TypeScript/tsc/internal/contentmapper", "cohere/TypeScript/tsc/internal/core",
	"cohere/TypeScript/tsc/internal/debug", "cohere/TypeScript/tsc/internal/diagnostics",
	"cohere/TypeScript/tsc/internal/evaluator", "cohere/TypeScript/tsc/internal/glob",
	"cohere/TypeScript/tsc/internal/ipc", "cohere/TypeScript/tsc/internal/jsnum", "cohere/TypeScript/tsc/internal/json",
	"cohere/TypeScript/tsc/internal/jsonrpc", "cohere/TypeScript/tsc/internal/locale",
	"cohere/TypeScript/tsc/internal/module", "cohere/TypeScript/tsc/internal/modulespecifiers",
	"cohere/TypeScript/tsc/internal/nativepath", "cohere/TypeScript/tsc/internal/nodebuilder",
	"cohere/TypeScript/tsc/internal/osutil", "cohere/TypeScript/tsc/internal/outputpaths",
	"cohere/TypeScript/tsc/internal/packagejson", "cohere/TypeScript/tsc/internal/parser",
	"cohere/TypeScript/tsc/internal/printer", "cohere/TypeScript/tsc/internal/pseudochecker",
	"cohere/TypeScript/tsc/internal/scanner", "cohere/TypeScript/tsc/internal/semver",
	"cohere/TypeScript/tsc/internal/sourcemap", "cohere/TypeScript/tsc/internal/spanmap",
	"cohere/TypeScript/tsc/internal/stringutil", "cohere/TypeScript/tsc/internal/symlinks",
	"cohere/TypeScript/tsc/internal/tracing", "cohere/TypeScript/tsc/internal/transformers",
	"cohere/TypeScript/tsc/internal/tsoptions", "cohere/TypeScript/tsc/internal/tspath",
	"cohere/TypeScript/tsc/internal/vfs",
	"cohere/internal/lint/checking", "cohere/internal/lint/ecmascript/binding", "cohere/internal/lint/ecmascript/comments",
	"cohere/internal/lint/ecmascript/control_flow_graph", "cohere/internal/lint/ecmascript/imports", "cohere/internal/lint/ecmascript/jsx",
	"cohere/internal/lint/ecmascript/module", "cohere/internal/lint/ecmascript/module_roots", "cohere/internal/lint/ecmascript/nextjs",
	"cohere/internal/lint/ecmascript/property", "cohere/internal/lint/ecmascript/react", "cohere/internal/lint/ecmascript/reference",
	"cohere/internal/lint/ecmascript/regexsyntax", "cohere/internal/lint/ecmascript/scope", "cohere/internal/lint/ecmascript/text",
	"cohere/internal/lint/linter", "cohere/internal/lint/rule", "cohere/internal/lint/rules/adamic", "cohere/internal/lint/rules/nexus",
	"cohere/internal/types/sourcename", "cohere/policy", "cohere/rule_runner",
}

// Lowering is the compiler's repository paths, relative to the repository root, as a key's Files names them: a fresh
// copy each call, so a caller appending its own files changes no one else's.
func Lowering() []string {
	return slices.Clone(lowering)
}
