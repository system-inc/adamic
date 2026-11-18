package lower

import (
	"errors"
	"strings"
	"testing"
)

const neverReadCallWrittenWhat = "a read the checker narrowed to never of a binding a call may write"

// An exhaustive switch reaches the never read with one discriminant read at the declared type.
// Keep the never diagnostic pinned for each of the original consumers.
func TestNeverReadOfACallWrittenLetIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, body string }{
		{"nv2", "const seen: unknown = node; return stray(seen);"},
		{"nv6", "const count: number = node; return \"count \" + String(count + 1);"},
		{"nv8", "const other: { readonly label: string } = node; return \"label \" + String(other.label === undefined);"},
		{"nv9", "const other: { readonly kind: string } = node; return other.kind.toUpperCase();"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := neverReadLetSwitch + "\t\t" + probe.body + neverReadLetSwitchEnd
			_, err := lowerSourceAs(t, source, ".ts")
			var notYet *NotYet
			want := neverReadWhere(t, source, "node; return")
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) || notYet.What != neverReadCallWrittenWhat {
				t.Fatalf("got %v, want a not yet at %s saying %q", err, want, neverReadCallWrittenWhat)
			}
		})
	}
}

// The exhaustive control, nv4: a parameter only its own function writes is narrowed to never by checks
// with no call between them, so its never is the checker's proof and it lowers as before. Mutant 2 in
// narrowed_never.mutants.md (the writtenAcrossFunctions gate dropped) stops it.
func TestNeverReadOfAnUnwrittenParameterLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
function describe(node: N): string {
	if (node.kind === K.A) return 'a';
	if (node.kind === K.B) return 'b';
	{ const seen: unknown = node; return stray(seen); }
}
console.log(describe({ kind: K.B, b: 2 }));
console.log(describe({ kind: K.A, a: 1 }));
`, ".ts"); err != nil {
		t.Fatalf("want nv4 to lower, got %v", err)
	}
}
