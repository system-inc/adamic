package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
	"testing"
)

// overloadCallbackNodes is a small stand-in for tsc's Node, its predicates and its assert, so each program
// below is TypeScript Node runs. check takes a third, lazy message, as tsc's assert does.
const overloadCallbackNodes = `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function isId(n: N): n is Id { return n.kind === 1; }
function check(ok: unknown, m: string, detail?: () => string): void { if (!ok) { throw new Error(detail === undefined ? m : m + " " + detail()); } }
`

// overloadCallbackAssertNode is tsc's debug.ts assertNode over N, its two overloads and an implementation
// with parameters and body, so a probe can change either and keep the overloads.
func overloadCallbackAssertNode(parameters, body string) string {
	return overloadCallbackNodes + `function assertNode<T extends N, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string): asserts node is U;
function assertNode(node: N | undefined, test: ((node: N) => boolean) | undefined, message?: string): void;
function assertNode(` + parameters + `): void {
` + body + `}
`
}

const overloadCallbackParameters = `node: N | undefined, test: ((node: N) => boolean) | undefined, message?: string`

const overloadCallbackCall = `const id: Id = { kind: 1, text: "a" };
const n: N = id;
assertNode(n, isId);
console.log(n.text);
`

// tsc's assertNode: overload 1 promises a test of `(node: T) => node is U`, the implementation takes a test
// of any Node, and the census once refused that callback as strictly contravariant (#1w5bve2's block 3). The
// body calls test only with node, once node !== undefined made it the overload's T, and its message arrow
// captures node and never test, so the implementation serves the overload and agrees with Node.
func TestOverloadCallbackAssertNode(t *testing.T) {
	t.Parallel()
	source := overloadCallbackAssertNode(overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.", () => "kind " + String(node?.kind));
`) + overloadCallbackCall
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A message arrow that captures test and only tests its presence is rule 4's capture: it may outlive the call
// (check holds it until it fails), and it still never calls test, so the body lowers and agrees with Node.
// Language's pin on #1w5bve2: escape isn't what the rule rests on.
func TestOverloadCallbackPresenceCaptureLowers(t *testing.T) {
	t.Parallel()
	source := overloadCallbackAssertNode(overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.", () => test === undefined ? "with no test" : "kind " + String(node?.kind));
`) + overloadCallbackCall
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// tsc's assertOptionalNode, its two predicate overloads over a body that tests test and node for presence
// before it calls test with node: overload 1 passes a T, overload 2 a T | undefined that the body's own
// node === undefined took away. Called once with a node and once with undefined, each agrees with Node.
func TestOverloadCallbackAssertOptionalNode(t *testing.T) {
	t.Parallel()
	source := overloadCallbackNodes + `function assertOptionalNode<T extends N, U extends T>(node: T, test: (node: T) => node is U, message?: string): asserts node is U;
function assertOptionalNode<T extends N, U extends T>(node: T | undefined, test: (node: T) => node is U, message?: string): asserts node is U | undefined;
function assertOptionalNode(node: N | undefined, test: ((node: N) => boolean) | undefined, message?: string): void {
    check(test === undefined || node === undefined || test(node), message || "Unexpected node.");
}
const id: Id = { kind: 1, text: "a" };
const n: N = id;
assertOptionalNode(n, isId);
console.log(n.text);
function pick(first: boolean): N | undefined { return first ? n : undefined; }
const maybe = pick(false);
assertOptionalNode(maybe, isId);
console.log(maybe === undefined ? "none" : maybe.text);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A test that says yes to a node of kind 2 is a lie Node believes: it prints undefined for the text a kind-2
// node doesn't have. The callback's claim is checked, so the program stops at exit 70 on native and the
// JavaScript backend with nothing printed and a predicate check's words, before the caller reads the node as
// an Id. Which check fires first, liar's own claim inside assertNode's body or the overload's `asserts node
// is U` at the call, both say "predicate"; this pins the stop, not which of the two.
func TestOverloadCallbackLyingTestStops(t *testing.T) {
	t.Parallel()
	source := overloadCallbackAssertNode(overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`) + `function liar(n: N): n is Id { return true; }
const other: N = { kind: 2 };
assertNode(other, liar);
console.log(other.text);
`
	brandedStops(t, source, "adamic: panic: predicate ")
}

// Every use the rule doesn't follow keeps the census's refusal of overload 1's test, word for word. Mutants,
// each in overload_callback_use.go (scratchpad b3-mutants): rule 1 taking the overload's type through any
// narrowing reds "a narrowing"; rule 1 without its never-assigned check reds "a closure assigns node"; rule 4
// admitting a nested call reds "a closure calls test"; the argument check against promised's parameters
// dropped reds "another node"; and any use admitted reds "passed on". Since #4czdn4q the body lowers
// before the call, and in "through call" another rule stops there first, so that row reads the census's
// verdict (overloadVerdictOf), and the program still stops.
func TestOverloadCallbackRefusals(t *testing.T) {
	t.Parallel()
	const refusal = "overload 1 of assertNode parameter test cannot be served by implementation parameter test"
	for _, probe := range []struct{ name, prelude, parameters, body string }{
		// A node the overload's caller never passed isn't a T.
		{"another node", "", overloadCallbackParameters, `    const other: N = { kind: 2 };
    check(node !== undefined && (test === undefined || test(other)), message || "Unexpected node.");
`},
		// A copy carries its own type, N, not the overload's: copies aren't followed.
		{"a copy", "", overloadCallbackParameters, `    const m = node;
    check(m !== undefined && (test === undefined || test(m)), message || "Unexpected node.");
`},
		// isId narrowed node to Id, which says nothing about the overload's T.
		{"a narrowing", "", overloadCallbackParameters, `    check(node !== undefined && isId(node) && (test === undefined || test(node)), message || "Unexpected node.");
`},
		// The callback proof rejects any closure write to node, even when the explicit call follows check.
		{"a closure assigns node", "", overloadCallbackParameters, `    const reset = () => { node = undefined; };
    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
    if (message === "reset") {
        reset();
    }
`},
		// A default replaces a missing node with one the overload's caller never passed.
		{"a default", "const fallback: Id = { kind: 1, text: \"f\" };\n", "node: N | undefined = fallback, test: ((node: N) => boolean) | undefined, message?: string", `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`},
		// A default on test itself is a value the overload's caller never passed (Compiler's tightening on
		// #1w5bve2: `test = isNode` is an assignment in all but name).
		{"a default test", "", "node: N | undefined, test: ((node: N) => boolean) | undefined = isId, message?: string", `    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`},
		// A call inside a nested function is refused even when its argument would serve (rule 4).
		{"a closure calls test", "", overloadCallbackParameters, `    const later = () => node !== undefined && (test === undefined || test(node));
    check(later(), message || "Unexpected node.");
`},
		// Passed on, as tsc's assertEachNode passes test to every, the callback is called by code the walk
		// doesn't read (rule 5). assertNode's shape keeps the other parameters as the census already serves them.
		{"passed on", "function every<E>(array: readonly E[], callback: (element: E) => boolean): boolean { for (const element of array) { if (!callback(element)) { return false; } } return true; }\n", overloadCallbackParameters, `    check(node !== undefined && (test === undefined || every([node], test)), message || "Unexpected node.");
`},
		// .call reads a member of the callback rather than calling it directly (rule 5).
		{"through call", "", overloadCallbackParameters, `    check(node !== undefined && (test === undefined || test.call(undefined, node)), message || "Unexpected node.");
`},
		// arguments reads every parameter without naming one.
		{"arguments read", "", overloadCallbackParameters, `    check(arguments.length > 0 && node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := probe.prelude + overloadCallbackAssertNode(probe.parameters, probe.body) + overloadCallbackCall
			if probe.name == "through call" {
				overloadVerdictRefused(t, source, ".ts", "assertNode", 1, refusal)
				return
			}
			_, err := lowerSourceAs(t, source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, refusal) {
				t.Fatalf("want %q, got %v", refusal, err)
			}
		})
	}
}

// An argument past the promised callback's parameters lands in the caller's function, which tsc lets declare
// a further optional parameter the promise never showed: `(dog: Dog, label?: string) => boolean` stands
// where `(value: T) => boolean` is promised. So the implementation's `test(value, 5)` would hand label a
// number, and `test(value, new Animal())` would hand a Dog parameter an Animal, each a TypeError on Node.
// Both keep the census's refusal (Language's must on #8n7eet3, x1 and x2). Mutant: an extra argument
// skipped as evaluated and ignored, and both lower.
func TestOverloadCallbackExtraArgumentRefuses(t *testing.T) {
	t.Parallel()
	const animals = `class Animal { readonly name: string = "animal"; }
class Dog extends Animal { bark(): string { return "woof"; } }
`
	const refusal = "overload 1 of check parameter test cannot be served by implementation parameter test"
	for _, probe := range []struct {
		name, source   string
		typeScriptOnly bool
	}{
		{"a number into an optional string", animals + `function check<T extends Animal>(value: T, test: (value: T) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, extra: number) => boolean) | undefined): boolean {
	return test === undefined || test(value, 5);
}
console.log(String(check(new Dog(), (dog: Dog, label?: string): boolean => {
	if (label === undefined) { return true; }
	console.log(label.toUpperCase());
	return false;
})));
`, false},
		{"an Animal into an optional Dog", animals + `function check<T extends Animal>(value: T, test: (value: T) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, other: Animal) => boolean) | undefined): boolean {
	return test === undefined || test(value, new Animal());
}
console.log(String(check(new Dog(), (dog: Dog, other?: Dog): boolean => {
	if (other === undefined) { return true; }
	console.log(other.bark());
	return false;
})));
`, false},
		// A rest of any promises nothing past the fixed positions: tsc lets `(dog: Dog, label?: string)` stand
		// there, and the implementation's 5 relates to any (Compiler's road on #8n7eet3). That caller is held
		// today by an earlier NotYet at its own argument, so this row's caller takes no extra parameter and
		// reaches the census, which must still refuse the extra. Mutant: the any test dropped, and it lowers.
		{"a number into a rest of any", animals + `function check<T extends Animal>(value: T, test: (value: T, ...rest: any[]) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, extra: number) => boolean) | undefined): boolean {
	return test === undefined || test(value, 5);
}
console.log(String(check(new Dog(), (dog: Dog): boolean => dog.bark() === "woof")));
`, true},
		// A fixed parameter written any promises nothing at its own position either. Where any is the whole
		// callback's only difference the census relates the pair before the use-proof runs, so this row's T at
		// position 0 sends it to the use-proof, which must refuse what lands at the any. The caller takes one
		// parameter, so nothing at its own argument is judged against any first. Mutant: the any test
		// dropped, and it lowers.
		{"an Animal into a parameter of any", animals + `function check<T extends Animal>(value: T, test: (value: T, other: any) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, other: Animal) => boolean) | undefined): boolean {
	return test === undefined || test(value, new Animal());
}
console.log(String(check(new Dog(), (dog: Dog): boolean => dog.bark() === "woof")));
`, true},
	} {
		// .a refuses a written any where it's written, before the census reads it.
		extensions := []string{".ts", ".a"}
		if probe.typeScriptOnly {
			extensions = extensions[:1]
		}
		for _, extension := range extensions {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, refusal) {
					t.Fatalf("want %q, got %v", refusal, err)
				}
			})
		}
	}
}

// overloadCallbackAssertEachNode is tsc's debug.ts assertEachNode over N as Correctness adapts it (Adamic's
// ruling on #8n7eet3, Nov 10): the body loops over nodes itself rather than passing test to every. Of its
// four predicate overloads, the two over NodeArray<T> are left out, since a NodeArray holds tsc's pos, end,
// hasTrailingComma and transformFlags and an element no wider than a number (NodeArrayLayout), and an N is
// an object; the two over readonly T[] and readonly T[] | undefined stand for them, beside the overload that
// takes any N. stackCrawlMark is left out as assertNode's twin leaves it: `stackCrawlMark || assertEachNode`
// reads an overload set as a value, which stage 0 stops. array is "readonly " for tsc's shape; a mutable
// array's overload is refused first at nodes itself, by the census's own invariance, so the twins that write
// keep it readonly and write through a cast. before runs ahead of the presence tests, and loop inside them.
func overloadCallbackAssertEachNode(array, before, loop string) string {
	return overloadCallbackNodes + `function shouldAssert(): boolean { return true; }
function assertEachNode<T extends N, U extends T>(nodes: ` + array + `T[], test: (node: T) => node is U, message?: string): asserts nodes is ` + array + `U[];
function assertEachNode<T extends N, U extends T>(nodes: ` + array + `T[] | undefined, test: (node: T) => node is U, message?: string): asserts nodes is ` + array + `U[] | undefined;
function assertEachNode(nodes: ` + array + `N[], test: ((node: N) => boolean) | undefined, message?: string): void;
function assertEachNode(nodes: ` + array + `N[] | undefined, test: ((node: N) => boolean) | undefined, message?: string): void {
    if (shouldAssert()) {
        let passed = true;
` + before + `        if (test !== undefined && nodes !== undefined) {
` + loop + `        }
        check(passed, message || "Unexpected node.", () => ` + "`Node array did not pass test.`" + `);
    }
}
`
}

// overloadCallbackEachFor is the adapted loop with head between its parentheses and inside ahead of its
// test, so a twin changes one thing.
func overloadCallbackEachFor(head, inside string) string {
	return `            for (` + head + `) {
` + inside + `                if (!test(n)) {
                    passed = false;
                    break;
                }
            }
`
}

// overloadCallbackEachCall passes overload 1 an array of Ids typed as N, then reads each element's text,
// which only the assertion lets it name.
func overloadCallbackEachCall(array string) string {
	return `const a: Id = { kind: 1, text: "a" };
const b: Id = { kind: 1, text: "b" };
const ids: ` + array + `N[] = [a, b];
assertEachNode(ids, isId);
for (const node of ids) {
    console.log(node.text);
}
`
}

// The adapted assertEachNode calls test only with n, a const bound by a for-of over nodes itself, and the
// body reads nodes nowhere else but in presence tests, so n holds the overload's caller's T and the
// implementation serves overloads 1 and 2. Called with Ids, with an empty array and with undefined, each
// agrees with Node.
func TestOverloadCallbackAssertEachNode(t *testing.T) {
	t.Parallel()
	source := overloadCallbackAssertEachNode("readonly ", "", overloadCallbackEachFor("const n of nodes", "")) + overloadCallbackEachCall("readonly ") + `const none: readonly N[] = [];
assertEachNode(none, isId);
console.log(String(none.length));
function pick(present: boolean): readonly N[] | undefined { return present ? ids : undefined; }
const missing = pick(false);
assertEachNode(missing, isId);
console.log(missing === undefined ? "none" : String(missing.length));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A test that says yes to a node of kind 2 inside the loop is the same lie TestOverloadCallbackLyingTestStops
// pins for assertNode, and it stops the same way, at exit 70 with nothing printed and a predicate check's
// words, before the caller reads an element as an Id.
func TestOverloadCallbackEachLyingTestStops(t *testing.T) {
	t.Parallel()
	source := overloadCallbackAssertEachNode("readonly ", "", overloadCallbackEachFor("const n of nodes", "")) + `function liar(n: N): n is Id { return true; }
const other: N = { kind: 2 };
const others: readonly N[] = [other];
assertEachNode(others, liar);
for (const node of others) {
    console.log(node.text);
}
`
	brandedStops(t, source, "adamic: panic: predicate ")
}

// Every loop the for-of rule doesn't follow keeps the census's refusal of overload 1's test, word for
// word: one row per edge of overloadCallbackOnlyIterated's allowlist and per lie. Mutants, each in
// overload_callback_use.go (scratchpad s2-mutants): any loop source accepted reds "a copy" and "another
// array"; overloadCallbackOnlyIterated dropped reds every row from "an index write in the loop" to "a
// length read"; the const check dropped reds "a let reassigned"; and the binding's flow check dropped reds
// "a narrowing". Since #4czdn4q the body lowers before the call, and in "for await" another rule stops there
// first, so that row reads the census's verdict (overloadVerdictOf), and the program still stops.
func TestOverloadCallbackEachRefusals(t *testing.T) {
	t.Parallel()
	const refusal = "overload 1 of assertEachNode parameter test cannot be served by implementation parameter test"
	const other = "        const other: N = { kind: 2 };\n"
	const writeRow = "a write "
	twin := func(prelude, array, before, loop string) string {
		return prelude + overloadCallbackAssertEachNode(array, before, loop) + overloadCallbackEachCall(array)
	}
	for _, probe := range []struct{ name, source string }{
		// A copy is another array, and the caller's elements in it say nothing about what else it holds.
		{"a copy", twin("", "readonly ", "", overloadCallbackEachFor("const n of nodes.slice()", ""))},
		// An array the overload's caller never passed holds no T.
		{"another array", twin("", "readonly ", "", `            const extra: readonly N[] = [{ kind: 2 }];
`+overloadCallbackEachFor("const n of extra", ""))},
		// An element written while the loop runs is one the overload's caller never passed. tsc's array is
		// readonly, so a write goes through a cast, and a mutable overload is refused first at nodes itself.
		{"a write by index in the loop", twin("", "readonly ", other, overloadCallbackEachFor("const n of nodes", "                (nodes as N[])[0] = other;\n"))},
		// So is one pushed, which the loop then reaches.
		{"a write by push in the loop", twin("", "readonly ", other, overloadCallbackEachFor("const n of nodes", "                (nodes as N[]).push(other);\n"))},
		// A write before the loop is outside it, and still lands in what the loop reads.
		{"a write before the loop", twin("", "readonly ", other+`        if (nodes !== undefined) {
            (nodes as N[])[0] = other;
        }
`, overloadCallbackEachFor("const n of nodes", ""))},
		// A call that takes the array could write it, whatever its parameter's type says.
		{"passed to a call", twin("function count(array: readonly N[]): number { return array.length; }\n", "readonly ", "", "            passed = count(nodes) >= 0;\n"+overloadCallbackEachFor("const n of nodes", ""))},
		// An alias is the same array under a name the rule doesn't follow.
		{"an alias", twin("", "readonly ", "", "            const q = nodes;\n"+overloadCallbackEachFor("const n of q", ""))},
		// A nested function that reads nodes hands it out of the body's sight.
		{"a nested function reads nodes", twin("", "readonly ", "        const held = (): readonly N[] | undefined => nodes;\n", "            passed = held() !== undefined;\n"+overloadCallbackEachFor("const n of nodes", ""))},
		// A member read is a member read, length or push: one allowlist, with no list of harmless members.
		{"a length read", twin("", "readonly ", "", "            passed = nodes.length >= 0;\n"+overloadCallbackEachFor("const n of nodes", ""))},
		// A let can be given another value inside the loop, which the flow type alone doesn't show: n is
		// still N after `n = other`.
		{"a let reassigned", twin("", "readonly ", other, overloadCallbackEachFor("let n of nodes", "                n = other;\n"))},
		// A value made from the binding is a new object, not the caller's element.
		{"a value derived", twin("", "readonly ", "", `            for (const n of nodes) {
                const m: N = { kind: n.kind + 1 };
                if (!test(m)) {
                    passed = false;
                    break;
                }
            }
`)},
		// isId narrowed n to Id, which says nothing about the overload's T.
		{"a narrowing", twin("", "readonly ", "", `            for (const n of nodes) {
                if (isId(n) && !test(n)) {
                    passed = false;
                    break;
                }
            }
`)},
		// for await hands n each element awaited, another value when an element is thenable. Only an async
		// implementation may loop so, so its overloads promise a Promise rather than assert.
		{"for await", overloadCallbackNodes + `function assertEachNode<T extends N, U extends T>(nodes: readonly T[], test: (node: T) => node is U, message?: string): Promise<void>;
function assertEachNode<T extends N, U extends T>(nodes: readonly T[] | undefined, test: (node: T) => node is U, message?: string): Promise<void>;
async function assertEachNode(nodes: readonly N[] | undefined, test: ((node: N) => boolean) | undefined, message?: string): Promise<void> {
    let passed = true;
    if (test !== undefined && nodes !== undefined) {
        for await (const n of nodes) {
            if (!test(n)) {
                passed = false;
                break;
            }
        }
    }
    check(passed, message || "Unexpected node.");
}
const a: Id = { kind: 1, text: "a" };
const ids: readonly N[] = [a];
await assertEachNode(ids, isId);
console.log("checked");
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if probe.name == "for await" {
				overloadVerdictRefused(t, probe.source, ".ts", "assertEachNode", 1, refusal)
				return
			}
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if errors.As(err, &refused) && strings.Contains(refused.What, refusal) {
				return
			}
			// A write to tsc's readonly array goes through a cast, and stage 0 stops that cast before the
			// census reaches the overload. Only the three write rows may stop there; once the cast lowers,
			// the allowlist refuses them as it refuses the member read and the pass on.
			var notYet *NotYet
			if strings.HasPrefix(probe.name, writeRow) && errors.As(err, &notYet) && strings.Contains(notYet.Error(), "a cast that makes a viewed readonly array writable") {
				return
			}
			t.Fatalf("want %q, got %v", refusal, err)
		})
	}
}

// An overload parameter written any admits whatever its caller passes, so an argument or a loop element
// typed through it promises the callback nothing: `assertNode(cat, isDog)` type-checks against
// `(node: any, test: (node: T) => node is U)`, and the body would hand isDog a cat. Both rule 1 and the for-of
// binding keep the flow type there, and the census refuses, at run's call, the one that resolves to the
// overload (#4czdn4q), which stops before any argument is lowered, so no argument-site stop comes first.
// Each row's place is measured (J-7bkh7xq). .ts only, since .a refuses a written any where it's written. Mutants:
// the any test dropped in rule 1 (g1) reddens "rule 1"; dropped for the element (g2) reddens "a for-of" if
// the checker gives an index type of any for any.
func TestOverloadCallbackAnyOverloadParameterRefuses(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, refusal, where string }{
		{"rule 1", overloadCallbackNodes + `function assertNode<T extends N, U extends T>(node: any, test: (node: T) => node is U): void;
function assertNode(node: N | undefined, test: ((node: N) => boolean) | undefined): void {
    check(node !== undefined && (test === undefined || test(node)), "Unexpected node.");
}
function run(n: N): void { assertNode(n, isId); }
console.log("declared");
`, "overload 1 of assertNode parameter test cannot be served by implementation parameter test", "main.ts:9:28"},
		{"a for-of", overloadCallbackNodes + `function assertEachNode<T extends N, U extends T>(nodes: any, test: (node: T) => node is U): void;
function assertEachNode(nodes: readonly N[] | undefined, test: ((node: N) => boolean) | undefined): void {
    let passed = true;
    if (test !== undefined && nodes !== undefined) {
        for (const n of nodes) {
            if (!test(n)) {
                passed = false;
                break;
            }
        }
    }
    check(passed, "Unexpected node.");
}
function run(nodes: readonly N[]): void { assertEachNode(nodes, isId); }
console.log("declared");
`, "overload 1 of assertEachNode parameter test cannot be served by implementation parameter test", "main.ts:18:43"},
		// An array of any holds any in its type argument, which the walk reads though it skips a library
		// type's members.
		{"a for-of over an array of any", overloadCallbackNodes + `function assertEachNode<T extends N, U extends T>(nodes: readonly any[], test: (node: T) => node is U): void;
function assertEachNode(nodes: readonly N[] | undefined, test: ((node: N) => boolean) | undefined): void {
    let passed = true;
    if (test !== undefined && nodes !== undefined) {
        for (const n of nodes) {
            if (!test(n)) {
                passed = false;
                break;
            }
        }
    }
    check(passed, "Unexpected node.");
}
function run(nodes: readonly N[]): void { assertEachNode(nodes, isId); }
console.log("declared");
`, "overload 1 of assertEachNode parameter test cannot be served by implementation parameter test", "main.ts:18:43"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.refusal) || !strings.HasSuffix(refused.Where, probe.where) {
				t.Fatalf("want %q at %s, got %v", probe.refusal, probe.where, err)
			}
		})
	}
}

// TestOverloadCallbackArrayProof pins the new roads separately from the older
// for-of tests. It uses only the JavaScript and source agreement helper.
func TestOverloadCallbackArrayProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		lowers       bool
	}{
		{"number-index", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"const-offset", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const start = 0;
for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i + start];
    if (node !== undefined) {
        return cb === undefined || cb(node);
    }
}
return false;

}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"let-offset", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const start = 0;
for (let i = 0; i < nodes.length; i++) {
    let node = nodes[i + start];
    if (node !== undefined) {
        return cb === undefined || cb(node);
    }
}
return false;

}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"not-element", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const other: N = { kind: 2 }; return cb === undefined || cb(other);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"string-index", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const key = "0"; return cb === undefined || cb(nodes[key]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"array-reassigned", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
nodes = []; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"forward", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const start = 0;
for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i + start];
    if (node !== undefined) {
        return cb === undefined || cb(node);
    }
}
return false;
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"forward-for-of", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
for (const node of nodes) if (cb !== undefined) return cb(node); return false;
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"forward-write", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
(nodes as N[])[0] = {kind: 2}; return cb === undefined || cb(nodes[0]!);
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"forward-push", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
(nodes as N[]).push({kind: 2}); return cb === undefined || cb(nodes[0]!);
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"forward-recursion", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"forward-spread", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return cb === undefined || cb(nodes[0]!);
}
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const pair: [readonly N[]] = [nodes]; return worker(...pair, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"callback-stored", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const held = cb; return held === undefined || held(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"truthiness", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
return cb ? cb(nodes[0]!) : false;
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"false-member", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | false): boolean {
return cb ? cb(nodes[0]!) : false;
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"and-truthiness", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
if (cb && cb(nodes[0]!)) return true; return false;
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"length", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
if (nodes.length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"identity", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const other: readonly N[] = []; if (other === (nodes as readonly N[])) return false; if (other !== nodes) { return cb === undefined || cb(nodes[0]!); } return false;
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"loose-identity", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const other: readonly N[] = []; if (other == nodes) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"library-slice", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const copy = nodes.slice(0, 1); if (copy.length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"interface-slice", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
interface OwnSlice { slice(): number; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const copy = nodes.slice(0, 1); if (copy.length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"class-slice", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
class OwnSlice extends Array<N> { slice(start?: number, end?: number): N[] { return []; } }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const copy = nodes.slice(0, 1); if (copy.length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"copy-element", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const copy = nodes.slice(0, 1); return cb === undefined || cb(copy[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"return-array", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): readonly N[];
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): readonly N[] {
if (cb !== undefined) cb(nodes[0]!); return nodes;
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: readonly N[] = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(String(answer.length));
`, true},
		{"nested-return", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const read = (): readonly N[] => nodes; if (read().length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"handoff", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
observe(nodes, kindTest); return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"handoff-push", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    (nodes as N[]).push({ kind: 2 });
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
observe(nodes, kindTest); return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"handoff-recursion", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    observe(nodes, test);
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
observe(nodes, kindTest); return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		// Rule 5 refuses storing the callback in a module variable.
		{"handoff-stored", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
let saved: ((node: N) => boolean) | undefined;
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
saved = cb; observe(nodes, kindTest); return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		// Rule 4 refuses invoking the callback from the handed-off closure.
		{"handoff-captured", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
observe(nodes, (node: N): boolean => cb === undefined || cb(node)); return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, false},
		{"worker-shape", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
let updated: N[] | undefined;
const start = 0;
const count = nodes.length;
for (let i = 0; i < count; i++) {
    const node = nodes[i + start];
    const visited = node !== undefined ? (cb ? cb(node) : true) : false;
    if (!visited && node !== undefined) {
        if (updated === undefined) updated = nodes.slice(0, i);
        updated.push(node);
    }
}
if (updated !== undefined) return false;
observe(nodes, kindTest);
return true;

}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"visit-array-shape", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function observe(nodes: readonly N[], test: (node: N) => boolean): void {
    for (const node of nodes) {
        if (!test(node)) break;
    }
}
function kindTest(node: N): boolean { return node.kind > 0; }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
let updated: N[] | undefined;
const start = 0;
const count = nodes.length;
for (let i = 0; i < count; i++) {
    const node = nodes[i + start];
    const visited = node !== undefined ? (cb ? cb(node) : true) : false;
    if (!visited && node !== undefined) {
        if (updated === undefined) updated = nodes.slice(0, i);
        updated.push(node);
    }
}
if (updated !== undefined) return false;
observe(nodes, kindTest);
return true;
}
function visit<T extends N, A extends readonly T[] | undefined>(nodes: A, cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[] | undefined, cb: ((node: N) => boolean) | undefined): boolean {
if (nodes === undefined) return false; const count = nodes.length; if (count === 0) return false; return worker(nodes, cb);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit<Id, readonly Id[]>(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`, true},
		{"visit-nodes-shape", `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
interface NodeArray<T extends N> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
function worker(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): readonly N[] {
    if (cb !== undefined && nodes.length !== 0) cb(nodes[0]!);
    return nodes;
}
function visit<T extends N, A extends NodeArray<T> | undefined>(nodes: A, cb: (node: T) => boolean): readonly N[] | undefined;
function visit(nodes: NodeArray<N> | undefined, cb: ((node: N) => boolean) | undefined): readonly N[] | undefined {
if (nodes === undefined) return nodes;
const length = nodes.length;
const end = nodes.end;
const start = nodes.pos;
const trailing = nodes.hasTrailingComma;
const updated = worker(nodes, cb);
if (updated !== (nodes as readonly N[])) return updated;
if (length === 0 || trailing || end < start) return updated;
return nodes;

}
export function use(nodes: NodeArray<Id>): number {
    const answer: readonly N[] | undefined = visit<Id, NodeArray<Id>>(nodes, (node: Id): boolean => node.text === "id");
    return answer === undefined ? 0 : answer.length;
}
console.log("declared");
`, true},
		{"reduce-left", `function reduceLeft<T, U>(array: readonly T[], f: (memo: U, value: T, i: number) => U, initial: U): U;
function reduceLeft<T>(array: readonly T[], f: (memo: T, value: T, i: number) => T): T | undefined;
function reduceLeft<T>(array: readonly T[], f: (memo: T | undefined, value: T, i: number) => T, initial?: T): T | undefined {
    if (array.length === 0) return undefined;
    let memo: T | undefined = arguments.length <= 2 ? array[0]! : initial;
    for (let i = 1; i < array.length; i++) memo = f(memo, array[i]!, i);
    return memo;
}
const result: number | undefined = reduceLeft([1, 2, 3], (memo, value) => memo + value);
if (result !== undefined) console.log(String(result));
`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if probe.lowers {
				lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
				return
			}
			switch probe.name {
			case "reduce-left":
				overloadVerdictRefused(t, probe.source, ".ts", "reduceLeft", 2,
					"overload 2 of reduceLeft parameter f cannot be served by implementation parameter f")
				return
			case "forward-write", "forward-push", "handoff-push", "loose-identity", "class-slice":
				// These programs can stop before the overload walk. Pin its
				// own verdict and also require the ordinary program to stop.
				checkArrayCallbackCensus(t, probe.source)
				return
			}
			overloadVerdictRefused(t, probe.source, ".ts", "visit", 1,
				"overload 1 of visit parameter cb cannot be served by implementation parameter cb")
		})
	}
}

// Some source restrictions run before overload census at a normal entry.
// Ask that census directly to pin the reason this proof refuses as well.
func checkArrayCallbackCensus(t *testing.T, source string) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	defer release()
	l := &lowering{program: program, checker: checked, entry: entry, result: &ir.Program{}}
	var overload, implementation *ast.Node
	for _, node := range entry.Statements.Nodes {
		if node.Kind != ast.KindFunctionDeclaration || node.Name().Text() != "visit" {
			continue
		}
		if node.Body() == nil {
			overload = node
		} else {
			implementation = node
		}
	}
	if implementation == nil || overload == nil {
		t.Fatal("missing visit overload and implementation")
	}
	var stopped *Refused
	err = l.censusOverload(implementation, overload, 1)
	const refusal = "overload 1 of visit parameter cb cannot be served by implementation parameter cb"
	if !errors.As(err, &stopped) || stopped.What != refusal {
		t.Fatalf("want %q, got %v", refusal, err)
	}
	if _, err := lowerSourceAs(t, source, ".ts"); err == nil {
		t.Fatal("program unexpectedly lowered")
	}
}

// The property obligation asks the same allocation graph as accessor
// dispatch. The two edges differ only in where the owning allocation goes.
func TestOverloadCallbackArrayProofGetterReach(t *testing.T) {
	t.Parallel()
	for _, owns := range []bool{false, true} {
		function := &ast.Node{}
		refusal := &Refused{What: "overload 1 of visit parameter cb cannot be served by implementation parameter cb"}
		owner := ir.ObjectLiteral{Class: 1}
		receiver := ir.ObjectLiteral{}
		if owns {
			receiver = owner
		}
		program := &ir.Program{
			Locals:    []ir.Local{{Type: ir.Object}},
			Classes:   []ir.Class{{Accessors: []ir.Accessor{{Name: "pos", Getter: -1, Setter: -1}}}},
			Functions: []ir.Function{{Parameters: []int{0}}},
			Main:      []ir.Statement{ir.Evaluate{Value: owner}, ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{receiver}}}},
		}
		l := &lowering{result: program, functionNodes: map[int]*ast.Node{0: function},
			callbackReadUses: []overloadCallbackReadProof{{reads: []overloadCallbackRead{{function, 0, "pos"}}, refusal: refusal}}}
		got := l.refuseReachedCallbackReads(l.newAccessorReach(program), map[string]bool{"pos": true})
		if owns && got != refusal || !owns && got != nil {
			t.Errorf("owning allocation reaches parameter %v, got %v", owns, got)
		}
	}
}

const overloadCallbackNodeSlice = `import { basename } from "node:path";
interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
function visit<T extends N>(nodes: readonly T[], cb: (node: T) => boolean): boolean;
function visit(nodes: readonly N[], cb: ((node: N) => boolean) | undefined): boolean {
const copy = nodes.slice(0, 1); if (copy.length === 0) return false; return cb === undefined || cb(nodes[0]!);
}
const nodes: readonly Id[] = [{ kind: 1, text: "id" }];
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
console.log(basename("folder/example"));
`

func TestOverloadCallbackArrayProofNodeSlice(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", overloadCallbackNodeSlice)})
	if err != nil {
		t.Fatal(err)
	}
	var declaresSlice ast.Visitor
	declaresSlice = func(node *ast.Node) bool {
		if node.Parent != nil && node.Parent.Kind == ast.KindInterfaceDeclaration && node.Name() != nil {
			if name, _ := memberKeyText(node); name == "slice" {
				return true
			}
		}
		return node.ForEachChild(declaresSlice)
	}
	for _, file := range program.SourceFiles() {
		if load.IsNodeLibrary(file) && file.AsNode().ForEachChild(declaresSlice) {
			t.Logf("bundled slice declaration in %s", file.FileName().AsString())
			lowersAndAgreesWithNodeAs(t, overloadCallbackNodeSlice, ".ts")
			return
		}
	}
	t.Fatal("the program did not load a bundled Node declaration of slice")
}

const overloadCallbackGetterFixture = `interface N { readonly kind: number }
interface Id extends N { readonly kind: 1; readonly text: string }
interface NodeArray<T extends N> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
class Holder { get pos(): number { return 7; } }
function visit<T extends N>(nodes: NodeArray<T>, cb: (node: T) => boolean): boolean;
function visit(nodes: NodeArray<N>, cb: ((node: N) => boolean) | undefined): boolean {
    if (nodes.pos === 0) return false;
    return cb === undefined || cb(nodes[0]!);
}
console.log(String(new Holder().pos));
`

func TestOverloadCallbackArrayProofDisjointGetter(t *testing.T) {
	t.Parallel()
	source := overloadCallbackGetterFixture + `
const item: Id = { kind: 1, text: "id" };
const nodes = [item] as NodeArray<Id>;
nodes.pos = 1;
const answer: boolean = visit(nodes, (node: Id): boolean => node.text === "id");
console.log(answer ? "true" : "false");
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

func TestOverloadCallbackArrayProofUnknownGetter(t *testing.T) {
	t.Parallel()
	source := overloadCallbackGetterFixture + `
export function use(nodes: NodeArray<Id>): boolean {
    return visit(nodes, (node: Id): boolean => node.text === "id");
}
`
	if err := overloadVerdictOf(t, source, ".ts", "visit", 1); err != nil {
		t.Fatalf("census must serve the overload: %v", err)
	}
	const want = "overload 1 of visit parameter cb cannot be served by implementation parameter cb"
	var refused *Refused
	_, err := lowerSourceAs(t, source, ".ts")
	if !errors.As(err, &refused) || refused.What != want || !strings.HasSuffix(refused.Where, "main.ts:5:50") {
		t.Fatalf("want %q at the overload parameter, got %v", want, err)
	}
}
