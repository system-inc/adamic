package split_test

import (
	"context"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The relaxed rules intentionally treat unwritten let globals and mutable arrays
// like their const and readonly twins. Element types must now remain distinguishable.
func TestBoundaryPreservesErasedFacts(t *testing.T) {
	path, err := filepath.Abs("testdata/erased.a")
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	compile := func(text string) *ir.Program {
		t.Helper()
		loaded, err := load.LoadOverlay([]string{path}, map[string]string{path: text})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		directory := t.TempDir()
		ordinary := filepath.Join(directory, "ordinary.a")
		module := filepath.Join(directory, "compiled.mjs")
		binary := filepath.Join(directory, "compiled")
		for path, content := range map[string]string{ordinary: text, module: javascript.JavaScript(program)} {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		runner, err := filepath.Abs("../../oracle/node.mjs")
		if err != nil {
			t.Fatal(err)
		}
		truth, err := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, ordinary).CombinedOutput()
		if err != nil {
			t.Fatalf("source oracle: %v: %s", err, truth)
		}
		if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		for _, command := range []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", runner, module), exec.Command(binary)} {
			output, err := command.CombinedOutput()
			if err != nil || string(output) != string(truth) {
				t.Fatalf("%v: %v: got %q, oracle %q", command.Args, err, output, truth)
			}
		}
		t.Logf("source Node, emitted JavaScript, sanitized native agree: %q", truth)
		return program
	}
	baseline := compile(source)
	for name, changed := range map[string]string{
		"mutable global": strings.Replace(source, "const limit", "let limit", 1),
		"mutable array":  strings.Replace(source, "readonly number[]", "number[]", 1),
		"boolean array":  strings.Replace(source, "readonly number[]", "readonly boolean[]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if name == "boolean array" {
				changed = strings.Replace(changed, "[1, 2, 3]", "[true, false, true]", 1)
			}
			other := compile(changed)
			// The boolean caller needs different literals; compare the analyzed
			// function boundary, which must retain the different element type.
			if name == "boolean array" {
				if reflect.DeepEqual(baseline.Functions[0].Boundary, other.Functions[0].Boundary) {
					t.Fatal("boundary lost array element type")
				}
			} else {
				baseCopy, otherCopy := *baseline, *other
				baseCopy.Functions = append([]ir.Function(nil), baseline.Functions...)
				otherCopy.Functions = append([]ir.Function(nil), other.Functions...)
				for i := range baseCopy.Functions {
					baseCopy.Functions[i].Boundary = ir.FunctionBoundary{}
					otherCopy.Functions[i].Boundary = ir.FunctionBoundary{}
				}
				if !reflect.DeepEqual(baseCopy, otherCopy) {
					t.Fatal("relaxed-equivalent programs have different runtime IR")
				}
			}
			t.Log("boundary retains the facts required by the relaxed rules")
		})
	}
}
