package parser

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
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

func TestClosedOptionalFunctionValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/5_optional_function_value.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	if string(expected) != "1\n" {
		t.Fatalf("Node optional function result %q", expected)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "optional")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASAN_OPTIONS", "detect_leaks=1")
	actual := execute(t, "", binary).output
	emitted := filepath.Join(t.TempDir(), "optional.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	backend := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted).output
	for _, side := range []struct {
		name   string
		output []byte
	}{{"native ASan/UBSan/LSan", actual}, {"JavaScript backend", backend}} {
		if !bytes.Equal(side.output, expected) {
			t.Fatalf("%s: %q, Node %q", side.name, side.output, expected)
		}
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
