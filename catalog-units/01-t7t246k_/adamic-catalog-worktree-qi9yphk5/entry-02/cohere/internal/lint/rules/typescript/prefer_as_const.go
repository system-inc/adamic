package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferAsConst = rule.Message{
	Id: "preferAsConst",
	Description: "This literal type restates a value the compiler can already read off the " +
		"literal itself. `as const` infers it, so the value and its type cannot drift apart; " +
		"writing the type out means every edit to the value has to be made twice, and the " +
		"compiler accepts the version where only one of them was.",
}

// PreferAsConst flags a literal type that merely restates the literal value beside it.
//
//	valid:   let foo = 'bar' as const;
//	valid:   let foo: string = 'bar';
//	valid:   let foo: 'bar' = 'baz';
//	invalid: let foo: 'bar' = 'bar';
//	invalid: let foo = 'bar' as 'bar';
//	invalid: class C { bar: 2 = 2; }
//
// Ported from oxc's `typescript/prefer-as-const`, which is itself a port of
// `@typescript-eslint/prefer-as-const`. Where the two disagree this follows oxc, because oxc is
// what the differential harness compares against. Three disagreements are recorded below.
//
// # Three anchors, and the positions that are deliberately not among them
//
// Upstream registers exactly three node kinds, and every other annotated position is out of scope
// rather than overlooked:
//
//	VariableDeclarator     let foo: 'bar' = 'bar'
//	PropertyDefinition     class C { bar: 'bar' = 'bar' }
//	TSAsExpression         'bar' as 'bar'
//
// Measured against the release binary, a parameter default (`function f(p: 'bar' = 'bar') {}`) and
// a generic argument (`g<'bar'>('bar')`) are both silent, and each has a fixture. They are the two
// positions a reader most expects to be covered, which is why they are pinned rather than left to
// the absence of a listener to imply.
//
// # oxc registers no TSTypeAssertion arm, and typescript-eslint does
//
// `let foo = <'bar'>'bar';` is reported by typescript-eslint and silent here. Upstream lists the
// three angle-bracket inputs among its failing cases, but its tester parses .tsx, where they are
// `Unexpected token` parse errors rather than rule findings, and its snapshot records them as such.
// That is the entirety of the extractor's "17 diagnostics against 20 inputs" warning.
//
// Measured in .ts, where the form parses cleanly, with a control rule firing on a neighbouring
// file: all three are silent. Upstream's own fix vectors for them are commented out, so the gap is
// known upstream rather than accidental. Reproduced as silence.
//
// # The comparison is on values, not on spellings
//
// oxc compares the parsed numbers (`(a - b).abs() < f64::EPSILON`) and the cooked string contents.
// typescript-eslint compares ESTree's `raw` text. So `let a: 0x10 = 16` reports here and would not
// there, as would `2 = 2.0`, `1_0 = 10` and `1e2 = 100`.
//
// No imported fixture separates the two readings, because the corpus writes only inputs where the
// spelling and the value agree. TestPreferAsConstMatchesNumericValuesNotSpellings is the fixture
// set that pins it, and it is invented rather than imported for exactly that reason.
//
// A parser is not needed for any of it. A NumericLiteral's Text field already holds the canonical
// rendering of the double it parsed to, so string equality on Text is the float comparison, and a
// StringLiteral's Text is the cooked value rather than the source spelling. Both were probed on the
// inputs above before this comparison was written.
var PreferAsConst = rule.Rule{
	Name: "@typescript-eslint/prefer-as-const",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				// A definite assignment carrying an initializer is not a program TypeScript
				// accepts, and upstream is silent on it. Measured against the release binary with
				// a control firing on a neighbouring file: `let a!: 'bar' = 'bar';` reports
				// nothing. Declining here also keeps the annotation range from swallowing the `!`.
				if declaration.ExclamationToken != nil {
					return
				}
				// A binding pattern reports without a repair. See reportAnnotation.
				reportAnnotation(ctx, declaration.Name(), declaration.Type, declaration.Initializer,
					declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if property.PostfixToken != nil {
					return
				}
				// Upstream passes `true` unconditionally here: a property name is never a binding
				// pattern, so the repair is always well formed.
				reportAnnotation(ctx, property.Name(), property.Type, property.Initializer, true)
			},
			ast.KindAsExpression: func(node *ast.Node) {
				expression := node.AsAsExpression()
				literalType := literalTypeRestatingValue(expression.Type, expression.Expression)
				if literalType == nil {
					return
				}
				// One range, so this is atomic by construction and cannot be split by the overlap
				// resolver. `'bar' as 'bar'` and `'bar' as const` mean the same thing to the
				// compiler, so nothing about the program changes and it is a fix rather than a
				// suggestion.
				ctx.ReportNodeWithFixes(literalType, messagePreferAsConst,
					ctx.ReplaceNode(literalType, "const"))
			},
		}
	},
}

// reportAnnotation handles the two arms that carry a separate type annotation.
//
// # Why the two-part edit is a fix rather than a suggestion
//
// This is the load-bearing decision in the rule, and it goes against the shape of the tree's
// warnings, so the reasoning is recorded rather than assumed.
//
// The repair is two edits at separated offsets: delete the annotation, and append ` as const` after
// the initializer. Our engine flattens a diagnostic's fixes into independent proposals, so a pair
// that only means something jointly is normally a Suggestion, whose Fixes field is the atomic
// container. core.NoCaseDeclarations is the shipped precedent.
//
// It is a fix here for two measured reasons.
//
// Upstream ships it as one. oxc calls diagnostic_with_fix, which is FixKind::SafeFix, on both arms,
// and there is no diagnostic_with_suggestion anywhere in the rule. Running the release binary with
// a bare --fix, which applies safe fixes only and never suggestions, rewrites every one of these
// inputs. So a port shipping a suggestion would decline to repair files upstream repairs unattended,
// which is a behavioral divergence rather than a conservative reading.
//
// (typescript-eslint does the opposite, and inverts the meaning of the same `canFix` parameter: it
// makes this arm a suggestion under a distinct message id and makes the as-expression arm the fix.
// oxc keeps one message id and fixes both. oxc wins.)
//
// And the halves cannot be separated in practice. The overlap resolver refuses a proposal that
// begins before the end of the previously kept one; these two are disjoint and ordered, because the
// initializer follows the annotation. The only edit that could land between them is this rule's own
// other arm, and the two arms are mutually exclusive on a single declarator: when the initializer is
// an AsExpression the annotation arm declines, because it compares the annotation against the
// initializer and the kinds disagree. `let foo: 'bar' = 'bar' as 'bar';` reports exactly once, which
// upstream's snapshot confirms and TestPreferAsConstReportsOnceWhenBothArmsCouldApply pins.
//
// Meaning is preserved: `let foo: 'bar' = 'bar'` and `let foo = 'bar' as const` give the variable
// the same type. That is the test the fix/suggestion split actually applies.
//
// # canRepair
//
// Upstream computes it as `matches!(id, BindingPattern::BindingIdentifier(_))`, so a destructuring
// target reports and offers nothing. It has to: deleting the annotation off `let []: 'bar' = 'bar'`
// and appending `as const` leaves a destructure of a string, which is a different program and would
// not compile. The absence of that input from upstream's fourteen fix vectors is the corpus saying
// the same thing.
func reportAnnotation(ctx rule.Context, name *ast.Node, typeNode *ast.Node, initializer *ast.Node, canRepair bool) {
	literalType := literalTypeRestatingValue(typeNode, initializer)
	if literalType == nil {
		return
	}

	if !canRepair {
		ctx.ReportNode(literalType, messagePreferAsConst)
		return
	}

	// The delete runs from the colon through the end of the type. InsertAfter lands at the
	// initializer's own end, before any trailing trivia. Both match upstream byte for byte,
	// measured on inputs carrying a tab before the colon, a comment after the type, and a comment
	// after the value.
	ctx.ReportNodeWithFixes(literalType, messagePreferAsConst,
		rule.RemoveRange(annotationRange(ctx, name, typeNode)),
		ctx.InsertAfter(initializer, " as const"))
}

// literalTypeRestatingValue returns the literal type node when it restates the value beside it, and
// nil otherwise.
//
// The kinds must agree as well as the text: `let a: 2 = '2'` and `let a: '2' = 2` are both silent,
// because upstream matches the type's literal kind first and only then compares against an
// initializer of that same kind.
//
// Only strings and numbers are matched. Upstream's match on TSLiteral falls through for every other
// variant, so a boolean (`let a: true = true`), a bigint (`let a: 2n = 2n`) and a negated number
// (`let a: -2 = -2`) are silent despite reading as the same shape. Each has a fixture.
//
// A parenthesized type is not a literal type. `let a: ('bar') = 'bar'` parses to ParenthesizedType
// here and declines without any explicit skip, which is what upstream does: measured silent. Adding
// a paren skip would read as a free correctness improvement and would be a divergence.
func literalTypeRestatingValue(typeNode *ast.Node, initializer *ast.Node) *ast.Node {
	if typeNode == nil || initializer == nil || typeNode.Kind != ast.KindLiteralType {
		return nil
	}
	literal := typeNode.AsLiteralTypeNode().Literal
	if literal == nil {
		return nil
	}

	switch literal.Kind {
	case ast.KindStringLiteral:
		if initializer.Kind != ast.KindStringLiteral {
			return nil
		}
		// Text is the cooked value on both sides, so `'bar'` and `"bar"` compare equal, which is
		// the input upstream ships as a failing case.
		if literal.AsStringLiteral().Text != initializer.AsStringLiteral().Text {
			return nil
		}
	case ast.KindNumericLiteral:
		if initializer.Kind != ast.KindNumericLiteral {
			return nil
		}
		// Text is the canonical rendering of the parsed double, so this equality is oxc's float
		// comparison rather than a comparison of source spellings.
		if literal.AsNumericLiteral().Text != initializer.AsNumericLiteral().Text {
			return nil
		}
	default:
		return nil
	}

	return typeNode
}

// annotationRange is the span of `: T`, from the colon through the end of the type.
//
// oxc deletes a TSTypeAnnotation node, which owns its colon. typescript-go has no such node: a
// declaration carries the type directly, and the type's Pos already sits past the colon. So the
// colon has to be found rather than inherited, and RemoveNode on the type alone leaves it behind
// as `let foo:  = 'bar' as const;`. That is the defect the fix-vector assertions caught and the
// message-id assertions could not, because the finding is byte-identical either way.
//
// The colon is the first token at or after the end of the name. Scanning from there rather than
// searching backwards from the type keeps the whitespace and any comment written before the colon
// out of the range, which is what upstream leaves in place: measured, `let a\t:\t'bar' = 'bar'`
// becomes `let a\t = 'bar' as const`, the tab before the colon surviving.
//
// Callers guard the definite-assignment token before reaching here, so the token found is a colon
// rather than a `!`.
func annotationRange(ctx rule.Context, name *ast.Node, typeNode *ast.Node) core.TextRange {
	colon := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, name.End())
	return colon.WithEnd(typeNode.End())
}
