package typescript

import (
	"math"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferFind flags `array.filter(...)[0]` and `array.filter(...).at(0)`.
//
//	valid:   arr.find(item => item === 'a');
//	valid:   arr.filter(item => item === 'a')[1];
//	valid:   arr.filter(item => item === 'a');
//	invalid: arr.filter(item => item === 'a')[0];
//	invalid: arr.filter(item => item === 'a').at(0);
//
// `filter` walks the whole array building a new one, then everything but the first element is
// thrown away. `find` stops at the first match and allocates nothing. On a long array that is the
// difference between one comparison and every comparison, and the intent reads better besides.
//
// # The repair is a SUGGESTION, and the reason is a real behavior change
//
// `filter(...)[0]` yields `undefined` when nothing matches, and so does `find(...)`, so the value
// is the same. The TYPE is not: `filter` returns `T[]` and indexing it gives `T` under a
// non-strict index signature, while `find` returns `T | undefined`. Swapping them can therefore
// turn code that compiled into code that does not, which is why upstream offers the rewrite rather
// than applying it. Shipping this as a fix would have the edit engine break builds unattended.
//
// # What counts as an index of zero, and where the constant comes from
//
// Upstream reads the index through ESLint's `getStaticValue`, which constant-folds through a
// binding, and then applies two different coercions:
//
//	member access   String(value) === "0"      so [0], ['0'] and a const holding either
//	the at method   Number(value) truncates to 0, and NaN counts    so .at(0), .at('a'), .at(0.2)
//
// The two disagree, deliberately, and the corpus pins the disagreement: `.at(0.2)` reports because
// truncating gives zero, while `[0.2]` does not because the string is not `"0"`. `.at('a')` reports
// because `Number('a')` is not a number at all, which upstream treats as zero.
//
// This tree has no scope-analysis constant folder, and it does not need one for the cases that
// matter. The CHECKER already folds a `const` to its literal type, which is the same answer from a
// different direction, probed before this rule was written:
//
//	const zero = 0;          the index expression's type is the literal `0`
//	const fltr = 'filter';   the member name's type is the literal `"filter"`
//	const n: number = 0;     widened to `number`, so no constant is available
//
// So a literal type is read where upstream reads a folded value. That is fidelity to what the rule
// DECIDES rather than to how upstream obtains it, and it is strictly better on one axis: the
// checker resolves the actual binding, where a name-based scan would confuse a shadowed one.
//
// # Where it is narrower than upstream, stated rather than silent
//
// Three shapes fold upstream and do not fold here, because the checker widens them to `number`
// rather than keeping a literal: a bigint (`const zero = 0n`), `NaN`, and `-Infinity`. All three
// are `.at()` arguments that upstream treats as zero, so this port is SILENT where upstream
// reports. Measured on the installed build; the corpus carries three such cases and they are
// recorded in the test file as divergences rather than deleted.
//
// The direction is the safe one for a rule whose repair is a suggestion: a missed finding costs a
// suggestion nobody sees, where a false one would propose a rewrite that changes types.
//
// # A chain of filters is followed through ternaries and sequences
//
// `(cond ? a.filter(f) : b.filter(g))[0]` reports, and the repair rewrites BOTH filters, because
// either branch could produce the array being indexed. A ternary reports only when both branches
// are filter calls; one bare array on either side and the whole thing is silent. A sequence
// expression contributes only its last operand, since that is the value being indexed.
//
// # Cost
//
// Two anchors, a call expression and a computed element access, both common. Each exits on a cheap
// syntactic test before the checker is consulted.
var PreferFind = rule.Rule{
	Name: "@typescript-eslint/prefer-find",

	// The receiver's type decides whether a `filter` call is Array.prototype.filter, and the index
	// constant is read as a literal type.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// isArrayish answers upstream's `isArrayish`: a possibly-nullable array or tuple, or a
		// union of those, where an intersection must be arrayish in EVERY part.
		//
		// The two composite forms are asymmetric and the corpus pins both. Null and undefined
		// members of a union are skipped rather than disqualifying, so `string[] | undefined`
		// reports, and at least one real array must remain, so `null | undefined` alone does not.
		var isArrayish func(subject *checker.Type) bool
		isArrayish = func(subject *checker.Type) bool {
			foundArray := false
			for unionPart := range type_checking.UnionTypePartsSeq(subject) {
				if type_checking.IsTypeFlagSet(unionPart, checker.TypeFlagsNull) ||
					type_checking.IsTypeFlagSet(unionPart, checker.TypeFlagsUndefined) {
					continue
				}

				// `checker.isArrayType(T[] & S[])` answers false, so an intersection is tested part
				// by part. Upstream carries the same comment for the same reason.
				everyPartIsArray := true
				for _, intersectionPart := range type_checking.IntersectionTypeParts(unionPart) {
					if !checker.Checker_isArrayType(ctx.TypeChecker, intersectionPart) &&
						!checker.IsTupleType(intersectionPart) {
						everyPartIsArray = false
						break
					}
				}
				if !everyPartIsArray {
					return false
				}
				foundArray = true
			}
			return foundArray
		}

		// staticStringOf renders the constant an expression denotes, where one is available.
		//
		// A literal node supplies it directly; an identifier or any other expression supplies it
		// through its literal TYPE, which is how a `const` is folded without scope analysis. The
		// second return reports whether a constant was found at all, which upstream distinguishes
		// from a constant that happens to be empty.
		staticStringOf := func(expression *ast.Node) (string, bool) {
			if expression == nil {
				return "", false
			}
			switch expression.Kind {
			case ast.KindStringLiteral, ast.KindNumericLiteral,
				ast.KindNoSubstitutionTemplateLiteral:
				return expression.Text(), true
			case ast.KindIdentifier:
				// `NaN` and `Infinity` are ordinary identifiers rather than literals, and the
				// checker widens both to `number`, so no literal type is available for them.
				// Upstream folds them through its global scope and treats the result as zero under
				// the `at` coercion, so they are named here directly.
				//
				// A local binding of the same name would shadow the global and make this wrong,
				// which is why the check is only reached when the checker could not supply a
				// literal type: a `const NaN = 5` resolves to the literal 5 above and never gets
				// here. Measured: the checker gives `NaN` the type `number` with no literal flag.
				if expression.Text() == "NaN" || expression.Text() == "Infinity" {
					return expression.Text(), true
				}
			case ast.KindPrefixUnaryExpression:
				// `-Infinity`, which is a unary minus over an identifier rather than a literal.
				unary := expression.AsPrefixUnaryExpression()
				if unary.Operator == ast.KindMinusToken && unary.Operand != nil &&
					unary.Operand.Kind == ast.KindIdentifier && unary.Operand.Text() == "Infinity" {
					return "-Infinity", true
				}
			}

			if ctx.TypeChecker == nil {
				return "", false
			}
			expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
			if expressionType == nil {
				return "", false
			}
			// Only a LITERAL type carries a value. A widened `number` or `string` does not, and
			// reading `TypeToString` on one would produce the word "number" and compare it as if it
			// were the index.
			//
			// A BIGINT literal counts, and leaving it out was a real gap rather than a nicety:
			// upstream folds `const zero = 0n` and reports, and the corpus carries two such cases.
			// The checker gives them `TypeFlagsBigIntLiteral` and renders `0n`, including for
			// `-0n`, which it normalises to the same text.
			if !type_checking.IsTypeFlagSet(expressionType, checker.TypeFlagsStringLiteral) &&
				!type_checking.IsTypeFlagSet(expressionType, checker.TypeFlagsNumberLiteral) &&
				!type_checking.IsTypeFlagSet(expressionType, checker.TypeFlagsBigIntLiteral) {
				return "", false
			}
			rendered := ctx.TypeChecker.TypeToString(expressionType)
			// A string literal type renders with its quotes, which are not part of the value.
			if len(rendered) >= 2 && (rendered[0] == '"' || rendered[0] == '\'') &&
				rendered[len(rendered)-1] == rendered[0] {
				rendered = rendered[1 : len(rendered)-1]
			}
			// A bigint literal type renders with its `n` suffix, which is not part of the value the
			// two coercions operate on: `Number(0n)` is 0 and `String(0n)` is "0".
			if len(rendered) >= 2 && rendered[len(rendered)-1] == 'n' &&
				type_checking.IsTypeFlagSet(expressionType, checker.TypeFlagsBigIntLiteral) {
				rendered = rendered[:len(rendered)-1]
			}
			return rendered, true
		}

		// memberNameOf answers upstream's `isStaticMemberAccessOfValue` for one expected name.
		//
		// A plain property access supplies the name directly. A computed one supplies it through
		// the same constant machinery as the index, which is how `arr['filter'](f)` and a const
		// holding `'filter'` both resolve.
		memberNameOf := func(callee *ast.Node) (name string, isComputed bool, nameNode *ast.Node, ok bool) {
			if ast.IsPropertyAccessExpression(callee) {
				access := callee.AsPropertyAccessExpression()
				if access.Name() == nil || !ast.IsIdentifier(access.Name()) {
					return "", false, nil, false
				}
				return access.Name().Text(), false, access.Name(), true
			}
			if ast.IsElementAccessExpression(callee) {
				access := callee.AsElementAccessExpression()
				value, hasValue := staticStringOf(access.ArgumentExpression)
				if !hasValue {
					return "", false, nil, false
				}
				return value, true, access.ArgumentExpression, true
			}
			return "", false, nil, false
		}

		// filterCall is one `.filter(...)` the repair has to rewrite.
		type filterCall struct {
			nameNode   *ast.Node
			isComputed bool
		}

		// collectFilters answers upstream's `parseArrayFilterExpressions`.
		//
		// It returns every filter call whose result could be the array being indexed, which is more
		// than one only for a ternary. An empty result means this is not a filter expression and
		// nothing is reported.
		var collectFilters func(expression *ast.Node) []filterCall
		collectFilters = func(expression *ast.Node) []filterCall {
			if expression == nil {
				return nil
			}

			// Parentheses are stepped through first, and this is the line the rule was missing.
			//
			// estree has no parenthesized node, so upstream recurses straight into a sequence, a
			// ternary or a call. Our parser keeps `KindParenthesizedExpression`, and every one of
			// upstream's sequence and ternary cases writes its expression in brackets because the
			// grammar requires them before an index. Without this step those cases silently found
			// no filter at all: ten of upstream's twenty-eight failing inputs went quiet, which is
			// the paren difference in its COSTS-FINDINGS direction rather than its over-reporting
			// one.
			//
			// Written as a loop with its own nil check rather than `ast.SkipParentheses`, which
			// dereferences its argument; `((x))` also nests, so one step would not be enough.
			node := expression
			for node.Kind == ast.KindParenthesizedExpression {
				inner := node.AsParenthesizedExpression().Expression
				if inner == nil {
					return nil
				}
				node = inner
			}

			// A sequence contributes only its last operand, since that is the value produced.
			if node.Kind == ast.KindBinaryExpression &&
				node.AsBinaryExpression().OperatorToken != nil &&
				node.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken {
				return collectFilters(node.AsBinaryExpression().Right)
			}

			// A ternary reports only when BOTH branches are filter calls, since either could be
			// the array indexed. One bare array on either side and the whole expression is silent.
			if node.Kind == ast.KindConditionalExpression {
				conditional := node.AsConditionalExpression()
				whenTrue := collectFilters(conditional.WhenTrue)
				if len(whenTrue) == 0 {
					return nil
				}
				whenFalse := collectFilters(conditional.WhenFalse)
				if len(whenFalse) == 0 {
					return nil
				}
				return append(whenTrue, whenFalse...)
			}

			if !ast.IsCallExpression(node) {
				return nil
			}
			call := node.AsCallExpression()
			// An optional CALL is excluded: upstream tests `!node.optional`, so `x?.(...)` never
			// matches even when the callee names filter.
			if call.QuestionDotToken != nil {
				return nil
			}
			callee := call.Expression
			if callee == nil {
				return nil
			}

			name, isComputed, nameNode, ok := memberNameOf(callee)
			if !ok || name != "filter" {
				return nil
			}

			receiver := receiverOfMemberAccess(callee)
			if receiver == nil || ctx.TypeChecker == nil {
				return nil
			}
			if !isArrayish(type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, receiver)) {
				return nil
			}
			return []filterCall{{nameNode: nameNode, isComputed: isComputed}}
		}

		// buildSuggestion assembles the whole repair: every filter becomes find, and the trailing
		// index access is deleted.
		buildSuggestion := func(filters []filterCall, arrayExpression *ast.Node, wholeExpression *ast.Node) (rule.Suggestion, bool) {
			fixes := make([]rule.Fix, 0, len(filters)+1)
			for _, filter := range filters {
				replacement := "find"
				if filter.isComputed {
					// A computed access keeps its quotes, so the replacement is a quoted string
					// rather than a bare name. Upstream writes double quotes regardless of what the
					// source used, and the corpus asserts that.
					replacement = `"find"`
				}
				fixes = append(fixes, rule.ReplaceRange(
					rule.TokenRange(ctx.SourceFile, filter.nameNode), replacement))
			}

			// The index access is removed from the `.` or `[` that begins it through the end of the
			// whole expression, which covers `[0]`, `.at(0)` and `["at"](0)` alike.
			accessStart, hasAccessStart := memberAccessTokenAfter(
				ctx.SourceFile.Text(),
				rule.TokenRange(ctx.SourceFile, arrayExpression).End(),
				rule.TokenRange(ctx.SourceFile, wholeExpression).End())
			if !hasAccessStart {
				// The token has to be there for this shape to have parsed, so this is unreachable
				// from any source that produced these nodes. Declining the whole suggestion rather
				// than emitting a partial one, because a repair that rewrote filter to find without
				// removing the index would change behavior.
				return rule.Suggestion{}, false
			}
			fixes = append(fixes, rule.RemoveRange(core.NewTextRange(
				accessStart, rule.TokenRange(ctx.SourceFile, wholeExpression).End())))

			return rule.Suggestion{Message: buildPreferFindSuggestionMessage(), Fixes: fixes}, true
		}

		report := func(wholeExpression *ast.Node, arrayExpression *ast.Node, filters []filterCall) {
			if suggestion, hasSuggestion := buildSuggestion(filters, arrayExpression, wholeExpression); hasSuggestion {
				ctx.ReportNodeWithSuggestions(wholeExpression, buildPreferFindMessage(), suggestion)
				return
			}
			ctx.ReportNode(wholeExpression, buildPreferFindMessage())
		}

		return rule.Listeners{
			// `filtered.at(0)`
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
					return
				}
				callee := call.Expression
				if callee == nil {
					return
				}
				// An optional member access is excluded, matching upstream's `!callee.optional`.
				if optionalMemberAccess(callee) {
					return
				}
				name, _, _, ok := memberNameOf(callee)
				if !ok || name != "at" {
					return
				}
				value, hasValue := staticStringOf(call.Arguments.Nodes[0])
				if !hasValue || !isTreatedAsZeroByArrayAt(value) {
					return
				}
				receiver := receiverOfMemberAccess(callee)
				if filters := collectFilters(receiver); len(filters) != 0 {
					report(node, receiver, filters)
				}
			},

			// `filtered[0]`
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				access := node.AsElementAccessExpression()
				// Upstream tests `!node.optional`, so `filtered?.[0]` never matches.
				if access.QuestionDotToken != nil {
					return
				}
				value, hasValue := staticStringOf(access.ArgumentExpression)
				if !hasValue || value != "0" {
					// The member-access coercion is `String(value) === "0"`, which is why `[0.2]`
					// is silent while `.at(0.2)` reports.
					return
				}
				if filters := collectFilters(access.Expression); len(filters) != 0 {
					report(node, access.Expression, filters)
				}
			},
		}
	},
}

// receiverOfMemberAccess returns the object a property or element access reads from.
func receiverOfMemberAccess(callee *ast.Node) *ast.Node {
	if ast.IsPropertyAccessExpression(callee) {
		return callee.AsPropertyAccessExpression().Expression
	}
	if ast.IsElementAccessExpression(callee) {
		return callee.AsElementAccessExpression().Expression
	}
	return nil
}

// optionalMemberAccess reports whether a callee is an optional-chained member access.
func optionalMemberAccess(callee *ast.Node) bool {
	if ast.IsPropertyAccessExpression(callee) {
		return callee.AsPropertyAccessExpression().QuestionDotToken != nil
	}
	if ast.IsElementAccessExpression(callee) {
		return callee.AsElementAccessExpression().QuestionDotToken != nil
	}
	return false
}

// isTreatedAsZeroByArrayAt implements the coercion `Array.prototype.at` performs on its argument.
//
// `Number(value)` is taken and truncated, and a value that is not a number at all counts as zero,
// which is why `.at('a')` reports. That differs from the member-access coercion, deliberately, and
// the corpus pins the difference on `0.2`.
func isTreatedAsZeroByArrayAt(value string) bool {
	asNumber, err := strconv.ParseFloat(value, 64)
	if err != nil {
		// `Number('a')` is not a number, which upstream treats as zero. Go's parser also refuses
		// the two spellings JavaScript accepts as numbers, so they are named before this point.
		return true
	}
	if math.IsNaN(asNumber) {
		return true
	}
	if math.IsInf(asNumber, 0) {
		// `Math.trunc(Infinity)` is Infinity rather than zero, so an infinite index is NOT treated
		// as zero. Go parses "Infinity" to +Inf where JavaScript's Number would too, so the two
		// agree; spelled out because the truncation below would otherwise read as covering it.
		return false
	}
	return math.Trunc(asNumber) == 0
}

// memberAccessTokenAfter finds the `.` or `[` that begins the index access being removed.
//
// Trivia is skipped with the scanner rather than by scanning characters, and that is not a
// refinement. Upstream searches TOKENS, so a comment sitting between the array and its index is
// stepped over; a raw character scan stops at the first `.` inside that comment and the repair
// truncates the file at it.
//
// Upstream ships a case built to catch exactly this, a multi-line comment between `.filter(cond)`
// and `.at('0')` whose text contains `[ . ?. <>` and a quote. Written the naive way first, the
// repair cut the source mid-comment and produced unparseable output that still passed every
// message and span assertion.
func memberAccessTokenAfter(text string, position int, upperBound int) (int, bool) {
	index := position
	for index < upperBound && index < len(text) {
		index = scanner.SkipTrivia(text, index)
		if index >= upperBound || index >= len(text) {
			return 0, false
		}
		if text[index] == '.' || text[index] == '[' {
			return index, true
		}
		index++
	}
	return 0, false
}

func buildPreferFindMessage() rule.Message {
	return rule.Message{
		Id:          "preferFind",
		Description: "Prefer .find(...) instead of .filter(...)[0].",
	}
}

func buildPreferFindSuggestionMessage() rule.Message {
	return rule.Message{
		Id:          "preferFindSuggestion",
		Description: "Use .find(...) instead of .filter(...)[0].",
	}
}
