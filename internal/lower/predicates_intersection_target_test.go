package lower

import (
	"strings"
	"testing"
)

const intersectionExpr = "interface Expr { readonly kind: number; readonly pos: number; readonly operator: { readonly kind: number } }\n"

// A checked predicate whose target is an intersection of plain object types with one literal kind takes its
// interface twin's road (#c7w6j0y's compound-node cause, tsc's Expr & { kind: 9; operator: { kind: 1 } } and
// EndOfFileToken's Token & JSDocContainer): the tag at the call, and the view at every read. Each honest row
// agrees with Node on the JavaScript backend and natively under the sanitizers. Mutant: intersections kept
// out of the branch, and every row is main's 'checked predicate overload target'.
func TestCheckedPredicateIntersectionTargetAgrees(t *testing.T) {
	t.Parallel()
	const show = "function show(n: Expr): void { if (isComma(n)) console.log(String(n.operator.kind)); else console.log('no'); }\n"
	const eof = "interface NodeBase { readonly kind: number }\ninterface Token extends NodeBase { readonly kind: 9 }\ninterface Docs { readonly doc?: string }\ninterface Block extends NodeBase { readonly kind: 10; readonly doc?: string }\ntype End = Token & Docs;\nfunction canHaveDocs(n: NodeBase): n is End | Block { return n.kind === 9 || n.kind === 10; }\nfunction show(n: NodeBase): void { if (canHaveDocs(n)) console.log(n.doc === undefined ? 'none' : n.doc); else console.log('no'); }\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"compound", intersectionExpr + "type Comma = Expr & { readonly kind: 9; readonly operator: { readonly kind: 1 } };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9 && n.operator.kind === 1; }\n" + show + "const e: Comma = { kind: 9, pos: 0, operator: { kind: 1 } };\nconst other: Expr = { kind: 3, pos: 0, operator: { kind: 1 } };\nshow(e);\nshow(other);\n", "1\nno\n"},
		{"flat", intersectionExpr + "type Comma = Expr & { readonly kind: 9 };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9; }\n" + show + "const e: Comma = { kind: 9, pos: 0, operator: { kind: 4 } };\nshow(e);\n", "4\n"},
		{"end of file", eof + "const token: End = { kind: 9, doc: 'end' };\nconst bare: End = { kind: 9 };\nconst block: Block = { kind: 10 };\nshow(token);\nshow(bare);\nshow(block);\n", "end\nnone\nnone\n"},
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

// A lie in the nested part never passes: the view checks the nested literal and the nested layout where
// they're read, as it does for the interface twin, and stops at exit 70 where Node reads on. Mutant: the
// view's registration skipped for an intersection, and these read the lie.
func TestCheckedPredicateIntersectionTargetNestedLiesStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stderr string }{
		{"nested value", intersectionExpr + "type Comma = Expr & { readonly kind: 9; readonly operator: { readonly kind: 1 } };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9; }\nfunction show(n: Expr): void { if (isComma(n)) console.log(String(n.operator.kind)); }\nconst e: Expr = { kind: 9, pos: 0, operator: { kind: 2 } };\nshow(e);\n", "adamic: panic: field read failed: n.operator.kind expected 1, found number 2"},
		{"nested layout", "interface Op { readonly kind: number }\ninterface Wide extends Op { readonly kind: 1; readonly text: string }\ninterface Expr { readonly kind: number; readonly operator: Op }\ntype Comma = Expr & { readonly kind: 9; readonly operator: Wide };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9; }\nfunction show(n: Expr): void { if (isComma(n)) console.log(n.operator.text); }\nconst e: Expr = { kind: 9, operator: { kind: 1 } };\nshow(e);\n", "adamic: panic: field read failed: n.operator.text is not initialized; expected string, found missing"},
		// Lower's optional-beside-required road: the merged pos is required, so a value without it is a lie
		// the view stops at the read (J-cyvtd2k, where Node printed "9 undefined").
		{"optional and required parts", "type Nine = { readonly kind: 9; readonly pos?: number } & { readonly pos: number };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\nfunction show(n: { readonly kind: number }): void { if (isNine(n)) console.log(String(n.kind) + ' ' + String(n.pos)); }\nconst bare: { readonly kind: number } = { kind: 9 };\nshow(bare);\n", "adamic: panic: field read failed: n.pos is not initialized; expected number, found missing"},
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

// An intersection that isn't one literal kind, or that has a class instance among its parts, keeps main's
// stop. Mutant: the literal test dropped, and the 9 | 10 row admits.
func TestCheckedPredicateIntersectionTargetOutsideStays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"two kinds", intersectionExpr + "type Comma = Expr & { readonly kind: 9 | 10 };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9 || n.kind === 10; }\nfunction show(n: Expr): void { if (isComma(n)) console.log(String(n.pos)); }\nshow({ kind: 9, pos: 0, operator: { kind: 1 } });\n"},
		{"class part", "class Base { readonly kind: number = 9; readonly pos: number = 0; }\ninterface Tagged { readonly kind: 9 }\ntype Comma = Base & Tagged;\nfunction isComma(n: Base): n is Comma { return n.kind === 9; }\nfunction show(n: Base): void { if (isComma(n)) console.log(String(n.pos)); }\nshow(new Base());\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), "checked predicate overload target") {
				t.Fatalf("want main's stop, got %v", err)
			}
		})
	}
}

// Lower's roads (#c7w6j0y, on the design): an intersection's merged property is a synthetic symbol that
// carries neither the getter nor the method flag of the part declaring it, so a getter must stop or
// dispatch, never read a slot. Each stops at today's words for its read (J-zr8d3qg): a getter in a part at
// the accessor alias, a method at the call. With the view skipped the two getter rows lower and dispatch,
// agreeing with Node, so their stop is the view's and fails closed.
func TestCheckedPredicateIntersectionTargetSyntheticSymbolsStop(t *testing.T) {
	t.Parallel()
	const tail = "function show(n: { readonly kind: number }): void { if (isNine(n)) console.log(String(n.kind) + ' ' + String(n.pos)); else console.log('no'); }\n"
	for _, probe := range []struct{ name, source, stop string }{
		{"getter part", "interface Token { readonly kind: number; readonly pos: number }\ntype Nine = Token & { get kind(): 9 };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\n" + tail + "class Getter { get kind(): 9 { return 9; } readonly pos: number = 3; }\nshow(new Getter());\nconst plain = { kind: 9, pos: 4 };\nshow(plain);\n", "a checked field alias of an accessor"},
		{"field and getter parts", "type Nine = { readonly kind: 9; readonly pos: number } & { get kind(): 9 };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\n" + tail + "const plain = { kind: 9, pos: 4 };\nshow(plain);\n", "a checked field alias of an accessor"},
		{"method and field parts", "type Nine = { readonly kind: 9; readonly pos: number; run(): number } & { readonly run: () => number };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\nfunction show(n: { readonly kind: number }): void { if (isNine(n)) console.log(String(n.run())); }\nconst plain = { kind: 9, pos: 1, run(): number { return 5; } };\nshow(plain);\n", "a call to a PropertyAccessExpression"},
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

// Language's storage-side roads (#c7w6j0y, on the design): parts whose members a view may not check. None
// lowers raw (J-f2jpj73). A phantom brand typed undefined is never stored, and an honest value agrees.
// Callable, array and record parts are kept out by plainObjectIntersection and keep main's stop. A nested
// getter stops at the accessor alias, and a method at rulings.md:27's check at the call. A generic part
// keeps the stop for a claim its body is lowered without knowing. The symbol-keyed part stops before the
// rule, at the symbol itself, so it says nothing of the rule until a symbol lowers; it's pinned so that day
// shows.
func TestCheckedPredicateIntersectionTargetStorageParts(t *testing.T) {
	t.Parallel()
	const value = "const e: Expr = { kind: 9, pos: 2, operator: { kind: 1 } };\n"
	const guard = "function isT(n: Expr): n is T { return n.kind === 9; }\n"
	const showPos = "function show(n: Expr): void { if (isT(n)) console.log(String(n.pos)); else console.log('no'); }\n"
	t.Run("phantom brand part", func(t *testing.T) {
		t.Parallel()
		program := lowersAndAgreesWithNodeAs(t, intersectionExpr+"type T = Expr & { readonly kind: 9; readonly __tagged: undefined };\n"+guard+showPos+value+"show(e);\n", ".ts")
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n" {
			t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "2\n")
		}
	})
	for _, probe := range []struct{ name, source, stop string }{
		{"callable part", intersectionExpr + "type T = Expr & ((x: number) => number);\n" + guard + showPos + value + "show(e);\n", "checked predicate overload target T"},
		{"array part", "interface Expr { readonly kind: number; readonly pos: number }\ntype T = readonly Expr[] & { readonly kind: 9 };\nfunction isT(n: { readonly kind: number }): n is T { return n.kind === 9; }\nfunction show(n: { readonly kind: number }): void { if (isT(n)) console.log(String(n.length)); else console.log('no'); }\nconst e: Expr = { kind: 9, pos: 2 };\nshow(e);\n", "checked predicate overload target T"},
		{"record part", intersectionExpr + "type T = Expr & Record<string, unknown> & { readonly kind: 9 };\n" + guard + showPos + value + "show(e);\n", "checked predicate overload target T"},
		{"nested getter part", intersectionExpr + "type T = Expr & { readonly kind: 9; readonly operator: { get kind(): 1 } };\n" + guard + "function show(n: Expr): void { if (isT(n)) console.log(String(n.operator.kind)); else console.log('no'); }\n" + value + "show(e);\n", "a checked field alias of an accessor"},
		{"method part", intersectionExpr + "type T = Expr & { readonly kind: 9; describe(): string };\n" + guard + "function show(n: Expr): void { if (isT(n)) console.log(n.describe()); else console.log('no'); }\nconst v = { kind: 9, pos: 2, operator: { kind: 1 }, describe(): string { return 'described'; } };\nshow(v);\n", "a callable field describe read through a checked view"},
		{"symbol keyed part", intersectionExpr + "const sym: unique symbol = Symbol('s');\ntype T = Expr & { readonly kind: 9; readonly [sym]: number };\n" + guard + "function show(n: Expr): void { if (isT(n)) console.log(String(n[sym])); else console.log('no'); }\n" + value + "show(e);\n", "a value of type unique symbol"},
		{"generic part", intersectionExpr + "function isK<T>(n: Expr): n is Expr & T & { readonly kind: 9 } { return n.kind === 9; }\nfunction show<T>(n: Expr): void { if (isK<T>(n)) console.log(String(n.pos)); else console.log('no'); }\n" + value + "show<{ readonly pos: number }>(e);\nshow<{ readonly kind: 8 }>(e);\n", "which this body is lowered without knowing"},
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
