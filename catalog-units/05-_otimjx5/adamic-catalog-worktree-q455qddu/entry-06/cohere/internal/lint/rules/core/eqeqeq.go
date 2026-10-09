package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// EqeqeqMode selects how strictly the rule asks for the identity operators.
type EqeqeqMode string

const (
	// EqeqeqAlways reports every `==` and `!=`. The default.
	EqeqeqAlways EqeqeqMode = "Always"

	// EqeqeqSmart exempts the comparisons where coercion cannot surprise anyone.
	//
	// Those are: either operand is a `typeof`, both operands are literals of the same type, or
	// either operand is `null`. The reasoning is that a `typeof` result is always a string, two
	// same-typed literals cannot coerce differently, and `== null` is the idiomatic nullish test.
	EqeqeqSmart EqeqeqMode = "Smart"
)

// EqeqeqNullPolicy selects what the rule does about a comparison against the `null` literal.
type EqeqeqNullPolicy string

const (
	// EqeqeqNullAlways reports `x == null` and asks for `===`. The default under Always.
	EqeqeqNullAlways EqeqeqNullPolicy = "Always"

	// EqeqeqNullNever inverts the rule for null, reporting `x === null` and asking for `==`.
	//
	// This is the only setting under which the rule reports an identity operator, and it exists for
	// codebases that treat `== null` as the deliberate way to spell "null or undefined".
	EqeqeqNullNever EqeqeqNullPolicy = "Never"

	// EqeqeqNullIgnore leaves every comparison against `null` alone, either way round.
	EqeqeqNullIgnore EqeqeqNullPolicy = "Ignore"
)

// EqeqeqOptions configures the rule.
//
// Upstream's option surface is a positional list whose legal second element depends on the first:
// `["always", {"null": ...}]` is valid and `["smart", {"null": ...}]` is not, which its schema
// expresses as a two-branch `anyOf`. `DecodeEqeqeqOptions` reads that list and enforces the same
// dependency, and this struct is what it decodes to.
//
// Upstream also accepts `"allow-null"` as a deprecated first element. Its own code reads it as
// `nullOption = config === "always" ? ... : "ignore"` with no smart exemption, which is Always plus
// a null policy of Ignore, so that is what it decodes to.
type EqeqeqOptions struct {
	// Mode is how strictly to ask. Absent means Always, which is upstream's default.
	Mode EqeqeqMode

	// Null is what to do about a comparison against the `null` literal.
	//
	// Absent means Always under the Always mode, matching upstream's `options.null || "always"`.
	// Under Smart it is forced to Ignore, because Smart already exempts every null comparison and
	// upstream hardcodes `nullOption = "ignore"` for it.
	Null EqeqeqNullPolicy
}

// eqeqeqSettings is the decoded form with the mode-dependency already resolved.
type eqeqeqSettings struct {
	mode EqeqeqMode
	null EqeqeqNullPolicy
}

// resolve applies upstream's defaults, including the one that depends on the mode.
func (options EqeqeqOptions) resolve() eqeqeqSettings {
	settings := eqeqeqSettings{mode: options.Mode, null: options.Null}
	if settings.mode == "" {
		settings.mode = EqeqeqAlways
	}
	// `nullOption = config === "always" ? options.null || "always" : "ignore"`. The null policy is
	// only consulted under Always; Smart exempts every null comparison on its own, so a null policy
	// written beside Smart is discarded rather than honoured. Upstream's schema makes that
	// combination unwritable; ours cannot, so the discard happens here.
	if settings.mode != EqeqeqAlways {
		settings.null = EqeqeqNullIgnore
	} else if settings.null == "" {
		settings.null = EqeqeqNullAlways
	}
	return settings
}

// DecodeEqeqeqOptions reads upstream's option list off the config: `["always", {"null": ...}]`,
// `["smart"]`, `["allow-null"]`, or nothing.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so an unrecognized mode or policy fails loudly.
// Every arm of this rule is selected by string equality, so a typo would silently pick a fourth
// behaviour of reporting nothing. A second element beside anything but "always" is refused, which is
// upstream's schema: its `anyOf` gives `smart` and `allow-null` a one-element list.
func DecodeEqeqeqOptions(list []byte) (any, error) {
	options := EqeqeqOptions{}
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return options, err
	}

	var mode string
	if err := json.Unmarshal(elements[0], &mode); err != nil {
		return options, fmt.Errorf("eqeqeq element 1 takes a string, got %s", elements[0])
	}
	switch mode {
	case "always":
		options.Mode = EqeqeqAlways
	case "smart":
		options.Mode = EqeqeqSmart
	case "allow-null":
		options.Mode = EqeqeqAlways
		options.Null = EqeqeqNullIgnore
	default:
		return options, fmt.Errorf(
			"eqeqeq: unknown mode %q, wanted one of always, smart, allow-null", mode)
	}
	if len(elements) < 2 {
		return options, nil
	}
	if mode != "always" {
		return options, fmt.Errorf(
			"eqeqeq: %q takes no second element, so %s would never be read; only \"always\" "+
				"reads a null policy", mode, elements[1])
	}

	var second struct {
		Null *string `json:"null"`
	}
	if err := rule.UnmarshalOptions(elements[1], &second); err != nil {
		return options, fmt.Errorf("eqeqeq element 2: %w", err)
	}
	if second.Null == nil {
		return options, nil
	}
	switch *second.Null {
	case "always":
		options.Null = EqeqeqNullAlways
	case "never":
		options.Null = EqeqeqNullNever
	case "ignore":
		options.Null = EqeqeqNullIgnore
	default:
		return options, fmt.Errorf(
			"eqeqeq: unknown null policy %q, wanted one of always, never, ignore", *second.Null)
	}
	return options, nil
}

// Eqeqeq flags a loose equality comparison and asks for the identity operator.
//
//	valid:   a === b
//	valid:   typeof a == 'number'          // under Smart
//	valid:   a == null                     // under Smart, or null: Ignore
//	invalid: a == b
//	invalid: a != b
//	invalid: a === null                    // under null: Never
//
// # Why this matters more than it looks
//
// `==` applies the abstract equality algorithm, which coerces across types before comparing, so
// `0 == '0'`, `0 == []` and an empty string against `false` are all true while `'0' == []` is false. The result is a
// relation that is not transitive, which is the property every reader unconsciously assumes a
// comparison has. `===` compares without coercion and is transitive.
//
// # The repair is a fix sometimes and a suggestion otherwise, and that split IS the rule
//
// Changing `==` to `===` changes what the code DOES wherever the operands can differ in type, so
// applying it unattended could turn a passing comparison into a failing one. Upstream draws the line
// where it can prove coercion is impossible:
//
//	either operand is a `typeof`     the result is always a string, so both sides are strings
//	both operands are literals       of the same `typeof` class, so no coercion can occur
//
// Those get a real Fix, applied unattended. Everything else gets a Suggestion, which a human
// chooses. Twenty of upstream's cases carry a fix and twenty-eight carry a suggestion, and no case
// carries both. Shipping the suggestion arm as a fix would let the edit engine change program
// behaviour on `a == b`, which is the single worst thing a linter can do.
//
// # `null` is its own axis
//
// `x == null` is the idiomatic test for "null or undefined", so it is broken out as a separate
// option rather than folded into the general judgment. Under `Null: Never` the rule INVERTS for
// null, reporting `x === null` and asking for `==` -- the only configuration in which this rule
// reports an identity operator at all.
//
// # Parentheses
//
// All three predicates see through them, and that is fidelity rather than a widening: espree gives a
// parenthesized expression no node, so upstream's direct field reads already look past them.
// Measured against the installed build -- `(typeof a) == 'number'` and `(1) == (2)` both take the
// FIX arm, and `x == (null)` takes the null arm. Reading the operands directly would put all three
// on the wrong arm while every unparenthesized case stayed green.
var Eqeqeq = rule.Rule{
	Name: "eqeqeq",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := EqeqeqOptions{}.resolve()
		if decoded, configured := rule.OptionsAs[EqeqeqOptions](options); configured {
			settings = decoded.resolve()
		}
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkEqeqeq(ctx, node, settings)
			},
		}
	},
}

// checkEqeqeq judges one binary expression.
func checkEqeqeq(ctx rule.Context, node *ast.Node, settings eqeqeqSettings) {
	binary := node.AsBinaryExpression()
	// The operator token is optional on the node type even though the parser always supplies one,
	// and the predicates below dereference their operands, so all three nil tests are here rather
	// than trusted away.
	if binary.OperatorToken == nil || binary.Left == nil || binary.Right == nil {
		return
	}

	isNull := isNullLiteralOperand(binary.Left) || isNullLiteralOperand(binary.Right)

	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
		// A loose operator, which is the ordinary case.
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		// An identity operator is reported ONLY under the inverted null policy, and only against
		// null. Everywhere else it is the repair this rule asks for.
		if settings.null == EqeqeqNullNever && isNull {
			reportEqeqeq(ctx, node, binary, eqeqeqLooseSpellingOf(binary.OperatorToken.Kind))
		}
		return
	default:
		// Every other binary operator, from `<` to `+` to `instanceof`.
		return
	}

	// Smart exempts the three shapes where coercion cannot surprise anyone.
	if settings.mode == EqeqeqSmart &&
		(eqeqeqHasTypeOfOperand(binary) || eqeqeqOperandsAreSameTypedLiterals(binary) || isNull) {
		return
	}

	// Under any policy but Always, a null comparison is left alone. Reached only under the Always
	// mode, because `resolve` forces Ignore for Smart.
	if settings.null != EqeqeqNullAlways && isNull {
		return
	}

	reportEqeqeq(ctx, node, binary, eqeqeqStrictSpellingOf(binary.OperatorToken.Kind))
}

// reportEqeqeq reports one comparison, with a fix where coercion is provably impossible and a
// suggestion everywhere else.
//
// The finding is anchored on the OPERATOR TOKEN rather than on the whole expression, which is
// upstream's `loc: operatorToken.loc`. That matters for a reader: a finding spanning `a == b` puts
// the caret on `a`, and the thing being complained about is the two characters in the middle.
func reportEqeqeq(ctx rule.Context, node *ast.Node, binary *ast.BinaryExpression, expected string) {
	// The operator's spelling comes from the SOURCE rather than from `Text()`, which panics on a
	// token node: `Unhandled case in Node.Text: *ast.Token`. That is the same hazard the shelf's
	// `property.Name` documents for a computed key, arriving through a different kind, and it takes
	// the whole file down rather than returning empty.
	operatorRange := rule.TokenRange(ctx.SourceFile, binary.OperatorToken)
	actual := ctx.SourceFile.Text()[operatorRange.Pos():operatorRange.End()]

	message := rule.Message{
		Id: "unexpected",
		Description: fmt.Sprintf(
			"This compares with `%s`, which coerces its operands to a common type before "+
				"comparing. That makes the comparison non-transitive, which is the one property "+
				"every reader assumes a comparison has: `0 == '0'` and `'0' == []` disagree "+
				"while `0 == []` is true. Write `%s`, which compares without coercing.",
			actual, expected),
	}

	// The fix arm: both operands are provably the same type, so swapping the operator cannot change
	// what the comparison answers. Everything else is offered rather than applied, because there
	// the swap CHANGES BEHAVIOUR and a human has to agree to it.
	if eqeqeqHasTypeOfOperand(binary) || eqeqeqOperandsAreSameTypedLiterals(binary) {
		ctx.ReportRangeWithFixes(operatorRange, message,
			rule.ReplaceRange(operatorRange, expected))
		return
	}

	ctx.ReportRangeWithSuggestions(operatorRange, message, rule.Suggestion{
		Message: rule.Message{
			Id: "replaceOperator",
			Description: fmt.Sprintf("Use `%s` instead of `%s`. This changes what the "+
				"comparison answers wherever the operands can differ in type, so it is offered "+
				"rather than applied.", expected, actual),
		},
		Fixes: []rule.Fix{rule.ReplaceRange(operatorRange, expected)},
	})
}

// eqeqeqStrictSpellingOf returns the identity operator matching a loose one.
func eqeqeqStrictSpellingOf(kind ast.Kind) string {
	if kind == ast.KindExclamationEqualsToken {
		return "!=="
	}
	return "==="
}

// eqeqeqLooseSpellingOf returns the loose operator matching an identity one.
//
// Reached only under the inverted null policy, which is the one configuration asking for `==`.
func eqeqeqLooseSpellingOf(kind ast.Kind) string {
	if kind == ast.KindExclamationEqualsEqualsToken {
		return "!="
	}
	return "=="
}

// eqeqeqHasTypeOfOperand answers whether either operand is a `typeof` expression.
//
// A `typeof` result is always one of a fixed set of strings, so comparing it against anything makes
// both sides strings and coercion cannot occur. That is what puts the comparison on the fix arm.
func eqeqeqHasTypeOfOperand(binary *ast.BinaryExpression) bool {
	return eqeqeqIsTypeOfExpression(binary.Left) || eqeqeqIsTypeOfExpression(binary.Right)
}

// eqeqeqIsTypeOfExpression answers whether an operand is `typeof x`.
func eqeqeqIsTypeOfExpression(operand *ast.Node) bool {
	if operand == nil {
		return false
	}
	unwrapped := ast.SkipParentheses(operand)
	return unwrapped.Kind == ast.KindTypeOfExpression
}

// eqeqeqOperandsAreSameTypedLiterals answers whether both operands are literals of one `typeof`
// class.
//
// Upstream's `areLiteralsAndSameType`, which compares `typeof node.value` for each side. Two string
// literals are the same type; a number and a string are not; a number and a bigint are NOT, because
// `typeof 1n` is `"bigint"` -- measured, `1 == 1n` takes the suggestion arm.
func eqeqeqOperandsAreSameTypedLiterals(binary *ast.BinaryExpression) bool {
	leftType := eqeqeqLiteralTypeOf(binary.Left)
	return leftType != "" && leftType == eqeqeqLiteralTypeOf(binary.Right)
}

// eqeqeqLiteralTypeOf returns the `typeof` class of a literal operand, or "" when it is not one.
//
// Upstream reads `typeof node.value` on an ESTree `Literal`, which is one node kind covering every
// literal form; our parser splits them by kind, so the mapping is written out. A no-substitution
// template counts as a string, matching upstream's `isStaticTemplateLiteral` branch, while a
// template WITH substitutions does not, because its value is not known until it runs.
//
// A regular expression literal answers "object", which is what `typeof /a/` evaluates to, and is why
// `/a/ == /a/` takes the fix arm -- measured against the installed build rather than reasoned about,
// because two regular expressions being the same `typeof` class while never being equal is exactly
// the kind of thing a port would tidy away.
func eqeqeqLiteralTypeOf(operand *ast.Node) string {
	if operand == nil {
		return ""
	}
	switch unwrapped := ast.SkipParentheses(operand); unwrapped.Kind {
	case ast.KindStringLiteral:
		return "string"
	case ast.KindNoSubstitutionTemplateLiteral:
		return "string"
	case ast.KindNumericLiteral:
		return "number"
	case ast.KindBigIntLiteral:
		return "bigint"
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return "boolean"
	case ast.KindNullKeyword:
		// `typeof null` is `"object"`, which is what makes `null == null` take the fix arm.
		return "object"
	case ast.KindRegularExpressionLiteral:
		return "object"
	}
	return ""
}
