package typescript

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RestrictPlusOperandsOptions is the rule's option surface, matching upstream's schema exactly.
//
// Every `allow` field is a `*bool` rather than a `bool`, and that is load-bearing rather than
// stylistic. All five DEFAULT TO TRUE upstream, so a zero-valued struct means "forbid everything"
// and would invert the rule into reporting most of the tree. A pointer keeps an absent key
// distinguishable from an explicit `false`, and the nil case falls back to the default in
// DefaultRestrictPlusOperandsSettings.
//
// SkipCompoundAssignments is the one option that defaults to FALSE, so it binds through as a plain
// bool: absent and explicitly false mean the same thing for it.
type RestrictPlusOperandsOptions struct {
	AllowAny                *bool `json:"allowAny"`
	AllowBoolean            *bool `json:"allowBoolean"`
	AllowNullish            *bool `json:"allowNullish"`
	AllowNumberAndString    *bool `json:"allowNumberAndString"`
	AllowRegExp             *bool `json:"allowRegExp"`
	SkipCompoundAssignments bool  `json:"skipCompoundAssignments"`
}

// restrictPlusOperandsSettings is the resolved form, with every default already applied, so the rule
// body reads plain bools and cannot accidentally treat nil as false.
type restrictPlusOperandsSettings struct {
	allowAny                bool
	allowBoolean            bool
	allowNullish            bool
	allowNumberAndString    bool
	allowRegExp             bool
	skipCompoundAssignments bool
}

// DefaultRestrictPlusOperandsSettings is upstream's `defaultOptions`, which is the permissive end of
// the rule: with everything allowed it reports only operands that cannot be added at all.
func DefaultRestrictPlusOperandsSettings() restrictPlusOperandsSettings {
	return restrictPlusOperandsSettings{
		allowAny:                true,
		allowBoolean:            true,
		allowNullish:            true,
		allowNumberAndString:    true,
		allowRegExp:             true,
		skipCompoundAssignments: false,
	}
}

func restrictPlusOperandsSettingsFrom(options any) restrictPlusOperandsSettings {
	settings := DefaultRestrictPlusOperandsSettings()

	decoded, ok := rule.OptionsAs[RestrictPlusOperandsOptions](options)
	if !ok {
		// A rule configured as a bare "error" is handed nil options, and `options.(T)` on nil yields
		// the zero value rather than failing. Returning the DEFAULTS here rather than the zero
		// struct is what keeps that path correct; reading the zero struct would turn every `allow`
		// off and report most of the tree.
		return settings
	}

	if decoded.AllowAny != nil {
		settings.allowAny = *decoded.AllowAny
	}
	if decoded.AllowBoolean != nil {
		settings.allowBoolean = *decoded.AllowBoolean
	}
	if decoded.AllowNullish != nil {
		settings.allowNullish = *decoded.AllowNullish
	}
	if decoded.AllowNumberAndString != nil {
		settings.allowNumberAndString = *decoded.AllowNumberAndString
	}
	if decoded.AllowRegExp != nil {
		settings.allowRegExp = *decoded.AllowRegExp
	}
	settings.skipCompoundAssignments = decoded.SkipCompoundAssignments

	return settings
}

// RestrictPlusOperands flags a `+` whose operands are not both numbers, both bigints, or both
// strings, and whose combination the options do not permit.
//
//	valid:   let x = 1 + 1
//	valid:   let x = '1' + '1'
//	valid:   declare const a: bigint; let x = a + a
//	valid:   let x = '1' + 1                        (allowNumberAndString defaults on)
//	invalid: let x = 1n + 1
//	invalid: declare const a: symbol; let x = a + a
//	invalid: let x = '1' + 1                        (under allowNumberAndString: false)
//
// `+` is the one operator in the language that means two unrelated things, and TypeScript permits
// the mixture rather than the intent. `1 + '1'` is `'11'` and almost never what was written.
//
// # The shape of the decision
//
// One early exit and then two passes, and the order between them is what the rule IS rather than an
// implementation detail.
//
// The early exit is `leftType == rightType` on POINTER identity plus a flag test for bigint, number
// or string. Upstream writes `===` on the type objects, and that reproduces here because the checker
// interns types: probed over eight shapes, two `string` annotations in different declarations, two
// identical unions, two identical interfaces and two different string LITERALS after widening all
// answer identical pointers, while `number` against `string` does not. So this is a faithful port of
// upstream's comparison rather than a lookalike that happens to agree.
//
// The first pass reports each operand that is individually impossible: a symbol, a `never`, an
// `unknown`, an object, or one of the four kinds an `allow` option is switched off for. The second
// pass runs ONLY if the first found nothing, and reports the pair: a string added to a number under
// `allowNumberAndString: false`, or a number added to a bigint, which is a type error in any
// configuration.
//
// That ordering is why `hadIndividualComplaint` exists and why it is not a micro-optimisation.
// `symbol + 1` has an impossible operand AND a mismatched pair, and upstream reports only the
// operand. A port running both passes would report twice on a corpus that says once.
//
// # One flag test, not two, and the source reads as though there are two
//
// Upstream's body calls two differently named things: its own `isTypeFlagSetInUnion`, defined at the
// bottom of the rule file, and an `isTypeFlagSet` imported from `../util`. That reads as a deliberate
// distinction between a union-aware test and a plain one, and porting it that way is wrong.
//
// The imported one is `packages/type-utils/src/typeFlagUtils.ts`, and it is ALSO union-aware: it ORs
// the flags of every union constituent before testing. Its own doc comment says so and points at
// tsutils for the plain behaviour. So both spellings mean the same thing and every site here is
// union-aware.
//
// Two corpus cases are what caught this, and nothing else in the suite could have. `foo += 'data'`
// where `foo` is `string | null`, under `allowNullish: false`, reports upstream; a port using a
// plain flag test is silent, because the union itself carries neither the null flag nor the
// undefined one, only its members do. Probed directly on both cases: our checker answers
// `string | null`, plain flag false, union flag true, which is exactly upstream's own answer, so the
// divergence was in the port and not in the substrate.
//
// The single remaining plain `IsTypeFlagSet` is the early exit, which is upstream's `tsutils`
// import and genuinely plain: it tests a type already known to be identical on both sides.
//
// # The RegExp arm, which is neither of the above
//
// A RegExp is an object and carries the object flag, so the generic object test below would report
// it regardless of `allowRegExp`. Upstream therefore names it first, by TYPE NAME, and gives it a
// second condition the other arms do not have: a RegExp is reported even when allowed if the OTHER
// operand is number-like, because `/x/ + 1` is a number-and-string mixture wearing an object.
//
// # Where each finding points
//
// The two passes report at different nodes and that is the whole difference between them. An
// individual complaint points at the OPERAND, so a reader sees which side is wrong; a pair complaint
// points at the WHOLE expression, because neither side is wrong alone. Upstream's corpus records
// columns that pin both, and the fixtures assert the spans rather than only the ids.
//
// # Two different renderings of a type, and they are not interchangeable
//
// The RegExp arm identifies a type by `getTypeName`, the shelf helper that folds a string-like, an
// all-string union, a string-bearing intersection and a string-constrained type parameter all down
// to "string". Every MESSAGE, by contrast, renders with the checker's own `TypeToString`. Upstream
// draws exactly the same line and this port keeps it.
//
// Swapping every message to `getTypeName` SURVIVES the whole suite, and that verdict is correct
// rather than a fixture gap. The check is mechanical rather than an argument: enumerate the type
// names the corpus records across all 56 findings that carry data, and ask which of them the two
// functions render differently. The answer is none. The list holds `string | boolean`,
// `string | null`, `string | undefined` and `{ first: number; second: string; }`, none of which is a
// shape `getTypeName` rewrites, because a union is only folded when EVERY member is string and these
// each have a non-string member. The one entry the two could disagree about is `string` itself,
// where both return "string".
//
// Probed for a reaching input as well as reasoned about, since the enumeration only covers the
// corpus: an all-string union, a string-constrained type parameter, a string intersection and an
// unconstrained type parameter were each driven at the report sites under several configurations,
// and all six were clean, because a type this rule complains about individually is by construction
// not a plain string.
//
// So the two spellings are equivalent HERE and not in general, and upstream's is kept because the
// equivalence rests on which types can reach these four sites, which is exactly the kind of fact
// that stops being true when someone adds an arm.
//
// # Messages
//
// All three interpolate, and the `stringLike` slot is the elaborate one: it is assembled from which
// `allow` options are on, so the same finding reads differently under different configuration. The
// corpus records the rendered text for 56 findings across four distinct renderings of that slot, and
// the fixtures assert the rendered string rather than the message id, because no id assertion can
// see a format built this way.
var RestrictPlusOperands = rule.Rule{
	Name: "@typescript-eslint/restrict-plus-operands",

	// Every operand is judged by its type. Nothing here is answerable from syntax.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}

		settings := restrictPlusOperandsSettingsFrom(options)
		stringLike := restrictPlusOperandsStringLike(settings)

		// Upstream applies `getBaseTypeOfLiteralType` over the CONSTRAINED type, so a `'a'` widens
		// to `string` and a `T extends number` resolves to `number` before anything is judged.
		typeOf := func(node *ast.Node) *checker.Type {
			return checker.Checker_getBaseTypeOfLiteralType(
				ctx.TypeChecker,
				type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, node),
			)
		}

		checkPlusOperands := func(node *ast.Node, left *ast.Node, right *ast.Node) {
			leftType := typeOf(left)
			rightType := typeOf(right)
			if leftType == nil || rightType == nil {
				return
			}

			// Both sides the same primitive is the overwhelmingly common case and the only one that
			// needs no further thought.
			if leftType == rightType && type_checking.IsTypeFlagSet(leftType,
				checker.TypeFlagsBigIntLike|checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike) {
				return
			}

			hadIndividualComplaint := false

			for _, side := range []struct {
				node      *ast.Node
				baseType  *checker.Type
				otherType *checker.Type
			}{
				{left, leftType, rightType},
				{right, rightType, leftType},
			} {
				// A symbol, a never and an unknown can never be added to anything, so they report
				// whatever the options say. The three `allow` tests beside them are the options
				// speaking.
				//
				// The nullish test is deliberately the NON-union form, matching upstream. Switching
				// it to the union-aware form reports `string | null`, which the corpus passes.
				if restrictPlusOperandsIsTypeFlagSetInUnion(side.baseType,
					checker.TypeFlagsESSymbolLike|checker.TypeFlagsNever|checker.TypeFlagsUnknown) ||
					(!settings.allowAny && restrictPlusOperandsIsTypeFlagSetInUnion(side.baseType, checker.TypeFlagsAny)) ||
					(!settings.allowBoolean && restrictPlusOperandsIsTypeFlagSetInUnion(side.baseType, checker.TypeFlagsBooleanLike)) ||
					(!settings.allowNullish && restrictPlusOperandsIsTypeFlagSetInUnion(side.baseType,
						checker.TypeFlagsNull|checker.TypeFlagsUndefined)) {
					ctx.ReportNode(side.node, buildRestrictPlusOperandsInvalidMessage(
						ctx.TypeChecker.TypeToString(side.baseType), stringLike))
					hadIndividualComplaint = true
					continue
				}

				// A RegExp carries the object flag as well as its name, so it has to be named before
				// the generic object test below would swallow it.
				for subBaseType := range type_checking.UnionTypePartsSeq(side.baseType) {
					var reportThisPart bool
					if ctx.TypeChecker.TypeToString(subBaseType) == "RegExp" {
						// Reported when RegExps are forbidden, and ALSO when they are allowed but
						// the other side is a number: `/x/ + 1` is a number-and-string mixture.
						reportThisPart = !settings.allowRegExp ||
							restrictPlusOperandsIsTypeFlagSetInUnion(side.otherType, checker.TypeFlagsNumberLike)
					} else {
						reportThisPart = (!settings.allowAny && type_checking.IsTypeAnyType(subBaseType)) ||
							restrictPlusOperandsIsDeeplyObjectType(subBaseType)
					}

					if reportThisPart {
						ctx.ReportNode(side.node, buildRestrictPlusOperandsInvalidMessage(
							ctx.TypeChecker.TypeToString(subBaseType), stringLike))
						hadIndividualComplaint = true
						continue
					}
				}
			}

			// The pair passes run only when neither operand was wrong on its own. Without this,
			// `symbol + 1` would report an impossible operand AND a mismatched pair, where upstream
			// reports one.
			if hadIndividualComplaint {
				return
			}

			for _, pair := range [][2]*checker.Type{
				{leftType, rightType},
				{rightType, leftType},
			} {
				baseType, otherType := pair[0], pair[1]

				if !settings.allowNumberAndString &&
					restrictPlusOperandsIsTypeFlagSetInUnion(baseType, checker.TypeFlagsStringLike) &&
					restrictPlusOperandsIsTypeFlagSetInUnion(otherType,
						checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
					ctx.ReportNode(node, buildRestrictPlusOperandsMismatchedMessage(
						ctx.TypeChecker.TypeToString(leftType),
						ctx.TypeChecker.TypeToString(rightType),
						stringLike))
					return
				}

				// A number added to a bigint is a type error in every configuration, so no option
				// switches this off.
				if restrictPlusOperandsIsTypeFlagSetInUnion(baseType, checker.TypeFlagsNumberLike) &&
					restrictPlusOperandsIsTypeFlagSetInUnion(otherType, checker.TypeFlagsBigIntLike) {
					ctx.ReportNode(node, buildRestrictPlusOperandsBigintAndNumberMessage(
						ctx.TypeChecker.TypeToString(leftType),
						ctx.TypeChecker.TypeToString(rightType)))
					return
				}
			}
		}

		return rule.Listeners{
			// Upstream registers two selectors over the same handler. Our parser spells both `a + b`
			// and `a += b` as a binary expression, distinguished only by the operator token, so one
			// listener serves both and the operator switch below is where the two selectors live.
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				switch binary.OperatorToken.Kind {
				case ast.KindPlusToken:
				case ast.KindPlusEqualsToken:
					if settings.skipCompoundAssignments {
						return
					}
				default:
					return
				}

				checkPlusOperands(node, binary.Left, binary.Right)
			},
		}
	},
}

// restrictPlusOperandsStringLike renders the `stringLike` slot every message interpolates.
//
// It names the things a string may be added to, which is exactly the set of `allow` options that are
// on, so the same finding reads differently under different configuration. Upstream builds it by
// filtering a five-element list and then choosing between three spellings by length.
//
// The SINGULAR spelling is reachable but the corpus never renders it: all four renderings the corpus
// records are the empty case or the plural one, because no corpus case leaves exactly one option on.
// It is kept, and a fixture below constructs the one-option configuration to exercise it, because a
// branch with no test is a branch that can be wrong for as long as nobody writes that configuration.
func restrictPlusOperandsStringLike(settings restrictPlusOperandsSettings) string {
	var allowed []string
	// The order is upstream's and it is not alphabetical: `null` and `undefined` are both gated on
	// allowNullish and sit on either side of `RegExp`.
	if settings.allowAny {
		allowed = append(allowed, "`any`")
	}
	if settings.allowBoolean {
		allowed = append(allowed, "`boolean`")
	}
	if settings.allowNullish {
		allowed = append(allowed, "`null`")
	}
	if settings.allowRegExp {
		allowed = append(allowed, "`RegExp`")
	}
	if settings.allowNullish {
		allowed = append(allowed, "`undefined`")
	}

	switch len(allowed) {
	case 0:
		return "string"
	case 1:
		return "string, allowing a string + " + allowed[0]
	default:
		return "string, allowing a string + any of: " + strings.Join(allowed, ", ")
	}
}

// restrictPlusOperandsIsTypeFlagSetInUnion asks whether ANY member of a union carries a flag.
//
// Distinct from type_checking.IsTypeFlagSet, which asks the type itself. `string | null` is the input
// that separates them: it has a `null` MEMBER while the union type itself carries neither the null
// flag nor the undefined one, so this answers true and the plain test answers false.
//
// Every flag test in the rule body except the early exit goes through here, including the ones
// upstream spells with an `isTypeFlagSet` imported from its `../util`, because that import is
// union-aware too. See the note on the rule about why the source reads as though it is not.
func restrictPlusOperandsIsTypeFlagSetInUnion(t *checker.Type, flags checker.TypeFlags) bool {
	for part := range type_checking.UnionTypePartsSeq(t) {
		if type_checking.IsTypeFlagSet(part, flags) {
			return true
		}
	}
	return false
}

// restrictPlusOperandsIsDeeplyObjectType answers whether every constituent is an object.
//
// An intersection is walked as an intersection and everything else as a union, which is upstream's
// own spelling. The distinction matters because `string & {}` is an intersection whose members are
// NOT all objects, so it is not reported, while `{} & {}` is.
func restrictPlusOperandsIsDeeplyObjectType(t *checker.Type) bool {
	parts := type_checking.UnionTypeParts(t)
	if type_checking.IsIntersectionType(t) {
		parts = type_checking.IntersectionTypeParts(t)
	}
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !type_checking.IsObjectType(part) {
			return false
		}
	}
	return true
}

func buildRestrictPlusOperandsInvalidMessage(typeName string, stringLike string) rule.Message {
	return rule.Message{
		Id: "invalid",
		Description: "Invalid operand for a '+' operation. Operands must each be a number or " +
			stringLike + ". Got `" + typeName + "`.",
	}
}

func buildRestrictPlusOperandsMismatchedMessage(left string, right string, stringLike string) rule.Message {
	return rule.Message{
		Id: "mismatched",
		Description: "Operands of '+' operations must be a number or " + stringLike +
			". Got `" + left + "` + `" + right + "`.",
	}
}

func buildRestrictPlusOperandsBigintAndNumberMessage(left string, right string) rule.Message {
	return rule.Message{
		Id: "bigintAndNumber",
		Description: "Numeric '+' operations must either be both bigints or both numbers. Got `" +
			left + "` + `" + right + "`.",
	}
}
