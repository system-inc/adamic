package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A call ends a proof's path only when its callee can't be replaced and its one body never returns
// (#8n7eet3, Adamic and Compiler, Nov 10): the callee names, bare or through a namespace, a symbol whose
// one declaration is a function declaration with a body (neverCall). The flow prover ends a path at such
// a call as at the prelude's panic, so debug.ts's assertIsDefined, whose failing branch calls fail, proves
// its claim by its body, and the calls semver.ts makes through checkDefined lower with no check where
// main stopped at "checked predicate overload target <T>". These pins hold that shape proven and agreeing
// with Node, through a value and through the uncaught "Debug Failure." stop, and hold every lying twin to
// its check: a never helper that returns on one path, an object-literal method, a const bound to an
// arrow, a method reached through an object or a class instance, and a namespace's function reached
// through a parameter typed as the namespace. Mutants, as patch files beside the brief: the flow prover
// ending a path at any call statement (n1) reddens the void helper's lowering and every twin's proof;
// neverCall taking a property access on any object (n2) reddens the parameter twin, the only one whose
// one declaration is a function declaration; main's neverCall restored (n2b) reddens the method twins;
// a const's arrow read as its declaration (n3) reddens the arrow twin.

// neverCallDebug is debug.ts's assertIsDefined and checkDefined inside namespace Debug, as tsc declares
// them without stackCrawlMark, with helper declared beside them and call in assertIsDefined's failing
// branch, where tsc calls fail.
func neverCallDebug(helper, call string) string {
	return `namespace Debug {
	` + helper + `

	export function assertIsDefined<T>(value: T, message?: string): asserts value is NonNullable<T> {
		if (value === undefined || value === null) {
			` + call + `;
		}
	}

	export function checkDefined<T>(value: T | null | undefined, message?: string): T {
		assertIsDefined(value, message);
		return value;
	}
}

`
}

// neverCallFail is debug.ts's fail, which builds its Error and throws it.
const neverCallFail = `export function fail(message?: string): never {
		const e = new Error(message ? "Debug Failure. " + message : "Debug Failure.");
		throw e;
	}`

// neverCallRun calls checkDefined as semver.ts does, on tryParseComponents' Version shape and on
// parseRange's Comparator[][], and prints fields of each result.
const neverCallRun = `interface Comparator {
	readonly operator: string;
	readonly operand: number;
}

function parse(text: string) {
	return text === "" ? undefined : { major: text.length, minor: 0, patch: 0, prerelease: "", build: text };
}

function rows(text: string): Comparator[][] | undefined {
	if (text === "") {
		return undefined;
	}
	return [[{ operator: ">=", operand: text.length }], [{ operator: "<", operand: 2 }]];
}

const version = Debug.checkDefined(parse("1.2.3"), "Invalid version");
console.log(version.build + " " + String(version.major));
const ranges = Debug.checkDefined(rows(">=1"), "Invalid range");
for (const row of ranges) {
	for (const comparator of row) {
		console.log(comparator.operator + String(comparator.operand));
	}
}
`

// neverCallMissing calls checkDefined on a string, then on an undefined, whose fail nothing catches.
const neverCallMissing = `function find(text: string): string | undefined {
	return text === "" ? undefined : text;
}

console.log(Debug.checkDefined(find("found")));
console.log(Debug.checkDefined(find("")));
console.log("unreached");
`

// neverCallVersionZero is the core of tsc's Version.zero (tsc_version_zero.ts on #1451npc's branch, at
// 096155e1): debug.ts's fail, assert, assertIsDefined and checkDefined as stage 3 adapts them, each
// template literal spelled as a concatenation, and semver.ts's Version with its static zero, reduced to
// the major, minor and patch its constructor checks, with tryParseComponents reduced to the same three.
const neverCallVersionZero = `import 'node:process';

type AnyFunction = (...args: never[]) => void;

namespace Debug {
	export function fail(message?: string, stackCrawlMark?: AnyFunction): never {
		debugger;
		const e = new Error(message ? "Debug Failure. " + message : "Debug Failure.");
		if (Error.captureStackTrace) {
			Error.captureStackTrace(e, stackCrawlMark || fail);
		}
		throw e;
	}

	export function assert(expression: unknown, message?: string, stackCrawlMark?: AnyFunction): asserts expression {
		if (!expression) {
			message = message ? "False expression: " + message : "False expression.";
			fail(message, stackCrawlMark || assert);
		}
	}

	export function assertIsDefined<T>(value: T, message?: string, stackCrawlMark?: AnyFunction): asserts value is NonNullable<T> {
		if (value === undefined || value === null) {
			fail(message, stackCrawlMark || assertIsDefined);
		}
	}

	export function checkDefined<T>(value: T | null | undefined, message?: string, stackCrawlMark?: AnyFunction): T {
		assertIsDefined(value, message, stackCrawlMark || checkDefined);
		return value;
	}
}

class Version {
	static readonly zero: Version = new Version(0, 0, 0);

	readonly major: number;
	readonly minor: number;
	readonly patch: number;

	constructor(text: string);
	constructor(major: number, minor?: number, patch?: number);
	constructor(major: number | string, minor = 0, patch = 0) {
		if (typeof major === "string") {
			const result = Debug.checkDefined(tryParseComponents(major), "Invalid version");
			({ major, minor, patch } = result);
		}

		Debug.assert(major >= 0, "Invalid argument: major");
		Debug.assert(minor >= 0, "Invalid argument: minor");
		Debug.assert(patch >= 0, "Invalid argument: patch");

		this.major = major;
		this.minor = minor;
		this.patch = patch;
	}
}

function tryParseComponents(text: string) {
	const parts = text.split(".");
	if (parts.length !== 3) return undefined;
	return { major: parseInt(parts[0]!, 10), minor: parseInt(parts[1]!, 10), patch: parseInt(parts[2]!, 10) };
}

console.log(String(Version.zero.major));
console.log(String(new Version("1.2.3").minor));
`

// neverCallKinds is the probe's two tagged members, which isA tells apart.
const neverCallKinds = `interface A {
	readonly kind: "a";
	readonly x: number;
}

interface B {
	readonly kind: "b";
	readonly s: string;
}

`

// neverCallNamespace is a namespace whose fail throws, a member nothing can replace.
const neverCallNamespace = `namespace Debug {
	export function fail(): never {
		throw new Error("no");
	}
}
`

// neverCallIsA is the probe's isA, with helper declared before it and call in the branch that isn't an A,
// then a call on an A whose x it prints.
func neverCallIsA(helper, call string) string {
	return neverCallKinds + helper + `
function isA(v: A | B): v is A {
	if (v.kind !== "a") {
		` + call + `;
	}
	return true;
}

function pick(n: number): A | B {
	return n > 0 ? { kind: "a", x: n } : { kind: "b", s: "b" };
}

const v = pick(1);
if (isA(v)) {
	console.log(String(v.x));
}
`
}

// neverCallProofs loads source as a file with extension, without lowering it, and asks both provers about
// function's claim: the flow prover (proveFlowPredicate) and the verifier (provePredicate), each nil when
// it proves the claim. A pin on the proof alone reads past whatever later stop the program meets.
func neverCallProofs(t *testing.T, source, extension, function string) (flow, body error) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	var annotation *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindTypePredicate && n.Parent.Name() != nil && n.Parent.Name().Text() == function {
			annotation = n
		}
		n.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if annotation == nil {
		t.Fatalf("no type predicate on %s", function)
	}
	_, body = l.provePredicate(annotation)
	return l.proveFlowPredicate(annotation), body
}

// neverCallProven fails t unless program calls function and every direction of every call is proven, so
// no check runs. Behavior can't tell a proven call from a checked one whose check passes, so the sites
// PredicateChecks records are read.
func neverCallProven(t *testing.T, program *ir.Program, function string) {
	t.Helper()
	calls := 0
	for _, site := range program.PredicateChecks.Sites {
		if site.Function != function {
			continue
		}
		calls++
		for _, direction := range site.Directions {
			if direction.Status != "proven" {
				t.Fatalf("%s at %s: its %s direction is %s, want proven (%s)", function, site.Where, direction.Direction, direction.Status, direction.Reason)
			}
		}
	}
	if calls == 0 {
		t.Fatalf("no call of %s in PredicateChecks: %+v", function, program.PredicateChecks)
	}
}

// neverCallKeepsCheck lowers source as main.ts and fails t unless function's claim keeps its check: the
// lowering stops at a checked predicate's NotYet, as a generic target does on main, or every call of
// function checks a direction.
func neverCallKeepsCheck(t *testing.T, source, function string) {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.Error(), "checked predicate") {
			t.Fatalf("got %v, want a checked predicate's NotYet or a program that checks %s", err, function)
		}
		return
	}
	calls := 0
	for _, site := range program.PredicateChecks.Sites {
		if site.Function != function {
			continue
		}
		calls++
		checked := false
		for _, direction := range site.Directions {
			checked = checked || direction.Status == "checked"
		}
		if !checked {
			t.Fatalf("%s at %s: no direction is checked (%+v), want its claim checked", function, site.Where, site.Directions)
		}
	}
	if calls == 0 {
		t.Fatalf("no call of %s in PredicateChecks: %+v", function, program.PredicateChecks)
	}
}

// tsc's shape: fail is a function declaration in the namespace, so assertIsDefined's body proves its claim,
// and checkDefined on the Version shape and on Comparator[][] lowers with every call of assertIsDefined
// proven, agreeing with Node on the JavaScript backend and natively. On an undefined, the uncaught fail
// stops all three alike, at exit 70 with Node's "Debug Failure." after the panic prefix.
func TestNeverCallProvesAssertIsDefined(t *testing.T) {
	t.Parallel()
	debug := neverCallDebug(neverCallFail, "fail(message)")
	t.Run("proof", func(t *testing.T) {
		t.Parallel()
		if flow, _ := neverCallProofs(t, debug+neverCallRun, ".ts", "assertIsDefined"); flow != nil {
			t.Fatalf("want assertIsDefined's body to prove its claim, got %v", flow)
		}
	})
	t.Run("checkDefined", func(t *testing.T) {
		t.Parallel()
		const want = "1.2.3 5\n>=3\n<2\n"
		program := lowersAndAgreesWithNodeAs(t, debug+neverCallRun, ".ts")
		neverCallProven(t, program, "assertIsDefined")
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != want || stderr != "" {
			t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", code, stdout, stderr, want)
		}
	})
	t.Run("checkDefined of undefined", func(t *testing.T) {
		t.Parallel()
		const stderr = "adamic: panic: Error: Debug Failure.\n"
		program := agreeEntry(t, writeAgreementSource(t, "main.ts", debug+neverCallMissing), nil, 70, nil, nil)
		neverCallProven(t, program, "assertIsDefined")
		if code, stdout, got := brandedNative(t, program); code != 70 || stdout != "found\n" || got != stderr {
			t.Fatalf("native: exit %d stdout %q stderr %q; want exit 70, %q and %q, as Node stops", code, stdout, got, "found\n", stderr)
		}
	})
}

// The core of tsc's Version.zero: assertIsDefined's body proves its claim through fail(message,
// stackCrawlMark || assertIsDefined), and assert's still proves its own through the same fail. Lowering
// stops where main stops the real file (at 92:38 there), the static zero's construction escaping before
// the static fields are initialized: #1451npc lifts that, and isn't on main. Refusals are judged before
// lowering's NotYets, so the static stop comes first whatever this change proves, and the proofs are read
// without lowering. Past #1451npc the next stop is no longer assertIsDefined's checked target, where its
// branch stopped the real file at 73:2.
func TestNeverCallVersionZeroCore(t *testing.T) {
	t.Parallel()
	if flow, _ := neverCallProofs(t, neverCallVersionZero, ".ts", "assertIsDefined"); flow != nil {
		t.Fatalf("want assertIsDefined's body to prove its claim, got %v", flow)
	}
	if _, body := neverCallProofs(t, neverCallVersionZero, ".ts", "assert"); body != nil {
		t.Fatalf("want assert's body to prove its condition, got %v", body)
	}
	_, err := lowerSourceAs(t, neverCallVersionZero, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a constructor object escaping before static fields are initialized" || !strings.HasSuffix(refused.Where, "main.ts:35:38") {
		t.Fatalf("got %v, want main's static construction stop at main.ts:35:38", err)
	}
}

// Each twin puts a never helper in assertIsDefined's failing branch that the body can't trust, so neither
// prover proves the claim and every call keeps its check: on main's generic target, its NotYet. A helper
// that returns on one path is typed never only through a cast, which is read here as proof alone; the
// same helper typed void, which isn't never, lowers to the check. An object-literal method and a const
// bound to an arrow can each be replaced, so their bodies prove nothing.
func TestNeverCallTwinsKeepTheirCheck(t *testing.T) {
	t.Parallel()
	for _, twin := range []struct {
		name, helper, call string
		lowered            bool
	}{
		{"a never helper that returns on one path", `export function warn(m?: string): never {
		if (m === "x") {
			throw new Error(m);
		}
		return undefined as never;
	}`, "warn(message)", false},
		{"a helper typed void", `export function warn(m?: string): void {
		if (m === "x") {
			throw new Error(m);
		}
	}`, "warn(message)", true},
		{"an object-literal method", `const H = { fail(m?: string): never { throw new Error(m); } };`, "H.fail(message)", true},
		{"a const bound to an arrow", `const fail2 = (m?: string): never => { throw new Error(m); };`, "fail2(message)", true},
	} {
		t.Run(twin.name, func(t *testing.T) {
			t.Parallel()
			source := neverCallDebug(twin.helper, twin.call) + neverCallRun
			if flow, body := neverCallProofs(t, source, ".ts", "assertIsDefined"); flow == nil || body == nil {
				t.Fatalf("want assertIsDefined unproven by both provers, got flow %v and body %v", flow, body)
			}
			if twin.lowered {
				neverCallKeepsCheck(t, source, "assertIsDefined")
			}
		})
	}
}

// The probe's isA, whose failing branch calls fail through an object-literal method or a class instance's
// method, either of which can be replaced through an alias, stays unproven: .a refuses it, and in .ts its
// call checks its claim. Main proved it through the method's body.
func TestNeverCallMethodTwinsStayUnproven(t *testing.T) {
	t.Parallel()
	for _, twin := range []struct{ name, helper string }{
		{"an object-literal method", `const H = { fail(): never { throw new Error("no"); } };
`},
		{"a class instance's method", `class F {
	fail(): never {
		throw new Error("no");
	}
}

const H = new F();
`},
	} {
		t.Run(twin.name, func(t *testing.T) {
			t.Parallel()
			source := neverCallIsA(twin.helper, "H.fail()")
			if flow, body := neverCallProofs(t, source, ".a", "isA"); flow == nil || body == nil {
				t.Fatalf("want isA unproven by both provers, got flow %v and body %v", flow, body)
			}
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "a type predicate whose return is not proven (true return narrows to A | B, not A)" {
				t.Fatalf("got %v, want .a's refusal of isA's unproven return", err)
			}
			neverCallKeepsCheck(t, source, "isA")
		})
	}
}

// The same isA failing through a namespace's fail, a member nothing can replace, is proven by its flow and
// lowers in .a, agreeing with Node. Reached through a parameter typed as the namespace, the same function
// declaration is a value a caller chooses, so the call proves nothing.
func TestNeverCallNamespaceMember(t *testing.T) {
	t.Parallel()
	t.Run("through the namespace", func(t *testing.T) {
		t.Parallel()
		source := neverCallIsA(neverCallNamespace, "Debug.fail()")
		if flow, _ := neverCallProofs(t, source, ".a", "isA"); flow != nil {
			t.Fatalf("want isA proven by its flow, got %v", flow)
		}
		lowersAndAgreesWithNode(t, source)
	})
	t.Run("through a parameter", func(t *testing.T) {
		t.Parallel()
		source := neverCallKinds + neverCallNamespace + `
function isA(v: A | B, d: typeof Debug): v is A {
	if (v.kind !== "a") {
		d.fail();
	}
	return true;
}
`
		if flow, body := neverCallProofs(t, source, ".a", "isA"); flow == nil || body == nil {
			t.Fatalf("want isA unproven by both provers, got flow %v and body %v", flow, body)
		}
	})
	// The proof rests on nothing replacing Debug.fail. tsc's TS2630 names only a bare name, so a write
	// through the namespace is held by lowering, which never rebinds a function declaration; if that write
	// ever lowered, a proof through Debug.fail would read a body the call doesn't run.
	for _, extension := range []string{".ts", ".a"} {
		t.Run("a write through the namespace stops "+extension, func(t *testing.T) {
			t.Parallel()
			source := neverCallIsA(neverCallNamespace, "Debug.fail()") + `Debug.fail = function (): never {
	throw new Error("replaced");
};
`
			_, err := lowerSourceAs(t, source, extension)
			if err == nil {
				t.Fatal("want the write to Debug.fail refused or not yet lowered, but the program lowered")
			}
			t.Logf("stopped by: %v", err)
		})
	}
}
