package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const misusedNewFile = "/repository/source/Thing.ts"

// TestNoMisusedNewFires carries every failing case from the oxc corpus verbatim, plus cases that
// pin decisions the corpus never exercises. Counts come from partitioning the seven snapshot
// diagnostics across six inputs by the source line each one prints.
func TestNoMisusedNewFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		messageIds []string
	}{
		// The corpus, verbatim. The first input is the one the snapshot reports twice: the
		// construct signature returning the interface, and the `constructor` method signature.
		{"a construct signature and a constructor signature in one interface",
			"interface I { new (): I; constructor(): void;}",
			[]string{"interfaceConstruct", "interfaceConstructor"}},
		{"a generic construct signature returning the interface",
			"interface G { new <T>(): G<T>;}",
			[]string{"interfaceConstruct"}},
		{"a constructor signature in a type literal",
			"type T = { constructor(): void;};",
			[]string{"interfaceConstructor"}},
		{"a class method named new returning the class",
			"class C { new(): C;}",
			[]string{"classNew"}},
		{"an abstract declared class method named new returning the class",
			"declare abstract class C { new(): C;}",
			[]string{"classNew"}},
		{"a constructor signature returning a string literal type",
			"interface I { constructor(): '';}",
			[]string{"interfaceConstructor"}},

		// Beyond the corpus. Each of these was measured against the oxc release binary before it
		// was written, because reading the source alone does not settle any of them.

		// A getter named `new` with a return type matching the class. oxc treats a getter as a
		// MethodDefinition, so `method.key.is_specific_id("new")` matches it. Our parser gives a
		// getter its own kind (KindGetAccessor), so a port anchored only on KindMethodDeclaration
		// goes silent here and no corpus case can see it: the corpus's only getter, in the first
		// pass case, has no return type and would pass either way.
		// Measured: reports at col 32.
		{"a getter named new returning the class",
			"declare abstract class C { get new(): C; }",
			[]string{"classNew"}},
		// A static method named `new`. oxc reads no static flag, so staticness is not a
		// discrimination. Measured: reports.
		{"a static method named new returning the class",
			"declare class C { static new(): C; }",
			[]string{"classNew"}},
		// The return type carries type arguments while its bare name matches. oxc compares only
		// the identifier's name text and never looks at the arguments, so this reports even
		// though `C<T>` is not `C`. Measured: reports.
		{"a class method named new whose return type carries type arguments",
			"declare class C { new(): C<T>; }",
			[]string{"classNew"}},
		// A class nested inside a namespace, to pin that the anchor is the class rather than the
		// top level of the file. Measured: reports.
		{"a class inside a namespace",
			"declare namespace N { class C { new(): C; } }",
			[]string{"classNew"}},
		// A NAMED class expression. Upstream anchors on AstKind::Class, which covers the
		// declaration and expression forms alike; our parser gives them separate kinds, so the
		// expression form is a second listener that no corpus case reaches. Every class
		// expression upstream writes is anonymous, and an anonymous one returns early on the
		// missing id, so all four of them pass under a port that handles only declarations.
		// The release binary cannot answer this either: the input raises TS2391, which suppresses
		// rule output. Established by adding it to oxc's own pass vector and running
		// `cargo test -p oxc_linter --lib no_misused_new`, which reported it at column 23. oxc's
		// tree was restored and verified clean afterwards.
		{"a named class expression with a method named new",
			"const foo = class C { new(): C; };",
			[]string{"classNew"}},
		// The comparison is by name text with no resolution, so an interface nested inside a
		// namespace matches its own name even though an outer interface of the same name exists.
		// Measured: reports.
		{"a shadowed interface name matching by text",
			"interface Outer {}\nnamespace M { interface Outer { new (): Outer; } }\n",
			[]string{"interfaceConstruct"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoMisusedNew, misusedNewFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.messageIds...)
		})
	}
}

// TestNoMisusedNewStaysSilent carries every passing case from the oxc corpus verbatim, plus the
// clean side of each decision the fires table exercises.
func TestNoMisusedNewStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The corpus, verbatim.
		{"a getter named new with no return type",
			"declare abstract class C { foo(); get new();bar();}"},
		{"a constructor overload signature in a class",
			"class C { constructor();}"},
		{"a constructor overload signature in a class expression",
			"const foo = class { constructor();};"},
		{"a class expression method named new returning another name",
			"const foo = class { new(): X;};"},
		{"a class method named new with a body",
			"class C { new() {} }"},
		{"a real constructor with a body",
			"class C { constructor() {} }"},
		{"a class expression method named new with a body",
			"const foo = class { new() {} };"},
		{"a class expression constructor with a body",
			"const foo = class { constructor() {} };"},
		{"a construct signature returning an object type",
			"interface I { new (): {}; }"},
		{"a construct signature in a type literal returning the alias name",
			"type T = { new (): T };"},
		{"an anonymous default-exported class with a constructor signature",
			"export default class { constructor(); }"},
		{"a generic construct signature returning another name",
			"interface foo { new <T>(): bar<T>; }"},
		{"a generic construct signature returning a string literal type",
			"interface foo { new <T>(): 'x'; }"},

		// Beyond the corpus. The clean half of every decision above, each measured.

		// The construct-signature arm anchors on an interface declaration alone. A type literal
		// carries construct signatures too, and oxc never visits them, which is why the corpus
		// pass case `type T = { new (): T }` is clean while `type T = { constructor(): void;}`
		// fails: the two arms of this rule reach type literals differently. Measured: silent.
		{"a construct signature in a type literal whose name matches", "type A = { new (): A };"},
		// The return name text must equal the enclosing name. Measured: silent.
		{"a construct signature returning a different interface", "interface I { new (): J; }"},
		{"a class method named new returning a different name", "declare class C { new(): D; }"},
		// The return type must be a bare type reference with an identifier name. A qualified name,
		// an array type, a literal type, and a parenthesized type all decline. Measured: silent.
		{"a construct signature returning a qualified name", "interface I { new (): I.K; }"},
		{"a class method named new returning an array of the class", "declare class C { new(): C[]; }"},
		{"a class method named new returning a literal type", "declare class C { new(): \"x\"; }"},
		{"a class method named new returning a parenthesized class name", "declare class C { new(): (C); }"},
		// No return type at all. Measured: silent.
		{"a class method named new with no return type", "declare class C { new(); }"},
		// The method name must be exactly `new`. Measured: silent.
		{"a class method whose name merely starts with new", "declare class C { newx(): C; }"},
		// A setter has no return type, so it cannot match even though its name is `new` and oxc
		// would accept it as a MethodDefinition. Measured: silent.
		{"a setter named new", "declare class C { set new(v: C); }"},
		// A setter carrying an illegal return type annotation. This is the single most valuable
		// case here and it is not in the corpus, because upstream's parser cannot produce it:
		// oxc refuses to attach a return type to a setter, so its rule never sees one. Ours
		// recovers from the parse error by attaching the node, so every other condition of the
		// class arm is met and a setter arm reports it. The first draft of this rule had one and
		// did report. Established by adding this case to oxc's own pass vector and running
		// `cargo test -p oxc_linter --lib no_misused_new`: the only diagnostic was TS1095 from the
		// parser, and no-misused-new reported nothing. The release binary cannot answer this,
		// because TS1095 suppresses rule output on the file and returns a silence that means
		// nothing.
		{"a setter named new carrying an illegal return type",
			"declare class C { set new(v: C): C; }"},
		// A call signature is not a construct signature. Measured: silent.
		{"a call signature returning the interface", "interface I { (): I; }"},
		// A plain method signature returning the interface is not a construct signature and is not
		// named `constructor`. Measured: silent.
		{"a method signature returning the interface", "interface I { m(): I; }"},
		// The `constructor` arm requires a static identifier key. oxc matches
		// PropertyKey::StaticIdentifier alone, so a string-literal key and a computed key both
		// decline. This is the pair that ast.TryGetTextOfPropertyName cannot express: it happily
		// answers "constructor" for both, so a port built on that helper over-reports here and no
		// corpus case can see it. Measured: both silent.
		{"a constructor signature with a string literal key", "interface I { \"constructor\"(): void; }"},
		{"a constructor signature with a computed key", "interface I { [\"constructor\"](): void; }"},
		// A property signature named `constructor` is not a method signature. Measured: silent.
		{"a property signature named constructor", "interface I { constructor: () => void; }"},
		// A real class constructor declaration, which is KindConstructor rather than a method
		// signature and is never the subject of either arm. Measured: silent.
		{"a declared class constructor returning the class", "declare class C { constructor(): C; }"},
		// The comparison is by NAME TEXT, not by resolution. `D` is an alias of `C`, so a port
		// resolving the return type would find the enclosing class and report. oxc compares the
		// written identifier and stays silent. This is the sharpest divergence available and no
		// imported fixture covers it. Measured: silent.
		{"a class method named new returning an alias that resolves to the class",
			"declare class C { new(): D; }\ntype D = C;\n"},
		{"a construct signature returning an alias that resolves to the interface",
			"interface J {}\ntype I = J;\ninterface I2 { new (): I; }\n"},
		// And the reverse: a name that differs textually but resolves to the interface. Measured:
		// silent, which a resolving port would also get right, so this is the weaker of the pair
		// and is kept as the control for the one above.
		{"a construct signature returning an alias of the interface",
			"interface I {}\ntype A = I;\ninterface I3 { new (): A; }\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoMisusedNew, misusedNewFile, testCase.sourceText))
		})
	}
}

// TestNoMisusedNewPointsAtTheRightToken asserts what ExpectFindings structurally cannot: where each
// finding lands. Both arms of this rule report a token that is not the node the listener receives,
// so a port that reported the whole signature would pass every table above while pointing at the
// wrong span in every case.
func TestNoMisusedNewPointsAtTheRightToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantTexts  []string
	}{
		// oxc spans the construct signature's first three bytes, which is the `new` keyword, and
		// the method signature's key, which is the whole `constructor` identifier.
		{"both arms in one interface", "interface I { new (): I; constructor(): void;}",
			[]string{"new", "constructor"}},
		// The generic case pins that the span is three bytes from the signature start rather than
		// the whole `new <T>()` head.
		{"a generic construct signature", "interface G { new <T>(): G<T>;}", []string{"new"}},
		{"a constructor signature in a type literal", "type T = { constructor(): void;};",
			[]string{"constructor"}},
		// The class arm spans the method key rather than the signature.
		{"a class method named new", "class C { new(): C;}", []string{"new"}},
		{"an abstract declared class method named new", "declare abstract class C { new(): C;}",
			[]string{"new"}},
		{"a getter named new", "declare abstract class C { get new(): C; }", []string{"new"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoMisusedNew, misusedNewFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantTexts) {
				t.Fatalf("got %d findings, want %d", len(result.Diagnostics), len(testCase.wantTexts))
			}
			for index, wantText := range testCase.wantTexts {
				diagnostic := result.Diagnostics[index]
				gotText := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotText != wantText {
					t.Errorf("finding %d spans %q, want %q", index, gotText, wantText)
				}
			}
		})
	}
}

// TestNoMisusedNewMessagesReadAsWritten asserts the message text against literal strings typed
// here rather than against the rule's own constants, so that a mutation moving a constant moves
// only one side of the comparison.
func TestNoMisusedNewMessagesReadAsWritten(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoMisusedNew, misusedNewFile,
		"interface I { new (): I; constructor(): void;}")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "interfaceConstruct" {
		t.Errorf("first finding id is %q, want %q", got, "interfaceConstruct")
	}
	if got := result.Diagnostics[1].Message.Id; got != "interfaceConstructor" {
		t.Errorf("second finding id is %q, want %q", got, "interfaceConstructor")
	}

	classResult := rule_testing.Run(t, NoMisusedNew, misusedNewFile, "class C { new(): C;}")
	if len(classResult.Diagnostics) != 1 {
		t.Fatalf("got %d class findings, want 1", len(classResult.Diagnostics))
	}
	if got := classResult.Diagnostics[0].Message.Id; got != "classNew" {
		t.Errorf("class finding id is %q, want %q", got, "classNew")
	}
}
