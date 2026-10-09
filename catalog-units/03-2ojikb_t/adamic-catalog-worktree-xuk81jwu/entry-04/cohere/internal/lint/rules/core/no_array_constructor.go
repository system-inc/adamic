package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoArrayConstructor = rule.Message{
	Id: "noArrayConstructor",
	Description: "This builds an array with the `Array` constructor, whose argument list means " +
		"two different things depending on how many arguments it gets: `Array(3)` is an empty " +
		"array of length three, while `Array(3, 4)` holds those two values. A refactor that " +
		"changes the count silently changes the kind of array produced. Use a literal, `[3, 4]`, " +
		"which reads the same at every length.",
}

var messageUseAnArrayLiteral = rule.Message{
	Id:          "useAnArrayLiteral",
	Description: "Replace the `Array` call with an array literal.",
}

// NoArrayConstructor flags a call to the `Array` constructor that is not the length form.
//
//	valid:   let a = [];
//	valid:   let a = new Array(9);
//	valid:   let a = Array(x);
//	valid:   let a = Array.from(iterable);
//	valid:   let a = new Array<Foo>(1, 2, 3);
//	valid:   let a = Array?.(1, 2, 3);
//	invalid: let a = new Array();
//	invalid: let a = new Array(1, 2);
//	invalid: let a = Array(...args);
//
// The defect is that the argument list is overloaded. One numeric argument sets a length and stores
// nothing; two or more store elements. So `Array(items)` and `Array(first, second)` are not the same
// function with different arity, they are different constructors sharing a spelling, and a change
// that adds or removes an argument crosses between them without any diagnostic. A literal has no
// such seam.
//
// # One argument is exempt, and that is the whole point of the rule
//
// `new Array(9)` is the only way to get a length-nine array in one expression, so it is deliberately
// left alone. That makes the predicate "argument count is not one" rather than "any Array call", and
// a port that reported every call would fire on the one form the rule exists to permit. `Array()`
// with zero arguments is reported, because `[]` says the same thing.
//
// # A spread argument is not one argument
//
// `Array(...args)` counts as a single argument syntactically while its runtime length is whatever
// `args` holds. If it holds exactly one number the call silently becomes the length form, which is
// the pitfall in its most dangerous shape, so a trailing spread is reported regardless of count.
// Upstream tests this in five shapes and each one is in the fixtures below.
//
// # What is not the global `Array`
//
// A property access is a different function, so `foo.Array()`, `Array.foo()`, and
// `new globalThis.Array` are untouched: the callee is not a bare identifier. Explicit type arguments
// mean `Array<Foo>(1, 2, 3)`, which is TypeScript asking for a typed array and not the untyped
// pitfall, so upstream exempts it and so does this. An optional call, `Array?.(1, 2, 3)`, is exempt
// too, and that one is upstream declining a case rather than judging it safe; see the note below.
//
// # Shadowing has to be resolved, and finding that out cost this port its first draft
//
// A local binding named `Array` shadows the global, and then `new Array()` constructs whatever that
// binding holds, which may be entirely reasonable. Upstream asks `is_global_reference_name`, a real
// scope analysis, and carries two such cases in its clean corpus: `const createArray = Array => new
// Array()` and `var Array; new Array;`.
//
// This rule was first written to match on the name alone, on the reasoning that shadowing `Array` is
// vanishingly rare and that neither of those two cases was a shape that fires anyway. The second
// half of that was simply false, and the fixtures said so immediately: both are zero-argument `new
// Array`, which is the most-reported shape there is. The reasoning was never checked before it was
// written down, and copying upstream's clean cases verbatim is the only reason it did not ship.
//
// So the check is the checker's, matching upstream's question rather than approximating it. There is
// no structural answer, as `no-new-native-nonconstructor` records at length: the shadow can be a
// parameter, an import, or a `var` in any enclosing scope, so no bounded walk finds it.
//
// # What that costs, and why it is still right
//
// `NeedsTypeChecker` takes an exclusive per-file lock, measured at roughly 50% of the lint phase
// across every file that acquires it, and this is the second hand-written rule to declare it. The
// price is paid because the alternative is not "a rarer false positive", it is two false positives
// on upstream's own clean corpus, on the single most common shape the rule reports. A rule that
// fires on `new Array()` inside any function taking an `Array` parameter is worse than no rule: it
// is the failure that gets a rule disabled tree-wide.
//
// The lock is acquired once per file, not per node, and the question is asked only for a call whose
// callee is the bare identifier `Array` and whose argument count already qualifies, so it is asked
// on almost no files in a normal tree. The dry run over the tree this ships into confirms that.
//
// `resolvesToAGlobal` is the existing helper and is used rather than respelled. Its conservative
// direction is the right one here too: a name the checker cannot resolve answers false and is left
// alone, because reporting on a binding nothing declares would be a guess.
//
// # The rewrite is a suggestion, and upstream shipping it as a fix is a live defect
//
// `Array()` becomes `[]`, which looks like the safest rewrite in the catalog and is not, because a
// call expression at the start of a line is a statement while a bracket at the start of a line
// continues the previous one. Upstream carries this input in its own corpus:
//
//	(function () {
//	    Fn
//	    Array() // ";" required
//	}) as Fn
//
// Rewritten, `Fn` and `[]` stop being two statements and become the single element access `Fn[]`.
// Measured directly on our parser: before the rewrite the block holds an identifier statement and a
// call statement, after it holds one `ElementAccessExpression`. Upstream comments the pair out of
// its fix table with "These currently produce invalid syntax, need to fix the fixer" but still
// declares the rule `fix`, so its engine applies the rewrite unattended on exactly the input it
// knows it gets wrong.
//
// This is the failure the brief singles out as the one an engine structurally cannot refuse: the
// output parses, so no syntax guard fires, and the meaning changed anyway. So the rewrite is offered
// as a suggestion a human accepts rather than a fix applied unattended. That is not caution about
// the common case, where `[]` is plainly right; it is that the rule cannot tell the common case from
// the ASI case without asking what precedes the statement, and a rewrite that is right except when
// it silently is not belongs on the human side of that line.
//
// The replaced span runs from the first argument to one character before the call's end, which
// drops the closing paren and copies everything between as source text, so the arguments and any
// comments among them survive with no per-shape arm.
//
// Two kinds of comment do not survive, both measured rather than assumed. A call whose parentheses
// hold only comments, `Array(/*a*/ /*b*/)`, has an empty argument list because comments are trivia,
// so it takes the no-arguments branch and becomes `[]`. And a comment between the callee and the
// open paren, `Array/*a*/()`, sits outside any span anchored on the first argument. Upstream loses
// both as well and carries the pairs commented out under "TODO: Preserve comments around callee".
// Acceptable only because this is a suggestion a human reads before accepting; it would be a defect
// in a fix.
var NoArrayConstructor = rule.Rule{
	Name: "@typescript-eslint/no-array-constructor",

	// See the doc above: a local binding named `Array` makes the reported shapes correct code, and
	// two of upstream's clean cases are exactly that. Nothing structural answers it.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			callee, arguments, typeArguments, isOptionalCall := arrayConstructorCallParts(node)

			// A bare identifier only. `foo.Array()` and `new globalThis.Array` call something else,
			// and this is the check that keeps them clean rather than a separate arm for each.
			//
			// Deliberately not `isIdentifierNamed`, which would see through parentheses: upstream
			// does not, and carries `(Array)(1, 2, 3)` commented out as a known miss. Reaching for
			// the shelf helper here would report a case upstream is silent on, which is improving on
			// upstream by accident rather than on purpose.
			if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "Array" {
				return
			}

			// `Array<Foo>()` is TypeScript asking for a typed array rather than the untyped pitfall.
			if typeArguments != nil {
				return
			}

			// `Array?.(1, 2, 3)`. Upstream carries every optional form as a pass case and marks them
			// "TODO: Catch optional chaining cases", so this is upstream declining to judge them,
			// not upstream judging them safe. Reproduced rather than improved on, because the brief
			// asks for upstream's behavior and a silent divergence here would be invisible.
			if isOptionalCall {
				return
			}

			// The exemption the rule exists around: exactly one argument is the length form, so
			// `Array(9)` is left alone and everything else is reported. A spread defeats it, because
			// `Array(...args)` counts as one argument while its runtime length is whatever `args`
			// holds, which is the length pitfall in its least visible form.
			//
			// Upstream writes the spread test as "is the *last* argument a spread". Written that way
			// here it is indistinguishable from asking about the first, and measurably so: a mutant
			// swapping `arguments.Nodes[count-1]` for `arguments.Nodes[0]` compiled, changed bytes,
			// and no fixture could see it. That is not a hole in the fixtures. The guard is only
			// consulted when the count is exactly one, and with one argument the first and the last
			// are the same node, so the two spellings cannot disagree on any input in any language.
			//
			// So it asks the question it actually means, about the single argument, rather than
			// carrying a positional index whose apparent generality is unreachable. Upstream's
			// spelling is not wrong; it is load bearing there because its condition is
			// `arg_len != 1 || last_arg_is_spread`, which evaluates the spread test for every count.
			// Ours returns early instead, and an index nothing can distinguish reads as a case
			// somebody handled.
			if arguments != nil && len(arguments.Nodes) == 1 &&
				arguments.Nodes[0].Kind != ast.KindSpreadElement {
				return
			}

			// Last, because it is the only question here that costs anything. Every check above is
			// a field read, so by the time this runs the node is already known to be a bare `Array`
			// call in a reported shape, which is a handful of nodes in a whole tree.
			//
			// The nil test is stated here rather than left to `resolvesToAGlobal`, which also guards
			// it. Without a checker this rule cannot tell the global from a local binding, and the
			// two available answers are opposite failures: report everything, including every
			// shadowed call, or report nothing. Declining is the honest one, and saying so at the
			// point of the decision keeps it from reading as an accident of the helper.
			//
			// It also states the dependency where the checker-declaration guard can see it. That
			// guard reads each rule file for a `.TypeChecker` selector, and this rule's only read
			// would otherwise sit in `resolvesToAGlobal` one file over, so the rule would declare a
			// need the guard could not confirm.
			if ctx.TypeChecker == nil || !resolvesToAGlobal(ctx, callee) {
				return
			}

			ctx.ReportNodeWithSuggestions(node, messageNoArrayConstructor, rule.Suggestion{
				Message: messageUseAnArrayLiteral,
				Fixes: []rule.Fix{
					ctx.ReplaceNode(node, arrayLiteralReplacing(ctx, node, arguments)),
				},
			})
		}

		return rule.Listeners{
			// Both call forms. `Array(1, 2)` and `new Array(1, 2)` build the same array, and a rule
			// watching only `new` is silent on the plainer half of upstream's corpus.
			ast.KindNewExpression:  report,
			ast.KindCallExpression: report,
		}
	},
}

// arrayConstructorCallParts pulls the four things this rule asks of either call form.
//
// `new` and plain calls carry the same fields under different types, and reading them at the two
// call sites separately is how the two arms drift. The optional flag is always false for `new`,
// because `new Array?.()` is not expressible.
func arrayConstructorCallParts(node *ast.Node) (
	callee *ast.Node, arguments *ast.NodeList, typeArguments *ast.NodeList, isOptionalCall bool,
) {
	switch node.Kind {
	case ast.KindNewExpression:
		newExpression := node.AsNewExpression()
		return newExpression.Expression, newExpression.Arguments, newExpression.TypeArguments, false
	case ast.KindCallExpression:
		callExpression := node.AsCallExpression()
		return callExpression.Expression, callExpression.Arguments, callExpression.TypeArguments,
			callExpression.QuestionDotToken != nil
	}
	return nil, nil, nil, false
}

// arrayLiteralReplacing builds the literal that replaces a reported call.
//
// The text between the first argument and one character before the call's end is copied verbatim,
// which is what carries the arguments, the separators, and any comments among them through without
// a case for each. One character back is the closing paren.
//
// `new Array` with no parentheses parses with a nil argument list rather than an empty one, which is
// a distinct shape from `new Array()` and reaches the same empty answer here. A port reading the
// list without that guard panics on the shortest input in the corpus.
func arrayLiteralReplacing(ctx rule.Context, node *ast.Node, arguments *ast.NodeList) string {
	if arguments == nil || len(arguments.Nodes) == 0 {
		return "[]"
	}
	return "[" + ctx.SourceFile.Text()[arguments.Nodes[0].Pos():node.End()-1] + "]"
}
