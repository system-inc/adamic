package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A checked view's fields are checked at a read under the view's receiver type, and a spread makes an object whose
// fields are no view's, so { ...l } of a value cast to Lit copied a lying layout raw: on main, Node printed 5 for
// view_spread.ts and sanitized native died in AddressSanitizer reading c.text. A whole-object copy of a value of a
// view's receiver type is now NotYet at the copy (internal/lower/view_copies.go), decided once lowering is done, so a
// copy in a function lowered before the cast (view_spread_parameter.ts) is caught too, as is a checked predicate's
// view (view_spread_predicate.ts). The fixtures sit in reland_refused, since they don't lower and the flow and fresh
// corpora lower every program in testdata. Mutant: checkViewCopies returning nil at once; with it applied, this test
// goes red, and ADAMIC_VIEW_SPREAD_MUTANT=1 holds the mutant's admitted programs to Node on sanitized and release
// native to show the crash.
func TestViewSpreadNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, at, view string }{
		{"view_spread.ts", "3:84", "Lit"},
		{"view_spread_parameter.ts", "3:34", "Lit"},
		{"view_spread_predicate.ts", "4:60", "Lit"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if disagreement(run{stdout: []byte("5\n")}, truth) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if os.Getenv("ADAMIC_VIEW_SPREAD_MUTANT") == "1" {
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				native, _ := nativelyUncached(t, program)
				for name, got := range map[string]run{"native": native, "release": releasedUncached(t, program)} {
					if disagreement(truth, got) == "" {
						t.Fatalf("the mutant's %s program agrees with Node: %+v", name, got)
					}
					t.Logf("mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q", name, got.exitCode, got.stdout, got.stderr, truth.stdout)
				}
				return
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.file+":"+probe.at) || !strings.Contains(err.Error(), "a whole-object copy of a value of checked view type "+probe.view) {
				t.Fatalf("want the whole-object copy NotYet at the spread, got %v", err)
			}
		})
	}
}

// A spread of a type no view registers is an honest copy, beside a view of another type, and still lowers and agrees
// with Node on every backend.
func TestHonestSpreadAgrees(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := `interface Node { readonly kind: number; readonly pos: number; }
interface Lit extends Node { readonly kind: 9; readonly text: string; }
interface Point { readonly x: number; readonly label: string; }
function show(n: Node): void { if (n.kind === 9) { const l = n as Lit; console.log(l.text); } }
const p: Point = { x: 1, label: "a" };
const c = { ...p, x: 2 };
console.log(` + "`${c.x} ${c.label} ${p.x}`" + `);
const v = { kind: 9, pos: 1, text: "t" };
show(v);
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if disagreement(run{stdout: []byte("2 a 1\nt\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("an honest spread must lower: %v", err)
	}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if problem := disagreement(truth, got); problem != "" {
			t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
		}
	}
}
