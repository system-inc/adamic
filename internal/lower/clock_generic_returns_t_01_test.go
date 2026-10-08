package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestClockGenericReturnsT01Shapes(t *testing.T) {
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
		_, err := lowerSource(t, `interface Base<T> { readonly token: T; }
function make(): (Base<"="> & { readonly left: { readonly text: string } }) | undefined | null { return null; }
console.log(typeof make());`)
		var stop *NotYet
		if !errors.As(err, &stop) {
			t.Fatalf("want a distinct null/undefined representation stop, got %v", err)
		}
	})

}

// Reject null at the signature proof itself. A later body or call refusal must
// not hide a mutant which merges null with the missing object representation.
func TestClockGenericReturnsT01RejectsNullBeforeBody(t *testing.T) {
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
