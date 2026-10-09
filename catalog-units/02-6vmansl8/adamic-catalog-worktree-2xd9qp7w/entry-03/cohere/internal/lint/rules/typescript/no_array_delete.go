package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoArrayDelete flags the `delete` operator applied to an element of an array or a tuple.
//
//	valid:   declare const obj: { a: 1; b: 2 }; delete obj.a
//	valid:   declare const obj: { a: 1; b: 2 }; delete obj['a']
//	valid:   declare const maybeArray: any; delete maybeArray[0]
//	valid:   declare const test: never; delete test[0]
//	invalid: declare const arr: number[]; delete arr[0]
//	invalid: declare const tuple: [number, string]; delete tuple[0]
//	invalid: delete [1, 2, 3][0]
//
// `delete arr[0]` leaves a hole rather than shortening the array: the length is unchanged, the slot
// reads back as `undefined`, and every later index keeps its old position. That is almost never
// what the author meant, which is why the repair is offered as a suggestion rather than applied.
//
// # Absorbed from tsgolint, which is the source of record for this rule
//
// Provenance: tsgolint `internal/rules/no_array_delete/no_array_delete.go`, vendored at commit
// `05b7fbc` and absorbed onto cohere's own rule interface here. It reaches for
// `GetConstrainedTypeAtLocation`, `TrimNodeTextRange` and `Checker_isArrayOrTupleType` because that
// is what upstream reaches for, and this note is why a reader finds those helpers in a file that
// otherwise looks native.
//
// tsgolint is not re-synced, so this file is now the only copy of the algorithm rather than a
// translation layer over a vendored one. The checker logic below is byte-identical to upstream's;
// what changed is the interface it speaks: `rule.RuleContext` became `rule.Context`,
// `RuleListeners` became `Listeners`, `rule.RuleMessage` became `rule.Message`, and the single
// report call site's suggestion moved from `rule.RuleSuggestion{FixesArr: []rule.RuleFix{...}}` to
// `rule.Suggestion{Fixes: []rule.Fix{...}}` with `RuleFixRemoveRange`/`RuleFixReplaceRange` becoming
// `rule.RemoveRange`/`rule.ReplaceRange`. Those are the same constructors under different names, so
// the three surgical ranges are unchanged. No predicate and no traversal was touched.
//
// The reported SPAN is unchanged too, and that is the part worth naming. While this rule was
// adapted, `upstream.Adapt` wrapped every node report in `rule.TokenRange(SourceFile, node)` because
// tsgolint's own runner does the same through `type_checking.TrimNodeTextRange`. Our native
// `ctx.ReportNodeWithSuggestions` applies exactly that trim itself, so absorbing the rule preserves
// the behavior rather than relying on the adapter to supply it. Passing `node.Loc` instead would
// include leading trivia and reintroduce the defect fixed in `8bdd70b`; the span test in this
// package pins all eighteen positioned cases against it.
//
// oxc declares `NoArrayDelete(tsgolint)` and carries no algorithm and no corpus, so the behavior
// oxlint exhibits for this rule IS tsgolint's: the release binary shells out to a `tsgolint`
// executable and refuses to run the rule when that binary is absent, which is what it does on this
// machine. That refusal is itself the proof, so oxlint cannot serve as ground truth here and
// tsgolint's own test file is the corpus instead. All thirty one of its cases are replayed in the
// test file, including all twenty two exact suggestion outputs.
//
// # What decides a finding
//
// The anchor is `KindDeleteExpression`. The operand has its parentheses skipped, so
// `delete ((a[b]))` reports, and it must then be an element access: `delete obj.a` is a property
// access and is silent even when `obj` is an array. The type of the RECEIVER is then resolved
// through `GetConstrainedTypeAtLocation`, which walks a type parameter to its base constraint, so
// both `getArray<T extends number[]>(): T` and `getArray<T extends number>(): T[]` report.
//
// The array test is asymmetric between the two composite forms and the corpus pins both directions.
// A union reports only when EVERY member is an array or a tuple, so `number[] | string[]` reports.
// An intersection reports when ANY member is, so `number[] & unknown` and `string & Array<number>`
// both report. Reading the rule file alone would suggest one predicate; it is two.
//
// `any`, `unknown` and `never` are silent, and they are silent as a consequence of the array test
// rather than through an explicit skip: none of them is an array or a tuple. That distinction
// matters for the fixture tsconfig, which pins `lib: ["ES2022"]`, because a type outside that
// library resolves to the error type whose flags report `any`. Nothing in this rule's corpus needs
// a type outside ES2022, so no case here is at risk of that; `Array` and tuples are all in scope.
//
// # Where the two upstreams disagree, and it is the REPAIR
//
// Both implementations register exactly ONE listener, carry the same two messages, ship the repair
// as a SUGGESTION rather than a fix, and compute `isUnderlyingTypeArray` identically, including the
// every/some asymmetry. Counted on both sides rather than assumed.
//
// They construct the repair completely differently, and the outputs differ observably.
//
// `@typescript-eslint` 8.x replaces the WHOLE node with a reconstructed string:
// it reads `object.getText()` and `property.getText()`, wraps the key in parentheses when it is a
// sequence expression, then hoists every comment inside the node up in front of the replacement,
// re-indented to the node's own column.
//
// tsgolint makes THREE surgical range edits instead: it removes the `delete` token, replaces the
// `[` token with `.splice(`, and replaces the `]` token with `, 1)`. Nothing else in the span is
// touched.
//
// Three consequences, all of them visible in the corpus and all of them reproduced here:
//
// A LEADING SPACE survives where `delete` was removed, because only the keyword token is taken and
// the space after it is not. Every one of upstream's twenty two expected outputs begins
// ` arr.splice(...)` rather than `arr.splice(...)`, and a repair that also ate the space would
// satisfy every message id while failing every output.
//
// COMMENTS STAY WHERE THEY WERE. Upstream's comment-heavy case keeps `/** multi line */` sitting
// between the removed keyword and the receiver, and keeps `/* inline */` inside the index
// expression, because those bytes were never in an edited range. The `@typescript-eslint` fixer
// would have collected them and stacked them above the call.
//
// A SEQUENCE EXPRESSION KEEPS ITS OWN PARENTHESES rather than being given new ones.
// `delete arr[(doWork(), 1)]` becomes ` arr.splice((doWork(), 1), 1)` because the parentheses are
// part of the untouched argument text, so tsgolint needs no special case where the other upstream
// has one. The two agree on the output for this input and reach it by different routes.
//
// tsgolint wins on every disagreement, because oxlint runs tsgolint and oxlint is what the
// differential harness compares against. Reproducing the other repair would be a difference the
// harness could see, for no gain.
//
// # The checker, and the nil guard that now lives here
//
// The listener reads `ctx.TypeChecker` unconditionally, so this needs the checker and its fixtures
// use `RunTyped`. While this rule was adapted, the standing `if ctx.TypeChecker == nil { return }`
// could not be written, because the listener was upstream's and editing it would have turned a
// re-sync into a merge. Absorbing the rule removes that constraint, so the guard is now written
// where the advice always wanted it, and it is the one addition to the body.
//
// The guard is unreachable through registration, since `NeedsTypeChecker` is declared right above.
// It is here for the harness path, where a Context can be built by hand, and because the dangerous
// direction for this rule is silence rather than a panic: under the untyped harness this rule goes
// quiet rather than crashing, and silence makes every clean fixture pass having proven nothing. A
// test in this package pins the declaration so a later revert fails loudly rather than going
// vacuously green.
//
// # Two mutants survive the sweep and both are equivalent rather than unseen
//
// Rewriting `type_checking.TrimNodeTextRange(sourceFile, n)` to `n.Loc` for either the receiver range or the
// argument range compiles, changes bytes, and no fixture notices. That reads as a blind spot and it
// is not one, which matters because the trivia-trimming difference is a real defect this project has
// already shipped once at the adapter.
//
// It cannot bite HERE. `TrimNodeTextRange` is
// `GetRangeOfTokenAtPosition(file, n.Pos()).WithEnd(n.End())`, so it differs from `n.Loc` only in
// where the range STARTS. Both spellings end at `n.End()`. And this rule reads nothing but the end:
// `expressionRange` and `argumentRange` are each consumed exactly once, as `.End()`, to locate the
// bracket tokens. No code path reads either `Pos()`, so no input can distinguish the two versions.
//
// That claim is checked rather than asserted. The same mutation applied to the value actually read,
// turning `expressionRange.End()` into `expressionRange.End() + 1`, is CAUGHT by twenty six failing
// assertions, so the fixtures can see this range and the survival is about the mutation rather than
// about the coverage.
//
// The trimming still belongs in the vendored file exactly as upstream wrote it. It is inert for
// these two call sites today and would stop being inert the moment anything read a start offset.
//
// # Cost
//
// `KindDeleteExpression` is a rare anchor and the listener exits on the second line for anything
// that is not an element access, so the checker is consulted only for a genuine `delete x[y]`.
var NoArrayDelete = rule.Rule{
	Name: "@typescript-eslint/no-array-delete",

	// The listener resolves the receiver's type on every `delete x[y]`, so the checker is required.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		isUnderlyingTypeArray := func(t *checker.Type) bool {
			if type_checking.IsTypeFlagSet(t, checker.TypeFlagsUnion) {
				for _, t := range t.Types() {
					if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) {
						return false
					}
				}
				return true
			}

			if type_checking.IsTypeFlagSet(t, checker.TypeFlagsIntersection) {
				for _, t := range t.Types() {
					if checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t) {
						return true
					}
				}
				return false
			}

			return checker.Checker_isArrayOrTupleType(ctx.TypeChecker, t)
		}

		return rule.Listeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				if node.Kind != ast.KindDeleteExpression {
					return
				}
				deleteExpression := ast.SkipParentheses(node.AsDeleteExpression().Expression)

				if !ast.IsElementAccessExpression(deleteExpression) {
					return
				}

				expression := deleteExpression.AsElementAccessExpression()

				argType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression.Expression)

				if !isUnderlyingTypeArray(argType) {
					return
				}

				expressionRange := type_checking.TrimNodeTextRange(ctx.SourceFile, expression.Expression)
				argumentRange := type_checking.TrimNodeTextRange(ctx.SourceFile, expression.ArgumentExpression)

				deleteTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				leftBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, expressionRange.End())
				rightBracketTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, argumentRange.End())

				ctx.ReportNodeWithSuggestions(node, buildNoArrayDeleteMessage(), rule.Suggestion{
					Message: buildUseSpliceMessage(),
					Fixes: []rule.Fix{
						rule.RemoveRange(deleteTokenRange),
						rule.ReplaceRange(leftBracketTokenRange, ".splice("),
						rule.ReplaceRange(rightBracketTokenRange, ", 1)"),
					},
				})
			},
		}
	},
}

func buildNoArrayDeleteMessage() rule.Message {
	return rule.Message{
		Id:          "noArrayDelete",
		Description: "Using the `delete` operator with an array expression is unsafe.",
	}
}

func buildUseSpliceMessage() rule.Message {
	return rule.Message{
		Id:          "useSplice",
		Description: "Use `array.splice()` instead.",
	}
}
