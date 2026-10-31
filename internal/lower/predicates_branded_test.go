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
// through S, stored or captured and read later, or cast or passed as a subtype. A write of the wrong type
// stops at the write. Mutants: no view registered, which reads every lie raw but the cast and the subtype
// parameter, which the cast's own view holds too; one of S's fields dropped from the checked set, which
// reads the layout, stored and captured lies raw.
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
	// tsc's brands are typed undefined (#ptd2v1r), so an undefined brand takes the same road (Correctness's
	// census on 8c7afca2 found the rule never firing on tsc). Mutant: undefined dropped from phantomField,
	// which reddens these with phantomField's own pins.
	undefinedBrands := strings.ReplaceAll(brandedPrelude, "Brand: any;", "Brand: undefined;")
	t.Run("undefined brand", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, undefinedBrands+`function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } else { console.log("no"); } }
const a = { kind: 171, pos: 1, name: "a", typeText: "t" }; show(a);
const b = { kind: 80, pos: 2, escapedText: "e" }; show(b);`, ".ts")
	})
	t.Run("undefined brand lie", func(t *testing.T) {
		t.Parallel()
		brandedStops(t, undefinedBrands+`function show(n: Node): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "adamic: panic: field read failed: n.name matches no member of string | undefined")
	})
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
		// A declared type S is assignable to takes the value by plain assignment, so only its own view
		// stops the lie (Language's s4, s5, s7, s8 on 7f4b7030). Mutant: the subtypes' views not
		// registered, and all four read the lie raw.
		{"assigned to a restatement", `interface Restated extends TypeElement {}
function show(n: Node): void { if (isTypeElementLie(n)) { const r: Restated = n; console.log(String(r.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "adamic: panic: field read failed: r.name matches no member of string | undefined"},
		{"assigned to an alias of its shape", `type Shape = { readonly kind: number; readonly pos: number; _declarationBrand: any; _typeElementBrand: any; readonly name?: string };
function show(n: Node): void { if (isTypeElementLie(n)) { const r: Shape = n; console.log(String(r.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "adamic: panic: field read failed: r.name matches no member of string | undefined"},
		{"passed as a restatement", `interface Restated extends TypeElement {}
function nameOf(r: Restated): string { return String(r.name); }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(nameOf(n)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "adamic: panic: field read failed: r.name matches no member of string | undefined"},
		{"stored as a restatement", `interface Restated extends TypeElement {}
const kept: Restated[] = [];
function show(n: Node): void { if (isTypeElementLie(n)) { kept.push(n); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);
for (const r of kept) { console.log(String(r.name)); }`, "adamic: panic: field read failed: r.name matches no member of string | undefined"},
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
		{"false branch", `function show(x: PropertySignature | Node): void { if (isTypeElement(x)) { console.log("element"); } else { console.log(String(x.pos)); } }
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
		// A recursive alias ends the false side's walk on its second visit (Lower's must on 9abcbdca): Nodes'
		// array holds Nodes, and no member is branded, so the walk would recurse forever without it.
		{"recursive alias", `type Nodes = Identifier | Nodes[];
function isNodes(x: unknown): x is Nodes { return x !== 0; }
function show(x: unknown): void { if (isNodes(x)) { console.log("nodes"); } else { console.log(String(x)); } }
show(1);`, ""},
		// A concrete source neither way related to the claim narrows to Source & S, which no view covers
		// (Compiler's must on 9abcbdca): name would be read as Computed while the value holds a Plain.
		// Mutant: the narrowing check dropped, and the call lowers.
		{"intersection", `interface Computed extends Node { readonly kind: 167; readonly expressionText: string; }
interface Plain extends Node { readonly kind: 80; readonly escapedText: string; }
interface PropertyDeclaration extends Declaration { readonly kind: 172; readonly name: Computed | Plain; }
interface LateBound extends Declaration { readonly name: Computed; }
function isLateLie(d: Declaration): d is LateBound { return kindOf(d) !== 0; }
function show(p: PropertyDeclaration): void { if (isLateLie(p)) { console.log(p.name.expressionText); } }
const p: PropertyDeclaration = { kind: 172, pos: 1, _declarationBrand: 0, name: { kind: 80, pos: 2, escapedText: "x" } }; show(p);`, "a checked predicate target LateBound that narrows PropertyDeclaration, which is neither its supertype nor a subtype its views cover"},
		// A member assignable to S but not its subtype, beside an unrelated one: the checker narrows both to
		// intersections, so the call is NotYet, never a model of which member it drops (Compiler's must on
		// 8c7afca2). Mutant: an unrelated member dropped beside a related one, and the call lowers.
		{"unrelated member", `interface Loose extends Declaration { _typeElementBrand: any; readonly kind: 174; readonly name: any; }
function show(x: Loose | Identifier): void { if (isTypeElementLie(x)) { console.log(String(x.pos)); } }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "a checked predicate target TypeElement that narrows Identifier, which is neither its supertype nor a subtype its views cover"},
		{"asserts over a union", `function assertTypeElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
function show(x: PropertySignature | Identifier): void { assertTypeElement(x); console.log(x.typeText); }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "a checked predicate target TypeElement that narrows Identifier, which is neither its supertype nor a subtype its views cover"},
		// A generic argument narrows to T & S, an intersection no view covers, so the call is NotYet
		// (Language's g1 to g3 on 7f4b7030). Mutant: the type-parameter guard dropped, and g1 lowers.
		{"generic argument", `function show<T extends Node>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`, "over a generic argument T, which narrows to an intersection no view covers"},
		{"generic argument through a supertype", `function show<T extends Declaration>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5, _declarationBrand: 0 }; show(w);`, "over a generic argument T, which narrows to an intersection no view covers"},
		{"generic argument given its type", `function show<T extends Node>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show<Node>(w);`, "over a generic argument T, which narrows to an intersection no view covers"},
		// A generic object type mentions its type parameter in an argument (Language's g4 and g5 on
		// 4a1fe092). Mutant: the guard judged by the type's own flags, and g4 lowers.
		{"generic object argument", `interface Holder<K extends number> { readonly kind: K; readonly pos: number; }
function show<K extends number>(n: Holder<K>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Holder<number> = w; show(v);`, "over a generic argument Holder<K>, which narrows to an intersection no view covers"},
		{"generic readonly argument", `function show<T extends Node>(n: Readonly<T>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`, "over a generic argument Readonly<T>, which narrows to an intersection no view covers"},
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
				var refused *Refused
				if probe.want == "" && errors.As(err, &refused) {
					return
				}
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// A subtype declared in a file the program imports only for its types is covered too, since programFiles
// follows type-only imports, so a call whose argument may be one lowers and agrees with Node. Mutant: the
// program's root files alone, and the subtype goes uncovered, so the call is NotYet.
func TestBrandedSupertypeCoversATypeOnlyImport(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "types.ts"), []byte(brandedPrelude[:strings.Index(brandedPrelude, "function kindOf")]+`export interface Signature extends TypeElement { readonly kind: 174; readonly arity: string; }
export type { Node, TypeElement };
`), 0o644); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(main, []byte(`import type { Node, TypeElement, Signature } from "./types";
function assertTypeElement(n: Node): asserts n is TypeElement { if (n.kind !== 174) { throw new Error("no"); } }
function show(x: Signature | Node): void { assertTypeElement(x); console.log(String(x.pos)); }
const s: Signature = { kind: 174, pos: 3, arity: "1", _declarationBrand: 0, _typeElementBrand: 0 }; show(s);
`), 0o644); err != nil {
		t.Fatal(err)
	}
	agreeEntry(t, main, nil, 0, nil, nil)
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
