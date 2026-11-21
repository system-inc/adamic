package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func viewDiagnosticLocation(t *testing.T, name, refusal, location string) {
	t.Helper()
	path := filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a")
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), refusal) || !strings.Contains(err.Error(), "oct9_views_"+name+".a:"+location+":") {
		t.Fatalf("want %q with source location %s, got %v", refusal, location, err)
	}
}

func TestViewDiagnosticP17(t *testing.T) {
	t.Parallel()
	// Element access is refused at its key before any contract is consulted.
	viewDiagnosticLocation(t, "p17", "view member read by element access is not yet checked", "6:16")
}

// viewDiagnosticTwinLocation lowers the fixture's source as a .ts file of the same name, where its cast
// stays a checked view, so the read it was written about is reached as .a reached it before #m8dvvmf.
func viewDiagnosticTwinLocation(t *testing.T, name, refusal, location string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "view_diagnostics", "oct9_views_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{writeAgreementSource(t, "oct9_views_"+name+".ts", string(source))})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), refusal) || !strings.Contains(err.Error(), "oct9_views_"+name+".ts:"+location+":") {
		t.Fatalf("twin: want %q with source location %s, got %v", refusal, location, err)
	}
}

// In .a p48's cast is a downcast whose checks no proof erases, refused (#m8dvvmf) before the
// destructured read; the .ts twin pins the read's own diagnostic.
func TestViewDiagnosticP48(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p48", "a downcast from Root to View in .a whose checks no proof erases", "5:14")
	viewDiagnosticTwinLocation(t, "p48", "unsupported collection contract", "6:9")
}

func TestViewDiagnosticP64(t *testing.T) {
	t.Parallel()
	viewDiagnosticLocation(t, "p64", "view member read by element access is not yet checked", "7:16")
}

func TestViewDiagnosticCheckedCallableField(t *testing.T) {
	t.Parallel()
	const q06 = `// q06_method_through_view.ts
enum K { A = 1, D = 4 }
interface Named { describe(): string }
interface A { readonly kind: K.A; describe(): string }
interface D { readonly kind: K.D; readonly d: number }
class Shown implements D { readonly kind = K.D as const; readonly d = 2; describe(): string { return "shown"; } }
function show(x: A | D): string { return (x as Named).describe(); }
const s: D = new Shown();
console.log(show(s));
`
	const compute = `// Function-typed field through the same union-to-interface view.
enum K { A = 1, D = 4 }
interface Named { compute: (value: number) => number }
interface A { readonly kind: K.A; compute: (value: number) => number }
interface D { readonly kind: K.D; readonly d: number }
class Shown implements D { readonly kind = K.D as const; readonly d = 2; compute = (value: number): number => value + 1; }
function show(x: A | D): number { return (x as Named).compute(2); }
const s: D = new Shown();
console.log(String(show(s)));
`
	for _, probe := range []struct {
		name, source, want string
		wantRefused        bool
	}{
		{"q06.ts", q06, "q06.ts:7:42: stage 0 can't lower a callable field describe read through a checked view, which needs a check at the call (rulings.md:27) yet", false},
		{"compute.ts", compute, "compute.ts:7:42: stage 0 can't lower a callable field compute read through a checked view, which needs a check at the call (rulings.md:27) yet", false},
		{"q06.a", q06, "q06.a:7:43: Adamic 0.1 refuses a downcast from A | D to Named in .a whose checks no proof erases; give Named a tag of its own in a discriminated union and narrow on it, as .a writes !, or build the value as a Named so the compiler can prove it, or move the file to .ts, where the cast stays a checked cast", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, probe.name, probe.source)
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			var refused *Refused
			var notYet *NotYet
			rightKind := errors.As(err, &notYet)
			if probe.wantRefused {
				rightKind = errors.As(err, &refused)
			}
			if !rightKind || err.Error() != filepath.Dir(path)+string(filepath.Separator)+probe.want {
				t.Fatalf("want %q, got %v", probe.want, err)
			}
		})
	}
}

// Source declarations of any stop before lowering. Exercise the checked read
// directly so this pin covers the any family in the same diagnostic branch.
func TestViewDiagnosticAnyField(t *testing.T) {
	t.Parallel()
	receiver := ir.Read{Of: ir.Object}
	read := ir.Property{
		Object: receiver, Name: "value", Of: ir.Object,
		ViewTypeID: 1, ViewReceiverTypeID: 2, ViewWhere: "any.ts:1:1",
	}
	program := &ir.Program{
		Locals:            []ir.Local{{Type: ir.Object}},
		ViewOrigins:       []ir.Expression{receiver},
		ViewContracts:     []ir.ViewContract{{Kind: ir.ViewUnknown, Name: "any", Unsupported: "any"}},
		ViewContractTypes: map[int]ir.ViewContractID{1: 1},
		CheckedFields:     map[string]bool{ir.CheckedFieldKey(2, "value"): true},
		Main:              []ir.Statement{ir.Evaluate{Value: read}},
	}
	l := &lowering{result: program}
	err := l.judgeCheckedViewReads(nil)
	const want = "any.ts:1:1: Adamic 0.1 refuses checked view read of field value with unsupported any contract; prove or implement the any contract before reading this field"
	var refused *Refused
	if !errors.As(err, &refused) || err.Error() != want {
		t.Fatalf("want %q, got %v", want, err)
	}
}

func TestViewDiagnosticScalarTwin(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, `enum K { A = 1, D = 4 }
interface Named { readonly name: string }
interface A { readonly kind: K.A; readonly name: string }
interface D { readonly kind: K.D; readonly d: number }
class Shown implements D { readonly kind = K.D as const; readonly d = 2; readonly name = "shown"; }
function show(x: A | D): string { return (x as Named).name; }
const s: D = new Shown();
console.log(show(s));
`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if len(program.ViewOrigins) == 0 {
		t.Fatal("expected the scalar twin to lower through a checked view")
	}
}
