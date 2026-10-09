package base

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// verifyArrayLevelDecorators operate on the array container itself, so their presence satisfies
// array parity and a property carrying one is never flagged.
//
// These strings are Base's decorator names and are matched literally. See the package doc: they are
// not this program's old name, and renaming them makes the rule match nothing and report nothing.
//
// `VerifyIsNotEmpty` is here rather than among the value-level rules, and the original states why:
// emptiness is a whole-container check, so on an array-typed property it evaluates the array itself
// and rejects `[]`. It therefore satisfies parity on its own.
//
// Its membership is REDUNDANT with the unknown-rule arm, and that took a paired mutation to see.
// Removing the name from this set alone SURVIVES every fixture, because the name then falls through
// to the unrecognised branch, which also suppresses. Removing it here AND adding it to the
// known-non-value set fails 5 lines. So the two sites cover for each other, and the name is kept
// here because this is where it belongs semantically rather than because only this placement works.
var verifyArrayLevelDecorators = map[string]struct{}{
	"VerifyIsArray":          {},
	"VerifyArrayMinimumSize": {},
	"VerifyArrayMaximumSize": {},
	"VerifyArrayUnique":      {},
	"VerifyIsNotEmpty":       {},
}

// verifyKnownValueLevelDecorators are the built-in rules that auto-iterate array elements.
//
// The rule reports only when one of THESE is present, which is a deliberate narrowing rather than a
// convenience: a custom `VerifyBy` rule's target cannot be read statically and may well be
// array-level, so keying on a known list is what keeps it from producing a false positive.
var verifyKnownValueLevelDecorators = map[string]struct{}{
	"VerifyIsBoolean":     {},
	"VerifyIsCountryCode": {},
	"VerifyIsDate":        {},
	"VerifyIsDefined":     {},
	"VerifyIsDomain":      {},
	"VerifyIsEmail":       {},
	"VerifyIsEnum":        {},
	"VerifyIsInteger":     {},
	"VerifyIsIpAddress":   {},
	"VerifyIsLocale":      {},
	"VerifyIsNumber":      {},
	"VerifyIsObject":      {},
	"VerifyIsPhoneNumber": {},
	"VerifyIsString":      {},
	"VerifyIsTimeZone":    {},
	"VerifyIsUrl":         {},
	"VerifyIsUuid":        {},
	"VerifyLength":        {},
	"VerifyMaximum":       {},
	"VerifyMaximumDate":   {},
	"VerifyMaximumLength": {},
	"VerifyMinimum":       {},
	"VerifyMinimumDate":   {},
	"VerifyMinimumLength": {},
	"VerifyStringMatches": {},
}

// verifyKnownNonValueDecorators are the names that are known and are not value-level element rules:
// every array-level rule, the optional sentinel, and the custom-rule factory.
//
// A `Verify*` name in NEITHER this set nor the value-level set is an unrecognised custom rule, and
// the rule suppresses on it. That is the whole reason this third set exists rather than testing the
// value-level set alone.
var verifyKnownNonValueDecorators = func() map[string]struct{} {
	known := map[string]struct{}{"VerifyIsOptional": {}, "VerifyBy": {}}
	for name := range verifyArrayLevelDecorators {
		known[name] = struct{}{}
	}
	return known
}()

// correctnessRequireVerifyArrayParityText is the rule's message, whose wording lives in
// `policy/messages/correctness-require-verify-array-parity.json`.
var correctnessRequireVerifyArrayParityText = policy.MessageOf("base/correctness-require-verify-array-parity", "missingArrayRule")

// CorrectnessRequireVerifyArrayParity requires an array-typed property with value-level rules to declare an
// array-level rule too.
//
//	valid:   @VerifyIsString() name: string                    (not an array)
//	valid:   @VerifyIsArray() @VerifyIsString() names: string[]
//	valid:   @VerifyIsNotEmpty() @VerifyIsString() names: string[]
//	valid:   @VerifyBy(check) names: string[]                  (unknown rule, may be array-level)
//	invalid: @VerifyIsString() names: string[]
//
// # Three sets, and the third is what prevents a false positive
//
// A decorator is array-level, a known value-level rule, or neither. The first suppresses, the
// second arms the finding, and the third ALSO suppresses: an unrecognised `Verify*` name is a
// custom rule whose target cannot be read from the source, and it may itself be array-level.
// Collapsing that third case into "ignore it" would report every property carrying a custom rule
// alongside a built-in one.
//
// # The array question is a type question, and it has to see through a union
//
// `string[] | null` is an array-typed property. Measured before this rule was written: neither
// `Checker_isArrayType` nor `IsTupleType` answers true for that union directly, and both answer
// true for its member, so the walk over union parts is load-bearing rather than defensive.
//
// The two predicates together cover every shape the original's `typeIsArray` accepts. Measured
// across twelve: `string[]`, `readonly string[]`, `[string, number]`, `Array<string>` and
// `ReadonlyArray<string>` all answer true, and the last two need no special case because the
// checker normalises them to the first two. `string`, `Set<string>` and `any` answer false.
//
// The original also falls back to reading the type's symbol name for `Array`/`ReadonlyArray`. Not
// reproduced, because the two checker predicates already answer true for both spellings here, and a
// name comparison would additionally match a user-defined class called `Array`.
//
// # No fix
//
// The original ships none and says why: which array-level decorator to add is a judgment call, and
// `@VerifyIsArray()` is only the most common answer rather than the right one everywhere.
var CorrectnessRequireVerifyArrayParity = rule.Rule{
	Name: "base/correctness-require-verify-array-parity",

	// Whether the property is array-typed is a question only the checker can answer, since the
	// annotation may be an alias, a generic, or a union.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkVerifyArrayParity(ctx, node, node.Name())
			},
			// A constructor parameter property. See isBaseParameterProperty for why the parameter
			// kind carries both forms here where the original has a node kind of its own.
			ast.KindParameter: func(node *ast.Node) {
				if !isBaseParameterProperty(node) {
					return
				}
				checkVerifyArrayParity(ctx, node, node.Name())
			},
		}
	},
}

// checkVerifyArrayParity judges one property or parameter property.
func checkVerifyArrayParity(ctx rule.Context, node *ast.Node, key *ast.Node) {
	if ctx.TypeChecker == nil || key == nil || !ast.IsIdentifier(key) {
		return
	}

	hasKnownValueRule := false
	suppressed := false
	for _, decorator := range decorators.Of(node) {
		name := decorators.CallName(decorator)
		if name == "" || !isVerifyDecoratorName(name) {
			continue
		}
		if _, arrayLevel := verifyArrayLevelDecorators[name]; arrayLevel {
			suppressed = true
			continue
		}
		if _, valueLevel := verifyKnownValueLevelDecorators[name]; valueLevel {
			hasKnownValueRule = true
			continue
		}
		if _, known := verifyKnownNonValueDecorators[name]; !known {
			// An unrecognised custom Verify* rule, which may be array-level.
			suppressed = true
		}
	}

	if !hasKnownValueRule || suppressed {
		return
	}

	if !verifyTypeIsArray(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(key)) {
		return
	}

	ctx.ReportNode(key, rule.Message{
		Id:          correctnessRequireVerifyArrayParityText.Id,
		Description: correctnessRequireVerifyArrayParityText.Render(map[string]string{"propertyName": key.Text()}),
	})
}

// verifyTypeIsArray answers whether a type is an array, a tuple, or a union containing one.
//
// `UnionTypeParts` returns the type itself for a non-union, so the loop covers both shapes without
// a separate arm. That matters: the union case is not an edge case here, since a nullable array
// property is the ordinary way to write an optional list.
func verifyTypeIsArray(typeChecker *checker.Checker, subject *checker.Type) bool {
	if subject == nil {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(subject) {
		if checker.Checker_isArrayType(typeChecker, part) || checker.IsTupleType(part) {
			return true
		}
	}
	return false
}
