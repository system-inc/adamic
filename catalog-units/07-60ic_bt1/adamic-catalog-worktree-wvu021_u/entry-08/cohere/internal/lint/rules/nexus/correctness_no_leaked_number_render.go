package nexus

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoLeakedNumberRenderId = "leakedNumberRender"

// correctnessNoLeakedNumberRenderText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-leaked-number-render.json`.
var correctnessNoLeakedNumberRenderText = policy.MessageOf("nexus/correctness-no-leaked-number-render", correctnessNoLeakedNumberRenderId)

func correctnessNoLeakedNumberRenderMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoLeakedNumberRenderId,
		Description: correctnessNoLeakedNumberRenderText.Render(nil),
	}
}

// CorrectnessNoLeakedNumberRender reports an `&&` in a JSX child whose falsy operand is typed as a
// number, so a `0` or `NaN` is what gets rendered.
//
//	invalid: {properties.itemsTotal && <p>{properties.itemsTotal} records</p>}   (itemsTotal?: number)
//	invalid: {items.length && <List items={items} />}
//	invalid: {flag && count && <Badge />}                                         (reports `count`)
//	invalid: {open ? total && <Total /> : null}
//	valid:   {properties.itemsTotal !== undefined && <p>...</p>}
//	valid:   {items.length > 0 && <List items={items} />}
//	valid:   {label && <span>{label}</span>}                                      (label: string)
//	valid:   {children && <div>{children}</div>}                                  (children: ReactNode)
//	valid:   {columns && <Grid />}                                                (columns: 1 | 2 | 3)
//	valid:   <Badge visible={count && true} />                                    (an attribute, not a child)
//
// # Where it came from
//
// The JS-catalog pass of the new-rules sweep (`#tevhg3f`, after sonarjs S6439 and Biome's
// `noLeakedRender`), built as task `#1g2wqpw`. Three sites in Structure:
//
//   - `source/components/navigation/pagination/PaginationControls.tsx`: `{properties.itemsTotal && (`
//     prints a bare "0" in the table footer when a table has no records. A bug.
//   - `source/components/forms/fields/multiple-checkbox-grid/FieldInputMultipleCheckboxGrid.tsx`:
//     `{properties.maximumSelectionsPerRow && (`, where `0` means no cap (the toggle handler tests
//     it the same way) and prints "0" under the grid. A bug.
//   - `source/ops/developers/web-sockets/WebSocketsPage.tsx`: `statistics.connectedAt &&`, typed
//     `number | null`. Not connected is `null`, and the number is `Date.now()`, so the `0` the type
//     allows does not happen at runtime. True to the type and harmless; `!== null` says what it means.
//
// eslint-plugin-react's `jsx-no-leaked-render` asks the same question without types and was
// declined at 428 findings in ahra (`react/jsx_no_leaked_render.md`): it reports every string,
// object and boolean too. This rule is that question asked of the checker, not that rule configured.
//
// # Which operands
//
// Only the value that `&&` returns when it is falsy, and only where that value lands in a JSX child
// (`<p>{...}</p>`, `<>{...}</>`), never an attribute and never a spread child. From the child
// expression the rule follows what is rendered: through parentheses, both branches of a ternary,
// the right side of `&&` and `||`, and both sides of `??`. At an `&&`, the left side is the operand
// that can leak, and it is read as the falsy result it produces: through parentheses, both branches
// of a ternary, both sides of a nested `&&` (`flag && count && <X />` checks `flag` and `count`),
// the right side of `||`, and both sides of `??`. Each leaf is judged by its own type at that point,
// which is the narrowed type, so `count !== undefined && count && <X />` judges `count` as `number`.
//
// # Which types
//
// A leaf is reported when its type, with `null`, `undefined` and `void` set aside, is made only of
// parts that render nothing when falsy or are numbers, and at least one part is a number that can
// be falsy:
//
//   - **Can be falsy:** `number`, `bigint`, a number or bigint literal whose value is `0`, `-0` or
//     `NaN`, an intersection with `number` or `bigint` in it (a branded number is still a number),
//     and a type parameter whose constraint is one of these.
//   - **Allowed beside them:** `true`, `false` (so `boolean`), a number literal that is not zero
//     (`1 | 2 | 3`), a numeric enum member, and `never`.
//
// Anything else in the union declines the whole leaf: `string`, an object, `any`, `unknown`, a type
// parameter with no constraint. `ReactNode` contains `number`: the research probe that counted every
// type containing it found 39 more hits typed `ReactNode` and excluded them all, because a value
// typed as a node is meant to be rendered, `0` included. `string | number` is declined for the same
// reason: the operand is something the author renders as text. A literal whose value the checker
// does not hold (a computed enum member) is allowed and never reported, which can miss a zero member
// but never invents one.
//
// # What it cannot see
//
// The type of a number does not say its range. `if(count > 0) { return <p>{count && 'some'}</p>; }`
// still reports, because `number` is all the checker knows there. The fix (`count > 0 &&`) costs
// nothing in that case and states the intent the type cannot. No such site exists in ahra today.
//
// # No fix
//
// Whether a `0` should render nothing or "0 records" is the author's call: `PaginationControls`
// arguably wants the count shown, and `!!count`, `count > 0` and `count !== undefined` each mean
// something different at a different site.
var CorrectnessNoLeakedNumberRender = rule.Rule{
	Name: "nexus/correctness-no-leaked-number-render",

	// The whole rule is a question about the operand's type; without the checker it would be
	// eslint-plugin-react's untyped rule, declined at 428.
	NeedsTypeChecker: true,
	// The rule asks only for the operand's type, never into an imported body.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindJsxExpression: func(node *ast.Node) {
				container := node.AsJsxExpression()
				if container.Expression == nil || container.DotDotDotToken != nil || node.Parent == nil {
					return
				}
				if node.Parent.Kind != ast.KindJsxElement && node.Parent.Kind != ast.KindJsxFragment {
					return
				}
				correctnessNoLeakedNumberRenderRendered(ctx, container.Expression)
			},
		}
	},
}

// correctnessNoLeakedNumberRenderRendered walks an expression whose value is rendered as a JSX
// child, and checks the falsy operand of every `&&` whose result is rendered.
func correctnessNoLeakedNumberRenderRendered(ctx rule.Context, expression *ast.Node) {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		correctnessNoLeakedNumberRenderRendered(ctx, conditional.WhenTrue)
		correctnessNoLeakedNumberRenderRendered(ctx, conditional.WhenFalse)
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			correctnessNoLeakedNumberRenderFalsy(ctx, binary.Left)
			correctnessNoLeakedNumberRenderRendered(ctx, binary.Right)
		case ast.KindBarBarToken:
			// The left side is rendered only when it is truthy, which is a value the author chose.
			correctnessNoLeakedNumberRenderRendered(ctx, binary.Right)
		case ast.KindQuestionQuestionToken:
			correctnessNoLeakedNumberRenderRendered(ctx, binary.Left)
			correctnessNoLeakedNumberRenderRendered(ctx, binary.Right)
		}
	}
}

// correctnessNoLeakedNumberRenderFalsy checks every leaf whose value an expression returns when it
// is falsy, and reports a leaf whose type says that value can be a number.
func correctnessNoLeakedNumberRenderFalsy(ctx rule.Context, expression *ast.Node) {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		correctnessNoLeakedNumberRenderFalsy(ctx, conditional.WhenTrue)
		correctnessNoLeakedNumberRenderFalsy(ctx, conditional.WhenFalse)
		return
	case ast.KindBinaryExpression:
		binary := expression.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			correctnessNoLeakedNumberRenderFalsy(ctx, binary.Left)
			correctnessNoLeakedNumberRenderFalsy(ctx, binary.Right)
			return
		case ast.KindBarBarToken:
			// `a || b` is falsy only when it returns `b`.
			correctnessNoLeakedNumberRenderFalsy(ctx, binary.Right)
			return
		}
	}
	if correctnessNoLeakedNumberRenderCanLeak(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(expression)) {
		ctx.ReportNode(expression, correctnessNoLeakedNumberRenderMessage())
	}
}

// correctnessNoLeakedNumberRenderPart is what one part of a union says about a falsy render.
type correctnessNoLeakedNumberRenderPart uint8

const (
	// correctnessNoLeakedNumberRenderDecline is a part this rule does not judge: a string, an
	// object, `any`, `unknown`. One of these anywhere declines the whole operand.
	correctnessNoLeakedNumberRenderDecline correctnessNoLeakedNumberRenderPart = iota
	// correctnessNoLeakedNumberRenderSilent renders nothing when falsy, or is never falsy.
	correctnessNoLeakedNumberRenderSilent
	// correctnessNoLeakedNumberRenderLeaks is a number that can be falsy, and renders when it is.
	correctnessNoLeakedNumberRenderLeaks
)

// correctnessNoLeakedNumberRenderCanLeak says whether a type is made only of numbers and parts that
// render nothing, with at least one number that can be falsy.
func correctnessNoLeakedNumberRenderCanLeak(typeChecker *checker.Checker, operandType *checker.Type) bool {
	leaks := false
	for part := range type_checking.UnionTypePartsSeq(operandType) {
		switch correctnessNoLeakedNumberRenderClassify(typeChecker, part, 0) {
		case correctnessNoLeakedNumberRenderDecline:
			return false
		case correctnessNoLeakedNumberRenderLeaks:
			leaks = true
		}
	}
	return leaks
}

// correctnessNoLeakedNumberRenderClassify reads one part of a union. depth bounds the walk through
// type parameter constraints, which can name each other.
func correctnessNoLeakedNumberRenderClassify(typeChecker *checker.Checker, part *checker.Type, depth int) correctnessNoLeakedNumberRenderPart {
	switch {
	case part == nil:
		return correctnessNoLeakedNumberRenderDecline
	case type_checking.IsTypeFlagSet(part, checker.TypeFlagsNullable|checker.TypeFlagsVoid|checker.TypeFlagsNever|checker.TypeFlagsBooleanLiteral):
		return correctnessNoLeakedNumberRenderSilent
	case type_checking.IsTypeFlagSet(part, checker.TypeFlagsNumber|checker.TypeFlagsBigInt):
		return correctnessNoLeakedNumberRenderLeaks
	case type_checking.IsTypeFlagSet(part, checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral):
		value, held := part.AsLiteralType().Value().(fmt.Stringer)
		if !held {
			// A computed enum member: a number, but the checker does not hold which.
			return correctnessNoLeakedNumberRenderSilent
		}
		switch value.String() {
		case "0", "-0", "NaN":
			return correctnessNoLeakedNumberRenderLeaks
		}
		return correctnessNoLeakedNumberRenderSilent
	case type_checking.IsIntersectionType(part):
		for _, member := range part.Types() {
			if type_checking.IsTypeFlagSet(member, checker.TypeFlagsNumber|checker.TypeFlagsBigInt) {
				return correctnessNoLeakedNumberRenderLeaks
			}
		}
		return correctnessNoLeakedNumberRenderDecline
	case type_checking.IsTypeFlagSet(part, checker.TypeFlagsInstantiable):
		if depth > 4 {
			return correctnessNoLeakedNumberRenderDecline
		}
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, part)
		if constraint == nil || constraint == part {
			return correctnessNoLeakedNumberRenderDecline
		}
		verdict := correctnessNoLeakedNumberRenderSilent
		for constraintPart := range type_checking.UnionTypePartsSeq(constraint) {
			switch correctnessNoLeakedNumberRenderClassify(typeChecker, constraintPart, depth+1) {
			case correctnessNoLeakedNumberRenderDecline:
				return correctnessNoLeakedNumberRenderDecline
			case correctnessNoLeakedNumberRenderLeaks:
				verdict = correctnessNoLeakedNumberRenderLeaks
			}
		}
		return verdict
	}
	return correctnessNoLeakedNumberRenderDecline
}
