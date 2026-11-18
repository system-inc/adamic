package lower

import (
	"strings"
	"testing"
)

// closedKindConstructorTypes is tsc's factory, small: an unbranded supertype Open, a literal of it at a
// literal kind, and a function constructor whose this is declared, writing a kind chosen at run time, as
// utilities.ts's Node(this: Mutable<Node>, kind, pos, end) does. kb chooses the kind at run time.
const closedKindConstructorTypes = closedKindTypes +
	"interface Open extends Node { readonly n: number }\n" +
	"const open: Open = { kind: K.Ident, pos: 0, n: 1 };\n" +
	"function kb(): K { return K.Ident; }\n" +
	"function OpenNode(this: { kind: K; pos: number; n: number }, kind: K): void { this.kind = kind; this.pos = 0; this.n = 1; }\n"

// A function constructor's new is born as a checked view, so its writes through a declared this, and a
// new whose callee isn't a class, add nothing to K(S), when every use of the function is a new callee or
// the operand of a cast to any (enum_tag_closed_kind.go, Adamic's ruling on #8eewz2n). Any other use keeps
// the write a build: a call, .call, storage, a getter returning it with no cast (Adamic's first condition:
// the road tsc's objectAllocator takes once its `as any` is adapted), or an argument. A class installed
// through a construct signature, as setObjectAllocator installs services' NodeObject, keeps S open by its
// own declaration when its kind is open (Adamic's third condition).
// Mutants: m1, the this write never a hatch, opens the unused and new-through-any rows; m2, every new a
// class's, opens the construct-signature row; m3, no use judged, closes the call, getter, argument and
// bind rows.
func TestClosedKindFunctionConstructorIsAHatch(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		closed       bool
		reason       string
	}{
		{"unused", "", true, ""},
		{"new through any", "const made: Open = new (OpenNode as any)(kb());\n", true, ""},
		{"construct signature", "declare const Make: new (kind: K) => Open;\nconst made = new Make(kb());\n", true, ""},
		{"call", "OpenNode.call(open, kb());\n", false, "a write of kind"},
		{"bind", "const bound = OpenNode.bind(open);\n", false, "a write of kind"},
		{"getter, no cast", "const allocator = { getOpen: () => OpenNode };\n", false, "a write of kind"},
		{"argument", "function keep(f: unknown): void {}\nkeep(OpenNode);\n", false, "a write of kind"},
		{"class through construct signature", "class OpenObject { kind: K; pos = 0; n = 1; constructor(kind: K) { this.kind = kind; } }\nconst Make: new (kind: K) => Open = OpenObject;\n", false, "OpenObject"},
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

// The walk reads the hatch: isOpenBody's narrowing over Body | Open lowers and agrees with Node beside an
// unused function constructor, and keeps the open-enum refusal when the constructor is called as a plain
// function, which could write a run-time kind onto an object built at a literal one.
// Mutant: m1 refuses the first row; m3 lowers the second.
func TestClosedKindFunctionConstructorNarrowing(t *testing.T) {
	t.Parallel()
	const narrowing = "function isOpenBody(x: Body | Open): x is Body { return x.kind === K.Body; }\n" +
		"function show(x: Body | Open): string { return isOpenBody(x) ? x.b : 'open'; }\n" +
		"console.log(show(body));\nconsole.log(show(open));\n"
	t.Run("unused", func(t *testing.T) {
		t.Parallel()
		program := lowersAndAgreesWithNodeAs(t, closedKindConstructorTypes+narrowing, ".ts")
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "block\nopen\n" {
			t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "block\nopen\n")
		}
	})
	t.Run("called", func(t *testing.T) {
		t.Parallel()
		source := closedKindConstructorTypes + "OpenNode.call(open, kb());\n" + narrowing
		if _, err := lowerSourceAs(t, source, ".ts"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
			t.Fatalf("want %q, got %v", kindPredicateRefusal, err)
		}
	})
}

// The fence the hatch leans on (Adamic's first condition on #8eewz2n): a new of a function constructor
// doesn't lower, by any road a program can write, until its object is born as a checked view. Should one
// lower as a plain object first, this goes red, since K(S) counts such a build as a hatch.
func TestClosedKindFunctionConstructorNewStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"through unknown", "const Ctor = OpenNode as unknown as new (kind: K) => Open;\nconst made = new Ctor(kb());\nconsole.log(String(made.n));\n"},
		{"through any", "const made: Open = new (OpenNode as any)(kb());\nconsole.log(String(made.n));\n"},
		{"beside a class", "class Real { readonly pos = 0; readonly n = 2; constructor(readonly kind: K) {} }\nconsole.log(String(new Real(K.Ident).n));\nconst Ctor = OpenNode as unknown as new (kind: K) => Open;\nconst made = new Ctor(kb());\nconsole.log(String(made.n));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, closedKindConstructorTypes+probe.source, ".ts")
			t.Logf("FENCE %s: %v", probe.name, err)
			if err == nil {
				t.Fatal("a new of a function constructor lowered: K(S) counts it a hatch only while it doesn't, or once it's born as a checked view")
			}
		})
	}
}
