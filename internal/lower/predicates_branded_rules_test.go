package lower

import (
	"errors"
	"strings"
	"testing"
)

// brandedRulesPrelude is core.ts's orPredicate2 (Correctness's adaptation) and tsc's overloaded
// findAncestor, whose implementation calls its callback through a plain boolean signature
// (utilitiesPublic.ts:786-788), beside predicates for an unbranded Literal and Keyword.
const brandedRulesPrelude = brandedPrelude + `interface Literal extends Node { readonly kind: 9; readonly text: string; }
interface Keyword extends Node { readonly kind: 100; readonly word: string; }
function isIdentifier(n: Node): n is Identifier { return kindOf(n) === 80; }
function isLiteral(n: Node): n is Literal { return kindOf(n) === 9; }
function isKeyword(n: Node): n is Keyword { return kindOf(n) === 100; }
function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 { return p => f1(p) || f2(p); }
function findAncestor<T extends Node>(node: Node | undefined, callback: (element: Node) => element is T): T | undefined;
function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined;
function findAncestor(node: Node | undefined, callback: (element: Node) => boolean | "quit"): Node | undefined {
	if (node === undefined) { return undefined; }
	const result = callback(node);
	if (result === "quit") { return undefined; }
	else if (result) { return node; }
	return undefined;
}
const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
const l: Literal = { kind: 9, pos: 4, text: "t" };
`

// eitherId is a checked predicate over a branded union, called in a statement that reads only its true
// side, so the call's check needs no false direction, which a branded supertype's membership can't give.
const eitherIdentifier = `function eitherId(n: Node): n is Identifier | TypeElement { return isIdentifier(n) || isTypeElementLie(n); }
`

// Rule 1 (Adamic's ruling on #fx9waqy, Nov 7, in Compiler's form): orPredicate2's body, `p => f1(p) ||
// f2(p)`, stands with no true-side read at a branded instance, since neither call's narrowing reaches a
// read and the arrow's result is checked at every call through orPredicate2's written type. A function
// whose chain leaves the binding for anything but another claimed call's argument stays NotYet: a field
// read after the chain (Compiler's pin), and a closure created after the narrowing, which TypeScript 5.4
// and later keeps narrowed (Language's pin, in the `||` form, where rule 1 is what judges it). Mutant: the
// use walk dropped from brandedChainCall, and both stops lower.
func TestBrandedCallRuleOneChain(t *testing.T) {
	t.Parallel()
	// The branded instance made and never called: rule 1 admits the body's own branded calls, and the
	// direction walk wants no uses of the arrow orPredicate2 returns under its written claim.
	lowersAndAgreesWithNodeAs(t, brandedRulesPrelude+`const element = orPredicate2(isIdentifier, isTypeElement);
console.log(typeof element);
`, ".ts")
	for _, probe := range []struct{ name, source string }{
		{"a field read after the chain", `function orRead<P extends Node, R1 extends P>(f1: (p1: P) => p1 is R1): (p: P) => p is R1 { return (p: P): p is R1 => f1(p) || p.pos !== 0; }
const element = orRead(isTypeElement);
console.log(typeof element);
`},
		{"a closure after the narrowing", `function orLater<P extends Node, R1 extends P>(f1: (p1: P) => p1 is R1): (p: P) => p is R1 { return (p: P): p is R1 => f1(p) || (() => p.pos !== 0)(); }
const element = orLater(isTypeElement);
console.log(typeof element);
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, brandedRulesPrelude+probe.source)
			var pending *NotYet
			if !errors.As(err, &pending) {
				t.Fatalf("want NotYet, got %v", err)
			}
			t.Logf("stops at %v", err)
		})
	}
}

// Rule 2(b): a call whose narrowing dropped nothing, by the checker's own identity of the read's type and
// the argument's, stands, since a lie changes no static type. PropertySignature | Identifier claimed as
// Identifier | TypeElement keeps both members, so even a lying TypeElement arm reads each value as what it
// is, in either order, and so does an argument narrowed by its assignment. PropertySignature | Literal
// drops Literal and keeps PropertySignature, a subtype of TypeElement no view checks, so it stays NotYet.
// Mutant: 2(b) compares the concreted types member by member, admitting the dropped member, and the last
// probe lowers.
func TestBrandedCallRuleTwoIdentity(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		eitherIdentifier + `function show(x: PropertySignature | Identifier): void { eitherId(x) && console.log(String(x.pos)); }
show(i);
`,
		eitherIdentifier + `function show(x: Identifier | PropertySignature): void { eitherId(x) && console.log(String(x.pos)); }
show(i);
`,
		`const p: PropertySignature = { kind: 171, pos: 1, name: "a", typeText: "t", _declarationBrand: undefined, _typeElementBrand: undefined };
function show(): void { let x: PropertySignature | Literal = p; isTypeElementLie(x) && console.log(x.typeText); }
show();
`,
	} {
		lowersAndAgreesWithNodeAs(t, brandedRulesPrelude+source, ".ts")
	}
	_, err := lowerTypeScriptAssertionSource(t, brandedRulesPrelude+eitherIdentifier+`function show(x: PropertySignature | Literal): void { eitherId(x) && console.log(x.typeText); }
show(l);
`)
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "which isn't its supertype") {
		t.Fatalf("want NotYet naming the dropped member, got %v", err)
	}
}

// Rule 2(c): Identifier | Literal claimed as Identifier | TypeElement drops Literal and keeps Identifier,
// an unbranded arm with its own tag, so it stands, and the Identifier arm's view, which the call's check
// registered, checks every read narrowed to it. A Literal said to be a TypeElement passes the check by the
// branded arm, which checks only that it's an object, and is then read as an Identifier: it stops at that
// read, on sanitized native and the JavaScript backend, never read raw. The honest twin agrees with Node.
// An arm whose tag shares its value with another member of its own enum tells nothing apart (rulings.md:13,
// tag_values.go), so it stays NotYet. Mutant: brandedTaggedArm without enumValueUnique, and the alias
// lowers.
func TestBrandedCallRuleTwoTaggedArm(t *testing.T) {
	t.Parallel()
	const show = eitherIdentifier + `function show(x: Identifier | Literal): void { eitherId(x) && console.log(x.escapedText); }
`
	lowersAndAgreesWithNodeAs(t, brandedRulesPrelude+show+"show(i);\n", ".ts")
	brandedStops(t, brandedRulesPrelude+show+"show(l);\n", "adamic: panic: ")
	_, err := lowerTypeScriptAssertionSource(t, brandedPrelude+`enum K { First = 1, Alias = 1 }
interface X extends Node { readonly kind: K.First; readonly x: string; }
interface Y extends Node { readonly kind: K.Alias; readonly y: number; }
function isX(n: Node): n is X { return kindOf(n) === 1; }
function eitherX(n: Node): n is X | TypeElement { return isX(n) || isTypeElementLie(n); }
function show(v: X | Y): void { eitherX(v) && console.log(v.x); }
show({ kind: K.Alias, pos: 0, y: 5 });
`)
	var pending *NotYet
	if !errors.As(err, &pending) {
		t.Fatalf("want NotYet for a tag another member of its enum holds, got %v", err)
	}
	t.Logf("the alias stops at %v", err)
}

// The producer rule (Adamic's ruling on #fx9waqy, Nov 7): a call of orPredicate2 handed to findAncestor's
// trusted callback proves its claim when orPredicate2's returned arrow proves it by flow and every
// predicate the call hands it proves its own, judged at that call's instance, a nested producer included.
// A lie carried in a value, which proves nothing, is refused where it's handed, as it would be passed
// directly. Inside the producer, a call through f1 keeps its check whoever calls it (Language's
// condition): the same lie handed to orPredicate2 itself, outside any trusted road, lowers, and stops at
// that check. Mutant: predicateProducerParameter answering no, so f1's claim is trusted by collecting the
// predicates passed at its call sites; the lie is then refused at orPredicate2's call, and the last probe
// goes red. One program holds an unbranded and a branded instance, each judged at its own door: the
// branded lie stops at its read through S's view (Language's per-instance pin).
func TestBrandedCallProducer(t *testing.T) {
	t.Parallel()
	// A measurement, not a pin: tsc's own findAncestor walks up by reassigning node, which the census's
	// overload proof may refuse as returning a value its predicate didn't check.
	_, tscShape := lowerTypeScriptAssertionSource(t, `interface N { readonly kind: number; readonly parent: N | undefined; }
interface I extends N { readonly kind: 80; readonly text: string; }
function findAncestor<T extends N>(node: N | undefined, callback: (element: N) => element is T): T | undefined;
function findAncestor(node: N | undefined, callback: (element: N) => boolean | "quit"): N | undefined;
function findAncestor(node: N | undefined, callback: (element: N) => boolean | "quit"): N | undefined {
	while (node) {
		const result = callback(node);
		if (result === "quit") { return undefined; }
		else if (result) { return node; }
		node = node.parent;
	}
	return undefined;
}
function isI(n: N): n is I { return n.kind === 80; }
const top: I = { kind: 80, text: "x", parent: undefined };
const leaf: N = { kind: 1, parent: top };
const found = findAncestor(leaf, isI);
console.log(found === undefined ? "none" : found.text);
`)
	t.Logf("tsc's findAncestor shape: %v", tscShape)
	// Road (B): orPredicate2's arrow is a checked closure under the producer's written claim, so findAncestor's
	// trusted callback takes the producer's result as it takes a checked closure passed directly, at an
	// unbranded, a nested and a branded instance alike.
	lowersAndAgreesWithNodeAs(t, brandedRulesPrelude+`const found = findAncestor(i, orPredicate2(isIdentifier, isLiteral));
console.log(found === undefined ? "none" : String(found.pos));
const nested = findAncestor(l, orPredicate2(orPredicate2(isIdentifier, isLiteral), isKeyword));
console.log(nested === undefined ? "none" : String(nested.pos));
const element = findAncestor(i, orPredicate2(isIdentifier, isTypeElement));
console.log(element === undefined ? "none" : String(element.pos));
`, ".ts")
	// A lie carried in a value is neither proven nor checked, so it's refused where it's handed, an unbranded
	// arm and a branded one alike. Mutant: predicateProducerProven without predicateArgumentProof's
	// recursion, and both lower.
	for _, liar := range []string{
		`const liar: (n: Node) => n is Literal = pick ? isLiteral : isLiteralLie;
const found = findAncestor(l, orPredicate2(liar, isKeyword));`,
		`const liar: (n: Node) => n is TypeElement = pick ? isTypeElement : isTypeElementLie;
const found = findAncestor(i, orPredicate2(isIdentifier, liar));`,
	} {
		_, err := lowerTypeScriptAssertionSource(t, brandedRulesPrelude+`function isLiteralLie(n: Node): n is Literal { return n.pos > 0; }
const pick = kindOf(i) > 1000;
`+liar+`
console.log(found === undefined ? "none" : String(found.pos));
`)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "an unproven predicate argument for parameter") {
			t.Fatalf("want the lie refused where it's handed, got %v", err)
		}
	}
	// A lying f2, a checked declaration handed by name, stops before any caller reads the value: at the
	// trusted site, through an immediate call, and through a const alias, which R-T80's erasure never
	// reaches, since a checked closure's return check isn't a signature's own check (Language's two roads).
	// Inside the producer f2's own call is checked too (Language's condition), and it stops first, so a
	// mutant dropping only the arrow's return check survives here; one dropping both reads the Keyword raw.
	const lie = `function isLiteralLie(n: Node): n is Literal { return n.pos > 0; }
const k: Keyword = { kind: 100, pos: 7, word: "w" };
`
	for _, source := range []string{
		`const found = findAncestor(k, orPredicate2(isIdentifier, isLiteralLie));
console.log(found === undefined ? "none" : String(found.pos));
`,
		`console.log(orPredicate2(isIdentifier, isLiteralLie)(k) ? "named" : "other");
`,
		`const either = orPredicate2(isIdentifier, isLiteralLie);
function show(x: Node): void { either(x) && console.log(String(x.pos)); }
show(k);
`,
	} {
		brandedStops(t, brandedRulesPrelude+lie+source, "adamic: panic: ")
	}
	// Language's condition: a lie handed to orPredicate2 itself through a value passes the scan, since f1's
	// claim is checked inside the producer, and stops at that check. Mutant: predicateProducerParameter
	// answering no, and the scan refuses it instead.
	brandedStops(t, brandedRulesPrelude+`function isLiteralLie(n: Node): n is Literal { return n.pos > 0; }
const pick = kindOf(i) > 1000;
const liar: (n: Node) => n is Literal = pick ? isLiteral : isLiteralLie;
const either = orPredicate2(liar, isKeyword);
function show(x: Node): void { either(x) && console.log(String(x.pos)); }
show(i);
`, "adamic: panic: ")
	// One program holds an unbranded and a branded instance, each judged at its own door: the branded lie
	// stops at its read through S's view (Language's per-instance pin), and so does findAncestor's branded
	// result read by a caller (Compiler's pin).
	brandedStops(t, brandedRulesPrelude+`const named = orPredicate2(isIdentifier, isLiteral);
const element = orPredicate2(isTypeElementLie, isTypeElement);
function show(n: Node): void { element(n) && console.log(String(n.name)); }
const w = { kind: 9, pos: 4, text: "t", name: 5 };
show(w);
`, "adamic: panic: ")
	brandedStops(t, brandedRulesPrelude+`const w = { kind: 9, pos: 4, text: "t", name: 5 };
const found = findAncestor(w, orPredicate2(isTypeElementLie, isTypeElement));
console.log(found === undefined ? "none" : String(found.name));
`, "adamic: panic: ")
}

// The inferred arrow's walk (road (B) on #fx9waqy, Nov 7): an arrow its producer returns written in place
// under its own written checked claim is a checked closure, which the direction walk reads as it reads an
// annotated one, so it wants none of its uses. A caller reading a
// lying instance's false side over a closed tag domain stops at compile time, as any checked closure's does. Every other
// shape keeps today's NotYet: an arrow held in a module array before it's returned (Compiler's pin), an
// unannotated helper whose result goes to an untyped const, an implementation written as a plain boolean
// function under an overload that claims, and a named function expression that stores itself (Language's
// pins). Mutants, each in checkedPredicateClaimNode: the written-type condition dropped (the untyped const
// lowers), the in-place condition dropped (the array lowers), overload 1's result read (the overload
// lowers), the name allowed (self lowers).
func TestInferredArrowUnderWrittenClaim(t *testing.T) {
	t.Parallel()
	// A caller reading the lying instance's false side is NotYet at compile time, as every checked closure's
	// false side read through its value is today (predicate_false_flow.go): measured on Cloud at fecfdfe4.
	wantStop(t, brandedPrelude+`type TokenKind = 1 | 2 | 3;
interface Token { readonly kind: TokenKind; readonly pos: number; }
interface Name extends Token { readonly kind: 1; readonly text: string; }
interface Count extends Token { readonly kind: 2; readonly value: number; }
interface Operator extends Token { readonly kind: 3; readonly op: string; }
type AnyToken = Name | Count | Operator;
function orPredicate2<P, R1 extends P, R2 extends P>(f1: (p1: P) => p1 is R1, f2: (p2: P) => p2 is R2): (p: P) => p is R1 | R2 { return p => f1(p) || f2(p); }
function isName(t: AnyToken): t is Name { return t.kind === 1; }
function isCountLie(t: AnyToken): t is Count { return t.pos > 100; }
const either = orPredicate2(isName, isCountLie);
function show(t: AnyToken): void { if (either(t)) { console.log("named"); } else { console.log(t.op); } }
const c: Count = { kind: 2, pos: 5, value: 1 };
show(c);
`, "a call through a predicate value whose false answer isn't checked yet")
	for _, probe := range []struct{ name, source string }{
		{"held in a module array", `const held: ((p: Node) => p is Identifier | Literal)[] = [];
export function orHeld(f1: (p: Node) => p is Identifier, f2: (p: Node) => p is Literal): (p: Node) => p is Identifier | Literal { const g = (p: Node) => f1(p) || f2(p); held.push(g); return g; }
const g = orHeld(isIdentifier, isLiteral);
console.log(typeof g);
`},
		{"an untyped const", `export function orBare(f1: (p: Node) => p is Identifier, f2: (p: Node) => p is Literal) { return (p: Node) => f1(p) || f2(p); }
const g = orBare(isIdentifier, isLiteral);
function show(x: Identifier | Literal | Keyword): void { if (!g(x)) { console.log(x.word); } }
show({ kind: 100, pos: 1, word: "w" });
`},
		{"an overload's claim", `export function orOver(f: (p: Node) => p is Identifier): (p: Node) => p is Identifier;
export function orOver(f: (p: Node) => p is Identifier): (p: Node) => boolean { return p => f(p); }
const g = orOver(isIdentifier);
console.log(typeof g);
`},
		{"a named function expression", `const cache: ((p: Node) => boolean)[] = [];
export function orSelf(f1: (p: Node) => p is Identifier, f2: (p: Node) => p is Literal): (p: Node) => p is Identifier | Literal { return function self(p: Node) { cache.push(self); return f1(p) || f2(p); }; }
const g = orSelf(isIdentifier, isLiteral);
console.log(typeof g);
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			wantStop(t, brandedRulesPrelude+probe.source, "whose uses the walk doesn't follow")
		})
	}
}

// wantStop lowers a .ts program and wants a NotYet holding want.
func wantStop(t *testing.T, source, want string) {
	t.Helper()
	_, err := lowerTypeScriptAssertionSource(t, source)
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), want) {
		t.Fatalf("want NotYet with %q, got %v", want, err)
	}
}
