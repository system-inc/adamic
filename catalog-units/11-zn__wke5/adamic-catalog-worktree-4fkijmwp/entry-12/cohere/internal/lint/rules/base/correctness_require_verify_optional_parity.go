package base

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// verifyIsOptionalDecorator is the sentinel that tells the validation engine to skip every other
// rule on a property when the value is null or undefined.
//
// Base's decorator name, matched literally. See the package doc for why it keeps that spelling.
const verifyIsOptionalDecorator = "VerifyIsOptional"

// verifyDecoratorPrefix is what marks a decorator as belonging to the validation engine.
const verifyDecoratorPrefix = "Verify"

// The rule's messages, whose wording lives in `policy/messages/correctness-require-verify-optional-parity.json`.
var (
	correctnessRequireVerifyOptionalParityOptionalButTypeNotText      = policy.MessageOf("base/correctness-require-verify-optional-parity", "optionalButTypeNot")
	correctnessRequireVerifyOptionalParityTypeOptionalButNoVerifyText = policy.MessageOf("base/correctness-require-verify-optional-parity", "typeOptionalButNoVerify")
)

// CorrectnessRequireVerifyOptionalParity requires the optional sentinel and the declared type to agree.
//
//	valid:   @VerifyIsString() name: string
//	valid:   @VerifyIsString() @VerifyIsOptional() name?: string
//	valid:   name?: string                                 (no Verify decorators, not validated)
//	invalid: @VerifyIsOptional() name: string              (validator tolerant, type is not)
//	invalid: @VerifyIsString() name?: string               (type tolerant, validator is not)
//
// # Two judgments, and the second needs a second decorator to fire
//
// `@VerifyIsOptional()` present with a non-nullable type is always a finding: the validator would
// accept a missing value the compiler forbids.
//
// The reverse only fires when the property carries at least one OTHER `Verify*` decorator. A
// nullable property with `@VerifyIsOptional()` alone is not a problem, and a nullable property with
// no validation at all is outside the rule's scope entirely. That "at least one other" test is what
// separates the two, and it is easy to drop into "has any Verify decorator", which would report
// every nullable property that carries only the sentinel.
//
// # A property with no Verify decorator is not validated
//
// The engine does not look at it, so neither does this rule. That is the first guard and it is what
// keeps the rule from reporting every optional field in the codebase.
//
// # The nullability question lives on the shelf
//
// `decorators.IsNullableType` is the shared predicate across the four base parity rules, and its
// mask is deliberately wider than null and undefined: `any`, `unknown` and `void` count, because
// they erase the distinction rather than answering it. A type of `any` beside a validator that
// forbids missing values is not a claim anyone checked.
//
// Measured before this rule was written: a union carrying `undefined` does NOT answer
// `IsTypeFlagSet` directly. `string | undefined` reports the union's own flags, and only its member
// carries `Undefined`. So the recursion inside `IsNullableType` is load-bearing rather than
// defensive, and a port testing the top-level flags alone goes silent on every optional property.
//
// # No fix
//
// The original ships none and none is possible: the repair is either to widen the type or to add
// the sentinel, and which one is right depends on what the field means.
var CorrectnessRequireVerifyOptionalParity = rule.Rule{
	Name: "base/correctness-require-verify-optional-parity",

	// The judgment is about the resolved TYPE of the property, not about its annotation. A property
	// written `name?: string` and one written `name: string | undefined` are the same question, and
	// only the checker collapses them.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkVerifyOptionalParity(ctx, node, node.Name())
			},
			// A constructor parameter property, `constructor(@VerifyIsString() public name: string)`.
			// The original hooks `TSParameterProperty`, which our parser does not have: it gives a
			// parameter the accessibility modifier directly, so the parameter kind carries both the
			// plain and the property form and the accessibility test tells them apart.
			ast.KindParameter: func(node *ast.Node) {
				if !isBaseParameterProperty(node) {
					return
				}
				checkVerifyOptionalParity(ctx, node, node.Name())
			},
		}
	},
}

// checkVerifyOptionalParity judges one property or parameter property.
func checkVerifyOptionalParity(ctx rule.Context, node *ast.Node, key *ast.Node) {
	if ctx.TypeChecker == nil || key == nil || !ast.IsIdentifier(key) {
		return
	}

	hasOptionalSentinel := false
	hasOtherVerifyRule := false
	for _, decorator := range decorators.Of(node) {
		name := decorators.CallName(decorator)
		if !isVerifyDecoratorName(name) {
			continue
		}
		if name == verifyIsOptionalDecorator {
			hasOptionalSentinel = true
			continue
		}
		hasOtherVerifyRule = true
	}

	// A property the validation engine never looks at is outside this rule's scope.
	if !hasOptionalSentinel && !hasOtherVerifyRule {
		return
	}

	keyType := ctx.TypeChecker.GetTypeAtLocation(key)
	typeIsNullable := decorators.IsNullableType(keyType)
	typeText := ctx.TypeChecker.TypeToString(keyType)

	if hasOptionalSentinel && !typeIsNullable {
		ctx.ReportNode(key, rule.Message{
			Id:          correctnessRequireVerifyOptionalParityOptionalButTypeNotText.Id,
			Description: correctnessRequireVerifyOptionalParityOptionalButTypeNotText.Render(map[string]string{"typeText": typeText}),
		})
		return
	}

	// The `hasOtherVerifyRule` test is the original's, and it is SUBSUMED here rather than
	// load-bearing. The line above already requires `!hasOptionalSentinel`, and the guard further up
	// already returned for a property with no Verify decorator at all, so by this point "not the
	// sentinel" and "some other rule" are the same fact.
	//
	// Measured rather than argued, because a single-site mutation cannot see it: dropping this test
	// SURVIVES, dropping the unvalidated-property guard SURVIVES, and dropping both together fails
	// 8 lines. Kept because it states the original's condition at the line a reader compares
	// against, and because deleting it would make the two implementations differ textually for no
	// behavioural gain.
	//
	// The first draft of this comment claimed the test stopped a sentinel-only property from
	// reporting. That was wrong -- such a property cannot reach this line at all -- and the sweep
	// is what caught it.
	if typeIsNullable && !hasOptionalSentinel && hasOtherVerifyRule {
		ctx.ReportNode(key, rule.Message{
			Id:          correctnessRequireVerifyOptionalParityTypeOptionalButNoVerifyText.Id,
			Description: correctnessRequireVerifyOptionalParityTypeOptionalButNoVerifyText.Render(map[string]string{"typeText": typeText}),
		})
	}
}

// isBaseParameterProperty says whether a constructor parameter also declares a class property.
//
// The original hooks estree's `TSParameterProperty`, a node kind our parser does not have:
// typescript-go gives the accessibility modifier to the parameter itself, so a plain parameter and
// a property-declaring one share `KindParameter`. `ModifierFlagsParameterPropertyModifier` is the
// set that makes the difference, which is `public`/`private`/`protected`/`readonly` and
// deliberately not `override` -- that one creates no property of its own.
//
// Without this test the rule would judge every constructor parameter, including ones that never
// become a validated property.
func isBaseParameterProperty(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier)
}

// isVerifyDecoratorName says whether a decorator belongs to the validation engine.
//
// Shared by both Verify parity rules rather than spelled twice in one package: they ask the
// identical question, and two spellings of it would be free to drift.
func isVerifyDecoratorName(name string) bool {
	return strings.HasPrefix(name, verifyDecoratorPrefix)
}
