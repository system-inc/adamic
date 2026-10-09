package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoFuncAssign = rule.Message{
	Id: "noFuncAssign",
	Description: "This assigns to the name a function declares. The declaration is hoisted, so the " +
		"assignment usually parses and runs without complaint, and what it costs shows up somewhere " +
		"else: every later call of the name invokes whatever was assigned instead of the function. " +
		"Inside the function's own body the mistake is quieter still, because the write happens on " +
		"the first call and the second call reaches a different value. A named function expression " +
		"is stricter, where the binding is immutable and the write throws in strict mode. Assign to " +
		"a separate variable, or declare the function as a variable if it is meant to be replaced.",
}

// NoFuncAssign flags an assignment to the name a function declares.
//
//	valid:   function foo() { var foo = bar; }
//	valid:   function foo(foo) { foo = bar; }
//	valid:   var foo = function() {}; foo = bar;
//	valid:   function foo() {} foo.x = 0;
//	invalid: function foo() {}; foo = bar;
//	invalid: function foo() { foo = bar; }
//	invalid: foo = bar; function foo() { };
//	invalid: var a = function foo() { foo = 123; };
//
// A function declaration binds its name in the enclosing scope, so the assignment parses and the
// mistake is silent at the line that makes it. What it costs shows up at the call sites: the name
// now holds whatever was assigned, so calling it calls that instead. The common real case is the
// write sitting inside the function's own body, where the first call replaces the function and
// every call after it reaches something else.
//
// A named function expression binds its name only inside its own body, and that binding is
// immutable, so `var a = function foo() { foo = 123; };` throws in strict mode rather than
// rebinding. Both shapes are in scope and both are upstream's.
//
// # Why this reads the checker
//
// The discrimination is entirely name resolution and none of it is structural. Four of upstream's
// seven clean cases write to a name spelled exactly like the function:
//
//	function foo(foo) { foo = bar; }              a parameter shadows it
//	function foo() { var foo = bar; }             a local var shadows it inside the body
//	function foo() { var foo; foo = bar; }        the same, written on a later line
//	var foo = function() { foo = bar; };          the variable, not any function name
//
// Their text is nearly identical to the failing forms, so a rule matching on the name reports all
// four. Upstream answers with `get_resolved_references` plus `is_write()`, which is a find-all
// -references index we do not have. The equivalent question our checker does answer is which
// declaration a given identifier binds to, so this anchors on the function node the listener
// already received and asks, for each write, whether the resolved symbol's first declaration is
// that same node.
//
// Node identity rather than declaration kind. Kind looks like it would work, since the anchor is a
// distinctive node type, and it fails silently the moment a shadow shares it: in
// `function foo() {} { function foo() {} foo = 1; }` both anchors are `KindFunctionDeclaration`, so
// a kind test calls the inner write a match for the outer function too and reports twice where
// identity reports once. Upstream's corpus contains no same-kind shadow pair at all and cannot
// catch that, which is why the case is written by hand in the fixtures. The same trap is sharper on
// the sibling rules over `const`, where a `const A` and a shadowing `let A` are both
// `KindVariableDeclaration` and kind cannot separate them even in principle.
//
// # Why the checker is not sufficient on its own
//
// Symbol identity says which binding an identifier names; it says nothing about whether the
// occurrence writes. `function foo() {} foo.x = 0;` and `function foo() {} foo();` both resolve to
// the function and neither reassigns it, so the rule needs the structural half too and reports only
// where the two agree. `reference.WritesToBinding` is that half, shared with five sibling rules.
//
// # One shape the checker answers differently
//
// `({foo} = {})` is a shorthand property in an object literal being used as a destructuring target,
// and `GetSymbolAtLocation` on its identifier returns the property's own symbol rather than the
// binding being written. TypeScript has a separate accessor for the value side and
// `resolvesToDeclaration` reaches for it. Without that, the case reads exactly like a correctly
// declined shadow, which is the quiet direction. Upstream's corpus has the object-pattern shape
// only with an explicit key (`({x: foo = 0} = bar)`), which the plain accessor handles, so this one
// is covered by a hand-written fixture rather than an imported one.
//
// No fix. The repair is a new binding, which means choosing a name and deciding which later uses of
// the old one meant the function, and the rule cannot know that.
var NoFuncAssign = rule.Rule{
	Name: "no-func-assign",

	// See the doc above: four of upstream's seven clean cases are textually near-identical to
	// failing ones and differ only in what the name resolves to.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Each function on the shared walk, reading the writes to its name from the file's identifier
		// index. Each function used to walk the whole file for writes to its name, which made this
		// simple rule one of the tree's slowest (177ms on ahra); then the file was walked twice from
		// the file node, once for the names and once for the writes, still 76ms of rule CPU on a cold
		// ahra run (#fcac58b). The index is built once per file for every rule that asks, so a name's
		// writes are read from the few identifiers spelled like it.
		//
		// The anchors already judged, because overloads of one function share one declaration, and
		// judging it once per overload would report each write once per overload.
		judged := map[*ast.Node]bool{}
		check := func(node *ast.Node) {
			// The engine hands every rule a nil checker when the program could not be built, and
			// this rule can answer nothing without one.
			if ctx.TypeChecker == nil {
				return
			}
			reportWritesToFunctionName(ctx, node, judged)
		}
		return rule.Listeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
		}
	},
}

// reportWritesToFunctionName reports every write to the name a function declares.
//
// Both kinds of function anchor, and they behave differently on purpose. A declaration's name is
// visible to the enclosing scope, so a write anywhere in the file can reach it. A function
// expression's name binds only inside its own body, so `var foo = function() {}; foo = bar;` writes to
// the variable and is clean while `var a = function foo() { foo = 123; };` writes to the function name
// and is not. Nothing here encodes that split: the checker resolves each write and the two cases fall
// out of the same code.
//
// Arrow functions are deliberately absent. They carry no name of their own, so
// `var foo = () => {}; foo = bar;` writes to the variable and upstream leaves it, which is what its
// `foo` pass case asserts. An anonymous function expression declares no name either, so
// `var foo = function() { foo = bar; };` writes to the variable, and upstream leaves it too.
//
// The whole file is searched, not the function's subtree. A write can sit before the declaration
// (`foo = bar; function foo() {}`), after it, or inside it, and hoisting makes all three the same
// binding. Anchoring the search on the file and the match on the symbol is what makes the position of
// the write irrelevant. An identifier resolves to one declaration, so each write is reported by at
// most one function.
func reportWritesToFunctionName(ctx rule.Context, function *ast.Node, judged map[*ast.Node]bool) {
	name := function.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	declaration := declarationAnchoredAt(ctx, name)
	if declaration == nil || judged[declaration] {
		return
	}
	judged[declaration] = true

	// A function's own name needs no exclusion here. It is an identifier whose text matches and which
	// resolves to its declaration, so only the structural half declines it, and that is enough: the
	// climb from a function name reaches the function itself, which is not an assignment. The
	// structural test runs first because it is far cheaper than the checker.
	for _, occurrence := range reference.IdentifiersNamed(ctx, name.Text()) {
		if reference.WritesToBinding(occurrence) && resolvedDeclarationOf(ctx, occurrence) == declaration {
			ctx.ReportNode(occurrence, messageNoFuncAssign)
		}
	}
}
