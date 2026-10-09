package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoConfusingVoidExpressionSettings is the decoded option surface.
//
// All three default to false, which is the zero value, so the generic decoder would answer
// correctly. Hand rolled anyway so that an explicit false stays distinguishable from an absent key
// in the wire type, which is what a later reader checking "did anyone turn this on" needs.
type NoConfusingVoidExpressionSettings struct {
	// IgnoreArrowShorthand exempts a concise arrow body, `() => f()`.
	IgnoreArrowShorthand bool
	// IgnoreVoidOperator exempts an expression already marked `void`, and changes three of the
	// messages into their wrap-with-void spellings.
	IgnoreVoidOperator bool
	// IgnoreVoidReturningFunctions exempts a return from a function whose return type includes
	// void, whether annotated or contextual.
	IgnoreVoidReturningFunctions bool
}

// DefaultNoConfusingVoidExpressionSettings is upstream's defaultOptions, all three false.
func DefaultNoConfusingVoidExpressionSettings() NoConfusingVoidExpressionSettings {
	return NoConfusingVoidExpressionSettings{}
}

type noConfusingVoidExpressionWire struct {
	IgnoreArrowShorthand         *bool `json:"ignoreArrowShorthand"`
	IgnoreVoidOperator           *bool `json:"ignoreVoidOperator"`
	IgnoreVoidReturningFunctions *bool `json:"ignoreVoidReturningFunctions"`
}

// DecodeNoConfusingVoidExpressionOptions reads the option object off the config.
func DecodeNoConfusingVoidExpressionOptions(raw []byte) (any, error) {
	settings := DefaultNoConfusingVoidExpressionSettings()
	if len(raw) == 0 {
		return settings, nil
	}
	var wire noConfusingVoidExpressionWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}
	if wire.IgnoreArrowShorthand != nil {
		settings.IgnoreArrowShorthand = *wire.IgnoreArrowShorthand
	}
	if wire.IgnoreVoidOperator != nil {
		settings.IgnoreVoidOperator = *wire.IgnoreVoidOperator
	}
	if wire.IgnoreVoidReturningFunctions != nil {
		settings.IgnoreVoidReturningFunctions = *wire.IgnoreVoidReturningFunctions
	}
	return settings, nil
}

var messageInvalidVoidExpr = rule.Message{
	Id: "invalidVoidExpr",
	Description: "Placing a void expression inside another expression is forbidden. Move it to " +
		"its own statement instead. A void expression evaluates to undefined, so whatever this " +
		"outer expression does with the value is working on undefined rather than on a result.",
}

var messageInvalidVoidExprArrow = rule.Message{
	Id: "invalidVoidExprArrow",
	Description: "Returning a void expression from an arrow function shorthand is forbidden. " +
		"Please add braces to the arrow function. The concise body makes the arrow return " +
		"undefined, which reads as deliberate when it is usually an accident of the syntax.",
}

var messageInvalidVoidExprArrowWrapVoid = rule.Message{
	Id: "invalidVoidExprArrowWrapVoid",
	Description: "Void expressions returned from an arrow function shorthand must be marked " +
		"explicitly with the `void` operator, so a reader can see the discard was intended.",
}

var messageInvalidVoidExprReturn = rule.Message{
	Id: "invalidVoidExprReturn",
	Description: "Returning a void expression from a function is forbidden. Please move it " +
		"before the `return` statement. Returning it says the value matters when it is undefined.",
}

var messageInvalidVoidExprReturnLast = rule.Message{
	Id: "invalidVoidExprReturnLast",
	Description: "Returning a void expression from a function is forbidden. Please remove the " +
		"`return` statement. It is the last statement, so the return adds nothing but the " +
		"suggestion that a value comes back.",
}

var messageInvalidVoidExprReturnWrapVoid = rule.Message{
	Id: "invalidVoidExprReturnWrapVoid",
	Description: "Void expressions returned from a function must be marked explicitly with the " +
		"`void` operator, so a reader can see the discard was intended.",
}

var messageInvalidVoidExprWrapVoid = rule.Message{
	Id: "invalidVoidExprWrapVoid",
	Description: "Void expressions used inside another expression must be moved to their own " +
		"statement or marked explicitly with the `void` operator.",
}

// NoConfusingVoidExpression requires an expression of type void to appear in statement position.
//
//	valid:   f();
//	valid:   cond && f();
//	valid:   maybe?.f();                  (types as void | undefined, not void)
//	invalid: const a = f();
//	invalid: const a = () => f();         (arrow shorthand)
//	invalid: function h() { return f(); } (final return)
//
// # The judgment is "walk up until the position is decided"
//
// Upstream's `findInvalidAncestor` climbs from the void expression through the parents that are
// TRANSPARENT to this question and stops at the first one that decides it. A statement position is
// valid; anything else is not. Three parents are transparent, and each is transparent for its own
// reason:
//
//	the RIGHT of a logical operator      `cond && f()` is a statement if the whole thing is
//	either arm of a conditional          `cond ? f() : f()` likewise
//	a non-final sequence element         `f(), g()` discards f()'s value by construction
//
// The left of a logical operator is NOT transparent, and that asymmetry is real rather than an
// oversight: `f() && cond` uses the void value as a condition. Measured against the installed rule,
// which reports it.
//
// # Two node kinds upstream distinguishes are one kind here
//
// Upstream reads `LogicalExpression` and `SequenceExpression` as separate node types. Our parser
// gives both `KindBinaryExpression` and distinguishes them by operator token, so the two arms have
// to key on `&&`/`||`/`??` and on the comma respectively. A port testing the kind alone would treat
// a comma expression as a logical one and exempt the wrong operand.
//
// # Upstream's ChainExpression arm has no counterpart, and does not need one
//
// Upstream unwraps a `ChainExpression` wrapper around an optional call. Our parser has no such
// node: it puts a `QuestionDotToken` on the access itself. The arm is therefore unreachable here,
// and it also would not matter: measured, `maybe?.f()` types as `void | undefined` rather than
// `void`, so it fails the void test before any ancestor walk. The installed rule agrees and reports
// nothing for it in statement, assigned, or arrow position.
//
// # Three of the eight messages are unreachable under the default configuration
//
// `invalidVoidExprArrowWrapVoid`, `invalidVoidExprReturnWrapVoid` and `invalidVoidExprWrapVoid`
// all require `ignoreVoidOperator`, and `voidExprWrapVoid` is a suggestion label that goes with the
// last of them. They are implemented rather than omitted, because the option is part of the rule's
// surface, but the corpus exercises them only under that option.
//
// # The fixes, and the one this port declines
//
// Upstream ships a fix on the arrow arm and on both return arms, and a SUGGESTION on the generic
// arm under `ignoreVoidOperator`. Two of the three fixes are reproduced here and the third is
// declined; see `noConfusingVoidExpressionReturnFix` for which and why.
var NoConfusingVoidExpression = rule.Rule{
	Name: "@typescript-eslint/no-confusing-void-expression",

	// Every judgment is a question about a type, so the checker is not optional here.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoConfusingVoidExpressionSettings](options)
		if !ok {
			settings = DefaultNoConfusingVoidExpressionSettings()
		}

		// One arrow can hold two reported expressions, and both would propose the identical
		// repair. Tracked per run so only the first carries it; see the arrow arm.
		repairedArrows := map[*ast.Node]bool{}

		check := func(node *ast.Node) {
			checkConfusingVoidExpression(ctx, node, settings, repairedArrows)
		}

		return rule.Listeners{
			ast.KindAwaitExpression:          check,
			ast.KindCallExpression:           check,
			ast.KindTaggedTemplateExpression: check,
		}
	},
}

// checkConfusingVoidExpression judges one void-capable expression.
func checkConfusingVoidExpression(ctx rule.Context, node *ast.Node,
	settings NoConfusingVoidExpressionSettings, repairedArrows map[*ast.Node]bool) {

	if ctx.TypeChecker == nil || ctx.SourceFile == nil {
		return
	}

	invalidAncestor := findConfusingVoidInvalidAncestor(node, settings)
	if invalidAncestor == nil {
		return
	}

	// The type test comes AFTER the ancestor walk, matching upstream's order. It matters for cost
	// rather than for verdicts: the walk is a few pointer hops and the type query is not, so
	// asking the cheap question first keeps the checker out of the common case.
	//
	// Deliberately quieter than upstream, which tests `TypeFlags.VoidLike` (void OR undefined) and so
	// reports an `undefined` value as though it were nothing. `void` says there is no result;
	// `undefined` is a value somebody returned on purpose, and the rule's name and every message are
	// about void. Upstream's corpus writes no undefined-typed expression, so no imported case moves.
	// The real sites were `Random.test.ts:6` and `TrackedPromise.test.ts:132` in ahra; see
	// TestNoConfusingVoidExpressionJudgesVoidNotUndefined. The fixers' own `canFix` guards below keep
	// VoidLike, because they are upstream's rewrite-safety test and only run once this one has passed.
	nodeType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
	if !type_checking.IsTypeFlagSet(nodeType, checker.TypeFlagsVoid) {
		return
	}

	if invalidAncestor.Kind == ast.KindArrowFunction {
		if settings.IgnoreVoidReturningFunctions &&
			isConfusingVoidReturningFunction(ctx, invalidAncestor) {
			return
		}
		if settings.IgnoreVoidOperator {
			ctx.ReportNodeWithFixes(node, messageInvalidVoidExprArrowWrapVoid,
				noConfusingVoidExpressionWrapFix(ctx, node))
			return
		}
		// Two void expressions inside one arrow body report twice and would each propose the SAME
		// repair, which this engine reads as an overlap rather than deduplicating. Upstream's own
		// vector for `foo => (foo ? a : b)` is a single rewrite against two findings, so the fix
		// is attached to the first finding for a given arrow and withheld from the rest.
		fix, fixable := noConfusingVoidExpressionArrowFix(ctx, invalidAncestor)
		if !fixable || repairedArrows[invalidAncestor] {
			ctx.ReportNode(node, messageInvalidVoidExprArrow)
			return
		}
		repairedArrows[invalidAncestor] = true
		ctx.ReportNodeWithFixes(node, messageInvalidVoidExprArrow, fix...)
		return
	}

	if invalidAncestor.Kind == ast.KindReturnStatement {
		if settings.IgnoreVoidReturningFunctions {
			if function := confusingVoidParentFunction(invalidAncestor); function != nil &&
				isConfusingVoidReturningFunction(ctx, function) {
				return
			}
		}
		if settings.IgnoreVoidOperator {
			ctx.ReportNodeWithFixes(node, messageInvalidVoidExprReturnWrapVoid,
				noConfusingVoidExpressionWrapFix(ctx, node))
			return
		}
		if isConfusingVoidFinalReturn(invalidAncestor) {
			fix, fixable := noConfusingVoidExpressionReturnFix(ctx, invalidAncestor, true)
			if !fixable {
				ctx.ReportNode(node, messageInvalidVoidExprReturnLast)
				return
			}
			ctx.ReportNodeWithFixes(node, messageInvalidVoidExprReturnLast, fix)
			return
		}
		fix, fixable := noConfusingVoidExpressionReturnFix(ctx, invalidAncestor, false)
		if !fixable {
			ctx.ReportNode(node, messageInvalidVoidExprReturn)
			return
		}
		ctx.ReportNodeWithFixes(node, messageInvalidVoidExprReturn, fix)
		return
	}

	if settings.IgnoreVoidOperator {
		// A SUGGESTION rather than a fix upstream, and the distinction is the port: wrapping in
		// `void` changes what the expression means to a reader, so a human chooses it.
		ctx.ReportNodeWithSuggestions(node, messageInvalidVoidExprWrapVoid, rule.Suggestion{
			Message: rule.Message{
				Id:          "voidExprWrapVoid",
				Description: "Mark with an explicit `void` operator.",
			},
			Fixes: []rule.Fix{noConfusingVoidExpressionWrapFix(ctx, node)},
		})
		return
	}

	ctx.ReportNode(node, messageInvalidVoidExpr)
}

// findConfusingVoidInvalidAncestor climbs to the first ancestor that decides the position, or nil
// when the expression is already in statement position.
//
// This is upstream's `findInvalidAncestor`, with two shape differences recorded at their arms.
func findConfusingVoidInvalidAncestor(node *ast.Node,
	settings NoConfusingVoidExpressionSettings) *ast.Node {

	parent := node.Parent
	if parent == nil {
		return nil
	}

	switch parent.Kind {
	case ast.KindExpressionStatement:
		// Already a statement, which is the whole point of the rule.
		return nil

	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return parent
		}
		switch binary.OperatorToken.Kind {
		case ast.KindCommaToken:
			// Upstream's SequenceExpression arm, and it is NOT transparent. A non-final element's
			// value is discarded by the operator itself, so that position is valid and the walk
			// stops with nil. The FINAL element falls through to "any other parent is invalid",
			// which means the comma expression itself is the deciding ancestor rather than
			// whatever encloses it.
			//
			// That distinction is visible only in return position and the corpus states it:
			// `return (input, console.log(input))` reports `invalidVoidExpr`, not a return-arm
			// message, while `return (console.log())` reports `invalidVoidExprReturnLast`.
			// Recursing here instead would produce the return arm for both. Measured against the
			// installed rule as well as read from the corpus.
			//
			// "The last element" cannot be answered by looking one level up, and that is the
			// difference our parser forces. ESTree gives a comma chain ONE SequenceExpression node
			// with a flat `expressions` array, so upstream compares against its last entry. Ours
			// nests left-associatively, so `(s, f(), s)` makes the call the RIGHT operand of the
			// inner comma while that inner comma is the LEFT operand of the outer one.
			//
			// Measured on exactly that shape, which is one of upstream's clean cases. Testing only
			// the immediate parent reports it, because the call looks final from one level up.
			if isConfusingVoidNonFinalCommaElement(node, parent) {
				return nil
			}
			return parent

		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// Upstream's LogicalExpression arm, and only the RIGHT operand is transparent. The
			// left is the condition, so a void value there is used rather than discarded, and
			// upstream reports it. Measured against the installed rule.
			if binary.Right == node {
				return findConfusingVoidInvalidAncestor(parent, settings)
			}
			return parent
		}
		return parent

	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		if conditional != nil &&
			(conditional.WhenTrue == node || conditional.WhenFalse == node) {
			return findConfusingVoidInvalidAncestor(parent, settings)
		}
		return parent

	case ast.KindArrowFunction:
		if settings.IgnoreArrowShorthand {
			return nil
		}
		return parent

	case ast.KindVoidExpression:
		if settings.IgnoreVoidOperator {
			return nil
		}
		return parent

	case ast.KindParenthesizedExpression:
		// Our parser keeps a node upstream's folds away, so `(f())` would otherwise report the
		// parenthesis as the deciding ancestor and take the generic arm even in statement
		// position. Transparent, which is what upstream gets for free.
		return findConfusingVoidInvalidAncestor(parent, settings)
	}

	return parent
}

// isConfusingVoidNonFinalCommaElement says whether a node's value is discarded by a comma chain.
//
// It is discarded when the node is a left operand, and also when every comma it is the right
// operand of is itself eventually a left operand. Walking up while the node remains a right operand
// answers both: the walk ends either at a comma where it is the left operand, which is discarded,
// or at a non-comma parent, which means it was the chain's final value.
func isConfusingVoidNonFinalCommaElement(node *ast.Node, comma *ast.Node) bool {
	current, parent := node, comma
	for parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil ||
			binary.OperatorToken.Kind != ast.KindCommaToken {
			return false
		}
		if binary.Left == current {
			return true
		}
		current, parent = parent, parent.Parent
	}
	return false
}

// isConfusingVoidFinalReturn is upstream's `isFinalReturn`: the return is the last statement of a
// function's own body block.
//
// Three conditions and each excludes a real shape. A return not directly inside a block, such as
// `if (cond) return f();`, is not final. A block that is not a function's own body, such as the
// consequent of an `if`, is not final either. And a return followed by another statement is not
// final even inside a function body.
func isConfusingVoidFinalReturn(returnStatement *ast.Node) bool {
	block := returnStatement.Parent
	if block == nil || block.Kind != ast.KindBlock {
		return false
	}
	blockParent := block.Parent
	if blockParent == nil {
		return false
	}
	switch blockParent.Kind {
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression:
	default:
		return false
	}
	statements := block.AsBlock().Statements
	if statements == nil || len(statements.Nodes) == 0 {
		return false
	}
	return statements.Nodes[len(statements.Nodes)-1] == returnStatement
}

// confusingVoidParentFunction returns the function a return statement belongs to.
func confusingVoidParentFunction(node *ast.Node) *ast.Node {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		switch ancestor.Kind {
		case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor:
			return ancestor
		}
	}
	return nil
}

// noConfusingVoidExpressionWrapFix prepends the `void` operator to the expression.
//
// Safe by construction and it is worth saying why, because this is the family that has destroyed
// type information twice in this tree. The fix INSERTS text before a span rather than replacing
// one: the expression's own bytes are untouched, so a type assertion, a `satisfies`, a non-null
// `!`, type arguments and any comment inside it all survive because nothing rewrites them.
//
// `TokenRange` rather than the node's own Pos, which includes leading trivia: inserting at the
// untrimmed position would put `void ` before a comment written above the expression rather than
// before the expression.
func noConfusingVoidExpressionWrapFix(ctx rule.Context, node *ast.Node) rule.Fix {
	span := rule.TokenRange(ctx.SourceFile, node)
	return rule.ReplaceRange(span.WithEnd(span.Pos()), "void ")
}

// noConfusingVoidExpressionArrowFix wraps a concise arrow body in braces.
//
// Two edits rather than one replacement, and that is the whole safety argument. Upstream replaces
// the gap between the arrow token and the body with ` { `, and the gap after the body with `; }`.
// The BODY ITSELF is never in either range, so nothing inside it can be lost -- the edits touch
// only whitespace on each side of it.
//
// Declines when the body's own type is not void-like, which is upstream's `canFix`. The case that
// needs it is a body whose type came from somewhere other than the reported expression, where
// wrapping in braces would change the arrow's return type from that value to undefined.
func noConfusingVoidExpressionArrowFix(ctx rule.Context, arrow *ast.Node) ([]rule.Fix, bool) {
	arrowFunction := arrow.AsArrowFunction()
	if arrowFunction == nil || arrowFunction.Body == nil {
		return nil, false
	}
	body := arrowFunction.Body
	if body.Kind == ast.KindBlock {
		// Already braced; there is nothing for this arm to repair.
		return nil, false
	}
	// Our parser keeps a parenthesis node upstream's folds away, so the body of
	// `foo => (foo ? a : b)` is a KindParenthesizedExpression here and the conditional itself
	// upstream. Unwrapping matters for the OUTPUT rather than for the verdict: upstream's own
	// vector for that case is `foo => { foo ? a : b; }`, with the parentheses gone, and wrapping
	// the parenthesized node instead writes `{ (foo ? a : b); }`.
	//
	// `wrappedBody` is what the braces actually enclose, and the comment scan below has to bound on
	// it: after unwrapping, the text between the arrow token and the inner body is `(`, which is
	// neither whitespace nor a comment, and scanning against the inner node alone would decline
	// every parenthesized arrow as if it held one.
	wrappedBody := body
	for body.Kind == ast.KindParenthesizedExpression {
		inner := body.AsParenthesizedExpression().Expression
		if inner == nil {
			return nil, false
		}
		body = inner
	}
	if ctx.TypeChecker == nil {
		return nil, false
	}
	bodyType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, body)
	if !type_checking.IsTypeFlagSet(bodyType, checker.TypeFlagsVoidLike) {
		return nil, false
	}

	arrowToken := arrowFunction.EqualsGreaterThanToken
	if arrowToken == nil {
		return nil, false
	}
	bodySpan := rule.TokenRange(ctx.SourceFile, body)
	wrappedSpan := rule.TokenRange(ctx.SourceFile, wrappedBody)
	arrowSpan := rule.TokenRange(ctx.SourceFile, arrowToken)
	functionSpan := rule.TokenRange(ctx.SourceFile, arrow)
	if arrowSpan.End() > wrappedSpan.Pos() || wrappedSpan.End() > functionSpan.End() ||
		bodySpan.Pos() < wrappedSpan.Pos() || bodySpan.End() > wrappedSpan.End() {
		return nil, false
	}

	// A comment between the arrow and the body, or after the body, sits inside a replaced range
	// and would be deleted. Declined rather than repaired, matching this tree's standing rule that
	// a fixer withholds rather than silently removing what it was not asked to remove. Upstream
	// does replace those gaps unconditionally, so this is a deliberate narrowing and it costs a
	// repair upstream would offer rather than producing a different one.
	source := ctx.SourceFile.Text()
	if confusingVoidGapHoldsComment(source[arrowSpan.End():wrappedSpan.Pos()]) ||
		confusingVoidGapHoldsComment(source[wrappedSpan.End():functionSpan.End()]) {
		return nil, false
	}

	return []rule.Fix{
		// The edits replace the gaps around the WRAPPED body, so a parenthesis pair is removed
		// along with them, which is what upstream's own vector for `foo => (a ? b : c)` shows.
		// The body's own bytes are in neither range, so nothing inside it can be lost.
		rule.ReplaceRange(arrowSpan.WithPos(arrowSpan.End()).WithEnd(bodySpan.Pos()), " { "),
		rule.ReplaceRange(bodySpan.WithPos(bodySpan.End()).WithEnd(functionSpan.End()), "; }"),
	}, true
}

// confusingVoidGapHoldsComment says whether a stretch between two tokens holds anything other than
// whitespace, which for a gap inside an expression means a comment.
func confusingVoidGapHoldsComment(gap string) bool {
	for index := 0; index < len(gap); index++ {
		switch gap[index] {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			continue
		}
		return true
	}
	return false
}

// isConfusingVoidReturningFunction is upstream's `isVoidReturningFunctionNode`, which only runs
// under `ignoreVoidReturningFunctions`.
//
// Two routes, in upstream's order. An explicit return annotation is read directly and answers on
// its own. Failing that, and only for a function EXPRESSION or an arrow, the contextual type is
// consulted, so `const handler: () => void = () => f()` is exempt while a function declaration
// with no annotation is not.
func isConfusingVoidReturningFunction(ctx rule.Context, function *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}

	if returnType := confusingVoidReturnTypeNode(function); returnType != nil {
		return confusingVoidTypeIncludesVoid(ctx.TypeChecker.GetTypeFromTypeNode(returnType))
	}

	if function.Kind == ast.KindFunctionDeclaration {
		return false
	}

	// The CHECKER's own contextual type, not the shelf's `type_checking.GetContextualType`.
	//
	// That shelf helper is a hand-rolled approximation covering the call-argument case and a few
	// declaration shapes, and it answers nil for every shape upstream's corpus uses here: an arrow
	// in an object literal typed by its annotation, the same under an `as` assertion, and a
	// parenthesized arrow asserted to a function type. Measured on seven shapes, it answered nil
	// on six of them and this one answered all seven.
	//
	// Named as a divergence rather than a fix to the shelf, because the shelf helper has other
	// callers whose expectations were not measured here.
	contextualType := checker.Checker_getContextualType(ctx.TypeChecker, function, checker.ContextFlagsNone)
	if contextualType == nil {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(contextualType) {
		for _, signature := range ctx.TypeChecker.GetSignaturesOfType(part, checker.SignatureKindCall) {
			if confusingVoidTypeIncludesVoid(ctx.TypeChecker.GetReturnTypeOfSignature(signature)) {
				return true
			}
		}
	}
	return false
}

// confusingVoidReturnTypeNode returns a function's explicit return annotation, or nil.
//
// The method and accessor kinds are here because of a parser difference rather than a widening.
// ESTree gives a class method or an object-literal method a `FunctionExpression` VALUE, so
// upstream's three-type walk finds one and reads its annotation. Our parser has no such inner
// node: `test(): void {}` is a single `KindMethodDeclaration` carrying the annotation itself.
// Measured: without these arms, upstream's two clean cases for a void-annotated method both
// report, because the annotation is never read and the contextual route does not cover a method.
func confusingVoidReturnTypeNode(function *ast.Node) *ast.Node {
	switch function.Kind {
	case ast.KindArrowFunction:
		return function.AsArrowFunction().Type
	case ast.KindFunctionDeclaration:
		return function.AsFunctionDeclaration().Type
	case ast.KindFunctionExpression:
		return function.AsFunctionExpression().Type
	case ast.KindMethodDeclaration:
		return function.AsMethodDeclaration().Type
	case ast.KindGetAccessor:
		return function.AsGetAccessorDeclaration().Type
	case ast.KindSetAccessor:
		return function.AsSetAccessorDeclaration().Type
	}
	return nil
}

// confusingVoidTypeIncludesVoid says whether a type has a void constituent.
//
// Upstream tests `isIntrinsicVoidType` per union constituent rather than a flag mask over the
// whole type, which is the narrower question: `undefined` is VoidLike and is not void, so a
// function annotated `undefined` is not void-returning.
func confusingVoidTypeIncludesVoid(subject *checker.Type) bool {
	if subject == nil {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(subject) {
		if type_checking.IsIntrinsicVoidType(part) {
			return true
		}
	}
	return false
}

// noConfusingVoidExpressionReturnFix rewrites a `return X;` whose value is void.
//
// Two shapes, and `final` picks between them. A final return becomes the bare expression, since
// falling off the end returns undefined anyway. A non-final one becomes the expression followed by
// a bare `return`, which preserves the control flow the return was there for.
//
// # Why constructing a span is safe here
//
// The replaced span is the whole return STATEMENT, and the text written back contains the
// argument's own source copied byte for byte. Nothing is rebuilt from node properties, so a type
// assertion, a `satisfies`, a non-null `!`, type arguments and any comment inside the argument all
// survive. What does NOT survive is anything between the `return` keyword and the argument, or
// between the argument and the semicolon, which is why a comment there declines the repair below.
//
// # The semicolon that goes in front
//
// Upstream prepends `;` when the argument begins with `(`, `[` or a backtick, because moving it to
// the start of a line would otherwise continue the previous line rather than start a new statement.
// One of upstream's cases is exactly that: a `return ['bar','baz'].forEach(...)` after an
// unterminated `console.log('foo')`.
//
// # The braces
//
// A non-final return that is not directly inside a block, `if (cond) return f();`, becomes two
// statements and therefore needs braces. Upstream adds them and so does this.
func noConfusingVoidExpressionReturnFix(ctx rule.Context, returnStatement *ast.Node,
	final bool) (rule.Fix, bool) {

	statement := returnStatement.AsReturnStatement()
	if statement == nil || statement.Expression == nil {
		return rule.Fix{}, false
	}
	argument := statement.Expression

	// Upstream's `canFix`: the ARGUMENT's own type must be void-like. On a final return the
	// argument is what remains, and on a non-final one it becomes its own statement, so a
	// non-void argument would have its value silently discarded.
	if ctx.TypeChecker == nil {
		return rule.Fix{}, false
	}
	argumentType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
	if !type_checking.IsTypeFlagSet(argumentType, checker.TypeFlagsVoidLike) {
		return rule.Fix{}, false
	}

	source := ctx.SourceFile.Text()
	statementSpan := rule.TokenRange(ctx.SourceFile, returnStatement)
	argumentSpan := rule.TokenRange(ctx.SourceFile, argument)
	if statementSpan.Pos() > argumentSpan.Pos() || argumentSpan.End() > statementSpan.End() {
		return rule.Fix{}, false
	}

	// A comment between `return` and the argument, or after the argument, sits inside the replaced
	// span and would be deleted. Declined rather than repaired, which is a deliberate narrowing:
	// upstream replaces those stretches unconditionally and would lose the comment.
	if confusingVoidGapHoldsComment(source[statementSpan.Pos()+len("return"):argumentSpan.Pos()]) ||
		confusingVoidGapHoldsCommentBeforeSemicolon(source[argumentSpan.End():statementSpan.End()]) {
		return rule.Fix{}, false
	}

	replacement := source[argumentSpan.Pos():argumentSpan.End()] + ";"
	if !final {
		replacement += " return;"
	}
	if confusingVoidIsPreventingAutomaticSemicolonInsertion(source, argumentSpan.Pos()) {
		replacement = ";" + replacement
	}
	if !final && returnStatement.Parent != nil && returnStatement.Parent.Kind != ast.KindBlock {
		replacement = "{ " + replacement + " }"
	}

	return rule.ReplaceRange(statementSpan, replacement), true
}

// confusingVoidGapHoldsCommentBeforeSemicolon is the trailing-gap variant, which tolerates the
// statement's own semicolon.
func confusingVoidGapHoldsCommentBeforeSemicolon(gap string) bool {
	for index := 0; index < len(gap); index++ {
		switch gap[index] {
		case ' ', '\t', '\n', '\r', '\v', '\f', ';':
			continue
		}
		return true
	}
	return false
}

// confusingVoidIsPreventingAutomaticSemicolonInsertion is upstream's `isPreventingASI`: whether the
// argument's first character would continue the previous line rather than start a statement.
func confusingVoidIsPreventingAutomaticSemicolonInsertion(source string, start int) bool {
	if start < 0 || start >= len(source) {
		return false
	}
	switch source[start] {
	case '(', '[', '`':
		return true
	}
	return false
}
