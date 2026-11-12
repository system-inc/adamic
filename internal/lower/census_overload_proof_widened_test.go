package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's getSupportedExtensions (utilities.ts:9993), cut to its proof: the first overload's caller passes
// no extra extensions, so `!extra || extra.length === 0` sends it to a return of the narrower constant,
// and the body proves overload 1's promise at the declaration (#z5d0jye step 2). Mutant: censusTrustedTest
// without its ! and || arms, and the declaration is refused.
func TestOverloadProofReadsAnOrOfTruthinessAndLength(t *testing.T) {
	t.Parallel()
	source := `type Extension = ".ts" | ".js";
const tsExtensions: readonly Extension[][] = [[".ts"]];
const allExtensions: readonly Extension[][] = [[".ts"], [".js"]];
interface Options { readonly allowJs?: boolean }
interface Extra { readonly extension: string }
function supported(options?: Options): readonly Extension[][];
function supported(options?: Options, extra?: readonly Extra[]): readonly string[][];
function supported(options?: Options, extra?: readonly Extra[]): readonly string[][] {
    const needJs = options !== undefined && options.allowJs === true;
    if (!extra || extra.length === 0) {
        return needJs ? allExtensions : tsExtensions;
    }
    return extra.map(x => [x.extension]);
}
const all: readonly Extension[][] = supported({ allowJs: true });
console.log(String(all.length));
console.log(String(supported(undefined).length));
console.log(String(supported(undefined, [{ extension: ".vue" }]).length));
`
	for _, extension := range []string{".ts", ".a"} {
		program := lowersAndAgreesWithNodeAs(t, source, extension)
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n1\n1\n" {
			t.Fatalf("%s native: exit %d, stdout %q, stderr %q; want 2, 1 and 1, as Node prints", extension, code, stdout, stderr)
		}
	}
}

// tsc's concatenate (core.ts:672): its honest overloads are proven through `b === undefined ||
// b.length === 0`, which narrows nothing but b.
const overloadConcatenateShape = `function concat<T>(a: T[], b: T[]): T[];
LIE
function concat<T>(a: readonly T[] | undefined, b: readonly T[] | undefined): readonly T[] | undefined;
function concat<T>(a: readonly T[] | undefined, b: readonly T[] | undefined): readonly T[] | undefined {
    if (b === undefined || b.length === 0) return a;
    if (a === undefined || a.length === 0) return b;
    return [...a, ...b];
}
const xs: number[] = concat([1], [2, 3]);
console.log(String(xs.length));
const ys: number[] = concat([4], []);
console.log(String(ys[0]));
`

func TestOverloadProofProvesConcatenate(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, strings.Replace(overloadConcatenateShape, "LIE", "", 1), ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "3\n4\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 3 and 4, as Node prints", code, stdout, stderr)
	}
}

// overloadProofRefused lowers source as extension and wants the named overload refused at its declaration.
func overloadProofRefused(t *testing.T, source, extension, what string) {
	t.Helper()
	_, err := lowerSourceAs(t, source, extension)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, what) {
		t.Fatalf("want %q refused, got %v", what, err)
	}
}

// Each promise here is one neither the call-site check (rulings.md:127, a literal, a tag or a brand) nor
// the presence check serves, so only the body proof could accept it.
func TestOverloadProofKeepsRefusing(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, extension, what string }{
		// tsc's concatenate overload 4: concat(undefined, []) returns undefined where it promises T[]. b's
		// test narrows nothing but b, so a stays possibly undefined at `return a`.
		{"concatenate's undefined first array", strings.Replace(overloadConcatenateShape, "LIE", "function concat<T>(a: T[] | undefined, b: T[]): T[];", 1), ".ts", "overload 2 of concat result T[] cannot be served"},
		// A call in the test stops the proof, a lying predicate's included. Mutant: censusTrustedTest
		// trusting any leaf, and the declaration is accepted.
		{"a predicate call in the test", `function isNumber(x: string | number): x is number { return x !== "never"; }
function only(x: string | number): number;
function only(x: string | number): string | number;
function only(x: string | number): string | number { if (isNumber(x)) return x; return 0; }
console.log(String(only("a")));
`, ".ts", "overload 1 of only result number cannot be served"},
		// A length with a literal part is a discriminant, so its comparison narrows: a tuple, an object
		// member like { length: 1 }, and a type parameter whose constraint has one. Mutant: censusLengthPlain
		// answering yes, and the One program is accepted.
		{"a tuple's length", `function head(xs: [string]): string;
function head(xs: [string] | [number, number]): string | number;
function head(xs: [string] | [number, number]): string | number { if (xs.length === 2) return 0; return "s"; }
console.log(String(head(["a"])));
`, ".ts", "overload 1 of head result string cannot be served"},
		{"a member whose length is a literal", `interface One { readonly length: 1; readonly first: string }
function first(x: One): string;
function first(x: One | readonly string[]): string | number;
function first(x: One | readonly string[]): string | number { if (x.length === 1) return "s"; return 0; }
console.log(String(first({ length: 1, first: "a" })));
`, ".ts", "overload 1 of first result string cannot be served"},
		{"a type parameter whose constraint's length is a literal", `function size<T extends { readonly length: 1 }>(x: T): string;
function size<T extends { readonly length: 1 }>(x: T): string | number;
function size<T extends { readonly length: 1 }>(x: T): string | number { if (x.length === 1) return "s"; return 0; }
console.log(String(size({ length: 1 })));
`, ".ts", "overload 1 of size result string cannot be served"},
		// A write inside the test is the body's, which the proof's scan refuses at any depth, and the leaves
		// the proof trusts can't hold one (Compiler's and Lower's condition on #z5d0jye step 2). In .a, where
		// a literal promise has no call-site check.
		{"an assignment in the test", "function one(x: number): 1;\nfunction one(x: number): number;\nfunction one(x: number): number { if ((x = 2) === 2) return 1; return 1; }\nconsole.log(String(one(3)));\n", ".a", "overload 1 of one result 1 cannot be served"},
		{"an increment in the test", "function one(x: number): 1;\nfunction one(x: number): number;\nfunction one(x: number): number { if (x++ > 0) return 1; return 1; }\nconsole.log(String(one(3)));\n", ".a", "overload 1 of one result 1 cannot be served"},
		{"a decrement in the test", "function one(x: number): 1;\nfunction one(x: number): number;\nfunction one(x: number): number { if (--x === 0) return 1; return 1; }\nconsole.log(String(one(3)));\n", ".a", "overload 1 of one result 1 cannot be served"},
		// A void promise over a result that owns something stays refused, so no built value is dropped
		// unreleased (Lower's condition).
		{"a void promise over a built string", `function name(n: number): void;
function name(n: number, full: boolean): string;
function name(n: number, full?: boolean): string { return "n" + String(n); }
name(1);
`, ".ts", "overload 1 of name result void cannot be served"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			overloadProofRefused(t, probe.source, probe.extension, probe.what)
		})
	}
}

// tsc's writeTokenText (emitter.ts:4949): overload 1 promises void over an implementation answering
// number. No read of a void value lowers (rulings.md:321), so the promise is served by any result, and
// a call statement keeps the implementation's representation. Mutant: overloadVoidPromise answering
// yes for every promise, and TestOverloadProofKeepsRefusing's programs are accepted.
const overloadVoidShape = `function write(text: string, out: (s: string) => void): void;
function write(text: string, out: (s: string) => void, position: number): number;
function write(text: string, out: (s: string) => void, position?: number): number {
    out(text);
    return position === undefined ? -1 : position + text.length;
}
const log = (s: string): void => { console.log(s); };
READ
`

func TestOverloadVoidPromiseIsServed(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadVoidShape, "READ", "write(\"ab\", log);\nconsole.log(String(write(\"cd\", log, 3)));", 1)
	for _, extension := range []string{".ts", ".a"} {
		program := lowersAndAgreesWithNodeAs(t, source, extension)
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "ab\ncd\n5\n" {
			t.Fatalf("%s native: exit %d, stdout %q, stderr %q; want ab, cd and 5, as Node prints", extension, code, stdout, stderr)
		}
	}
}

// The two places a void result is handed on rather than read: returned from a void function, and an
// arrow's body held as () => void (Lower's seam pins). Each agrees with Node natively and on JavaScript.
func TestOverloadVoidPromiseHandedOn(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadVoidShape, "READ", "function emit(): void { return write(\"x\", log); }\nemit();\nconst later: () => void = () => write(\"y\", log);\nlater();", 1)
	for _, extension := range []string{".ts", ".a"} {
		program := lowersAndAgreesWithNodeAs(t, source, extension)
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "x\ny\n" {
			t.Fatalf("%s native: exit %d, stdout %q, stderr %q; want x and y, as Node prints", extension, code, stdout, stderr)
		}
	}
}

// A read of the void promise's result is still refused, by :321's rule, never served.
func TestOverloadVoidPromiseReadStaysRefused(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadVoidShape, "READ", "const got = write(\"ab\", log);\nconsole.log(String(got));", 1)
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "a read of a void value") {
		t.Fatalf("want the void read refused, got %v", err)
	}
}
