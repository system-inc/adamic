package parser

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestStrongAstParentSupported(t *testing.T) {
	path, err := filepath.Abs("gaps/1_strong_ast_parent.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "root\n" {
		t.Fatalf("Node cycle result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(lowered)
	for _, side := range []struct {
		name     string
		sanitize bool
	}{{"release", false}, {"sanitized", true}} {
		t.Run(side.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "cycle")
			if err := native.Build(code, binary, native.Options{Sanitize: side.sanitize}); err != nil {
				t.Fatal(err)
			}
			answer := execute(t, "", binary)
			if !bytes.Equal(answer.output, result.output) {
				t.Fatalf("%s cycle output %q differs from Node %q", side.name, answer.output, result.output)
			}
		})
	}
}

func TestPushSpreadGap(t *testing.T) {
	path, err := filepath.Abs("gaps/2_push_spread.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1,2,3\n" {
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a SpreadElement" {
		t.Fatalf("GAPS.md records push spread NotYet, got %v", err)
	}
}

func TestTypeImportCycleGap(t *testing.T) {
	path, err := filepath.Abs("gaps/3_import_cycle.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1\n" {
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || refusal.What != "an import cycle" {
		t.Fatalf("GAPS.md records type-only import cycle refusal, got %v", err)
	}
}

func TestClassMethodInterfaceGap(t *testing.T) {
	path, err := filepath.Abs("gaps/4_class_interface_method.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1\n" {
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a class method through a view that erases its prototype origin" {
		t.Fatalf("GAPS.md records an erased class-method origin NotYet, got %v", err)
	}
	t.Logf("Node prints 1; Adamic reports: %v", notYet)
}

func TestOptionalFunctionValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/5_optional_function_value.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1\n" {
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a function value with an optional parameter" {
		t.Fatalf("GAPS.md records optional function value NotYet, got %v", err)
	}
}

func TestConditionalEmptyArrayGap(t *testing.T) {
	path, err := filepath.Abs("gaps/6_conditional_empty_array.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(result.output) != "1\n" {
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "an array of never" {
		t.Fatalf("GAPS.md records empty conditional array NotYet, got %v", err)
	}
}
