package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoNullishStrippingAssertionId = "nullishStrippedByAssertion"

// correctnessNoNullishStrippingAssertionText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-nullish-stripping-assertion.json`. It gives the repairs in the
// order `no-non-null-assertion` does, since the two rules report the same claim through two doors.
var correctnessNoNullishStrippingAssertionText = policy.MessageOf("nexus/correctness-no-nullish-stripping-assertion", correctnessNoNullishStrippingAssertionId)

func correctnessNoNullishStrippingAssertionMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoNullishStrippingAssertionId,
		Description: correctnessNoNullishStrippingAssertionText.Render(nil),
	}
}

// CorrectnessNoNullishStrippingAssertion reports a type assertion that removes `undefined` or `null`
// from a value's type and also changes the rest of it, which is the non-null assertion the house
// bans, written as a cast.
//
//	invalid: const days = parseInt(options.days as string, 10);   // string | boolean | Pending | undefined
//	invalid: const element = event.target as HTMLElement;          // EventTarget | null
//	invalid: const state = cache.get(key) as 'Ready' | 'Stale';    // string | undefined
//	valid:   const days = numberFlag(options, 'days', 7);
//	valid:   const name = maybe as string;                        // the pure removal, the style rule's
//	valid:   const state = maybe as 'Ready' | undefined;
//	valid:   const cents = row.value_cents as number;              // unknown, no nullish constituent
//	valid:   if(options.start) { use(options.start as string); }   // narrowed, the undefined is gone
//
// # Where it came from
//
// The cross-language pass of the new-rules sweep (`#tevhg3f`, after rust clippy's `unwrap_used` and
// Kotlin's `!!`), built as task `#c3afxe9`. ahra bans `!` at error through
// `@typescript-eslint/no-non-null-assertion`, and `x as T` where `x` is `T | undefined` makes the same
// unchecked claim through the other door. The cited site is `modules/google/analytics/AnalyticsApi.ts:175`,
// `parseInt(options.days as string, 10)`, where the flag's type is `string | boolean |
// PendingCommandArgumentFlagValueType | undefined`: the cast silences a bare `--days` (a boolean) and
// a deferred value (an object) along with the missing one, and each reaches `parseInt` as NaN.
//
// # The partition with `@typescript-eslint/non-nullable-type-assertion-style`
//
// An assertion whose only effect is removing null and undefined (`maybe as string` where `maybe` is
// `string | undefined`) is that rule's finding, and this one declines it. The two must never report
// the same cast, so the decline is that rule's own predicate, copied here verbatim
// (`isSameTypeWithoutNullish` in `rules/typescript/non_nullable_type_assertion_style.go`, which is
// unexported): every asserted constituent is a non-nullish constituent of the original by type
// IDENTITY, and every non-nullish original constituent is asserted. Anything this rule reports fails
// that predicate, so the style rule declines it; anything the style rule reports passes it, so this
// rule declines it. The fixtures assert both directions by running both rules on the same cast.
//
// The style rule is `off` in ahra's live configuration (enabled for a measurement on 2026-10-03, it
// found 0: the pure removals were cleaned up before this rule was built). Its fix rewrites the cast to
// `!`, which `no-non-null-assertion` then bans, so turning it on would need its message, not its
// predicate, changed. Either way the partition holds, because it is decided by the predicate.
//
// # A fallback right after the assertion does not excuse it
//
// `(flags.size as OpenAiImageSizeType) || 'auto'` and `(properties.variant as ButtonVariantType) ?? 'A'`
// report, though the missing value is caught on the next token. That is the same contract the house
// already enforces for `!`: `maybe! ?? fallback` is an error under
// `no-non-null-asserted-nullish-coalescing`. The assertion makes the fallback look unreachable to the
// compiler and to the next reader, and the narrowing it carries (`string` to a literal union, a bare
// `--flag`'s `true` to `string`) is unchecked either way. Declining the shape would be exact too; it is
// left reporting so the `as` door and the `!` door stay the same width.
//
// # What counts as stripping
//
// The original (the expression's type at the assertion, after narrowing) must be a union holding at
// least one constituent flagged Null or Undefined AND at least one that is not. The asserted type must
// have no constituent that could be nullish:
//
//   - **`any`, `unknown` on either side decline.** Asserting from `unknown` is a parse at a boundary,
//     `no-unsafe-type-assertion`'s concern; asserting to either keeps the absence representable.
//     `row.value_cents as number` on a `Record<string, unknown>` row is therefore silent. The research's
//     cited site, `row.value_cents === null ? null : (row.value_cents as number)`
//     (`FinanceSqliteDatabase.ts:5390`), reports all the same, and correctly: ruling out `null` narrows
//     `unknown` to `{} | undefined`, so the assertion removes the `undefined` the check left in.
//   - **An all-nullish original declines.** `null as ResponseType` (LinkedInClient's 204 branch) and
//     `undefined as never` manufacture a placeholder rather than strip a value, and neither has a present
//     case to lie about.
//   - **`as never` declines.** It erases the whole type to get past an assignment, the escape hatch
//     `as any` is, and says nothing about presence in particular. `no-unsafe-type-assertion` owns it.
//     ahra's three (`FinancePositionCommandLineInterface.ts:487`, `:648`, `GoogleAdsKeywordApi.ts:282`)
//     are left to that rule.
//   - **`void` is not counted, on either side.** A `void` constituent says the result is to be ignored,
//     not that it is absent, and `fn() as Promise<void>` over `void | Promise<void>` is a different
//     question. On the asserted side `void` is treated as possibly nullish, so an assertion to it declines.
//   - **A generic asserted type answers through its constraint.** A type parameter, indexed access or
//     conditional type with no constraint could be instantiated with `undefined`, so `cache.get(key) as
//     T` declines; one constrained to `string` reports. This is upstream's `couldBeNullish`, widened from
//     type parameters to every instantiable type, because `M[K]` can carry `undefined` exactly as `T` can.
//   - **`as const` declines with no guard of its own.** The type the checker answers for the `const`
//     type node is the expression's own, so a nullable original asserts a nullable type and the
//     asserted-side test declines it; a non-nullable one has nothing to strip. The style rule keeps a
//     syntactic guard for this and records it as subsumed; this rule does not carry one.
//   - **An assignment target declines.** `(maybe as string) = 'x'` writes a value rather than reading
//     one, so nothing missing is passed on.
//
// # Cost
//
// Two type lookups per assertion, and the asserted side only once the original is known to hold a
// nullish constituent, which most assertions in real code do not.
//
// # No fix
//
// Whether a missing value should throw, default, return early, or stay in the type is the author's
// call, and a fixer choosing one would be wrong for most sites.
var CorrectnessNoNullishStrippingAssertion = rule.Rule{
	Name: "nexus/correctness-no-nullish-stripping-assertion",

	// Every finding is decided by the types on both sides of the assertion; there is no syntactic
	// subset of this rule that could run without the checker.
	NeedsTypeChecker: true,
	// The rule asks only for the types at the assertion itself, never into an imported body.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		check := func(node *ast.Node, expression *ast.Node, typeNode *ast.Node) {
			if correctnessNoNullishStrippingAssertionStrips(ctx.TypeChecker, node, expression, typeNode) {
				ctx.ReportNode(node, correctnessNoNullishStrippingAssertionMessage())
			}
		}
		return rule.Listeners{
			ast.KindAsExpression: func(node *ast.Node) {
				check(node, node.AsAsExpression().Expression, node.AsAsExpression().Type)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				check(node, node.AsTypeAssertion().Expression, node.AsTypeAssertion().Type)
			},
		}
	},
}

// correctnessNoNullishStrippingAssertionStrips is the whole decision for one assertion, in the order
// the doc comment gives it: the syntactic declines, the original side, the asserted side, then the
// partition with the style rule.
func correctnessNoNullishStrippingAssertionStrips(
	typeChecker *checker.Checker,
	node *ast.Node,
	expression *ast.Node,
	typeNode *ast.Node,
) bool {
	if expression == nil || typeNode == nil {
		return false
	}
	if ast.IsAssignmentTarget(node) {
		return false
	}

	// `any` and `unknown` need no decline of their own: neither is a union, neither carries the Null or
	// Undefined flag, so the nothing-to-strip test below declines both.
	originalType := typeChecker.GetTypeAtLocation(expression)
	if originalType == nil {
		return false
	}
	originalTypes := type_checking.UnionTypeParts(originalType)
	nonNullishOriginalTypes := make([]*checker.Type, 0, len(originalTypes))
	for _, originalType := range originalTypes {
		if type_checking.IsTypeFlagSet(originalType, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		nonNullishOriginalTypes = append(nonNullishOriginalTypes, originalType)
	}
	// Nothing nullish to strip, or nothing but nullish.
	if len(nonNullishOriginalTypes) == len(originalTypes) || len(nonNullishOriginalTypes) == 0 {
		return false
	}

	assertedType := typeChecker.GetTypeAtLocation(typeNode)
	if assertedType == nil || type_checking.IsTypeFlagSet(assertedType, checker.TypeFlagsNever) {
		return false
	}
	assertedTypes := type_checking.UnionTypeParts(assertedType)
	for _, assertedPart := range assertedTypes {
		if correctnessNoNullishStrippingAssertionCouldBeNullish(typeChecker, assertedPart, 0) {
			return false
		}
	}

	// The style rule's finding: the assertion removes the nullish constituents and nothing else.
	if correctnessNoNullishStrippingAssertionIsPureRemoval(assertedTypes, nonNullishOriginalTypes) {
		return false
	}
	return true
}

// correctnessNoNullishStrippingAssertionCouldBeNullish answers whether one asserted constituent can
// hold `undefined` or `null`, so that asserting to it keeps the absence representable.
//
// `any`, `unknown` and `void` can. An instantiable type (a type parameter, `M[K]`, a conditional)
// answers for its base constraint, and one with no constraint answers yes, because a caller can
// instantiate it with `undefined`. The depth bound is a guard against a constraint chain that
// refers back to itself; a chain that deep answers yes, which declines.
func correctnessNoNullishStrippingAssertionCouldBeNullish(typeChecker *checker.Checker, candidate *checker.Type, depth int) bool {
	if candidate == nil || depth > 8 {
		return true
	}
	if type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid|
		checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return true
	}
	if type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsInstantiable) {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, candidate)
		if constraint == nil || constraint == candidate {
			return true
		}
		for constraintPart := range type_checking.UnionTypePartsSeq(constraint) {
			if correctnessNoNullishStrippingAssertionCouldBeNullish(typeChecker, constraintPart, depth+1) {
				return true
			}
		}
	}
	return false
}

// correctnessNoNullishStrippingAssertionIsPureRemoval is `isSameTypeWithoutNullish` from
// `rules/typescript/non_nullable_type_assertion_style.go`, minus the two checks this rule has already
// made by the time it is asked (something nullish was dropped, and no asserted constituent could be
// nullish, which is a strictly wider test than the style rule's). What remains is the identity
// comparison both ways, by pointer, because the checker interns a type and hands back the same object
// for it from anywhere in one program.
func correctnessNoNullishStrippingAssertionIsPureRemoval(assertedTypes []*checker.Type, nonNullishOriginalTypes []*checker.Type) bool {
	for _, assertedPart := range assertedTypes {
		if !correctnessNoNullishStrippingAssertionContains(nonNullishOriginalTypes, assertedPart) {
			return false
		}
	}
	for _, originalPart := range nonNullishOriginalTypes {
		if !correctnessNoNullishStrippingAssertionContains(assertedTypes, originalPart) {
			return false
		}
	}
	return true
}

func correctnessNoNullishStrippingAssertionContains(types []*checker.Type, wanted *checker.Type) bool {
	for _, candidate := range types {
		if candidate == wanted {
			return true
		}
	}
	return false
}
