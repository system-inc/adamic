package lower

import (
	"strings"
	"testing"
)

// The open-enum tag walk proves a property-access reference (`h.n`) only under Lower's local rule
// (#qeanycw, Adamic's ruling on #00rd54t, Nov 4): its narrowing sits in the test of a conditional,
// an `&&` or an if that holds the read, with no loop, label or function between, and nothing that
// can run program code or write a property completes between that narrowing and the read. A call
// leaves no flow node, so `reset(h)` between the test and the read could store a B in h.n while the
// checker still reads it as an A. kindPredicateTypes is the kind predicate pins' N = A | B.

const accessHolder = `interface H { readonly n: N }
interface M { n: N }
function isA(n: N): n is A { return n.kind === K.A; }
const a: A = { kind: K.A, flag: F.X, a: 1 };
const b: B = { kind: K.B, b: 'x' };
const ha: H = { n: a };
const hb: H = { n: b };
`

// Each local shape lowers and agrees with Node, sanitized natively and on the JavaScript backend,
// by an exact predicate's call and by the written comparison. Mutant: enumRefusal handing the walk
// the property's name again, so every row is refused as before #qeanycw.
func TestEnumTagAccessLocalShapes(test *testing.T) {
	test.Parallel()
	const twoHolders = "console.log(show(ha));\nconsole.log(show(hb));\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{
			"call in a conditional",
			"function show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }\n" + twoHolders,
			"a 1\nb\n",
		},
		{
			"comparison in a conditional",
			"function show(h: H): string { return h.n.kind === K.A ? `a ${h.n.a}` : 'b'; }\n" + twoHolders,
			"a 1\nb\n",
		},
		{
			"call in an if",
			"function show(h: H): string { if (isA(h.n)) { return `a ${h.n.a}`; } return 'b'; }\n" + twoHolders,
			"a 1\nb\n",
		},
		{
			"false side in an if",
			"function show(h: H): string { if (!isA(h.n)) { return `b ${h.n.b}`; } return 'a'; }\n" + twoHolders,
			"a\nb x\n",
		},
		{
			"&& chain",
			"function show(h: H): string { return isA(h.n) && h.n.a > 0 ? 'positive' : 'other'; }\n" + twoHolders,
			"positive\nother\n",
		},
		{
			"passed on",
			"function take(x: A): number { return x.a; }\nfunction show(h: H): string { return isA(h.n) ? `a ${take(h.n)}` : 'b'; }\n" +
				twoHolders,
			"a 1\nb\n",
		},
		{
			"mutable field, nothing between",
			"function show(m: M): string { return isA(m.n) ? `a ${m.n.a}` : 'b'; }\nconst ma: M = { n: a };\nconsole.log(show(ma));\n",
			"a 1\n",
		},
		{
			"object narrowed first",
			"interface O1 { readonly kind: K.A; readonly n: N }\ninterface O2 { readonly kind: K.B; readonly n: N; readonly z: number }\ntype O = O1 | O2;\nfunction isO1(o: O): o is O1 { return o.kind === K.A; }\nfunction show(o: O): string { return isO1(o) && isA(o.n) ? `a ${o.n.a}` : 'b'; }\nconst o1: O1 = { kind: K.A, n: a };\nconsole.log(show(o1));\n",
			"a 1\n",
		},
		{
			"two links",
			"interface G { readonly h: H }\nfunction show(g: G): string { return isA(g.h.n) ? `a ${g.h.n.a}` : 'b'; }\nconst g: G = { h: ha };\nconsole.log(show(g));\n",
			"a 1\n",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			source := kindPredicateTypes + accessHolder + probe.source
			for _, extension := range []string{".ts", ".a"} {
				if _, lowerSourceAsError := lowerSourceAs(test, source, extension); lowerSourceAsError != nil {
					test.Fatalf("%s: want it lowered, got %v", extension, lowerSourceAsError)
				}
			}
			program := lowersAndAgreesWithNodeAs(test, source, ".ts")
			if code, stdout, stderr := brandedNative(test, program); code != 0 || stdout != probe.stdout {
				test.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// What the local rule still refuses, each with the open-enum refusal. Mutants, each named in the
// commit: codeBetween always false (the resetting calls, the alias write and the call in the chain
// lower); proves counting a step outside the local tests (the loop row lowers, where reset runs
// after the read in source order and before it on the next pass); codeBetween counting from where
// the step ends (the call inside the step lowers, which runs after the step reads h.n); and the
// exact predicate's field allowed a getter (the getter of the compared field by predicate lowers).
// A predicate whose body writes the path (Lower's attack) is never exact, so it never proves.
func TestEnumTagAccessRefusals(test *testing.T) {
	test.Parallel()
	const reset = "function reset(m: M): void { m.n = b; }\n"
	for _, probe := range []struct{ name, source string }{
		{"mutable field, resetting call between", reset +
			"function show(m: M): string { if (isA(m.n)) { reset(m); return `a ${m.n.a}`; } return 'b'; }\nconsole.log(show({ n: a }));\n"},
		{
			"readonly field, resetting call between",
			"const raw: M = { n: a };\nfunction reset(): void { raw.n = b; }\nfunction show(h: H): string { if (isA(h.n)) { reset(); return `a ${h.n.a}`; } return 'b'; }\nconsole.log(show(raw));\n",
		},
		{
			"alias write between",
			"function show(h: H, m: M): string { if (isA(h.n)) { m.n = b; return `a ${h.n.a}`; } return 'b'; }\nconst both: M = { n: a };\nconsole.log(show(both, both));\n",
		},
		{
			"call in the chain",
			"function log(): boolean { return true; }\nfunction show(h: H): string { return isA(h.n) && log() ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(ha));\n",
		},
		{
			"early return",
			"function show(h: H): string { if (!isA(h.n)) { return 'b'; } return `a ${h.n.a}`; }\nconsole.log(show(ha));\n",
		},
		{"read in a loop under the test", reset +
			"function show(m: M): string { let out = ''; if (isA(m.n)) { let i = 0; while (i < 2) { if (i === 1) { out = `a ${m.n.a}`; } reset(m); i = i + 1; } } return out; }\nconsole.log(show({ n: a }));\n"},
		{
			"getter of that name in the program",
			"class Lazy implements H { get n(): N { return a; } }\nfunction show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(new Lazy()));\n",
		},
		{
			"writing predicate",
			"const holder: M = { n: a };\nfunction isAw(n: N): n is A { holder.n = b; return n.kind === K.A; }\nfunction show(h: H): string { return isAw(h.n) ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(holder));\n",
		},
		{
			"call inside the step",
			"function kindOf(): K.A { return K.A; }\nfunction show(h: H): string { return h.n.kind === kindOf() ? `a ${h.n.a}` : 'b'; }\nconsole.log(show(ha));\n",
		},
		{
			"getter of the compared field, by predicate",
			"class Odd { get kind(): K.A { return K.A; } }\nfunction show(h: H): string { return isA(h.n) ? `a ${h.n.a}` : 'b'; }\nconst odd = new Odd();\nconsole.log(show(ha) + String(odd.kind));\n",
		},
		{
			"getter of the compared field, by comparison",
			"class Odd { get kind(): K.A { return K.A; } }\nfunction show(h: H): string { return h.n.kind === K.A ? `a ${h.n.a}` : 'b'; }\nconst odd = new Odd();\nconsole.log(show(ha) + String(odd.kind));\n",
		},
		{
			"constructor's this",
			"class C { readonly n: N; readonly out: string; constructor(n: N) { this.n = n; this.out = isA(this.n) ? `a ${this.n.a}` : 'b'; } }\nconsole.log(new C(a).out);\n",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceAsError := lowerSourceAs(test, kindPredicateTypes+accessHolder+probe.source, ".ts")
			if lowerSourceAsError == nil || !strings.Contains(lowerSourceAsError.Error(), kindPredicateRefusal) {
				test.Fatalf("want %q, got %v", kindPredicateRefusal, lowerSourceAsError)
			}
		})
	}
}
