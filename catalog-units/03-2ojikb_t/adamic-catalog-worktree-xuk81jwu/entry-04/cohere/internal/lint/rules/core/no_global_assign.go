package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoGlobalAssignOptions is the decoded option surface.
//
// The inventory column said this rule takes no options and that is wrong. oxc carries
// `NoGlobalAssignConfig { exceptions: Vec<CompactStr> }` and ESLint carries the matching
// `schema: [{ properties: { exceptions: { type: "array", items: { type: "string" } } } }]`, both
// exercised by the same pass case: `Object = 0;` with `{"exceptions": ["Object"]}`.
type NoGlobalAssignOptions struct {
	// Exceptions names globals that may be assigned to. A project that deliberately polyfills or
	// stubs a builtin lists it here rather than turning the rule off everywhere.
	Exceptions []string `json:"exceptions"`
}

// NoGlobalAssign flags an assignment to a read-only global.
//
//	valid:   string = '1';
//	valid:   var string;
//	valid:   window[parseInt('42', 10)] = 99;
//	valid:   function f(Object) { Object = 1; }
//	valid:   Object.x = 1;
//	valid:   foo(Object);
//	invalid: String = 'hello world';
//	invalid: String++;
//	invalid: ({Object = 0, String = 0} = {});
//	invalid: function f() { Object = 1; }
//	invalid: Array = 1;
//
// Assigning to `Object`, `String`, or `Array` replaces the binding every other piece of code on the
// page reaches through. It does not fail at the line that does it: it fails later and somewhere
// else, when a library calls `Object.keys` and gets whatever was written instead, or when
// `x instanceof Array` answers about a different object. In a module or a class body the write
// throws outright, which is the better outcome, because the silent case takes the rest of the
// program down with it in a way that is very hard to trace back here.
//
// # Why this reads the checker
//
// The discrimination is name resolution and nothing structural answers it. `Object = 1` and
// `function f(Object) { Object = 1; }` are the same six characters at the write, and the first is a
// defect while the second is ordinary code shadowing a name it happens to reuse. The shadow can be a
// parameter, a `let`, a `var`, a class, a function declaration, an import, or a catch parameter,
// anywhere in an enclosing scope, so no bounded walk finds it. `no-ex-assign` gets away with name
// matching only because a catch clause hands it a subtree to stop at, and this rule's scope is the
// whole file.
//
// The predicate is which file declares the name. `Object`, `String`, `Math`, `JSON`, `Promise`,
// `NaN` and the rest are declared in the TypeScript standard library, which is a declaration file;
// anything bound in source is a shadow, whatever its kind. That is one test rather than one arm per
// shadow shape, and it is `resolvesToAGlobal`, already shipped for `no-new-native-nonconstructor`.
//
// # Upstream's corpus cannot see the bug this guards against
//
// Upstream iterates `root_unresolved_references()`, and a shadowed name *resolves*, so it never
// enters the iteration at all. The shadow filter is the choice of collection rather than a test,
// which is why upstream's eight clean cases contain zero shadowing cases: it is testing what the
// collection does not contain. A port matching on the name alone therefore passes every visible
// clean case and still reports `function f(Object) { Object = 1; }` on real code. The fixtures write
// that case and eleven more like it, because the absent test is the hazard here rather than a
// present one.
//
// # What this port does not carry, and why
//
// Upstream answers "is this a read-only global" from a vendored `javascript_globals` table selected
// by an `env` config (`browser`, `node`) and overridden by a user `globals` map. Neither table nor
// either config surface exists in this tree. Five of upstream's sixteen cases turn on nothing else,
// and three of those five share their source text with a case on the other side, separated only by
// the lint configuration:
//
//	top = 0;         clean with no env, reported under `env: {browser: true}`
//	require = 0;     clean with no env, reported under `env: {node: true}`
//	a = 1            clean under `globals: {a: true}`, reported under `globals: {a: false}`
//
// There is no information in the file that separates them, so this port reports on none of them: a
// name it cannot prove is a read-only global is left alone. That direction costs coverage rather
// than false positives, which is the right way to be short. `TestNoGlobalAssignBoundary` pins it so
// the gap stays a decision rather than becoming a drift, and the standard library covers everything
// upstream's always-on `GLOBALS_BUILTIN` table covers.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names and says nothing about whether the
// occurrence writes. `Object.x = 1;` and `foo(Object);` both resolve to the global and neither
// reassigns it, so the rule needs the structural half too and reports only where the two agree.
// `reference.WritesToBinding` is that half, shared with five sibling rules.
//
// # The one shape the checker answers differently
//
// `({Object = 0, String = 0} = {});` is upstream's two-diagnostic case, and `GetSymbolAtLocation` on
// either identifier returns the *shorthand property's own* symbol, declared in source. That reads
// exactly like a correctly declined shadow, so the plain accessor drops upstream's most interesting
// fail case in the quiet direction. Measured on the corpus rather than assumed: with
// `GetShorthandAssignmentValueSymbol` both names resolve into the standard library, and
// `function f(Object) { ({Object} = {}); }` still resolves to the parameter, so the accessor
// recovers the finding without costing the shadow test.
//
// No fix. The repair is a new binding, which means choosing a name and deciding which later uses of
// the old one meant the global.
var NoGlobalAssign = rule.Rule{
	Name: "no-global-assign",

	// See the doc above: `Object = 1` and `function f(Object) { Object = 1; }` are textually
	// identical at the write and differ only in what the name resolves to.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// An absent or malformed option block means no exceptions, which is upstream's default
		// (`defaultOptions: [{ exceptions: [] }]`). A config typo must not silently exempt
		// everything, so the failure direction here is "report as usual".
		exempt := map[string]bool{}
		if decoded, ok := rule.OptionsAs[NoGlobalAssignOptions](options); ok {
			for _, name := range decoded.Exceptions {
				exempt[name] = true
			}
		}

		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built,
				// and this rule can answer nothing without one. A mutant removing this survives
				// the fixtures, because typescript-go tolerates a nil receiver on this path and
				// returns no symbol, so the rule goes silent either way. That is an implementation
				// detail of a vendored compiler rather than a promise, so the guard stays and the
				// rule does not depend on it. `no-class-assign` measured the same survivor.
				if ctx.TypeChecker == nil {
					return
				}
				if exempt[node.Text()] {
					return
				}
				// The structural half first, because it is far cheaper than a checker call and this
				// listener sees every identifier in the file. Most of them are reads.
				if !reference.WritesToBinding(node) {
					return
				}
				if !resolvesToAGlobalThroughShorthand(ctx, node) {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: "noGlobalAssign",
					Description: fmt.Sprintf(
						"This assigns to `%s`, which is a read-only global that the whole program "+
							"shares. Nothing fails here: the failure lands later and elsewhere, "+
							"when other code reaches for `%s` and gets whatever was written "+
							"instead. In a module or a class body the write throws instead of "+
							"landing at all. Use a local variable.", node.Text(), node.Text()),
				})
			},
		}
	},
}

// resolvesToAGlobalThroughShorthand asks `resolvesToAGlobal` about the binding a write actually
// targets.
//
// The two differ on exactly one shape. A shorthand property in a destructuring target
// (`({Object} = {})`) resolves through `GetSymbolAtLocation` to the property rather than to the
// value being written, and the property is declared in source, so the plain accessor answers "not a
// global" for a case upstream reports. TypeScript has a separate accessor for the value side and
// reaching for it is the difference between carrying upstream's two-diagnostic case and silently
// dropping it.
//
// Delegating the file test rather than restating it keeps one answer to "is this a global" in the
// package. `no-class-assign` reaches the same accessor for the same shape against a different
// anchor.
func resolvesToAGlobalThroughShorthand(ctx rule.Context, identifier *ast.Node) bool {
	if identifier.Parent == nil || identifier.Parent.Kind != ast.KindShorthandPropertyAssignment {
		return resolvesToAGlobal(ctx, identifier)
	}
	symbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	declaringFile := ast.GetSourceFileOfNode(symbol.Declarations[0])
	return declaringFile != nil && declaringFile.IsDeclarationFile
}
