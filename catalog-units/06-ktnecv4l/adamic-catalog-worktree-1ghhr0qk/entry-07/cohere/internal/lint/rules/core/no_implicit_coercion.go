package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

const messageNoImplicitCoercionId = "implicitCoercion"
const messageNoImplicitCoercionSuggestionId = "useRecommendation"

// noImplicitCoercionMessage renders the finding.
//
// Upstream's template is "Unexpected implicit coercion encountered. Use `{{recommendation}}`
// instead." and the recommendation is computed per finding from the operand's own source text, so
// the tests assert the rendered sentence rather than only the id.
func noImplicitCoercionMessage(recommendation string) rule.Message {
	return rule.Message{
		Id: messageNoImplicitCoercionId,
		Description: fmt.Sprintf("Unexpected implicit coercion encountered. Use `%s` instead. "+
			"The shorthand converts a type as a side effect of an operator chosen for something "+
			"else, so a reader has to know the coercion table to see what it produces. The "+
			"explicit call says which type is wanted, and says it where the conversion happens.",
			recommendation),
	}
}

// noImplicitCoercionSuggestionMessage is upstream's second message, offered rather than applied.
func noImplicitCoercionSuggestionMessage(recommendation string) rule.Message {
	return rule.Message{
		Id:          messageNoImplicitCoercionSuggestionId,
		Description: fmt.Sprintf("Use `%s` instead.", recommendation),
	}
}

// NoImplicitCoercionOptions carries upstream's five options.
//
// Three of the four booleans default to TRUE, which is the trap this project has been bitten by:
// `rule.DecodeOptionsInto` on a default-true option yields a zero-value struct that silently turns
// the rule off, and every fixture built from a struct rather than routed through the decoder passes
// anyway. Pointers keep an absent key distinguishable from an explicit false.
type NoImplicitCoercionOptions struct {
	Boolean                   *bool    `json:"boolean"`
	Number                    *bool    `json:"number"`
	String                    *bool    `json:"string"`
	DisallowTemplateShorthand *bool    `json:"disallowTemplateShorthand"`
	Allow                     []string `json:"allow"`
}

// DefaultNoImplicitCoercionSettings is upstream's `defaultOptions`.
func DefaultNoImplicitCoercionSettings() NoImplicitCoercionOptions {
	enabled := true
	disabled := false
	return NoImplicitCoercionOptions{
		Boolean:                   &enabled,
		Number:                    &enabled,
		String:                    &enabled,
		DisallowTemplateShorthand: &disabled,
		Allow:                     nil,
	}
}

// DecodeNoImplicitCoercionOptions turns the configured object into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` because three of the four booleans default to
// true. The generic helper errors on empty input, the config layer turns that into nil, and the
// rule reads nil as a zero-valued struct with every boolean false, which disables three of the four
// judgments. That is the live config's own shape for a rule configured as a bare severity, so it is
// the ordinary path rather than a corner.
func DecodeNoImplicitCoercionOptions(raw []byte) (any, error) {
	settings := DefaultNoImplicitCoercionSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var configured NoImplicitCoercionOptions
	if err := rule.UnmarshalOptions(raw, &configured); err != nil {
		return settings, err
	}
	if configured.Boolean != nil {
		settings.Boolean = configured.Boolean
	}
	if configured.Number != nil {
		settings.Number = configured.Number
	}
	if configured.String != nil {
		settings.String = configured.String
	}
	if configured.DisallowTemplateShorthand != nil {
		settings.DisallowTemplateShorthand = configured.DisallowTemplateShorthand
	}
	settings.Allow = configured.Allow
	return settings, nil
}

// noImplicitCoercionAllows answers whether an operator is on the allow list.
func noImplicitCoercionAllows(settings NoImplicitCoercionOptions, operator string) bool {
	for _, allowed := range settings.Allow {
		if allowed == operator {
			return true
		}
	}
	return false
}

// NoImplicitCoercion flags the shorthand type conversions and recommends the explicit call.
//
//	valid:   var b = Boolean(foo);
//	valid:   var n = Number(foo);
//	valid:   var s = String(foo);
//	valid:   var n = +1;              a numeric operand is already a number
//	valid:   var s = '' + 'foo';      concatenating two strings is not a coercion
//	invalid: var b = !!foo;           fixed to Boolean(foo)
//	invalid: var n = +foo;            suggested, not fixed
//	invalid: var s = '' + foo;        suggested, not fixed
//	invalid: var b = ~foo.indexOf(bar);   reported bare, neither fixed nor suggested
//
// # Nine arms, and only one of them carries a fix
//
// This is the fact that decides how the rule is tested, and it is not what the rule's `fixable`
// metadata suggests. Measured against the installed build at 10.8.1, arm by arm:
//
//	!!foo               FIX, and only when `Boolean` resolves to the global
//	+foo                suggestion only
//	-(-foo)             suggestion only
//	1 * foo             suggestion only
//	foo - 0             suggestion only
//	'' + foo            suggestion only
//	foo + ''            suggestion only
//	foo += ''           suggestion only
//	~foo.indexOf(bar)   NEITHER, reported bare
//
// The corpus agrees: 48 reporting cases carry 4 non-null outputs and 44 declines. A fix is applied
// unattended and a suggestion is not, so shipping the eight as fixes would rewrite code upstream
// only offers to rewrite.
//
// # Why `!!foo` is the only fixable one
//
// `Boolean(foo)` produces exactly what `!!foo` produces, so the rewrite preserves meaning.
// `Number(foo)` does not always match a unary plus, and `String(foo)` does not always match a
// concatenation with an empty string, once `Symbol`, `BigInt` or a custom `valueOf` is
// involved, which is why upstream offers those
// rather than applying them. Reproduced rather than improved on.
//
// # The `Boolean` global, and what withdraws the fix
//
// Upstream asks its scope manager whether `Boolean` has any declaration in source, and withholds
// the fix when it does, because `Boolean(foo)` would then call something else. Measured here
// through `ResolveName`: a `var`, `let`, parameter, function or class named `Boolean` each withdraw
// the fix, while a shadow in a SIBLING scope does not.
//
// Upstream's corpus writes the same input twice with opposite verdicts, separated only by
// `languageOptions.globals.Boolean = "off"`, which is the same question asked from the other side:
// the fix is withheld when `Boolean` is not a global, rather than only when it is shadowed. That
// pair reads as a contradiction until the language options are rendered beside the source.
//
// # The `~indexOf` arm, which has real semantics behind it
//
// `~n` is falsy exactly when `n` is `-1`, which is what makes `~foo.indexOf(bar)` a membership test.
// The recommendation depends on whether the call is optionally chained, and this is not cosmetic:
// `foo?.indexOf(bar)` yields `undefined` when `foo` is nullish, and `undefined !== -1` is TRUE,
// which would invert the test. So upstream recommends `>= 0` for the chained form and `!== -1`
// otherwise, and offers neither as a fix nor as a suggestion because the reader has to choose.
//
// Measured: `~foo.indexOf(x)` and `~foo.lastIndexOf(x)` report, `~foo['indexOf'](x)` reports because
// the key is static, and `~foo[indexOf](x)`, `~foo.someOther(x)`, `~foo` and `~foo.indexOf` are all
// clean.
var NoImplicitCoercion = rule.Rule{
	Name:             "no-implicit-coercion",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[NoImplicitCoercionOptions](options)
		defaults := DefaultNoImplicitCoercionSettings()
		if settings.Boolean == nil {
			settings.Boolean = defaults.Boolean
		}
		if settings.Number == nil {
			settings.Number = defaults.Number
		}
		if settings.String == nil {
			settings.String = defaults.String
		}
		if settings.DisallowTemplateShorthand == nil {
			settings.DisallowTemplateShorthand = defaults.DisallowTemplateShorthand
		}

		return rule.Listeners{
			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				noImplicitCoercionCheckUnary(ctx, node, settings)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				noImplicitCoercionCheckBinary(ctx, node, settings)
			},
			ast.KindTemplateExpression: func(node *ast.Node) {
				noImplicitCoercionCheckTemplate(ctx, node, settings)
			},
		}
	},
}

// noImplicitCoercionReport emits the finding, with a fix, a suggestion, or neither.
//
// Upstream funnels every arm through one `report` taking two booleans, and this is the same shape:
// `shouldFix` wins over `shouldSuggest`, so an arm never offers both.
func noImplicitCoercionReport(
	ctx rule.Context,
	node *ast.Node,
	recommendation string,
	shouldSuggest bool,
	shouldFix bool,
) {
	message := noImplicitCoercionMessage(recommendation)
	replacement := noImplicitCoercionReplacementText(ctx, node, recommendation)

	switch {
	case shouldFix:
		ctx.ReportNodeWithFixes(node, message,
			rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node).WithEnd(node.End()), replacement))
	case shouldSuggest:
		ctx.ReportNodeWithSuggestions(node, message, rule.Suggestion{
			Message: noImplicitCoercionSuggestionMessage(recommendation),
			Fixes: []rule.Fix{
				rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node).WithEnd(node.End()), replacement),
			},
		})
	default:
		ctx.ReportNode(node, message)
	}
}

// noImplicitCoercionReplacementText prepends a space when the recommendation would otherwise lex
// into the token before it.
//
// Upstream asks a tokenizer whether the two can sit adjacent. The only shapes that can reach this
// are a recommendation opening with `B`, `N`, `S` or the operand's own first character, so the
// hazard is an identifier or keyword running into the replacement: `typeof+foo` becoming
// `typeofNumber(foo)`. Testing whether the preceding character can continue an identifier covers
// exactly that.
func noImplicitCoercionReplacementText(ctx rule.Context, node *ast.Node, recommendation string) string {
	start := rule.TokenRange(ctx.SourceFile, node).Pos()
	if start == 0 {
		return recommendation
	}
	text := ctx.SourceFile.Text()
	previous := text[start-1]
	if noImplicitCoercionContinuesIdentifier(previous) && noImplicitCoercionContinuesIdentifier(recommendation[0]) {
		return " " + recommendation
	}
	return recommendation
}

// noImplicitCoercionContinuesIdentifier answers whether a byte can sit inside an identifier.
func noImplicitCoercionContinuesIdentifier(character byte) bool {
	switch {
	case character >= 'a' && character <= 'z',
		character >= 'A' && character <= 'Z',
		character >= '0' && character <= '9',
		character == '_', character == '$':
		return true
	}
	return false
}

// noImplicitCoercionOperandText is upstream's `getOperandText`.
//
// A sequence expression must be parenthesized or its commas would be read as argument separators,
// turning `!!(a, b)` into `Boolean(a, b)`, which evaluates something else. Our parser gives a
// sequence expression a comma-operator binary expression rather than its own kind.
func noImplicitCoercionOperandText(ctx rule.Context, node *ast.Node) string {
	// Unwrap first. Our parser keeps `KindParenthesizedExpression`, which ESTree folds away, so
	// upstream's `getText(node)` on `!!(foo + bar)` yields `foo + bar` while ours yielded
	// `(foo + bar)` and the recommendation came out as `Boolean((foo + bar))`. The loop rather
	// than `ast.SkipParentheses` because `((x))` nests and that helper dereferences its argument.
	node = noImplicitCoercionUnwrapParentheses(node)
	if node == nil {
		return ""
	}
	text := ctx.SourceFile.Text()
	span := rule.TokenRange(ctx.SourceFile, node)
	source := text[span.Pos():node.End()]
	if node.Kind == ast.KindBinaryExpression {
		if operator := node.AsBinaryExpression().OperatorToken; operator != nil &&
			operator.Kind == ast.KindCommaToken {
			return "(" + source + ")"
		}
	}
	return source
}

// noImplicitCoercionResolvesToTheGlobalBoolean answers whether `Boolean` at this point is the
// global one.
//
// Upstream asks its scope manager for a variable named `Boolean` and requires it to have zero
// identifiers, meaning no declaration in source. Asked here through `ResolveName`, which resolves
// the name the way the checker would at that position, and requiring every declaration to be
// ambient.
//
// Measured across seven shapes before being built on: no shadow answers global; a `var`, `let`,
// parameter, function or class named `Boolean` each answer not-global; and a shadow in a sibling
// function does not suppress. Neutralizing the ambient test failed five of those and cutting the
// resolution failed the other two, so the predicate is falsifiable in both directions.
func noImplicitCoercionResolvesToTheGlobalBoolean(ctx rule.Context, at *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.ResolveName("Boolean", at, ast.SymbolFlagsValue, false)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !file.IsDeclarationFile {
			return false
		}
	}
	return true
}

// noImplicitCoercionRewrite is which repair a `!!x` can take without changing what the program means
// to the compiler.
type noImplicitCoercionRewrite int

const (
	noImplicitCoercionRewriteDecline noImplicitCoercionRewrite = iota
	noImplicitCoercionRewriteBoolean
	noImplicitCoercionRewriteComparison
)

// noImplicitCoercionNarrowingRewrite chooses the repair for `!!x` that keeps TypeScript's narrowing.
//
// `Boolean(x)` evaluates to what `!!x` evaluates to, which is the whole of upstream's argument for
// the fix, and it is not the whole of what the expression means. TypeScript narrows a reference
// through `!!x` and through a comparison, including through a const that aliases the result, and it
// does not narrow through a call. So `const hasData = !!data; hasData ? data.energy : 0` compiles and
// its `Boolean(data)` rewrite is TS18048. That broke www-phi-health's
// `AppSidebarEnergyBalance.tsx` and seven ahra sites in one morning, each repaired by hand to the
// comparison this now writes.
//
// Three answers, decided by the operand and its type:
//
//	not a reference (a call, a literal)            Boolean(x), since nothing there can be narrowed
//	narrowing changes nothing (string, any, Foo)   Boolean(x), upstream's fix
//	nullish plus always-truthy (Foo | undefined)   the comparison: x !== undefined, x !== null, or both
//	anything else (number | undefined, boolean)    declined: reported and suggested, never applied
//
// The comparison is chosen only when it is exactly `!!x`: every constituent besides null and
// undefined must be one no value of which is falsy, so removing the nullish ones is removing every
// falsy value. A type also holding `0`, `false` or an empty string has no comparison that means `!!x`, and the
// truthiness narrowing there is visible, so neither rewrite is safe and a person decides.
func noImplicitCoercionNarrowingRewrite(ctx rule.Context, node *ast.Node, operand *ast.Node) (noImplicitCoercionRewrite, string) {
	if !noImplicitCoercionIsNarrowableReference(operand) {
		return noImplicitCoercionRewriteBoolean, ""
	}
	if ctx.TypeChecker == nil {
		return noImplicitCoercionRewriteDecline, ""
	}
	operandType := ctx.TypeChecker.GetTypeAtLocation(operand)
	if operandType == nil {
		return noImplicitCoercionRewriteDecline, ""
	}

	hasNull, hasUndefined, hasPossiblyFalsy := false, false, false
	for constituent := range type_checking.UnionTypePartsSeq(operandType) {
		switch {
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsAny):
			// `any` absorbs a union, and truthiness does not narrow it, so nothing can be lost.
			return noImplicitCoercionRewriteBoolean, ""
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsUnknown|checker.TypeFlagsInstantiable):
			// `unknown` narrows to `{}` and a type parameter to `T & {}`, neither expressible as a
			// comparison and both visible to a later line.
			return noImplicitCoercionRewriteDecline, ""
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsNull):
			hasNull = true
		case type_checking.IsTypeFlagSet(constituent, checker.TypeFlagsUndefined|checker.TypeFlagsVoid):
			hasUndefined = true
		case noImplicitCoercionIsFalsyLiteral(constituent):
			// `false`, `0` or `''`: truthiness removes it, so the narrowing is visible, and no
			// comparison against null or undefined removes it.
			return noImplicitCoercionRewriteDecline, ""
		case !noImplicitCoercionIsAlwaysTruthy(constituent):
			// A wide `string`, `number` or `bigint`: truthiness narrows it to itself, so on its own
			// it costs nothing, and beside a nullish member it rules the comparison out.
			hasPossiblyFalsy = true
		}
	}

	switch {
	case !hasNull && !hasUndefined:
		return noImplicitCoercionRewriteBoolean, ""
	case hasPossiblyFalsy:
		return noImplicitCoercionRewriteDecline, ""
	}

	reference := noImplicitCoercionOperandText(ctx, operand)
	var comparison string
	switch {
	case hasNull && hasUndefined:
		comparison = reference + " !== null && " + reference + " !== undefined"
	case hasNull:
		comparison = reference + " !== null"
	default:
		comparison = reference + " !== undefined"
	}
	if noImplicitCoercionComparisonNeedsParentheses(node) {
		comparison = "(" + comparison + ")"
	}
	return noImplicitCoercionRewriteComparison, comparison
}

// noImplicitCoercionIsNarrowableReference answers whether TypeScript can narrow this expression.
//
// Narrowing applies to references: a name, `this`, and property or element access built on them.
// Anything else, a call above all, is evaluated fresh wherever it appears, so there is no narrowing
// for `Boolean(x)` to lose and upstream's fix stands.
func noImplicitCoercionIsNarrowableReference(node *ast.Node) bool {
	for node != nil {
		switch node.Kind {
		case ast.KindIdentifier, ast.KindThisKeyword, ast.KindSuperKeyword:
			return true
		case ast.KindPropertyAccessExpression:
			node = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			node = node.AsElementAccessExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		default:
			return false
		}
	}
	return false
}

// noImplicitCoercionIsFalsyLiteral answers whether a type is one falsy literal: `false`, `0`, `-0`,
// `NaN`, the empty string, or a bigint literal. A bigint literal counts whatever its value, which is the safe
// direction: it declines a rewrite rather than writing a comparison that might not mean `!!x`.
func noImplicitCoercionIsFalsyLiteral(candidate *checker.Type) bool {
	switch {
	case type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsBigIntLiteral):
		return true
	case type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsBooleanLiteral|checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral):
		switch value := candidate.AsLiteralType().Value().(type) {
		case bool:
			return !value
		case string:
			return value == ""
		case nil:
			// A computed enum member, whose value the checker does not hold.
			return true
		default:
			rendered := fmt.Sprint(value)
			return rendered == "0" || rendered == "-0" || rendered == "NaN"
		}
	}
	return false
}

// noImplicitCoercionIsAlwaysTruthy answers whether no value of a type is falsy: an object, a symbol,
// or a truthy literal. An intersection qualifies only when every part does, so a branded primitive
// such as `number & { unit: 'pixels' }` is still a number that can be zero.
func noImplicitCoercionIsAlwaysTruthy(candidate *checker.Type) bool {
	if type_checking.IsIntersectionType(candidate) {
		for _, part := range candidate.Types() {
			if !noImplicitCoercionIsAlwaysTruthy(part) {
				return false
			}
		}
		return true
	}
	switch {
	case type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive|checker.TypeFlagsESSymbolLike):
		return true
	case type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsBooleanLiteral|checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral):
		return !noImplicitCoercionIsFalsyLiteral(candidate)
	}
	return false
}

// noImplicitCoercionComparisonNeedsParentheses answers whether the comparison must be wrapped where
// `!!x` stood.
//
// `!!x` binds tighter than anything around it, and a comparison does not: `fallback ?? !!x` becomes
// `fallback ?? x !== null && x !== undefined`, which is a syntax error, and `await !!x` would compare
// the awaited value. So the comparison goes bare only where the surrounding grammar takes any
// expression, or where `&&` and `||` would bind looser than it anyway, and is wrapped everywhere else.
func noImplicitCoercionComparisonNeedsParentheses(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindParenthesizedExpression, ast.KindIfStatement,
		ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindReturnStatement,
		ast.KindArrowFunction, ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAssignment,
		ast.KindJsxExpression, ast.KindConditionalExpression, ast.KindExpressionStatement,
		ast.KindArrayLiteralExpression, ast.KindTemplateSpan, ast.KindExportAssignment,
		ast.KindPropertyDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindThrowStatement,
		ast.KindCaseClause, ast.KindSwitchStatement:
		return false
	case ast.KindBinaryExpression:
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindEqualsToken, ast.KindCommaToken:
			return false
		}
	}
	return true
}

// noImplicitCoercionUnwrapParentheses strips every layer of parentheses.
//
// Written as a loop with its own nil check rather than calling `ast.SkipParentheses`, which
// dereferences its argument, and every operand reaching here can be nil from a recovered parse.
func noImplicitCoercionUnwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// noImplicitCoercionCheckUnary handles the four prefix-operator arms.
func noImplicitCoercionCheckUnary(ctx rule.Context, node *ast.Node, settings NoImplicitCoercionOptions) {
	unary := node.AsPrefixUnaryExpression()
	operand := noImplicitCoercionUnwrapParentheses(unary.Operand)
	if operand == nil {
		return
	}

	// !!foo
	if *settings.Boolean && !noImplicitCoercionAllows(settings, "!!") &&
		unary.Operator == ast.KindExclamationToken &&
		operand.Kind == ast.KindPrefixUnaryExpression &&
		operand.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
		inner := noImplicitCoercionUnwrapParentheses(operand.AsPrefixUnaryExpression().Operand)
		// Already a boolean, `!!b` converts nothing and `Boolean(b)` is no better advice.
		if inner != nil && !noImplicitCoercionIsAlready(ctx, inner, checker.TypeFlagsBooleanLike) {
			recommendation := "Boolean(" + noImplicitCoercionOperandText(ctx, inner) + ")"
			switch rewrite, comparison := noImplicitCoercionNarrowingRewrite(ctx, node, inner); rewrite {
			case noImplicitCoercionRewriteComparison:
				noImplicitCoercionReport(ctx, node, comparison, true, true)
			case noImplicitCoercionRewriteBoolean:
				// The fix is withheld when `Boolean` is not the global, because the rewrite would
				// then call whatever the name binds to instead. The finding still fires, as a
				// suggestion.
				booleanIsGlobal := noImplicitCoercionResolvesToTheGlobalBoolean(ctx, node)
				noImplicitCoercionReport(ctx, node, recommendation, true, booleanIsGlobal)
			default:
				noImplicitCoercionReport(ctx, node, recommendation, true, false)
			}
		}
	}

	// ~foo.indexOf(bar)
	if *settings.Boolean && !noImplicitCoercionAllows(settings, "~") &&
		unary.Operator == ast.KindTildeToken {
		if call, chained, ok := noImplicitCoercionIndexOfCall(operand); ok {
			comparison := "!== -1"
			if chained {
				// A chained call yields undefined when the receiver is nullish, and
				// `undefined !== -1` is true, which would invert the membership test.
				comparison = ">= 0"
			}
			recommendation := ctx.NodeText(call) + " " + comparison
			noImplicitCoercionReport(ctx, node, recommendation, false, false)
		}
	}

	// +foo
	if *settings.Number && !noImplicitCoercionAllows(settings, "+") &&
		unary.Operator == ast.KindPlusToken && !noImplicitCoercionIsNumeric(operand) &&
		!noImplicitCoercionIsAlready(ctx, operand, checker.TypeFlagsNumberLike) {
		recommendation := "Number(" + noImplicitCoercionOperandText(ctx, operand) + ")"
		noImplicitCoercionReport(ctx, node, recommendation, true, false)
	}

	// -(-foo)
	if *settings.Number && !noImplicitCoercionAllows(settings, "- -") &&
		unary.Operator == ast.KindMinusToken &&
		operand.Kind == ast.KindPrefixUnaryExpression &&
		operand.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		inner := noImplicitCoercionUnwrapParentheses(operand.AsPrefixUnaryExpression().Operand)
		// A bigint too: `-(-x)` is an identity on a bigint, where `+x`, `1 * x` and `x - 0` throw.
		if inner != nil && !noImplicitCoercionIsNumeric(inner) &&
			!noImplicitCoercionIsAlready(ctx, inner, checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
			recommendation := "Number(" + noImplicitCoercionOperandText(ctx, inner) + ")"
			noImplicitCoercionReport(ctx, node, recommendation, true, false)
		}
	}
}

// noImplicitCoercionCheckBinary handles the three binary arms and the compound assignment.
func noImplicitCoercionCheckBinary(ctx rule.Context, node *ast.Node, settings NoImplicitCoercionOptions) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken == nil {
		return
	}
	operator := binary.OperatorToken.Kind
	left, right := binary.Left, binary.Right
	if left == nil || right == nil {
		return
	}

	// foo += ""
	if operator == ast.KindPlusEqualsToken {
		if *settings.String && !noImplicitCoercionAllows(settings, "+") &&
			noImplicitCoercionIsEmptyString(right) &&
			!noImplicitCoercionIsAlready(ctx, left, checker.TypeFlagsStringLike) {
			code := ctx.NodeText(left)
			noImplicitCoercionReport(ctx, node, code+" = String("+code+")", true, false)
		}
		return
	}

	// 1 * foo
	if *settings.Number && !noImplicitCoercionAllows(settings, "*") &&
		operator == ast.KindAsteriskToken &&
		noImplicitCoercionIsMultiplyByOne(binary) &&
		!noImplicitCoercionIsMultiplyByFractionOfOne(node) {
		if operand := noImplicitCoercionNonNumericOperand(binary); operand != nil &&
			!noImplicitCoercionIsAlready(ctx, operand, checker.TypeFlagsNumberLike) {
			recommendation := "Number(" + noImplicitCoercionOperandText(ctx, operand) + ")"
			noImplicitCoercionReport(ctx, node, recommendation, true, false)
		}
	}

	// foo - 0
	if *settings.Number && !noImplicitCoercionAllows(settings, "-") &&
		operator == ast.KindMinusToken &&
		right.Kind == ast.KindNumericLiteral && right.Text() == "0" &&
		!noImplicitCoercionIsNumeric(left) &&
		!noImplicitCoercionIsAlready(ctx, left, checker.TypeFlagsNumberLike) {
		recommendation := "Number(" + noImplicitCoercionOperandText(ctx, left) + ")"
		noImplicitCoercionReport(ctx, node, recommendation, true, false)
	}

	// "" + foo
	if *settings.String && !noImplicitCoercionAllows(settings, "+") &&
		operator == ast.KindPlusToken &&
		noImplicitCoercionIsConcatWithEmptyString(binary) {
		operand := noImplicitCoercionNonEmptyOperand(binary)
		if !noImplicitCoercionIsAlready(ctx, operand, checker.TypeFlagsStringLike) {
			recommendation := "String(" + noImplicitCoercionOperandText(ctx, operand) + ")"
			noImplicitCoercionReport(ctx, node, recommendation, true, false)
		}
	}
}

// noImplicitCoercionIndexOfCall answers whether a node is a call to `indexOf` or `lastIndexOf` on
// some receiver, and whether that call is optionally chained.
//
// Upstream's `isBinaryNegatingOfIndexOf` skips a chain expression and then asks
// `isSpecificMemberAccess(callee, null, /^(?:i|lastI)ndexOf$/)`. The `null` receiver means ANY
// object, and the static-key requirement is why `foo['indexOf'](x)` reports and `foo[indexOf](x)`
// does not.
func noImplicitCoercionIndexOfCall(node *ast.Node) (call *ast.Node, chained bool, ok bool) {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil, false, false
	}
	// The chain flag is read from the callee BEFORE unwrapping, and the name after. Parenthesizing
	// a callee ends the optional chain, so `(foo?.indexOf)(1)` is not short-circuited and upstream
	// recommends `!== -1` for it while recommending `>= 0` for `foo?.indexOf(1)`. Measured against
	// the installed build; the two spellings differ only by those parentheses.
	rawCallee := node.AsCallExpression().Expression
	if rawCallee == nil {
		return nil, false, false
	}
	callee := noImplicitCoercionUnwrapParentheses(rawCallee)
	if callee == nil {
		return nil, false, false
	}
	parenthesizedCallee := rawCallee != callee

	var name string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access.Name() == nil {
			return nil, false, false
		}
		name = access.Name().Text()
		chained = access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		argument := access.ArgumentExpression
		if argument == nil || argument.Kind != ast.KindStringLiteral {
			// A computed key that is not a string literal is not a static property name, so
			// `foo[indexOf](x)` is clean.
			return nil, false, false
		}
		name = argument.Text()
		chained = access.QuestionDotToken != nil
	default:
		return nil, false, false
	}

	if name != "indexOf" && name != "lastIndexOf" {
		return nil, false, false
	}
	if parenthesizedCallee {
		chained = false
	}
	return node, chained, true
}

// noImplicitCoercionIsNumeric is upstream's `isNumeric`.
//
// A number literal, or a call to `Number`, `parseInt` or `parseFloat`. Named with the rule prefix
// because a package-level `isNumeric` already exists in this package and answers a different
// question.
func noImplicitCoercionIsNumeric(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindNumericLiteral {
		return true
	}
	if node.Kind == ast.KindCallExpression {
		callee := node.AsCallExpression().Expression
		if callee != nil && callee.Kind == ast.KindIdentifier {
			switch callee.Text() {
			case "Number", "parseInt", "parseFloat":
				return true
			}
		}
	}
	return false
}

// noImplicitCoercionIsAlready reports whether the checker proves an operand already has the type an
// arm would coerce it to, in which case nothing is coerced and there is nothing to report.
//
// This is where cohere is deliberately quieter than upstream, whose `isNumeric` is syntactic (a
// numeric literal or a call to `Number`, `parseInt` or `parseFloat`). The real sites are
// `Map.tsx:747/765/784` and `MapDrawing.ts:220` in ahra, `1 * zoom` with `zoom: number`, where ESLint
// reports and recommends `Number(zoom)`, which converts nothing. The string arms (`"" + s`, `s + ""`,
// `s += ""`) and the boolean arm (`!!b`) ask the same question with string-like and boolean-like
// flags (#vsy2eym): `String(s)` on a string and `Boolean(b)` on a boolean convert nothing either.
//
// Every union constituent must carry one of `kinds`; an intersection constituent qualifies when any of
// its parts does, which is a branded number such as `number & { unit: 'pixels' }`. So `any`,
// `unknown`, `number | undefined`, `string | undefined`, `boolean | null` and an error type all keep
// reporting. The type is read through a
// type parameter's constraint, since `T extends number` is a number at run time.
func noImplicitCoercionIsAlready(ctx rule.Context, operand *ast.Node, kinds checker.TypeFlags) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	operandType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, operand)
	if operandType == nil {
		return false
	}
	for constituent := range type_checking.UnionTypePartsSeq(operandType) {
		if type_checking.IsTypeFlagSet(constituent, kinds) {
			continue
		}
		branded := false
		if type_checking.IsIntersectionType(constituent) {
			for _, part := range constituent.Types() {
				if type_checking.IsTypeFlagSet(part, kinds) {
					branded = true
					break
				}
			}
		}
		if !branded {
			return false
		}
	}
	return true
}

// noImplicitCoercionIsEmptyString is upstream's `isEmptyString`.
//
// An empty string literal, or a template literal with no substitutions and empty text. Upstream
// tests the COOKED value, which `Text()` already holds.
func noImplicitCoercionIsEmptyString(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text() == ""
	}
	return false
}

// noImplicitCoercionIsStringType is upstream's `isStringType`.
//
// A string literal of any kind, or a call to `String`. Concatenating two strings is not a coercion,
// which is what keeps an empty string added to a string literal, and to a String call, clean.
func noImplicitCoercionIsStringType(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsStringLiteralLike(node) || node.Kind == ast.KindTemplateExpression {
		return true
	}
	if node.Kind == ast.KindCallExpression {
		callee := node.AsCallExpression().Expression
		return callee != nil && callee.Kind == ast.KindIdentifier && callee.Text() == "String"
	}
	return false
}

// noImplicitCoercionIsConcatWithEmptyString is upstream's `isConcatWithEmptyString`.
func noImplicitCoercionIsConcatWithEmptyString(binary *ast.BinaryExpression) bool {
	return (noImplicitCoercionIsEmptyString(binary.Left) && !noImplicitCoercionIsStringType(binary.Right)) ||
		(noImplicitCoercionIsEmptyString(binary.Right) && !noImplicitCoercionIsStringType(binary.Left))
}

// noImplicitCoercionNonEmptyOperand is upstream's `getNonEmptyOperand`.
func noImplicitCoercionNonEmptyOperand(binary *ast.BinaryExpression) *ast.Node {
	if noImplicitCoercionIsEmptyString(binary.Left) {
		return binary.Right
	}
	return binary.Left
}

// noImplicitCoercionIsMultiplyByOne is upstream's `isMultiplyByOne`.
func noImplicitCoercionIsMultiplyByOne(binary *ast.BinaryExpression) bool {
	return (binary.Left != nil && binary.Left.Kind == ast.KindNumericLiteral && binary.Left.Text() == "1") ||
		(binary.Right != nil && binary.Right.Kind == ast.KindNumericLiteral && binary.Right.Text() == "1")
}

// noImplicitCoercionIsMultiplyByFractionOfOne is upstream's `isMultiplyByFractionOfOne`.
//
// `a * 1 / b` parses as `(a * 1) / b`, so the `* 1` is technically a multiply by one while the
// expression reads as `a * (1 / b)`. Upstream declines it, and only when the multiplication is not
// itself parenthesized.
func noImplicitCoercionIsMultiplyByFractionOfOne(node *ast.Node) bool {
	binary := node.AsBinaryExpression()
	if binary.Right == nil || binary.Right.Kind != ast.KindNumericLiteral || binary.Right.Text() != "1" {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind == ast.KindParenthesizedExpression {
		// A parenthesized multiplication is its own expression, so the fraction reading does not
		// apply and upstream reports it.
		return false
	}
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	outer := parent.AsBinaryExpression()
	return outer.OperatorToken != nil && outer.OperatorToken.Kind == ast.KindSlashToken &&
		outer.Left == node
}

// noImplicitCoercionNonNumericOperand is upstream's `getNonNumericOperand`.
//
// Upstream checks the RIGHT operand first, which decides which one is named in the recommendation
// when neither is numeric.
func noImplicitCoercionNonNumericOperand(binary *ast.BinaryExpression) *ast.Node {
	// Both operands are unwrapped before the KIND test, which is the parser difference rather than
	// a convenience. Upstream skips an operand that is itself a `BinaryExpression`, because
	// `a * b * 1` should name the whole product rather than one factor. ESTree folds parentheses
	// away, so `(index + 1) * 1` reaches that test AS a binary expression and is declined; ours
	// wraps it, so without the unwrap the kind test missed it and the rule reported a site the
	// installed build leaves alone. Found on the real tree rather than in the corpus, which writes
	// no parenthesized operand here.
	right := noImplicitCoercionUnwrapParentheses(binary.Right)
	left := noImplicitCoercionUnwrapParentheses(binary.Left)
	if right != nil && !noImplicitCoercionIsEstreeBinary(right) && !noImplicitCoercionIsNumeric(right) {
		return right
	}
	if left != nil && !noImplicitCoercionIsEstreeBinary(left) && !noImplicitCoercionIsNumeric(left) {
		return left
	}
	return nil
}

// noImplicitCoercionIsEstreeBinary answers upstream's `type === "BinaryExpression"`.
//
// Our parser puts three ESTree node types under one kind, and the difference decides this rule:
// ESTree's `BinaryExpression` covers the arithmetic and comparison operators, while `&& || ??` are
// a `LogicalExpression` and the comma operator is a `SequenceExpression`. Measured against the
// installed build, `(a || b) * 1` and `(a, b) * 1` both REPORT while `(i + 1) * 1` is clean, so
// testing the kind alone is wrong in both directions.
func noImplicitCoercionIsEstreeBinary(node *ast.Node) bool {
	if node.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken
	if operator == nil {
		return false
	}
	switch operator.Kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
		ast.KindCommaToken:
		return false
	}
	return !ast.IsAssignmentOperator(operator.Kind)
}

// noImplicitCoercionCheckTemplate handles the `disallowTemplateShorthand` arm, which is off by
// default.
//
// The shape is a template whose entire content is one substitution and nothing else, so it exists
// only to convert: an empty head, one expression, an empty tail. A tagged template is exempt
// because the tag decides what the parts mean, and an expression that is already a string is not
// being coerced.
//
// Our parser splits a template into a head span and a list of `TemplateSpan`, each carrying an
// expression and the literal text that follows it. Upstream reads `quasis[0]` and `quasis[1]`,
// which are those two literal pieces.
func noImplicitCoercionCheckTemplate(ctx rule.Context, node *ast.Node, settings NoImplicitCoercionOptions) {
	if !*settings.DisallowTemplateShorthand {
		return
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
		return
	}

	template := node.AsTemplateExpression()
	if template.Head == nil || template.Head.Text() != "" {
		return
	}
	spans := template.TemplateSpans.Nodes
	if len(spans) != 1 {
		return
	}
	span := spans[0].AsTemplateSpan()
	if span.Literal == nil || span.Literal.Text() != "" {
		return
	}

	expression := noImplicitCoercionUnwrapParentheses(span.Expression)
	if expression == nil || noImplicitCoercionIsStringType(expression) {
		return
	}
	recommendation := "String(" + noImplicitCoercionOperandText(ctx, expression) + ")"
	noImplicitCoercionReport(ctx, node, recommendation, true, false)
}
