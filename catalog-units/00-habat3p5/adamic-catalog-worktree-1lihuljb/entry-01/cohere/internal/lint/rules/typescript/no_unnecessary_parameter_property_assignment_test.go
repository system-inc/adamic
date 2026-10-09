package typescript

import (
	"fmt"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// parameterPropertyFile is where the fixtures pretend to live.
const parameterPropertyFile = "/repository/source/ParameterProperty.ts"

// parameterPropertyCaseName names a case by its position in upstream's own list.
//
// The sources are multi-line class bodies, so a descriptive name would be a paraphrase that drifts
// from the case it names. The index is what lets a failure be looked up in the corpus directly.
func parameterPropertyCaseName(number int) string {
	return fmt.Sprintf("upstream case %02d", number)
}

// The corpus is oxc's, copied byte for byte rather than retyped.
//
// Every case below came out of the extractor's dump and was then compared byte against byte with
// the literals in oxc's own rule file by a script: 29 pass, 21 fail, 21 fix vectors, zero
// mismatches over all 50 rule cases. That script also asserted that no backslash appears anywhere
// in the corpus, which is what rules out an escape having been cooked on the way in.
//
// The snapshot records 25 diagnostics from those 21 failing inputs, so one finding per input is
// wrong here. The per-input counts below were recovered by walking the snapshot's location headers
// in input order: one input reports three times because it writes the same assignment three times,
// and two report twice because each nests a second class with its own parameter property.

// The clean cases are the whole discrimination and each one fails for a different reason.
//
// Two are worth naming, because a wrong port passes all the others without them. Case 15
// (`this.bar = () => { this.foo = foo; }`) is clean because the enclosing ASSIGNMENT is a dead end
// in upstream's visitor, NOT because arrows are skipped, and reading it the other way is the single
// most available wrong conclusion here. Case 23 (a block redeclaring `foo`) is the only case in the
// corpus that turns on resolving the identifier rather than matching its spelling.
func TestNoUnnecessaryParameterPropertyAssignmentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		number     int
		sourceText string
	}{
		{1, "\n        class Foo {\n          constructor(public name: unknown) {}\n        }\n        "},
		{2, "\n        class Foo {\n          constructor(public name: unknown, other: unknown) {\n            this.other = other;\n          }\n        }\n        "},
		{3, "\n        class Foo {\n          constructor(public name: unknown) {\n            this.other = name;\n          }\n        }\n        "},
		{4, "\n        class Foo {\n          constructor(name: unknown) {\n            this.name = name;\n          }\n        }\n        "},
		{5, "\n        class Foo {\n          constructor(public name: unknown) {\n            this.name = 'other';\n          }\n        }\n        "},
		{6, "\n        class Foo {\n          constructor(public name: unknown) {\n            this.name = name + 'edited';\n          }\n        }\n        "},
		{7, "\n        class Foo {\n          constructor(public name: unknown) {\n            if (maybeTrue) {\n              this.name = name + 'edited';\n            }\n          }\n        }\n        "},
		{8, "\n        class Foo {\n          constructor(foo: string) {}\n        }\n        "},
		{9, "\n        class Foo {\n          constructor(private foo: string) {}\n        }\n        "},
		{10, "\n        class Foo {\n          constructor(private foo: string) {\n            this.foo = bar;\n          }\n        }\n        "},
		{11, "\n        class Foo {\n          constructor(private foo: any) {\n            this.foo = foo.bar;\n          }\n        }\n        "},
		{12, "\n        class Foo {\n          constructor(private foo: string) {\n            this.foo = this.bar;\n          }\n        }\n        "},
		{13, "\n        class Foo {\n          foo: string;\n          constructor(foo: string) {\n            this.foo = foo;\n          }\n        }\n        "},
		{14, "\n        class Foo {\n          bar: string;\n          constructor(private foo: string) {\n            this.bar = foo;\n          }\n        }\n        "},
		{15, "\n        class Foo {\n          constructor(private foo: string) {\n            this.bar = () => {\n              this.foo = foo;\n            };\n          }\n        }\n        "},
		{16, "\n        class Foo {\n          constructor(private foo: string) {\n            this[`${foo}`] = foo;\n          }\n        }\n        "},
		{17, "\n        function Foo(foo) {\n          this.foo = foo;\n        }\n        "},
		{18, "\n        const foo = 'foo';\n        this.foo = foo;\n        "},
		{19, "\n        class Foo {\n          constructor(public foo: number) {\n            this.foo += foo;\n            this.foo -= foo;\n            this.foo *= foo;\n            this.foo /= foo;\n            this.foo %= foo;\n            this.foo **= foo;\n          }\n        }\n        "},
		{20, "\n        class Foo {\n          constructor(public foo: number) {\n            this.foo += 1;\n            this.foo = foo;\n          }\n        }\n        "},
		{21, "\n        class Foo {\n          constructor(\n            public foo: number,\n            bar: boolean,\n          ) {\n            if (bar) {\n              this.foo += 1;\n            } else {\n              this.foo = foo;\n            }\n          }\n        }\n        "},
		{22, "\n        class Foo {\n          constructor(public foo: number) {\n            this.foo = foo;\n          }\n          init = (this.foo += 1);\n        }\n        "},
		{23, "\n        class Foo {\n          constructor(public foo: number) {\n            {\n              const foo = 1;\n              this.foo = foo;\n            }\n          }\n        }\n        "},
		{24, "\n        declare const name: string;\n        class Foo {\n          constructor(public foo: number) {\n            this[name] = foo;\n          }\n        }\n        "},
		{25, "\n        declare const name: string;\n        class Foo {\n          constructor(public foo: number) {\n            Foo.foo = foo;\n          }\n        }\n        "},
		{26, "\n        class Foo {\n          constructor(public foo: number) {\n            this.foo = foo;\n          }\n          init = (() => {\n            this.foo += 1;\n          })();\n        }\n        "},
		{27, "\n        declare const name: string;\n        class Foo {\n          constructor(public foo: number) {\n            this[name] = foo;\n          }\n          init = (this[name] = 1);\n          init2 = (Foo.foo = 1);\n        }\n        "},
		{28, "\n        class Foo {\n          constructor(public foo: number) {\n            this.foo = foo;\n          }\n          init = (function() {\n            console.log('hi');\n            this.foo += 1;\n          })();\n        }\n        "},
		{29, "\n        class Foo {\n          constructor(private foo: string) {\n            function bar() {\n              this.foo = foo;\n            }\n          }\n        }\n        "},
	}

	for _, testCase := range cases {
		t.Run(parameterPropertyCaseName(testCase.number), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.sourceText))
		})
	}
}

// The reporting cases, with the per-input diagnostic count recovered from the snapshot.
func TestNoUnnecessaryParameterPropertyAssignmentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		number     int
		sourceText string
		findings   int
	}{
		{1, "\n        class Foo {\n          constructor(public name: unknown) {\n            this.name = name;\n          }\n        }\n        ", 1},
		{2, "\n        class Foo {\n          constructor(other: unknown, public name: unknown) {\n            this.other = other;\n            this.name = name;\n          }\n        }\n        ", 1},
		{3, "\n        class Foo {\n          constructor(public name: unknown) {\n            this.name = name;\n            this.name = name;\n            this.name = name;\n          }\n        }\n        ", 3},
		{4, "\n        class Foo {\n          constructor(public name: unknown) {\n            if (maybeTrue) {\n              this.name = name;\n            } else {\n              this.name = name + 'edited';\n            }\n          }\n        }\n        ", 1},
		{5, "\n        class Foo {\n          constructor(public foo: string) {\n            this.foo = foo;\n          }\n        }\n        ", 1},
		{6, "\n        class Foo {\n          constructor(public foo?: string) {\n            this.foo = foo!;\n          }\n        }\n        ", 1},
		{7, "\n        class Foo {\n          constructor(public foo?: string) {\n            this.foo = foo as any;\n          }\n        }\n        ", 1},
		{8, "\n        class Foo {\n          constructor(public foo = '') {\n            this.foo = foo;\n          }\n        }\n        ", 1},
		{9, "\n        class Foo {\n          constructor(public foo = '') {\n            this.foo = foo;\n            this.foo += 'foo';\n          }\n        }\n        ", 1},
		{10, "\n        class Foo {\n          constructor(public foo: string) {\n            this.foo ||= foo;\n          }\n        }\n        ", 1},
		{11, "\n        class Foo {\n          constructor(public foo: string) {\n            this.foo ??= foo;\n          }\n        }\n        ", 1},
		{12, "\n        class Foo {\n          constructor(public foo: string) {\n            this.foo &&= foo;\n          }\n        }\n        ", 1},
		{13, "\n        class Foo {\n          constructor(private foo: string) {\n            this['foo'] = foo;\n          }\n        }\n        ", 1},
		{14, "\n        class Foo {\n          constructor(private foo: string) {\n            function bar() {\n              this.foo = foo;\n            }\n            this.foo = foo;\n          }\n        }\n        ", 1},
		{15, "\n        class Foo {\n          constructor(private foo: string) {\n            this.bar = () => {\n              this.foo = foo;\n            };\n            this.foo = foo;\n          }\n        }\n        ", 1},
		{16, "\n        class Foo {\n          constructor(private foo: string) {\n            class Bar {\n              constructor(private foo: string) {\n                this.foo = foo;\n              }\n            }\n            this.foo = foo;\n          }\n        }\n        ", 2},
		{17, "\n        class Foo {\n          constructor(private foo: string) {\n            this.foo = foo;\n          }\n          bar = () => {\n            this.foo = 'foo';\n          };\n        }\n        ", 1},
		{18, "\n        class Foo {\n          constructor(private foo: string) {\n            this.foo = foo;\n          }\n          init = foo => {\n            this.foo = foo;\n          };\n        }\n        ", 1},
		{19, "\n        class Foo {\n          constructor(private foo: string) {\n            this.foo = foo;\n          }\n          init = class Bar {\n            constructor(private foo: string) {\n              this.foo = foo;\n            }\n          };\n        }\n        ", 2},
		{20, "\n        class Foo {\n          constructor(private foo: string) {\n            {\n              this.foo = foo;\n            }\n          }\n        }\n        ", 1},
		{21, "\n        class Foo {\n          constructor(private foo: string) {\n            (() => {\n              this.foo = foo;\n            })();\n          }\n        }\n        ", 1},
	}

	for _, testCase := range cases {
		t.Run(parameterPropertyCaseName(testCase.number), func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "unnecessaryAssign"
			}
			rule_testing.ExpectFindings(t, rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.sourceText),
				wantIds...)
		})
	}
}

// The repair, asserted by applying it rather than by comparing its text.
//
// A repair writing the right string over the wrong span passes a text comparison, so these apply
// every suggestion back to front and compare the resulting source with upstream's own after-text.
// Upstream ships these as SUGGESTIONS rather than fixes, which is why `ExpectFixedSource` cannot be
// used at all: it applies `Fixes`, and this rule proposes none. The applier is hand-rolled below.
func TestNoUnnecessaryParameterPropertyAssignmentSuggestsTheRepair(t *testing.T) {
	t.Parallel()

	cases := []struct {
		number int
		before string
		after  string
	}{
		{1, "\n            class Foo {\n              constructor(public name: unknown) {\n                this.name = name;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public name: unknown) {\n                \n              }\n            }\n            "},
		{2, "\n            class Foo {\n              constructor(other: unknown, public name: unknown) {\n                this.other = other;\n                this.name = name;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(other: unknown, public name: unknown) {\n                this.other = other;\n                \n              }\n            }\n            "},
		{3, "\n            class Foo {\n              constructor(public name: unknown) {\n                this.name = name;\n                this.name = name;\n                this.name = name;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public name: unknown) {\n                \n                \n                \n              }\n            }\n            "},
		{4, "\n            class Foo {\n              constructor(public name: unknown) {\n                if (maybeTrue) {\n                  this.name = name;\n                } else {\n                  this.name = name + 'edited';\n                }\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public name: unknown) {\n                if (maybeTrue) {\n                  \n                } else {\n                  this.name = name + 'edited';\n                }\n              }\n            }\n            "},
		{5, "\n            class Foo {\n              constructor(public foo: string) {\n                this.foo = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo: string) {\n                \n              }\n            }\n            "},
		{6, "\n            class Foo {\n              constructor(public foo?: string) {\n                this.foo = foo!;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo?: string) {\n                \n              }\n            }\n            "},
		{7, "\n            class Foo {\n              constructor(public foo?: string) {\n                this.foo = foo as any;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo?: string) {\n                \n              }\n            }\n            "},
		{8, "\n            class Foo {\n              constructor(public foo = '') {\n                this.foo = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo = '') {\n                \n              }\n            }\n            "},
		{9, "\n            class Foo {\n              constructor(public foo = '') {\n                this.foo = foo;\n                this.foo += 'foo';\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo = '') {\n                \n                this.foo += 'foo';\n              }\n            }\n            "},
		{10, "\n            class Foo {\n              constructor(public foo: string) {\n                this.foo ||= foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo: string) {\n                \n              }\n            }\n            "},
		{11, "\n            class Foo {\n              constructor(public foo: string) {\n                this.foo ??= foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo: string) {\n                \n              }\n            }\n            "},
		{12, "\n            class Foo {\n              constructor(public foo: string) {\n                this.foo &&= foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(public foo: string) {\n                \n              }\n            }\n            "},
		{13, "\n            class Foo {\n              constructor(private foo: string) {\n                this['foo'] = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                \n              }\n            }\n            "},
		{14, "\n            class Foo {\n              constructor(private foo: string) {\n                function bar() {\n                  this.foo = foo;\n                }\n                this.foo = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                function bar() {\n                  this.foo = foo;\n                }\n                \n              }\n            }\n            "},
		{15, "\n            class Foo {\n              constructor(private foo: string) {\n                this.bar = () => {\n                  this.foo = foo;\n                };\n                this.foo = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                this.bar = () => {\n                  this.foo = foo;\n                };\n                \n              }\n            }\n            "},
		{16, "\n            class Foo {\n              constructor(private foo: string) {\n                class Bar {\n                  constructor(private foo: string) {\n                    this.foo = foo;\n                  }\n                }\n                this.foo = foo;\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                class Bar {\n                  constructor(private foo: string) {\n                    \n                  }\n                }\n                \n              }\n            }\n            "},
		{17, "\n            class Foo {\n              constructor(private foo: string) {\n                this.foo = foo;\n              }\n              bar = () => {\n                this.foo = 'foo';\n              };\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                \n              }\n              bar = () => {\n                this.foo = 'foo';\n              };\n            }\n            "},
		{18, "\n            class Foo {\n              constructor(private foo: string) {\n                this.foo = foo;\n              }\n              init = foo => {\n                this.foo = foo;\n              };\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                \n              }\n              init = foo => {\n                this.foo = foo;\n              };\n            }\n            "},
		{19, "\n            class Foo {\n              constructor(private foo: string) {\n                this.foo = foo;\n              }\n              init = class Bar {\n                constructor(private foo: string) {\n                  this.foo = foo;\n                }\n              };\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                \n              }\n              init = class Bar {\n                constructor(private foo: string) {\n                  \n                }\n              };\n            }\n            "},
		{20, "\n            class Foo {\n              constructor(private foo: string) {\n                {\n                  this.foo = foo;\n                }\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                {\n                  \n                }\n              }\n            }\n            "},
		{21, "\n            class Foo {\n              constructor(private foo: string) {\n                (() => {\n                  this.foo = foo;\n                })();\n              }\n            }\n            ", "\n            class Foo {\n              constructor(private foo: string) {\n                (() => {\n                  \n                })();\n              }\n            }\n            "},
	}

	for _, testCase := range cases {
		t.Run(parameterPropertyCaseName(testCase.number), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.before)
			got := applyEveryParameterPropertySuggestion(t, testCase.before, result)
			if got != testCase.after {
				t.Fatalf("applying the suggestions gave\n%q\nwanted\n%q", got, testCase.after)
			}
		})
	}
}

// applyEveryParameterPropertySuggestion rewrites the source with every suggestion this rule offered.
//
// Applied back to front so that an earlier edit does not shift the offsets a later one was computed
// against. `rule_testing` has no suggestion applier, only `ExpectFixedSource` for fixes, so this exists
// rather than the assertion being skipped: a rule whose only repair is a suggestion would otherwise
// ship with nothing checking where the repair points.
func applyEveryParameterPropertySuggestion(t *testing.T, source string, result rule_testing.Result) string {
	t.Helper()

	type edit struct {
		start int
		end   int
		text  string
	}

	var edits []edit
	for _, diagnostic := range result.Diagnostics {
		if len(diagnostic.Suggestions) != 1 {
			t.Fatalf("wanted exactly one suggestion per finding, got %d", len(diagnostic.Suggestions))
		}
		for _, fix := range diagnostic.Suggestions[0].Fixes {
			edits = append(edits, edit{start: fix.Range.Pos(), end: fix.Range.End(), text: fix.Text})
		}
	}

	sort.Slice(edits, func(a, b int) bool { return edits[a].start > edits[b].start })

	for _, e := range edits {
		if e.start < 0 || e.end > len(source) || e.start > e.end {
			t.Fatalf("suggestion range [%d,%d) is outside the source of length %d", e.start, e.end, len(source))
		}
		source = source[:e.start] + e.text + source[e.end:]
	}

	return source
}

// A parameter property written as a destructuring pattern, which crashes the rule without a guard.
//
// TypeScript rejects this with error 1187 so it never appears in code that compiles, but the parser
// still produces the node with ModifierFlagsParameterPropertyModifier set, and `Node.Text()` panics
// on the resulting KindObjectBindingPattern. A mutation removing the kind guard survived every
// other fixture in this file, because no `ExpectFindings` assertion can observe a panic. This test
// exists to make that guard visible: it fails by crashing rather than by mismatching.
func TestNoUnnecessaryParameterPropertyAssignmentSurvivesADestructuredParameterProperty(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.Run(t,
		NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile,
		"class Foo {\n  constructor(public { a }: { a: string }) {\n    this.a = a;\n  }\n}\n"))
}

// The message text and id, asserted against literals typed here rather than against the rule's own
// constants.
//
// Comparing a diagnostic with the very constant it was reported from is equality that looks correct
// and moves in lockstep under mutation, so it can never fail. These are literals.
func TestNoUnnecessaryParameterPropertyAssignmentMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile,
		"class Foo {\n  constructor(public name: unknown) {\n    this.name = name;\n  }\n}\n")

	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}

	if got := result.Diagnostics[0].Message.Id; got != "unnecessaryAssign" {
		t.Fatalf("message id was %q", got)
	}
	wantDescription := "A constructor parameter carrying `public`, `private`, `protected`, " +
		"`readonly` or `override` is already assigned to the matching property before the " +
		"constructor body runs. Writing it again does nothing, and it invites the next reader to " +
		"believe the two names could diverge. Delete the assignment."
	if got := result.Diagnostics[0].Message.Description; got != wantDescription {
		t.Fatalf("message description was %q", got)
	}

	if got := result.Diagnostics[0].Suggestions[0].Message.Id; got != "removeAssignment" {
		t.Fatalf("suggestion id was %q", got)
	}
}

// The span, which the message-id fixtures above cannot see.
//
// Sliced out of the source with the finding's own range and compared with the text it should cover.
// Upstream's span is the whole assignment expression, which the second case pins in the one shape
// where that is not obvious: a parenthesized target puts the leading `(` inside the reported range.
func TestNoUnnecessaryParameterPropertyAssignmentSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"a plain assignment",
			"class Foo {\n  constructor(public name: unknown) {\n    this.name = name;\n  }\n}\n",
			"this.name = name",
		},
		{
			// oxc's parser drops the parentheses around an assignment target while ours keeps a
			// real node, so this shape needs a skip that upstream's source does not show. Measured
			// on the release binary: it reports, and the span covers the leading parenthesis.
			"a parenthesized target",
			"class Foo {\n  constructor(public name: unknown) {\n    (this.name) = name;\n  }\n}\n",
			"(this.name) = name",
		},
		{
			// The right-hand unwrap does not widen the span past the assignment.
			"a non-null assertion on the right",
			"class Foo {\n  constructor(public name?: unknown) {\n    this.name = name!;\n  }\n}\n",
			"this.name = name!",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Fatalf("the finding pointed at %q, wanted %q", reported, testCase.want)
			}
		})
	}
}

// Cases the imported corpus does not write, each measured on the release oxlint binary first.
//
// The corpus writes no parentheses anywhere and no `override` or `readonly` parameter with a body,
// so a port could fall either way on all of these and every imported fixture would stay green.
func TestNoUnnecessaryParameterPropertyAssignmentMeasuredAgainstUpstream(t *testing.T) {
	t.Parallel()

	reporting := []struct {
		name       string
		sourceText string
	}{
		{
			// `override` is in upstream's `has_modifier` and is NOT in
			// `ModifierFlagsAccessibilityModifier`, so a port reaching for the accessibility flag
			// alone goes silent here. Measured: reports.
			"an override parameter",
			"class Base { foo!: string; }\nclass Foo extends Base {\n  constructor(override foo: string) {\n    super();\n    this.foo = foo;\n  }\n}\n",
		},
		{
			// Same trap, other half. Measured: reports.
			"a readonly parameter",
			"class Foo {\n  constructor(readonly foo: string) {\n    this.foo = foo;\n  }\n}\n",
		},
		{
			// `get_inner_expression()` unwraps the parentheses on the right. Measured: reports.
			"a parenthesized right side",
			"class Foo {\n  constructor(public foo: string) {\n    this.foo = (foo);\n  }\n}\n",
		},
		{
			// A prior write to the PARAMETER does not rescue the assignment, because upstream's
			// bookkeeping records only `this.X` targets. The property and the parameter genuinely
			// differ here and upstream reports anyway. Measured: reports. Reproduced deliberately.
			"a parameter reassigned before the property assignment",
			"class Foo {\n  constructor(public foo: string) {\n    foo = foo.trim();\n    this.foo = foo;\n  }\n}\n",
		},
		{
			// The arrow case that the corpus's clean version hides. An arrow NOT under an
			// assignment is descended into, because arrows keep `this`. Measured: reports. A port
			// that skipped arrows passes every imported fixture and goes silent here.
			"an arrow passed to a call",
			"declare function call(f: () => void): void;\nclass Foo {\n  constructor(private foo: string) {\n    call(() => {\n      this.foo = foo;\n    });\n  }\n}\n",
		},
		{
			// The same, with no call at all, which isolates the assignment as the stopping
			// condition rather than the arrow. Measured: reports.
			"a bare arrow expression statement",
			"class Foo {\n  constructor(private foo: string) {\n    (() => { this.foo = foo; });\n  }\n}\n",
		},
		{
			// The parenthesized arrow expression body. Upstream unwraps parentheses on the
			// initializer and on the callee and NOWHERE ELSE, so this assignment is not recognized
			// by the class body scan and the constructor still reports. Measured: REPORTS, while
			// the identical form without the inner parentheses is silent. A port that helpfully
			// unwrapped here passes every imported fixture and goes silent on this.
			"a parenthesized assignment in an arrow expression body",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  init = (() => (this.foo = 1))();\n}\n",
		},
		{
			// The same divergence through the block arm. Measured: REPORTS.
			"a parenthesized assignment in an arrow block body",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  init = (() => { (this.foo = 1); })();\n}\n",
		},
		{
			// A sibling method assigning the same property does NOT suppress, because the class
			// body scan reads property definitions only. Measured: reports.
			"a sibling method assigning the same property",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  m() { this.foo = 2; }\n}\n",
		},
	}

	for _, testCase := range reporting {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.sourceText),
				"unnecessaryAssign")
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		{
			// The other half of the parenthesis measurement, and the direction a free "skip
			// parentheses everywhere" improvement gets wrong. Upstream matches the `this` object by
			// kind with no unwrapping. Measured on the release binary: SILENT.
			"a parenthesized this object",
			"class Foo {\n  constructor(public foo: string) {\n    (this).foo = foo;\n  }\n}\n",
		},
		{
			// The bare arrow expression body, which IS recognized. This is the other half of the
			// paren measurement above and the reason the divergence is about parentheses rather
			// than about expression bodies. Measured: silent.
			"a bare assignment in an arrow expression body",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  init = ((() => this.foo = 1))();\n}\n",
		},
		{
			// And the bare block body. Measured: silent.
			"a bare assignment in an arrow block body",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  init = (() => { this.foo = 1; })();\n}\n",
		},
		{
			// A property definition BELOW the constructor suppresses just as one above it does, so
			// the class body scan is position independent. Measured: silent.
			"a property initializer assigning the same property below the constructor",
			"class Foo {\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n  init = (this.foo = 1);\n}\n",
		},
		{
			// And above it. Measured: silent.
			"a property initializer assigning the same property above the constructor",
			"class Foo {\n  init = (this.foo = 1);\n  constructor(public foo: number) {\n    this.foo = foo;\n  }\n}\n",
		},
		{
			// An arrow parameter shadowing the parameter property. Upstream's symbol cross-check
			// declines this; the corpus writes no such case, and a mutation making arrow parameters
			// invisible to the shadow walk survived every other fixture here. Measured on the
			// release binary: SILENT, while the identical body with the arrow parameter renamed
			// reports. This is the only shape in the rule where an arrow both keeps `this` and
			// rebinds the name on the right.
			"an arrow parameter shadowing the parameter property",
			"declare function call(f: (foo: string) => void): void;\nclass Foo {\n  constructor(public foo: string) {\n    call((foo) => {\n      this.foo = foo;\n    });\n  }\n}\n",
		},
		{
			// An accessor holds a `Function` upstream, so the same override that stops a method
			// stops it. Measured on the release binary: SILENT. Nothing in the corpus writes an
			// accessor inside a constructor body.
			"an object literal getter inside the constructor",
			"class Foo {\n  constructor(private foo: string) {\n    const o = { get x() { this.foo = foo; return 1; } };\n  }\n}\n",
		},
		{
			// A decorated parameter carries a modifier list but is not a parameter property, and
			// `ModifierFlagsParameterPropertyModifier` declines it. Pinned because a port reaching
			// for "does this parameter have any modifier at all" reports here.
			"a decorated parameter",
			"declare const dec: any;\nclass Foo {\n  constructor(@dec foo: string) {\n    this.foo = foo;\n  }\n}\n",
		},
	}

	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t,
				NoUnnecessaryParameterPropertyAssignment, parameterPropertyFile, testCase.sourceText))
		})
	}
}
