package parser

import (
	"bytes"
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
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
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "method-gap")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	var stderr bytes.Buffer
	command := exec.Command(binary)
	command.Stdout = output
	command.Stderr = &stderr
	err = command.Run()
	var status *exec.ExitError
	if !errors.As(err, &status) || status.ExitCode() != 70 || !strings.Contains(stderr.String(), "compiler bug: a field the checker proved is there is missing") {
		t.Fatalf("GAPS.md records native structural method dispatch panic, got %v %q", err, stderr.String())
	}
}
