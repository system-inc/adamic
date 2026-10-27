package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A view field whose interface has a length is held as a union, since a string, an array and an object all satisfy
// it, and its contract is an object's, so neither the union adapter nor the in-place check can emit the read. Lowering
// stops at the read, on a cast-reached object and on one the cast never sees, where emission used to panic with an
// incomplete checked field contract and the JavaScript backend aborted where Node prints the length. The probes are
// .ts, the reviewed program's text among them, where the cast stays a checked view; the reviewed program itself is .a,
// where the cast is refused first, since its tag test, kind === 'Holder', is one every Root passes (#q97a1rr).
func TestViewFieldHeldAsUnionStopsAtTheRead(t *testing.T) {
	t.Parallel()
	reviewed, err := filepath.Abs(filepath.Join(reviewRoot, "refused", "fxspptb_oct9_views_p70_length_view_over_tuple.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, reviewed)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.Where != reviewed+":8:16" || refused.What != `a downcast to View whose tag test is vacuous: every Root has kind "Holder"` {
		t.Fatalf("want the cast's vacuous downcast refusal at %s:8:16, got %v", reviewed, err)
	}
	reviewedSource, err := os.ReadFile(reviewed)
	if err != nil {
		t.Fatal(err)
	}
	twin := filepath.Join(t.TempDir(), "reviewed.ts")
	if err := os.WriteFile(twin, reviewedSource, 0o644); err != nil {
		t.Fatal(err)
	}
	direct := filepath.Join(t.TempDir(), "direct.ts")
	source := `interface Sized { readonly length: number; }
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: Sized; }
const pair: [string, number] = ['first', 1];
const raw = { kind: 'Holder' as const, value: pair };
const base: Root = raw;
const viewed = base as View;
const direct: View = { kind: 'Holder', value: 'abc' };
const value = direct.value;
console.log(` + "`${value.length}`" + `);
console.log(` + "`${viewed.kind}`" + `);
`
	if err := os.WriteFile(direct, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ path, stdout, where string }{
		{twin, "2\n", twin + ":9:15"},
		{direct, "3\nHolder\n", direct + ":9:15"},
	} {
		node := onNode(t, probe.path)
		if node.exitCode != 0 || string(node.stdout) != probe.stdout || len(node.stderr) != 0 {
			t.Fatalf("Node on %s: exit=%d stdout=%q stderr=%q", probe.path, node.exitCode, node.stdout, node.stderr)
		}
		_, err := lowered(t, probe.path)
		var pending *lower.NotYet
		if !errors.As(err, &pending) || pending.Where != probe.where || pending.What != "a checked view read of field value of type Sized, which is a union of differently held members outside a union view contract" {
			t.Fatalf("want the read's NotYet at %s, got %v", probe.where, err)
		}
	}
}
