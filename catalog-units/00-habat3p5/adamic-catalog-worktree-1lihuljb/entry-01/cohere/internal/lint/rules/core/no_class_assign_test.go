package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// classAssignFile is where the fixtures pretend to live.
//
// A real path matters more here than for a syntactic rule: this rule reads the checker, so the
// harness builds an actual program and the file has to sit somewhere a tsconfig can reach.
const classAssignFile = "/repository/source/ClassAssign.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_class_assign.rs`:
// one Tester block, 17 pass and 8 fail. The snapshot records 9 diagnostics from those 8 inputs, so
// one finding per input is wrong here, and the extractor said so before any code was written. The
// extra diagnostic belongs to `class A { } A = 0; A = 1;`, which writes twice and reports twice.
func TestNoClassAssignFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a plain reassignment after the declaration", "class A { } A = 0;",
			[]string{"noClassAssign"}},
		{"a shorthand destructuring target", "class A { } ({A} = 0);",
			[]string{"noClassAssign"}},
		{"a destructuring target with a default", "class A { } ({b: A = 0} = {});",
			[]string{"noClassAssign"}},
		{"a reassignment hoisted above the declaration", "A = 0; class A { }",
			[]string{"noClassAssign"}},
		{"a reassignment from inside a method", "class A { b() { A = 0; } }",
			[]string{"noClassAssign"}},
		{"a named class expression reassigned inside its own body",
			"let A = class A { b() { A = 0; } }", []string{"noClassAssign"}},
		// The discrepancy case. Two writes, two findings.
		{"two reassignments of the same class", "class A { } A = 0; A = 1;",
			[]string{"noClassAssign", "noClassAssign"}},
		{"a reassignment inside the declaring block", "if (foo) { class A {} A = 1; }",
			[]string{"noClassAssign"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoClassAssign, classAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases are the whole discrimination, and most of them are a shadow of some shape.
//
// `class A { b(A) { A = 0; } }` writes to a parameter, `class A { b() { let A; A = 0; } }` to a
// local, and `let A = class A { }; A = 1;` to the variable rather than to the class name, which only
// binds inside the class body. Those three are the reason this rule reads the checker at all: no
// bounded walk distinguishes them from the fail cases, because the text is identical.
//
// The if/else pair is the sharpest. Two sibling classes named `A` in disjoint blocks, and an `A = 1`
// after both that binds to neither, so a rule matching on name reports twice on code upstream leaves
// alone.
func TestNoClassAssignStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a read of the class", "class A { } foo(A);"},
		{"a read of a variable holding a named class expression", "let A = class A { }; foo(A);"},
		{"a write to a parameter shadowing the class", "class A { b(A) { A = 0; } }"},
		{"a write to a local shadowing the class", "class A { b() { let A; A = 0; } }"},
		{"a write to the variable from inside an anonymous class expression",
			"let A = class { b() { A = 0; } }"},
		{"a write to the variable from inside a differently named class expression",
			"let A = class B { foo() { A = 0; } }"},
		{"a write to the variable holding a named class expression", "let A = class A {}; A = 1"},
		{"a write to a var", "var x = 0; x = 1;"},
		{"a write to a let", "let x = 0; x = 1;"},
		{"a write to a const", "const x = 0; x = 1;"},
		{"a write to a function declaration", "function x() {} x = 1;"},
		{"a write to a parameter", "function foo(x) { x = 1; }"},
		{"a write to a catch parameter", "try {} catch (x) { x = 1; }"},
		{"a write binding to neither of two sibling classes",
			"if (foo) { class A {} } else { class A {} } A = 1;"},
		// Sequence expression
		{"a write beside a class expression in a sequence", "(class A {}, A = 1)"},
		// Class expressions
		{"a write to a variable holding an anonymous class", "let A = class { }; A = 1;"},
		{"a write to a variable holding a differently named class", "let A = class B { }; A = 1;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoClassAssign, classAssignFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Upstream carries two labels per diagnostic, the declaration and the write, and our engine has one
// range per finding. The write is the one chosen: it is the line the reader has to change, and the
// declaration is usually correct code. A port pointing at the declaration passes every fixture
// above, because `ExpectFindings` asserts ids and count and nothing else.
//
// Sliced out of the source with the finding's own range rather than compared against an offset,
// since an offset computed by the test is wrong in the same direction as the code that produced it.
func TestNoClassAssignPointsAtTheWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a plain reassignment", "class A { } A = 0;", []string{"A"}},
		// A parenthesized assignment target, which neither upstream's corpus nor this one had until
		// a sibling rule tripped over it. `reference.WritesToBinding` climbs and reassigns `child` to
		// the parent, so at the binary node `child` was the parenthesis while the arm compared
		// against `ast.SkipParentheses(binary.Left)`, which is the identifier. The two never
		// matched and `(A) = 1` read as a read. The span assertion is the useful half here: it
		// pins the finding to the name rather than to the parentheses around it.
		{"a parenthesized target", "class A { } (A) = 1;", []string{"A"}},
		{"a hoisted reassignment", "A = 0; class A { }", []string{"A"}},
		{"a shorthand destructuring target", "class A { } ({A} = 0);", []string{"A"}},
		{"two writes point at their own sites", "class A { } A = 0; A = 1;", []string{"A", "A"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoClassAssign, classAssignFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.want), len(result.Diagnostics))
			}
			// The harness trims and appends a newline, so read positions against that same text.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			for i, want := range testCase.want {
				reported := source[result.Diagnostics[i].Range.Pos():result.Diagnostics[i].Range.End()]
				if reported != want {
					t.Errorf("finding %d pointed at %q, wanted %q", i, reported, want)
				}
			}
		})
	}
}

// The two writes must point at different places, which the fixture above cannot quite prove.
//
// `[]string{"A", "A"}` is satisfied by a rule reporting the same identifier twice, and a rule
// looping over the wrong collection does exactly that. This pins that they are distinct offsets.
func TestNoClassAssignReportsEachWriteAtItsOwnOffset(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoClassAssign, classAssignFile, "class A { } A = 0; A = 1;")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted 2 findings, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Fatalf("both findings pointed at offset %d, so one write went unreported",
			result.Diagnostics[0].Range.Pos())
	}
}

// The typed harness is load-bearing, and a revert to `rule_testing.Run` must fail loudly.
//
// This rule declares NeedsTypeChecker, so the plain harness hands it a nil checker and it goes
// completely silent. Every StaysSilent case above would then pass for the wrong reason, and the
// Fires cases would fail in a way that reads as a rule bug rather than as a harness mistake. This
// asserts the silence directly, so the vacuous configuration is a named, tested state rather than
// something a later edit can drift into unnoticed.
func TestNoClassAssignNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoClassAssign.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so the engine will not lock the file")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoClassAssign, classAssignFile, "class A { } A = 0;"))
}

// Cases written from reading our code rather than upstream's, each covering a write shape the
// imported corpus never exercises.
//
// Upstream gets these for free: `is_write()` is asked of a reference index that already classified
// every occurrence, so an array pattern and a for-of head cost it nothing to support. Ours has to
// name each shape, which means each is a place the port can be silently short, and the corpus
// cannot tell us because it contains no such case.
//
// The reads are here for the same reason from the other side. `A.x = 0` writes to a property of the
// class and leaves the binding alone, and it resolves to the class symbol exactly as a real
// reassignment does, so symbol identity alone reports it. That one is the reason this rule needs
// structural write detection as well as the checker.
func TestNoClassAssignCoversWriteShapesUpstreamNeverExercises(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a compound assignment", "class A { } A += 1;", []string{"noClassAssign"}},
		{"a logical assignment", "class A { } A ||= 1;", []string{"noClassAssign"}},
		{"a postfix increment", "class A { } A++;", []string{"noClassAssign"}},
		{"a prefix decrement", "class A { } --A;", []string{"noClassAssign"}},
		{"an array destructuring target", "class A { } [A] = [0];", []string{"noClassAssign"}},
		{"a rest element in an array pattern", "class A { } [...A] = [0];",
			[]string{"noClassAssign"}},
		{"a rest property in an object pattern", "class A { } ({...A} = {});",
			[]string{"noClassAssign"}},
		{"a shorthand target carrying a default", "class A { } ({A = 1} = {});",
			[]string{"noClassAssign"}},
		{"a for-of loop head", "class A { } for (A of []) {}", []string{"noClassAssign"}},
		{"a for-in loop head", "class A { } for (A in {}) {}", []string{"noClassAssign"}},
		{"a write from a static block", "class A { static { A = 0; } }",
			[]string{"noClassAssign"}},
		{"a write from a field initializer", "class A { p = (A = 0); }",
			[]string{"noClassAssign"}},
		{"a nested destructuring target", "class A { } ({b: {c: A}} = {b: {}});",
			[]string{"noClassAssign"}},
		// Two classes in one file, one of them written to. The finding must be one, not two: a
		// rule looping the file per class and matching on text alone reports the write once for
		// each anchor, and `wantIds` of length one is what catches that.
		{"one write with two classes in the file", "class A { } class B {} B = 0;",
			[]string{"noClassAssign"}},
		// A class shadowing another class of the same name, which is the case that separates node
		// identity from declaration kind. Both anchors are KindClassDeclaration, so a rule
		// comparing kinds instead of nodes calls the write a match for the outer class as well and
		// reports it twice. Upstream's corpus cannot catch that: its only same-kind pair is the
		// if/else one, whose write resolves to nothing at all and exits before the comparison.
		//
		// A mutant replacing the node comparison with a kind comparison survived the entire
		// upstream corpus and every invented case above before this was added.
		{"a block-scoped class shadowing an outer class of the same name",
			"class A {} { class A {} A = 1; }", []string{"noClassAssign"}},
		{"a function-scoped class shadowing an outer class of the same name",
			"class A {} function f() { class A {} A = 1; }", []string{"noClassAssign"}},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoClassAssign, classAssignFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		{"a property write through the class", "class A { } A.x = 0;"},
		{"a read on the right of an assignment", "class A { } let b; b = A;"},
		{"an extends clause naming the class", "class A { } class B extends A {}"},
		{"a same-named property of an unrelated object",
			"const o = { A: 1 }; class A {} o.A = 2;"},
		{"a shorthand property in an object being built rather than destructured",
			"class A { } const o = {A};"},
		{"a type-position reference", "class A { } let b: A;"},
		// The class on the left of a binary expression that is not an assignment. A mutant dropping
		// the assignment-operator gate treats every binary with the name on its left as a write,
		// and survives the whole corpus: upstream has no case reading the class in left position,
		// because its `is_write()` classified the reference and never had to ask about operators.
		{"a comparison with the class on the left", "class A { } if (A === 0) {}"},
		{"an arithmetic use with the class on the left", "class A { } let b = A + 1;"},
		// A prefix unary that reads rather than writes. Same gap from the other side: a mutant
		// dropping the update-operator gate calls `-A` and `!A` writes, and nothing upstream or in
		// the invented cases above exercises a non-update prefix operator on a class name.
		{"a negation of the class", "class A { } let b = !A;"},
		{"a unary minus on the class", "class A { } let b = -A;"},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoClassAssign, classAssignFile, testCase.sourceText))
		})
	}
}
