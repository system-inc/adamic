package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistencyRequireMatchingReturnTypeFile = "/repository/source/ConsistencyRequireMatchingReturnType.ts"

// consistencyRequireMatchingReturnTypePrelude declares every name the cases use, so the checker
// resolves each one and no verdict rests on an unresolved identifier (an unresolved name is the
// error type, which carries the Any flag and would read as unknowable).
//
// The React declarations mirror @types/react's shape rather than importing it: `EffectCallback`
// returns `void | Destructor`, which is the union a real effect is contextually typed by.
var consistencyRequireMatchingReturnTypePrelude = strings.Join([]string{
	"declare const ready: boolean;",
	"declare function work(): void;",
	"declare function name(): string;",
	"type Destructor = () => void;",
	"type EffectCallback = () => void | Destructor;",
	"declare function useEffect(effect: EffectCallback, dependencies?: unknown[]): void;",
	"declare function compute<T>(factory: () => T): T;",
	"declare function on(listener: (...values: any[]) => any): void;",
	"declare function mount(options: { onReady?: () => void; onClick?: (event: string) => void }): void;",
	"declare function choose(callback: () => string | undefined): void;",
	"type Shape = 'Circle' | 'Square';",
	"declare function forEachChild<T>(node: string, callback: (child: string) => T | undefined): T | undefined;",
	"declare function log(text: string): void;",
	"",
}, "\n")

func consistencyRequireMatchingReturnTypeSource(lines ...string) string {
	return consistencyRequireMatchingReturnTypePrelude + strings.Join(lines, "\n") + "\n"
}

type consistencyRequireMatchingReturnTypeCase struct {
	name  string
	lines []string
	// reported is the text each finding must span, in order.
	reported []string
	// fixed is the whole statement each fix must leave, or "" when the finding must carry no fix.
	fixed string
}

// The value direction: a bare `return;` where the type says the function returns a value.
var consistencyRequireMatchingReturnTypeBareCases = []consistencyRequireMatchingReturnTypeCase{
	{"declared string | undefined", []string{
		"export function find(): string | undefined {",
		"    if (!ready) return;",
		"    return name();",
		"}",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"async Promise<string | undefined> is judged by the awaited type", []string{
		"export async function load(): Promise<string | undefined> {",
		"    if (!ready) return;",
		"    return name();",
		"}",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"unannotated, mixing a bare return with a value, infers string | undefined", []string{
		"export function find() {",
		"    if (!ready) return;",
		"    return name();",
		"}",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"contextual from a generic factory, instantiated by the call", []string{
		"export const found = compute(function() {",
		"    if (!ready) return;",
		"    return name();",
		"});",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"contextual from a non-generic callback type", []string{
		"choose(function() {",
		"    if (!ready) return;",
		"    return name();",
		"});",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"a contextual any says nothing, so the inferred value type decides", []string{
		"on(function() {",
		"    if (!ready) return;",
		"    return name();",
		"});",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	{"declared undefined alone is a value", []string{
		"export function nothing(): undefined {",
		"    return;",
		"}",
	}, []string{"return;"}, "    return undefined;"},
	{"a getter is an ordinary function", []string{
		"export class Box {",
		"    get label(): string | undefined {",
		"        if (!ready) return;",
		"        return name();",
		"    }",
		"}",
	}, []string{"return;"}, "        if (!ready) return undefined;"},
	{"an unannotated getter is judged by the property type it infers", []string{
		"export class Box {",
		"    get label() {",
		"        if (!ready) return;",
		"        return name();",
		"    }",
		"}",
	}, []string{"return;"}, "        if (!ready) return undefined;"},
	{"an object-literal method takes its context", []string{
		"export const handlers: { pick(): string | undefined } = {",
		"    pick() {",
		"        if (!ready) return;",
		"        return name();",
		"    },",
		"};",
	}, []string{"return;"}, "        if (!ready) return undefined;"},
	{"a generator is judged by its TReturn", []string{
		"export function* count(): Generator<number, string | undefined> {",
		"    if (!ready) return;",
		"    yield 1;",
		"    return name();",
		"}",
	}, []string{"return;"}, "    if (!ready) return undefined;"},
	// Reported without a fix: `undefined` is not in `number`, so `return undefined;` would be a type
	// error. The bare return only compiles without noImplicitReturns, which this harness leaves off.
	{"a type that does not admit undefined reports without a fix", []string{
		"export function size(): number {",
		"    if (!ready) return;",
		"    return 1;",
		"}",
	}, []string{"return;"}, ""},
	// Reported without a fix: the comment would be eaten or displaced by a textual rewrite.
	{"a bare return holding a comment reports without a fix", []string{
		"export function find(): string | undefined {",
		"    if (!ready) return /* not yet */;",
		"    return name();",
		"}",
	}, []string{"return /* not yet */;"}, ""},
	// Reported without a fix: inserting ` undefined` after a `return` with no semicolon could join
	// the next line into the returned expression.
	{"a bare return with no semicolon reports without a fix", []string{
		"export function find(): string | undefined {",
		"    if (!ready) {",
		"        return",
		"    }",
		"    return name();",
		"}",
	}, []string{"return"}, ""},
	{"each return is judged against its own function, not the outer one", []string{
		"export function outer(): void {",
		"    const inner = function(): string | undefined {",
		"        if (!ready) return;",
		"        return name();",
		"    };",
		"    if (!inner()) return;",
		"}",
	}, []string{"return;"}, "        if (!ready) return undefined;"},
}

// The void direction: `return undefined;` where the type says `void`.
var consistencyRequireMatchingReturnTypeUndefinedCases = []consistencyRequireMatchingReturnTypeCase{
	{"declared void", []string{
		"export function save(): void {",
		"    if (!ready) return undefined;",
		"    work();",
		"}",
	}, []string{"return undefined;"}, "    if (!ready) return;"},
	{"async Promise<void> is judged by the awaited type", []string{
		"export async function save(): Promise<void> {",
		"    if (!ready) return undefined;",
		"    work();",
		"}",
	}, []string{"return undefined;"}, "    if (!ready) return;"},
	{"a React effect's early exit, contextually void | Destructor", []string{
		"useEffect(function() {",
		"    if (!ready) return undefined;",
		"    work();",
		"    return function() { work(); };",
		"});",
	}, []string{"return undefined;"}, "    if (!ready) return;"},
	{"an optional callback property, whose undefined member has no signature", []string{
		"mount({ onClick: function(event) {",
		"    if (!event) return undefined;",
		"    work();",
		"} });",
	}, []string{"return undefined;"}, "    if (!event) return;"},
	{"a constructor's early exit", []string{
		"export class Engine {",
		"    constructor() {",
		"        if (!ready) return undefined;",
		"        work();",
		"    }",
		"}",
	}, []string{"return undefined;"}, "        if (!ready) return;"},
	{"a setter's early exit", []string{
		"export class Engine {",
		"    set speed(value: number) {",
		"        if (!ready) return undefined;",
		"        work();",
		"    }",
		"}",
	}, []string{"return undefined;"}, "        if (!ready) return;"},
	{"a void-and-value union is still void", []string{
		"export function maybe(): string | void {",
		"    if (!ready) return undefined;",
		"    return name();",
		"}",
	}, []string{"return undefined;"}, "    if (!ready) return;"},
	// Reported without a fix: only the canonical text is rewritten.
	{"a parenthesized undefined reports without a fix", []string{
		"export function save(): void {",
		"    if (!ready) return (undefined);",
		"    work();",
		"}",
	}, []string{"return (undefined);"}, ""},
}

// Silent: the spelling already matches the type, or the type cannot say.
var consistencyRequireMatchingReturnTypeSilentCases = []struct {
	name  string
	lines []string
}{
	{"return undefined in a string | undefined function", []string{
		"export function find(): string | undefined {",
		"    if (!ready) return undefined;",
		"    return name();",
		"}",
	}},
	{"return undefined in an async Promise<string | undefined> function", []string{
		"export async function load(): Promise<string | undefined> {",
		"    if (!ready) return undefined;",
		"    return name();",
		"}",
	}},
	{"a bare return in a void function", []string{
		"export function save(): void {",
		"    if (!ready) return;",
		"    work();",
		"}",
	}},
	{"a bare return in an async Promise<void> function", []string{
		"export async function save(): Promise<void> {",
		"    if (!ready) return;",
		"    work();",
		"}",
	}},
	{"a bare return in an unannotated function with no value returns, which infers void", []string{
		"export function save() {",
		"    if (!ready) return;",
		"    work();",
		"}",
	}},
	// The case `consistent-return` cannot get right: the inferred type would be
	// `(() => void) | undefined`, a value, and only the contextual `void | Destructor` says void.
	{"a React effect exiting early and otherwise returning its cleanup", []string{
		"useEffect(function() {",
		"    if (!ready) return;",
		"    work();",
		"    return function() { work(); };",
		"});",
	}},
	{"an object-literal method contextually void through an optional property", []string{
		"mount({ onReady() {",
		"    if (!ready) return;",
		"    work();",
		"} });",
	}},
	{"a contextual any falls through to an inferred void", []string{
		"on(function() {",
		"    if (!ready) return;",
		"    work();",
		"});",
	}},
	// No bare return exists and the fall-off end is never judged: that is noImplicitReturns' job, and
	// here the checker proves the end unreachable anyway.
	{"an exhaustive switch with no default in a union-returning function", []string{
		"export function sides(shape: Shape): number {",
		"    switch (shape) {",
		"        case 'Circle':",
		"            return 0;",
		"        case 'Square':",
		"            return 4;",
		"    }",
		"}",
	}},
	{"a declared any cannot say which spelling is meant, either way", []string{
		"export function loose(): any {",
		"    if (!ready) return;",
		"    if (ready) return undefined;",
		"    return name();",
		"}",
	}},
	{"a declared unknown cannot say either", []string{
		"export function opaque(): unknown {",
		"    if (!ready) return;",
		"    return name();",
		"}",
	}},
	// Compiles only without noImplicitReturns, which this harness leaves off.
	{"a bare type parameter could be instantiated as void", []string{
		"export function pass<T>(value: T): T {",
		"    if (!ready) return;",
		"    return value;",
		"}",
	}},
	// A `never` function cannot reach a return without a type error, and there is no spelling to offer.
	{"a never function is unknowable", []string{
		"export function fail(): never {",
		"    if (!ready) return;",
		"    throw new Error('stop');",
		"}",
	}},
	// Inferred alone this is `string | undefined`, a value; only the contextual `() => void` says void,
	// so the case sees whether object-literal methods take their context.
	{"an object-literal method whose context is void while its returns infer a value", []string{
		"mount({ onReady() {",
		"    if (!ready) return;",
		"    return name();",
		"} });",
	}},
	{"an arrow React effect exiting early and otherwise returning its cleanup", []string{
		"useEffect(() => {",
		"    if (!ready) return;",
		"    work();",
		"    return () => work();",
		"});",
	}},
	{"an unannotated generator with only bare returns infers TReturn void", []string{
		"export function* count() {",
		"    if (!ready) return;",
		"    yield 1;",
		"}",
	}},
	{"a generator declared Iterable has an any TReturn", []string{
		"export function* count(): Iterable<number> {",
		"    if (!ready) return;",
		"    yield 1;",
		"}",
	}},
	// The ahra dispatcher shape: `return void log(...)` makes the inferred type `undefined`, which is
	// the spelling read back rather than a value anybody receives.
	{"an unannotated async function whose value returns are all void expressions", []string{
		"export async function dispatch(command: string) {",
		"    if (!command) return void log('usage');",
		"    if (command === 'list') {",
		"        work();",
		"        return;",
		"    }",
		"    work();",
		"}",
	}},
	// A generic context instantiated from the callback it types: `T | undefined` resolves from the
	// callback's own returns, the same circle as inference one call away.
	{"a generic context instantiated from the callback's own returns", []string{
		"forEachChild('node', function(child) {",
		"    if (!child) return;",
		"    work();",
		"});",
	}},
	// Both spellings inside one function whose inferred type is exactly `undefined`: unknowable, so
	// neither is reported.
	{"an unannotated function whose returns are all undefined is unknowable", []string{
		"export function save(flag: boolean) {",
		"    if (!ready) return undefined;",
		"    if (flag) return;",
		"    work();",
		"}",
	}},
	// The ahra shapes the void reading got wrong: `undefined` is the value here, and `return;` would
	// retype the callback as `void` and break `Array<undefined>` and the stub's consumer.
	{"a generic mapper instantiated as undefined from its own return", []string{
		"export const slots = compute(function() {",
		"    return undefined;",
		"});",
	}},
	{"a stub method returning undefined in an unannotated object", []string{
		"export const stub = {",
		"    lookup: function() {",
		"        return undefined;",
		"    },",
		"};",
	}},
	{"a constructor's bare early exit", []string{
		"export class Engine {",
		"    constructor() {",
		"        if (!ready) return;",
		"        work();",
		"    }",
		"}",
	}},
	{"a setter's bare early exit", []string{
		"export class Box {",
		"    #label = '';",
		"    set label(value: string) {",
		"        if (!ready) return;",
		"        this.#label = value;",
		"    }",
		"}",
	}},
	{"an arrow with an expression body has no return statement", []string{
		"export const pick = (): string | undefined => (ready ? name() : undefined);",
	}},
	// The value overload is first on purpose: with the void one first, ignoring the disagreement
	// would still land on void and stay silent, so the case could not see the guard.
	{"overloads that disagree on void against value are unknowable", []string{
		"export function either(flag: boolean): string | undefined;",
		"export function either(): void;",
		"export function either(flag?: boolean) {",
		"    if (!flag) return;",
		"    return name();",
		"}",
	}},
	// A `return` in a class static block is a grammar error, and the parser still hands back the
	// statement. It belongs to no function, so it must not be judged against the function enclosing
	// the class, whose type here would make it a value finding.
	{"a return in a class static block is not attributed to the enclosing function", []string{
		"export function outer(): string | undefined {",
		"    class Holder { static { if (!ready) return; work(); } }",
		"    return Holder.name;",
		"}",
	}},
	{"a void function returning the value of a void call", []string{
		"export function save(): void {",
		"    if (!ready) return work();",
		"    work();",
		"}",
	}},
}

func consistencyRequireMatchingReturnTypeAssertCase(t *testing.T, subject consistencyRequireMatchingReturnTypeCase, id string) {
	t.Helper()
	source := consistencyRequireMatchingReturnTypeSource(subject.lines...)
	result := rule_testing.RunTyped(t, ConsistencyRequireMatchingReturnType, consistencyRequireMatchingReturnTypeFile, source)
	want := make([]string, len(subject.reported))
	for index := range want {
		want[index] = id
	}
	rule_testing.ExpectFindings(t, result, want...)
	if len(result.Diagnostics) != len(subject.reported) {
		return
	}
	for index, diagnostic := range result.Diagnostics {
		reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != subject.reported[index] {
			t.Errorf("finding %d spans %q, want %q", index, reported, subject.reported[index])
		}
	}
	if subject.fixed == "" {
		for _, diagnostic := range result.Diagnostics {
			if len(diagnostic.Fixes) != 0 {
				t.Errorf("the finding must carry no fix, got %+v", diagnostic.Fixes)
			}
		}
		return
	}
	// The whole file after the fix, with exactly the one line changed, so a fix landing on the wrong
	// span cannot pass.
	fixedLines := append([]string{}, subject.lines...)
	changed := 0
	for index, line := range fixedLines {
		if strings.Contains(line, subject.reported[0]) && changed == 0 {
			fixedLines[index] = subject.fixed
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("the case names no line holding %q to fix", subject.reported[0])
	}
	rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(consistencyRequireMatchingReturnTypeSource(fixedLines...))+"\n")
}

func TestConsistencyRequireReturnMatchingTypeFiresOnBareReturnInValueFunction(t *testing.T) {
	t.Parallel()
	for _, subject := range consistencyRequireMatchingReturnTypeBareCases {
		t.Run(subject.name, func(t *testing.T) {
			t.Parallel()
			consistencyRequireMatchingReturnTypeAssertCase(t, subject, consistencyRequireMatchingReturnTypeBareId)
		})
	}
}

func TestConsistencyRequireReturnMatchingTypeFiresOnUndefinedReturnInVoidFunction(t *testing.T) {
	t.Parallel()
	for _, subject := range consistencyRequireMatchingReturnTypeUndefinedCases {
		t.Run(subject.name, func(t *testing.T) {
			t.Parallel()
			consistencyRequireMatchingReturnTypeAssertCase(t, subject, consistencyRequireMatchingReturnTypeUndefinedId)
		})
	}
}

func TestConsistencyRequireReturnMatchingTypeStaysSilent(t *testing.T) {
	t.Parallel()
	for _, subject := range consistencyRequireMatchingReturnTypeSilentCases {
		t.Run(subject.name, func(t *testing.T) {
			t.Parallel()
			source := consistencyRequireMatchingReturnTypeSource(subject.lines...)
			result := rule_testing.RunTyped(t, ConsistencyRequireMatchingReturnType, consistencyRequireMatchingReturnTypeFile, source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The message names the kind of function, the type, where the type came from, and what to write,
// asserted whole because the computed parts are where a rendering defect would live.
func TestConsistencyRequireReturnMatchingTypeMessages(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"value, declared", []string{
			"export async function load(): Promise<string | undefined> {",
			"    if (!ready) return;",
			"    return name();",
			"}",
		}, "This async function returns a value: its return type is `Promise<string | undefined>` (from its " +
			"annotation), so a bare `return;` hides that this path hands back `undefined`. Write `return " +
			"undefined;`. A bare return is reserved for functions whose type says `void`, so a reader can tell " +
			"at the return which kind of function this is."},
		{"void, contextual", []string{
			"useEffect(function() {",
			"    if (!ready) return undefined;",
			"    return function() { work(); };",
			"});",
		}, "This callback returns nothing: its return type is `void | Destructor` (from the signature it is " +
			"passed to), which says `void`, so `return undefined;` dresses an early exit up as a value. Write " +
			"`return;`. `return undefined;` is reserved for functions whose type includes `undefined` as a " +
			"value, so a reader can tell at the return which kind of function this is."},
	}
	for _, subject := range cases {
		t.Run(subject.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, ConsistencyRequireMatchingReturnType, consistencyRequireMatchingReturnTypeFile,
				consistencyRequireMatchingReturnTypeSource(subject.lines...))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != subject.want {
				t.Errorf("message\n got: %s\nwant: %s", got, subject.want)
			}
		})
	}
}

// The rule declares the checker and declines a file without one, rather than dereferencing nil.
func TestConsistencyRequireReturnMatchingTypeDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()
	if !ConsistencyRequireMatchingReturnType.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker")
	}
	if listeners := ConsistencyRequireMatchingReturnType.Run(rule.Context{}, nil); listeners != nil {
		t.Errorf("with no checker the rule must register no listeners, got %d", len(listeners))
	}
}
