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
func TestBoundaryPreservesErasedFacts(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/erased.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	contents, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	source := string(contents)
	compile := func(text string) *ir.Program {
		test.Helper()
		loaded, loadOverlayError := load.LoadOverlay([]string{path}, map[string]string{path: text})
		if loadOverlayError != nil {
			test.Fatal(loadOverlayError)
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			test.Fatal(lowerError)
		}
		directory := test.TempDir()
		ordinary := filepath.Join(directory, "ordinary.a")
		module := filepath.Join(directory, "compiled.mjs")
		binary := filepath.Join(directory, "compiled")
		for path, content := range map[string]string{ordinary: text, module: javascript.JavaScript(program)} {
			if writeError := os.WriteFile(path, []byte(content), 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
		runner, absolutePathError := filepath.Abs("../../oracle/node.mjs")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		truthCommand := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, ordinary)
		truthCommand.Env = append(os.Environ(), jsonTypesEnvironment(test, ordinary)...)
		truth, combinedOutputError := truthCommand.CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("source oracle: %v: %s", combinedOutputError, truth)
		}
		if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
			test.Fatal(buildError)
		}
		for _, command := range []*exec.Cmd{
			exec.Command("node", "--disable-warning=ExperimentalWarning", runner, module),
			exec.Command(binary),
		} {
			output, combinedOutputError := command.CombinedOutput()
			if combinedOutputError != nil || string(output) != string(truth) {
				test.Fatalf("%v: %v: got %q, oracle %q", command.Args, combinedOutputError, output, truth)
			}
		}
		test.Logf("source Node, emitted JavaScript, sanitized native agree: %q", truth)
		return program
	}
	baseline := compile(source)
	for name, changed := range map[string]string{
		"mutable global": strings.Replace(source, "const limit", "let limit", 1),
		"mutable array":  strings.Replace(source, "readonly number[]", "number[]", 1),
		"boolean array":  strings.Replace(source, "readonly number[]", "readonly boolean[]", 1),
	} {
		test.Run(name, func(test *testing.T) {
			if name == "boolean array" {
				changed = strings.Replace(changed, "[1, 2, 3]", "[true, false, true]", 1)
			}
			other := compile(changed)
			// The boolean caller needs different literals; compare the analyzed
			// function boundary, which must retain the different element type.
			if name == "boolean array" {
				if reflect.DeepEqual(baseline.Functions[0].Boundary, other.Functions[0].Boundary) {
					test.Fatal("boundary lost array element type")
				}
			} else {
				baseCopy, otherCopy := *baseline, *other
				baseCopy.Functions = append([]ir.Function(nil), baseline.Functions...)
				otherCopy.Functions = append([]ir.Function(nil), other.Functions...)
				for index := range baseCopy.Functions {
					baseCopy.Functions[index].Boundary = ir.FunctionBoundary{}
					otherCopy.Functions[index].Boundary = ir.FunctionBoundary{}
				}
				// Main records checker-local type IDs in the IR; compare emitted runtime code,
				// whose equality does not depend on those IDs. Both backends must agree.
				if javascript.JavaScript(&baseCopy) != javascript.JavaScript(&otherCopy) ||
					native.C(&baseCopy) != native.C(&otherCopy) {
					test.Fatal("relaxed-equivalent programs have different runtime IR")
				}
			}
			test.Log("boundary retains the facts required by the relaxed rules")
		})
	}
}
