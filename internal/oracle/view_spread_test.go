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
// view (view_spread_predicate.ts). A copy through a wider type that names a checked field, of a value that may be a
// viewed allocation, is NotYet too (view_spread_supertype.ts). The fixtures sit in reland_refused, since they don't
// lower and the flow and fresh corpora lower every program in testdata. Mutant: checkViewCopies returning nil at once;
// with it applied, this test goes red, and ADAMIC_VIEW_SPREAD_MUTANT=1 holds the mutant's admitted programs to Node on
// sanitized and release native to show the crash.
func TestViewSpreadNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{"view_spread.ts", "3:84", "a whole-object copy of a value of checked view type Lit"},
		{"view_spread_parameter.ts", "3:34", "a whole-object copy of a value of checked view type Lit"},
		{"view_spread_predicate.ts", "4:60", "a whole-object copy of a value of checked view type Lit"},
		{
			"view_spread_supertype.ts",
			"4:106",
			"a whole-object copy of a value of type HasText that may hold a checked view's value",
		},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				test.Fatal(err)
			}
			truth := onNode(test, path)
			if disagreement(run{stdout: []byte("5\n")}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(test, path)
			if os.Getenv("ADAMIC_VIEW_SPREAD_MUTANT") == "1" {
				if err != nil {
					test.Fatalf("the mutant must admit the program: %v", err)
				}
				native, _ := nativelyUncached(test, program)
				for name, got := range map[string]run{"native": native, "release": releasedUncached(test, program)} {
					if disagreement(truth, got) == "" {
						test.Fatalf("the mutant's %s program agrees with Node: %+v", name, got)
					}
					test.Logf(
						"mutant admits it: %s exit %d, stdout %q, stderr %q where Node prints %q",
						name,
						got.exitCode,
						got.stdout,
						got.stderr,
						truth.stdout,
					)
				}
				return
			}
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) ||
				!strings.Contains(err.Error(), probe.file+":"+probe.at) ||
				!strings.Contains(err.Error(), probe.what) {
				test.Fatalf("want the whole-object copy NotYet at the spread, got %v", err)
			}
		})
	}
}

// The other whole-object copies of a viewed value stop before they lower: Object.assign takes only a source whose
// shape a literal proves, and destructuring takes no rest binding. They stay refused at those rules.
func TestViewCopyVariantsRefused(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ file, at, what string }{
		{
			"view_spread_assign.ts",
			"3:122",
			"each source must have its complete shape proven by a literal or const binding",
		},
		{"view_spread_rest.ts", "3:86", "a destructured name that isn't plain"},
	} {
		test.Run(probe.file, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused", probe.file))
			if err != nil {
				test.Fatal(err)
			}
			if truth := onNode(test, path); disagreement(run{stdout: []byte("5\n")}, truth) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			_, err = lowered(test, path)
			if err == nil ||
				!strings.Contains(err.Error(), probe.file+":"+probe.at) ||
				!strings.Contains(err.Error(), probe.what) {
				test.Fatalf("want the copy refused at %s, got %v", probe.at, err)
			}
		})
	}
}

// A spread of a type no view registers is an honest copy, beside a view of another type, and still lowers and agrees
// with Node on every backend.
func TestHonestSpreadAgrees(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.ts")
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
		test.Fatal(err)
	}
	truth := onNode(test, path)
	if disagreement(run{stdout: []byte("2 a 1\nt\n")}, truth) != "" {
		test.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatalf("an honest spread must lower: %v", err)
	}
	native, _ := natively(test, program)
	for name, got := range map[string]run{
		"native":     native,
		"release":    released(test, program),
		"javascript": onJavaScriptBackend(test, program),
	} {
		if problem := disagreement(truth, got); problem != "" {
			test.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, problem, got.exitCode, got.stdout, got.stderr)
		}
	}
}
