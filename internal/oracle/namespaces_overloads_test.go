package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

func TestNamespaceIdenticalOverloadMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/namespaces/debug-groups/identical-overloads.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i := range program.Functions {
		if program.Functions[i].Name == "assertEachNode" {
			program.Functions[i].Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 99}}}
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed no overload implementation")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", result)
	}
	expected := onNode(t, path)
	if difference := disagreement(expected, result); difference != "stdout differs" {
		t.Fatalf("native mutant not caught: %s", difference)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant not caught: %s", difference)
	}
	t.Log("wrong overload implementation caught by Node in both backends")
}
