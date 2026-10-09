package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noLoneBlocksFile is where the fixtures pretend to live.
const noLoneBlocksFile = "/repository/source/NoLoneBlocks.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Verbatim from `eslint/tests/lib/rules/no-lone-blocks.js`, extracted by loading that file with the
// rule tester stubbed out. 23 valid and 27 invalid.
//
// # The one case that is NOT imported, and why
//
// Upstream's corpus runs `{ function bar() {} }` twice, at the SAME ecmaVersion 6, and gets
// opposite verdicts. The discriminator is strictness, not the language version, and it is easy to
// misread as the latter because the two rows differ in a nested `parserOptions` key:
//
//	valid    ecmaVersion 6 + ecmaFeatures.impliedStrict: true
//	invalid  ecmaVersion 6                                (sloppy)
//
// A function declaration in a block binds to the block in strict mode and hoists out of it in
// sloppy mode, so the braces are load-bearing in one and redundant in the other. That is upstream's
// `sourceCode.getScope(node).isStrict` test, which is why the rule reads scope for this one arm.
//
// Measured against the installed eslint 10.8.1 build rather than inferred: `sourceType: "module"`
// gives zero findings and `sourceType: "script"` gives one, on identical source at ecmaVersion
// 2022. A TypeScript file is a module and is therefore always strict, so the sloppy row has no
// reachable configuration here. It is the single upstream case deliberately not imported, and
// `TestNoLoneBlocksTreatsEveryFileAsStrict` pins the verdict we do produce for it.

// The clean cases. Each block below is load-bearing for a different reason.
//
// The first three are blocks that belong to a statement rather than standing alone. The rest carry
// a block-scoped binding -- `let`, `const`, a class, a function declaration, `using` -- which is
// what the braces exist to contain. The static-block rows are the same question one scope in.
func TestNoLoneBlocksStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"if (foo) { if (bar) { baz(); } }",
		"do { bar(); } while (foo)",
		"function foo() { while (bar) { baz() } }",
		"{ let x = 1; }",
		"{ const x = 1; }",
		"'use strict'; { function bar() {} }",
		"{ function bar() {} }",
		"{ class Bar {} }",
		"{ {let y = 1;} let x = 1; }",
		"\n          switch (foo) {\n            case bar: {\n              baz;\n            }\n          }\n        ",
		"\n          switch (foo) {\n            case bar: {\n              baz;\n            }\n            case qux: {\n              boop;\n            }\n          }\n        ",
		"\n          switch (foo) {\n            case bar:\n            {\n              baz;\n            }\n          }\n        ",
		"function foo() { { const x = 4 } const x = 3 }",
		"class C { static {} }",
		"class C { static { foo; } }",
		"class C { static { if (foo) { block; } } }",
		"class C { static { lbl: { block; } } }",
		"class C { static { { let block; } something; } }",
		"class C { static { something; { const block = 1; } } }",
		"class C { static { { function block(){} } something; } }",
		"class C { static { something; { class block {}  } } }",
		"\n{\n  using x = makeDisposable();\n}",
		"\n{\n  await using x = makeDisposable();\n}",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, sourceText))
		})
	}
}

// The failing cases, with the exact ids the corpus names and in its order.
//
// Two messages, and which one fires is decided by the block's PARENT: a block sitting directly
// inside another block or inside a static block is `redundantNestedBlock`, everything else is
// `redundantBlock`. The three-finding case pins both the count and the interleaving.
func TestNoLoneBlocksFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
	}{
		{"{}", []string{"redundantBlock"}},
		{"{var x = 1;}", []string{"redundantBlock"}},
		{"foo(); {} bar();", []string{"redundantBlock"}},
		{"if (foo) { bar(); {} baz(); }", []string{"redundantNestedBlock"}},
		{"{ \n{ } }", []string{"redundantBlock", "redundantNestedBlock"}},
		{"function foo() { bar(); {} baz(); }", []string{"redundantNestedBlock"}},
		{"while (foo) { {} }", []string{"redundantNestedBlock"}},
		{"{var x = 1;}", []string{"redundantBlock"}},
		{"{ \n{var x = 1;}\n let y = 2; } {let z = 1;}", []string{"redundantNestedBlock"}},
		{"{ \n{let x = 1;}\n var y = 2; } {let z = 1;}", []string{"redundantBlock"}},
		{"{ \n{var x = 1;}\n var y = 2; }\n {var z = 1;}", []string{"redundantBlock", "redundantNestedBlock", "redundantBlock"}},
		{"\n              switch (foo) {\n                case 1:\n                    foo();\n                    {\n                        bar;\n                    }\n              }\n            ", []string{"redundantBlock"}},
		{"\n              switch (foo) {\n                case 1:\n                {\n                    bar;\n                }\n                foo();\n              }\n            ", []string{"redundantBlock"}},
		{"\n              function foo () {\n                {\n                  const x = 4;\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              function foo () {\n                {\n                  var x = 4;\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  if (foo) {\n                    {\n                        let block;\n                    }\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  if (foo) {\n                    {\n                        block;\n                    }\n                    something;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    block;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    let block;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    const block = 1;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    function block() {}\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    class block {}\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    var block;\n                  }\n                  something;\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  something;\n                  {\n                    var block;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  {\n                    block;\n                  }\n                  something;\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
		{"\n              class C {\n                static {\n                  something;\n                  {\n                    block;\n                  }\n                }\n              }\n            ", []string{"redundantNestedBlock"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The strictness row the corpus splits, resolved for a TypeScript file.
//
// Upstream runs `{ function bar() {} }` twice at ecmaVersion 6 and gets opposite verdicts, differing
// only in `impliedStrict`. Measured against the installed eslint 10.8.1 build: zero findings under
// `sourceType: "module"`, one under `"script"`. A TypeScript file is a module and a module body is
// always strict, so the block-scoped reading is the only reachable one and the block is clean.
func TestNoLoneBlocksTreatsEveryFileAsStrict(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, "{ function bar() {} }"))

	// The control, so the silence above is about the function declaration rather than about the
	// rule declining every block: the same shape holding a `var` reports.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"{ var bar = 1; }"), "redundantBlock")
}

// TypeScript's own block-scoped declarations, which eslint cannot parse and therefore has no
// opinion about.
//
// `{ enum E {} }`, `{ interface I {} }` and `{ type T = number; }` all return a fatal parse error
// from eslint 10.8.1, so there is no upstream behaviour to port and the rule has to decide. Settled
// with tsc rather than by reading: a file referencing `E`, `I` or `T` below such a block gets
// TS2304 "Cannot find name" for each, so all three bind to the block the way `let` does and the
// braces are load-bearing. Reporting them would be a false positive on correct code, and no
// imported fixture can see it because upstream's parser cannot produce the input.
//
// `namespace` inside a block is TS1235 and not legal at all; it is exempted so a recovered parse
// does not add a finding to source that is already broken.
func TestNoLoneBlocksExemptsTypeScriptBlockScopedDeclarations(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"{ enum E { A } }",
		"{ interface I {} }",
		"{ type T = number; }",
		"{ namespace N {} }",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, sourceText))
		})
	}

	// The control: a block holding only a statement still reports, so the exemptions above are
	// about the declaration kinds rather than about the rule going quiet on this shape.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"{ foo(); }"), "redundantBlock")
}

// A static block's own body is not a lone block, and this is a PARSER DIFFERENCE rather than a rule
// decision.
//
// ESTree gives a `StaticBlock` node that IS the statement list, so upstream never sees a block in
// that position. typescript-go gives a `ClassStaticBlockDeclaration` whose `Body` is a separate
// `KindBlock`, which looks exactly like a redundant block inside a static block. Without the guard,
// eight of upstream's clean cases reported and eleven of its failing ones reported twice.
//
// Kept here rather than only in the corpus table, because the corpus rows read as ordinary static
// block cases and this names what they are actually protecting.
func TestNoLoneBlocksDoesNotReportAStaticBlockBody(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{
		"class C { static {} }",
		"class C { static { foo; } }",
		"class C { static { if (foo) { block; } } }",
	} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, sourceText))
		})
	}

	// A real block INSIDE a static block still reports, and as nested, which is what shows the
	// guard is about the body node rather than about static blocks generally.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"class C { static { { block; } } }"), "redundantNestedBlock")
}

// Cases measured against the installed eslint 10.8.1 build that the corpus does not write.
func TestNoLoneBlocksMatchesTheInstalledBuild(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// The second arm in isolation: the inner block holds a `let` and is still reported, because
		// it is the outer block's only statement. A port with only the lone-block test finds one
		// here instead of two.
		{"a bound block as its parent's only statement", "{ {let y = 1;} }",
			[]string{"redundantBlock", "redundantNestedBlock"}},

		// The same shape with `var`, which reports for both reasons at once and still gives two
		// findings rather than three, because upstream's arms are an if/else.
		{"an unbound block as the only statement", "{ {var y = 1;} }",
			[]string{"redundantBlock", "redundantNestedBlock"}},

		// A default clause behaves as a case clause does: the whole-clause block exempts.
		{"a block wrapping a whole default clause", "switch (foo) { default: { bar; } }", nil},
		{"a block after a statement in a default clause", "switch (foo) { default: foo(); { bar; } }",
			[]string{"redundantBlock"}},

		// A labeled block is the label's body, not a lone block, and the label is what makes the
		// braces reachable by `break lbl`.
		{"a labeled block", "label: { foo; }", nil},

		// Two sibling bound blocks: neither is its parent's only statement, so only the outer
		// reports, and it reports because it holds no binding of its own.
		{"two bound sibling blocks", "{ { let x = 1; } { let y = 2; } }", []string{"redundantBlock"}},

		// A bound block beside a binding: the outer holds a `let`, so nothing reports.
		{"a binding beside a bound block", "{ let x = 1; { let y = 2; } }", nil},

		// A block as a function body's only statement is nested, not lone-at-program-level.
		{"a bound block alone in a function body", "function f() { { let x; } }",
			[]string{"redundantNestedBlock"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The spans, which no message-id fixture can see.
//
// Upstream reports on the block node, so the finding covers the braces and everything between them.
// The two-finding case is asserted because it is the one where a port could anchor the outer
// finding on the inner block and still satisfy every id assertion.
func TestNoLoneBlocksSpansTheBlock(t *testing.T) {
	t.Parallel()

	const sourceText = "{ {let y = 1;} }"

	result := rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, sourceText)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted two diagnostics, got %d", len(result.Diagnostics))
	}

	outer := result.Diagnostics[0]
	if got := sourceText[outer.Range.Pos():outer.Range.End()]; got != sourceText {
		t.Fatalf("outer finding spans %q, wanted the whole block %q", got, sourceText)
	}
	if outer.Message.Id != "redundantBlock" {
		t.Fatalf("outer finding is %q, wanted redundantBlock", outer.Message.Id)
	}

	inner := result.Diagnostics[1]
	if got := sourceText[inner.Range.Pos():inner.Range.End()]; got != "{let y = 1;}" {
		t.Fatalf("inner finding spans %q, wanted %q", got, "{let y = 1;}")
	}
	if inner.Message.Id != "redundantNestedBlock" {
		t.Fatalf("inner finding is %q, wanted redundantNestedBlock", inner.Message.Id)
	}
}

// The identity test in both arms: "the parent's only statement" means THIS block, not merely that
// the parent has one statement.
//
// Found as two surviving mutants, one per arm, both deleting `Nodes[0] == block` while keeping
// `len(Nodes) == 1`. The distinguishing input is a block nested one construct deeper inside a
// single-statement parent, where `Nodes[0]` is that construct rather than the block:
//
//	{ if (a) { bar; } }        Nodes[0] is the IfStatement, not the consequent block
//	{ while(a) { bar; } }      Nodes[0] is the WhileStatement
//	{ label: { bar; } }        Nodes[0] is the LabeledStatement
//
// Without the identity test the consequent block satisfies "the only statement" and reports, which
// would be a false positive on the block a construct requires. All four rows measured against the
// installed eslint 10.8.1 build: the OUTER block reports once and the inner one never does.
//
// The switch rows are the same question in the clause arm, and they invert: a clause holding one
// statement exempts the block only when the block IS that statement, so `case 1: if (a) { bar; }`
// is clean throughout while `case 1: { bar; }` is clean for a different reason.
func TestNoLoneBlocksIdentityNotJustLength(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a required consequent block", "{ if (a) { bar; } }", []string{"redundantBlock"}},
		{"a required loop body", "{ while(a) { bar; } }", []string{"redundantBlock"}},
		{"a labeled block", "{ label: { bar; } }", []string{"redundantBlock"}},
		{"a try block", "{ try { bar; } catch(e) {} }", []string{"redundantBlock"}},
		{"a bound consequent block", "{ if (a) { let x; } }", []string{"redundantBlock"}},

		// The clause arm, where the same shapes are clean because the clause is not a block.
		{"a consequent block in a clause", "switch (foo) { case 1: if (a) { bar; } }", nil},
		{"a loop body in a clause", "switch (foo) { case 1: while(a) { bar; } }", nil},
		{"a labeled block in a clause", "switch (foo) { case 1: label: { bar; } }", nil},
		{"a try block in a clause", "switch (foo) { case 1: try { bar; } catch(e) {} }", nil},

		// A function declaration's body is required too, and the outer block reports for holding a
		// block-scoped binding's sibling rather than for the body.
		{"a function body", "{ function f() { g; } }", nil},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// A block-scoped block still reports when it is a function body's ONLY statement, and the
// TypeScript exemption does not rescue it.
//
// Written after a seeded probe made this look like a defect in the TypeScript exemption. It is not:
// the second arm reports a block that is all its parent contains regardless of what it holds, so a
// `let` and a `type` are both reported there, and upstream does the same. Measured on eslint 10.8.1:
// `function f() { { let x = 1; void x; } }` reports `redundantNestedBlock`, and adding one statement
// beside the block makes it clean.
//
// The pair matters because the exemption and the second arm answer at different points, and only
// the sibling statement separates them. Without the second row, a reader hitting the first would
// reasonably conclude the exemption was broken.
func TestNoLoneBlocksSecondArmOutranksTheBindingExemption(t *testing.T) {
	t.Parallel()

	// Sole statement of a function body: reports, bindings notwithstanding.
	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"function f() { { let x = 1; void x; } }"), "redundantNestedBlock")

	rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"function f() { { type Local = number; const v: Local = 1; void v; } }"),
		"redundantNestedBlock")

	// One statement beside it, so the second arm no longer applies and the binding exemption is
	// what decides. Both go clean.
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"function f() { { let x = 1; void x; } g(); }"))

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoLoneBlocks, noLoneBlocksFile,
		"function f() { g(); { type Local = number; const v: Local = 1; void v; } }"))
}
