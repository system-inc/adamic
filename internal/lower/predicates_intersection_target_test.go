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
	for _, probe := range []struct{ name, source string }{
		{"nested value", intersectionExpr + "type Comma = Expr & { readonly kind: 9; readonly operator: { readonly kind: 1 } };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9; }\nfunction show(n: Expr): void { if (isComma(n)) console.log(String(n.operator.kind)); }\nconst e: Expr = { kind: 9, pos: 0, operator: { kind: 2 } };\nshow(e);\n"},
		{"nested layout", "interface Op { readonly kind: number }\ninterface Wide extends Op { readonly kind: 1; readonly text: string }\ninterface Expr { readonly kind: number; readonly operator: Op }\ntype Comma = Expr & { readonly kind: 9; readonly operator: Wide };\nfunction isComma(n: Expr): n is Comma { return n.kind === 9; }\nfunction show(n: Expr): void { if (isComma(n)) console.log(n.operator.text); }\nconst e: Expr = { kind: 9, operator: { kind: 1 } };\nshow(e);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			brandedStopsProgram(t, program, "field read failed")
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
// carries neither the getter nor the method flag of the part declaring it. Logged on the first run, to pin
// what each does: a getter must stop or dispatch, never read a slot.
func TestCheckedPredicateIntersectionTargetSyntheticSymbolsLogged(t *testing.T) {
	const tail = "function show(n: { readonly kind: number }): void { if (isNine(n)) console.log(String(n.kind) + ' ' + String(n.pos)); else console.log('no'); }\n"
	for _, probe := range []struct{ name, source string }{
		{"getter part", "interface Token { readonly kind: number; readonly pos: number }\ntype Nine = Token & { get kind(): 9 };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\n" + tail + "class Getter { get kind(): 9 { return 9; } readonly pos: number = 3; }\nshow(new Getter());\nshow({ kind: 9, pos: 4 });\n"},
		{"field and getter parts", "type Nine = { readonly kind: 9; readonly pos: number } & { get kind(): 9 };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\n" + tail + "show({ kind: 9, pos: 4 });\n"},
		{"method and field parts", "type Nine = { readonly kind: 9; readonly pos: number; run(): number } & { readonly run: () => number };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\nfunction show(n: { readonly kind: number }): void { if (isNine(n)) console.log(String(n.run())); }\nshow({ kind: 9, pos: 1, run(): number { return 5; } });\n"},
		{"optional and required parts", "type Nine = { readonly kind: 9; readonly pos?: number } & { readonly pos: number };\nfunction isNine(n: { readonly kind: number }): n is Nine { return n.kind === 9; }\n" + tail + "const bare: { readonly kind: number } = { kind: 9 };\nshow(bare);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			program, err := lowerSourceAs(t, probe.source, ".ts")
			t.Logf("ROW %s lower: %v", probe.name, err)
			if err == nil {
				observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", probe.source))
				t.Logf("ROW %s node: exit %d stdout %q", probe.name, observed.code, observed.stdout)
				code, stdout, stderr := brandedNative(t, program)
				t.Logf("ROW %s native: exit %d stdout %q stderr %q", probe.name, code, stdout, stderr)
			}
		})
	}
}
