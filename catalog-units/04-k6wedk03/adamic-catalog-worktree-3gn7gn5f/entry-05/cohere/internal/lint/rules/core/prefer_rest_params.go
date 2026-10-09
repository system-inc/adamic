package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferRestParams flags reads of the implicit `arguments` object inside a function.
//
//	valid:   function foo(...args) { args; }
//	valid:   function foo(arguments) { arguments; }
//	valid:   function foo() { var arguments; arguments; }
//	valid:   arguments;
//	valid:   var foo = () => arguments;
//	valid:   function foo() { arguments.length; }
//	invalid: function foo() { arguments; }
//	invalid: function foo() { arguments[0]; }
//	invalid: function foo() { function bar() { arguments; } }
//
// `arguments` is array-like rather than an array, so it carries none of `Array.prototype`. Code
// that wants to iterate, slice, or map it has to launder it through
// `Array.prototype.slice.call(arguments)` first, and a rest parameter is a real array at the point
// of declaration with no conversion and no aliasing surprises.
//
// # Why this reads the checker
//
// The rule's whole discrimination is whether a given `arguments` is the implicit binding or a name
// somebody declared, and that is name resolution rather than a scope flag. Upstream asks
// `root_unresolved_references()`, which is the same question phrased for a semantic layer we do not
// have.
//
// There is no bounded syntactic answer. The shadow can be a parameter, a `var` in any enclosing
// block, a `let` in a nested block, a catch clause binding, an import, or a declaration in an outer
// function, and a walk would have to reimplement each arm. It would also have to know that the
// implicit binding is *reintroduced* at every non-arrow function boundary, so an outer function's
// parameter named `arguments` does not shadow an inner function's implicit one, while a catch
// clause in the same function does.
//
// This was determined by probe rather than by reading upstream's call site. Measured on all
// thirteen upstream cases plus five shadow-reach cases before this was written, the checker
// reproduces every one of those distinctions, including the arrow cases, for free.
//
// # The predicate, and why resolvesToAGlobal is not it
//
// `resolvesToAGlobal` is the shipped helper for "is this the global or a local shadowing it", and
// it is the wrong test here. The implicit `arguments` is not declared in the standard library: the
// checker resolves it to a symbol carrying **zero declarations**, so `resolvesToAGlobal` answers
// false on exactly the case this rule must report and the rule would go silent. A shadow, by
// contrast, always carries a source declaration. The predicate is therefore the inverse: a symbol
// with no declarations at all is the implicit binding.
//
// A nil symbol is not a finding. That is what an `arguments` with no enclosing non-arrow function
// resolves to, which is upstream's first valid case and the top-level arrow.
//
// # What this deliberately does not catch
//
// Normal member access, `arguments.length` and `arguments.callee`, is permitted by both upstream
// implementations. Reading a property off the object is not the array-like abuse the rule targets,
// and rewriting it would change what the code means. Computed access is reported, because
// `arguments[0]` is indexing the array-like directly.
//
// # No fix
//
// Neither upstream implementation offers one, and neither should. The repair is to add a rest
// parameter to the enclosing function and rewrite every use, which changes the signature, is not a
// single-span edit, and has to pick a parameter name. Under this tree's rule that is not even a
// suggestion: it is a change of meaning at a site the rule does not own.
var PreferRestParams = rule.Rule{
	Name: "prefer-rest-params",

	// See the doc above: telling the implicit binding from a declared one is name resolution, and
	// the probe that established it is recorded there.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if node.Text() != "arguments" {
					return
				}
				if !isImplicitArgumentsBinding(ctx, node) {
					return
				}
				if isNormalMemberAccess(node) {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: "preferRestParams",
					Description: "This reads the implicit `arguments` object, which is array-like " +
						"rather than an array and carries none of `Array.prototype`, so iterating " +
						"or slicing it needs a conversion first. Declare a rest parameter " +
						"(`...args`) and use that instead.",
				})
			},
		}
	},
}

// isImplicitArgumentsBinding reports whether an identifier spelled `arguments` is the binding every
// non-arrow function gets for free, rather than a name somebody declared.
//
// The checker answers this directly and the answer is a declaration count. A declared `arguments`,
// whatever its kind and wherever in the enclosing scopes it sits, resolves to a symbol holding that
// declaration. The implicit binding resolves to a symbol holding none, because nothing in source or
// in the standard library declares it. An identifier the checker cannot resolve at all answers
// false: that is `arguments` with no enclosing non-arrow function to supply it, which is not a
// finding.
func isImplicitArgumentsBinding(ctx rule.Context, identifier *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}
	return len(symbol.Declarations) == 0
}

// isNormalMemberAccess reports whether an identifier is the object of a dotted property read.
//
// `arguments.length` is normal member access and upstream permits it. `arguments[0]` is computed
// access and upstream reports it, so element access is not this. The identifier also has to be the
// *object* rather than the property name, since `foo.arguments` reads a property that happens to
// share the spelling.
func isNormalMemberAccess(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	return parent.AsPropertyAccessExpression().Expression == identifier
}
