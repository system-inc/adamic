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

func hiddenTNodeConstraint(t *testing.T, constraint string, want ir.Type, wantKnown bool) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte("function visit<TNode"+constraint+">(node: TNode): TNode { return node; }"), 0644); err != nil {
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
	if held, known := l.representation(parameter); held != want || known != wantKnown {
		t.Fatalf("constraint representation = (%v, %v), want (%v, %v)", held, known, want, wantKnown)
	}
	// Parameters and their binders must use the same constraint storage rule.
	parameterNode := file.Statements.Nodes[0].Parameters()[0].Name()
	held, parameterErr := l.typeOf(parameterNode)
	if wantKnown && (parameterErr != nil || held != want) {
		t.Fatalf("parameter storage = %v, %v; want %v", held, parameterErr, want)
	}
	if !wantKnown && parameterErr == nil {
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
}

func TestHiddenTNodeConstraintNumber(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends number", ir.Number, true)
}

func TestHiddenTNodeConstraintString(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends string", ir.String, true)
}

func TestHiddenTNodeConstraintBoolean(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends boolean", ir.Boolean, true)
}

func TestHiddenTNodeConstraintObject(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends { readonly name: string }", ir.Union, true)
}

func TestHiddenTNodeConstraintObjectUnion(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends { readonly kind: 1 } | { readonly kind: 2 }", ir.Union, true)
}

func TestHiddenTNodeConstraintStructuralLength(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends { readonly length: number }", 0, false)
}

func TestHiddenTNodeConstraintUnconstrained(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, "", 0, false)
}

func TestHiddenTNodeConstraintUnknown(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends unknown", 0, false)
}

func TestHiddenTNodeConstraintAny(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends any", 0, false)
}

func TestHiddenTNodeConstraintArrayLayout(t *testing.T) {
	t.Parallel()
	hiddenTNodeConstraint(t, " extends readonly number[]", 0, false)
}

func TestHiddenTNodeGenericValueRemainsNotYet(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/notyet/hidden_boundary_generic_tnode_value.a")
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
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/notyet/hidden_boundary_generic_tnode_mutation.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "assigning a field of a union of differently held members" {
		t.Fatalf("want constrained mutation refusal, got %v", err)
	}
}
