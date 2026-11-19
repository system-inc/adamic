package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// closedKindConstructorTypes is tsc's factory, small: an unbranded supertype Open, a literal of it at a
// literal kind, and a function constructor whose this is declared, writing a kind chosen at run time (kb), as
// utilities.ts's Node(this: Mutable<Node>, kind, pos, end) does. It takes no other parameter, since lowering
// refuses a function whose this parameter is followed by another (dynamic receivers), so the narrowing rows
// can lower beside it.
const closedKindConstructorTypes = closedKindTypes +
	"interface Open extends Node { readonly n: number }\n" +
	"const open: Open = { kind: K.Ident, pos: 0, n: 1 };\n" +
	"function kb(): K { return K.Ident; }\n" +
	"function OpenNode(this: { kind: K; pos: number; n: number }): void { this.kind = kb(); this.pos = 0; this.n = 1; }\n"

// A function constructor's new is born as a checked view, so its writes through a declared this, and a
// new whose callee isn't a class, add nothing to K(S), when every use of the function is a new callee or
// the operand of a cast to any (enum_tag_closed_kind.go, Adamic's ruling on #8eewz2n). Any other use keeps
// the write a build: a call, .call, storage, a getter returning it with no cast (Adamic's first condition:
// the road tsc's objectAllocator takes once its `as any` is adapted), or an argument. A class installed
// through a construct signature, as setObjectAllocator installs services' NodeObject, keeps S open by its
// own declaration when its kind is open (Adamic's third condition).
// Mutants: m1, the this write never a hatch, opens the unused and new-through-any rows; m2, every new a
// class's, opens the construct-signature row; m3, no use judged, closes the call, getter, argument, bind
// and shorthand rows.
func TestClosedKindFunctionConstructorIsAHatch(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		closed       bool
		reason       string
	}{
		{"unused", "", true, ""},
		{"new through any", "const made: Open = new (OpenNode as any)();\n", true, ""},
		{"construct signature", "declare const Make: new (kind: K) => Open;\nconst made = new Make(kb());\n", true, ""},
		{"call", "OpenNode.call(open);\n", false, "a write of kind"},
		{"bind", "const bound = OpenNode.bind(open);\n", false, "a write of kind"},
		{"getter, no cast", "const allocator = { getOpen: () => OpenNode };\n", false, "a write of kind"},
		{"argument", "function keep(f: unknown): void {}\nkeep(OpenNode);\n", false, "a write of kind"},
		{"shorthand storage", "const box = { OpenNode };\n", false, "a write of kind"},
		{"class through construct signature", "class OpenObject { kind: K; pos = 0; n = 1; constructor(kind: K) { this.kind = kind; } }\nconst Make: new (kind: K) => Open = OpenObject;\n", false, ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, _, _, closed, reason := closedKindsAt(t, closedKindConstructorTypes+probe.source, "Open")
			if closed != probe.closed || probe.reason != "" && !strings.Contains(reason, probe.reason) {
				t.Fatalf("want Open closed %v, reason holding %q; got closed %v, reason %q", probe.closed, probe.reason, closed, reason)
			}
		})
	}
}

// The walk reads the hatch. No program holding a function constructor lowers today: lowering stops at its
// this ("this outside a method", J-np4frks), and at a this parameter followed by another. So the hatch shows
// in which stop a program reaches, and the narrowing's refusal comes first when the walk refuses: beside an
// unused constructor the walk proves isOpenBody's narrowing over Body | Open, and the program stops at the
// constructor's this; beside a constructor called as a plain function, which could write a run-time kind
// onto an object built at a literal one, the walk keeps the open-enum refusal.
// Mutants: m1 turns the first row into the refusal; m3 the second into the this stop.
func TestClosedKindFunctionConstructorNarrowing(t *testing.T) {
	t.Parallel()
	const narrowing = "function isOpenBody(x: Body | Open): x is Body { return x.kind === K.Body; }\n" +
		"function show(x: Body | Open): string { return isOpenBody(x) ? x.b : 'open'; }\n" +
		"console.log(show(body));\nconsole.log(show(open));\n"
	for _, probe := range []struct{ name, source, stop string }{
		{"unused", closedKindConstructorTypes + narrowing, "can't lower this outside a method"},
		{"called", closedKindConstructorTypes + "OpenNode.call(open);\n" + narrowing, kindPredicateRefusal},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, ".ts"); err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want a stop holding %q, got %v", probe.stop, err)
			}
		})
	}
}

// The fence the hatch leans on (Adamic's first condition on #8eewz2n): a new of a function constructor
// lowers by no road but an allocator getter's, whose object is born as a checked view (#837zjkt,
// function_constructor_test.go). Every road is a cast, since the checker gives a function declaration no
// construct signature, and lowering refuses each cast outside that road before the new (J-dtkcpw6: through
// unknown, through any, and beside a class whose construct signature it matches). Should one lower as a
// plain object first, this goes red, since K(S) counts such a build as a hatch.
func TestClosedKindFunctionConstructorNewStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"through unknown", "const Ctor = OpenNode as unknown as new () => Open;\nconst made = new Ctor();\nconsole.log(String(made.n));\n"},
		{"through any", "const made: Open = new (OpenNode as any)();\nconsole.log(String(made.n));\n"},
		{"beside a class", "class Real { readonly pos = 0; readonly n = 2; constructor(readonly kind: K) {} }\nconsole.log(String(new Real(K.Ident).n));\nconst Ctor = OpenNode as unknown as new () => Open;\nconst made = new Ctor();\nconsole.log(String(made.n));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, closedKindConstructorTypes+probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "refuses a cast the runtime can't check") {
				t.Fatalf("a new of a function constructor must stop at the cast that reaches it, got %v: K(S) counts it a hatch only while it doesn't lower, or once it's born as a checked view", err)
			}
		})
	}
}

// A use under another name is judged where it's used (Language's first should on #8eewz2n): a function
// constructor exported from lib.ts and called under a renamed import, or a default import, keeps Open open,
// and newed through a cast to any under the renamed import, it doesn't. Mutant: the use scan filtering
// identifiers by the function's own name, as 8d8e651c did, closes the renamed and default rows.
func TestClosedKindFunctionConstructorUsesUnderAnotherName(t *testing.T) {
	t.Parallel()
	const constructor = "(this: { kind: number; pos: number; n: number }): void { this.kind = Number('1'); this.pos = 0; this.n = 1; }\n"
	for _, probe := range []struct {
		name, lib, use string
		closed         bool
	}{
		{"renamed import, called", "export function Make" + constructor, "import { Make as M } from './lib';\nM.call(open);\n", false},
		{"default import, called", "export default function Make" + constructor, "import Anything from './lib';\nAnything.call(open);\n", false},
		{"renamed import, newed through any", "export function Make" + constructor, "import { Make as M } from './lib';\nconst made: Open = new (M as any)();\n", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			main := writeAgreementSource(t, "main.ts", closedKindConstructorTypes+probe.use)
			if err := os.WriteFile(filepath.Join(filepath.Dir(main), "lib.ts"), []byte(probe.lib), 0o644); err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{main})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			checked, release := loaded.Checker(context.Background(), loaded.Files()[0])
			defer release()
			l := &lowering{program: loaded, checker: checked, result: &ir.Program{}}
			var open *ast.Node
			for _, file := range loaded.Files() {
				for _, statement := range file.Statements.Nodes {
					if statement.Kind == ast.KindInterfaceDeclaration && statement.Name().Text() == "Open" {
						open = statement
					}
				}
			}
			if open == nil {
				t.Fatal("no interface Open")
			}
			_, closed, reason := l.closedKinds(l.checker.GetDeclaredTypeOfSymbol(l.checker.GetSymbolOfDeclaration(open)), "kind")
			if closed != probe.closed {
				t.Fatalf("want Open closed %v, got closed %v, reason %q", probe.closed, closed, reason)
			}
		})
	}
}
