package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestViewArrayContract(t *testing.T) {
	for _, test := range []struct {
		declaration, element string
		array, unsupported   bool
	}{
		{"type Target = number[];", "number", true, false},
		{"type Target = readonly string[];", "string", true, false},
		{"type Target = { readonly name: string }[];", "{ readonly name: string; }", true, false},
		{"type Target = (string | number)[];", "string | number", true, false},
		{"type Target = number;", "", false, false},
		{"type Target = readonly [number, string?];", "", true, true},
	} {
		t.Run(test.declaration, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(test.declaration), 0600); err != nil {
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
			node := file.Statements.Nodes[0]
			target := checked.GetTypeAtLocation(node.Name())
			calls := 0
			sentinel := errors.New("nested contract unavailable")
			handled, err := l.viewArrayContract(node, target, func(element *checker.Type) error {
				calls++
				if got := checked.TypeToString(element); got != test.element {
					t.Fatalf("element %q, want %q", got, test.element)
				}
				return sentinel
			})
			if handled != test.array {
				t.Fatalf("handled %t, want %t", handled, test.array)
			}
			if test.unsupported {
				if _, ok := err.(*NotYet); !ok || calls != 0 {
					t.Fatalf("tuple must fail closed, got %v, calls %d", err, calls)
				}
			} else if test.array {
				if !errors.Is(err, sentinel) || calls != 1 {
					t.Fatalf("nested contract lost: %v, calls %d", err, calls)
				}
			} else if err != nil || calls != 0 {
				t.Fatalf("non-array: %v, calls %d", err, calls)
			}
		})
	}
}
