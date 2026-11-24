package lower

import (
	"strings"
	"testing"
)

// The node world of ruling 9's pins (Nov 24, #em1vyta): tsc's NamedDeclaration is a branded supertype whose
// name is optional, and isNamedDeclaration claims the base with a required, refined name.
const refinementWorld = "interface Node { readonly kind: number; readonly pos: number }\n" +
	"interface Declaration extends Node { readonly _declarationBrand: undefined }\n" +
	"interface Identifier extends Node { readonly kind: 1; readonly text: string }\n" +
	"interface StringLiteral extends Node { readonly kind: 2; readonly text: string }\n" +
	"type DeclarationName = Identifier | StringLiteral;\n" +
	"interface MemberName extends Node { readonly kind: 3; readonly left: Identifier }\n" +
	"interface NamedDeclaration extends Declaration { readonly name?: DeclarationName }\n" +
	"const id: Identifier = { kind: 1, pos: 0, text: 'id' };\n" +
	"const literal: StringLiteral = { kind: 2, pos: 0, text: 'literal' };\n"

// isNamedDeclaration as tsc 6.0.3 writes it (utilitiesPublic.ts:946), and a reader of its true side.
const refinementSite = "function isNamedDeclaration(node: Node): node is NamedDeclaration & { name: DeclarationName } { return !!(node as NamedDeclaration).name; }\n" +
	"function show(n: Node): void { if (isNamedDeclaration(n)) console.log(n.name.text); else console.log('no'); }\n"

// Language's r05 (#em1vyta, her J-8bgv5n0): a refined field whose reduced type still holds an intersection
// member, Identifier & Extra, takes predicateMembership's recursion on it.
const refinementReduced = "interface Node { readonly kind: number; readonly pos: number }\n" +
	"interface Declaration extends Node { readonly _declarationBrand: undefined }\n" +
	"interface Identifier extends Node { readonly kind: 1; readonly text: string }\n" +
	"interface StringLiteral extends Node { readonly kind: 2; readonly text: string }\n" +
	"interface Extra { readonly extra: string }\n" +
	"interface Wide extends Declaration { readonly name?: (Identifier & Extra) | StringLiteral }\n" +
	"interface Plain extends Declaration { readonly name?: Node }\n" +
	"function isNamed(node: Node): node is Wide & { name: Identifier | StringLiteral } { return !!(node as Wide).name; }\n" +
	"function show(n: Node): string { return isNamed(n) && n.name.kind === 1 ? n.name.extra.toUpperCase() : 'no'; }\n"

// A branded base with refined parts takes ruling 9's road: the base's road at the call, each refined
// field's own membership beside it, and the views at every read. tsc's site agrees with Node on the
// JavaScript backend and natively under the sanitizers, and so do the shapes the rule admits that aren't
// tsc's. Mutant: brandedRefinement answering nil, and every row is the family's 'checked predicate
// overload target'.
func TestBrandedRefinementAgrees(t *testing.T) {
	t.Parallel()
	const declarations = "const named: NamedDeclaration = { kind: 40, pos: 0, _declarationBrand: undefined, name: id };\n" +
		"const quoted: NamedDeclaration = { kind: 41, pos: 0, _declarationBrand: undefined, name: literal };\n" +
		"const bare: NamedDeclaration = { kind: 42, pos: 0, _declarationBrand: undefined };\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"site", refinementWorld + refinementSite + declarations + "show(named);\nshow(quoted);\nshow(bare);\nshow(id);\n", "id\nliteral\nno\nno\n"},
		{"two refined fields", refinementWorld + "interface Typed extends NamedDeclaration { readonly type?: Identifier }\n" +
			"function isTyped(node: Node): node is Typed & { name: DeclarationName; type: Identifier } { const t = node as Typed; return t.name !== undefined && t.type !== undefined; }\n" +
			"function show(n: Node): void { if (isTyped(n)) console.log(n.name.text + ':' + n.type.text); else console.log('no'); }\n" +
			"const both: Typed = { kind: 43, pos: 0, _declarationBrand: undefined, name: literal, type: id };\nconst half: Typed = { kind: 43, pos: 0, _declarationBrand: undefined, name: literal };\nshow(both);\nshow(half);\n", "literal:id\nno\n"},
		{"reduced intersection member", refinementReduced + "const good: Wide = { kind: 40, pos: 0, _declarationBrand: undefined, name: { kind: 1, pos: 0, text: 't', extra: 'e' } };\nconsole.log(show(good));\n", "E\n"},
		{"in a loop", refinementWorld + "function isNamedDeclaration(node: Node): node is NamedDeclaration & { name: DeclarationName } { return !!(node as NamedDeclaration).name; }\n" +
			"const all: Node[] = [{ kind: 40, pos: 0, _declarationBrand: undefined, name: id } as NamedDeclaration, id, { kind: 41, pos: 0, _declarationBrand: undefined, name: literal } as NamedDeclaration];\n" +
			"let index = 0;\nwhile (index < all.length) { const n = all[index]; if (n !== undefined && isNamedDeclaration(n)) console.log(n.name.text); index++; }\n", "id\nliteral\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// Every lie stops at exit 70 where Node reads on, and each stops where the rule says it does. A JSDocLink
// whose name is a member name, which tsc's body answers true, stops in that body, at its read through
// NamedDeclaration's view, and at the call when the body doesn't read the name as the base's type (the
// ruling's own case).
// Language's condition (a) is two rows: a field the base and a refining part both declare is checked as
// the narrower of the two, so a base narrower than the part stops a value the part alone would pass, and
// a part narrower than the base stops a value the base alone would pass, both at the call. A base field a
// lie carries passes the call and stops at its read through the view. Mutants: the refined fields not
// checked at the call (the JSDocLink and both condition rows read on, or stop at the read, not the call);
// the field checked as the base's type alone (the part-narrower row passes the call); as the part's type
// alone (the base-narrower row passes the call); the intersection's view not registered (the base-field
// row reads a string as a number, or is NotYet as a widened read).
func TestBrandedRefinementLiesStop(t *testing.T) {
	t.Parallel()
	const loose = "interface Named extends Node { readonly name?: Node }\n"
	const jsDocLink = "interface JSDocLink extends Node { readonly kind: 4; readonly name?: Identifier | MemberName }\n" +
		"const member: MemberName = { kind: 3, pos: 0, left: id };\nconst link: JSDocLink = { kind: 4, pos: 0, name: member };\n"
	for _, probe := range []struct{ name, source, stderr string }{
		// tsc's own body reads the name through NamedDeclaration's view, which stops the lie there, before
		// the call's check is reached.
		{"misnamed link in tsc's body", refinementWorld + refinementSite + jsDocLink + "show(link);\n", "adamic: panic: field read failed: (node as NamedDeclaration).name matches no member of DeclarationName | undefined"},
		// A body that never reads the name as the base's type leaves the lie to the call's check.
		{"misnamed link at the call", refinementWorld + loose + jsDocLink +
			"function isNamed(node: Named): node is NamedDeclaration & { name: DeclarationName } { return node.name !== undefined; }\n" +
			"function show(n: Named): void { if (isNamed(n)) console.log(n.name.text); else console.log('no'); }\nshow(link);\n", "adamic: panic: predicate isNamed at "},
		{"base narrower than the part", strings.Replace(refinementWorld, "readonly name?: DeclarationName", "readonly name?: Identifier", 1) + loose +
			"function isNamed(node: Named): node is NamedDeclaration & { name: Identifier | StringLiteral } { return node.name !== undefined; }\n" +
			"function show(n: Named): void { if (isNamed(n)) console.log(n.name.text); else console.log('no'); }\n" +
			"const other: Named = { kind: 50, pos: 0, name: literal };\nshow(other);\n", "adamic: panic: predicate isNamed at "},
		{"part narrower than the base", refinementWorld + loose +
			"function isNamed(node: Named): node is NamedDeclaration & { name: Identifier } { return node.name !== undefined; }\n" +
			"function show(n: Named): void { if (isNamed(n)) console.log(n.name.text); else console.log('no'); }\n" +
			"const quoted: NamedDeclaration = { kind: 41, pos: 0, _declarationBrand: undefined, name: literal };\nshow(quoted);\n", "adamic: panic: predicate isNamed at "},
		{"reduced intersection member", refinementReduced + "const id: Identifier = { kind: 1, pos: 0, text: 'id' };\nconst lie: Plain = { kind: 40, pos: 0, _declarationBrand: undefined, name: id };\nconsole.log(show(lie));\n", "adamic: panic: field read failed: (node as Wide).name"},
		{"base field", refinementWorld + "interface Loose { readonly kind: number; readonly pos: number | string; readonly name?: Node }\n" +
			"function isNamed(node: Loose): node is NamedDeclaration & { name: DeclarationName } { return node.name !== undefined; }\n" +
			"function show(n: Loose): void { if (isNamed(n)) console.log(String(n.pos + 1)); else console.log('no'); }\n" +
			"const raw = { kind: 40, pos: 'p', name: id };\nconst wrong: Loose = raw;\nshow(wrong);\n", "adamic: panic: field read failed: n.pos is not a number"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			brandedStopsProgram(t, program, probe.stderr)
		})
	}
}

// What the rule doesn't admit keeps a stop. Language's condition (b): a refined field a getter may answer,
// a class's or an object literal's, keeps a NotYet in its own words, naming the getter. A refined field
// with no check, a Map, keeps the whole target at the family's stop, as does a refining part the ruling
// doesn't cover: an optional refined field, a branded refining part, two named parts. Mutants: the getter
// test dropped (the two getter rows lower, or stop in other words); the no-check fallback dropped.
func TestBrandedRefinementOutsideStays(t *testing.T) {
	t.Parallel()
	const getterShow = "function isNamedDeclaration(node: Node): node is NamedDeclaration & { name: DeclarationName } { return !!(node as NamedDeclaration).name; }\n" +
		"function show(n: Node): void { if (isNamedDeclaration(n)) console.log(n.name.text); else console.log('no'); }\n"
	for _, probe := range []struct{ name, source, stop string }{
		{"class getter", refinementWorld + getterShow + "class Decl { readonly kind: number = 40; readonly pos: number = 0; readonly _declarationBrand: undefined = undefined; get name(): Identifier { return id; } }\nshow(new Decl());\n", "whose refined field name a getter may answer"},
		{"literal getter", refinementWorld + getterShow + "const decl: NamedDeclaration = { kind: 40, pos: 0, _declarationBrand: undefined, get name(): Identifier { return id; } };\nshow(decl);\n", "whose refined field name a getter may answer"},
		{"no check", refinementWorld + "function isMapped(node: Node): node is NamedDeclaration & { table: Map<string, number> } { return node.kind === 40; }\nfunction show(n: Node): void { if (isMapped(n)) console.log(String(n.table.size)); else console.log('no'); }\nshow(id);\n", "checked predicate overload target"},
		{"optional refined field", refinementWorld + "function isMaybe(node: Node): node is NamedDeclaration & { name?: Identifier } { return node.kind === 40; }\nfunction show(n: Node): void { if (isMaybe(n)) console.log(String(n.pos)); else console.log('no'); }\nshow(id);\n", "checked predicate overload target"},
		{"branded part", refinementWorld + "function isBranded(node: Node): node is NamedDeclaration & { name: DeclarationName; readonly _otherBrand: undefined } { return node.kind === 40; }\nfunction show(n: Node): void { if (isBranded(n)) console.log(n.name.text); else console.log('no'); }\nshow(id);\n", "checked predicate overload target"},
		{"two named parts", refinementWorld + "interface Exported extends Node { readonly _exportBrand: undefined }\nfunction isBoth(node: Node): node is NamedDeclaration & Exported & { name: DeclarationName } { return node.kind === 40; }\nfunction show(n: Node): void { if (isBoth(n)) console.log(n.name.text); else console.log('no'); }\nshow(id);\n", "checked predicate overload target"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want the stop %q, got %v", probe.stop, err)
			}
		})
	}
}
