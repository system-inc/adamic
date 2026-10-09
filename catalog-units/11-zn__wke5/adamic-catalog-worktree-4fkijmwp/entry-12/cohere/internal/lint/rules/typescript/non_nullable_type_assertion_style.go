package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NonNullableTypeAssertionStyle flags a type assertion whose only effect is removing null and
// undefined, which a `!` says in one character.
//
//	valid:   declare const original: number | string; const cast = original as string;
//	valid:   declare const original: number | undefined; const cast = original as any;
//	valid:   type T = string | null; declare const x: T; const y = x as NonNullable<T>;
//	valid:   const foo = [] as const;
//	invalid: declare const maybe: string | undefined; const bar = maybe as string;
//	invalid: declare const maybe: string | null | undefined; const bar = maybe as string;
//
// An assertion that only strips nullish is `!` written the long way, and the long way is worse than
// verbose: `as string` keeps asserting `string` after somebody widens the variable to
// `string | number`, silently, while `!` narrows whatever is there and keeps failing loudly when the
// underlying type changes. So this is a rule about which assertion survives a refactor, not about
// character count.
//
// # The judgment is type IDENTITY, and identity is why the checker is required
//
// Upstream compares the asserted type's constituents against the original's with `includes`, which
// is reference equality on interned `ts.Type` objects rather than any structural comparison. That is
// reproduced exactly: our checker interns the same way, measured by printing pointers for the same
// `string` reached through an expression and through a type node and finding one address. A
// structural comparison would be a different rule, and a looser one.
//
// The report fires only when all three hold: the original has at least one nullish constituent, every
// asserted constituent is a non-nullish constituent of the original, and every non-nullish original
// constituent is asserted. That third clause is what keeps `number | null | undefined as number|null`
// silent, and the corpus tests it.
//
// # The fixture harness must be one file per program, and measuring this cost a cycle
//
// Two of upstream's own failing cases assert to a `type` alias and to an `interface`, and both depend
// on the asserted object type being the SAME interned object as the one in the original union. A
// replay harness that put all twenty cases in one program had several files each declaring
// `type Type`, the checker interned one per file, identity failed, and those two cases came back
// clean. It reads exactly like a rule that cannot see object types. Running each case in its own
// program reproduced all twenty. rule_testing.RunTyped builds a one-file program per call, so the
// fixtures below are naturally isolated, and this is recorded because the same instrument bug is
// available to anyone re-measuring this rule.
//
// # Parentheses, where our tree keeps a node upstream's parser drops
//
// Upstream decides whether its rewrite needs wrapping by asking whether the asserted expression binds
// tighter than a unary operator, and it takes the replacement text from that expression. Its parser
// folds parentheses away, so `(await p()) as string` reaches its rule as the await itself. Ours keeps
// KindParenthesizedExpression, so without an unwrap the rule would be reading a different node than
// upstream reads and asking the precedence question about the wrong thing.
//
// Reading our node directly happens to produce upstream's answer on most shapes, which is what makes
// this worth writing down: a parenthesized expression's precedence is above unary, so nothing more
// gets added and the parentheses already in the text survive into the output. That accident covers
// `(await p())`, `(c ? a : a)` and `(0, a)`, and it fails on exactly one shape, `(a) as string`, where
// upstream drops the redundant parentheses and writes `a!`. The unwrap in the body puts every shape on
// upstream's own footing rather than relying on the accident.
//
// Measured on the installed build across bare await, a parenthesized await, a parenthesized
// conditional, a comma expression, redundant parentheses around an identifier, an angle-bracket
// assertion and a plain call, with the precedence our own parser assigns each one printed beside it.
// Recorded at this length because "our AST keeps a node upstream's drops, and the fixer reads that
// node's text" is exactly the shape that ships a wrong repair while every message id assertion stays
// green.
//
// # `as const` is guarded separately because the types are equal
//
// A const assertion yields the same type object as the expression, so nothing downstream would report
// it and the guard reads as redundant. It is not: `getTypeAtLocation` on the `const` type node
// answers the expression's own type, so for `x as const` where x is `string | null` BOTH sides are
// that union, the original has nullish, and the asserted constituents are a subset. Without the guard
// the rule reports `const foo = [] as const;` and rewrites it to `[]!`, changing what the code means.
// Upstream guards on the syntax rather than the type for the same reason.
//
// # Cost
//
// Two rare anchors, and the checker is consulted only when one is found. Both type lookups are
// declined early when the expression is any or unknown, which is the common shape in the code this
// rule runs on.
var NonNullableTypeAssertionStyle = rule.Rule{
	Name: "@typescript-eslint/non-nullable-type-assertion-style",

	// Every finding is decided by comparing two types by identity, so there is no syntactic subset of
	// this rule that could run without the checker.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// report is shared by the two assertion spellings, which differ only in where the expression
		// and the type node live. `x as T` and `<T>x` are the same judgment and upstream registers
		// one visitor over both.
		report := func(node *ast.Node, expression *ast.Node, typeNode *ast.Node) {
			// Without this the rule is not merely wrong, it is silently wrong: the shim answers nil
			// from GetSymbolAtLocation and friends rather than panicking, so a missing guard buys a
			// vacuous green rather than an obvious crash.
			if ctx.TypeChecker == nil {
				return
			}
			if expression == nil || typeNode == nil {
				return
			}

			if isConstTypeAssertion(typeNode) {
				return
			}

			originalTypes, isUsable := unionConstituentsIfNotLoose(ctx.TypeChecker, expression)
			if !isUsable {
				return
			}
			assertedTypes, isUsable := unionConstituentsIfNotLoose(ctx.TypeChecker, typeNode)
			if !isUsable {
				return
			}

			if !isSameTypeWithoutNullish(ctx.TypeChecker, assertedTypes, originalTypes) {
				return
			}

			// The replacement is the expression's own text plus `!`, wrapped when the expression does
			// not bind tighter than a unary operator.
			//
			// The unwrap ahead of both is what makes this agree with upstream byte for byte rather
			// than nearly. Upstream's parser folds a parenthesis away, so its `expression` is
			// whatever the parentheses held and its `getText` never includes them, while ours keeps
			// the KindParenthesizedExpression and its text carries them. Reading our node directly
			// happens to produce upstream's answer on most shapes, because a parenthesized
			// expression's precedence is above unary so nothing more is added and the parentheses
			// already present survive. It disagrees on exactly one: `(a) as string`, where upstream
			// drops the redundant parentheses and writes `a!` while reading our node writes `(a)!`.
			// Both are valid and mean the same thing, and this is a rule whose fix is applied
			// unattended across a tree, so producing upstream's exact text is worth one loop.
			//
			// After the unwrap the precedence test is asked of the same node upstream asks it of, so
			// `await p() as string` and `(await p()) as string` both reach the await at precedence
			// 16, fail the test, and are rewritten as `(await p())!`. Measured both ways against the
			// installed build.
			//
			// Written as a loop rather than a single step because parentheses nest, and deliberately
			// not `ast.SkipParentheses`: that helper dereferences its argument, and reaching it with
			// nil is a panic that costs every rule this file.
			assertedExpression := expression
			for assertedExpression.Kind == ast.KindParenthesizedExpression {
				inner := assertedExpression.AsParenthesizedExpression().Expression
				if inner == nil {
					break
				}
				assertedExpression = inner
			}

			// `rule.TokenRange` supplies the text rather than `Pos()`, which includes leading trivia
			// and would pull the preceding whitespace into the replacement as well as into the span.
			expressionRange := rule.TokenRange(ctx.SourceFile, assertedExpression)
			expressionText := ctx.SourceFile.Text()[expressionRange.Pos():expressionRange.End()]

			precedence := ast.GetOperatorPrecedence(assertedExpression.Kind, ast.KindUnknown,
				ast.OperatorPrecedenceFlagsNone)
			replacement := expressionText + "!"
			if precedence <= ast.OperatorPrecedenceUnary {
				replacement = "(" + expressionText + ")!"
			}

			ctx.ReportNodeWithFixes(node, preferNonNullAssertionMessage,
				rule.ReplaceRange(rule.TokenRange(ctx.SourceFile, node), replacement))
		}

		return rule.Listeners{
			ast.KindAsExpression: func(node *ast.Node) {
				assertion := node.AsAsExpression()
				report(node, assertion.Expression, assertion.Type)
			},
			ast.KindTypeAssertionExpression: func(node *ast.Node) {
				assertion := node.AsTypeAssertion()
				report(node, assertion.Expression, assertion.Type)
			},
		}
	},
}

var preferNonNullAssertionMessage = rule.Message{
	Id:          "preferNonNullAssertion",
	Description: "Use a ! assertion to more succinctly remove null and undefined from the type.",
}

// unionConstituentsIfNotLoose is upstream's `getTypesIfNotLoose`.
//
// A type that is `any` or `unknown` carries no information the comparison could use, and asserting
// away from one is the whole point of an assertion, so the rule declines rather than reporting. The
// second return says whether the answer is usable, which upstream expresses as undefined.
//
// # The early decline is SUBSUMED here, and it is kept anyway
//
// Neutralizing this branch survived the whole fixture set, which is the sweep telling the truth: the
// comparison below already declines every shape it could see. Measured over six inputs rather than
// argued, with a reporting control in the same run. Asserting TO `any` or `unknown` leaves an asserted
// list holding one type that is not a constituent of the original, so the membership loop declines;
// asserting FROM one leaves an original list of a single non-nullish type, so the nothing-was-dropped
// guard declines first. Both routes reach silence.
//
// It stays because it is upstream's, because it is the cheaper of the two routes on the shape this
// rule meets most often in real code, and because the subsumption is a fact about the current
// comparison rather than about the rule. A future change to the membership test that made `any`
// compare loosely would turn this from redundant into load-bearing with nothing to notice.
func unionConstituentsIfNotLoose(
	typeChecker *checker.Checker,
	node *ast.Node,
) ([]*checker.Type, bool) {
	nodeType := typeChecker.GetTypeAtLocation(node)
	if nodeType == nil {
		return nil, false
	}
	if type_checking.IsTypeFlagSet(nodeType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return nil, false
	}
	// A non-union is its own single constituent, which is what upstream's unionConstituents answers
	// and what makes the comparison below read the same for both shapes.
	return type_checking.UnionTypeParts(nodeType), true
}

// couldBeNullish is upstream's `couldBeNullish`.
//
// A type parameter answers for its constraint, and an unconstrained one answers true, because nothing
// rules out a caller instantiating it with null. That is the branch keeping `function f<T>(x: T)
// { return x as string; }` silent, and it is reachable: measured against the installed build, the
// unconstrained form is clean while `T extends string | null` reports.
func couldBeNullish(typeChecker *checker.Checker, candidate *checker.Type) bool {
	if candidate == nil {
		return false
	}

	if type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsTypeParameter) {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, candidate)
		if constraint == nil {
			return true
		}
		return couldBeNullish(typeChecker, constraint)
	}

	if type_checking.IsUnionType(candidate) {
		for part := range type_checking.UnionTypePartsSeq(candidate) {
			if couldBeNullish(typeChecker, part) {
				return true
			}
		}
		return false
	}

	return type_checking.IsTypeFlagSet(candidate,
		checker.TypeFlagsNull|checker.TypeFlagsUndefined)
}

// isSameTypeWithoutNullish is upstream's `sameTypeWithoutNullish`.
//
// The two lists are compared by IDENTITY rather than structurally, which is upstream's `includes` on
// an array of interned type objects. Three conditions, all required:
//
//	the original must lose something          otherwise the assertion is not about nullish at all
//	every asserted type is a non-nullish      otherwise the assertion widens or changes the type
//	  constituent of the original
//	every non-nullish original is asserted    otherwise the assertion narrows further than `!` would
func isSameTypeWithoutNullish(
	typeChecker *checker.Checker,
	assertedTypes []*checker.Type,
	originalTypes []*checker.Type,
) bool {
	nonNullishOriginalTypes := make([]*checker.Type, 0, len(originalTypes))
	for _, originalType := range originalTypes {
		if type_checking.IsTypeFlagSet(originalType,
			checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		nonNullishOriginalTypes = append(nonNullishOriginalTypes, originalType)
	}

	// Nothing was dropped, so the assertion is not removing nullish and this rule has no opinion.
	if len(nonNullishOriginalTypes) == len(originalTypes) {
		return false
	}

	for _, assertedType := range assertedTypes {
		if couldBeNullish(typeChecker, assertedType) {
			return false
		}
		if !containsType(nonNullishOriginalTypes, assertedType) {
			return false
		}
	}

	for _, originalType := range nonNullishOriginalTypes {
		if !containsType(assertedTypes, originalType) {
			return false
		}
	}

	return true
}

// containsType is upstream's `Array.prototype.includes` on a list of types.
//
// Pointer equality rather than any structural test, because the checker interns a type and hands back
// the same object for the same type reached from anywhere in one program. Measured by printing
// addresses for `string` reached through an expression and through a type node in the same file.
func containsType(types []*checker.Type, wanted *checker.Type) bool {
	for _, candidate := range types {
		if candidate == wanted {
			return true
		}
	}
	return false
}

// isConstTypeAssertion is upstream's `isConstAssertion`.
//
// It keys on the SYNTAX, `as const`, rather than on the resulting type.
//
// # This guard is SUBSUMED here too, and the reason is worth stating because the doc comment it
// replaces asserted the opposite
//
// The obvious reading is that the guard is load-bearing: a const assertion yields the expression's
// own type, so the comparison cannot tell it from an assertion that strips nullish, and without the
// guard `const foo = [] as const;` would be rewritten to `[]!`. That reading is wrong, and only a
// probe says so. Neutralizing the guard survived every fixture, so ten const-assertion shapes were
// run through the comparison directly with the verdict printed: an empty array, a nullable string, a
// nullable union, a numeric literal, a nullable object, a constrained type parameter, `any`, an object
// literal, and a nullable readonly array. All ten declined, none reported.
//
// The mechanism is that `as const` returns the expression's own type, so the asserted list is the
// original list. When the original holds nothing nullish the nothing-was-dropped guard declines, and
// when it does hold something nullish that constituent is in the ASSERTED list too, where couldBeNullish
// rejects it. There is no third arrangement.
//
// It is kept because it is upstream's, because it declines in one syntactic step what otherwise costs
// two type queries and a pair of loops, and because the subsumption belongs to today's comparison
// rather than to the rule.
func isConstTypeAssertion(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return false
	}
	typeName := typeNode.AsTypeReferenceNode().TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return false
	}
	return typeName.AsIdentifier().Text == "const"
}
