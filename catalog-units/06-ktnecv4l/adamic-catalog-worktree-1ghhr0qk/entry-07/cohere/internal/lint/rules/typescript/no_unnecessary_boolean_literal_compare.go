package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The five reporting messages, plus the configuration complaint.
//
// Upstream's ids are kept verbatim because a suppression comment names them, and the wording says
// what to write instead rather than restating the rule: each arm has a different replacement and a
// reader who cannot tell which arm fired cannot act on the finding.

func messageNoUnnecessaryBooleanLiteralCompareDirect() rule.Message {
	return rule.Message{
		Id: "direct",
		Description: "This compares a boolean to a boolean literal, which asks the same question " +
			"the value already answers. Use the expression on its own.",
	}
}

func messageNoUnnecessaryBooleanLiteralCompareNegated() rule.Message {
	return rule.Message{
		Id: "negated",
		Description: "This compares a boolean to a boolean literal to ask whether it is false, " +
			"which the value already answers. Negate the expression instead.",
	}
}

func messageNoUnnecessaryBooleanLiteralCompareNullableToTrueDirect() rule.Message {
	return rule.Message{
		Id: "comparingNullableToTrueDirect",
		Description: "This compares a nullable boolean to `true`, which is a roundabout way of " +
			"asking whether it is exactly true. Use the expression directly, or supply a default " +
			"with the nullish coalescing operator so the nullish case is stated rather than implied.",
	}
}

func messageNoUnnecessaryBooleanLiteralCompareNullableToTrueNegated() rule.Message {
	return rule.Message{
		Id: "comparingNullableToTrueNegated",
		Description: "This compares a nullable boolean to `true` to ask whether it is anything " +
			"else, which lumps `false` together with null and undefined. Negate the expression " +
			"instead, so the nullish case is stated rather than implied.",
	}
}

func messageNoUnnecessaryBooleanLiteralCompareNullableToFalse() rule.Message {
	return rule.Message{
		Id: "comparingNullableToFalse",
		Description: "This compares a nullable boolean to `false`, which treats null and undefined " +
			"the same as `true`. Supply a default with the nullish coalescing operator so the " +
			"nullish case is a decision rather than a side effect of the comparison.",
	}
}

func messageNoUnnecessaryBooleanLiteralCompareNoStrictNullCheck() rule.Message {
	return rule.Message{
		Id: "noStrictNullCheck",
		Description: "This rule needs the `strictNullChecks` compiler option to tell a boolean " +
			"from a nullable one. Without it every nullable boolean looks like a plain boolean, so " +
			"the rule would propose repairs that silently change what the code does on null.",
	}
}

// NoUnnecessaryBooleanLiteralCompare flags a comparison between a boolean and a boolean literal.
//
//	valid:   declare const b: boolean; if (b) {}
//	valid:   declare const s: string; if (s === 'true') {}
//	valid:   declare const n: boolean | undefined; if (n === true) {}   allowed by default
//	invalid: declare const b: boolean; if (b === true) {}    becomes if (b) {}
//	invalid: declare const b: boolean; if (b !== true) {}    becomes if (!b) {}
//
// Ported from `@typescript-eslint/no-unnecessary-boolean-literal-compare`, reading the clone at
// `packages/eslint-plugin/src/rules/no-unnecessary-boolean-literal-compare.ts` and measuring every
// verdict and every repair against the installed 8.67.0 build driven over a real TypeScript program,
// since the rule is type-aware and the plain Linter path cannot reach it.
//
// # The repair copies the compared expression rather than re-rendering it
//
// This is the shape that lost type information twice in this project, and it is safe here for a
// structural reason worth stating: the replacement is built from the compared expression's OWN
// SOURCE TEXT plus punctuation, so nothing inside it is re-rendered and nothing inside it can be
// lost. The span being replaced holds only the comparison operator and the boolean literal, neither
// of which carries a type.
//
// Measured on the installed build over the shapes a rewrite could damage, and every one survives:
//
//	f<string>() === true        becomes f<string>()          a generic call
//	(x as boolean) === true     becomes x as boolean          a type assertion
//	(b satisfies boolean) === true  becomes b satisfies boolean
//	b! === true                 becomes b!                    a non-null assertion
//	o[`b`] === true             becomes o[`b`]                a template element access
//	/* c */ b === true          becomes /* c */ b             a comment before the expression
//
// # Parentheses, and the one place our parser forces a step upstream does not have
//
// TSESTree excludes the parentheses from an expression's range, so `(a || c) === true` hands
// upstream a `LogicalExpression` whose text is `a || c`, and upstream then re-adds parentheses when
// the context needs them. Our parser gives a real `KindParenthesizedExpression`, so the comparison's
// left side is the wrapper and its text already carries the parentheses.
//
// Reading it without unwrapping would be wrong twice over: the text would carry parentheses upstream
// does not write, and the precedence test would answer about the wrapper rather than about what is
// inside it. `type_checking.IsStrongPrecedenceNode` on the shelf treats a parenthesized expression as
// STRONG, which is correct for a caller handed the wrapper and wrong here, because upstream's own
// list has no such arm and cannot: its parser deleted the node. So the unwrap happens first and the
// shelf helper is then asked about the inner node, where it agrees with upstream's list.
//
// # The nullish arms produce different code, not just different messages
//
// A plain boolean rewrites to the expression or its negation. A NULLABLE one has to say what the
// nullish case means, and upstream picks the default that preserves the original truth table:
//
//	b === true    in a value position   becomes b ?? false
//	b === true    in a conditional test becomes b            the test coerces anyway
//	b === false                          becomes !(b ?? true)
//
// Both nullish arms are ALLOWED by default, so the rule is silent on them unless configured. That is
// the one place a Go zero value would invert the rule rather than silence it, which is why the
// decoder is hand-rolled with pointer fields.
//
// # The strictNullChecks arm is unreachable through this project's test harness
//
// Upstream reports once per file, at line zero, when `strictNullChecks` is off and the escape option
// is not set. Without it every nullable boolean looks like a plain boolean and the rule would
// propose repairs that change behavior on null.
//
// It is reproduced, and it cannot be fixtured through `rule_testing`: the harness writes its own
// tsconfig AFTER the setup hook runs, so a fixture cannot turn the option off. Probed directly,
// with a hook writing `strict:false` and `strictNullChecks:false` into the directory: the resolved
// option is still true, in both the default path and the hook path. The predicate is therefore
// pinned by a unit test on the option resolution rather than by a rule fixture, and that is stated
// rather than left as a gap.
var NoUnnecessaryBooleanLiteralCompare = rule.Rule{
	Name:             "@typescript-eslint/no-unnecessary-boolean-literal-compare",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// The compiler options.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The zero value is not the default and this is the normal path, not a defensive branch.
		// Both nullish allowances default to TRUE, so a zero-valued struct would turn them off and
		// the rule would report cases upstream is silent on. A rule configured as a bare `"error"`
		// is handed nil options, because `DecodeOptionsInto` errors on empty input and the config
		// layer turns that into nil.
		parsed, decoded := rule.OptionsAs[NoUnnecessaryBooleanLiteralCompareOptions](options)
		if !decoded {
			parsed = DefaultNoUnnecessaryBooleanLiteralCompareOptions()
		}

		reportedStrictNullChecks := false

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				if reportedStrictNullChecks {
					return
				}
				if noUnnecessaryBooleanLiteralCompareStrictNullChecksEnabled(ctx) {
					return
				}
				if parsed.AllowRuleToRunWithoutStrictNullChecks {
					return
				}
				reportedStrictNullChecks = true
				// Upstream reports at line zero, column zero, which is a location rather than a
				// node. The file's first token is the nearest thing this tree has to that and it
				// puts the finding where a reader will look.
				ctx.ReportRange(ctx.SourceFile.AsNode().Loc.WithEnd(ctx.SourceFile.AsNode().Loc.Pos()),
					messageNoUnnecessaryBooleanLiteralCompareNoStrictNullCheck())
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				reportNoUnnecessaryBooleanLiteralCompare(ctx, node, parsed)
			},
		}
	},
}

// noUnnecessaryBooleanLiteralCompareComparison is upstream's `BooleanComparison`, plus the type
// verdict its `WithTypeInformation` extension adds.
type noUnnecessaryBooleanLiteralCompareComparison struct {
	// expression is the side being compared, parentheses already unwrapped.
	expression *ast.Node

	// literalIsTrue records which boolean literal the other side held.
	literalIsTrue bool

	// negated is true for `!=` and `!==`.
	negated bool

	// expressionIsNullableBoolean separates the two families of message and repair.
	expressionIsNullableBoolean bool
}

// reportNoUnnecessaryBooleanLiteralCompare is upstream's `BinaryExpression` visitor.
func reportNoUnnecessaryBooleanLiteralCompare(
	ctx rule.Context,
	node *ast.Node,
	options NoUnnecessaryBooleanLiteralCompareOptions,
) {
	comparison, found := noUnnecessaryBooleanLiteralCompareDeconstruct(ctx, node)
	if !found {
		return
	}

	if comparison.expressionIsNullableBoolean {
		if comparison.literalIsTrue && options.AllowComparingNullableBooleansToTrue {
			return
		}
		if !comparison.literalIsTrue && options.AllowComparingNullableBooleansToFalse {
			return
		}
	}

	message := noUnnecessaryBooleanLiteralCompareMessageFor(comparison)

	replacement, mutated, buildable := noUnnecessaryBooleanLiteralCompareReplacement(ctx, node, comparison)
	if !buildable {
		ctx.ReportNode(node, message)
		return
	}

	ctx.ReportNodeWithFixes(node, message,
		rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, mutated), replacement))
}

// noUnnecessaryBooleanLiteralCompareMessageFor is upstream's nested conditional over the ids.
func noUnnecessaryBooleanLiteralCompareMessageFor(
	comparison noUnnecessaryBooleanLiteralCompareComparison,
) rule.Message {
	if comparison.expressionIsNullableBoolean {
		if !comparison.literalIsTrue {
			return messageNoUnnecessaryBooleanLiteralCompareNullableToFalse()
		}
		if comparison.negated {
			return messageNoUnnecessaryBooleanLiteralCompareNullableToTrueNegated()
		}
		return messageNoUnnecessaryBooleanLiteralCompareNullableToTrueDirect()
	}
	if comparison.negated {
		return messageNoUnnecessaryBooleanLiteralCompareNegated()
	}
	return messageNoUnnecessaryBooleanLiteralCompareDirect()
}

// noUnnecessaryBooleanLiteralCompareDeconstruct is upstream's `getBooleanComparison`, which folds
// `deconstructComparison` and the type test together.
func noUnnecessaryBooleanLiteralCompareDeconstruct(
	ctx rule.Context,
	node *ast.Node,
) (noUnnecessaryBooleanLiteralCompareComparison, bool) {
	var empty noUnnecessaryBooleanLiteralCompareComparison

	binary := node.AsBinaryExpression()
	if binary.OperatorToken == nil {
		return empty, false
	}

	positive, isEquality := noUnnecessaryBooleanLiteralCompareEqualsKind(binary.OperatorToken.Kind)
	if !isEquality {
		return empty, false
	}

	// Upstream tries the right side as the literal first, then the left, and returns on the first
	// match. The order decides `true === true`, where both sides are literals: the RIGHT one is
	// taken as the literal and the LEFT becomes the compared expression.
	sides := [2][2]*ast.Node{
		{binary.Right, binary.Left},
		{binary.Left, binary.Right},
	}

	for _, side := range sides {
		against, expression := side[0], side[1]
		literalIsTrue, isBooleanLiteral := noUnnecessaryBooleanLiteralCompareBooleanLiteral(
			noUnnecessaryBooleanLiteralCompareSkipParentheses(against))
		if !isBooleanLiteral {
			continue
		}

		// Our parser keeps a parenthesized expression where TSESTree folds it away, so the
		// comparison's side is the wrapper and upstream is handed what is inside. Unwrapping here
		// makes both the text and the precedence test ask about the same node upstream asks about.
		expression = noUnnecessaryBooleanLiteralCompareSkipParentheses(expression)
		if expression == nil {
			return empty, false
		}

		comparison := noUnnecessaryBooleanLiteralCompareComparison{
			expression:    expression,
			literalIsTrue: literalIsTrue,
			negated:       !positive,
		}

		expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
		if expressionType == nil {
			return empty, false
		}

		// Upstream's `getConstraintInfo`: a type parameter is judged by its constraint, and one
		// with no constraint is declined rather than treated as its own type.
		//
		// The nil check is upstream's and it is INERT here, which is recorded rather than left for
		// the next reader to rediscover. Two mutants establish it: deleting the check survives, and
		// so does replacing it with a fallback to the type parameter itself. Both shelf predicates
		// below guard nil and answer false, and a bare type parameter carries neither the Boolean
		// flag nor the Union flag, so all three versions reach the same `return empty, false` by
		// different routes. It is kept because it is upstream's and because it states the intent,
		// not because any input can see it.
		constraint := expressionType
		if type_checking.IsTypeFlagSet(expressionType, shimchecker.TypeFlagsTypeParameter) {
			constraint = shimchecker.Checker_getBaseConstraintOfType(ctx.TypeChecker, expressionType)
			if constraint == nil {
				return empty, false
			}
		}

		if noUnnecessaryBooleanLiteralCompareIsBoolean(constraint) {
			comparison.expressionIsNullableBoolean = false
			return comparison, true
		}
		if noUnnecessaryBooleanLiteralCompareIsNullableBoolean(constraint) {
			comparison.expressionIsNullableBoolean = true
			return comparison, true
		}
		return empty, false
	}

	return empty, false
}

// noUnnecessaryBooleanLiteralCompareEqualsKind is upstream's `getEqualsKind`.
//
// Only the four equality operators reach this rule. `isStrict` is computed upstream and never read,
// so it is not carried here; the boolean returned is upstream's `isPositive`.
func noUnnecessaryBooleanLiteralCompareEqualsKind(operator ast.Kind) (positive bool, isEquality bool) {
	switch operator {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken:
		return true, true
	case ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
		return false, true
	}
	return false, false
}

// noUnnecessaryBooleanLiteralCompareBooleanLiteral answers whether a node is `true` or `false`.
//
// TSESTree folds both into a `Literal` whose value is a JavaScript boolean; typescript-go gives each
// its own keyword kind, so the two arms are written out. Anything else, including a `Boolean(...)`
// call or a `1`, is not a boolean literal to this rule.
func noUnnecessaryBooleanLiteralCompareBooleanLiteral(node *ast.Node) (isTrue bool, ok bool) {
	if node == nil {
		return false, false
	}
	switch node.Kind {
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword:
		return false, true
	}
	return false, false
}

// noUnnecessaryBooleanLiteralCompareIsBoolean is upstream's `isBooleanType`.
//
// `Boolean` and `BooleanLiteral` together, so `true`, `false` and `boolean` all answer yes. Note
// this is a FLAG test on the type as a whole rather than on its constituents, which is why a
// `boolean | string` union answers no here and falls through to the nullable test.
func noUnnecessaryBooleanLiteralCompareIsBoolean(expressionType *shimchecker.Type) bool {
	return type_checking.IsTypeFlagSet(expressionType,
		shimchecker.TypeFlagsBoolean|shimchecker.TypeFlagsBooleanLiteral)
}

// noUnnecessaryBooleanLiteralCompareIsNullableBoolean is upstream's `isNullableBoolean`.
//
// A union that holds at least one nullish constituent, at least one non-nullish constituent, and
// whose non-nullish constituents are all boolean. The three conditions are upstream's and each
// declines a different shape: `null | undefined` alone has no non-nullish part, a plain `boolean`
// is not a union at all, and `boolean | string | undefined` has a non-boolean survivor.
func noUnnecessaryBooleanLiteralCompareIsNullableBoolean(expressionType *shimchecker.Type) bool {
	if !type_checking.IsUnionType(expressionType) {
		return false
	}

	parts := type_checking.UnionTypeParts(expressionType)
	nonNullish := make([]*shimchecker.Type, 0, len(parts))
	for _, part := range parts {
		if type_checking.IsTypeFlagSet(part,
			shimchecker.TypeFlagsUndefined|shimchecker.TypeFlagsNull) {
			continue
		}
		nonNullish = append(nonNullish, part)
	}

	if len(nonNullish) == 0 {
		return false
	}
	// This one is SUBSUMED by the loop below and the verdict is recorded rather than the guard
	// deleted. A union with no nullish constituent reaches the loop with every constituent intact,
	// and any union of only booleans has already been answered by the flag test on the whole type
	// before this function is called, so whatever survives here is a non-boolean the loop declines.
	// A mutant removing this line survives, correctly. It is kept because it is upstream's and
	// because it says "not nullable" in one place rather than leaving that conclusion to be
	// reconstructed from the loop.
	if len(nonNullish) == len(parts) {
		return false
	}
	for _, part := range nonNullish {
		if !noUnnecessaryBooleanLiteralCompareIsBoolean(part) {
			return false
		}
	}
	return true
}

// noUnnecessaryBooleanLiteralCompareReplacement is upstream's `fix` closure.
//
// The replacement is assembled from the compared expression's own source text outward, so the only
// thing this function decides is punctuation. Returning the node to replace as well as the text,
// because a comparison wrapped in `!` is replaced together with its wrapper.
func noUnnecessaryBooleanLiteralCompareReplacement(
	ctx rule.Context,
	node *ast.Node,
	comparison noUnnecessaryBooleanLiteralCompareComparison,
) (replacement string, mutated *ast.Node, buildable bool) {
	// The wrapping negation may sit outside a parenthesis, because our parser keeps a node TSESTree
	// deletes. `!(b !== false)` hands upstream a negation whose argument IS the comparison, and
	// hands this rule a negation whose argument is a parenthesized expression. Walking out through
	// the parentheses finds the same negation upstream found, and the node replaced then has to be
	// that negation rather than the comparison, or the parentheses would be left stranded.
	outermost := node
	for outermost.Parent != nil && outermost.Parent.Kind == ast.KindParenthesizedExpression {
		outermost = outermost.Parent
	}

	wrappedInNegation := noUnnecessaryBooleanLiteralCompareIsUnaryNegation(outermost.Parent)
	mutated = node
	if wrappedInNegation {
		mutated = outermost.Parent
	}

	// Upstream's `booleanXor` over three flags: whether the whole expression's truth table is
	// inverted, ignoring the nullish cases. Written as a parity count because that is what the fold
	// computes and it reads with fewer moving parts.
	inversions := 0
	if wrappedInNegation {
		inversions++
	}
	if comparison.negated {
		inversions++
	}
	if !comparison.literalIsTrue {
		inversions++
	}
	overallNegated := inversions%2 == 1

	text, readable := noUnnecessaryBooleanLiteralCompareNodeText(ctx, comparison.expression)
	if !readable {
		return "", nil, false
	}

	mayNeedParentheses := !type_checking.IsStrongPrecedenceNode(comparison.expression)

	// A nullable expression that is not being negated overall becomes the expression with an
	// explicit nullish default, EXCEPT in a conditional test, where the test coerces the nullish
	// value the same way and the default would be noise.
	if !overallNegated && comparison.expressionIsNullableBoolean &&
		!noUnnecessaryBooleanLiteralCompareIsConditionalTest(mutated) {
		if mayNeedParentheses {
			text = "(" + text + ")"
		}
		text += " ?? false"
		mayNeedParentheses = true
	} else {
		// Comparing a nullable to `false` gives a nullish value the same truth table as `true`, so
		// the default written here is the opposite of the one above.
		if comparison.expressionIsNullableBoolean && !comparison.literalIsTrue {
			if mayNeedParentheses {
				text = "(" + text + ")"
			}
			text += " ?? true"
			mayNeedParentheses = true
		}

		if overallNegated {
			if mayNeedParentheses {
				text = "(" + text + ")"
			}
			text = "!" + text
			mayNeedParentheses = false
		}
	}

	if mayNeedParentheses && noUnnecessaryBooleanLiteralCompareIsWeakPrecedenceParent(mutated) {
		text = "(" + text + ")"
	}

	return text, mutated, true
}

// noUnnecessaryBooleanLiteralCompareSkipParentheses unwraps a parenthesized expression.
//
// Written out rather than reaching for `ast.SkipParentheses`, which dereferences its argument and is
// how this project lost 167 files to a nil panic. A loop rather than one step, because `((b))`
// nests.
//
// Every caller exists because TSESTree deletes the node this unwraps and upstream therefore never
// meets it. Three sites need it and each fails differently without it: the literal side (`(false)`
// stops being recognized as a literal at all), the parent walk (`!(b !== false)` stops being seen as
// negated), and the compared expression (its text would carry parentheses upstream does not write).
func noUnnecessaryBooleanLiteralCompareSkipParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// noUnnecessaryBooleanLiteralCompareIsUnaryNegation answers whether a node is `!x`.
func noUnnecessaryBooleanLiteralCompareIsUnaryNegation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	return node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
}

// noUnnecessaryBooleanLiteralCompareIsConditionalTest is upstream's `isConditionalTest`.
//
// It walks upward through the shapes that pass a value straight into a test position: a logical
// operator, the branches of a conditional, the last element of a sequence, and a negation. The walk
// stops at a node whose parent tests it.
func noUnnecessaryBooleanLiteralCompareIsConditionalTest(node *ast.Node) bool {
	for node != nil {
		parent := node.Parent
		if parent == nil {
			return false
		}

		switch parent.Kind {
		case ast.KindBinaryExpression:
			// Upstream distinguishes a LogicalExpression from a BinaryExpression and typescript-go
			// does not, so the operator is asked instead. Only `&&`, `||` and `??` pass a value
			// through to a surrounding test; an arithmetic or comparison operator consumes it.
			operator := parent.AsBinaryExpression().OperatorToken
			if operator == nil {
				return false
			}
			switch operator.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				node = parent
				continue
			}
			return false

		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional.Condition == node {
				return true
			}
			if conditional.WhenTrue == node || conditional.WhenFalse == node {
				node = parent
				continue
			}
			return false

		case ast.KindPrefixUnaryExpression:
			if parent.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
				return false
			}
			node = parent
			continue

		case ast.KindParenthesizedExpression:
			// No upstream counterpart and it cannot have one: TSESTree has no parenthesized node,
			// so a parenthesized test reaches upstream's walk already unwrapped. Without this arm
			// `if ((b === true)) {}` would read as not-a-test and the repair would add a nullish
			// default the conditional does not need.
			node = parent
			continue

		case ast.KindIfStatement:
			return parent.AsIfStatement().Expression == node
		case ast.KindWhileStatement:
			return parent.AsWhileStatement().Expression == node
		case ast.KindDoStatement:
			return parent.AsDoStatement().Expression == node
		case ast.KindForStatement:
			return parent.AsForStatement().Condition == node
		}

		return false
	}
	return false
}

// noUnnecessaryBooleanLiteralCompareIsWeakPrecedenceParent is upstream's `isWeakPrecedenceParent`.
//
// It answers whether the parent could bind differently once the child is replaced, in which case the
// replacement is parenthesized. Upstream's `UpdateExpression` covers `++x` and `x++`, which are two
// kinds here.
func noUnnecessaryBooleanLiteralCompareIsWeakPrecedenceParent(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindBinaryExpression, ast.KindConditionalExpression, ast.KindAwaitExpression:
		return true

	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == node
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == node
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression == node
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression == node
	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag == node
	}
	return false
}

// noUnnecessaryBooleanLiteralCompareStrictNullChecksEnabled is upstream's
// `isStrictCompilerOptionEnabled(compilerOptions, 'strictNullChecks')`.
//
// `GetStrictOptionValue` is the same resolution: an explicit value wins, and an unset one falls back
// to `strict` unless `strict` is explicitly false. Reading it through the program handle is why this
// rule declares that it reads the program.
func noUnnecessaryBooleanLiteralCompareStrictNullChecksEnabled(ctx rule.Context) bool {
	if ctx.Program == nil {
		return true
	}
	compilerOptions := ctx.Program.Options()
	if compilerOptions == nil {
		return true
	}
	return compilerOptions.GetStrictOptionValue(compilerOptions.StrictNullChecks)
}

// noUnnecessaryBooleanLiteralCompareNodeText slices a node's own source text, trivia excluded.
func noUnnecessaryBooleanLiteralCompareNodeText(ctx rule.Context, node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, node)
	sourceText := ctx.SourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(sourceText) || trimmed.Pos() > trimmed.End() {
		return "", false
	}
	return sourceText[trimmed.Pos():trimmed.End()], true
}

// NoUnnecessaryBooleanLiteralCompareOptions is the rule's configuration.
//
// Both allowances default to TRUE, which is the trap this rule's decoder exists for: a Go zero value
// turns them off and the rule reports two whole families of comparison upstream is silent on.
type NoUnnecessaryBooleanLiteralCompareOptions struct {
	AllowComparingNullableBooleansToFalse bool
	AllowComparingNullableBooleansToTrue  bool
	AllowRuleToRunWithoutStrictNullChecks bool
}

// DefaultNoUnnecessaryBooleanLiteralCompareOptions is upstream's `defaultOptions`.
func DefaultNoUnnecessaryBooleanLiteralCompareOptions() NoUnnecessaryBooleanLiteralCompareOptions {
	return NoUnnecessaryBooleanLiteralCompareOptions{
		AllowComparingNullableBooleansToFalse: true,
		AllowComparingNullableBooleansToTrue:  true,
		AllowRuleToRunWithoutStrictNullChecks: false,
	}
}

// noUnnecessaryBooleanLiteralCompareRawOptions is the wire shape.
//
// Every field is a pointer so an absent key stays distinguishable from an explicit `false`. That
// distinction is the whole content of this decoder: two of the three default to true, so a
// non-pointer field cannot tell "not configured" from "turned off" and the generic helper would
// silently invert the rule on a bare `"error"` configuration.
type noUnnecessaryBooleanLiteralCompareRawOptions struct {
	AllowComparingNullableBooleansToFalse *bool `json:"allowComparingNullableBooleansToFalse"`
	AllowComparingNullableBooleansToTrue  *bool `json:"allowComparingNullableBooleansToTrue"`
	AllowRuleToRunWithoutStrictNullChecks *bool `json:"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing"`
}

// DecodeNoUnnecessaryBooleanLiteralCompareOptions maps the wire keys onto the rule's settings.
//
// Hand-written rather than `rule.DecodeOptionsInto` because two of the three defaults are true.
// The generic helper on a non-pointer struct yields false for every absent key, which would turn
// both allowances off and make the rule report the two nullish families upstream allows by default.
// Every fixture built from a struct rather than routed through this decoder would pass anyway,
// which is the failure this project has shipped before.
func DecodeNoUnnecessaryBooleanLiteralCompareOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noUnnecessaryBooleanLiteralCompareRawOptions]()(raw)
	if err != nil {
		return DefaultNoUnnecessaryBooleanLiteralCompareOptions(), err
	}

	wire, _ := decoded.(noUnnecessaryBooleanLiteralCompareRawOptions)
	options := DefaultNoUnnecessaryBooleanLiteralCompareOptions()

	if wire.AllowComparingNullableBooleansToFalse != nil {
		options.AllowComparingNullableBooleansToFalse = *wire.AllowComparingNullableBooleansToFalse
	}
	if wire.AllowComparingNullableBooleansToTrue != nil {
		options.AllowComparingNullableBooleansToTrue = *wire.AllowComparingNullableBooleansToTrue
	}
	if wire.AllowRuleToRunWithoutStrictNullChecks != nil {
		options.AllowRuleToRunWithoutStrictNullChecks = *wire.AllowRuleToRunWithoutStrictNullChecks
	}

	return options, nil
}
