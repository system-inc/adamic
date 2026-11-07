package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// overloadRestNodes is a small node family in tsc's shape: tagged nodes on a parent chain, honest
// predicates on their kind, and findAncestor walking the chain with a callback that claims what it
// found. findAncestor is exported, as tsc's is, so its callback is a checked claim at every call
// through it rather than one every producer proves. isB is left to each source, so a lie can take its
// place.
const overloadRestNodes = `interface N { readonly kind: number; readonly parent: N | undefined; }
interface A extends N { readonly kind: 1; readonly a: string; }
interface B extends N { readonly kind: 2; readonly b: string; }
interface C extends N { readonly kind: 3; readonly c: string; }
function isA(n: N): n is A { return n.kind === 1; }
function isC(n: N): n is C { return n.kind === 3; }
export function findAncestor<T extends N>(node: N | undefined, callback: (element: N) => element is T): T | undefined {
  while (node !== undefined) {
    if (callback(node)) { return node; }
    node = node.parent;
  }
  return undefined;
}
`

const overloadRestHonestB = "function isB(n: N): n is B { return n.kind === 2; }\n"

// overloadRestOr is tsc's core.ts or, its three overloads and its implementation as tsc writes them.
const overloadRestOr = `export function or<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2;
export function or<P, R1 extends P, R2 extends P, R3 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2, f3: (p3: P) => p3 is R3): (p: P) => p is R1 | R2 | R3;
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U;
export function or<T extends unknown[], U>(...fs: ((...args: T) => U)[]): (...args: T) => U {
    return (...args) => {
        let lastResult: U;
        for (const f of fs) {
            lastResult = f(...args);
            if (lastResult) {
                return lastResult;
            }
        }
        return lastResult!;
    };
}
`

const overloadRestCalls = `const root: C = { kind: 3, parent: undefined, c: "root" };
const middle: B = { kind: 2, parent: root, b: "middle" };
const leaf: N = { kind: 9, parent: middle };
const two = findAncestor(leaf, or(isA, isB));
console.log(two === undefined ? "none" : String(two.kind));
const none = findAncestor(root, or(isA, isB));
console.log(none === undefined ? "none" : String(none.kind));
const three = findAncestor(root, or(isA, isB, isC));
console.log(three === undefined ? "none" : String(three.kind));
`

// overloadRestJudgments loads source as main.ts and judges each type predicate written on an overload's
// parameter, in source order, by the first two roads predicateRefusal tries: "trusted" where
// predicateParameter trusts it, "checked" where checkedFunctionTypeClaim makes it a checked claim, and
// "refused" where neither does, which leaves it to a body proof an overload's parameter doesn't have.
// Behavior can't say which road admitted a site, and the two mutants this file guards move a site from
// one road to another without changing whether the program lowers everywhere.
func overloadRestJudgments(t *testing.T, source string) string {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := loaded.Files()[0]
	checked, release := loaded.Checker(context.Background(), file)
	t.Cleanup(release)
	l := &lowering{program: loaded, checker: checked, result: &ir.Program{}}
	var judgments []string
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypePredicate && node.Parent.Kind == ast.KindFunctionType && node.Parent.Parent.Kind == ast.KindParameter {
			parameter := node.Parent.Parent
			if parameter.Parent.Kind == ast.KindFunctionDeclaration && parameter.Parent.Body() == nil {
				judgment := "refused"
				if l.predicateParameter(node) != nil {
					judgment = "trusted"
				} else if l.checkedFunctionTypeClaim(node) {
					judgment = "checked"
				}
				judgments = append(judgments, parameter.Name().Text()+" "+judgment)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	return strings.Join(judgments, ", ")
}

// overloadRestStops runs program on sanitized native and the JavaScript backend, each of which must stop
// at exit 70 with nothing printed and a predicate check's words, the same words on both.
func overloadRestStops(t *testing.T, program *ir.Program) {
	t.Helper()
	directory := t.TempDir()
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	var words []string
	for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var out, errs bytes.Buffer
		command.Stdout, command.Stderr = &out, &errs
		code := 0
		if err := command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		t.Logf("%s: exit %d, stderr %q", command.Path, code, errs.String())
		if code != 70 || out.Len() != 0 || !strings.HasPrefix(errs.String(), "adamic: panic: predicate ") {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want 70, nothing printed and a predicate check's words", command.Path, code, out.String(), errs.String())
		}
		words = append(words, errs.String())
	}
	if words[0] != words[1] {
		t.Fatalf("native stopped with %q and JavaScript with %q; want the same words", words[0], words[1])
	}
}

// tsc's core.ts or takes its predicates through overloads whose implementation holds every one in a
// rest of plain callbacks, `...fs: ((...args: T) => U)[]` (#1w5bve2). With its five parameter sites
// trusted, it stops next at the census's overload check, which maps the implementation's binders to the
// overload's by position when it has no more of them: T, the arguments' tuple, against P, the predicate's
// subject, which isn't an unknown[]. That's the wall this head measured on Pupa at 4df2fed1, before the
// implementation's f(...args) is ever read; the fix is the census's to make (inference over the whole
// signature, rest elements included), and this pin turns to agreement with Node when it lands.
func TestOrOverloadNextStop(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + overloadRestHonestB + overloadRestOr + overloadRestCalls
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "main.ts:15:20: Adamic 0.1 refuses overload 1 of or type parameter P cannot satisfy implementation parameter T; make implementation constraints accept every type admitted by the overload") {
		t.Fatalf("got %v, want a refusal ending %q", err, "main.ts:15:20: Adamic 0.1 refuses overload 1 of or type parameter P cannot satisfy implementation parameter T; make implementation constraints accept every type admitted by the overload")
	}
}

// overloadRestOrByPosition is or with an implementation whose binders rename the overload's by
// position, P, R1 and R2, holding its predicates in a rest of plain callbacks called with the subject
// alone, so the census's binder check passes and the rest rule is what's left to judge.
const overloadRestOrByPosition = `export function or<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2;
export function or<P, R1 extends P, R2 extends P>(...fs: ((p: P) => boolean)[]): (p: P) => boolean {
    return (p: P): boolean => {
        for (const f of fs) {
            if (f(p)) {
                return true;
            }
        }
        return false;
    };
}
`

// The same calls through or by position, honest: agrees with Node on the JavaScript backend and
// natively. If the census's overload-result check refuses the predicate-returning overload over a
// boolean-returning implementation, that's the wall after the binder check, measured here first.
func TestOrByPositionAgreesWithNode(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + overloadRestHonestB + overloadRestOrByPosition + `const root: C = { kind: 3, parent: undefined, c: "root" };
const middle: B = { kind: 2, parent: root, b: "middle" };
const leaf: N = { kind: 9, parent: middle };
const two = findAncestor(leaf, or(isA, isB));
console.log(two === undefined ? "none" : String(two.kind));
`
	if _, err := lowerSourceAs(t, source, ".ts"); err != nil {
		t.Logf("lowering or by position: %v", err)
		t.Fatal("or by position didn't lower")
	}
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "2\n")
	}
}

// A lie through or stops at a checked call before its answer is read. isB claims a kind-3 node is a B,
// and findAncestor(leaf, or(isA, isB)) reaches the kind-3 root, where isA says no and isB says yes; the
// result is then read as a B, which Node prints as undefined. A lie on a kind-1 node would never be
// heard: isA answers first for one of those, and or doesn't ask isB. Natively and on JavaScript the run
// stops at exit 70 with the same words. It goes through or by position, since tsc's or stops first at the
// census's binder check (TestOrOverloadNextStop).
func TestOrOverloadLieStopsAtTheCheckedCall(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + "function isB(n: N): n is B { return n.kind === 2 || n.kind === 3; }\n" + overloadRestOrByPosition + `const root: C = { kind: 3, parent: undefined, c: "root" };
const leaf: N = { kind: 9, parent: root };
const found = findAncestor(leaf, or(isA, isB));
if (found !== undefined && found.kind !== 1) { console.log(found.b); }
`
	if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); node.code != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node: exit %d, stdout %q; want 0 and the lie read as %q", node.code, node.stdout, "undefined\n")
	}
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Logf("lowering the lie through or: %v", err)
		t.Fatal("the lie through or didn't lower")
	}
	overloadRestStops(t, program)
}

// or's five overload parameters, f1 and f2 in its first overload and f1, f2 and f3 in its second, each
// land in the implementation's rest, whose element is a plain callback, so predicateParameter trusts
// every one, and none is refused as a claim with no body proving it. Whatever stops the program, if
// anything, isn't that. Mutant: the helper's rest branch dropped, which refuses all five as before.
func TestOrOverloadParametersClear(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + overloadRestHonestB + overloadRestOr + overloadRestCalls
	if got, want := overloadRestJudgments(t, source), "f1 trusted, f2 trusted, f1 trusted, f2 trusted, f3 trusted"; got != want {
		t.Fatalf("or's overload parameters: %s; want %s", got, want)
	}
	if _, err := lowerSourceAs(t, source, ".ts"); err != nil {
		t.Logf("lowering or: %v", err)
		if strings.Contains(err.Error(), "there is no body proving this parameter") {
			t.Fatalf("an overload parameter is still refused: %v", err)
		}
	}
}

// Trusting the parameter isn't trusting the result. keepAll's overload promises U[] from the claim its
// predicate makes, and its implementation keeps whatever every plain callback in its rest said yes to,
// a T[]. The parameter is trusted, as or's are, and the overload's result proof still judges the
// promise and refuses it.
func TestOverloadRestPlainResultStaysJudged(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + `function keepAll<T, U extends T>(xs: readonly T[], p: (x: T) => x is U): U[];
function keepAll<T>(xs: readonly T[], ...ps: ((x: T) => boolean)[]): T[] {
  const kept: T[] = [];
  for (const x of xs) {
    let all = true;
    for (const p of ps) { if (!p(x)) { all = false; } }
    if (all) { kept.push(x); }
  }
  return kept;
}
const root: C = { kind: 3, parent: undefined, c: "root" };
const a: A = { kind: 1, parent: root, a: "a" };
const nodes: readonly N[] = [root, a];
console.log(String(keepAll(nodes, isA).length));
`
	if got, want := overloadRestJudgments(t, source), "p trusted"; got != want {
		t.Fatalf("keepAll's overload parameter: %s; want %s", got, want)
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of keepAll result U[] cannot be served by implementation result T[]") {
		t.Fatalf("want keepAll's result refused by the overload's result proof, got %v", err)
	}
}

// A rest element block 2 doesn't read as a plain callback leaves the overload's parameter untrusted
// (#1w5bve2): a union of callables, which plainCallbackType keeps out written alone too; an element
// carrying a predicate of its own; and a type parameter, which holds whatever each call instantiates
// it as, not a written plain callback. A reference such as Array<X> isn't read through at all. The union,
// the type parameter and the reference are refused as claims with no body proving them. The element
// with its own predicate is a checked claim instead, through checkedFunctionTypeClaim's overload rule:
// every call the body makes through the element checks that element's claim. It was refused before an
// overload's parameter was read through a rest, and TestOverloadRestCheckedElement holds it end to end.
// Mutant mA, the plain-callback test dropped for a rest element, trusts every row but the reference,
// which the mapping never reads into an element.
func TestOverloadRestElementNotPlain(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, generic, implementation, call, want string }{
		{"union element", "", "...tests: (((x: N) => boolean) | ((x: N, y: number) => boolean))[]", "test(n, 0)", "test refused"},
		{"element with its own predicate", "", "...tests: ((x: N) => x is A)[]", "test(n)", "test checked"},
		{"generic element", "<F extends (x: N) => boolean>", "...tests: F[]", "test(n)", "test refused"},
		{"array reference", "", "...tests: Array<(x: N) => boolean>", "test(n)", "test refused"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := overloadRestNodes + "function anyOf(n: N, test: (x: N) => x is A): boolean;\nfunction anyOf" + probe.generic + "(n: N, " + probe.implementation + "): boolean {\n  for (const test of tests) { if (" + probe.call + ") { return true; } }\n  return false;\n}\n" + `const root: C = { kind: 3, parent: undefined, c: "root" };
console.log(String(anyOf(root, isA)));
`
			if got := overloadRestJudgments(t, source); got != probe.want {
				t.Fatalf("anyOf's overload parameter: %s; want %s", got, probe.want)
			}
			if probe.want != "test refused" {
				return
			}
			_, err := lowerSourceAs(t, source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "a type predicate whose return is not proven (there is no body proving this parameter)" {
				t.Fatalf("want the overload's claim refused, got %v", err)
			}
		})
	}
}

// A tuple rest holds each argument at its own position, so each overload parameter is judged by the
// tuple's element at that position, never by one merged across the tuple: two plain callbacks trust
// both parameters, and a second element claiming a predicate of its own leaves the second parameter
// to checkedFunctionTypeClaim's overload rule, as an implementation's parameter written out would. A
// named member reads as its type; an optional element is one a call can leave out, and it fails closed.
// Mutant mC, the tuple merged to its first element, trusts the second parameter in the second row.
func TestOverloadRestTuple(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, tuple, want string }{
		{"plain", "[(x: N) => boolean, (x: N) => boolean]", "first trusted, second trusted"},
		{"second claims", "[(x: N) => boolean, (x: N) => x is A]", "first trusted, second checked"},
		{"named", "[first: (x: N) => boolean, second: (x: N) => boolean]", "first trusted, second trusted"},
		{"optional second", "[(x: N) => boolean, ((x: N) => boolean)?]", "first trusted, second refused"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := overloadRestNodes + "function both(n: N, first: (x: N) => x is A, second: (x: N) => x is A): boolean;\nfunction both(n: N, ...tests: " + probe.tuple + "): boolean {\n  const second = tests[1];\n  return tests[0](n) && second !== undefined && second(n);\n}\n" + `const root: C = { kind: 3, parent: undefined, c: "root" };
console.log(String(both(root, isA, isA)));
`
			if got := overloadRestJudgments(t, source); got != probe.want {
				t.Fatalf("both's overload parameters: %s; want %s", got, probe.want)
			}
		})
	}
}

// predicateParameter and checkedFunctionTypeClaim read one mapping (overloadImplementationParameter),
// and this is where they'd answer apart if they didn't. pick's implementation holds both its overload's
// predicates in `...tests: ((x: N) => x is A)[]`, an element carrying a checked claim of its own, which
// plainCallbackType doesn't read as plain, so predicateParameter leaves both to
// checkedFunctionTypeClaim's overload rule. Read through the rest, both land on the element, whose
// claim checks every call the body makes through it, so both are checked claims and the program
// lowers; a lying arrow then stops at a check, natively and on JavaScript, where Node prints undefined.
// Mutant mB, checkedFunctionTypeClaim back on its own index loop, maps first to the rest's array type
// and second past the implementation's end, answers false for both, and leaves both refused as claims
// with no body proving them.
func TestOverloadRestCheckedElement(t *testing.T) {
	t.Parallel()
	source := overloadRestNodes + `function pick(n: N, first: (x: N) => x is A, second: (x: N) => x is A): A | undefined;
function pick(n: N, ...tests: ((x: N) => x is A)[]): A | undefined {
  for (const test of tests) {
    if (test(n)) { return n; }
  }
  return undefined;
}
const root: C = { kind: 3, parent: undefined, c: "root" };
const found = pick(root, isA, (x: N): x is A => true);
console.log(found === undefined ? "none" : found.a);
`
	if got, want := overloadRestJudgments(t, source), "first checked, second checked"; got != want {
		t.Fatalf("pick's overload parameters: %s; want %s", got, want)
	}
	if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); node.code != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node: exit %d, stdout %q; want 0 and the lie read as %q", node.code, node.stdout, "undefined\n")
	}
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want pick lowered with its claims checked, got %v", err)
	}
	overloadRestStops(t, program)
}

// A this parameter takes no argument, so an overload's predicate parameter is matched to the
// implementation's by the arguments a call passes, counted past a leading this on each side alone
// (Language's attack on 4df2fed1). With `this: void` on the overload only, or on the implementation
// only, `first` still lands in the implementation's plain callback and is trusted. Counted by
// Parameters() index, the first case ran off the implementation's end and the second met `n`.
func TestOverloadRestThisParameter(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"this on the overload only", `function pick(this: void, n: N, first: (x: N) => x is A): A | undefined;
function pick(n: N, first: (x: N) => boolean): A | undefined {
  return first(n) && n.kind === 1 ? (n as A) : undefined;
}
`},
		{"this on the implementation only", `function pick(n: N, first: (x: N) => x is A): A | undefined;
function pick(this: void, n: N, first: (x: N) => boolean): A | undefined {
  return first(n) && n.kind === 1 ? (n as A) : undefined;
}
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if got, want := overloadRestJudgments(t, overloadRestNodes+probe.source), "first trusted"; got != want {
				t.Fatalf("%s: %s; want %s", probe.name, got, want)
			}
		})
	}
}
