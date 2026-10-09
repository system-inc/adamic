package type_checking

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// CompilerImpliesStrict answers whether the compiler makes this file strict however it is written.
//
// The compiler's `alwaysStrict` binds every TypeScript file as strict and emits the directive into
// output that is not an ES module, which is ESLint's `ecmaFeatures.impliedStrict` (#hks3djf). It is on
// under `strict`, and tsgo has removed `alwaysStrict: false`. A JavaScript file is left out: the option
// governs what the compiler emits, and a JavaScript file the compiler only checks runs as written.
//
// Lifted from `strict`, which asks it to decide whether a script needs a directive, when
// `no-implicit-globals` needed it to decide whether an assignment to an undeclared name can leak.
//
// A run with no program, which is the syntax-only test harness, reads the default options, where an
// unset `strict` is on. So the harness and a repository with an empty tsconfig agree.
func CompilerImpliesStrict(ctx rule.Context, source *ast.SourceFile) bool {
	if ast.IsSourceFileJS(source) {
		return false
	}
	options := &core.CompilerOptions{}
	if ctx.Program != nil {
		options = ctx.Program.Options()
	}
	return IsStrictCompilerOptionEnabled(options, options.AlwaysStrict)
}
