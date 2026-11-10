package lower

import (
	"errors"
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
// dropped reds "another node"; and any use admitted reds "passed on".
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
		// A closure that writes node could run before the call, so node may not hold what the caller passed.
		{"a closure assigns node", "", overloadCallbackParameters, `    const reset = () => { node = undefined; };
    if (message === "reset") {
        reset();
    }
    check(node !== undefined && (test === undefined || test(node)), message || "Unexpected node.");
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
		name, source string
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
		// A rest of any promises nothing past the fixed positions, and tsc lets the same function stand there
		// (Compiler's road on #8n7eet3). Mutant: the any test dropped, and this row lowers.
		{"a number into a rest of any", animals + `function check<T extends Animal>(value: T, test: (value: T, ...rest: any[]) => boolean): boolean;
function check(value: Animal, test: ((value: Animal, extra: number) => boolean) | undefined): boolean {
	return test === undefined || test(value, 5);
}
console.log(String(check(new Dog(), (dog: Dog, label?: string): boolean => {
	if (label === undefined) { return true; }
	console.log(label.toUpperCase());
	return false;
})));
`, true},
		// A fixed parameter written any promises nothing at its own position either.
		{"an Animal into a parameter of any", animals + `function check<T extends Animal>(value: T, test: (value: any) => boolean): boolean;
function check(value: Animal, test: ((value: Animal) => boolean) | undefined): boolean {
	return test === undefined || test(new Animal());
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
