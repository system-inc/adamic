package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoConstAssign = rule.Message{
	Id: "noConstAssign",
	Description: "This assigns to a binding declared `const`, `using`, or `await using`, which is " +
		"immutable. The assignment throws a TypeError at runtime rather than storing anything, so " +
		"the line does not do what it reads as doing and everything after it in that function is " +
		"skipped. TypeScript reports this too, but only where the file is type checked, and the " +
		"failure is a hard one at the moment it runs. Declare the binding `let` if it needs to " +
		"change, or assign to a new binding.",
}

// NoConstAssign flags an assignment to a binding declared const, using, or await using.
//
//	valid:   const x = 0; foo(x);
//	valid:   const x = {key: 0}; x.key = 1;
//	valid:   const x = 0; { let x; x = 1; }
//	valid:   for (const x of [1,2,3]) { foo(x); }
//	invalid: const x = 0; x = 1;
//	invalid: const x = 0; ++x;
//	invalid: using x = foo(); x ??= bar();
//	invalid: const [a, b, ...[c, ...d]] = [1,2,3,4,5]; d = 123
//
// A `const` binding cannot be rebound, so the write throws a TypeError the moment it runs. That is
// worse than it sounds in a place the rule is most useful: the throw aborts the rest of the
// enclosing function, so a line that reads like a harmless counter bump takes out everything after
// it, and in code that only runs on an uncommon branch it ships.
//
// `using` and `await using` are included, matching both upstream ports. Their bindings are const in
// exactly the same way, and they are newer, so the mistake is likelier rather than less likely.
// Their flags are the reason this rule reads the flag mask carefully rather than testing for `const`
// by name; see the constant-binding note below.
//
// # Why this reads the checker
//
// The discrimination is name resolution, not structure. Upstream's clean cases include
//
//	const x = 0; { let x; x = 1; }        a block-scoped shadow
//	const x = 0; function a(x) { x = 1; } a parameter shadow
//	const a = 1; { let a = 2; { a += 1; } }  a shadow written from a nested block
//
// whose text is indistinguishable from the failing forms. A rule matching on the name reports all
// three. Upstream answers with `get_resolved_references`, a find-all-references index we do not
// have. The question our checker does answer is which declaration a given identifier binds to, so
// this anchors on the declaration node the listener already received and asks, for each write,
// whether the resolved symbol's first declaration is that same node.
//
// # Node identity, never declaration kind
//
// The anchor here is a variable declaration and so is every shadow that matters, so comparing the
// *kind* of the resolved declaration cannot separate them at all: a `const A` and a shadowing
// `let A` are both `KindVariableDeclaration`. This is sharper than it was for the sibling rule on
// classes, where kind is merely a near miss. Nothing in the imported corpus catches the difference,
// because upstream's shadow cases resolve to a declaration this rule never anchored on and so exit
// through a different path. The input that separates them is a nested redeclaration of the same
// kind,
//
//	const A = 1; { const A = 2; A = 3; }
//
// where kind reports twice, once for each anchor, and identity reports once. It has its own fixture
// and its own mutant.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names; it says nothing about whether the
// occurrence writes to it. `const x = {key: 0}; x.key = 1;` and `const x = 0; foo(x);` both resolve
// to the anchored declaration and neither reassigns it, so a rule built on identity alone reports
// every property write and every read. The structural half is `ast.IsWriteAccess`, and the rule
// reports only where the two agree.
//
// # The structural half comes mostly off the shelf, and where it does not is measured
//
// `ast.IsWriteAccess` is typescript-go's own answer to the question oxc asks as `is_write()`, and it
// covers most of it: compound assignment, both fixities of `++`/`--`, destructuring through array
// and object literals, property assignments, shorthand properties, and the `for (x of ...)` head.
// Two shapes it gets right that a hand-written climb tends to get wrong in the quiet direction:
//
//	`const FOO = 1; ({ files = FOO } = arg1);`  FOO is the default value, a read. Clean upstream.
//	`for (const x of [1,2,3]) { foo(x); }`      the head declares rather than assigns. Clean upstream.
//
// It is deliberately `IsWriteAccess` rather than `IsWriteOnlyAccess`. The latter is
// `AccessKind == Write`, which excludes `ReadWrite`, and every compound assignment, every `++`, and
// every `x ??= y` is `ReadWrite`. Six of upstream's failing cases are that shape, so the stricter
// accessor silently drops a quarter of the corpus.
//
// # Where the shelf function is short, which was measured rather than assumed
//
// It answers `false` for a rest element in a destructuring target. Upstream's `accessKind` switch
// carries no arm for `KindSpreadElement` or `KindSpreadAssignment`, so both fall through to the
// default and are classified as reads. That is defensible for its own callers, which use it for
// unused-locals reporting, and it is wrong for this rule: two of upstream's failing cases are
// exactly that shape,
//
//	const d = 123; [a, b, ...[c, ...d]] = [1, 2, 3, 4, 5]
//	const b = 0; ({a, ...b} = {a: 1, c: 2, d: 3})
//
// and both went silent on the first run against the imported corpus.
//
// `reference.WritesToBinding` is the shelf function with that one gap closed. This rule wrote the
// wrapper first and two sibling rules wrote their own, differently; a census that read the whole
// corpus at once found four implementations of the one decision and lifted the union. The lifted
// version also corrects a shape this rule's own wrapper got wrong, the shorthand default value, and
// its doc records how.
//
// This is the reason the corpus goes in verbatim before the rule is written. Nothing about the name
// `IsWriteAccess` suggests it declines a rest element, and a fixture set invented alongside the port
// would have encoded the same blind spot and passed.
//
// # The constant-binding mask
//
// `NodeFlagsAwaitUsing` is not a distinct bit. It is defined upstream as `NodeFlagsConst |
// NodeFlagsUsing`, because the two flags are mutually exclusive on any single node and the pair is
// free to mean the third thing. So `flags&NodeFlagsAwaitUsing != 0` is true for a plain `const` and
// for a plain `using` alike, and a rule testing it as though it were one bit misclassifies. The
// right mask is `NodeFlagsConstant`, which is the same `Const|Using` pair and is named for the
// question this rule asks: is this binding constant. It matches ESLint's `CONSTANT_BINDINGS` set of
// const, using, and await using exactly.
//
// No fix. The repair is either changing the declaration to `let`, which the rule cannot know is
// wanted and which is wrong when the write was the mistake, or introducing a new binding, which
// means choosing a name. Both are judgment.
var NoConstAssign = rule.Rule{
	Name: "no-const-assign",

	// See the doc above: upstream's shadow cases are textually identical to failing ones and differ
	// only in what the name resolves to.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Each constant declaration on the shared walk, reading the writes to its names from the
		// file's identifier index.
		//
		// This listened on `KindVariableDeclarationList` once before and walked the whole file from
		// each one, which is quadratic in how many constants a file declares: 881ms and 37.4% of all
		// rule time on the ahra tree. It then gathered every constant in one walk from the file node
		// and matched the writes in a second, still 86ms of rule CPU on a cold ahra run (#fcac58b).
		// The index is built once per file for every rule that asks, so each name's writes are read
		// from the few identifiers spelled like it, and the rule walks nothing itself.
		//
		// The anchors already judged, because a constant declared twice in one scope (a parse that
		// recovers from the error) is one symbol whose anchor both declarations name.
		judged := map[*ast.Node]bool{}
		return rule.Listeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// The engine hands every rule a nil checker when the program could not be built,
				// and this rule can answer nothing without one.
				if ctx.TypeChecker == nil {
					return
				}
				// The flags live on the declaration list rather than on the individual declarator,
				// which is why this reads the list. See the constant-binding note above for why the
				// mask is `NodeFlagsConstant` and not `NodeFlagsAwaitUsing`.
				if node.Flags&ast.NodeFlagsConstant == 0 {
					return
				}
				for _, declaration := range node.AsVariableDeclarationList().Declarations.Nodes {
					// Every name the list binds, which for a destructuring pattern is several and
					// nested arbitrarily deep. `const [a, b, ...[c, ...d]] = x` binds `d` two rest
					// elements down, and upstream fails on exactly that.
					forEachBoundName(declaration.AsVariableDeclaration().Name(), func(boundName *ast.Node) {
						if boundName != nil && boundName.Kind == ast.KindIdentifier {
							reportConstWrites(ctx, boundName, judged)
						}
					})
				}
			},
		}
	},
}

// reportConstWrites reports every write in the file that binds to one constant's name.
//
// The whole file rather than any bounded subtree. A write can sit before the declaration, after it,
// or nested inside a function several scopes down, and all three are the same binding. Upstream has
// a failing case for the write coming first (`x = 123; const x = 1;`) and notes that reporting it
// aligns with ESLint. Anchoring the search on the file and the match on symbol identity is what
// makes the position of the write irrelevant.
func reportConstWrites(ctx rule.Context, boundName *ast.Node, judged map[*ast.Node]bool) {
	// Resolved through the checker rather than taken as the declarator node directly, so that both
	// sides of the comparison are answers to the same question: asking the AST for one side and the
	// checker for the other compares two things that happen to agree today.
	anchor := declarationAnchoredAt(ctx, boundName)
	if anchor == nil || judged[anchor] {
		return
	}
	judged[anchor] = true

	// The declarator's own name needs no exclusion. It is an identifier whose text matches and which
	// resolves to this very declaration, and `WritesToBinding` declines it: a declaration name's
	// parent is the declarator, which is not an assignment. The structural test runs first because it
	// is far cheaper than a checker call.
	for _, occurrence := range reference.IdentifiersNamed(ctx, boundName.Text()) {
		if reference.WritesToBinding(occurrence) && resolvesToDeclaration(ctx, occurrence, anchor) {
			ctx.ReportNode(occurrence, messageNoConstAssign)
		}
	}
}

// forEachBoundName calls back with every identifier a binding name introduces.
//
// A plain identifier binds itself. A binding pattern binds whatever its elements bind, recursively
// and to any depth, and a nested rest element is still a pattern. The property key side of
// `const {a: x} = ...` is deliberately not visited: `a` is a key naming a property of the object
// being destructured, and only `x` becomes a binding.
//
// Omitted elements are skipped. `const [, x] = pair` holds a hole whose element node carries no
// name, and treating it as a binding would dereference nothing.
func forEachBoundName(name *ast.Node, callback func(*ast.Node)) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		callback(name)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindOmittedExpression {
				continue
			}
			forEachBoundName(element.Name(), callback)
		}
	}
}
