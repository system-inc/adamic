package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// unmodifiedLoopFile is where the fixtures pretend to live.
const unmodifiedLoopFile = "/repository/source/LoopConditions.ts"

// The corpus is ESLint's own, loaded by running its tester file with the RuleTester stubbed so the
// cases are the file's values rather than retyped, then driven through the installed eslint 10.8.1
// build to record what it reports.
//
// 40 cases. 37 agreed with the corpus's own stated verdict on the installed build. The other 3
// carry an option that build does not have, and they are held separately in
// unmodifiedLoopOptionOnlyCases rather than asserted.

// TestNoUnmodifiedLoopConditionFires runs every corpus input upstream reports on.
//
// The assertion names the variables and their order, not only the count. Two cases report twice, on
// two different names, and a rule finding the right number of problems while naming the wrong
// binding passes a count fixture.
func TestNoUnmodifiedLoopConditionFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range unmodifiedLoopFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)

			wantIds := make([]string, len(testCase.variables))
			for index := range wantIds {
				wantIds[index] = "loopConditionNotModified"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for index, wantVariable := range testCase.variables {
				want := "The variable is '" + wantVariable + "'."
				got := result.Diagnostics[index].Message.Description
				if !strings.HasSuffix(got, want) {
					t.Errorf("finding %d message: got %q, want it to end with %q", index, got, want)
				}
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionStaysSilent runs every corpus input upstream leaves alone.
//
// This half is the rule. A port that reports any loop-condition variable without a write in the
// body passes most of the firing table and fails here on every call, every member access, every
// group with one changing member, and the function called from inside the loop.
func TestNoUnmodifiedLoopConditionStaysSilent(t *testing.T) {
	t.Parallel()

	for _, source := range unmodifiedLoopCleanCases {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, source))
		})
	}
}

// TestNoUnmodifiedLoopConditionOptionOnlyCasesAreRecordedNotAsserted keeps the version gap visible.
//
// The clone at 10.9.1 adds `checkConditionalExpressions`; the installed build at 10.8.1 has an empty
// schema and throws at config validation when the option is set, so these three cases have no
// oracle verdict to import. This rule ports the installed build and takes no options.
//
// The test asserts only that the three cases still parse and run without crashing, which is the
// most that can honestly be claimed about them. It deliberately does NOT assert a finding count:
// under 10.8.1 semantics a ConditionalExpression is always a group, and what the option would
// change is exactly the thing nothing here can measure.
//
// When the installed build moves to 10.9, this test is the marker for the follow-up.
func TestNoUnmodifiedLoopConditionOptionOnlyCasesAreRecordedNotAsserted(t *testing.T) {
	t.Parallel()

	if len(unmodifiedLoopOptionOnlyCases) == 0 {
		t.Fatal("the option-only table is empty, so the version gap it records has gone unnoticed")
	}
	for _, testCase := range unmodifiedLoopOptionOnlyCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			// Runs the rule for its own sake: these are real sources and a crash on one is a
			// defect whatever the option question is.
			rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
		})
	}
}

// TestNoUnmodifiedLoopConditionSpans asserts where the finding points.
//
// Upstream reports on the IDENTIFIER inside the condition, not on the loop and not on the
// condition, so the span is one name. The two-finding case pins both spans and their order.
func TestNoUnmodifiedLoopConditionSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports []string
	}{
		{
			"the identifier in the condition, not the whole condition",
			"var foo = 0; while (foo) { } foo = 1;",
			[]string{"foo"},
		},
		{
			"the operand inside a negation",
			"var foo = 0; while (!foo) { } foo = 1;",
			[]string{"foo"},
		},
		{
			"both members of an unmodified group, in source order",
			"var foo = 0, bar = 9; while (foo < bar) { } foo = 1;",
			[]string{"foo", "bar"},
		},
		{
			"only the unmodified member when its partner changes",
			"var foo = 0, bar = 0; while (foo && bar) { ++bar; } foo = 1;",
			[]string{"foo"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			if len(result.Diagnostics) != len(testCase.reports) {
				t.Fatalf("expected %d findings, got %d", len(testCase.reports), len(result.Diagnostics))
			}
			// Sliced from the source the harness wrote. RunTyped trims, so a leading newline in
			// the literal would offset every position; these cases have none, and going through
			// the same transform keeps that from being load-bearing.
			onDisk := strings.TrimSpace(testCase.source) + "\n"
			for index, want := range testCase.reports {
				diagnostic := result.Diagnostics[index]
				got := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if got != want {
					t.Errorf("finding %d span: got %q, want %q", index, got, want)
				}
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionMessage asserts the message id and both halves of the description.
//
// The description interpolates the variable name, so the id assertion cannot see anything the
// format string does. Asserted against literals typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from is an equality both
// sides of which move together under mutation.
func TestNoUnmodifiedLoopConditionMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile,
		"var foo = 0; while (foo) { } foo = 1;")
	rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
	message := result.Diagnostics[0].Message

	if message.Id != "loopConditionNotModified" {
		t.Errorf("message id: got %q, want %q", message.Id, "loopConditionNotModified")
	}
	const wantPrefix = "This variable is read by the loop condition and never changed inside the loop,"
	if !strings.HasPrefix(message.Description, wantPrefix) {
		t.Errorf("description: got %q, want it to start with %q", message.Description, wantPrefix)
	}
	const wantSuffix = "The variable is 'foo'."
	if !strings.HasSuffix(message.Description, wantSuffix) {
		t.Errorf("description: got %q, want it to end with %q", message.Description, wantSuffix)
	}
	if strings.Count(message.Description, "'foo'") != 1 {
		t.Errorf("description names 'foo' %d times, want 1: %q",
			strings.Count(message.Description, "'foo'"), message.Description)
	}
}

// TestNoUnmodifiedLoopConditionRequiresTheTypedHarness pins that the rule declines without a
// checker.
//
// `GetSymbolAtLocation` on a nil checker returns nil rather than crashing, so a rule missing its
// guard goes silently inert instead of announcing itself, and every StaysSilent case would then
// pass vacuously.
func TestNoUnmodifiedLoopConditionRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "var foo = 0; while (foo) { } foo = 1;"

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, source),
		"loopConditionNotModified")

	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, source))

	if !NoUnmodifiedLoopCondition.NeedsTypeChecker {
		t.Error("the rule must declare NeedsTypeChecker, or the live run hands it a nil checker")
	}
}

// TestNoUnmodifiedLoopConditionForInAndForOfAreNeverJudged covers a shape the corpus omits.
//
// Only `while`, `do`/`while` and `for` have a test. A `for...in` or `for...of` has no condition, and
// upstream's LOOP_PATTERN excludes them by name with a comment saying exactly that. The corpus
// writes no `for...in` or `for...of` at all, so a port that treated their iterated expression as a
// condition would pass every imported case.
//
// Measured against the installed eslint 10.8.1 build: silent on both.
func TestNoUnmodifiedLoopConditionForInAndForOfAreNeverJudged(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		"var foo = {}; for (var key in foo) { } foo = 1;",
		"var foo = []; for (var item of foo) { } foo = 1;",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, source))
		})
	}
}

// TestNoUnmodifiedLoopConditionUnnamedFunctionDeclarationHasNoNameToReach covers a guard the corpus
// never exercises.
//
// `getEncloseFunctionDeclaration` returns null when the declaration it lands on has no id, because
// the whole mechanism is looking that name up to see whether it is referenced from the loop. The
// only unnamed function declaration in the language is `export default function () {}`, which the
// corpus does not write, so a mutant dropping the guard survives all 37 imported cases.
//
// The three rows separate the three things that could be doing the work: the name's absence, the
// name being referenced, and the name existing but unreferenced.
//
// Measured against the installed eslint 10.8.1 build, every row.
func TestNoUnmodifiedLoopConditionUnnamedFunctionDeclarationHasNoNameToReach(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		fires  bool
	}{
		{
			"an unnamed export default function cannot be reached by name, so its write does not count",
			"var foo = 0; while (foo) { } export default function () { ++foo; }",
			true,
		},
		{
			"the same function named and called in the loop does count",
			"var foo = 0; while (foo) { d(); } export default function d() { ++foo; }",
			false,
		},
		{
			"named but never called from the loop does not",
			"var foo = 0; while (foo) { } export default function d() { ++foo; }",
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			if testCase.fires {
				rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionClimbStopsAtTheNearestFunctionDeclaration covers where the search
// for a caller stops.
//
// `getEncloseFunctionDeclaration` climbs to the FIRST function declaration and stops there, rather
// than continuing outward to one that happens to be referenced from the loop. The corpus never
// nests, so a mutant that skipped past a declaration survived all 37 imported cases and the
// unnamed-function fixtures too.
//
// The three rows separate the two ways a write can sit inside a called function:
//
//	a named inner declaration      the climb stops at `inner`, which the loop never names, so it
//	                               reports even though the outer `d` IS called from the loop
//	an unnamed function expression not a declaration at all, so the climb passes through it to `d`
//	a write directly in `d`        the plain case, clean
//
// Measured against the installed eslint 10.8.1 build, every row.
func TestNoUnmodifiedLoopConditionClimbStopsAtTheNearestFunctionDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		fires  bool
	}{
		{
			"a named inner declaration stops the climb, and it is not called from the loop",
			"var foo = 0; while (foo) { d(); } function d() { function inner() { ++foo; } inner(); }",
			true,
		},
		{
			"a function expression is not a declaration, so the climb reaches the called one",
			"var foo = 0; while (foo) { d(); } function d() { (function () { ++foo; })(); }",
			false,
		},
		{
			"a write directly in the called function",
			"var foo = 0; while (foo) { d(); } function d() { ++foo; }",
			false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			if testCase.fires {
				rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionDynamicCheckDoesNotDescendIntoFunctions covers upstream's
// SKIP_PATTERN, which the corpus never exercises.
//
// The dynamic-expression check abandons a group holding a call, a member access, a `new`, a tagged
// template or a `yield`, because any of those may be changing what the rule is about to call
// unchanging. It deliberately does NOT look inside a nested function, since a function that is
// merely written is not called.
//
// Every dynamic case in the corpus writes the call at the top level of the condition, so a mutant
// deleting the skip survived all 37 of them. These rows put the call inside a function in the
// condition: with the skip the group survives and the rule reports, without it the group is
// abandoned and the rule goes silent.
//
// Measured against the installed eslint 10.8.1 build, every row.
func TestNoUnmodifiedLoopConditionDynamicCheckDoesNotDescendIntoFunctions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		source  string
		reports int
	}{
		{
			"a call inside an arrow in the condition does not abandon the group",
			"var foo = 0; while (foo === (() => q())) { } foo = 1;",
			1,
		},
		{
			"a call inside a function expression does not either",
			"var foo = 0; while (foo === (function(){ return q(); })) { } foo = 1;",
			1,
		},
		{
			"a member access inside one does not either",
			"var foo = 0, bar = 0; while (foo === (function(){ return x.y; })) { } foo = 1;",
			1,
		},
		{
			"a call written directly in the condition DOES abandon it",
			"var foo = 0; while (foo === q()) { } foo = 1;",
			0,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			wantIds := make([]string, testCase.reports)
			for index := range wantIds {
				wantIds[index] = "loopConditionNotModified"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// TestNoUnmodifiedLoopConditionVarInitializerCountsAsAWrite covers upstream's initializer exception.
//
// `isWriteReference` declines a reference marked as an initializer UNLESS the binding is a `var`.
// The corpus exercises that only through `for (var foo = 0; foo < 10; ) { } foo = 1;`, where the
// exception makes the head's `foo = 0` a write that then fails the in-the-loop test anyway, so the
// case reports either way and a mutant removing the exception survives it.
//
// The shape that separates them is a `var` declaration with an initializer INSIDE the loop body,
// where the write does count as in the loop. The `let` row is the pure contrast: same source except
// for the keyword, and the `let` shadows rather than writing the outer binding.
//
// Measured against the installed eslint 10.8.1 build, every row.
func TestNoUnmodifiedLoopConditionVarInitializerCountsAsAWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		fires  bool
	}{
		{
			"a var initializer inside the loop body is a write, so the condition is modified",
			"while (foo) { var foo = 0; }",
			false,
		},
		{
			"the same with the declaration also written above",
			"var foo; while (foo) { var foo = 1; }",
			false,
		},
		{
			"a let inside the body shadows instead, so the outer binding is unmodified",
			"var foo; while (foo) { let foo = 1; }",
			true,
		},
		{
			"an initializer OUTSIDE the loop is a write that cannot run in it",
			"var foo = 0; while (foo) { }",
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			if testCase.fires {
				rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionModifierReachedThroughACall covers the pair the corpus states from
// only one side at a time.
//
// A write inside a named function declaration counts when that function's name is referenced from
// inside the loop. The corpus has the clean row and the shadowed row, and it does not have the row
// where the function exists but is never called from the loop, which is what separates "there is a
// function that writes it" from "that function can run here".
//
// Measured against the installed eslint 10.8.1 build, every row.
func TestNoUnmodifiedLoopConditionModifierReachedThroughACall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		fires  bool
	}{
		{
			"a function called in the loop that writes the variable makes it modified",
			"var foo = 0; while (foo) { update(); } function update() { ++foo; }",
			false,
		},
		{
			"the same function whose parameter shadows the name does not",
			"var foo = 0; while (foo) { update(); } function update(foo) { ++foo; }",
			true,
		},
		{
			"a function that writes it but is never called from the loop does not either",
			"var foo = 0; while (foo) { other(); } function update() { ++foo; } function other() {}",
			true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, testCase.source)
			if testCase.fires {
				rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoUnmodifiedLoopConditionSeesAWriterThatRunsWhileTheLoopIsSuspended is where cohere is
// deliberately quieter than ESLint.
//
// Upstream credits a write only when it sits in the loop, or in a function DECLARATION whose name is
// referenced from the loop. It cannot see the most common modern cancellation shape: a flag set by a
// signal handler or an abort callback while the loop is suspended at an `await`. A suspended loop
// hands the thread to whatever else is queued, so a write in another function activation, registered
// before the loop could finish, can change the condition between iterations. ESLint reports all three
// such loops in ahra, and all three are false.
//
// The first three silent rows are those sites, cut to the lines that decide them:
// `modules/tasks/TasksWatchCommandLineInterface.ts:305` (a hoisted SIGINT handler),
// `modules/os/sensation/AhraOsMonitors.ts:548` (an `abort` closure in the enclosing function of an async
// IIFE), and `modules/os/boot-screens/RainbowMatrix.ts:689` (`teardown`, reached through `onSignal`).
//
// The reporting rows are the controls that keep the loop detector alive: an awaiting loop with no
// writer at all, the same handler beside a loop that never suspends, an `await` that belongs to a
// nested function rather than the loop, a write in the loop's own activation before the loop, a
// closure created only after the loop, a hoisted writer nothing registers, one registered only after
// the loop, and a closure writing a shadowing binding.
func TestNoUnmodifiedLoopConditionSeesAWriterThatRunsWhileTheLoopIsSuspended(t *testing.T) {
	t.Parallel()

	const ambient = "declare const process: { on(event: string, handler: () => void): void };\n" +
		"declare function sleep(): Promise<void>;\n" +
		"declare function setTimeout(callback: (value?: unknown) => void, delay: number): void;\n"

	cases := []struct {
		name   string
		source string
		fires  bool
	}{
		{"TasksWatchCommandLineInterface.ts:305, a hoisted signal handler", `
async function watch() {
    let stopping = false;
    function handleShutdown(): void {
        if(stopping) return;
        stopping = true;
    }
    process.on('SIGINT', handleShutdown);
    while(!stopping) {
        await new Promise(function(resolve) {
            setTimeout(resolve, 10);
        });
        if(stopping) break;
    }
}
`, false},
		{"AhraOsMonitors.ts:548, an abort closure beside an async IIFE", `
declare function driveMonitor(row: number): Promise<void>;
declare function nextRow(row: number): number | null;
function superviseMonitor(register: (child: { abort(): void }) => void): void {
    let abortRequested = false;
    const child = {
        abort: function() {
            abortRequested = true;
        },
    };
    register(child);
    (async function() {
        let currentRow: number | null = 1;
        while(!abortRequested && currentRow !== null) {
            await driveMonitor(currentRow);
            if(abortRequested) break;
            currentRow = nextRow(currentRow);
        }
    })();
}
`, false},
		{"RainbowMatrix.ts:689, teardown reached through a registered handler", `
async function play() {
    let stopped = false;
    function teardown(): void {
        if(stopped) return;
        stopped = true;
    }
    function onSignal(): void {
        teardown();
    }
    process.on('SIGINT', onSignal);
    while(!stopped) {
        await sleep();
    }
}
`, false},
		{"a generator suspended at a yield", `
let stop = false;
process.on('SIGINT', () => { stop = true; });
function* ticks() {
    while(!stop) {
        yield 1;
    }
}
`, false},
		{"an enclosing function's own code, after it starts the loop", `
function outer() {
    let stop = false;
    async function run() {
        while(!stop) {
            await sleep();
        }
    }
    run();
    stop = true;
}
`, false},
		{"a hoisted handler declared after the loop and registered before it", `
async function watch() {
    let stop = false;
    process.on('SIGINT', halt);
    while(!stop) {
        await sleep();
    }
    function halt(): void {
        stop = true;
    }
}
`, false},
		{"an await in the update clause", `
async function poll() {
    let stop = false;
    process.on('SIGINT', () => { stop = true; });
    for(let index = 0; !stop; index = index + await sleep().then(() => 1)) {
    }
}
`, false},

		{"an enclosing function registers a closure after starting the loop", `
function outer() {
    let stop = false;
    async function run() {
        while(!stop) {
            await sleep();
        }
    }
    run();
    process.on('SIGINT', () => { stop = true; });
}
`, false},
		{"a hoisted handler further out, registered after the loop starts", `
let stop = false;
async function run() {
    while(!stop) {
        await sleep();
    }
}
run();
process.on('SIGINT', halt);
function halt(): void {
    stop = true;
}
`, false},

		{"an await only in a for initializer, which runs once", `
async function poll() {
    let stop = false;
    process.on('SIGINT', () => { stop = true; });
    for(let index = await sleep(); !stop; ) {
        Math.random();
    }
}
`, true},
		{"an awaiting loop with no writer at all", `
async function watch() {
    let stop = false;
    while(!stop) {
        await sleep();
    }
}
`, true},
		{"a registered handler beside a loop that never suspends", `
function spin() {
    let stop = false;
    process.on('SIGINT', () => { stop = true; });
    while(!stop) {
        Math.random();
    }
}
`, true},
		{"an await that belongs to a nested function, not the loop", `
declare const items: number[];
function spin() {
    let stop = false;
    process.on('SIGINT', () => { stop = true; });
    while(!stop) {
        items.forEach(async () => { await sleep(); });
    }
}
`, true},
		{"a write in the loop's own activation before the loop", `
async function watch() {
    let stop = true;
    stop = false;
    while(!stop) {
        await sleep();
    }
}
`, true},
		{"a closure created only after the loop", `
async function watch() {
    let stop = false;
    while(!stop) {
        await sleep();
    }
    process.on('SIGINT', () => { stop = true; });
}
`, true},
		{"a hoisted writer nothing registers", `
async function watch() {
    let stop = false;
    while(!stop) {
        await sleep();
    }
    function halt(): void {
        stop = true;
    }
}
`, true},
		{"a hoisted writer declared before the loop that nothing registers", `
async function watch() {
    let stop = false;
    function halt(): void {
        stop = true;
    }
    while(!stop) {
        await sleep();
    }
}
`, true},
		{"a hoisted writer registered only after the loop", `
async function watch() {
    let stop = false;
    while(!stop) {
        await sleep();
    }
    process.on('SIGINT', halt);
    function halt(): void {
        stop = true;
    }
}
`, true},
		{"a closure writing a shadowing binding", `
async function watch() {
    let stop = false;
    process.on('SIGINT', () => { let stop = false; stop = true; });
    while(!stop) {
        await sleep();
    }
}
`, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnmodifiedLoopCondition, unmodifiedLoopFile, ambient+testCase.source)
			if testCase.fires {
				rule_testing.ExpectFindings(t, result, "loopConditionNotModified")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}
