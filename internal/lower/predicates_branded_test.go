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
		// A generic argument narrows to T & S, and its first read on the true side concretes to the instance's
		// intersection. Here T is the object literal's own type, which S isn't assignable to, so the call is
		// NotYet (one of Language's g1 to g3 on 7f4b7030). The other two, at Node inferred and given, and the
		// generic object argument, g4 on 4a1fe092, read S intersected only with supertypes of S, so they're
		// checked now (TestBrandedSupertypeGenericIntersection). Mutant: the type-parameter guard dropped, and
		// this is NotYet by the wrong rule.
		{"generic argument through a supertype", `function show<T extends Declaration>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5, _declarationBrand: undefined }; show(w);`, "over a generic argument T"},
		// g5 on 4a1fe092: its read concretes to Readonly<T> & TypeElement with T still unmapped (measured on
		// Golem, Nov 1), so a member mentions a type parameter and the call stays NotYet (Adamic's ruling,
		// Nov 1: B's unread rule).
		{"generic readonly argument", `function show<T extends Node>(n: Readonly<T>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`, "read on the true side as Readonly<T> & TypeElement"},
		// An anonymous object holding a type parameter is held by the related-member rule, not the generic
		// one, since mentionsTypeParameter doesn't walk members (Language's g7 on 59d9c44b, a should), and K,
		// read from no reference or signature, is unread, so no mapper makes a new type of it. It's safe
		// either way, so this pins the reason it has, and a later loosening of either rule shows here.
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

// The checker narrows a call's argument by the claim however the call reaches the predicate, and every
// call is judged when its body is lowered, under its instance (judgeBrandedCalls, a walk over the body):
// a direct call, the predicate held as a value, a function-type parameter, a class method, an
// arrow, an assertion direct and held, an optional call, and claims through a generic predicate. Each
// narrows a PropertySignature | Identifier argument to PropertySignature, whose typeText S's view doesn't
// check, so each is NotYet at the call, by Identifier, an unrelated member the checker drops, as much as
// by the subtype. Mutant: the walk dropped, and the roads marked opened lower; with
// ADAMIC_BRANDED_CALLS_MUTANT=1 each of those must lower instead and its sanitized native must disagree
// with Node, which prints undefined (on Golem it stops with "compiler bug: a field the checker proved is
// there is missing"). The others are held then by older rules: the direct call by the false side's guard
// and the optional call as a call main doesn't lower. The class method is held by the indirect-call rule
// whichever way, since that rule runs before any body is lowered, and the two generic claims by the unread
// rule, since U stands nowhere in g's signature.
func TestBrandedSupertypeEveryCallPathIsJudged(t *testing.T) {
	t.Parallel()
	const lie = `const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
`
	// brandedNarrowing names the first member that isn't a supertype, PropertySignature or Identifier by the
	// union's order, so the pin holds the reason's own words.
	const walk = "which isn't its supertype"
	const unread = "which this body is lowered without knowing"
	for _, probe := range []struct {
		name, source, want string
		opened             bool
	}{
		{"direct", `function show(x: PropertySignature | Identifier): void { if (isTypeElementLie(x)) { console.log(String(x.typeText)); } }
show(i);`, walk, false},
		{"held as a value", `const guard = isTypeElementLie;
function show(x: PropertySignature | Identifier): void { if (guard(x)) { console.log(String(x.typeText)); } }
show(i);`, walk, true},
		{"function-type parameter", `function show(check: (n: Node) => n is TypeElement, x: PropertySignature | Identifier): void { if (check(x)) { console.log(String(x.typeText)); } }
show(isTypeElementLie, i);`, walk, true},
		{"class method", `class Judge { isElement(n: Node): n is TypeElement { return kindOf(n) !== 0; } }
const judge = new Judge();
function show(x: PropertySignature | Identifier): void { if (judge.isElement(x)) { console.log(String(x.typeText)); } }
show(i);`, "an indirect call of a checked predicate", false},
		{"arrow", `const isElement = (n: Node): n is TypeElement => kindOf(n) !== 0;
function show(x: PropertySignature | Identifier): void { if (isElement(x)) { console.log(String(x.typeText)); } }
show(i);`, walk, true},
		// An assertion narrows after the call statement, with no if around it (Compiler, on 7530a354).
		{"asserts", `function assertElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
function show(x: PropertySignature | Identifier): void { assertElement(x); console.log(String(x.typeText)); }
show(i);`, walk, true},
		{"asserts held as a value", `function assertElement(n: Node): asserts n is TypeElement { if (kindOf(n) === 0) { throw new Error("no"); } }
const held: (n: Node) => asserts n is TypeElement = assertElement;
function show(x: PropertySignature | Identifier): void { held(x); console.log(String(x.typeText)); }
show(i);`, walk, true},
		// A generic predicate's claim over g's U, which stands nowhere in g's signature, so g's body is
		// lowered with U unread and shared by every instance: NotYet, whatever U is (Adamic's ruling, Nov 1).
		{"generic claim", `function isT<T extends Node>(n: Node): n is T { return kindOf(n) !== 0; }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log("element"); } }
function g<U extends Node>(x: PropertySignature | Identifier): void { if (isT<U>(x)) { console.log(String(x.pos)); } }
g<TypeElement>(i);
show(i);`, unread, false},
		// A generic branded supertype: Tagged<U> at U = number is S itself, and U is unread in g, so the claim
		// is NotYet whatever U's constraint (Compiler's must on 810ce7d9).
		{"generic branded supertype", `interface Tagged<V> { readonly kind: number; readonly pos: number; _taggedBrand: undefined; readonly value: V }
interface Sub extends Tagged<number> { readonly kind: 7; readonly extra: string; }
function isTagged(n: Node): n is Tagged<number> { return kindOf(n) !== 0; }
function isTaggedOf<U extends number>(n: Node): n is Tagged<U> { return kindOf(n) !== 0; }
function show(n: Node): void { if (isTagged(n)) { console.log(String(n.pos)); } }
function g<U extends number>(x: Sub | Identifier): void { if (isTaggedOf<U>(x)) { console.log(String(x.pos)); } }
show(i);
g<number>(i);`, unread, false},
		{"optional call", `const maybe: ((n: Node) => n is TypeElement) | undefined = isTypeElementLie;
function show(x: PropertySignature | Identifier): void { if (maybe?.(x)) { console.log(String(x.typeText)); } }
show(i);`, walk, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, brandedPrelude+lie+probe.source)
			if os.Getenv("ADAMIC_BRANDED_CALLS_MUTANT") == "1" && probe.opened {
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

// tsc's binarySearchKey calls `some(array)`, which claims `readonly T[]` of an argument already typed
// `readonly T[]`, so the narrowing is the argument's own type whatever T is (computeLineOfPosition and
// indexOfNode, Correctness's census on 9e4f2184). With S's view registered in the program, the walk reads
// the call as written, a claim naming firstPos's T, and passes it because the argument is the claim's own
// type, so T at Node and at number lowers and agrees with Node, as on main (Adamic's ruling). An empty
// array isn't run: `some` answering false of one is a lie its checked false side stops, which B doesn't
// touch for an unbranded T. Mutant: that identity test dropped, and the program is NotYet as a generic
// claim S may instantiate. At T = TypeElement the instance's claim is an array of S, whose false side the
// false side's guard keeps NotYet.
func TestBrandedSupertypeIdentityNarrowingLowers(t *testing.T) {
	t.Parallel()
	const some = `function some<T>(array: readonly T[] | undefined): array is readonly T[] { return array !== undefined && array.length > 0; }
function firstPos<T extends Node>(array: readonly T[]): number { if (!some(array)) { return -1; } const first = array[0]; return first === undefined ? -1 : first.pos; }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(String(n.pos)); } }
const element: TypeElement = { kind: 171, pos: 4, _declarationBrand: undefined, _typeElementBrand: undefined };
show(element);
`
	lowersAndAgreesWithNodeAs(t, brandedPrelude+some+`function firstOf(array: readonly number[]): number { if (!some(array)) { return -1; } const first = array[0]; return first === undefined ? -1 : first; }
const nodes: Node[] = [{ kind: 80, pos: 6 }];
console.log(String(firstPos(nodes)));
console.log(String(firstOf([7])));
`, ".ts")
	_, err := lowerTypeScriptAssertionSource(t, brandedPrelude+some+`const elements: TypeElement[] = [element];
console.log(String(firstPos(elements)));
`)
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "checked predicate membership of a branded supertype on its false branch") {
		t.Fatalf("want T at TypeElement NotYet on the false side, got %v", err)
	}
}

// tsc's tryCast and cast take a predicate as a function type and call it on a value of the type its claim
// extends: `test(value)` claims TOut of a TIn, with TOut extends TIn. Each instance is judged where it's
// lowered (core.ts:1779, Adamic's ruling on 810ce7d9). Both read the value, after `value !== undefined`,
// as TIn & TOut, which concretes to Node & TypeElement: S intersected only with a supertype of S, read
// through S's checked view (brandedIntersection, Adamic's ruling, Nov 1). So both lower and agree with Node
// for isTypeElement and isClassElement alike, and so does TOut = TypeElement over a PropertySignature |
// Node (Compiler's union pin), whose result reads through TypeElement's view. A lie stops at its first bad
// read: inside the helper's body when the body reads a field, as castName does, and through tryCast, whose
// body reads nothing, at the caller's. Mutants: the read compared as written, unconcreted, and these are
// NotYet; the intersection's reads not routed to S's view (brandedIntersectionReceiver), and castName's
// value.name, read through Positioned & TypeElement, is NotYet as a widened read, honest or lying.
func TestBrandedSupertypeTryCastIsJudgedPerInstance(t *testing.T) {
	t.Parallel()
	// castName is cast reading a field of its value after the narrowing. Its TIn holds undefined, so
	// `value !== undefined` reads it as TIn & {}, which the claim narrows to TIn & {} & TOut, as cast's.
	const helpers = `interface ClassElement extends Declaration { _classElementBrand: undefined; readonly name?: string; }
interface Positioned { readonly pos: number; readonly name?: string; }
function isClassElement(n: Node): n is ClassElement { const k = kindOf(n); return k === 172 || k === 174; }
function tryCast<TOut extends TIn, TIn = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut | undefined { return value !== undefined && test(value) ? value : undefined; }
function cast<TOut extends TIn, TIn = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut { if (value !== undefined && test(value)) { return value; } throw new Error("Invalid cast"); }
function castName<TOut extends TIn, TIn extends Positioned | undefined = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut { if (value !== undefined && test(value)) { console.log(String(value.name)); return value; } throw new Error("Invalid cast"); }
`
	for _, probe := range []struct{ name, source string }{
		{"tryCast", `function show(n: Node): void { const e = tryCast(n, isTypeElement); console.log(e === undefined ? "no" : String(e.pos)); }
show({ kind: 171, pos: 1 }); show({ kind: 80, pos: 2 });`},
		{"tryCast class element", `function show(n: Node): void { const c = tryCast(n, isClassElement); console.log(c === undefined ? "no" : String(c.pos)); }
show({ kind: 172, pos: 2 }); show({ kind: 80, pos: 3 });`},
		{"cast", `const a = { kind: 171, pos: 1, name: "a", typeText: "t" };
const e = cast(a as Node, isTypeElement); console.log(String(e.pos) + " " + String(e.name));`},
		{"cast class element", `const a = { kind: 172, pos: 4, name: "c" };
const c = cast(a as Node, isClassElement); console.log(String(c.pos) + " " + String(c.name));`},
		{"castName", `const a = { kind: 171, pos: 1, name: "a", typeText: "t" };
console.log(String(castName(a as Node, isTypeElement).pos));`},
		{"union", `const a = { kind: 171, pos: 1, name: "a", typeText: "t" };
const p: PropertySignature | Node = a;
const u = tryCast(p, isTypeElement); console.log(u === undefined ? "no" : String(u.name));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, brandedPrelude+helpers+probe.source, ".ts")
		})
	}
	const lie = `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w;
`
	for _, probe := range []struct{ name, source, stderr string }{
		{"castName lie", `console.log(String(castName(n, isTypeElementLie).pos));`, "adamic: panic: field read failed: value.name matches no member of string | undefined"},
		{"tryCast lie", `const e = tryCast(n, isTypeElementLie); console.log(e === undefined ? "no" : String(e.name));`, "adamic: panic: field read failed: e.name matches no member of string | undefined"},
		{"union lie", `const p: PropertySignature | Node = n; const u = tryCast(p, isTypeElementLie); console.log(u === undefined ? "no" : String(u.name));`, "adamic: panic: field read failed: u.name matches no member of string | undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			brandedStops(t, brandedPrelude+helpers+lie+probe.source, probe.stderr)
		})
	}
}

// A generic argument narrows to T & S, which concretes to the instance's own intersection. At T = Node,
// Declaration, or another supertype of S whose every field is S's, it's S intersected only with supertypes
// of S, read through S's view, so an honest value agrees with Node and a lie stops at the read (Language's
// g1 at Node, and its g4 at Holder<number>, Adamic's ruling, Nov 1). At T = PropertySignature, a subtype
// of S, at T = Identifier, which S isn't assignable to, at T = Named, whose name is required where S's is
// optional, and at T = Bag, whose index signature an interface S has no implicit one for, the
// intersection holds a member that isn't a supertype, so the call stays NotYet. So it does at T = Lettered,
// whose optional a S lacks, at T = Present, whose name S has as an optional field of the same type, and,
// for a claim of Shape, a type literal's alias S that does have an implicit index signature, at T = Bag,
// whose index signature is no property (brandedSupertypeMember, ruled Nov 2). Mutants: any member let
// through, and g1 at PropertySignature, Identifier and Named lower; comparability in place of
// assignability, and g1 at Present lowers; the reads not routed to S's view, and the honest g1 is NotYet
// as a widened read; the property-subset test dropped, and g1 at Lettered lowers; the signature test
// dropped, and Shape at Bag lowers.
func TestBrandedSupertypeGenericIntersection(t *testing.T) {
	t.Parallel()
	const g1 = `function show<T extends Node>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
`
	const honest = `function show<T extends Node>(n: T): void { if (isTypeElement(n)) { console.log(String(n.name)); } else { console.log("no"); } }
`
	lowersAndAgreesWithNodeAs(t, brandedPrelude+honest+`const w = { kind: 171, pos: 2, name: "a" }; const v: Node = w; show(v);
const x = { kind: 80, pos: 3 }; const y: Node = x; show(y);
`, ".ts")
	const lie = `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 };
`
	for _, probe := range []struct{ name, source string }{
		{"g1 at Node", g1 + `const v: Node = w; show(v);`},
		{"g1 given Node", g1 + `show<Node>(w);`},
		{"g4 at Holder<number>", `interface Holder<K extends number> { readonly kind: K; readonly pos: number; }
function show<K extends number>(n: Holder<K>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const v: Holder<number> = w; show(v);`},
		{"g1 at Declaration", `function show<T extends Declaration>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const d = { kind: 80, pos: 2, _declarationBrand: undefined, name: 5 }; const v: Declaration = d; show(v);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			brandedStops(t, brandedPrelude+lie+probe.source, "adamic: panic: field read failed: n.name matches no member of string | undefined")
		})
	}
	// Lettered's a is optional, so TypeElement, which lacks it, is assignable to Lettered. On 021e43f3 the
	// honest g1 at Lettered lowered and read x.a raw through TypeElement's id, unchecked (Pupa, Nov 2). A
	// lying a can't be given a Lettered today: method bivariance, JSON.parse, a cast through unknown and
	// any are each refused before this rule is reached, so the honest program is the pin.
	const lettered = `interface Lettered extends Node { readonly a?: string; }
function show<T extends Lettered>(x: T): void { if (isTypeElementLie(x)) { console.log(String(x.a)); } }
`
	for _, probe := range []struct{ name, source, want string }{
		{"g1 at PropertySignature", g1 + `const p: PropertySignature = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a", typeText: "t" }; show(p);`, "read on the true side as PropertySignature & TypeElement"},
		{"g1 at Identifier", g1 + `const i: Identifier = { kind: 80, pos: 2, escapedText: "e" }; show(i);`, "read on the true side as Identifier & TypeElement"},
		{"g1 at Named", `interface Named extends Node { readonly name: string; }
` + g1 + `const m: Named = { kind: 80, pos: 2, name: "m" }; show(m);`, "read on the true side as Named & TypeElement"},
		{"g1 at Lettered", lettered + `const w = { kind: 80, pos: 2, a: "x" }; const v: Lettered = w; show(v);`, "read on the true side as Lettered & TypeElement"},
		{"g1 at Bag", `interface Bag extends Node { [k: string]: unknown; }
function show<T extends Bag>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const b: Bag = { kind: 80, pos: 2 }; show(b);`, "read on the true side as Bag & TypeElement"},
		{"g1 at Present", `interface Present extends Node { readonly name: string | undefined; }
` + g1 + `const q: Present = { kind: 80, pos: 2, name: "q" }; show(q);`, "read on the true side as Present & TypeElement"},
		{"Shape at Bag", `type Shape = { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _shapeBrand: undefined; readonly name?: string };
function isShapeLie(n: Node): n is Shape { return kindOf(n) !== 0; }
interface Bag extends Node { [k: string]: unknown; }
function show<T extends Bag>(n: T): void { if (isShapeLie(n)) { console.log(String(n.name)); } }
const b: Bag = { kind: 80, pos: 2 }; show(b);`, "read on the true side as Bag & Shape"},
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

// brandedIntersectionReceiver routes a read to the one branded member every other member is a supertype
// of, so with two branded members it's the narrower or neither (Compiler, on 021e43f3). TypeElement &
// ClassElement holds two that neither is a supertype of the other, so the receiver stays the intersection
// and x.name, with both views live, is no checked read: a plain one, since x holds only an allocation no
// view admitted, and the program agrees with Node. MethodLike extends TypeElement, both branded, so g1's
// read through TypeElement & MethodLike goes through MethodLike's view, which checks arity, and a lie stops
// there. Mutant: the receiver takes the first branded member found, and TypeElement & ClassElement's read
// is checked, and MethodLike's arity is read through TypeElement, NotYet as a widened read.
func TestBrandedSupertypeIntersectionReceiver(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, brandedPrelude+`interface ClassElement extends Declaration { _classElementBrand: undefined; readonly name?: string; }
function isClassElementLie(n: Node): n is ClassElement { return kindOf(n) !== 0; }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(String(n.pos)); } if (isClassElementLie(n)) { console.log(String(n.pos)); } }
function nameOf(x: TypeElement & ClassElement): string { return String(x.name); }
const n: Node = { kind: 2, pos: 3 }; show(n);
const b: TypeElement & ClassElement = { kind: 1, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, _classElementBrand: undefined, name: "b" };
console.log(nameOf(b));
`, ".ts")
	reads := 0
	inspect := func(node any) bool {
		if read, ok := node.(ir.Property); ok && read.Name == "name" {
			reads++
			if program.CheckedField(read.ViewReceiverTypeID, read.Name) {
				t.Errorf("x.name through TypeElement & ClassElement is a checked read of receiver %d", read.ViewReceiverTypeID)
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if reads != 1 {
		t.Fatalf("want the one read of name, x.name, found %d", reads)
	}
	brandedStops(t, brandedPrelude+`interface MethodLike extends TypeElement { _methodLikeBrand: undefined; readonly arity?: number; }
function isMethodLikeLie(n: Node): n is MethodLike { return kindOf(n) !== 0; }
function show<T extends TypeElement>(n: T): void { if (isMethodLikeLie(n)) { console.log(String(n.arity)); } }
const w = { kind: 80, pos: 2, _declarationBrand: undefined, _typeElementBrand: undefined, name: "m", arity: "x" }; const e: TypeElement = w; show(e);
`, "adamic: panic: field read failed: n.arity matches no member of number | undefined")
}

// Two branded members each a supertype of the other, TypeElement and Twin2 with the same fields, both
// qualify, so brandedIntersectionReceiver routes a declared TypeElement & Twin2 to its first member and
// Twin2 & TypeElement to its own first, a different id. The outcome is the same either way with both views
// live, and that's what's pinned, never the id: x.name is a checked read, an honest value agrees with Node,
// and a lie stops with the same message on sanitized native and the JavaScript backend (probed on Pupa,
// Nov 2; Compiler's accept). Mutant: the routing removed, and each lie is NotYet as a widened read.
func TestBrandedSupertypeMutualIntersectionOrder(t *testing.T) {
	t.Parallel()
	const twin = `interface Twin2 extends Declaration { _typeElementBrand: undefined; readonly name?: string; }
function isTwin2Lie(n: Node): n is Twin2 { return kindOf(n) !== 0; }
function reg(n: Node): number { if (isTwin2Lie(n)) { return n.pos; } return 0; }
const r: Node = { kind: 3, pos: 3 }; reg(r);
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(nameOf(n)); } }
`
	// reg puts Twin2's view in the program and prints nothing, since a stop must come before any output.
	for _, order := range []string{"TypeElement & Twin2", "Twin2 & TypeElement"} {
		t.Run(order, func(t *testing.T) {
			t.Parallel()
			source := brandedPrelude + twin + "function nameOf(x: " + order + "): string { return String(x.name); }\n"
			program := lowersAndAgreesWithNodeAs(t, source+`const w = { kind: 171, pos: 2, name: "a" }; const v: Node = w; show(v);
`, ".ts")
			reads := 0
			inspect := func(node any) bool {
				if read, ok := node.(ir.Property); ok && read.Name == "name" {
					reads++
					if !program.CheckedField(read.ViewReceiverTypeID, read.Name) {
						t.Errorf("x.name through %s isn't a checked read", order)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if reads != 1 {
				t.Fatalf("want the one read of name, x.name, found %d", reads)
			}
			brandedStops(t, source+`const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);
`, "adamic: panic: field read failed: x.name matches no member of string | undefined")
		})
	}
}

// Only a read is routed to S's view. A write through S intersected with its supertypes keeps the
// intersection's own id, since S's checked write checks a slot a value admitted as S needn't have: routed,
// Language's y4 stopped at x.name = "w" with "<write>.name is not initialized" on both backends where Node
// prints w (Optimizer's ruling, Nov 2). Mut's name is optional and Branded's value has none, or has "a".
// Unrouted, the write is a plain one, and with #krv1hgh on main a plain write to an optional field some
// literal left out is NotYet at the write, so both y4 rows are NotYet there, "writing a possibly absent
// optional own field", natively and on the JavaScript backend alike (Language, merged on main 24616705,
// Nov 2). Before #krv1hgh the plain write reached main's missing slot and native panicked. A
// required field, MutCount's count over Counted, always has its slot: its write stays a plain one and
// agrees on both backends, and it isn't the widened NotYet, since checkViewWidenedReads weighs reads,
// never writes (Pupa, Nov 2). Mutant: the four write sites routed again, and the required field's write is
// routed to the checked write.
func TestBrandedSupertypeWriteIsNotRouted(t *testing.T) {
	t.Parallel()
	const y4 = `interface Mut extends Node { name?: string; }
function show<T extends Mut>(x: T): void { if (isTypeElement(x)) { x.name = "w"; console.log(String(x.name)); } }
type Branded = Mut & { _declarationBrand: undefined; _typeElementBrand: undefined };
`
	const counted = `interface Counted extends Declaration { _countedBrand: undefined; count: number; }
function isCountedLie(n: Node): n is Counted { return kindOf(n) !== 0; }
interface MutCount extends Node { count: number; }
function show<T extends MutCount>(x: T): void { if (isCountedLie(x)) { x.count = 7; console.log(String(x.count)); } }
`
	for _, probe := range []struct {
		name, source, field, node string
		notYet                    bool
	}{
		{"y4", y4 + `const v: Branded = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined };
const mv: Mut = v; show(mv);`, "name", "w\n", true},
		{"y4 holding a name", y4 + `const v: Branded = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" };
const mv: Mut = v; show(mv);`, "name", "w\n", true},
		{"required field", counted + `const w = { kind: 5, pos: 1, count: 3 }; const m: MutCount = w; show(m);`, "count", "7\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := brandedPrelude + probe.source
			if probe.notYet {
				_, err := lowerTypeScriptAssertionSource(t, source)
				var pending *NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), "writing a possibly absent optional own field") {
					t.Fatalf("want NotYet at the write, writing a possibly absent optional own field, got %v", err)
				}
				return
			}
			program := lowersAndAgreesWithNodeAs(t, source, ".ts")
			if code, out, errs := brandedNative(t, program); code != 0 || out != probe.node || errs != "" {
				t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q where Node prints %q", code, out, errs, probe.node)
			}
			reads, writes := 0, 0
			inspect := func(node any) bool {
				if read, ok := node.(ir.Property); ok && read.Name == probe.field {
					reads++
					if !program.CheckedField(read.ViewReceiverTypeID, read.Name) {
						t.Errorf("x.%s's read isn't checked through the branded supertype's view", probe.field)
					}
				}
				if write, ok := node.(ir.SetProperty); ok && write.Name == probe.field {
					writes++
					if program.CheckedField(write.ViewReceiverTypeID, write.Name) {
						t.Errorf("the write of x.%s is routed to the branded supertype's checked write", probe.field)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if reads != 1 || writes != 1 {
				t.Fatalf("want one read and one write of %s, found %d and %d", probe.field, reads, writes)
			}
		})
	}
}

// A union or optional field read in a program with record storage is NotYet when a view registered its
// field (rule a, checkRecordGuardedReads), and the guard asks the id the read is checked under: a read
// through Sized & Node is routed to Sized's view, so it's held as one through Sized is. Asking the raw id,
// 148e6df3 lowered the routed read where the plain one stopped (Lower's must, probed on Pupa, Nov 2).
// Mutant: the record guard back on the raw id, and the routed reads lower.
func TestBrandedSupertypeRecordGuardIsRouted(t *testing.T) {
	t.Parallel()
	const records = `const counts: Record<string, number> = {};
counts["a"] = 1;
interface Sized extends Declaration { _sizedBrand: undefined; readonly size?: number; readonly label: string | number; }
function isSizedLie(n: Node): n is Sized { return kindOf(n) !== 0; }
`
	const body = ` { if (isSizedLie(n)) { console.log(n.size === undefined ? "none" : String(n.size)); console.log(String(n.label)); } }
`
	for _, probe := range []struct{ name, show, value string }{
		{"plain honest", "function show(n: Node): void", `{ kind: 80, pos: 2, size: 3, label: "x" }`},
		{"plain lying", "function show(n: Node): void", `{ kind: 80, pos: 2, size: "big", label: true }`},
		{"routed honest", "function show<T extends Node>(n: T): void", `{ kind: 80, pos: 2, size: 3, label: "x" }`},
		{"routed lying", "function show<T extends Node>(n: T): void", `{ kind: 80, pos: 2, size: "big", label: true }`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, brandedPrelude+records+probe.show+body+"const w = "+probe.value+"; const v: Node = w; show(v);\nconsole.log(String(counts[\"a\"]));\n")
			var pending *NotYet
			if !errors.As(err, &pending) || pending.What != "a union or optional field read in a program with record storage" {
				t.Fatalf("want NotYet at the n.size read, held by record storage, got %v", err)
			}
		})
	}
}

// A generic helper is judged per instance: pick at Identifier lowers, and pick at TypeElement over the
// same PropertySignature | Identifier argument is NotYet at its own call, in a program that lowers the
// Identifier instance first. Mutant: one judgment kept for every instance of a declaration, and the second
// program lowers. A helper whose T stands nowhere in its signature is lowered with T unread, one body for
// every instance, and a generic arrow has no instance door, so its T is never mapped: both are NotYet at
// their claim whatever they're called with (Adamic's ruling and Lower's pin, Nov 1). Mutant: the unread
// NotYet removed, and both lower, the lying instance included. A helper the program never instantiates
// isn't judged at all, so a program that registers S and never calls it lowers, unlike under the
// whole-program walk. A closure made after a stopped call that reads the narrowed argument stops with its
// function (Compiler's condition).
func TestBrandedSupertypeIsJudgedWhereItIsLowered(t *testing.T) {
	t.Parallel()
	const pick = `function isIdentifier(n: Node): n is Identifier { return kindOf(n) === 80; }
function pick<T extends Node>(x: PropertySignature | Identifier, test: (n: Node) => n is T): T | undefined { if (test(x)) { return x; } return undefined; }
function never<T extends Node>(x: PropertySignature | Identifier, test: (n: Node) => n is T): string { if (test(x)) { return "yes"; } return "no"; }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log(String(n.pos)); } }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
show(i);
const picked = pick(i, isIdentifier);
console.log(picked === undefined ? "no" : picked.escapedText);
`
	lowersAndAgreesWithNodeAs(t, brandedPrelude+pick, ".ts")
	for _, probe := range []struct{ name, source, want string }{
		{"second instance", pick + `const element = pick(i, isTypeElementLie);
console.log(element === undefined ? "no" : String(element.pos));
`, "which isn't its supertype"},
		{"unread", `function isIdentifier(n: Node): n is Identifier { return kindOf(n) === 80; }
function pickName<T extends Node>(x: PropertySignature | Identifier, test: (n: Node) => n is T): string { if (test(x)) { return "yes"; } return "no"; }
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
console.log(pickName(i, isIdentifier));
console.log(pickName(i, isTypeElementLie));
`, "which this body is lowered without knowing"},
		{"generic arrow", `function isIdentifier(n: Node): n is Identifier { return kindOf(n) === 80; }
const pickArrow = <T extends Node>(x: PropertySignature | Identifier, test: (n: Node) => n is T): string => test(x) ? "yes" : "no";
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
console.log(pickArrow(i, isIdentifier));
console.log(pickArrow(i, isTypeElementLie));
`, "which this body is lowered without knowing"},
		{"closure after the call", `const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
function show(x: PropertySignature | Identifier): () => string { if (isTypeElementLie(x)) { return () => x.typeText; } return () => "no"; }
console.log(show(i)());
`, "which isn't its supertype"},
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

// A member S is assignable to and that's assignable to S keeps its own type when the checker narrows it
// (getNarrowedType takes a subtype of the candidate as it is), so brandedNarrowing passes it and S's view
// never sees its reads (Compiler's pin on 9e4f2184): a restatement, an alias of S's shape, and a twin whose
// name is any, each called directly and through the predicate held as a value. A direct call's false side
// narrows such a member to never, so it stops at the false side's guard; held as a value, only the true side
// is checked, so the read runs. Each is held to never read raw: a program lowers and agrees with Node, or stops at exit 70
// with a located failure on sanitized native and the JavaScript backend, or is NotYet. The restatement and
// the alias can only hold a lie a cast of its own put there, which that cast's view checks; the twin holds
// one by plain assignment.
func TestBrandedSupertypeMutualMemberNeverReadsRaw(t *testing.T) {
	t.Parallel()
	const mutual = `interface Restated extends TypeElement {}
type Shape = { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: string };
interface Twin { readonly kind: number; readonly pos: number; _declarationBrand: undefined; _typeElementBrand: undefined; readonly name?: any }
`
	// Each probe declares only the function it calls, so one probe's stop can't hide another's read.
	const direct = `function via(x: T): string { if (isTypeElementLie(x)) { return String(x.name); } return "no"; }
`
	const held = `const guard = isTypeElementLie;
function via(x: T): string { if (guard(x)) { return String(x.name); } return "no"; }
`
	for _, probe := range []struct{ name, via, of, source string }{
		{"restatement honest", direct, "Restated", `const r: Restated = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(via(r));`},
		{"restatement through a cast", direct, "Restated", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(via(n as Restated));`},
		{"alias honest", direct, "Shape", `const r: Shape = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(via(r));`},
		{"alias through a cast", direct, "Shape", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(via(n as Shape));`},
		{"twin", direct, "Twin", `const t: Twin = { kind: 80, pos: 2, _declarationBrand: undefined, _typeElementBrand: undefined, name: 5 }; console.log(via(t));`},
		// Held as a value, the claim's false side isn't checked, so the true side's read is what runs.
		{"held restatement honest", held, "Restated", `const r: Restated = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(via(r));`},
		{"held restatement through a cast", held, "Restated", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(via(n as Restated));`},
		{"held alias honest", held, "Shape", `const r: Shape = { kind: 171, pos: 1, _declarationBrand: undefined, _typeElementBrand: undefined, name: "a" }; console.log(via(r));`},
		{"held alias through a cast", held, "Shape", `const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const n: Node = w; console.log(via(n as Shape));`},
		{"held twin", held, "Twin", `const t: Twin = { kind: 80, pos: 2, _declarationBrand: undefined, _typeElementBrand: undefined, name: 5 }; console.log(via(t));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := brandedPrelude + mutual + strings.ReplaceAll(probe.via, "x: T", "x: "+probe.of) + probe.source
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
