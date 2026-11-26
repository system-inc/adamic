package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

// enumTagSwingViewTypes is row 258's shape (rulings.md:258) for the swing's condition 2 (#x2cjvy9, Adamic's
// ruling of Nov 25): a direct tag test of a field, m.n.kind === K.A, over a union of members whose tags are an
// open numeric enum's, with a call between the test and the read that the walk can't see past. reset writes b,
// a B, where the read takes an A, so the refined value is a lie at the read; keep writes a, an A, so its twin is
// honest. Every cast that would build the lie directly is refused before the swing (J-3y6bcjn), so a call that
// rewrites the field is the road. A's flag holds the whole open enum F, as kindPredicateTypes' A does, so the
// walk raises a hazard and the swing's view is what reads: without it closedKinds closes N, the walk proves the
// refinement, and the swing never runs (J-zy5mseq read the lie wrong at the head and under the mutant alike, a
// road #8nj5pn4 holds). Each row checks its .a copy is refused with family 1's words, so a pin can't miss the
// swing's road again.
const enumTagSwingViewTypes = `enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly a: number }
interface B { readonly kind: K.B; readonly b: string }
type N = A | B;
interface M { n: N }
const a: A = { kind: K.A, flag: F.X, a: 1 };
const b: B = { kind: K.B, b: "x" };
function reset(m: M): void { m.n = b; }
function keep(m: M): void { m.n = a; }
function use(x: A): string { return String(x.a + 1); }
`

// The swing reads an unproven open-enum refinement through each narrowed member's checked view, so a lie stops
// loudly at its read, never reading B's storage through A's layout: read directly, handed to a function taking
// the member, stored in an array of the member and read back, and read through a wider structural view. Each
// honest twin agrees with Node on both backends, and each lie stops with a deliberate panic at exit 70 on sanitized
// native and the JavaScript backend, where Node prints what reading a missing field gives. Mutant: the view
// registered without narrows (enums.go's enumTagViewProbe), under which the array and wider-view lies read B's
// storage raw and the test goes red.
func TestEnumTagSwingLiesStopAtEveryRead(t *testing.T) {
	t.Parallel()
	for _, road := range []struct{ name, read string }{
		{"direct", "return String(m.n.a + 1);"},
		{"through a function", "return use(m.n);"},
		{"stored in an array", "const held: A[] = [m.n]; let out = ''; for (const x of held) { out += String(x.a + 1); } return out;"},
		{"through a wider view", "const wide: { readonly a: number } = m.n; return String(wide.a + 1);"},
	} {
		for _, side := range []struct{ name, call, want string }{
			{"honest", "keep", "AGREE"},
			{"lie", "reset", "STOP"},
		} {
			t.Run(road.name+" "+side.name, func(t *testing.T) {
				t.Parallel()
				source := enumTagSwingViewTypes + "function show(m: M): string { if (m.n.kind === K.A) { " + side.call + "(m); " + road.read + " } return 'b'; }\nconsole.log(show({ n: a }));\n"
				if _, err := lowerSourceAs(t, source, ".a"); err == nil || !strings.Contains(err.Error(), kindPredicateRefusal) {
					t.Fatalf("want the .a copy refused with %q, so the swing's road is what lowers the .ts copy, got %v", kindPredicateRefusal, err)
				}
				program, err := lowerSourceAs(t, source, ".ts")
				if err != nil {
					t.Fatalf("want it lowered, got %v", err)
				}
				node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source))
				nativeCode, nativeStdout, nativeStderr := brandedNative(t, program)
				generated := runAgreementNode(t, writeAgreementSource(t, "generated.mjs", javascript.JavaScript(program)))
				nativeClass := enumTagSwingClass(nativeCode, nativeStdout, nativeStderr, node)
				javascriptClass := enumTagSwingClass(generated.code, string(generated.stdout), string(generated.stderr), node)
				if nativeClass != side.want || javascriptClass != side.want {
					t.Fatalf("want %s on both backends, got native %s (exit %d stdout %q stderr %q) and JavaScript %s (exit %d stdout %q stderr %q); Node exit %d stdout %q", side.want, nativeClass, nativeCode, nativeStdout, nativeStderr, javascriptClass, generated.code, generated.stdout, generated.stderr, node.code, node.stdout)
				}
			})
		}
	}
}
