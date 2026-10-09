package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// nativeNonConstructors are the globals that are callable and throw under `new`.
var nativeNonConstructors = map[string]bool{
	"Symbol": true,
	"BigInt": true,
}

// NoNewNativeNonconstructor flags `new Symbol()` and `new BigInt()`.
//
//	valid:   var foo = Symbol('foo');
//	valid:   function bar(Symbol) { var baz = new Symbol('baz'); }
//	valid:   function Symbol() {} new Symbol();
//	valid:   new foo(Symbol);
//	invalid: var foo = new Symbol('foo');
//	invalid: var foo = new BigInt(9007199254740991);
//
// Both are functions that produce a primitive and both throw a TypeError under `new`, so this is a
// crash rather than a style preference: the code never worked and the failure is at the call.
//
// # This is the first hand-written rule to read the checker, and why it must
//
// Six of the ten valid cases above are the same shape: a local binding named `Symbol` or `BigInt`
// that shadows the global. `new Symbol()` inside a function taking `Symbol` as a parameter
// constructs whatever that parameter holds, which may be perfectly constructible, and reporting it
// is a false positive on correct code.
//
// Upstream answers this with `root_unresolved_references()`, which is a scope analysis rather than a
// spelling test. **There is no structural answer.** The shadow can be a parameter, a function
// declaration, an import, or a `var` anywhere in an enclosing scope, so no bounded walk finds it:
// `no-ex-assign` gets away with name matching only because a catch clause gives it a subtree to
// stop at, and this has none.
//
// The predicate is where the name's declaration lives. A global comes from a declaration file, since
// `Symbol` and `BigInt` are declared in the TypeScript standard library; a shadow comes from source.
// That is `IsDeclarationFile` on the declaration's own file, which is the checker's answer to
// upstream's question rather than an approximation of it.
//
// Measured on all ten of upstream's valid cases and all four invalid ones before this was written,
// with a probe reporting which side each landed on.
//
// # Cost
//
// Declaring `NeedsTypeChecker` takes an exclusive per-file lock, and it is the only hand-written
// rule that does. The lock is acquired once per file rather than per node, and this rule asks the
// checker only for a `new` expression whose callee is one of two identifiers, so the question is
// asked on almost no files in a normal tree.
var NoNewNativeNonconstructor = rule.Rule{
	Name: "no-new-native-nonconstructor",

	// See the doc above: the rule's whole discrimination is whether a name is the global or a local
	// shadowing it, and nothing structural answers that.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindNewExpression: func(node *ast.Node) {
				callee := node.AsNewExpression().Expression
				if callee == nil || callee.Kind != ast.KindIdentifier {
					return
				}
				name := callee.Text()
				if !nativeNonConstructors[name] {
					return
				}
				if !resolvesToAGlobal(ctx, callee) {
					return
				}

				ctx.ReportNode(callee, rule.Message{
					Id: "noNewNativeNonconstructor",
					Description: fmt.Sprintf(
						"This calls `new %s()`. %s is a function that returns a primitive and "+
							"throws a TypeError when constructed, so this line fails at runtime "+
							"every time it is reached. Call it without `new`.", name, name),
				})
			},
		}
	},
}

// resolvesToAGlobal reports whether an identifier names a global rather than a local shadowing it.
//
// The question is which file declares the name. A global is declared in the TypeScript standard
// library, which is a declaration file; any binding written in source is a shadow, whatever its
// kind. That covers the parameter, the function declaration, and the `var` forms with one test
// rather than one arm each.
//
// An identifier the checker cannot resolve answers false. That is the conservative direction: a name
// nothing declares is not provably the global, and reporting on it would be a guess.
func resolvesToAGlobal(ctx rule.Context, identifier *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	declaringFile := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return declaringFile != nil && declaringFile.IsDeclarationFile
}
