package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnsafeDeclarationMerging = rule.Message{
	Id: "unsafeMerging",
	Description: "A class and an interface sharing a name are merged into one type, and the " +
		"members the interface contributes are never initialized by the class constructor. " +
		"TypeScript does not check them, so reading one compiles and returns undefined at " +
		"runtime. Give the interface a different name, or declare the members on the class.",
}

// NoUnsafeDeclarationMerging flags a class and an interface declared with the same name in the same
// scope.
//
//	valid:   interface Foo {}; class Bar implements Foo {}
//	valid:   namespace Foo {}; namespace Foo {}
//	valid:   interface Foo {}; { class Foo {} }
//	invalid: interface Foo {}; class Foo {}
//	invalid: class Foo {}; interface Foo {}
//
// Ported from oxc's `no_unsafe_declaration_merging`, which is what the gate runs.
//
// # Which merge is unsafe
//
// Only class with interface, which the message explains and which the corpus pins from both sides:
// `namespace Foo {}` twice passes, `enum Foo {}` beside `namespace Foo {}` passes, and
// `namespace Fooo {}` beside `function Foo() {}` passes. Measured on the release oxlint binary
// rather than inferred, an interface beside a function is silent and a namespace beside a class is
// silent, so the pairing is exactly the two kinds and nothing else. Interface beside interface is
// silent too, which is the intended TypeScript feature.
//
// # Why this reads only the FIRST declaration, and why that is a loop rather than an index
//
// Upstream reaches exactly one node. `ctx.scoping().get_binding(scope, name)` yields the symbol and
// `symbol_declaration(symbol_id)` yields the single declaration oxc's binder recorded first, which
// is the first one in source. The rule then reports only when that first declaration is of the
// opposite kind to the node being visited. That is not an implementation detail it would be safe to
// improve on, because it decides real verdicts, and all of these were pinned on the release binary:
//
//	interface Foo {}; class Foo {}                        1 finding
//	class Foo {}; interface Foo {}                        1 finding
//	interface Foo {}; interface Foo {}; class Foo {}      1 finding   both interfaces decline
//	class Foo {}; interface Foo {}; interface Foo {}      2 findings  both interfaces fire
//	namespace Foo {}; interface Foo {}; class Foo {}      0 findings  the namespace is first
//	import { Foo } from './x'; interface Foo {}; class Foo {}   0 findings  the import is first
//
// The last two are upstream going silent on a genuinely unsafe merge, because the first declaration
// is neither a class nor an interface and so neither arm matches. That is reproduced here rather
// than corrected: the gate runs oxc, so improving on it would read as a differential difference for
// no gain. It is worth naming as a real upstream limitation rather than a quirk of the encoding.
//
// # Where typescript-eslint disagrees, and it is not a small disagreement
//
// The original asks `defs.some(def => def.node.type === unsafeKind)`, which examines EVERY
// definition rather than the first one. So the two implementations part company on exactly the
// inputs above where oxc goes silent: under typescript-eslint,
// `namespace Foo {}; interface Foo {}; class Foo {}` reports twice, once from each arm, because
// each arm finds an opposite-kind definition somewhere in the list. Under oxc it reports zero
// times. oxc is what the gate runs, so oxc is what this reproduces, and the difference is recorded
// here rather than silently resolved. Reading only the original would produce a rule that is
// arguably more correct and reads to the differential harness as a defect.
//
// The original also declares `requiresTypeChecking: false`. That is not a disagreement about
// semantics, only about what each host hands the rule: ESLint supplies a scope table without a
// program, and typescript-go populates symbols only when one is built. See the checker note below.
//
// The standing advice in this tree is never to index `symbol.Declarations[0]` and to loop instead,
// because merging can put an unexpected declaration at index zero. Both halves of that matter here
// and they point in different directions, so this was measured rather than assumed. Probed in
// typescript-go on this rule's own shapes, `Declarations` IS in strict source-position order for
// every one of them, including the two above where the surprising entry (a module declaration, an
// import specifier) lands at index zero exactly as its source position says. So an index-zero read
// would in fact reproduce upstream here.
//
// It is still written as a loop taking the lowest position, for two reasons. The ordering is a
// property of the binder that no test in this package guards, and a neighbouring rule measured the
// opposite behavior for a different shape: `no_async_client_component` found a function sorting
// ahead of an interface written above it, because a value declaration can outrank a type one. Both
// findings are real and they are about different shapes, which is precisely why neither should be
// relied on implicitly. Asking for the earliest position states the requirement in the code, costs
// one pass over a slice that is almost always length two, and cannot silently change meaning if the
// binder's ordering does.
//
// # Scope
//
// Upstream binds the name in the node's own scope, so a class in a nested block does not merge with
// an interface outside it, and the corpus carries three cases for this: a class returned from a
// function body, a class inside an immediately-invoked function, and an interface inside
// `declare global` beside a class outside it. All three pass. Nothing in this rule implements that:
// typescript-go's binder already scopes the symbols, so a block-scoped `Foo` resolves to a symbol
// carrying one declaration and the opposite-kind test simply finds nothing. Probed, not assumed.
//
// # The checker
//
// Required, and measured rather than reasoned. Under the untyped harness `node.Symbol()` is nil on
// every one of these inputs, because symbol tables are populated when the program is built. So this
// declares NeedsTypeChecker, its fixtures use RunTyped, and both listeners guard against a nil
// checker: a rule that dereferences a checker result without that guard panics rather than going
// quiet.
var NoUnsafeDeclarationMerging = rule.Rule{
	Name:             "@typescript-eslint/no-unsafe-declaration-merging",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node, opposite ast.Kind) {
			// The nil-checker guard, kept even though it is inert on today's shim, and the
			// measurement is here because the standing advice says a missing one PANICS.
			//
			// Probed rather than believed: under the untyped harness `ctx.TypeChecker` is nil and an
			// unguarded `GetSymbolAtLocation` on it returns nil rather than crashing, so a mutant
			// neutralizing this guard survived the whole fixture set. That makes the guard's value
			// today entirely about not depending on that: nil tolerance is a property of the shim's
			// receiver, not a promise, and the failure it would produce if it changed is a crash
			// rather than a quiet wrong answer. It also keeps a `.TypeChecker` selector in this
			// file, which is what the registry's declaration guard reads.
			if ctx.TypeChecker == nil {
				return
			}
			name := node.Name()
			// An anonymous `export default class {}` is the one declaration of either kind that can
			// have no name, probed rather than assumed: an interface always carries one, and a class
			// expression is a different node kind. It binds nothing, so it merges with nothing.
			if name == nil {
				return
			}
			symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
			if symbol == nil {
				return
			}

			// An EXPORTED declaration is handed its own symbol carrying only itself, while the
			// merged list lives on its local symbol. This is the single most surprising thing about
			// porting this rule and no imported fixture can see it, because upstream's corpus writes
			// no `export` anywhere. Probed on three shapes:
			//
			//	interface Foo {}; export class Foo {}
			//	  visiting the interface   GetSymbolAtLocation holds both, LocalSymbol is nil
			//	  visiting the class       GetSymbolAtLocation holds only the class,
			//	                           and LocalSymbol holds both
			//
			// So `GetSymbolAtLocation` alone goes silent whenever the merged partner is exported and
			// written second, which is three of the seven export shapes measured against the release
			// binary, all of which upstream reports. oxc does not have this problem because its scope
			// binding hands every declaration one symbol.
			//
			// `LocalSymbol` is nil on every non-exported declaration, so this is a fallback rather
			// than a replacement and it cannot broaden anything it is not needed for.
			//
			// The length comparison is not load-bearing and is kept for what it states rather than
			// for what it decides. A mutant dropping it survived, and the reason is measured: over
			// sixteen shapes, `LocalSymbol` when non-nil holds either the same merged list or a
			// strict superset, never a smaller one, so taking it unconditionally reads identical
			// declarations. The test says "only widen", which is the property the fallback depends
			// on and which nothing else in the file would reveal if it stopped holding.
			declarations := symbol.Declarations
			if local := node.LocalSymbol(); local != nil && len(local.Declarations) > len(declarations) {
				declarations = local.Declarations
			}

			// Upstream consults exactly ONE declaration, the earliest, and reports only when that
			// one is of the opposite kind. That single test carries three behaviors at once, all
			// pinned on the release binary:
			//
			//	interface Foo {}; class Foo {}                      1  the class sees an interface
			//	interface Foo {}; interface Foo {}; class Foo {}    1  both interfaces see an interface
			//	class Foo {}; interface Foo {}; interface Foo {}    2  both interfaces see a class
			//	namespace Foo {}; interface Foo {}; class Foo {}    0  everything sees a namespace
			//
			// The last is upstream going quiet on a genuinely unsafe merge, reproduced rather than
			// corrected. A predicate asking whether ANY declaration is the opposite kind reads as an
			// improvement and gets the second and fourth lines wrong.
			//
			// This is a loop rather than `declarations[0]`, and the standing advice to loop is right
			// here for a reason that had to be measured rather than assumed. Probed over eighteen
			// class-and-interface shapes in typescript-go, `Declarations` IS in strict source-position
			// order, including the two where a third kind lands at index zero, so an index read would
			// in fact agree today. It stays a loop because a neighbouring rule measured the opposite
			// for its own shape, where a value declaration sorted ahead of a type one written above
			// it. Both findings are real and about different shapes, so neither ordering should be
			// relied on implicitly. Asking for the lowest position states the requirement in code and
			// cannot silently change meaning if the binder's ordering does.
			earliest := earliestDeclaration(declarations)
			if earliest == nil || earliest.Kind != opposite {
				return
			}

			// The span. oxc attaches two labels and sorts them, so its primary is the declaration
			// written first, which is the opposite-kind one the arm just consulted rather than the
			// node being visited. Reporting the visited node is the natural way to write this and
			// puts every finding on the wrong declaration, which no message-id fixture can see.
			// No nil check on the name here, and its absence is deliberate. `earliest.Kind` has
			// already been narrowed to exactly the opposite kind, so it is a class declaration or an
			// interface declaration, and reaching this line at all required resolving a symbol
			// through a name. Probed over sixteen shapes including every export arrangement and both
			// anonymous default forms: no merged declaration of either kind ever has a nil name. A
			// guard written here survived the whole fixture set because nothing can reach it, which
			// is a different thing from nothing covering it.
			ctx.ReportNode(earliest.Name(), messageUnsafeDeclarationMerging)
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				report(node, ast.KindInterfaceDeclaration)
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				report(node, ast.KindClassDeclaration)
			},
		}
	},
}

// earliestDeclaration returns the declaration with the lowest source position, or nil for none.
//
// This states in code what upstream gets from its binder: the declaration recorded first is the one
// written first. See the ordering note on the rule for why this is asked rather than assumed.
func earliestDeclaration(declarations []*ast.Node) *ast.Node {
	var earliest *ast.Node
	for _, declaration := range declarations {
		if earliest == nil || declaration.Pos() < earliest.Pos() {
			earliest = declaration
		}
	}
	return earliest
}
