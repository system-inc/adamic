package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// Upstream's own rows, all of them, replay in no_shadow_corpus_test.go. The tests here pin shapes
// the corpus does not write, each measured against the installed rule. Every one runs under RunTyped
// rather than Run, and not because the rule asks a type question: the locals tables this rule reads
// are populated by the binder, and the binder runs with the program, so under the plain harness every
// table is empty and every case would pass vacuously clean. TestNoShadowNeedsTheTypedHarness pins it.

// TestNoShadowDeclaresTheTypeChecker pins the declaration itself.
//
// The rule reads no type, so a later reader can reasonably conclude the flag is over-declared and
// remove it. Removing it makes every locals table empty and the whole rule silently, vacuously
// clean, which no fixture above could distinguish from a clean input. This asserts the flag so that
// revert fails loudly.
func TestNoShadowDeclaresTheTypeChecker(t *testing.T) {
	t.Parallel()

	if !NoShadow.NeedsTypeChecker {
		t.Fatal("NoShadow must declare NeedsTypeChecker: the binder populates the locals tables " +
			"this rule reads, and without the program every table is empty and the rule goes " +
			"vacuously clean on every file")
	}
}

// TestNoShadowNeedsTheTypedHarness is the measurement behind the comment above.
//
// It runs a case that reports under RunTyped through the plain harness and asserts the rule finds
// nothing there. That is not a property worth having, it is the failure mode being recorded: if
// this ever starts reporting, the untyped path became viable and the declaration can be revisited.
func TestNoShadowNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "var a = 3;\nfunction b() {\n  var a = 10;\n}\n"

	typedResult := rule_testing.RunTyped(t, NoShadow, "file.ts", source)
	if len(typedResult.Diagnostics) != 1 {
		t.Fatalf("control failed: the typed harness must report this case once, got %d",
			len(typedResult.Diagnostics))
	}

	untypedResult := rule_testing.Run(t, NoShadow, "file.ts", source)
	if len(untypedResult.Diagnostics) != 0 {
		t.Fatalf("the untyped harness now reports %d findings on a case that needs the binder; "+
			"the NeedsTypeChecker reasoning in the doc comment is stale",
			len(untypedResult.Diagnostics))
	}
}

// TestNoShadowSpanAndMessage asserts WHERE a finding points and WHAT it says.
//
// ExpectFindings asserts ids and count and nothing else, so a rule pointing at the wrong node or
// rendering the wrong text passes every case above. The message interpolates the shadowed
// declaration line and column, so it is asserted for equality rather than by prefix: a Contains
// check on an interpolated value is weaker than the property it guards.
func TestNoShadowSpanAndMessage(t *testing.T) {
	t.Parallel()

	source := "var a = 3;\nfunction b() {\n  var a = 10;\n}\n"
	result := rule_testing.RunTyped(t, NoShadow, "file.ts", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}

	diagnostic := result.Diagnostics[0]

	// The harness writes strings.TrimSpace(source)+"\n" to disk, so slice what it wrote rather than
	// the literal above, or the span reads one byte off.
	written := strings.TrimSpace(source) + "\n"
	reported := written[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != "a" {
		t.Errorf("finding should point at the inner binding name, got %q", reported)
	}

	// The outer `var a` is on line 1, and its name is at column 5.
	wantMessage := "This binding takes a name that is already bound in an enclosing scope. Every " +
		"read of the name inside this scope now resolves here instead of to the outer binding, " +
		"so code written against the outer one silently reads the inner one. Rename the inner " +
		"binding. The name 'a' is already declared in the upper scope on line 1 column 5."
	if diagnostic.Message.Description != wantMessage {
		t.Errorf("message text\n got: %q\nwant: %q", diagnostic.Message.Description, wantMessage)
	}
	if diagnostic.Message.Id != "noShadow" {
		t.Errorf("message id: got %q, want %q", diagnostic.Message.Id, "noShadow")
	}
}

// TestNoShadowEnumMessage asserts the second message arm, which no span test above reaches.
//
// The enum arm is reported through a different accessor from every other binding, because an enum
// declaration is not a locals container and its members live on the enum symbol Exports table. Its
// message differs too, so both are pinned here.
func TestNoShadowEnumMessage(t *testing.T) {
	t.Parallel()

	source := "enum A {\n  A,\n  B,\n}\n"
	result := rule_testing.RunTyped(t, NoShadow, "file.ts", source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}

	diagnostic := result.Diagnostics[0]
	if diagnostic.Message.Id != "noEnumShadow" {
		t.Errorf("message id: got %q, want %q", diagnostic.Message.Id, "noEnumShadow")
	}

	wantMessage := "This enum member takes a name that is already bound in an enclosing scope. " +
		"Enum members are added to the enum's own scope, so a later member initializer that " +
		"names it resolves to this member rather than to the outer binding. Rename the member. " +
		"The name 'A' is already declared in the upper scope on line 1 column 6."
	if diagnostic.Message.Description != wantMessage {
		t.Errorf("message text\n got: %q\nwant: %q", diagnostic.Message.Description, wantMessage)
	}
}

// TestNoShadowHandlesShapesTheCorpusDoesNotWrite covers inputs upstream had no reason to write.
//
// Upstream parser folds parentheses away, so its corpus cannot express a parenthesized initializer
// at all, and the exemption that keeps `var a = function a() {}` clean is reached through an
// identity test that a KindParenthesizedExpression would break. A destructured parameter is the
// other shape: reading a binding name without a kind guard panics on a binding pattern, and the
// walk recovers per FILE, so one such parameter would cost every rule in the package its verdict.
func TestNoShadowHandlesShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name   string
		source string
	}{
		{
			// Parenthesized, so the identity test must see through the parentheses the way it sees
			// through `||` and a conditional. Upstream is silent on the unparenthesized form and
			// its parser cannot produce this one.
			name:   "parenthesized function expression initializer",
			source: "var a = (function a() {});\n",
		},
		{
			// Doubly parenthesized, because parentheses nest and a single unwrap would miss it.
			name:   "doubly parenthesized function expression initializer",
			source: "var a = ((function a() {}));\n",
		},
		{
			// A destructured parameter name is a binding pattern rather than an identifier, and
			// reading Text() on one panics. Nothing here shadows, so the assertion is that the rule
			// survives the shape at all.
			name:   "destructured parameter",
			source: "var options = 1;\nfunction foo({ a, b }: { a: number; b: number }) {}\n",
		},
		{
			// An array pattern and a rest parameter, same reason.
			name:   "array pattern and rest parameter",
			source: "function foo([first, second]: number[], ...rest: number[]) {}\n",
		},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoShadow, "file.ts", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}

	// Control for the two parenthesized cases above: without the parentheses the exemption applies
	// and the case is clean, so a wrapped call must still REPORT. If this went silent the unwrap
	// would be over-reaching rather than correct.
	t.Run("parenthesized wrapped call still reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoShadow, "file.ts", "var a = (wrap(function a() {}));\n")
		rule_testing.ExpectFindings(t, result, "noShadow")
	})
}

// TestNoShadowInferTypeParameters pins the infer-over-infer exemption and its boundary.
//
// Upstream default corpus contains exactly ONE infer case, and it is clean, so a port that
// exempted `infer` outright would pass the whole imported set while silencing two shapes upstream
// reports. Every verdict below was measured by driving the installed rule at 8.67.0 rather than
// read off the source; the commands are recorded in the helper doc comment.
func TestNoShadowInferTypeParameters(t *testing.T) {
	t.Parallel()

	t.Run("infer over infer is clean", func(t *testing.T) {
		t.Parallel()
		source := "export type A<F> = F extends (a: Array<infer T>) => any\n" +
			"  ? T[]\n" +
			"  : F extends (...a: infer T) => any\n" +
			"    ? T\n" +
			"    : never;\n"
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source))
	})

	// The two boundary cases. Both REPORT on the installed rule, and both would go silent under a
	// blanket infer exemption.
	t.Run("infer over an outer type alias reports", func(t *testing.T) {
		t.Parallel()
		source := "type T = string;\nexport type A<F> = F extends (a: infer T) => any ? T : never;\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})

	t.Run("infer over an outer type parameter reports", func(t *testing.T) {
		t.Parallel()
		source := "export type A<T> = T extends (a: infer T) => any ? T : never;\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})

	// Clean, but for the type-versus-value reason rather than for the infer one. Recorded so a
	// later reader does not attribute it to this exemption.
	t.Run("infer over an outer value is clean", func(t *testing.T) {
		t.Parallel()
		source := "const T = 1;\nexport type A<F> = F extends (a: infer T) => any ? T : never;\n"
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source))
	})
}

// TestNoShadowMergedEnumReportsOnce pins the merged-enum deduplication.
//
// Two `enum A` declarations merge onto one symbol, so the Exports table reached from either lists
// every member of both. Without the containment test in collectEnumMembers this input reports
// TWICE at the identical span, which reads as a correct finding in any count and is visible only
// in the count itself.
func TestNoShadowMergedEnumReportsOnce(t *testing.T) {
	t.Parallel()

	source := "enum A {\n  B = 2,\n}\n\nenum A {\n  A = 1,\n}\n"
	result := rule_testing.RunTyped(t, NoShadow, "file.ts", source)
	rule_testing.ExpectFindings(t, result, "noEnumShadow")
}

// TestNoShadowFunctionTypeParameterNames covers a shape upstream corpus does not write.
//
// The defaults corpus tests `ignoreFunctionTypeParameterNameValueShadow` only where the outer
// binding is a TYPE, and that case is already clean for a different reason: a type shadowed by a
// value crosses the declaration-space boundary. So the guard for this option survived a mutation of
// the whole imported set, and the shape that separates them is an outer VALUE.
//
// Every verdict was measured by driving the installed rule at 8.67.0. Five type-position parameter
// forms are clean over an outer value; a real function parameter over the same value reports.
func TestNoShadowFunctionTypeParameterNames(t *testing.T) {
	t.Parallel()

	clean := []struct {
		name   string
		source string
	}{
		{"function type", "const cb = 1;\ntype F = (cb: number) => void;\n"},
		{"constructor type", "const cb = 1;\ntype F = new (cb: number) => void;\n"},
		{"method signature", "const cb = 1;\ninterface I {\n  m(cb: number): void;\n}\n"},
		{"call signature", "const cb = 1;\ninterface I {\n  (cb: number): void;\n}\n"},
		{
			"type parameter constraint",
			"const Args = 1;\nfunction foo<T extends (Args: any) => void>(arg: T) {}\n",
		},
	}
	for _, testCase := range clean {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", testCase.source))
		})
	}

	// The control that makes the five above mean something: the same outer value shadowed by a
	// REAL parameter reports, so the exemption is about type position rather than about the name.
	t.Run("control: a real function parameter reports", func(t *testing.T) {
		t.Parallel()
		source := "const cb = 1;\nfunction foo(cb: number) {}\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})
}

// TestNoShadowSiblingExpressionNames covers the shape that makes the same-scope containment test
// load bearing, and it is a shape upstream corpus does not write.
//
// A function-expression name is collected from AST shape rather than from a locals table, so two of
// them in the same scope land in the same slice. They are SIBLINGS, not nested, and upstream gives
// each its own scope so neither can see the other. Without the containment test in
// findEnclosingBinding the second would be paired with the first and reported.
//
// Measured on the installed rule at 8.67.0: the two sibling forms are clean, and the three forms
// where one name genuinely encloses or precedes a real declaration all report.
func TestNoShadowSiblingExpressionNames(t *testing.T) {
	t.Parallel()

	t.Run("two sibling function expressions of one name are clean", func(t *testing.T) {
		t.Parallel()
		source := "var b = function a() {};\nvar c = function a() {};\n"
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source))
	})

	t.Run("two sibling class expressions of one name are clean", func(t *testing.T) {
		t.Parallel()
		source := "var b = class a {};\nvar c = class a {};\n"
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source))
	})

	// The controls. Each reports upstream, so the exemption above is about sibling position rather
	// than about expression names being exempt.
	t.Run("control: expression name over a preceding declaration reports", func(t *testing.T) {
		t.Parallel()
		source := "function a() {}\nvar b = function a() {};\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})

	t.Run("control: expression name over a following declaration reports", func(t *testing.T) {
		t.Parallel()
		source := "var b = function a() {};\nfunction a() {}\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})
}

// TestNoShadowInterfaceTypeParameters covers the third substrate gap of the enum family.
//
// An interface declaration is NOT a locals container, so its type parameters are invisible to the
// scope walk exactly the way an enum's members are, and they need collecting from AST shape. A type
// ALIAS is a locals container and needs none of this, which is what makes the absence easy to miss:
// the two read as the same construct and behave differently.
//
// Upstream default corpus writes no interface type parameter anywhere, so nothing imported can see
// this. It was found by driving the installed rule at 8.67.0 over shapes the corpus does not write.
func TestNoShadowInterfaceTypeParameters(t *testing.T) {
	t.Parallel()

	t.Run("interface type parameter over an outer type reports", func(t *testing.T) {
		t.Parallel()
		source := "type T = 1;\ninterface I<T> {\n  x: T;\n}\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})

	// The control that pins the comparison: the type alias form reports too, and it reaches the
	// answer through the locals table rather than through the shape collection above.
	t.Run("type alias type parameter over an outer type reports", func(t *testing.T) {
		t.Parallel()
		source := "type T = 1;\ntype A<T> = T;\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})

	// An interface with NO type parameters is the common case and its TypeParameters field is nil.
	// Dereferencing it would panic, and the walk recovers per FILE, so one such interface would
	// cost every rule in the package its verdict on that file. The assertion is that a binding
	// nested INSIDE the interface's members is still reached, which a bare early return would lose.
	t.Run("plain interface does not abandon the subtree", func(t *testing.T) {
		t.Parallel()
		source := "type T = 1;\ninterface Plain {\n  x: number;\n}\nfunction f<T>(a: T) {}\n"
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoShadow, "file.ts", source), "noShadow")
	})
}

// TestNoShadowTemporalDeadZone covers the two halves of the default hoist setting that upstream
// corpus exercises only partially.
//
// `hoist` defaults to `functions-and-types`, so an inner binding written textually BEFORE the outer
// one it shadows reports only when the outer declaration hoists. The corpus writes the function
// case and the plain `var`/`let` cases; it writes no interface or type alias in that position, so
// the other two thirds of the hoisting set survived a mutation of the whole imported suite.
//
// Every verdict measured against the installed rule at 8.67.0.
func TestNoShadowTemporalDeadZone(t *testing.T) {
	t.Parallel()

	reports := []struct {
		name   string
		source string
	}{
		// The three declaration kinds that hoist under the default setting.
		{"outer function declaration", "{\n  let a;\n}\nfunction a() {}\n"},
		{"outer interface", "{\n  type a = 1;\n}\ninterface a {}\n"},
		{"outer type alias", "{\n  type a = 1;\n}\ntype a = 2;\n"},
		// Not a dead zone at all: the parameter's name sits inside the function's own name range,
		// so the inner binding does not precede the outer one. This is the case that separates a
		// comparison against the outer name POSITION from one against its END.
		{"overlapping positions", "function a(a) {}\n"},
	}
	for _, testCase := range reports {
		t.Run(testCase.name+" reports", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoShadow, "file.ts", testCase.source), "noShadow")
		})
	}

	// The counterpart: a declaration that does NOT hoist leaves the earlier inner binding clean.
	// Without these the hoisting set could be widened to everything and nothing would notice.
	clean := []struct {
		name   string
		source string
	}{
		{"outer class", "{\n  type a = 1;\n}\nclass a {}\n"},
		{"outer let", "{\n  let a;\n}\nlet a;\n"},
		{"outer var", "{\n  let a;\n}\nvar a;\n"},
	}
	for _, testCase := range clean {
		t.Run(testCase.name+" is clean", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoShadow, "file.ts", testCase.source))
		})
	}
}
