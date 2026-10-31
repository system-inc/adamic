package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A branded supertype, tsc's TypeElement shape: a wide kind and a brand field typed any, over subtypes
// each with a literal kind. Identifier isn't one, since it lacks the brands.
const brandedPrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: any; }
interface TypeElement extends Declaration { _typeElementBrand: any; readonly name?: string; }
interface PropertySignature extends TypeElement { readonly kind: 171; readonly name: string; readonly typeText: string; }
interface MethodSignature extends TypeElement { readonly kind: 173; readonly name: string; readonly count: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function kindOf(n: Node): number { return n.kind; }
function isTypeElement(n: Node): n is TypeElement { const k = kindOf(n); return k === 171 || k === 173; }
function isTypeElementLie(n: Node): n is TypeElement { return kindOf(n) !== 0; }
`

// A checked predicate claiming a branded supertype checks at the call only that the value is an object,
// and every field where it's read, through S's view and its subtypes' (#c7w6j0y change B, ruled Oct 31).
// An honest value agrees with Node whatever its kind, a value whose kind no subtype names included, as
// tsc's isLiteralExpression takes JsxText and isLateBound a declaration whose name it refines. A lie stops
// at its first bad read on sanitized native and the JavaScript backend, however it got its type: read
// through S, stored or captured and read later, cast or passed as a subtype, or narrowed to a subtype by
// an asserts call, which no cast covers, so only the subtypes' views stop it. A write of the wrong type
// stops at the write. Mutants: no view registered, which reads every lie raw but the cast and the subtype
// parameter, which the cast's own view holds too; S's view alone, which reads the asserts lie raw (the
// JavaScript backend prints undefined, native calls it a compiler bug); one of S's fields dropped from the
// checked set, which reads the layout, stored and captured lies raw.
func TestBrandedSupertypeIsCheckedByItsViews(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"honest", `function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } else { console.log("no"); } }
const a = { kind: 171, pos: 1, name: "a", typeText: "t" }; show(a);
const b = { kind: 80, pos: 2, escapedText: "e" }; show(b);`},
		{"unnamed kind", `function isNamed(n: Node): n is TypeElement { return kindOf(n) >= 171; }
function show(n: Node): void { if (isNamed(n)) { console.log(String(n.name)); } }
const v = { kind: 300, pos: 1, name: "jsx" }; show(v);`},
		{"refined field", `interface Computed extends Node { readonly kind: 167; readonly expressionText: string; }
interface Plain extends Node { readonly kind: 80; readonly escapedText: string; }
interface PropertyDeclaration extends Declaration { readonly kind: 172; readonly name: Computed | Plain; }
interface LateBound extends Declaration { readonly name: Computed; }
function isLate(d: Declaration): d is LateBound { const n = (d as PropertyDeclaration).name; return n.kind === 167; }
function show(d: Declaration): void { if (isLate(d)) { console.log(d.name.expressionText); } }
const p: PropertyDeclaration = { kind: 172, pos: 1, _declarationBrand: 0, name: { kind: 167, pos: 2, expressionText: "Symbol.iterator" } }; show(p);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, brandedPrelude+probe.source, ".ts")
		})
	}
	for _, probe := range []struct{ name, source, stderr string }{
		{"layout", `function show(n: Node): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "adamic: panic: field read failed: n.name matches no member of string | undefined"},
		{"stored", `const kept: TypeElement[] = [];
function show(n: Node): void { if (isTypeElementLie(n)) { kept.push(n); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);
for (const t of kept) { console.log(String(t.name)); }`, "adamic: panic: field read failed: t.name matches no member of string | undefined"},
		{"captured", `function show(n: Node): () => string { if (isTypeElementLie(n)) { return () => String(n.name); } return () => "no"; }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; console.log(show(v)());`, "adamic: panic: field read failed: n.name matches no member of string | undefined"},
		{"cast to a subtype", `function show(n: Node): void { if (isTypeElementLie(n) && n.kind === 171) { const p = n as PropertySignature; console.log(p.typeText); } }
const v = { kind: 171, pos: 2 }; show(v);`, "adamic: panic: field read failed: p.typeText is not initialized"},
		{"passed as a subtype", `function read(p: PropertySignature): void { console.log(p.typeText); }
function show(n: Node): void { if (isTypeElementLie(n) && n.kind === 171) { read(n as PropertySignature); } }
const v = { kind: 171, pos: 2 }; show(v);`, "adamic: panic: field read failed: p.typeText is not initialized"},
		{"narrowed to a subtype", `function assertTypeElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
function show(x: PropertySignature | Identifier): void { assertTypeElement(x); console.log(x.typeText); }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "adamic: panic: field read failed: x.typeText is not initialized"},
		{"write", `interface Counter extends Declaration { _counterBrand: any; count: number; }
function isCounter(n: Node): n is Counter { return kindOf(n) === 5; }
function bump(n: Node): void { if (isCounter(n)) { n.count = 7; } }
const v = { kind: 5, pos: 1, count: "s" }; bump(v); console.log(v.count);`, "adamic: panic: field read failed: <write>.count is not a number"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			brandedStops(t, brandedPrelude+probe.source, probe.stderr)
		})
	}
}

// What the views can't hold stays NotYet under its own reason: a false side, since an object test can't
// be negated (mutant: the guard dropped, and the honest Identifier panics on the false branch); a whole
// copy of a covered value, which no view checks until #yq6fkp5 (mutant: the copy scan dropped, and the
// spread lowers); a class or a generic subtype, which no view registered here covers. A supertype with a
// wide kind and no brand keeps the old reason, so the rule is keyed on shape (mutant: the brand test
// dropped, and Expression lowers).
func TestBrandedSupertypeStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"false branch", `function show(x: PropertySignature | Identifier): void { if (isTypeElement(x)) { console.log("element"); } else { console.log(x.escapedText); } }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "checked predicate membership of a branded supertype on its false branch"},
		{"spread", `function show(n: Node): void { if (isTypeElementLie(n)) { const c = { ...n }; console.log(String(c.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "a whole-object copy of a branded supertype's value, which no view checks"},
		{"rest", `function show(n: Node): void { if (isTypeElementLie(n)) { const { kind, ...rest } = n; console.log(String(rest.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "a whole-object copy of a branded supertype's value, which no view checks"},
		{"assign", `function show(n: Node): void { if (isTypeElementLie(n)) { const c = Object.assign({}, n); console.log(String(c.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "a whole-object copy of a branded supertype's value, which no view checks"},
		{"class", `class Signature implements TypeElement { _declarationBrand: any; _typeElementBrand: any; kind = 171; pos = 1; }
function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } }
show(new Signature());`, "a checked predicate target TypeElement that class Signature is assignable to"},
		{"generic", `interface Tagged<K extends number> extends TypeElement { readonly kind: K; readonly extra: number; }
function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } }
const a = { kind: 171, pos: 1, name: "a", typeText: "t" }; show(a);`, "a checked predicate target TypeElement that generic Tagged<K> is assignable to"},
		{"no brand", `interface Expression extends Node { readonly text: string; }
function isExpression(n: Node): n is Expression { return kindOf(n) === 9; }
function show(n: Node): void { if (isExpression(n)) { console.log(n.text); } }
const v = { kind: 9, pos: 1, text: "x" }; show(v);`, "checked predicate overload target Expression"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, brandedPrelude+probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// brandedStops lowers a .ts program and runs it on sanitized native and the JavaScript backend, each of
// which must stop at exit 70 with nothing printed and stderr starting with the located failure.
func brandedStops(t *testing.T, source, stderr string) {
	t.Helper()
	program, err := lowerTypeScriptAssertionSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
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
		if code != 70 || out.Len() != 0 || !strings.HasPrefix(errs.String(), stderr) {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want 70, nothing printed and %q", command.Path, code, out.String(), errs.String(), stderr)
		}
	}
}
