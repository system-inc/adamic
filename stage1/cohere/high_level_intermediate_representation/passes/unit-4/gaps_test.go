package unit4

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestArrayNeverGap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct{ file, want string }{{"array-never.a", "0\n"}, {"array-never-coalesce.a", "1\n"}} {
		entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4/gaps", fixture.file)
		if got := string(unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", entry)); got != fixture.want {
			t.Fatalf("Node %s: %q", fixture.file, got)
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		_, err = lower.Lower(context.Background(), program)
		if err == nil {
			t.Fatal("this gap lowers now: mark it closed and undo the port's gap 5 workaround")
		}
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || notYet.What != "an array of never" {
			t.Fatalf("stage 0 refuses another way now: %v", err)
		}
	}
}

func TestRecursiveInitializerGap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-4/gaps/recursive-initializer.a")
	if got := string(unit4Run(t, root, "node", "--no-warnings", "oracle/node.mjs", entry)); got != "0\n" {
		t.Fatalf("Node: %q", got)
	}
	// Compiler #dv99xzy retains the original initializer shape.
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a function value that captures the variable its own initializer declares" {
		t.Fatalf("recursive initializer gap closed or changed: %v", err)
	}
}
