package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageNoUnsafeEnumComparisonMismatchedCondition = rule.Message{
		Id: "mismatchedCondition",
		Description: "The two values in this comparison do not have a shared enum type. Comparing an " +
			"enum member against a bare literal type-checks and then stops protecting you: the " +
			"literal is not tied to the enum, so renaming a member, changing its value, or removing " +
			"it leaves this comparison compiling and silently false. That is the failure enums exist " +
			"to prevent. Compare against the enum member itself.",
	}
	messageNoUnsafeEnumComparisonMismatchedCase = rule.Message{
		Id: "mismatchedCase",
		Description: "The case statement does not have a shared enum type with the switch predicate. " +
			"A `case` holding a bare literal where the discriminant is an enum compiles and stops " +
			"matching the moment the enum's value changes, and nothing reports it: the switch simply " +
			"falls through. Compare against the enum member instead, which also lets exhaustiveness " +
			"checking see the case.",
	}
)

// NoUnsafeEnumComparison flags a comparison between an enum value and a non-enum value.
//
//	valid:   Fruit.Apple === Fruit.Banana;
//	valid:   1 === 2;
//	valid:   declare const fruit: Fruit | 0; fruit === 0;
//	invalid: declare const fruit: Fruit; fruit === 'apple';
//	invalid: Count.One === 1;
//	invalid: switch (fruit) { case 'apple': break; }
//
// Ported from `@typescript-eslint/no-unsafe-enum-comparison`, read from the clone at
// `packages/eslint-plugin/src/rules/no-unsafe-enum-comparison.ts` plus its judgment in
// `src/rules/enum-utils/shared.ts`. No options (`schema: []`), two messages, `hasSuggestions: true`
// and NO `fixable` key.
//
// The whole 85-case corpus was extracted from upstream's own tester and verified byte for byte
// through a second independent extraction.
//
// # The rule file is thin and the judgment is not
//
// `no-unsafe-enum-comparison.ts` is 106 lines and mostly plumbing: it reads two types and calls
// `isMismatchedEnumComparisonTypes`. That predicate lives in `enum-utils/shared.ts` and is where the
// whole port is. Sizing this rule by its own file would have been wrong by a factor of three, which
// is the trap the standard records for a thin rule over a shared substrate.
//
// # Every type capability this needs was probed before any code was written
//
// The constraint on this batch was that a missing type capability is a finding about the shelf
// rather than something to work around inside a rule. All of them are present, established with a
// compiling probe over four discriminating inputs rather than by grepping:
//
//	UnionTypeParts            splits `Fruit | Vegetable` into constituents
//	IsTypeFlagSet             EnumLiteral, NumberLike, StringLike, NumberLiteral all reachable
//	IsSymbolFlagSet           SymbolFlagsEnumMember distinguishes a member from its enum
//	Type_symbol               reaches a type's symbol
//	Symbol.ValueDeclaration   a plain field, used by four shipped rules
//	GetTypeAtLocation         used 81 times across this package
//
// The probe's four rows discriminated exactly as the rule requires: `fruit === 'apple'` shows two
// enum literals on the left against a stringLike non-enum on the right, `fruit === Fruit.Apple`
// shows an enum member on both sides, `Count.One === 1` shows the numeric form, and `1 === 2` shows
// neither side carrying an enum literal, which is the control that keeps the other three from being
// vacuous.
//
// # The suggestion is NOT ported, and that is a stated decline
//
// `meta.hasSuggestions` is true and `meta.fixable` is absent, so upstream offers a repair a human
// chooses rather than one the edit engine applies. 16 corpus cases carry one.
//
// The suggestion rewrites the literal side into an enum member reference, which needs
// `getEnumKeyForLiteral`: it matches a static value against each enum literal, then reconstructs a
// source-level name from the member's declaration, with three spellings depending on whether the
// member name is an identifier, a string literal, or a computed property. Reconstructing a
// REFERENCE rather than editing a span is the shape this repository has repeatedly got wrong, and
// `ReportNodeWithSuggestions` would need each of those three spellings pinned separately.
//
// Reporting without the suggestion is the subset that can be shown correct. The finding is what
// carries the value here; the suggestion is a convenience, and a wrong one rewrites a comparison
// into a different comparison. Recorded rather than silently omitted, and the 16 cases are in the
// corpus waiting if somebody wants it.
var NoUnsafeEnumComparison = rule.Rule{
	Name:             "@typescript-eslint/no-unsafe-enum-comparison",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declared AND guarded. `NeedsTypeChecker` governs the registration path; the harness can
		// still build a `rule.Context` by hand, and `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes`
		// does exactly that. Guarding in `Run` declines the file once rather than once per node.
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()
				if expression == nil || expression.OperatorToken == nil {
					return
				}
				// Upstream's selector is `[operator=/^[<>!=]?={0,2}$/]`, which matches the six
				// comparison operators and, because the regex is unanchored in effect, also plain
				// `=`. Assignment cannot appear as a BinaryExpression operator in a position this
				// rule reaches, so the set below is the reachable half spelled out.
				if !noUnsafeEnumComparisonIsComparisonOperator(expression.OperatorToken.Kind) {
					return
				}
				leftType := ctx.TypeChecker.GetTypeAtLocation(expression.Left)
				rightType := ctx.TypeChecker.GetTypeAtLocation(expression.Right)
				if noUnsafeEnumComparisonIsMismatched(ctx.TypeChecker, leftType, rightType) {
					ctx.ReportNode(node, messageNoUnsafeEnumComparisonMismatchedCondition)
				}
			},

			ast.KindCaseClause: func(node *ast.Node) {
				clause := node.AsCaseOrDefaultClause()
				// A `default` clause has no test. Upstream returns early on `node.test == null`,
				// and our AST spells the two as different kinds, so registering only `KindCaseClause`
				// already excludes the default. The nil check remains because a malformed parse can
				// produce a case clause without one.
				if clause == nil || clause.Expression == nil {
					return
				}
				switchStatement := noUnsafeEnumComparisonEnclosingSwitch(node)
				if switchStatement == nil {
					return
				}
				leftType := ctx.TypeChecker.GetTypeAtLocation(switchStatement.Expression)
				rightType := ctx.TypeChecker.GetTypeAtLocation(clause.Expression)
				if noUnsafeEnumComparisonIsMismatched(ctx.TypeChecker, leftType, rightType) {
					ctx.ReportNode(node, messageNoUnsafeEnumComparisonMismatchedCase)
				}
			},
		}
	},
}

// noUnsafeEnumComparisonIsComparisonOperator answers upstream's operator selector.
//
// The selector is `[operator=/^[<>!=]?={0,2}$/]`, and the `{0,2}` is the part to read carefully:
// the `=` group may be EMPTY, so a bare `<` and a bare `>` both match. An earlier version of this
// function excluded them, with a doc comment asserting they were not matched -- which was an
// argument rather than a measurement, and it was wrong.
//
// Evaluated rather than re-reasoned, by running upstream's own pattern:
//
//	<  ==  ===  !=  !==  <=  >=  >  =    all match
//	<<  &&  +                            do not
//
// Corpus cases invalid[0] and invalid[1] are `Fruit.Apple < 1` and `Fruit.Apple > 1`, which is what
// caught it. An ordering comparison against an enum is exactly as unsafe as an equality one, so the
// inclusion is deliberate upstream rather than an artifact of a loose regex.
//
// Spelled out rather than regex-matched because the operator arrives as a token kind here rather
// than as text. `=` is in upstream's matched set and is absent below: an assignment is not a
// BinaryExpression comparison in a position this rule reaches, and including it would mean testing
// the types of an assignment's two sides, which is a different question.
func noUnsafeEnumComparisonIsComparisonOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsEqualsToken,
		ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken,
		ast.KindExclamationEqualsEqualsToken,
		ast.KindLessThanToken,
		ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken,
		ast.KindGreaterThanEqualsToken:
		return true
	}
	return false
}

// noUnsafeEnumComparisonEnclosingSwitch finds the switch a case clause belongs to.
//
// The clause's parent is the case block; its parent is the statement. Walked rather than assumed,
// because a malformed parse can leave either link absent.
func noUnsafeEnumComparisonEnclosingSwitch(clause *ast.Node) *ast.SwitchStatement {
	if clause.Parent == nil || clause.Parent.Parent == nil {
		return nil
	}
	if clause.Parent.Parent.Kind != ast.KindSwitchStatement {
		return nil
	}
	return clause.Parent.Parent.AsSwitchStatement()
}

// noUnsafeEnumComparisonIsMismatched answers upstream's `isMismatchedEnumComparisonTypes`.
//
// Four gates in upstream's order, and the order is load-bearing because each one exists to let
// something through that a later gate would otherwise reject.
func noUnsafeEnumComparisonIsMismatched(
	typeChecker *checker.Checker,
	leftType *checker.Type,
	rightType *checker.Type,
) bool {
	if leftType == nil || rightType == nil {
		return false
	}

	// Gate one: nothing to do with enums at all. `1 === 2`.
	leftEnumTypes := noUnsafeEnumComparisonEnumTypes(typeChecker, leftType)
	rightEnumTypes := noUnsafeEnumComparisonEnumTypes(typeChecker, rightType)
	if len(leftEnumTypes) == 0 && len(rightEnumTypes) == 0 {
		return false
	}

	// Gate two: the two sides share an enum type. `Fruit.Apple === Fruit.Banana`.
	//
	// Compared by POINTER identity, which is what upstream's `Set.has` does on `ts.Type` objects.
	// The checker interns one type object per enum per program, so identity is the right test and a
	// structural comparison would be both slower and wrong for two enums with identical members.
	rightEnumSet := map[*checker.Type]bool{}
	for _, rightEnumType := range rightEnumTypes {
		rightEnumSet[rightEnumType] = true
	}
	for _, leftEnumType := range leftEnumTypes {
		if rightEnumSet[leftEnumType] {
			return false
		}
	}

	// Gate three: a union constituent appears on both sides. `declare const fruit: Fruit.Apple | 0;
	// fruit === 0`. Splitting the union is what makes this reachable, and upstream's comment says
	// exactly why: without it, a value typed as a union of an enum and a literal cannot be compared
	// against that literal.
	leftTypeParts := type_checking.UnionTypeParts(leftType)
	rightTypeParts := type_checking.UnionTypeParts(rightType)
	for _, leftTypePart := range leftTypeParts {
		for _, rightTypePart := range rightTypeParts {
			if leftTypePart == rightTypePart {
				return false
			}
		}
	}

	// Gate four: the actual violation. One side carries an enum whose VALUE kind matches the other
	// side's primitive kind, which is what makes the comparison type-check while meaning nothing.
	return noUnsafeEnumComparisonTypeViolates(leftTypeParts, rightType) ||
		noUnsafeEnumComparisonTypeViolates(rightTypeParts, leftType)
}

// noUnsafeEnumComparisonTypeViolates answers upstream's `typeViolates`.
//
// A number-valued enum on one side against a number-like type on the other, or the string pair.
// Note it is the enum's VALUE kind that matters rather than the enum itself: a string enum compared
// against a number is not this rule's business, because TypeScript already rejects it.
func noUnsafeEnumComparisonTypeViolates(leftTypeParts []*checker.Type, rightType *checker.Type) bool {
	sawNumberEnum := false
	sawStringEnum := false
	for _, part := range leftTypeParts {
		switch noUnsafeEnumComparisonEnumValueKind(part) {
		case noUnsafeEnumValueNumber:
			sawNumberEnum = true
		case noUnsafeEnumValueString:
			sawStringEnum = true
		}
	}
	if sawNumberEnum && noUnsafeEnumComparisonIsPrimitiveLike(rightType, checker.TypeFlagsNumberLike) {
		return true
	}
	return sawStringEnum && noUnsafeEnumComparisonIsPrimitiveLike(rightType, checker.TypeFlagsStringLike)
}

// noUnsafeEnumComparisonIsPrimitiveLike answers upstream's `isNumberLike` and `isStringLike`, which
// differ only in the flag they carry.
//
// The shape is EVERY union constituent, SOME intersection constituent -- and testing the flag on
// the composite type instead is the defect this replaced. Two shapes make the difference visible,
// and both are in the corpus:
//
//	number & {}       an intersection. The composite carries neither NumberLike nor StringLike, so
//	                  a flag test on it answers false and the comparison goes unreported. Measured:
//	                  `IsTypeFlagSet(NumberLike)` is false while the intersection's number part is
//	                  NumberLike. Corpus invalid[37] through invalid[40].
//	'foo' | 'bar'     a union. Same story: the composite is neither, while every constituent is
//	                  StringLike. Corpus invalid[41] through invalid[43].
//
// The asymmetry between the two quantifiers is upstream's and is not arbitrary. A union is
// number-like only if EVERY branch is, because any branch that is not gives the value a way to be
// something else. An intersection is number-like if ANY constituent is, because an intersection is
// all of its parts at once and `number & {}` is still a number.
func noUnsafeEnumComparisonIsPrimitiveLike(t *checker.Type, flag checker.TypeFlags) bool {
	unionParts := type_checking.UnionTypeParts(t)
	if len(unionParts) == 0 {
		return false
	}
	for _, unionPart := range unionParts {
		intersectionParts := type_checking.IntersectionTypeParts(unionPart)
		if len(intersectionParts) == 0 {
			return false
		}
		matched := false
		for _, intersectionPart := range intersectionParts {
			if type_checking.IsTypeFlagSet(intersectionPart, flag) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// noUnsafeEnumValueKind is what an enum member's value is, if the type is an enum at all.
type noUnsafeEnumValueKind int

const (
	noUnsafeEnumValueNone noUnsafeEnumValueKind = iota
	noUnsafeEnumValueNumber
	noUnsafeEnumValueString
)

// noUnsafeEnumComparisonEnumValueKind answers upstream's `getEnumValueType`.
//
// Upstream returns `ts.TypeFlags.Number` or `ts.TypeFlags.String` and uses them as set members,
// which reads as a flag test and is really a three-valued answer. Named as one here, because
// returning a flag that is never tested as a flag invites the next reader to mask against it.
func noUnsafeEnumComparisonEnumValueKind(t *checker.Type) noUnsafeEnumValueKind {
	if !type_checking.IsTypeFlagSet(t, checker.TypeFlagsEnumLike) {
		return noUnsafeEnumValueNone
	}
	if type_checking.IsTypeFlagSet(t, checker.TypeFlagsNumberLiteral) {
		return noUnsafeEnumValueNumber
	}
	return noUnsafeEnumValueString
}

// noUnsafeEnumComparisonEnumTypes answers upstream's `getEnumTypes`.
//
// The enum literals in a type, mapped to the enums that DECLARE them. `Fruit.Apple` yields `Fruit`,
// which is what makes gate two work: two members of one enum share a base type and compare safely.
//
// # Which reading of `isTypeFlagSet` this needs, and why the answer differs by call site
//
// Upstream imports TWO functions of that name into `enum-utils/shared.ts` and uses each
// deliberately. `../../util`'s decomposes a union and ORs its constituents' flags; `tsutils`' reads
// the type's own flags. `getEnumLiterals` uses the first, `getEnumValueType` uses the second, and
// the file imports both on purpose.
//
// Our `type_checking.IsTypeFlagSet` is the own-flags reading, so reaching for it BY NAME at both
// sites would be wrong at one of them. Measured rather than reasoned, on `Fruit.Apple | Veg.Leek`:
//
//	whole type, own flags     false      <- what our helper answers on the union
//	whole type, union-ORed    true       <- what upstream's ../../util answers
//	per constituent           2 literals <- what this loop answers
//
// The loop is what makes it correct here: iterating `UnionTypeParts` first and testing each part
// reaches the same answer as upstream's union-decomposing helper, because decomposing is exactly
// what the loop already does. And `noUnsafeEnumComparisonEnumValueKind` is only ever called ON a
// constituent, from `typeViolates`, so the own-flags reading is the right one there and matches
// upstream's deliberate `tsutils` call.
//
// So both sites agree with upstream for different reasons, and neither would if the helper were
// swapped for the other reading. `TestNoUnsafeEnumComparisonReadsUnionsConstituentwise` pins it.
func noUnsafeEnumComparisonEnumTypes(typeChecker *checker.Checker, t *checker.Type) []*checker.Type {
	var enumTypes []*checker.Type
	for part := range type_checking.UnionTypePartsSeq(t) {
		if !type_checking.IsTypeFlagSet(part, checker.TypeFlagsEnumLiteral) {
			continue
		}
		enumTypes = append(enumTypes, noUnsafeEnumComparisonBaseEnumType(typeChecker, part))
	}
	return enumTypes
}

// noUnsafeEnumComparisonBaseEnumType answers upstream's `getBaseEnumType`.
//
// Given an enum MEMBER type, returns the type of the enum that declares it; given anything else,
// returns it unchanged. Upstream asserts the symbol is non-null with a `!`; here the nil is checked,
// because the walk recovers per FILE rather than per rule and one nil dereference costs every rule
// its verdict on that file.
func noUnsafeEnumComparisonBaseEnumType(typeChecker *checker.Checker, t *checker.Type) *checker.Type {
	symbol := checker.Type_symbol(t)
	if symbol == nil {
		return t
	}
	if !type_checking.IsSymbolFlagSet(symbol, ast.SymbolFlagsEnumMember) {
		return t
	}
	declaration := symbol.ValueDeclaration
	if declaration == nil || declaration.Parent == nil {
		return t
	}
	return typeChecker.GetTypeAtLocation(declaration.Parent)
}
