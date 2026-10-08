package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestHiddenTNodeConstraintRepresentation(t *testing.T) {
	for _, probe := range []struct {
		name, constraint string
		want             ir.Type
		known            bool
	}{
		{"number", " extends number", ir.Number, true},
		{"string", " extends string", ir.String, true},
		{"boolean", " extends boolean", ir.Boolean, true},
		{"object", " extends { readonly name: string }", ir.Union, true},
		{"object union", " extends { readonly kind: 1 } | { readonly kind: 2 }", ir.Union, true},
		{"structural length", " extends { readonly length: number }", 0, false},
		{"unconstrained", "", 0, false},
		{"unknown", " extends unknown", 0, false},
		{"any", " extends any", 0, false},
		{"array layout", " extends readonly number[]", 0, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte("function visit<TNode"+probe.constraint+">(node: TNode): TNode { return node; }"), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			check, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: check}
			parameter := check.GetTypeAtLocation(file.Statements.Nodes[0].TypeParameters()[0].Name())
			if held, known := l.representation(parameter); held != probe.want || known != probe.known {
				t.Fatalf("constraint representation = (%v, %v), want (%v, %v)", held, known, probe.want, probe.known)
			}
			// signature declares this actual parameter through typeOf; the binder
			// and parameter must use the same constraint storage path.
			parameterNode := file.Statements.Nodes[0].Parameters()[0].Name()
			held, parameterErr := l.typeOf(parameterNode)
			if probe.known && (parameterErr != nil || held != probe.want) {
				t.Fatalf("parameter storage = %v, %v; want %v", held, parameterErr, probe.want)
			}
			if !probe.known && parameterErr == nil {
				t.Fatalf("unsupported parameter constraint accepted: %v", held)
			}
			// A concrete substitution must win even when its storage differs from the constraint.
			l.substitution = map[*checker.Type]ir.Type{parameter: ir.Array}
			if held, known := l.representation(parameter); !known || held != ir.Array {
				t.Fatalf("lost concrete substitution: %v, %v", held, known)
			}
			l.substitution = nil
			l.typeMapper = newTypeMapper([]*checker.Type{parameter}, []*checker.Type{check.GetNumberType()})
			if held, known := l.representation(parameter); !known || held != ir.Number {
				t.Fatalf("lost concrete checker mapping: %v, %v", held, known)
			}
		})
	}
}

func TestHiddenTNodeGenericValueRemainsNotYet(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/hidden_boundary_generic_tnode_value.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "a generic function as a value" {
		t.Fatalf("want generic value refusal, got %v", err)
	}
}

func TestHiddenTNodeConstraintMutationRemainsNotYet(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/hidden_boundary_generic_tnode_mutation.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "assigning a field of a union of differently held members" {
		t.Fatalf("want constrained mutation refusal, got %v", err)
	}
}
