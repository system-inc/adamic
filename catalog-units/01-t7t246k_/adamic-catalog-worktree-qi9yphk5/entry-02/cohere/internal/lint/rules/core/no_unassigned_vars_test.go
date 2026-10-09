package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unassignedVarsFile is where the fixtures pretend to live.
//
// A real path matters here because this rule reads the checker, so the harness builds an actual
// program and the file has to sit somewhere a tsconfig can reach.
const unassignedVarsFile = "/repository/source/UnassignedVars.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_unassigned_vars.rs`:
// one Tester block, 13 pass and 11 fail. The snapshot records 12 diagnostics from those 11 inputs,
// so one finding per input is wrong here, and the extractor said so before any code was written.
// The extra diagnostic belongs to `let x; let a = x, b; log(x, a, b);`, which declares two
// never-assigned readable bindings and reports twice.
func TestNoUnassignedVarsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		// The discrepancy case. Two unassigned readable declarations, two findings. `a` has an
		// initializer so it is exempt, which is what makes the count two rather than three.
		{"two unassigned declarations in one statement pair", "let x; let a = x, b; log(x, a, b);",
			[]string{"noUnassignedVars", "noUnassignedVars"}},
		{"a comparison that never assigns", "const foo = (two) => { let one; if (one === two) {} }",
			[]string{"noUnassignedVars"}},
		{"passed to a call and never assigned", "let user; greet(user);",
			[]string{"noUnassignedVars"}},
		{"read through a logical fallback", "function test() { let error; return error || 'Unknown error'; }",
			[]string{"noUnassignedVars"}},
		{"destructured out of a never-assigned binding", "let options; const { debug } = options || {};",
			[]string{"noUnassignedVars"}},
		{"read as a loop condition", "let flag; while (!flag) { }",
			[]string{"noUnassignedVars"}},
		{"read from inside a nested function", "let config; function init() { return config?.enabled; }",
			[]string{"noUnassignedVars"}},
		{"a type annotation is not an assignment", "let x: number; log(x);",
			[]string{"noUnassignedVars"}},
		{"an optional type annotation is not an assignment", "let x: number | undefined; log(x);",
			[]string{"noUnassignedVars"}},
		{"an annotated local inside an arrow that only compares",
			"const foo = (two: string): void => { let one: string | undefined; if (one === two) {} }",
			[]string{"noUnassignedVars"}},
		// The ambient declaration inside `declare module` is exempt; the one outside it is not.
		// A rule exempting the whole file once it sees an ambient block reports zero here.
		{"an ambient module beside a real declaration", `
                            declare module 'module' {
                                let x: string;
                            }
                            let y: string;
                            console.log(y);
                        `, []string{"noUnassignedVars"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The clean cases, verbatim from the same block.
//
// Two groups do the real work. The first is the never-read group: `let x;` on its own is not
// reported, because a binding nobody reads is `no-unused-vars`' finding rather than this one's, and
// reporting it here would double up on every unused local in the tree. The second is the
// assigned-somewhere group, where a write exists and the rule has to find it: upstream's
// `one = two` inside an `if` is the only shape its corpus covers, which is why the invented cases
// below enumerate the rest.
func TestNoUnassignedVarsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"declared and never read", "let x;"},
		{"a var declared and never read", "var x;"},
		{"a const initialized to undefined", "const x = undefined; log(x);"},
		{"a let initialized to undefined", "let y = undefined; log(y);"},
		{"a var initialized to undefined", "var y = undefined; log(y);"},
		{"two initialized lets", "let a = x, b = y; log(a, b);"},
		{"two initialized vars", "var a = x, b = y; log(a, b);"},
		{"assigned inside a conditional", "const foo = (two) => { let one; if (one !== two) one = two; }"},
		{"an annotated let initialized to undefined", "let z: number | undefined = undefined; log(z);"},
		{"an ambient declaration", "declare let c: string | undefined; log(c);"},
		{"an annotated local assigned inside a block", `
                        const foo = (two: string): void => {
                            let one: string | undefined;
                            if (one !== two) {
                                one = two;
                            }
                        }
                    `},
		{"a declaration inside an ambient module", `
                        declare module 'module' {
                            import type { T } from 'module';
                            let x: T;
                            export = x;
                        }
                    `},
		{"a for-of loop variable", "for (let p of pathToRemove) { p.remove() }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText))
		})
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Upstream labels the declared identifier rather than the whole declarator, and that is the choice
// reproduced here: the name is what the reader searches for, and a span covering `x: number` would
// swallow the annotation into a finding that is not about the annotation. ESLint reports the whole
// declarator node, so this is a deliberate divergence toward the tighter of the two.
//
// Sliced out of the source with the finding's own range rather than compared against an offset,
// since an offset computed by the test is wrong in the same direction as the code that produced it.
func TestNoUnassignedVarsPointsAtTheDeclaredName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       []string
	}{
		{"a bare declaration", "let user; greet(user);", []string{"user"}},
		{"an annotated declaration points at the name and not the type",
			"let config: number; log(config);", []string{"config"}},
		{"two declarations point at their own names", "let x; let a = x, b; log(x, a, b);",
			[]string{"x", "b"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText)
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

// The two findings must point at different places, which the fixture above cannot quite prove.
//
// `[]string{"x", "b"}` would also be satisfied by a rule that happened to order them that way while
// reporting one of them at the other's offset. This pins that they are distinct.
func TestNoUnassignedVarsReportsEachDeclarationAtItsOwnOffset(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile,
		"let x; let a = x, b; log(x, a, b);")
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted 2 findings, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
		t.Fatalf("both findings pointed at offset %d, so one declaration went unreported",
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
func TestNoUnassignedVarsNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoUnassignedVars.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so the engine will not lock the file")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnassignedVars, unassignedVarsFile, "let user; greet(user);"))
}

// Every write shape, each one a way this rule could report a variable that is in fact assigned.
//
// This is the half of the rule upstream gets for free and we do not. oxc asks `is_write()` of a
// reference index its semantic layer already built, so an array pattern, a for-in head and a
// postfix `++` all cost it nothing and none of them appear in its corpus. Ours has to recognize
// each shape structurally, and a shape it fails to recognize is a variable reported as
// never-assigned while a line three characters away assigns it. That is a false positive on correct
// code, which is the worse direction for this rule specifically: a missed write here does not
// quietly under-report, it actively accuses.
//
// The sibling rule `no-ex-assign` shipped for weeks missing `x++` entirely because neither corpus
// tested it. Each arm below exists so that gap cannot recur silently here.
func TestNoUnassignedVarsFindsEveryWriteShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a plain assignment", "let x; x = 1; log(x);"},
		{"a compound assignment", "let x; x += 1; log(x);"},
		{"a logical assignment", "let x; x ??= 1; log(x);"},
		{"a postfix update", "let x; x++; log(x);"},
		{"a prefix update", "let x; ++x; log(x);"},
		{"a postfix decrement", "let x; x--; log(x);"},
		{"a prefix decrement", "let x; --x; log(x);"},
		{"an array destructuring target", "let x; [x] = a; log(x);"},
		{"a shorthand object destructuring target", "let x; ({x} = o); log(x);"},
		{"a renamed object destructuring target", "let x; ({a: x} = o); log(x);"},
		{"a destructuring target with a default", "let x; ({a: x = 1} = o); log(x);"},
		{"a rest destructuring target", "let x; ({...x} = o); log(x);"},
		{"a for-of loop target that is not a declaration", "let x; for (x of a) {} log(x);"},
		{"a for-in loop target that is not a declaration", "let x; for (x in o) {} log(x);"},
		{"an assignment through parentheses", "let x; (x) = 1; log(x);"},
		{"a nested array destructuring target", "let x; [[x]] = a; log(x);"},
		// A write inside a nested function still writes the binding. oxc's reference index does not
		// care where the reference sits, and neither does this: the walk covers the whole file.
		{"a write from inside a nested function", "let x; function g() { x = 1; } log(x);"},
		{"a write from inside a nested arrow", "let x; const g = () => { x = 1; }; log(x);"},
		// A write textually before the declaration. `let` is in its temporal dead zone here so this
		// throws at runtime, but it is still a write and upstream's index counts it, so reporting
		// the declaration as never-assigned would be wrong about the code as written.
		{"a write inside a function hoisted above the declaration",
			"function g() { x = 1; } let x; log(x);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText))
		})
	}
}

// The shapes that look like writes and are not, which is the other direction the write detector can
// be wrong in.
//
// Every case here declares a binding that is read and never assigned, so the rule must still report
// despite there being an identifier of the same name in a position that resembles a write. A
// detector that is too eager goes silent on real findings, which is the quiet failure: nothing in
// the corpus notices a rule that stopped firing.
func TestNoUnassignedVarsDoesNotMistakeReadsForWrites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a property write is not a binding write", "let x; x.y = 1; log(x);"},
		{"an element write is not a binding write", "let x; x[0] = 1; log(x);"},
		{"the right side of an assignment is a read", "let x; let b; b = x; log(x, b);"},
		{"a call argument is a read", "let x; foo(x); log(x);"},
		{"a key in a destructuring pattern is not the target",
			"let x; let out; ({x: out} = o); log(x, out);"},
		{"a negation is a read and not an update", "let x; log(!x);"},
		{"a unary minus is a read and not an update", "let x; log(-x);"},
		{"a for-of head that declares its own binding does not write the outer one",
			"let x; for (const x of a) { log(x); } log(x);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText),
				"noUnassignedVars")
		})
	}
}

// Shadowing, which is why this rule reads the checker rather than matching names.
//
// A write to a different binding spelled the same way is not a write to this one, and a rule
// matching on text alone cannot tell them apart. The direction of the mistake decides how much this
// matters: text matching sees the inner write, believes the outer binding is assigned, and goes
// silent. That is a missed finding rather than a false accusation, so it is the gentler failure,
// but it is also invisible, and it is most of what this rule would miss in real code where the same
// short name is reused in two scopes.
//
// The reverse case is the one that accuses. `{ let x; log(x); }` inside a file whose outer `x` is
// assigned must report the inner binding, and a rule anchoring on the wrong declaration reports
// nothing or reports the wrong one.
func TestNoUnassignedVarsSeparatesShadowedBindings(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"a parameter of the same name absorbs the write",
			"let x; function f(x) { x = 1; } log(x);", []string{"noUnassignedVars"}},
		{"a block-scoped local of the same name absorbs the write",
			"let x; { let x = 0; x = 1; log(x); } log(x);", []string{"noUnassignedVars"}},
		// The nested-redeclaration case named in the porting brief, written for this declaration
		// type. Both anchors are `let` bindings, so a port comparing declaration kind rather than
		// node identity calls the inner declaration a match for the outer one, decides the outer is
		// assigned, and reports only the inner. Identity reports the outer, which is the one that is
		// genuinely never assigned.
		{"two same-kind anchors where only the inner one is written",
			"let x; log(x); { let x; x = 1; log(x); }", []string{"noUnassignedVars"}},
		// Both bindings unassigned and both read. Two anchors, two findings, and a rule that
		// confuses them reports one.
		{"two shadowed bindings both unassigned", "let x; log(x); { let x; log(x); }",
			[]string{"noUnassignedVars", "noUnassignedVars"}},
	}

	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText),
				testCase.wantIds...)
		})
	}
}

// The nested-redeclaration case again, this time asserting which binding was named.
//
// The fixture above pins the count at one, and a port comparing declaration kind also produces one
// finding here. It just produces the wrong one: it reports the inner `x`, which is assigned, and
// stays quiet about the outer `x`, which is not. Only the offset separates the two behaviors, so
// the count assertion alone would ship the kind comparison.
func TestNoUnassignedVarsNamesTheOuterBindingWhenOnlyTheInnerIsWritten(t *testing.T) {
	t.Parallel()

	const sourceText = "let x; log(x); { let x; x = 1; log(x); }"
	result := rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted 1 finding, got %d", len(result.Diagnostics))
	}
	// The outer `x` is the fifth byte of the file. The inner one sits past `log(x); { let `.
	if position := result.Diagnostics[0].Range.Pos(); position != 4 {
		t.Fatalf("reported the binding at offset %d, wanted the outer one at 4; "+
			"reporting the inner binding means the rule matched on declaration kind rather than identity",
			position)
	}
}

// Cases written from reading our own tree rather than upstream's corpus.
//
// Two families. The first is the declaration forms upstream's exemptions imply but never state:
// a binding pattern has no single name to report and `Name().Text()` panics outright on one, so
// that is a crash rather than a wrong finding. The second is the ambient and namespace surface,
// where oxc exempts any enclosing module declaration whether or not it carries `declare`, and our
// checker's ambient flag alone would not.
func TestNoUnassignedVarsExemptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// `Name()` panics on a binding pattern, so this is the difference between a clean run and a
		// crashed one on any file destructuring without an initializer. It cannot occur in valid
		// TypeScript, since a pattern requires an initializer, but the rule sees files mid-edit.
		{"an object binding pattern", "let {a, b} = o; log(a, b);"},
		{"an array binding pattern", "let [a, b] = o; log(a, b);"},
		{"a const is exempt even when never assigned again", "const x = 1; log(x);"},
		{"a using declaration is const-like", "using x = r(); log(x);"},
		// A catch parameter is bound by the runtime, so no assignment to it exists in the source.
		// That is the exact shape this rule reports, and it produced 445 findings on our own tree
		// with a 100% false positive rate before this case existed. Upstream never faces it: oxc
		// requires a declarator whose parent is a VariableDeclaration, and a catch parameter fails
		// that parent check. TypeScript spells the declarator KindVariableDeclaration, so the port
		// collapsed both anchors into one and lost the exemption with it.
		{"a catch parameter", "try { risky(); } catch(error) { log(error); }"},
		{"a typed catch parameter", "try { risky(); } catch(error: unknown) { log(error); }"},
		{"a destructured catch parameter", "try { risky(); } catch({message}) { log(message); }"},

		// The for-head exemptions. A loop variable is assigned by the loop itself on every
		// iteration, so it is never the finding even though no initializer is written.
		{"a for-in loop variable", "for (let k in o) { log(k) }"},
		{"a bare for-statement loop variable", "for (let i; ; ) { log(i) }"},
		// oxc exempts any namespace declaration, not only a `declare` one. Our checker's ambient
		// flag is not set on a plain `namespace`, so an ambient-flag-only exemption reports here
		// and diverges. Reproduced rather than improved on, per the porting brief.
		{"a plain namespace, which oxc exempts without requiring declare",
			"namespace N { let x: string; log(x); }"},
		{"a nested plain namespace", "namespace N { namespace M { let x: string; log(x); } }"},
		{"a declare global block", "declare global { let g: string; }"},
		{"a declare namespace", "declare namespace N { let x: string; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText))
		})
	}
}

// A read is required, and this pins the boundary rather than trusting the corpus's `let x;`.
//
// Upstream stops on a binding nobody reads and says why in a comment: `no-unused-vars` owns that
// finding. Dropping the read requirement would report every declared-then-unused local in the tree,
// which is a large and entirely duplicate finding set. The corpus covers it with `let x;` alone, at
// the top level of a file, which is also the shape a rule could decline for unrelated reasons.
func TestNoUnassignedVarsRequiresARead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"never read inside a function", "function f() { let x; }"},
		{"never read inside a block", "{ let x; }"},
		{"never read while a same-named binding elsewhere is read",
			"function f() { let x; } function g() { let x; log(x); }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnassignedVars, unassignedVarsFile, testCase.sourceText)
			// The third case has one genuine finding; the first two have none. Asserting the count
			// against the number of readable declarations keeps this one table.
			wanted := strings.Count(testCase.sourceText, "log(x)")
			if len(result.Diagnostics) != wanted {
				t.Fatalf("wanted %d findings, got %d", wanted, len(result.Diagnostics))
			}
		})
	}
}
