package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// The C runtime accepts reference slots before the source constructor does.
// Start from a lowered, dense string array, replace only its allocation with the
// runtime's length-form allocation, and retain every logical element type.
func arrayHolesReferenceRuntimeFixture(t *testing.T) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_array_holes_parallel_references.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var pending *lower.NotYet
	if !errors.As(err, &pending) || pending.What != "holey Array of this element representation" {
		t.Fatalf("constructor refusal: %v", err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), "new Array<string>(4294967295)", "[]", 1))
	shadow := filepath.Join(t.TempDir(), "runtime-witness.a")
	if err := os.WriteFile(shadow, source, 0644); err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, shadow)
	if err != nil {
		t.Fatal(err)
	}
	replacements := 0
	for i, statement := range p.Main {
		switch d := statement.(type) {
		case ir.Declare:
			if p.Locals[d.Local].Name == "cells" {
				d.Value = ir.ArrayHoles{Length: ir.NumberConstant{Value: 4294967295}, Element: ir.String}
				p.Main[i] = d
				replacements++
			}
		case ir.Assign:
			if p.Locals[d.Local].Name == "cells" {
				d.Value = ir.ArrayHoles{Length: ir.NumberConstant{Value: 4294967295}, Element: ir.String}
				p.Main[i] = d
				replacements++
			}
		}
	}
	if replacements != 1 {
		t.Fatalf("reference allocation replacements: %d", replacements)
	}
	p.ClosuresMayThrow = true
	if !usesParallelMap(p) {
		t.Fatal("fixture did not lower actual parallel work")
	}
	return p, path
}

// Not parallel: sanitizer variants run worker pools and measure leak cleanup.
func TestArrayHolesReferenceParallel(t *testing.T) {
	p, path := arrayHolesReferenceRuntimeFixture(t)
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "4294967295:hole:ok:ok\n" {
		t.Fatalf("Node: %#v", truth)
	}
	checkParallelVariants(t, p, truth)
	if difference := disagreement(truth, onJavaScriptBackend(t, p)); difference != "" {
		t.Fatal(difference)
	}
}

func TestArrayHolesSparseShareMutant(t *testing.T) {
	p, path := arrayHolesReferenceRuntimeFixture(t)
	truth := onNode(t, path)
	mutant := arrayRuntimeMutant(t, native.C(p), "share.c",
		"if (array->sparse != NULL) {\n\t\t\t\tappend(&pending, array->sparse);\n\t\t\t} else if (array->references)",
		"if (array->references)", native.Options{Sanitize: true})
	if disagreement(truth, mutant) == "" || mutant.exitCode == 0 {
		t.Fatalf("sparse sharing omission escaped: %#v", mutant)
	}
	if !strings.Contains(string(mutant.stderr), "AddressSanitizer") && !strings.Contains(string(mutant.stderr), "runtime error:") {
		t.Fatalf("no sanitizer evidence: %s", mutant.stderr)
	}
	t.Log("sparse sharing walk omitted; reference array witness catches the null elements access")
}

func TestArrayHolesSparseFreeMutant(t *testing.T) {
	p, path := arrayHolesReferenceRuntimeFixture(t)
	truth := onNode(t, path)
	mutant := arrayRuntimeMutant(t, native.C(p), "heap.c",
		"if (array->sparse != NULL) {\n\t\t\tlet_go(array->sparse);\n\t\t} else if (array->references)",
		"if (array->references)", native.Options{Sanitize: true})
	if disagreement(truth, mutant) == "" || mutant.exitCode == 0 {
		t.Fatalf("sparse child cleanup omission escaped: %#v", mutant)
	}
	if !strings.Contains(string(mutant.stderr), "AddressSanitizer") && !strings.Contains(string(mutant.stderr), "runtime error:") {
		t.Fatalf("no sanitizer evidence: %s", mutant.stderr)
	}
	t.Log("free_one sparse branch omitted; reference array witness catches the null elements access")
}

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		p, _ := arrayHolesReferenceRuntimeFixture(t)
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(native.C(p), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		name, args := pinnedStack(binary)
		got := executeWith(t, []string{"ADAMIC_THREADS=1"}, name, args...)
		match := countsLine.FindSubmatch(got.stderr)
		if got.exitCode != 0 || match == nil {
			t.Fatalf("counts: %#v", got)
		}
		return []string{fmt.Sprintf("| internal/oracle/testdata/library_array_holes_parallel_references.a | %s | %s | %s | %s | %s | %s |", match[1], match[2], match[3], match[4], match[5], match[6])}
	})
}
