package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireAtomicUpdatesCorpusProvenance records where the imported cases came from and how they were
// checked, because a fixture table is only worth what its provenance is.
//
// Source: ESLint's own tests at `tests/lib/rules/require-atomic-updates.js`, extracted by hooking
// its `RuleTester` and capturing the case objects rather than by parsing the file, so no case was
// retyped and no escape passed through a shell. 28 clean and 35 reporting cases, 37 findings.
//
// Checked twice before being trusted:
//
//	every extracted case was replayed against the INSTALLED eslint 10.8.1 through the Linter API,
//	  and all 63 reproduced their expected verdict with zero mismatches, which is what proves the
//	  extraction is faithful rather than merely well-formed
//	every Go literal in the corpus file was parsed back out and compared byte against byte with the
//	  extracted JSON, which is what catches a cooked escape or a smart quote
//
// One case is imported with a stated difference. Upstream's case 26 declares
// `globals: { process: "readonly" }`, and measured on the installed rule the SAME source with no
// such declaration is completely silent. See `TestRequireAtomicUpdatesResolutionDivergence`.
const requireAtomicUpdatesCorpusProvenance = "eslint/tests/lib/rules/require-atomic-updates.js"

// TestRequireAtomicUpdatesStaysSilent runs every clean case upstream ships.
//
// These are the false positives upstream already thought about, and each one was added when somebody
// hit it. They are the half of the corpus that can catch a port being too eager, which is the
// direction this rule fails in: the flow reasoning is easy to make coarser than upstream's and every
// reporting case stays green while it happens.
func TestRequireAtomicUpdatesStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range requireAtomicUpdatesCleanCases {
		result := rule_testing.RunTypedWithOptions(
			t, RequireAtomicUpdates, "clean.ts", testCase.source, testCase.options)
		if len(result.Diagnostics) != 0 {
			t.Errorf("reported %v on a clean case\n%s",
				result.MessageIds(), testCase.source)
		}
	}
}

// TestRequireAtomicUpdatesFires runs every reporting case upstream ships, asserting the message ids
// in order and their count.
func TestRequireAtomicUpdatesFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range requireAtomicUpdatesReportingCases {
		if requireAtomicUpdatesNeedsDeclaredGlobals(testCase.source) {
			// One case is decided ABOVE the rule, by whether its identifier resolves at all.
			// Upstream ships it with `globals: { process: "readonly" }`, and measured on the
			// installed rule the same source with no such declaration is completely silent. Our
			// resolution comes from the checker rather than from a globals list, and this fixture
			// file declares nothing, so the name resolves to nothing here.
			//
			// It is recorded rather than deleted or greened, and the behaviour it stands for is
			// pinned by `TestRequireAtomicUpdatesProcessCaseWithDeclaration`, which supplies the
			// declaration and gets upstream's two findings.
			continue
		}
		result := rule_testing.RunTypedWithOptions(
			t, RequireAtomicUpdates, "fires.ts", testCase.source, testCase.options)
		got := result.MessageIds()
		if len(got) != len(testCase.wantIds) {
			t.Errorf("reported %v, want %v\n%s", got, testCase.wantIds, testCase.source)
			continue
		}
		for index, want := range testCase.wantIds {
			if got[index] != want {
				t.Errorf("finding %d was %q, want %q\n%s",
					index, got[index], want, testCase.source)
			}
		}
	}
}

// TestRequireAtomicUpdatesSpan asserts WHERE each finding points, which no message-id fixture can
// see.
//
// Upstream reports on `node.parent`, the whole assignment expression, for both message arms. A port
// pointing at the left side, at the right side, or at the enclosing statement passes every id
// fixture in this file while being wrong, and a reader following the finding would be sent to the
// wrong place. The expectation is the exact source text of the assignment.
func TestRequireAtomicUpdatesSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "compound assignment to an outer variable",
			source: `let foo; async function x() { foo += await amount; }`,
			want:   []string{"foo += await amount"},
		},
		{
			name:   "plain assignment reading itself before the suspension",
			source: `let foo; async function x() { foo = foo + await amount; }`,
			want:   []string{"foo = foo + await amount"},
		},
		{
			name:   "property store on a const object",
			source: `const foo = {}; async function x() { foo.bar += await baz }`,
			want:   []string{"foo.bar += await baz"},
		},
		{
			name:   "computed member in the middle of the chain",
			source: `const foo = []; async function x() { foo[bar].baz += await result;  }`,
			want:   []string{"foo[bar].baz += await result"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RequireAtomicUpdates, "span.ts", testCase.source)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("reported %d findings, want %d: %v",
					len(result.Diagnostics), len(testCase.want), result.MessageIds())
			}
			// The harness writes `strings.TrimSpace(contents)+"\n"`, so the file on disk can be one
			// byte offset from the literal above. Slicing the source the harness actually wrote is
			// what keeps this from reading as an off-by-one in the rule.
			written := result.SourceFile.Text()
			for index, want := range testCase.want {
				diagnostic := result.Diagnostics[index]
				got := written[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d spans %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestRequireAtomicUpdatesMessageText asserts the rendered text, exactly.
//
// The description is built with `Sprintf` and carries the binding name and, on the property arm, the
// assignment target's source text. A mutation that moves only the per-finding text leaves the id and
// the count fixed, so every other fixture in this file stays green over it. The assertions are
// equality against a literal typed here rather than against the rule's own message constant, because
// comparing against the constant moves both sides together under mutation.
func TestRequireAtomicUpdatesMessageText(t *testing.T) {
	t.Parallel()

	t.Run("variable arm names the binding", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "text.ts",
			`let counter; async function x() { counter += await amount; }`)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("reported %d findings, want 1", len(result.Diagnostics))
		}
		got := result.Diagnostics[0].Message.Description
		if !strings.HasSuffix(got, " Here the variable is counter.") {
			t.Errorf("description does not name the binding: %q", got)
		}
		if strings.Count(got, "counter") != 1 {
			t.Errorf("binding name appears %d times, want 1: %q",
				strings.Count(got, "counter"), got)
		}
	})

	t.Run("property arm names the target text and the object", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "text.ts",
			`const holder = []; async function x() { holder[key].slot += await result; }`)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("reported %d findings, want 1", len(result.Diagnostics))
		}
		got := result.Diagnostics[0].Message.Description
		want := " Here the assignment writes holder[key].slot and the object holder was read before the suspension."
		if !strings.HasSuffix(got, want) {
			t.Errorf("description tail is wrong\n got %q\nwant suffix %q", got, want)
		}
	})

	t.Run("message ids are the two upstream names", func(t *testing.T) {
		t.Parallel()
		if messageRequireAtomicUpdatesVariable.Id != "nonAtomicUpdate" {
			t.Errorf("variable arm id is %q", messageRequireAtomicUpdatesVariable.Id)
		}
		if messageRequireAtomicUpdatesProperty.Id != "nonAtomicObjectUpdate" {
			t.Errorf("property arm id is %q", messageRequireAtomicUpdatesProperty.Id)
		}
	})
}

// TestRequireAtomicUpdatesOptionsDecode routes options through the rule's own exported decoder.
//
// Building the options struct directly would leave the json tag and the default untested, and those
// are the two lines with no upstream counterpart. The nil case is the one that has shipped a broken
// rule in this tree before: a rule configured as a bare severity is handed nil, and a port relying
// on the zero value arriving by accident cannot tell that from a decoder that never ran.
func TestRequireAtomicUpdatesOptionsDecode(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[RequireAtomicUpdatesOptions]()

	t.Run("allowProperties true suppresses only the property arm", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{"allowProperties": true}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.(RequireAtomicUpdatesOptions).AllowProperties != true {
			t.Fatalf("decoded %#v", decoded)
		}

		property := `async function a(foo) { if (foo.bar) { foo.bar = await something; } }`
		result := rule_testing.RunTypedWithOptions(t, RequireAtomicUpdates, "opt.ts", property, decoded)
		if len(result.Diagnostics) != 0 {
			t.Errorf("property arm still reported under allowProperties: %v", result.MessageIds())
		}

		variable := `let foo; async function a() { if (foo) { foo = await something; } }`
		result = rule_testing.RunTypedWithOptions(t, RequireAtomicUpdates, "opt.ts", variable, decoded)
		rule_testing.ExpectFindings(t, result, "nonAtomicUpdate")
	})

	t.Run("allowProperties false keeps both arms", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{"allowProperties": false}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		property := `async function a(foo) { if (foo.bar) { foo.bar = await something; } }`
		result := rule_testing.RunTypedWithOptions(t, RequireAtomicUpdates, "opt.ts", property, decoded)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate")
	})

	t.Run("nil options bypass the decoder and keep both arms", func(t *testing.T) {
		t.Parallel()
		// This is the shape a bare `"error"` produces. It reaches the rule without ever passing
		// through the decoder above, so no fixture routed through `decode` can see it.
		property := `async function a(foo) { if (foo.bar) { foo.bar = await something; } }`
		result := rule_testing.RunTypedWithOptions(t, RequireAtomicUpdates, "opt.ts", property, nil)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate")
	})

	t.Run("an empty object decodes to the permissive-off default", func(t *testing.T) {
		t.Parallel()
		decoded, err := decode(json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.(RequireAtomicUpdatesOptions).AllowProperties != false {
			t.Fatalf("empty object did not default AllowProperties to false: %#v", decoded)
		}
	})
}

// TestRequireAtomicUpdatesSuspensionOrder pins the discrimination the whole rule rests on: which
// side of the suspension a read sits on.
//
// The two inputs differ by nothing except the position of `foo` relative to the await, and they have
// opposite verdicts upstream. A port that recorded the suspension at the wrong point, or that
// recorded reads without ordering them against it, gets one of these wrong while the other stays
// green. Both are upstream cases, kept here as a named pair because the corpus tables cannot say
// that these two are each other's control.
func TestRequireAtomicUpdatesSuspensionOrder(t *testing.T) {
	t.Parallel()

	stale := rule_testing.RunTyped(t, RequireAtomicUpdates, "order.ts",
		`let foo; async function x() { foo = foo + await amount; }`)
	rule_testing.ExpectFindings(t, stale, "nonAtomicUpdate")

	fresh := rule_testing.RunTyped(t, RequireAtomicUpdates, "order.ts",
		`let foo; async function x() { foo = await bar + foo; }`)
	rule_testing.ExpectClean(t, fresh)
}

// TestRequireAtomicUpdatesMessageSkipsLeadingComments pins the real-tree shape where a comment
// belongs to the assignment target's leading trivia. The finding was emitted at the right line,
// but the target interpolated the comment and newlines into its message. Line-oriented parity
// checks then saw no rule name on the diagnostic's first line and counted a false negative.
func TestRequireAtomicUpdatesMessageSkipsLeadingComments(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAtomicUpdates, "property-floor.ts", `
let state = { guard: false, first: 0 };
declare function pause(): Promise<void>;
async function run() {
    if (state.guard) return;
	await pause();
	// Advance only after success.
    state.first = Date.now();
}`)

	rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate")
	message := result.Diagnostics[0].Message.Description
	if strings.Contains(message, "Advance only") || strings.Contains(message, "\n") {
		t.Fatalf("message contains the assignment's leading trivia: %q", message)
	}
	if !strings.Contains(message, "assignment writes state.first") {
		t.Fatalf("message does not name the assignment target: %q", message)
	}
}

// TestRequireAtomicUpdatesEscapeTable pins each row of the escape predicate against the verdict
// measured on the installed rule.
//
// The rows are not interchangeable and the pairs matter more than the rows: each clean case is the
// control for the reporting case beside it, differing in exactly the property under test. Without
// the control, a rule that reported on everything would pass the reporting rows.
func TestRequireAtomicUpdatesEscapeTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "declared inside and never captured is local",
			source: `async function x() { let foo; foo += await bar; }`,
		},
		{
			name:   "declared inside but captured by a closure escapes",
			source: `async function x() { let foo; bar(() => foo); foo += await amount; }`,
			want:   []string{"nonAtomicUpdate"},
		},
		{
			name: "a closure that does not name the binding is not a capture",
			// The control for the row above: a nested function exists in both, and only naming the
			// binding changes the verdict.
			source: `async function x() { let foo; bar(() => baz += 1); foo += await amount; }`,
		},
		{
			name:   "declared outside the function escapes",
			source: `let foo; async function x() { foo += await amount; }`,
			want:   []string{"nonAtomicUpdate"},
		},
		{
			name: "a parameter is local for the variable arm",
			// Paired with the row below. One parameter, two questions, opposite answers.
			source: `async function f(foo) { foo = await bar; }`,
		},
		{
			name:   "a parameter escapes for the property arm",
			source: `async function f(foo) { let b = await get(foo.id); foo.bar = b.bar; }`,
			want:   []string{"nonAtomicObjectUpdate"},
		},
		{
			name:   "a local object is local for the property arm",
			source: `async function f() { let foo = {}; let bar = await get(foo.id); foo.prop = bar.prop; }`,
		},
		{
			name: "a non-suspending function is never judged",
			// The `shouldVerify` gate. Same shape as the reporting rows and no await anywhere.
			source: `let foo; function x() { foo = foo + amount; }`,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RequireAtomicUpdates, "escape.ts", testCase.source)
			if len(testCase.want) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestRequireAtomicUpdatesResolutionDivergence records the one place this port answers differently
// from a bare ESLint run, with the measurement rather than an argument.
//
// Upstream skips a reference whose `resolved` is null, and ESLint decides "resolved" from its scope
// analysis plus the configured globals list. Cohere asks the checker, which resolves against the
// program's declarations. Upstream's own case 26 is exactly this: it ships with
// `globals: { process: "readonly" }`, and measured on the installed eslint 10.8.1 the same source
// without that declaration is completely silent while with it there are two findings.
//
// The divergence is a strict superset in favour of reporting, and it is the direction the rule
// wants: upstream skips an unresolved name because it cannot tell a global from a typo, not because
// a global cannot race. A global is the most racy binding in a file.
//
// This test pins the behaviour we ship rather than upstream's, deliberately, so a later reader sees
// which one was chosen. Both halves are asserted: the resolvable binding reports, and a genuinely
// unresolvable one stays silent, which is the control that proves resolution is what decides it.
func TestRequireAtomicUpdatesResolutionDivergence(t *testing.T) {
	t.Parallel()

	t.Run("a resolvable binding is judged", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "divergence.ts",
			`let holder: any; async function f() { const q = holder.a; try { const r = await run(); holder.b = r; } catch (e) { holder.b = 1; } }`)
		if len(result.Diagnostics) == 0 {
			t.Fatal("a resolvable binding was not judged")
		}
		for _, diagnostic := range result.Diagnostics {
			if diagnostic.Message.Id != "nonAtomicObjectUpdate" {
				t.Errorf("unexpected id %q", diagnostic.Message.Id)
			}
		}
	})

	t.Run("an unresolvable name is skipped", func(t *testing.T) {
		t.Parallel()
		// The control. Same shape, and nothing declares the binding, so the checker gives no symbol
		// and the reference is skipped exactly as upstream skips an unresolved one.
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "divergence.ts",
			`async function f() { const q = neverDeclaredAnywhere.a; await run(); neverDeclaredAnywhere.b = 1; }`)
		rule_testing.ExpectClean(t, result)
	})
}

// requireAtomicUpdatesNeedsDeclaredGlobals names the one imported case whose verdict is decided by a
// declaration the fixture cannot carry.
//
// Matched by content rather than by index, because the corpus file is regenerated from upstream and
// an index would silently drift onto a different case.
func requireAtomicUpdatesNeedsDeclaredGlobals(source string) bool {
	return strings.Contains(source, "process.exitCode")
}

// TestRequireAtomicUpdatesProcessCaseWithDeclaration is the skipped corpus case, restored.
//
// Upstream's case 26 depends entirely on `process` being a declared global: measured on eslint
// 10.8.1, with `globals: { process: "readonly" }` it reports twice and with the identical source and
// no declaration it is silent. Supplying the declaration the way this tree actually gets one, through
// the program's type declarations, reproduces upstream's two findings exactly.
//
// This is what makes the divergence a resolution difference rather than a rule difference, and it is
// the reason the skip above is a stated scope note instead of a defect.
func TestRequireAtomicUpdatesProcessCaseWithDeclaration(t *testing.T) {
	t.Parallel()

	source := `declare const process: any;
declare function run(options: any): Promise<any>;
async function main() {
    const opts: any = {};
    opts.spec = process.stdin;
    try {
        const { exit_code } = await run(opts);
        process.exitCode = exit_code;
    } catch (e) {
        process.exitCode = 1;
    }
}`
	result := rule_testing.RunTyped(t, RequireAtomicUpdates, "process.ts", source)
	rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate", "nonAtomicObjectUpdate")
}

// TestRequireAtomicUpdatesFinallyReportsOnce pins the duplicate-layout defect the real tree found.
//
// The graph lays a `finally` block out twice, once for normal completion and once for the path
// leaving the `try` abruptly, and both copies carry the same source positions. A write inside a
// `finally` is therefore judged twice, and before the deduplication in `analyzeRootNonAtomicUpdates`
// this reported the identical span twice.
//
// No imported case can see it: upstream's corpus writes no assignment inside a `finally` at all, and
// its one try/catch case puts the writes in the arms. This came off `cohere --lint` over the real
// tree, where the guard-flag shape below appears in several files.
//
// Measured against the installed eslint 10.8.1: one finding, at the `g = false` in the `finally`.
// The three controls beside it are what prove the shape is reported for the right reason, since each
// removes exactly one element and upstream goes silent for two of them.
func TestRequireAtomicUpdatesFinallyReportsOnce(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "a guard flag reset in finally reports once, not once per layout",
			source: `function o() { let g = false; async function f() { if (g) return; g = true; try { await s(); } finally { g = false; } } }`,
			want:   []string{"nonAtomicUpdate"},
		},
		{
			name: "without the guard read there is nothing stale",
			// The control: the `finally` write is still there and the read before the await is gone.
			source: `function o() { let g = false; async function f() { g = true; try { await s(); } finally { g = false; } } }`,
		},
		{
			name: "without the finally write there is nothing to report",
			// The other control: the read is there and the write is gone.
			source: `function o() { let g = false; async function f() { if (g) return; g = true; try { await s(); } finally { } } }`,
		},
		{
			name:   "the same shape without a try still reports once",
			source: `function o() { let g = false; async function f() { if (g) return; await s(); g = false; } }`,
			want:   []string{"nonAtomicUpdate"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RequireAtomicUpdates, "finally.ts", testCase.source)
			if len(testCase.want) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.want...)
			// The span is asserted too, because the defect this pins was two findings over ONE
			// span, which a count assertion alone would pass once the count was fixed by any means.
			written := result.SourceFile.Text()
			got := written[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if got != "g = false" {
				t.Errorf("finding spans %q, want %q", got, "g = false")
			}
		})
	}
}

// TestRequireAtomicUpdatesCatchDoesNotInheritTryRefresh pins the one place this rule reaches a
// different JUDGMENT from ESLint, rather than a different set of resolvable names.
//
// ESLint clears a variable's outdated mark at a read and carries that clearing into the `catch` of
// an enclosing `try`, so a read at the end of a try block silences a write in the catch. The catch
// is reached precisely on the paths where that read did not run, so the refresh it credits never
// happened.
//
// Measured on the installed eslint 10.8.1, and the pair below is the reduction that isolates it:
// without the trailing `use(entry.position)` ESLint reports both writes, and with it ESLint reports
// only the try-block one. Reading an unrelated variable in that position does not change ESLint's
// verdict, and adding a further await after the read restores it, which together show the mechanism
// is the refresh rather than statement count or position.
//
// Both directions are asserted. The second case is the divergence; the first is the control that
// keeps it from passing on a rule that simply reports every catch-block write.
//
// This is `modules/kingdom/KingdomShadeController.ts:203` on the real tree, and it is one of the 8
// findings cohere reports there that ESLint does not. See the rule's doc comment.
func TestRequireAtomicUpdatesCatchDoesNotInheritTryRefresh(t *testing.T) {
	t.Parallel()

	const declarations = `declare function a(): Promise<void>;
declare function b(): Promise<number>;
declare function use(value: unknown): void;
`

	t.Run("both linters report when the try block ends at the await", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "catch.ts", declarations+
			`async function f(entry: any) { if (entry.status !== 1) return;
    try { await a(); entry.position = await b(); } catch (error) { entry.error = 1; } }`)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate", "nonAtomicObjectUpdate")
	})

	t.Run("a read at the end of the try does not refresh the catch", func(t *testing.T) {
		t.Parallel()
		// ESLint reports only `entry.position` here. Cohere reports both, because the fork to the
		// handler is a real successor edge and the read on the normal path is not on it.
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "catch.ts", declarations+
			`async function f(entry: any) { if (entry.status !== 1) return;
    try { await a(); entry.position = await b(); use(entry.position); } catch (error) { entry.error = 1; } }`)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate", "nonAtomicObjectUpdate")

		// The span is asserted so the catch-block write is named specifically rather than the count
		// being satisfied by two findings on the try-block one.
		written := result.SourceFile.Text()
		var sawCatchWrite bool
		for _, diagnostic := range result.Diagnostics {
			if written[diagnostic.Range.Pos():diagnostic.Range.End()] == "entry.error = 1" {
				sawCatchWrite = true
			}
		}
		if !sawCatchWrite {
			t.Error("the catch-block write was not among the findings")
		}
	})

	t.Run("a guard read is still what makes the object outdated", func(t *testing.T) {
		t.Parallel()
		// The control for the whole test. With no read before the suspension there is nothing stale,
		// and both linters are silent, which is what keeps the two rows above from passing on a rule
		// that reports every property write in a catch.
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "catch.ts", declarations+
			`async function f(entry: any) {
    try { await a(); } catch (error) { entry.error = 1; } }`)
		rule_testing.ExpectClean(t, result)
	})
}

// TestRequireAtomicUpdatesRestoreInFinallyIsJudged pins the shape behind seven of the eight findings
// cohere reports on the real tree that ESLint does not.
//
// A handler saved into a local, replaced, and restored in a `finally` after an await is a real
// last-writer-wins race: two overlapping calls restore in the wrong order and the second installs a
// handler that was already torn down. `modules/art/ArtTerminal.ts:414-416` and
// `modules/phi/social/PhiSocialTerminal.ts:480-482` are that pattern over `console.log`,
// `console.info` and `process.stdout.write`.
//
// ESLint is silent on it here, and the cause is which globals a `.ts` file receives from this
// project's config rather than a difference in judgment. The rule's doc comment carries the two
// lists and the measurement; what this test pins is that the judgment itself does not depend on the
// binding being a global. The two cases are the same shape over a local and over an ambient
// declaration, and both report, which is what makes the ESLint difference attributable to resolution
// rather than to the shape being one we get wrong.
//
// This mirrors the control seeded into `ArtTerminal.ts` itself: two structurally identical writes in
// one `finally`, one over a locally declared object and one over `console`, produced exactly one
// ESLint finding, the local one. Same file, same run, opposite verdicts.
func TestRequireAtomicUpdatesRestoreInFinallyIsJudged(t *testing.T) {
	t.Parallel()

	t.Run("restoring a local object's property is judged", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "restore.ts",
			`declare function callback(): Promise<void>;
const holder = { slot: 0 };
async function withRestore() {
    const saved = holder.slot;
    try { await callback(); }
    finally { holder.slot = saved; }
}`)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate")
	})

	t.Run("restoring through an ambient declaration is judged the same way", func(t *testing.T) {
		t.Parallel()
		// The binding is declared rather than local, which is the only thing that changes. Both
		// report here, so the shape is not what separates cohere from ESLint on the real tree.
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "restore.ts",
			`declare function callback(): Promise<void>;
declare const ambientHolder: { slot: number };
async function withRestore() {
    const saved = ambientHolder.slot;
    try { await callback(); }
    finally { ambientHolder.slot = saved; }
}`)
		rule_testing.ExpectFindings(t, result, "nonAtomicObjectUpdate")
	})

	t.Run("with no read before the suspension there is nothing stale", func(t *testing.T) {
		t.Parallel()
		// The control. Without the saved read there is no pre-suspension value for the restore to be
		// built from, and this is silent, which keeps the two rows above from passing on a rule that
		// reports every property write in a finally.
		result := rule_testing.RunTyped(t, RequireAtomicUpdates, "restore.ts",
			`declare function callback(): Promise<void>;
const holder = { slot: 0 };
async function withRestore() {
    try { await callback(); }
    finally { holder.slot = 0; }
}`)
		rule_testing.ExpectClean(t, result)
	})
}
