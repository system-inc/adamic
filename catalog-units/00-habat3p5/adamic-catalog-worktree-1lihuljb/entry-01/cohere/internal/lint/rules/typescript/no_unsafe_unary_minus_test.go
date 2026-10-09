package typescript

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeUnaryMinusFile = "/repository/source/Negate.ts"

// TestNoUnsafeUnaryMinusStaysSilent carries tsgolint's fourteen valid inputs verbatim.
//
// Copied out of `no_unsafe_unary_minus_test.go` by parsing it rather than by reading it, and the
// bytes on disk were compared against upstream's afterwards. Each was additionally driven through
// `@typescript-eslint` 8.67.0 on a real program and all fourteen were clean there too.
//
// The cases that matter most here are the ones that look like they should report. `(a: any) => -a`
// and `(a: never) => -a` are exempt because `Any` and `Never` sit in the same flag set as the
// numeric ones, which is not obvious from the rule's name. `(a: number | bigint) => -a` is the case
// that proves the union walk: the union's own flags carry `Union` and nothing else, so a version
// asking the whole type rather than its parts would report this valid input.
func TestNoUnsafeUnaryMinusStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"unary plus on a number literal", "+42;"},
		{"negating a number literal", "-42;"},
		{"negating a bigint literal", "-42n;"},
		{"negating a number parameter", "(a: number) => -a;"},
		{"negating a bigint parameter", "(a: bigint) => -a;"},
		{"negating a number or bigint union", "(a: number | bigint) => -a;"},
		{"negating an any parameter", "(a: any) => -a;"},
		{"negating a union of numeric literal types", "(a: 1 | 2) => -a;"},
		{"unary plus on a string parameter", "(a: string) => +a;"},
		{"negating an element of a number array", "(a: number[]) => -a[0];"},
		{"negating a type parameter intersected with number", "<T,>(t: T & number) => -t;"},
		{"negating a number property", "(a: { x: number }) => -a.x;"},
		{"negating a never parameter", "(a: never) => -a;"},
		{"negating a type parameter constrained to number", "<T extends number>(t: T) => -t;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeUnaryMinus,
				noUnsafeUnaryMinusFile, testCase.sourceText))
		})
	}
}

// TestNoUnsafeUnaryMinusFires carries tsgolint's nine invalid inputs verbatim.
//
// Upstream records exactly one diagnostic per input and the rule `break`s at the first offending
// union part, so no input here can report twice. `@typescript-eslint` agreed on all nine.
func TestNoUnsafeUnaryMinusFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"negating a string parameter", "(a: string) => -a;"},
		{"negating an empty object type", "(a: {}) => -a;"},
		{"negating a whole number array", "(a: number[]) => -a;"},
		{"negating a string literal", "-'hello';"},
		{"negating a template literal", "-`hello`;"},
		{"negating an object with a number property", "(a: { x: number }) => -a;"},
		{"negating an unknown parameter", "(a: unknown) => -a;"},
		{"negating a void parameter", "(a: void) => -a;"},
		{"negating an unconstrained type parameter", "<T,>(t: T) => -t;"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeUnaryMinus,
				noUnsafeUnaryMinusFile, testCase.sourceText), "unaryMinus")
		})
	}
}

// TestNoUnsafeUnaryMinusTypeBoundary walks the flag boundary the corpus leaves unwritten.
//
// The rule turns entirely on which type flags count as safe, and upstream's twenty three cases
// cover perhaps half of that surface. Every row below was measured twice before it became a
// fixture: once with a flags probe that printed the constrained type, its flags, and the flags of
// each union part, and once by driving `@typescript-eslint` 8.67.0 over the same source. Both
// references agree on every verdict here.
//
// Three rows contradict the reading a person would arrive at from the rule's name:
//
//	`unknown` REPORTS while `any` does not, because Unknown is a separate flag and only Any is exempt
//	`T extends any` REPORTS, because the constraint resolves to `unknown` rather than staying `any`
//	a whole numeric enum is SILENT, because it splits into members that are each NumberLike
func TestNoUnsafeUnaryMinusTypeBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// Safe: the union parts are NumberLike even though the union itself is not.
		{"a number or string union reports on the string part", "(a: number | string) => -a;", []string{"unaryMinus"}},
		{"a three way union reports once and stops", "(a: number | string | bigint) => -a;", []string{"unaryMinus"}},
		// `string | any` collapses to `any` inside the checker, so no union survives to walk.
		{"a union with any collapses to any and is silent", "(a: string | any) => -a;", nil},
		{"a union with never drops the never and is silent", "declare const x: number | never;\n-x;", nil},
		// Enums: a member is NumberLiteral|EnumLiteral, and the whole enum is a union of members.
		{"a numeric enum member is silent", "enum E { A = 1, B = 2 }\ndeclare const e: E.A;\n-e;", nil},
		{"a whole numeric enum is silent", "enum E { A = 1, B = 2 }\ndeclare const e: E;\n-e;", nil},
		{"a string enum reports", "enum S { A = 'a' }\ndeclare const s: S;\n-s;", []string{"unaryMinus"}},
		// Literal types sit inside NumberLike and BigIntLike.
		{"a numeric literal type is silent", "declare const n: 5;\n-n;", nil},
		{"a bigint literal type is silent", "declare const n: 5n;\n-n;", nil},
		// A boxed wrapper is an Object, which is the point of the rule.
		{"the boxed Number wrapper reports", "declare const n: Number;\n-n;", []string{"unaryMinus"}},
		{"the boxed BigInt wrapper reports", "declare const n: BigInt;\n-n;", []string{"unaryMinus"}},
		// `boolean` is internally `false | true`, so the walk splits it.
		{"a boolean reports", "declare const b: boolean;\n-b;", []string{"unaryMinus"}},
		{"null reports", "declare const x: null;\n-x;", []string{"unaryMinus"}},
		{"undefined reports", "declare const x: undefined;\n-x;", []string{"unaryMinus"}},
		// Constraint resolution happens before the walk.
		{"a type parameter constrained to a union reports", "<T extends number | string>(t: T) => -t;", []string{"unaryMinus"}},
		{"a type parameter constrained to any reports because the constraint is unknown", "<T extends any>(t: T) => -t;", []string{"unaryMinus"}},
		{"a type parameter constrained to unknown reports", "<T extends unknown>(t: T) => -t;", []string{"unaryMinus"}},
		// Only the minus operator is anchored, which is the first line of the rule.
		{"bitwise not on a string is silent", "(a: string) => ~a;", nil},
		{"logical not on a string is silent", "(a: string) => !a;", nil},
		{"unary plus on an object is silent", "(a: {}) => +a;", nil},
		{"a postfix decrement is silent", "let z = 1;\nz--;", nil},
		// Negating a string yields a number, so only the inner expression is unsafe.
		{"a doubled negation reports only on the inner one", "(a: string) => - -a;", []string{"unaryMinus"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, testCase.sourceText)
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoUnsafeUnaryMinusTheErrorTypeIsAny records the hazard this harness creates for this rule.
//
// The fixture tsconfig pins `lib: ["ES2022"]` and cannot be raised, so a type it does not carry
// resolves to the error type. That type's flags are exactly `Any`, and `Any` is in this rule's safe
// set, so the finding disappears. For a flags-based rule that is the dangerous direction: such a
// fixture would sit in the silent table, pass, and assert the opposite of upstream.
//
// This is pinned rather than merely written down, because the whole point is that it is invisible.
// `Disposable` is a real ES2022 absentee and `NotDefinedAnywhere` is a name that resolves nowhere at
// all, and the rule cannot tell them apart from a genuinely safe `any`. Upstream's own corpus names
// no such type, so no imported case needed a second file, but the next person adding one to this
// table needs to know why their new clean case might be hollow.
func TestNoUnsafeUnaryMinusTheErrorTypeIsAny(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"declare const x: Disposable;\n-x;",
		"declare const x: NotDefinedAnywhere;\n-x;",
	} {
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, sourceText)
		if len(result.Diagnostics) != 0 {
			t.Errorf("the error type stopped resolving to Any for %q, which changes what a hollow fixture looks like", sourceText)
		}
	}

	// The control: a type the lib DOES carry, negated, still reports. Without this the assertions
	// above would pass just as well if the rule had stopped running altogether.
	control := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, "declare const x: string;\n-x;")
	if len(control.Diagnostics) != 1 {
		t.Fatalf("the control found %d findings, want one, so the assertions above proved nothing", len(control.Diagnostics))
	}
}

// TestNoUnsafeUnaryMinusSpansAndText asserts where each finding points and exactly what it renders.
//
// ExpectFindings compares ids and a count and nothing else, so a rule pointing at the operand
// rather than at the whole expression passes every fixture above while being wrong. Reporting the
// operand is the natural way to write this rule, and the id fixtures cannot see the difference.
//
// The message text is asserted by equality against a literal typed here rather than against the
// rule's own message constant, because a constant compared to itself moves under mutation and stays
// green. It matters more than usual for this rule: the type name is interpolated per finding, and
// it is exactly the half where the two upstream references disagree.
func TestNoUnsafeUnaryMinusSpansAndText(t *testing.T) {
	t.Parallel()

	t.Run("the finding spans the whole unary expression, not the operand", func(t *testing.T) {
		t.Parallel()
		source := "declare const a: string;\nconst negated = -a;"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		finding := result.Diagnostics[0]
		reported := source[finding.Range.Pos():finding.Range.End()]
		if reported != "-a" {
			t.Errorf("the finding points at %q, want the whole unary expression", reported)
		}
		if finding.Message.Id != "unaryMinus" {
			t.Errorf("message id is %q", finding.Message.Id)
		}
		if finding.Message.Description != "Argument of unary negation should be assignable to number | bigint but is string instead." {
			t.Errorf("message text is %q", finding.Message.Description)
		}
	})

	t.Run("the span includes parentheses around the operand", func(t *testing.T) {
		t.Parallel()
		// Our parser keeps the parenthesized expression as a real node and upstream's does too, so
		// the operand is the ParenthesizedExpression and the report covers it. Measured against
		// `@typescript-eslint`, which spans `-(a)` here as well. This is written down because the
		// corpus contains no parenthesized form at all and a port could fall either way unnoticed.
		source := "declare const a: string;\nconst negated = -(a);"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "-(a)" {
			t.Errorf("the finding points at %q, want the parentheses included", reported)
		}
	})

	t.Run("a parenthesized whole expression reports the inner span", func(t *testing.T) {
		t.Parallel()
		source := "declare const a: string;\nconst negated = (-a);"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		if reported != "-a" {
			t.Errorf("the finding points at %q, want the unary expression without the outer parens", reported)
		}
	})

	t.Run("the message names the offending union part rather than the whole type", func(t *testing.T) {
		t.Parallel()
		// This is the measured divergence from `@typescript-eslint`, which renders
		// "is string | number instead" for the same input because it passes the WHOLE argument type
		// to typeToString while tsgolint passes the loop's current part. tsgolint wins because
		// oxlint runs tsgolint, so this assertion is what records which reference was ported.
		source := "declare const a: number | string;\nconst negated = -a;"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		if result.Diagnostics[0].Message.Description != "Argument of unary negation should be assignable to number | bigint but is string instead." {
			t.Errorf("message text is %q, want the union PART named", result.Diagnostics[0].Message.Description)
		}
	})

	t.Run("a boolean renders as false because boolean is the union false or true", func(t *testing.T) {
		t.Parallel()
		// Nothing about this input looks like a union, and this row was not predicted from reading
		// either implementation. `@typescript-eslint` renders "is boolean instead" here. It is the
		// clearest evidence that the union walk is what produces the text, and no fixture asserting
		// an id or a count could see it.
		source := "declare const b: boolean;\nconst negated = -b;"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		if result.Diagnostics[0].Message.Description != "Argument of unary negation should be assignable to number | bigint but is false instead." {
			t.Errorf("message text is %q, want the first union part named", result.Diagnostics[0].Message.Description)
		}
	})

	t.Run("each unsafe expression reports separately", func(t *testing.T) {
		t.Parallel()
		// The rule `break`s within one expression's union walk, so at most one finding per
		// expression, but two expressions produce two findings at two different offsets. Without
		// the offset comparison a rule reporting the same node twice would pass this.
		source := "declare const a: string;\ndeclare const b: {};\nconst x = -a;\nconst y = -b;"
		result := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
		if len(result.Diagnostics) != 2 {
			t.Fatalf("want two findings, got %d", len(result.Diagnostics))
		}
		if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
			t.Fatal("both findings point at the same offset")
		}
		first := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
		second := source[result.Diagnostics[1].Range.Pos():result.Diagnostics[1].Range.End()]
		if first != "-a" || second != "-b" {
			t.Errorf("findings point at %q and %q", first, second)
		}
	})
}

// TestNoUnsafeUnaryMinusRequiresTheTypedHarness pins the checker declaration and the rule name.
//
// This rule is the silent kind rather than the panicking kind under a nil checker, which is the
// more dangerous of the two: moved to `rule_testing.Run`, every Fires case would fail and every silent
// case would pass VACUOUSLY. The nil guard the standing advice asks for now lives at the top of the
// listener, because absorbing the rule off the adapter made that listener ours to edit. It is
// unreachable through registration, since `NeedsTypeChecker` is declared; it covers the harness
// path, where a Context is built by hand. This test pins the declaration AND the guard.
func TestNoUnsafeUnaryMinusRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoUnsafeUnaryMinus.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}
	if NoUnsafeUnaryMinus.Name != "@typescript-eslint/no-unsafe-unary-minus" {
		t.Errorf("the registered name is %q, and the inventory writes typescript/no-unsafe-unary-minus", NoUnsafeUnaryMinus.Name)
	}

	source := "declare const a: string;\nconst negated = -a;"
	typed := rule_testing.RunTyped(t, NoUnsafeUnaryMinus, noUnsafeUnaryMinusFile, source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness found %d findings, want one", len(typed.Diagnostics))
	}

	// The guard the absorption made possible. Driving the listener with a checker-less Context must
	// return rather than dereference nil, and this is the only path that reaches that branch, since
	// registration always supplies a checker.
	listeners := NoUnsafeUnaryMinus.Run(rule.Context{SourceFile: typed.SourceFile}, nil)
	listener, hasListener := listeners[ast.KindPrefixUnaryExpression]
	if !hasListener {
		t.Fatal("the rule stopped listening on prefix unary expressions")
	}
	for _, statement := range typed.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindVariableStatement {
			continue
		}
		for _, declarator := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if initializer := declarator.Initializer(); initializer != nil {
				listener(initializer)
			}
		}
	}
}
