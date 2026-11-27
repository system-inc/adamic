package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// enumTagSwingTwoSites refines one declared union to one member at two sites, each with a call between the
// tag test and the read, so the walk can't prove either and the swing takes both. The member holds a callable
// field, which a checked view can't read: its view registers, and each read stops on its own.
const enumTagSwingTwoSites = `enum K { A = 1, B = 2 }
enum F { X = 1 }
interface A { readonly kind: K.A; readonly flag: F; readonly f: () => string }
interface B { readonly kind: K.B; readonly b: string }
type N = A | B;
interface M { n: N }
function touch(): void {}
function one(m: M): string { if (m.n.kind === K.A) { touch(); return m.n.f(); } return 'b'; }
function two(m: M): string { if (m.n.kind === K.A) { touch(); return m.n.f(); } return 'b'; }
const a: A = { kind: K.A, flag: F.X, f: () => 'a' };
console.log(one({ n: a }));
console.log(two({ n: a }));
`

// A first-link lowering goes on past a stop, and each site the swing takes stops in its own unit, on its own
// line. It holds placement only: a callable field's view registers and its stop is at the read, so it never
// reaches a view that fails at registration, and the mutant that keeps such a failure's error for the pair
// survives here (J-51b5pqs). A view failing at registration is #3bf1yxs's to pin (Compiler's should on
// #x2cjvy9): enumTagCheckedViews keeps only views that registered, so such a pair is asked again at each site.
func TestEnumTagSwingSitesStopInTheirOwnUnits(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", enumTagSwingTwoSites)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := LowerWithOptions(context.Background(), loaded, Options{FirstLinkTraps: true})
	if err != nil {
		t.Fatalf("want a first-link lowering, got %v", err)
	}
	var one, two bool
	for _, trap := range program.Traps {
		switch {
		case strings.HasSuffix(trap.Unit, " one") && strings.Contains(trap.Site, "main.ts:8:"):
			one = true
		case strings.HasSuffix(trap.Unit, " two") && strings.Contains(trap.Site, "main.ts:9:"):
			two = true
		}
	}
	if !one || !two {
		t.Fatalf("want a stop in one at line 8 and one in two at line 9, got %+v", program.Traps)
	}
}
