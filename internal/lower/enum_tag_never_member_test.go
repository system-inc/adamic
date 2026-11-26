package lower

import (
	"testing"
)

// An intersection of unions distributes into a union whose members include intersections that hold
// no value, since their kinds conflict (`P & C` below). The checker's narrowing drops them, so the
// walk of enum_tag_flow.go reads each type through the checker's reduction (enumTagReducedType):
// asked for kind, such a member has no field, and the step failed for every value. That was 57 of
// family 1's refusals on tsc (#x2cjvy9), 45 of them nodeFactory.ts's replaceDecoratorsAndModifiers,
// whose node is `HasModifiers & HasDecorators`. F is a one-member enum, so P's flag is an open tag
// and each refinement here needs the walk. Mutant: enumTagReducedType returning its type unreduced,
// which refuses every row but the control with the open-enum refusal.
const neverMemberTypes = `enum K { P = 1, C = 2, M = 3 }
enum F { X = 1 }
interface P { readonly kind: K.P; readonly flag: F; readonly p: number }
interface C { readonly kind: K.C; readonly c: string }
interface M { readonly kind: K.M; readonly m: string }
type HasX = P | C | M;
type HasY = P | C;
function isP(n: HasX): n is P { return n.kind === K.P; }
function isC(n: HasX): n is C { return n.kind === K.C; }
const p: P = { kind: K.P, flag: F.X, p: 1 };
const c: C = { kind: K.C, c: 'x' };
`

func TestEnumTagWalkDropsMembersThatHoldNoValue(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		// kindPredicateStep: the predicate chain replaceDecoratorsAndModifiers writes.
		{"predicate chain", "function show(n: HasX & HasY): string { return isP(n) ? `p ${n.p}` : isC(n) ? `c ${n.c}` : 'none'; }\nconsole.log(show(p));\nconsole.log(show(c));\n", "p 1\nc x\n"},
		// fieldRead: the written comparison on the same union.
		{"comparison", "function show(n: HasX & HasY): string { return n.kind === K.P ? `p ${n.p}` : n.kind === K.C ? `c ${n.c}` : 'none'; }\nconsole.log(show(p));\nconsole.log(show(c));\n", "p 1\nc x\n"},
		// plainDataLink: an access reference whose object holds the union.
		{"access reference", "interface H { readonly n: HasX & HasY }\nfunction show(h: H): string { return isP(h.n) ? `p ${h.n.p}` : isC(h.n) ? `c ${h.n.c}` : 'none'; }\nconsole.log(show({ n: p }));\nconsole.log(show({ n: c }));\n", "p 1\nc x\n"},
		// The control: a union with no empty member, which main lowered too.
		{"control", "function show(n: HasY): string { return isP(n) ? `p ${n.p}` : `c ${n.c}`; }\nconsole.log(show(p));\nconsole.log(show(c));\n", "p 1\nc x\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, neverMemberTypes+probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// The reduction drops only members that hold no value. A member that holds one keeps its hazard: Q's
// kind is the whole of K, and wide, a Q whose kind is chosen at run time, keeps it open (closedKinds,
// enum_tag_closed_kind.go), so it may hold K.P with Q's layout, and the refinement stays refused.
func TestEnumTagWalkKeepsMembersThatHoldAValue(t *testing.T) {
	t.Parallel()
	source := neverMemberTypes + "interface Q { readonly kind: K; readonly q: string }\nfunction numeric(): number { return 1; }\nconst wide: Q = { kind: numeric(), q: 'wide' };\nfunction isPQ(n: P | Q): n is P { return n.kind === K.P; }\nfunction show(n: (P | Q) & (P | Q | C)): string { return isPQ(n) ? `p ${n.p}` : 'other'; }\nconsole.log(show(p));\n"
	// .a stops earlier at the predicate under rulings.md:142, so the walk is judged in .ts, where a refinement it
	// can't prove takes the swing's checked view (enumTagSwingSites).
	if _, sites, err := lowerSourceSwung(t, source); err != nil || len(sites) == 0 {
		t.Fatalf("want the walk to keep Q and the swing to take the refinement, got sites %v and %v", sites, err)
	}
}
