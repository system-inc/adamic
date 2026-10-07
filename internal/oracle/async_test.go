package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Generated-C mutants keep the lowering and runtime builds unchanged; warning failures are not kills.
func TestAsyncCompilerChecksCatchMutants(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_three.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	oracle := onNode(t, path)
	t.Run("wrong-resumed-value", func(t *testing.T) {
		code := strings.Replace(control, "resumed.number", "(resumed.number + 1)", 1)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only the Node stdout check must catch the mutant: %+v", result)
		}
		t.Log("wrong resumed value caught only by Node stdout comparison")
	})
	t.Run("missing-frame-parameter-retain", func(t *testing.T) {
		code := regexp.MustCompile(`(frame->cells\[[0-9]+\].value.reference = )adamic_retain\((adamic_local_[0-9]+_seed)\)`).ReplaceAllString(control, `${1}${2}`)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
		if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-use-after-free") {
			t.Fatalf("ASan did not catch a borrowed dynamic parameter across await: %+v", result)
		}
		t.Log("missing frame parameter retain caught by ASan heap-use-after-free")
	})
}

// The source oracle covers the uncaught throw. This internal root harness consumes that rejection
// instead of panicking, so Linux LSan can check frame cleanup even though the source exits nonzero.
func TestAsyncThrowReleasesFrame(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_throw.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	original := "if (root->rejected) { adamic_thrown = adamic_retain(root->value.reference); adamic_uncaught(); }"
	if !strings.Contains(code, original) {
		t.Fatal("missing root cleanup hook")
	}
	code = strings.Replace(code, original, "if (!root->rejected) return 2;", 1)
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(t *testing.T) {
			source := code
			if mutant {
				start := strings.Index(source, "adamic_release(adamic_slot_")
				end := strings.Index(source[start:], ";") + start
				if start < 0 || end < start {
					t.Fatal("mutant did not change input")
				}
				source = source[:start] + "(void)" + source[start+len("adamic_release("):end-1] + source[end:]
			}
			binary := filepath.Join(t.TempDir(), "throw")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if mutant {
				if result.exitCode == 0 || !strings.Contains(string(result.stderr), "LeakSanitizer: detected memory leaks") {
					t.Fatalf("leak mutant survived: %+v", result)
				}
				t.Log("throw frame local release mutant caught by LeakSanitizer")
			} else if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("throw root cleanup leaked: %+v", result)
			}
		})
	}
}

// Each callable and control-flow surface must expose a wrong fulfillment payload to Node.
func TestAsyncExpandedChecksCatchMutants(t *testing.T) {
	for _, fixture := range []string{"async_control", "async_operand_order", "async_methods", "async_closures", "async_conditions"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			control := native.C(program)
			code := strings.ReplaceAll(control, "resumed.number", "(resumed.number + 1)")
			if code == control {
				t.Fatal("no numeric suspension payload to mutate")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if disagreement(onNode(t, path), result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Node output did not exclusively catch fulfillment mutant: %+v", result)
			}
		})
	}
}
