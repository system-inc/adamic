package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A branded supertype, tsc's TypeElement shape: a wide kind and a brand field typed undefined, over subtypes
// each with a literal kind. Identifier isn't one, since it lacks the brands.
const brandedPrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface TypeElement extends Declaration { _typeElementBrand: undefined; readonly name?: string; }
interface PropertySignature extends TypeElement { readonly kind: 171; readonly name: string; readonly typeText: string; }
interface MethodSignature extends TypeElement { readonly kind: 173; readonly name: string; readonly count: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function kindOf(n: Node): number { return n.kind; }
function isTypeElement(n: Node): n is TypeElement { const k = kindOf(n); return k === 171 || k === 173; }
function isTypeElementLie(n: Node): n is TypeElement { return kindOf(n) !== 0; }
`

// A checked predicate claiming a branded supertype checks at the call only that the value is an object,
// and every field where it's read, through S's view alone (#c7w6j0y change B, road 2, ruled Nov 1). An
// honest value agrees with Node whatever its kind, a value whose kind no subtype names included, as tsc's
// isLiteralExpression takes JsxText and isLateBound a declaration whose name it refines, and so does one
// a generic subtype's type could be given. A lie stops at its first bad read on sanitized native and the
// JavaScript backend when it's read through S, or stored or captured as S and read later. A write of the
// wrong type stops at the write. Mutants: no view registered, which reads every lie raw; one of S's
// fields dropped from the checked set, which reads the layout, stored and captured lies raw.
func TestBrandedSupertypeIsCheckedByItsViews(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"honest", `function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } else { console.log("no"); } }
const a = { kind: 171, pos: 1, name: "a", typeText: "t" }; show(a);
const b = { kind: 80, pos: 2, escapedText: "e" }; show(b);`},
		{"unnamed kind", `function isNamed(n: Node): n is TypeElement { return kindOf(n) >= 171; }
function show(n: Node): void { if (isNamed(n)) { console.log(String(n.name)); } }
const v = { kind: 300, pos: 1, name: "jsx" }; show(v);`},
		// A generic subtype is no different from any other subtype now that S's view is the only one: its
		// values are read through S's view once narrowed, and through #80cs7rz's rule when widened.
		{"generic subtype", `interface Tagged<K extends number> extends TypeElement { readonly kind: K; readonly extra: number; }
function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } }
const a = { kind: 171, pos: 1, name: "a", typeText: "t" }; show(a);`},
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
		{"write", `interface Counter extends Declaration { _counterBrand: undefined; count: number; }
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

// S's view is the only one, so a value it admitted and then gave another type with no cast is read
// through #80cs7rz's rule, NotYet at the read wherever the value may come from S's entry (road 2, ruled
// Nov 1). That holds Language's assignment roads on 7f4b7030, a field-less restatement (s4), an alias of
// S's shape (s5), an inline type of it (s6), a parameter (s7) and an array element (s8), and #80cs7rz's
// twelve roads entered through the predicate instead of a cast, each reading name through a wider type.
// Each lie prints 5 on Node. Mutants: a read through a subtype of a view's target passed by
// checkViewWidenedReads, which turns s4 to s8 red; checkViewWidenedReads off, which turns every one red.
// With ADAMIC_BRANDED_WIDENED_MUTANT=1 a probe must lower instead and its sanitized native must disagree
// with Node, which shows what the rule holds back.
func TestBrandedSupertypeWidenedReadIsNotYet(t *testing.T) {
	t.Parallel()
	const lie = `interface HasName { readonly name?: string; }
interface Restated extends TypeElement {}
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 };
`
	for _, probe := range []struct{ name, source, through, node string }{
		{"s4 assigned to a restatement", `function show(n: Node): void { if (isTypeElementLie(n)) { const r: Restated = n; console.log(String(r.name)); } }
show(v);`, "Restated", "5\n"},
		{"s5 assigned to an alias of its shape", `type Shape = { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: string };
function show(n: Node): void { if (isTypeElementLie(n)) { const r: Shape = n; console.log(String(r.name)); } }
show(v);`, "Shape", "5\n"},
		{"s6 assigned to an inline type of its shape", `function show(n: Node): void { if (isTypeElementLie(n)) { const r: { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: string } = n; console.log(String(r.name)); } }
show(v);`, "{ readonly kind: number;", "5\n"},
		{"s7 passed as a restatement", `function nameOf(r: Restated): string { return String(r.name); }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(nameOf(n)); } }
show(v);`, "Restated", "5\n"},
		{"s8 stored as a restatement", `const kept: Restated[] = [];
function show(n: Node): void { if (isTypeElementLie(n)) { kept.push(n); } }
show(v);
for (const r of kept) { console.log(String(r.name)); }`, "Restated", "5\n"},
		{"widened", `function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName = n; console.log(String(h.name)); } }
show(v);`, "HasName", "5\n"},
		{"array element", `function show(n: Node): void { if (isTypeElementLie(n)) { const xs: HasName[] = [n]; for (const x of xs) { console.log(String(x.name)); } } }
show(v);`, "HasName", "5\n"},
		{"map value", `function show(n: Node): void { if (isTypeElementLie(n)) { const m = new Map<string, HasName>([["k", n]]); const g = m.get("k"); console.log(g === undefined ? "none" : String(g.name)); } }
show(v);`, "HasName", "5\n"},
		{"closure", `function show(n: Node): () => string { if (isTypeElementLie(n)) { const h: HasName = n; return () => String(h.name); } return () => "no"; }
console.log(show(v)());`, "HasName", "5\n"},
		{"parameter", `function nameOf(x: HasName): string { return String(x.name); }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(nameOf(n)); } }
show(v);`, "HasName", "5\n"},
		{"return", `function view(n: Node): HasName | undefined { if (isTypeElementLie(n)) { return n; } return undefined; }
const h = view(v); console.log(h === undefined ? "none" : String(h.name));`, "HasName", "5\n"},
		{"field", `interface Box { readonly item: HasName; }
function show(n: Node): void { if (isTypeElementLie(n)) { const b: Box = { item: n }; console.log(String(b.item.name)); } }
show(v);`, "HasName", "5\n"},
		{"destructure", `function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName = n; const { name } = h; console.log(String(name)); } }
show(v);`, "HasName", "5\n"},
		{"template", `function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName = n; console.log(` + "`t=${String(h.name)}`" + `); } }
show(v);`, "HasName", "t=5\n"},
		{"conditional", `const other: HasName = { name: "o" };
function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName = kindOf(n) > 0 ? n : other; console.log(String(h.name)); } }
show(v);`, "HasName", "5\n"},
		{"class field", `class Printer { readonly h: HasName; constructor(h: HasName) { this.h = h; } run(): void { console.log(String(this.h.name)); } }
function show(n: Node): void { if (isTypeElementLie(n)) { new Printer(n).run(); } }
show(v);`, "HasName", "5\n"},
		{"satisfies", `function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName = n satisfies HasName; console.log(String(h.name)); } }
show(v);`, "HasName", "5\n"},
		{"optional chain", `function show(n: Node): void { if (isTypeElementLie(n)) { const h: HasName | undefined = n; console.log(String(h?.name)); } }
show(v);`, "HasName", "5\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, brandedPrelude+lie+probe.source)
			if os.Getenv("ADAMIC_BRANDED_WIDENED_MUTANT") == "1" {
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				code, out, errs := brandedNative(t, program)
				if code == 0 && out == probe.node {
					t.Fatalf("the mutant's sanitized native agrees with Node: %q", out)
				}
				t.Logf("the mutant admits it: sanitized native exit %d, stdout %q, stderr %.300q where Node prints %q", code, out, errs, probe.node)
				return
			}
			what := "a read of field name through type " + probe.through
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), what) || !strings.Contains(err.Error(), "which may hold a checked view's value") {
				t.Fatalf("want NotYet at the read with %q, got %v", what, err)
			}
		})
	}
}

// What S's view can't hold stays NotYet under its own reason: a false side, since an object test can't
// be negated (mutant: the guard dropped, and the honest Identifier panics on the false branch); a whole
// copy of S's value, which #yq6fkp5's rule holds at the copy now that it's on main, so B's own copy scan
// is gone; a class assignable to S, kept failing closed: an honest instance read through S's view agreed
// with Node with this rule off (probed on Golem, Nov 1), but no lying instance is pinned yet (mutant: the
// rule off, and the class lowers). A subtype of S in the argument's type keeps its own type when
// narrowed, with fields S's view doesn't check (mutant: a subtype member passed, and it lowers). A
// supertype with a wide kind and no brand keeps the old reason, so the rule is keyed on shape (mutant:
// the brand test dropped, and Expression lowers).
func TestBrandedSupertypeStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		// The false side narrows only when a member of the argument's type is assignable to S, and every such
		// member but one S is also assignable to (a restatement) is NotYet at the call by brandedNarrowing, so
		// a restatement is how the false side's guard is reached.
		{"false branch", `interface Restated extends TypeElement {}
function show(x: Restated | Node): void { if (isTypeElement(x)) { console.log("element"); } else { console.log(String(x.pos)); } }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "checked predicate membership of a branded supertype on its false branch"},
		{"spread", `function show(n: Node): void { if (isTypeElementLie(n)) { const c = { ...n }; console.log(String(c.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "a whole-object copy of a value of checked view type TypeElement"},
		{"rest", `function show(n: Node): void { if (isTypeElementLie(n)) { const { kind, ...rest } = n; console.log(String(rest.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "a destructured name that isn't plain"},
		{"assign", `function show(n: Node): void { if (isTypeElementLie(n)) { const c = Object.assign({}, n); console.log(String(c.name)); } }
const v = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(v);`, "refuses Object.assign"},
		{"class", `class Signature implements TypeElement { _declarationBrand: undefined; _typeElementBrand: undefined; kind = 171; pos = 1; }
function show(n: Node): void { if (isTypeElement(n)) { console.log(String(n.pos)); } }
show(new Signature());`, "a checked predicate target TypeElement that class Signature is assignable to"},
		// A cast beside S's view registers a view of its own, and the prelude's kindOf reads n.kind through
		// Node. Once these two views are in the program, #80cs7rz's rule finds that read may hold a viewed
		// value it can't vouch for, and makes it NotYet before any leg runs (Lower asked each of these be
		// named by the rule that holds it). These stopped at the read on 4a1fe092, before #80cs7rz;
		// #zpzm9ta is where such reads become checked.
		{"cast to a subtype", `function show(n: Node): void { if (isTypeElementLie(n) && n.kind === 171) { const p = n as PropertySignature; console.log(p.typeText); } }
const v = { kind: 171, pos: 2 }; show(v);`, "a read of field kind through type Node, which may hold a checked view's value"},
		{"passed as a subtype", `function read(p: PropertySignature): void { console.log(p.typeText); }
function show(n: Node): void { if (isTypeElementLie(n) && n.kind === 171) { read(n as PropertySignature); } }
const v = { kind: 171, pos: 2 }; show(v);`, "a read of field kind through type Node, which may hold a checked view's value"},
		// An honest program, held by the same rule at the same read: the price #80cs7rz charges, which the
		// subtype views paid too (177dcf9f stopped here as well).
		{"refined field", `interface Computed extends Node { readonly kind: 167; readonly expressionText: string; }
interface Plain extends Node { readonly kind: 80; readonly escapedText: string; }
interface PropertyDeclaration extends Declaration { readonly kind: 172; readonly name: Computed | Plain; }
interface LateBound extends Declaration { readonly name: Computed; }
function isLate(d: Declaration): d is LateBound { const n = (d as PropertyDeclaration).name; return n.kind === 167; }
function show(d: Declaration): void { if (isLate(d)) { console.log(d.name.expressionText); } }
const p: PropertyDeclaration = { kind: 172, pos: 1, _declarationBrand: undefined, name: { kind: 167, pos: 2, expressionText: "Symbol.iterator" } }; show(p);`, "a read of field kind through type Node, which may hold a checked view's value"},
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
const p: PropertyDeclaration = { kind: 172, pos: 1, _declarationBrand: undefined, name: { kind: 80, pos: 2, escapedText: "x" } }; show(p);`, "a checked predicate target LateBound that narrows PropertyDeclaration, which isn't its supertype"},
		// A member assignable to S beside an unrelated one: the checker keeps the first and drops the second,
		// so the call is NotYet, never a model of which member it drops (Compiler's must on 8c7afca2). It's
		// held by both members now, so passing a subtype member alone still leaves Identifier to stop it.
		{"unrelated member", `interface Loose extends Declaration { _typeElementBrand: undefined; readonly kind: 174; readonly name: any; }
function show(x: Loose | Identifier): void { if (isTypeElementLie(x)) { console.log(String(x.pos)); } }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "a checked predicate target TypeElement that narrows Identifier, which isn't its supertype"},
		// A subtype in the argument's type keeps its own type, so the lie's typeText or arity would be read
		// raw through it: S's view doesn't check them, and no entry of S's names the value.
		{"subtype member", `interface Signature extends TypeElement { readonly kind: 174; readonly arity: string; }
function assertTypeElement(n: Node): asserts n is TypeElement { if (n.kind !== 174) { throw new Error("no"); } }
function show(x: Signature | Node): void { assertTypeElement(x); console.log(String(x.pos)); }
const s: Signature = { kind: 174, pos: 3, arity: "1", _declarationBrand: undefined, _typeElementBrand: undefined }; show(s);`, "a checked predicate target TypeElement that narrows Signature, which isn't its supertype"},
		{"asserts over a union", `function assertTypeElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
function show(x: PropertySignature | Identifier): void { assertTypeElement(x); console.log(x.typeText); }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "a checked predicate target TypeElement that narrows Identifier, which isn't its supertype"},
		// A generic argument narrows to T & S, an intersection no view covers, so the call is NotYet
		// (Language's g1 to g3 on 7f4b7030). Mutant: the type-parameter guard dropped, and g1 lowers.
		{"generic argument", `function show<T extends Node>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`, "over a generic argument T, which narrows to an intersection no view covers"},
		{"generic argument through a supertype", `function show<T extends Declaration>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5, _declarationBrand: undefined }; show(w);`, "over a generic argument T, which narrows to an intersection no view covers"},
		{"generic argument given its type", `function show<T extends Node>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show<Node>(w);`, "over a generic argument T, which narrows to an intersection no view covers"},
		// A generic object type mentions its type parameter in an argument (Language's g4 and g5 on
		// 4a1fe092). Mutant: the guard judged by the type's own flags, and g4 lowers.
		{"generic object argument", `interface Holder<K extends number> { readonly kind: K; readonly pos: number; }
function show<K extends number>(n: Holder<K>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Holder<number> = w; show(v);`, "over a generic argument Holder<K>, which narrows to an intersection no view covers"},
		{"generic readonly argument", `function show<T extends Node>(n: Readonly<T>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`, "over a generic argument Readonly<T>, which narrows to an intersection no view covers"},
		// An anonymous object holding a type parameter is held by the related-member rule, not the generic
		// one, since mentionsTypeParameter doesn't walk members (Language's g7 on 59d9c44b, a should). It's
		// safe either way, so this pins the reason it has, and a later loosening of either rule shows here.
		{"generic anonymous argument", `function show<K extends number>(n: { readonly kind: K; readonly pos: number }): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(w);`, "which isn't its supertype"},
		// A member typed any is no brand (Adamic's correction on #tmdhz1d: #bw3xg7c refuses any), so a
		// supertype whose only brand-like member is any keeps the old reason. Mutant: any taken as a brand
		// again, and it lowers.
		{"any is no brand", `interface Opaque extends Node { _opaqueBrand: any; readonly text?: string; }
function isOpaque(n: Node): n is Opaque { return kindOf(n) === 9; }
function show(n: Node): void { if (isOpaque(n)) { console.log(String(n.text)); } }
const v = { kind: 9, pos: 1, text: "x" }; show(v);`, "checked predicate overload target Opaque"},
		{"no brand", `interface Expression extends Node { readonly text: string; }
function isExpression(n: Node): n is Expression { return kindOf(n) === 9; }
function show(n: Node): void { if (isExpression(n)) { console.log(n.text); } }
const v = { kind: 9, pos: 1, text: "x" }; show(v);`, "checked predicate overload target Expression"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, brandedPrelude+probe.source)
			var pending *NotYet
			var refused *Refused
			if !errors.As(err, &pending) && !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet or Refused with %q, got %v", probe.want, err)
			}
		})
	}
}

// The checker narrows a call's argument by the claim however the call reaches the predicate, so every
// path is judged by brandedNarrowing, from the one place they all reach (brandedCalls): a direct call, the
// predicate held as a value, a function-type parameter, a class method, an arrow, an assertion direct and
// held, and an optional call (Language's must on 9e4f2184, Compiler's roads on 7530a354). Each narrows a
// PropertySignature | Identifier argument to PropertySignature, whose typeText S's view doesn't check,
// so each is NotYet at the call. Three are held first by older rules, each pinned by its reason: the
// direct call by the false side's guard, which runs before S's view registers, the class method as an
// indirect call, and the optional call as a call main doesn't lower. Mutant: only a call through a declared function's name judged, as
// 9e4f2184 did, and the other roads lower; with ADAMIC_BRANDED_CALLS_MUTANT=1 each of those must lower
// instead and its sanitized native must disagree with Node, which prints undefined.
func TestBrandedSupertypeEveryCallPathIsJudged(t *testing.T) {
	t.Parallel()
	const lie = `const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
`
	const walk = "a checked predicate target TypeElement that narrows PropertySignature, which isn't its supertype"
	for _, probe := range []struct{ name, source, want string }{
		{"direct", `function show(x: PropertySignature | Identifier): void { if (isTypeElementLie(x)) { console.log(String(x.typeText)); } }
show(i);`, "checked predicate membership of a branded supertype on its false branch"},
		{"held as a value", `const guard = isTypeElementLie;
function show(x: PropertySignature | Identifier): void { if (guard(x)) { console.log(String(x.typeText)); } }
show(i);`, walk},
		{"function-type parameter", `function show(check: (n: Node) => n is TypeElement, x: PropertySignature | Identifier): void { if (check(x)) { console.log(String(x.typeText)); } }
show(isTypeElementLie, i);`},
		{"class method", `class Judge { isElement(n: Node): n is TypeElement { return kindOf(n) !== 0; } }
const judge = new Judge();
function show(x: PropertySignature | Identifier): void { if (judge.isElement(x)) { console.log(String(x.typeText)); } }
show(i);`, "an indirect call of a checked predicate"},
		{"arrow", `const isElement = (n: Node): n is TypeElement => kindOf(n) !== 0;
function show(x: PropertySignature | Identifier): void { if (isElement(x)) { console.log(String(x.typeText)); } }
show(i);`, walk},
		// An assertion narrows after the call statement, with no if around it (Compiler, on 7530a354).
		{"asserts", `function assertElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
function show(x: PropertySignature | Identifier): void { assertElement(x); console.log(String(x.typeText)); }
show(i);`, walk},
		{"asserts held as a value", `function assertElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
const held: (n: Node) => asserts n is TypeElement = assertElement;
function show(x: PropertySignature | Identifier): void { held(x); console.log(String(x.typeText)); }
show(i);`, walk},
		{"optional call", `const maybe: ((n: Node) => n is TypeElement) | undefined = isTypeElementLie;
function show(x: PropertySignature | Identifier): void { if (maybe?.(x)) { console.log(String(x.typeText)); } }
show(i);`, "a call through ?. (an optional call)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, brandedPrelude+lie+probe.source)
			if os.Getenv("ADAMIC_BRANDED_CALLS_MUTANT") == "1" && probe.want == walk {
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				code, out, errs := brandedNative(t, program)
				if code == 0 && out == "undefined\n" {
					t.Fatalf("the mutant's sanitized native agrees with Node: %q", out)
				}
				t.Logf("the mutant admits it: sanitized native exit %d, stdout %q, stderr %.300q where Node prints undefined", code, out, errs)
				return
			}
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet with %q, got %v", probe.want, err)
			}
		})
	}
}

// A member S is assignable to and that's assignable to S keeps its own type when the checker narrows it
// (getNarrowedType takes a subtype of the candidate as it is), so brandedNarrowing passes it and S's view
// never sees its reads (Compiler's pin on 9e4f2184): a restatement, an alias of S's shape, and a twin whose
// name is any. Each is held to never read raw: a program lowers and agrees with Node, or stops at exit 70
// with a located failure on sanitized native and the JavaScript backend, or is NotYet. The restatement and
// the alias can only hold a lie a cast of its own put there, which that cast's view checks; the twin holds
// one by plain assignment.
func TestBrandedSupertypeMutualMemberNeverReadsRaw(t *testing.T) {
	t.Parallel()
	const mutual = `interface Restated extends TypeElement {}
type Shape = { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: string };
interface Twin { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: any }
function viaRestated(x: Restated): string { if (isTypeElementLie(x)) { return String(x.name); } return "no"; }
function viaShape(x: Shape): string { if (isTypeElementLie(x)) { return String(x.name); } return "no"; }
function viaTwin(x: Twin): string { if (isTypeElementLie(x)) { return String(x.name); } return "no"; }
`
	for _, probe := range []struct{ name, source string }{
		{"restatement honest", `const r: Restated = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(viaRestated(r));`},
		{"restatement through a cast", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(viaRestated(n as Restated));`},
		{"alias honest", `const r: Shape = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(viaShape(r));`},
		{"alias through a cast", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(viaShape(n as Shape));`},
		{"twin", `const t: Twin = { kind: 80, pos: 2, _declarationBrand: undefined, _typeElementBrand: undefined, name: 5 }; console.log(viaTwin(t));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := brandedPrelude + mutual + probe.source
			program, err := lowerTypeScriptAssertionSource(t, source)
			var pending *NotYet
			var refused *Refused
			if errors.As(err, &pending) || errors.As(err, &refused) {
				t.Logf("stops before it lowers: %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(probe.name, "honest") {
				brandedStopsProgram(t, program, "adamic: panic: ")
				return
			}
			lowersAndAgreesWithNodeAs(t, source, ".ts")
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
	brandedStopsProgram(t, program, stderr)
}

// brandedStopsProgram runs a lowered program on sanitized native and the JavaScript backend, each of which
// must stop at exit 70 with nothing printed and stderr starting with stderr.
func brandedStopsProgram(t *testing.T, program *ir.Program, stderr string) {
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

// brandedNative builds a program's sanitized native and runs it, returning its exit code, stdout and stderr.
func brandedNative(t *testing.T, program *ir.Program) (int, string, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
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
	return code, out.String(), errs.String()
}
