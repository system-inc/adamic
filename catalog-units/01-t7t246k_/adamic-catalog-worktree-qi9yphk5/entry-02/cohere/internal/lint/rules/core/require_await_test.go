package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// requireAwaitFile is where the fixtures pretend to live.
const requireAwaitFile = "/repository/source/RequireAwait.ts"

// requireAwaitCase is one imported corpus row.
type requireAwaitCase struct {
	sourceText string
	// wantNames is the description fragment each finding uses to name the function, in order.
	// nil means the case is clean. This is half the rendered message and no message-id
	// assertion can see it.
	wantNames []string
	// wantReplacements is what each finding's suggestion writes in place of `async`. Almost
	// always empty; a semicolon where removing the keyword would let automatic semicolon
	// insertion join two statements.
	wantReplacements []string
}

func runRequireAwait(t *testing.T, testCase requireAwaitCase) rule_testing.Result {
	t.Helper()
	return rule_testing.Run(t, RequireAwait, requireAwaitFile, testCase.sourceText)
}

// The corpus is ESLint's own, extracted mechanically rather than retyped.
//
// `tests/lib/rules/require-await.js` was loaded with its RuleTester stubbed so every case came out
// as data, then replayed against the INSTALLED rule to record what it reports, how each finding
// names its function, and what each suggestion writes. All 41 reproduced.
//
// Every finding this rule produces carries a suggestion and none carries a fix, so there is no
// `ExpectFixedSource` here: `cohereAndFix` reports no change at all for this rule, because a
// suggestion is offered to a human rather than applied. Asserting the suggestion's replacement text
// is the equivalent, and it is done per finding below.
func requireAwaitFiresCases() []requireAwaitCase {
	return []requireAwaitCase{
		{"async function foo() { doSomething() }", []string{"Async function 'foo'"}, []string{""}},
		{"(async function() { doSomething() })", []string{"Async function"}, []string{""}},
		{"async () => { doSomething() }", []string{"Async arrow function"}, []string{""}},
		{"async () => doSomething()", []string{"Async arrow function"}, []string{""}},
		{"({ async foo() { doSomething() } })", []string{"Async method 'foo'"}, []string{""}},
		{"class A { async foo() { doSomething() } }", []string{"Async method 'foo'"}, []string{""}},
		{"(class { async foo() { doSomething() } })", []string{"Async method 'foo'"}, []string{""}},
		{"(class { async ''() { doSomething() } })", []string{"Async method ''"}, []string{""}},
		{"async function foo() { async () => { await doSomething() } }", []string{"Async function 'foo'"}, []string{""}},
		{"async function foo() { await (async () => { doSomething() }) }", []string{"Async arrow function"}, []string{""}},
		{"const obj = { async: async function foo() { bar(); } }", []string{"Async method 'async'"}, []string{""}},
		{"async    /* test */ function foo() { doSomething() }", []string{"Async function 'foo'"}, []string{""}},
		{"class A {\n                a = 0\n                async [b](){ return 0; }\n            }", []string{"Async method"}, []string{";"}},
		{"class A {\n                a\n                async [b](){ return 0; }\n            }", []string{"Async method"}, []string{""}},
		{"class A {\n                a = 0\n                async in(){ return 0; }\n            }", []string{"Async method 'in'"}, []string{";"}},
		{"const obj = {\n                foo,\n                async in(){ return 0; }\n            }", []string{"Async method 'in'"}, []string{""}},
		{"foo\n                async () => { return 0; }\n            ", []string{"Async arrow function"}, []string{";"}},
		{"class A {\n                foo() {}\n                async [bar] () { baz; }\n            }", []string{"Async method"}, []string{""}},
		{"async function run() { using resource = getResource(); }", []string{"Async function 'run'"}, []string{""}},
	}
}

func requireAwaitSilentCases() []requireAwaitCase {
	return []requireAwaitCase{
		{"async function foo() { await doSomething() }", nil, nil},
		{"(async function() { await doSomething() })", nil, nil},
		{"async () => { await doSomething() }", nil, nil},
		{"async () => await doSomething()", nil, nil},
		{"({ async foo() { await doSomething() } })", nil, nil},
		{"class A { async foo() { await doSomething() } }", nil, nil},
		{"(class { async foo() { await doSomething() } })", nil, nil},
		{"async function foo() { await (async () => { await doSomething() }) }", nil, nil},
		{"async function foo() {}", nil, nil},
		{"async () => {}", nil, nil},
		{"function foo() { doSomething() }", nil, nil},
		{"async function foo() { for await (x of xs); }", nil, nil},
		{"await foo()", nil, nil},
		{"\n                for await (let num of asyncIterable) {\n                    console.log(num);\n                }\n            ", nil, nil},
		{"async function* run() { yield * anotherAsyncGenerator() }", nil, nil},
		{"async function* run() {\n                await new Promise(resolve => setTimeout(resolve, 100));\n                yield 'Hello';\n                console.log('World');\n            }\n            ", nil, nil},
		{"async function* run() { }", nil, nil},
		{"const foo = async function *(){}", nil, nil},
		{"const foo = async function *(){ console.log(\"bar\") }", nil, nil},
		{"async function* run() { console.log(\"bar\") }", nil, nil},
		{"await using resource = getResource();", nil, nil},
		{"async function run() { await using resource = getResource(); }", nil, nil},
	}
}

func TestRequireAwaitFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range requireAwaitFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, len(testCase.wantNames))
			for index := range wantIds {
				wantIds[index] = "missingAwait"
			}
			rule_testing.ExpectFindings(t, runRequireAwait(t, testCase), wantIds...)
		})
	}
}

func TestRequireAwaitStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range requireAwaitSilentCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runRequireAwait(t, testCase))
		})
	}
}

// How each finding NAMES its function, which the message-id assertions cannot see.
//
// Upstream renders nine distinct descriptions from `getFunctionNameWithKind`, and the differences
// between them are real judgments rather than formatting: a method is named by its property rather
// than by the function bound to it, an empty string literal name renders as a name rather than as
// an absent one, and a computed name renders as no name at all.
func TestRequireAwaitNamesTheFunction(t *testing.T) {
	t.Parallel()

	for _, testCase := range requireAwaitFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runRequireAwait(t, testCase)
			if len(result.Diagnostics) != len(testCase.wantNames) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantNames),
					len(result.Diagnostics))
			}
			for index, wantName := range testCase.wantNames {
				got := result.Diagnostics[index].Message.Description
				// Anchored on the sentence opening rather than a bare Contains, so a
				// description that merely mentioned the name somewhere would not satisfy it.
				if !strings.HasPrefix(got, wantName+" has no ") {
					t.Errorf("finding %d rendered\n  %q\nwhich does not open by naming %q",
						index, got, wantName)
				}
			}
		})
	}
}

// What each suggestion WRITES, which is the repair half of this rule.
//
// Every finding carries exactly one suggestion and none carries a fix, so a rule that offered its
// repair as a fix would rewrite source the engine should only have offered to rewrite. Both halves
// are asserted: the count and kind of repair, and the text it writes.
//
// The semicolon rows are the ones to read. Removing `async` can let automatic semicolon insertion
// join the construct to the line above, and upstream substitutes a semicolon rather than emitting
// nothing. Three of its own cases exercise it, all measured.
func TestRequireAwaitSuggestsRemovingAsync(t *testing.T) {
	t.Parallel()

	semicolons := 0
	for _, testCase := range requireAwaitFiresCases() {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runRequireAwait(t, testCase)
			if len(result.Diagnostics) != len(testCase.wantReplacements) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantReplacements),
					len(result.Diagnostics))
			}
			for index, wantText := range testCase.wantReplacements {
				diagnostic := result.Diagnostics[index]
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d carried %d fixes; this rule only ever suggests",
						index, len(diagnostic.Fixes))
				}
				if len(diagnostic.Suggestions) != 1 {
					t.Fatalf("finding %d carried %d suggestions, wanted exactly one",
						index, len(diagnostic.Suggestions))
				}
				fixes := diagnostic.Suggestions[0].Fixes
				if len(fixes) != 1 {
					t.Fatalf("finding %d's suggestion carried %d fixes, wanted one", index,
						len(fixes))
				}
				if got := fixes[0].Text; got != wantText {
					t.Errorf("finding %d's suggestion writes %q, wanted %q", index, got,
						wantText)
				}
			}
		})
		for _, text := range testCase.wantReplacements {
			if text == ";" {
				semicolons++
			}
		}
	}
	// The semicolon arm is the one most easily lost, so its coverage is asserted rather than
	// assumed: a change that stopped emitting it would otherwise leave every row still green
	// on the empty-string cases alone.
	if semicolons != 3 {
		t.Errorf("%d rows exercised the semicolon arm, wanted 3", semicolons)
	}
}

// The two halves of the semicolon condition, separated.
//
// Written for a surviving mutant: removing the continuation test left all 41 imported rows green,
// because in every corpus case the two halves agree. They only disagree when the previous member IS
// a continuable expression and the token after `async` is NOT one that could continue it, which the
// corpus never writes.
//
// `a = 0` followed by `async m()` is that shape. A plain identifier cannot extend `0` into a larger
// expression, so no semicolon is needed even though the property has an unterminated initializer.
// Every verdict below was measured against the installed rule before the row was written.
func TestRequireAwaitSemicolonNeedsBothHalves(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantText   string
	}{
		// The continuation test alone decides these three: same previous member, different
		// following token.
		{"class A {\n a = 0\n async m(){ return 0; }\n}", ""},
		{"class A {\n a = 0\n async foo(){ return 0; }\n}", ""},
		{"class A {\n a = 0\n async [b](){ return 0; }\n}", ";"},
		{"class A {\n a = 0\n async in(){ return 0; }\n}", ";"},
		{"class A {\n a = 0\n async instanceof(){ return 0; }\n}", ";"},
		// And the previous-member test alone decides these: same following token, different
		// predecessor. Without an initializer there is no open expression to continue.
		{"class A {\n a\n async [b](){ return 0; }\n}", ""},
		{"class A {\n foo() {}\n async [bar] () { baz; }\n}", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runRequireAwait(t, requireAwaitCase{sourceText: testCase.sourceText})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
				t.Fatalf("wanted one suggestion carrying one fix, got %d suggestions",
					len(suggestions))
			}
			if got := suggestions[0].Fixes[0].Text; got != testCase.wantText {
				t.Errorf("suggestion writes %q, wanted %q", got, testCase.wantText)
			}
		})
	}
}

// A function carrying modifiers BEFORE `async`, which the corpus never writes.
//
// Written for a surviving mutant: the loop that finds the `async` keyword skips any modifier that is
// not `async`, and deleting that skip left all 48 imported rows green. Every corpus case has `async`
// as its only modifier, so the first modifier IS the right one and a loop that takes the first
// blindly agrees by accident.
//
// Real code does not look like that. `static async`, `public async` and `export async` are ordinary,
// and there the first modifier is the wrong one: the suggestion would delete `static` and leave the
// function async, which is both wrong and silent, since the finding still appears and still carries
// a repair. Upstream was driven over each shape below and removes exactly "async " every time.
func TestRequireAwaitFindsAsyncPastOtherModifiers(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class A { static async foo() { return 1; } }",
		"class A { public async foo() { return 1; } }",
		"export async function foo() { return 1; }",
		"class A { private static async foo() { return 1; } }",
	}
	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := runRequireAwait(t, requireAwaitCase{sourceText: sourceText})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
				t.Fatalf("wanted one suggestion carrying one fix")
			}
			fix := suggestions[0].Fixes[0]
			// The span is asserted rather than the resulting text, because a repair that
			// removed the wrong keyword would still produce compiling source.
			if removed := sourceText[fix.Range.Pos():fix.Range.End()]; removed != "async " {
				t.Errorf("suggestion removes %q, wanted %q", removed, "async ")
			}
			if fix.Text != "" {
				t.Errorf("suggestion writes %q, wanted the empty string", fix.Text)
			}
		})
	}
}

// WHERE each finding is reported, which no message-id assertion can see.
//
// This is the arm the imported corpus was blind to. `rule_testing` compares message ids, so all 48
// rows passed while five of these eleven shapes reported on the wrong column: the span started at
// the `async` keyword where upstream starts at the property that owns the function. The differential
// against the installed rule over the ahra tree agreed on every LINE and disagreed on 490 of 767
// COLUMNS, which is what a line-only comparison reports as perfect agreement.
//
// The branch order is the substance. Upstream tests the parent BEFORE the arrow case, so a property
// holding an arrow reports on the property rather than on the arrow token. Both orderings produce a
// plausible span and only one matches. Every expectation below was measured against the installed
// rule.
func TestRequireAwaitReportsOnTheFunctionHead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantHead   string
	}{
		// The parent branch, which wins over both branches below it.
		{"const o = { run: async function(a) { return 1; } };", "run: async function"},
		{"class A { foo = async function(a) { return 1; } }", "foo = async function"},
		{"class A { static foo = async function(a) { return 1; } }", "static foo = async function"},
		// The same branch with an arrow, where the ordering is load-bearing: an arrow reached
		// through a property reports on the property, NOT on the arrow token.
		{"const o = { run: async (a) => { return 1; } };", "run: async "},
		{"class A { foo = async (a) => { return 1; } }", "foo = async "},
		// The arrow branch, reached only when no parent claimed the head.
		{"const f = async (a) => { return 1; };", "=>"},
		// The plain branch: the function's own start through the end of its name, or through the
		// opening parenthesis when it has none.
		{"const o = { async run(a) { return 1; } };", "async run"},
		{"class A { static async foo(a) { return 1; } }", "static async foo"},
		{"async function foo(a) { return 1; }", "async function foo"},
		{"const f = async function(a) { return 1; };", "async function"},
		{"const f = async function named(a) { return 1; };", "async function named"},
		// The parent branch requires the function to BE the value, not merely to sit somewhere
		// inside it. A function nested in an array or a call within the initializer reports on
		// itself, and an arrow in that position falls through to the arrow branch. Without the
		// value test these four would take the property's start and every span would be wrong by
		// the width of the property.
		{"class A { foo = [async function(a) { return 1; }] }", "async function"},
		{"class A { foo = bar(async function(a) { return 1; }) }", "async function"},
		{"const o = { run: [async function(a) { return 1; }] };", "async function"},
		{"const o = { run: bar(async (a) => { return 1; }) };", "=>"},
		// An `export` is a wrapper around the declaration rather than part of it, so the span
		// starts at `async`; a class member modifier IS part of the head and is kept. Both
		// directions are pinned because skipping every modifier looks just as correct from the
		// inside and breaks the other half.
		{"export async function foo(a) { return 1; }", "async function foo"},
		{"export default async function foo(a) { return 1; }", "async function foo"},
		{"export default async function(a) { return 1; }", "async function"},
		{"class A { public async foo(a) { return 1; } }", "public async foo"},
		{"class A { private static async foo(a) { return 1; } }", "private static async foo"},
		{"class A { async foo(a) { return 1; } }", "async foo"},
	}
	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := runRequireAwait(t, requireAwaitCase{sourceText: testCase.sourceText})
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			span := result.Diagnostics[0].Range
			if got := testCase.sourceText[span.Pos():span.End()]; got != testCase.wantHead {
				t.Errorf("reported on %q, wanted %q", got, testCase.wantHead)
			}
		})
	}
}

// TestRequireAwaitExemptsAPromiseContract covers the case upstream cannot see.
//
// A function written into a position whose declared type demands a promise has to be async, and
// reporting it would be telling the author to make their code not compile. The exemption is only
// sound if it is narrow, so both directions are asserted here: the contract case stays silent, and
// the same body in a position that does NOT demand a promise still reports.
func TestRequireAwaitExemptsAPromiseContract(t *testing.T) {
	t.Parallel()

	// The stub matches a map whose declared signature returns a promise, exactly the shape a
	// not-yet-wired collector takes beside siblings that really do I/O.
	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
type CollectorType = (key: string) => Promise<number>;
const collectors: Record<string, CollectorType> = {
    wired: async function (key: string): Promise<number> {
        return Promise.resolve(key.length);
    },
    notWiredYet: async function (key: string): Promise<number> {
        return 0;
    },
};
void collectors;
`)
	rule_testing.ExpectClean(t, result)
}

// TestRequireAwaitStillReportsWithoutAContract is the control for the exemption above.
//
// A detector that cannot fire is indistinguishable from a clean corpus, so the same body written
// where nothing demands a promise has to still report. Without this the exemption could be
// swallowing every finding and the suite would look identical.
func TestRequireAwaitStillReportsWithoutAContract(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
async function notWiredYet(key: string): Promise<number> {
    return 0;
}
void notWiredYet;
`)
	rule_testing.ExpectFindings(t, result, "missingAwait")
}

// TestRequireAwaitReportsWhenTheContractAcceptsEither guards the narrowness of the exemption.
//
// A position typed `T | Promise<T>` gives the author a choice, so the keyword is optional there and
// the finding is real. This matters because that union is what a deliberately widened interface
// looks like, and widening is the correct repair when no implementation awaits. Exempting it would
// make the rule unable to see the very thing the widening was meant to expose.
func TestRequireAwaitReportsWhenTheContractAcceptsEither(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
type EitherType = (key: string) => number | Promise<number>;
const handler: EitherType = async function (key: string): Promise<number> {
    return key.length;
};
void handler;
`)
	rule_testing.ExpectFindings(t, result, "missingAwait")
}

// TestRequireAwaitDoesNotLetAGenericInferItsOwnContract guards the exemption against self-inference.
//
// A generic position's type parameter is inferred FROM the function being judged, so asking for the
// contextual type after inference reads the function's own `Promise` back as the position's demand,
// and every async callback handed to `map`, `useMemo`, `useCallback` or an identity helper exempted
// itself. Measured on ahra before the fix: `useMetricsExportMenu.tsx:74` and `:92`, an
// `onSelected: async function() {...}` inside `React.useMemo(() => [...])` that awaits nothing,
// were silent here and reported by ESLint, and ESLint was right. `[1, 2].map(async (n) => ...)` is
// the rule's classic true positive and was silent too. Each shape below reported nothing before the
// contextual type was asked for with inference from the node blocked.
func TestRequireAwaitDoesNotLetAGenericInferItsOwnContract(t *testing.T) {
	t.Parallel()

	shapes := []struct {
		name       string
		sourceText string
	}{
		{"Array.map", `
function sync(): number {
    return 1;
}
const values = [1, 2].map(async (n) => n + sync());
void values;
`},
		{"identity", `
function identity<T>(value: T): T {
    return value;
}
const handler = identity(async () => {
    return 1;
});
void handler;
`},
		{"useCallback", `
function useCallback<T extends Function>(callback: T, dependencies: unknown[]): T {
    void dependencies;
    return callback;
}
const handler = useCallback(async () => {
    return 1;
}, []);
void handler;
`},
		{"useMemo over an object literal", `
function useMemo<T>(factory: () => T, dependencies: unknown[]): T {
    void dependencies;
    return factory();
}
function exportFile(): void {}
const items = useMemo(function () {
    return [
        {
            label: 'Export',
            onSelected: async function () {
                exportFile();
            },
        },
    ];
}, []);
void items;
`},
	}

	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", shape.sourceText)
			rule_testing.ExpectFindings(t, result, "missingAwait")
		})
	}
}

// TestRequireAwaitKeepsAGenericContractThatDemandsAPromise is the control for the test above.
//
// Blocking inference from the node must not throw away a promise the position genuinely demands. A
// generic callback parameter typed `(api) => Promise<T>` still demands a promise when `T` cannot be
// inferred, and a type parameter constrained to a promise-returning function falls back to that
// constraint. Measured on ahra: `withDiscord<T>(callback: (api) => Promise<T>)` and its callers in
// `DiscordCommandLineInterface.ts` are the shape, and they must stay silent.
func TestRequireAwaitKeepsAGenericContractThatDemandsAPromise(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
async function withClient<T>(callback: (client: { name: string }) => Promise<T>): Promise<T> {
    return await callback({ name: 'client' });
}
function runTask<T extends () => Promise<void>>(task: T): T {
    return task;
}
void withClient(async function (client) {
    return client.name;
});
void runTask(async function () {
    return;
});
`)
	rule_testing.ExpectClean(t, result)

	// A method of a generic INTERFACE is an instantiated signature with no type parameters of its
	// own: `T` was fixed by the declared receiver, not inferred from this argument. Reading its
	// uninstantiated parameter type would lose the demand and report code that cannot drop the
	// keyword.
	interfaceMethod := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
interface RegistryInterface<T> {
    register(handler: T): void;
}
declare const registry: RegistryInterface<() => Promise<number>>;
registry.register(async () => {
    return 1;
});
`)
	rule_testing.ExpectClean(t, interfaceMethod)

	// Explicit type arguments are the author's, so nothing was inferred from the callback and the
	// instantiated parameter type is the real demand.
	explicit := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
function identity<T>(value: T): T {
    return value;
}
const handler = identity<() => Promise<number>>(async () => {
    return 1;
});
void handler;
`)
	rule_testing.ExpectClean(t, explicit)
}

// TestRequireAwaitReadsAGenericReturnFromItsTypedPosition covers the return-type stand-in.
//
// `const items: ItemType[] = useMemo(() => [...])` fixes `T` from the annotation as well as from
// the factory. When the annotated member returns a promise, removing the keyword breaks the
// assignment, so the demand is real and the case stays silent. When the annotated type does not
// declare the member at all, which is the ahra shape (`MenuItemButtonProperties[]` has no
// `onSelected`), nothing demands a promise and the finding stands.
func TestRequireAwaitReadsAGenericReturnFromItsTypedPosition(t *testing.T) {
	t.Parallel()

	prelude := `
function useMemo<T>(factory: () => T, dependencies: unknown[]): T {
    void dependencies;
    return factory();
}
function exportFile(): void {}
`

	demanded := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", prelude+`
const items: { label: string; onSelected: () => Promise<void> }[] = useMemo(function () {
    return [
        {
            label: 'Export',
            onSelected: async function () {
                exportFile();
            },
        },
    ];
}, []);
void items;
`)
	rule_testing.ExpectClean(t, demanded)

	// A factory with its own return annotation fixes the demand without the call's help, so the walk
	// stops there and the annotation answers.
	annotatedFactory := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", prelude+`
const items = useMemo(function (): { onSelected: () => Promise<void> }[] {
    return [
        {
            onSelected: async function () {
                exportFile();
            },
        },
    ];
}, []);
void items;
`)
	rule_testing.ExpectClean(t, annotatedFactory)

	undeclared := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", prelude+`
const items: { label: string }[] = useMemo(function () {
    return [
        {
            label: 'Export',
            onSelected: async function () {
                exportFile();
            },
        },
    ];
}, []);
void items;
`)
	rule_testing.ExpectFindings(t, undeclared, "missingAwait")
}

// TestRequireAwaitExemptsAnInterfaceMethod covers the class-method half of the contract check.
//
// `implements` puts the contract on the class rather than on the method, so the method has no
// contextual type of its own and the object-literal path above does not reach it. The control
// below asserts the exemption is not simply swallowing every method.
func TestRequireAwaitExemptsAnInterfaceMethod(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
interface AdapterInterface {
    fetchBalances: () => Promise<number[]>;
}
class ScaffoldedAdapter implements AdapterInterface {
    async fetchBalances(): Promise<number[]> {
        return [];
    }
}
void ScaffoldedAdapter;
`)
	rule_testing.ExpectClean(t, result)
}

// TestRequireAwaitReportsAMethodWithNoInterface is the control for the exemption above.
//
// The same method on a class that implements nothing has no contract to satisfy, so the keyword is
// the author's own choice and the finding stands.
func TestRequireAwaitReportsAMethodWithNoInterface(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, RequireAwait, "Subject.ts", `
class PlainAdapter {
    async fetchBalances(): Promise<number[]> {
        return [];
    }
}
void PlainAdapter;
`)
	rule_testing.ExpectFindings(t, result, "missingAwait")
}
