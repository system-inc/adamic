package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoAsyncClientComponent = rule.Message{
	Id: "noAsyncClientComponent",
	Description: "A client component declared as an async function is rendered on the client, " +
		"where React cannot await it the way the server can. The promise it returns is not an " +
		"element, so rendering fails or hydration diverges from what the server produced. Move " +
		"the awaiting into an effect or an event handler and keep the component itself synchronous.",
}

// NoAsyncClientComponent flags the default export of a `use client` file when it is an async
// function with a capitalized name.
//
//	valid:   "use client"; export default function MyComponent() {}
//	valid:   "use client"; export default async function myFunction() {}
//	valid:   export default async function MyComponent() {}
//	invalid: "use client"; export default async function MyComponent() {}
//	invalid: "use client"; async function MyComponent() {}; export default MyComponent
//	invalid: "use client"; const MyComponent = async () => {}; export default MyComponent
//
// Ported from `@next/next/no-async-client-component`, keying on oxc's implementation.
//
// # The rule does not test componenthood, despite its name
//
// Three of upstream's five failing cases return a string and contain no markup at all, and they
// report. The entire discrimination is a capitalized name bound to an async function that is the
// file's default export, in a file whose prologue carries `use client`. Reaching for a component
// predicate such as `react.EnclosingComponent` here would look like an improvement and would
// silently drop three of the five imported failures.
//
// # What the gate is
//
// A directive, not a path. Several rules in this package gate on which Next.js file a path is, and
// `internal/utilities/nextjs/paths.go` holds two deliberately disagreeing predicates for that question.
// Neither applies here: oxc's `run_once` reads `program.directives` and never consults the path, so
// this rule runs on every file and decides from the prologue alone. Recorded because the shape of
// the neighbours invites the wrong assumption.
//
// # Capitalization is Unicode-aware here, unlike upstream's own react rules
//
// oxc tests `char::is_uppercase`, which accepts non-ASCII. Pinned against the release oxlint
// binary: a client component named `Фoo` reports and one named `фoo` is silent. That makes
// `react.IsLikelyComponentName`, which decodes the first rune and asks `unicode.IsUpper`, an exact
// match for this rule rather than the divergence it is elsewhere. The helper's own doc records an
// open ASCII question because oxc's *component* path uses `is_ascii_uppercase`; this rule is not on
// that path, so the question does not arise and the helper is simply right. The `@next` original
// spells it `/[A-Z]/` and is ASCII-only, so it and oxc disagree on this input. oxc is the port
// target and wins.
//
// # Parenthesized forms are silent, and that is reproduced rather than repaired
//
// oxc destructures an arrow initializer and an exported identifier directly, with no parenthesis
// skipping, so both of these are silent upstream and both are silent here, for free rather than by
// a check. Measured on the release binary, since the corpus writes no parenthesized forms:
//
//	"use client"; const MyThing = (async () => {}); export default MyThing     silent
//	"use client"; async function MyComponent() {}; export default (MyComponent) silent
//
// Our parser gives the first a `KindParenthesizedExpression` initializer, which the kind check
// declines, and the second a parenthesized export expression the checker resolves to no symbol.
// Adding a skip in either place would be a divergence in the direction of reporting more.
//
// No fix. The repair is to remove `async` and rewrite whatever the body awaited, which the rule
// cannot know how to do.
var NoAsyncClientComponent = rule.Rule{
	Name: "@next/next/no-async-client-component",

	// The indirect shapes resolve an exported identifier back to its declaration, which upstream
	// does through `get_declaration_of_variable`. Measured: a block-scoped async `MyComponent`
	// shadowed by a top-level synchronous one is silent upstream, so a name-matching scan over
	// top-level statements, which is what the `@next` original does, is not the same rule.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole rule is one pass over the top-level statements, matching upstream's
			// `run_once`. A `KindSourceFile` listener fires before its children, so the walk is
			// done here rather than through per-kind listeners that would also see nested
			// declarations upstream never looks at.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				sourceFile := node.AsSourceFile()
				if !nextjs.HasFileDirective(sourceFile, "use client") {
					return
				}

				for _, statement := range sourceFile.Statements.Nodes {
					switch statement.Kind {
					case ast.KindFunctionDeclaration:
						// `export default async function MyComponent() {}`. The parser folds the
						// export into the declaration's modifiers rather than wrapping it, so this
						// is a function declaration carrying the default modifier and not an
						// export assignment.
						reportDirectAsyncDefaultExport(ctx, statement)
					case ast.KindExportAssignment:
						// `export default MyComponent`, where the name is declared elsewhere.
						reportIndirectAsyncDefaultExport(ctx, statement)
					}
				}
			},
		}
	},
}

// reportDirectAsyncDefaultExport handles `export default async function MyComponent() {}`.
//
// An anonymous default export is silent because there is no name to test, which is upstream's
// behavior rather than an omission: its capitalization check reads through `func_decl.id`, and a
// missing id fails the `is_some_and`.
func reportDirectAsyncDefaultExport(ctx rule.Context, statement *ast.Node) {
	if statement.ModifierFlags()&ast.ModifierFlagsDefault == 0 {
		return
	}
	if !ast.IsAsyncFunction(statement) {
		return
	}
	name := statement.AsFunctionDeclaration().Name()
	if name == nil || !react.IsLikelyComponentName(name.Text()) {
		return
	}
	// The identifier, not the statement. oxc labels `func_decl.id.span` and the `@next` original
	// reports the whole export declaration; the snapshot pins oxc's narrower span and that is the
	// one ported.
	ctx.ReportNode(name, messageNoAsyncClientComponent)
}

// reportIndirectAsyncDefaultExport handles `export default MyComponent`, resolving the name.
//
// Upstream reads two declaration shapes here, an async function declaration and a variable
// declarator holding an async arrow, and it writes them as two sequential blocks with no `continue`
// between them. That reads as a double-report waiting to happen and is not one: a declaration
// cannot be both kinds at once, so the second block's binding always fails when the first fired.
// Written as a switch here, which says the same thing without the dead sequencing.
func reportIndirectAsyncDefaultExport(ctx rule.Context, statement *ast.Node) {
	expression := statement.AsExportAssignment().Expression
	if expression == nil || expression.Kind != ast.KindIdentifier {
		return
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(expression)
	if symbol == nil {
		return
	}

	// The declaration that comes FIRST IN SOURCE ORDER, which is neither a loop nor an index-zero
	// read, and every part of this was measured rather than reasoned.
	//
	// The standing advice in this tree is to loop over `symbol.Declarations` because indexing zero
	// goes silent when declarations merge. That advice is right for a rule anchored on a
	// declaration asking "does any write reach me", where more declarations can only mean more
	// chances to match. It is wrong here, where the question is "what single thing is this name",
	// and it is wrong in both of its halves.
	//
	// First, declarations are NOT in source order in typescript-go. Measured: an
	// `interface MyComponent {}` written above `async function MyComponent() {}` still puts the
	// function at index 0, because a value declaration sorts ahead of a type one no matter where it
	// was written. So the input the loop advice exists to catch does not arise for this shape at
	// all, and a fixture written for it asserts nothing. That is why a mutant replacing the loop
	// with an index-zero read survived a fixture written specifically to kill it.
	//
	// Second, upstream reaches exactly one node, through `symbol_declaration(symbol_id)`, and which
	// node that is depends on source order: oxc's binder records the first declaration it sees. So
	// the rule is order-sensitive upstream, and all four of these were pinned on the release
	// binary:
	//
	//	interface MyComponent {}; async function MyComponent() {}   silent   the interface is first
	//	async function MyComponent() {}; interface MyComponent {}   REPORTS  the function is first
	//	namespace MyComponent {}; async function MyComponent() {}   silent   the namespace is first
	//	let MyThing = async () => {}; namespace MyThing {}          REPORTS  the binding is first
	//
	// A loop reports all four, because one merged declaration is always the async one. An
	// index-zero read reports the first three and is right twice by accident. Sorting by position
	// and taking the earliest reproduces upstream on all four, for the same reason upstream gives
	// rather than by coincidence.
	declaration := earliestDeclaration(symbol.Declarations)
	if declaration == nil {
		return
	}

	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		if !ast.IsAsyncFunction(declaration) {
			return
		}
		name := declaration.AsFunctionDeclaration().Name()
		if name == nil || !react.IsLikelyComponentName(name.Text()) {
			return
		}
		ctx.ReportNode(name, messageNoAsyncClientComponent)
		return
	case ast.KindVariableDeclaration:
		variableDeclaration := declaration.AsVariableDeclaration()
		name := variableDeclaration.Name()
		// No binding-identifier kind check here, and its absence is deliberate.
		//
		// Upstream destructures a `BindingPattern::BindingIdentifier` at this point, so writing the
		// equivalent kind guard is the obvious port and it is inert. Probed rather than reasoned: a
		// name bound by a destructuring pattern, whether from an object or an array, resolves to a
		// `KindBindingElement` and never to a `KindVariableDeclaration`, so the switch arm above has
		// already declined every input that guard could have seen. A mutant deleting it survived the
		// whole fixture set, and the reason was that nothing can reach it rather than that nothing
		// covers it.
		//
		// The nil check stays, since a declaration with no name is a different question from a
		// declaration whose name is a pattern.
		if name == nil {
			return
		}
		if !react.IsLikelyComponentName(name.Text()) {
			return
		}
		// An arrow specifically. An async *function expression* assigned to the same name is
		// silent upstream, pinned on the release binary, and declining it is a kind check
		// rather than an oversight.
		initializer := variableDeclaration.Initializer
		if initializer == nil || initializer.Kind != ast.KindArrowFunction {
			return
		}
		if !ast.IsAsyncFunction(initializer) {
			return
		}
		ctx.ReportNode(name, messageNoAsyncClientComponent)
		return
	}
}

// earliestDeclaration returns the declaration with the lowest source position, or nil for none.
//
// This stands in for oxc's `symbol_declaration(symbol_id)`, which yields the one node its binder
// recorded for a symbol, and the binder records the first one it walks past. typescript-go instead
// keeps every merged declaration and orders them by kind rather than by position, so recovering
// upstream's answer means asking for the earliest rather than taking the list's head.
//
// Linear rather than a sort because a merged symbol carries two or three declarations in practice
// and allocating to order them would cost more than the scan.
func earliestDeclaration(declarations []*ast.Node) *ast.Node {
	var earliest *ast.Node
	for _, declaration := range declarations {
		if earliest == nil || declaration.Pos() < earliest.Pos() {
			earliest = declaration
		}
	}
	return earliest
}
