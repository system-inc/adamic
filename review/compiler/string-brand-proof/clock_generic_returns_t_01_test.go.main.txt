package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestClockGenericReturnsT01Shapes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, shape string
		supported   bool
	}{
		{"nested readonly fields", `{ readonly left: { readonly text: string; readonly generated: boolean }; readonly children: readonly string[] }`, true},
		{"callable", `{ (): string; readonly left: { readonly text: string } }`, false},
		{"constructable", `{ new (): { text: string }; readonly left: { readonly text: string } }`, false},
		{"indexed", `{ readonly [key: string]: unknown; readonly left: { readonly text: string } }`, false},
		{"container intersection", `readonly string[]`, false},
		{"nested callable", `{ readonly field: () => string }`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := `interface Base<T> { readonly token: T; } function make(): (Base<"="> & ` + probe.shape + `) | undefined { return undefined; } console.log(typeof make());`
			_, err := lowerSource(t, source)
			if probe.supported {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var stop *NotYet
			var refusal *Refused
			if !errors.As(err, &stop) && !errors.As(err, &refusal) {
				t.Fatalf("want unsupported shape stop, got %v", err)
			}
		})
	}
	t.Run("null remains distinct", func(t *testing.T) {
		t.Parallel()
		_, err := lowerSource(t, `interface Base<T> { readonly token: T; }
function make(): (Base<"="> & { readonly left: { readonly text: string } }) | undefined | null { return null; }
console.log(typeof make());`)
		var stop *NotYet
		if !errors.As(err, &stop) {
			t.Fatalf("want a distinct null/undefined representation stop, got %v", err)
		}
	})

}

// A branded primitive is an intersection whose parts aren't all type references. Asking the
// string part for type arguments made the checker dereference nil (x1 of views slice 1,
// notyet_element_access/paths.a); the proof now declines it and lowering stops normally.
func TestClockGenericReturnsT01DeclinesBrandedPrimitives(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`type Path = string & { __pathBrand: void }; function read(values: readonly Path[], index: number): Path | undefined { return values[index]; } console.log(typeof read([], 0));`,
		`type Count = number & { __countBrand: void }; function read(values: readonly Count[], index: number): Count | undefined { return values[index]; } console.log(typeof read([], 0));`,
	} {
		_, err := lowerSource(t, source)
		var stop *NotYet
		if !errors.As(err, &stop) {
			t.Fatalf("want an explicit NotYet for a branded primitive result, got %v", err)
		}
	}
}

// Reject null at the signature proof itself. A later body or call refusal must
// not hide a mutant which merges null with the missing object representation.
func TestClockGenericReturnsT01RejectsNullBeforeBody(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "probe.a")
	source := `interface Base<T> { readonly token: T; }
function make(): (Base<"="> & { readonly left: { readonly text: string } }) | undefined | null { return null; }
console.log(typeof make());`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	requireClockSignaturePositiveControl(t)
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Text() == "make" {
			found = true
			result := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(node))
			if held, known := l.clockGenericReturnsT01(result); known {
				t.Errorf("null and undefined admitted with one representation: %v", held)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if !found {
		t.Fatal("missing checked make signature")
	}
}

// An index signature must be rejected by this proof even when the body returns
// undefined and later passes independently refuse the erased index shape.
func TestClockGenericReturnsT01RejectsIndexBeforeBody(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "probe.a")
	source := `interface Base<T> { readonly token: T; }
function make(): (Base<"="> & { readonly [key: string]: unknown; readonly left: { readonly text: string } }) | undefined { return undefined; }
console.log(typeof make());`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	requireClockSignaturePositiveControl(t)
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Text() == "make" {
			found = true
			result := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(node))
			if held, known := l.clockGenericReturnsT01(result); known {
				t.Errorf("indexed object admitted by the finite-shape proof: %v", held)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if !found {
		t.Fatal("missing checked make signature")
	}
}

// The rejecting proof must still admit its supported finite object signature.
func requireClockSignaturePositiveControl(t *testing.T) {
	t.Helper()
	graph, statements := namespaceGraphForTest(t, `interface Base<T> { readonly token: T; }
 function make(): (Base<"="> & { readonly left: { readonly text: string } }) | undefined { return undefined; }`)
	for _, node := range statements {
		if node.Kind == ast.KindFunctionDeclaration {
			checked := graph.lowering.checker
			result := checked.GetReturnTypeOfSignature(checked.GetSignatureFromDeclaration(node))
			if held, known := graph.lowering.clockGenericReturnsT01(result); !known || held != ir.Object {
				t.Fatalf("supported signature lost its object representation: held=%v known=%t", held, known)
			}
			return
		}
	}
	t.Fatal("missing supported make signature")
}
