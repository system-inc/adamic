package parser

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestStrongAstParentGap(t *testing.T) {
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
		t.Fatalf("Node gap result %q", result.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "a cycle reference counting can't free") {
		t.Fatalf("GAPS.md says strong parent/child cycle is refused, got %v", err)
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
