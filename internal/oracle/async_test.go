package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
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
		code := strings.Replace(control, "= resumed.number;", "= resumed.number + 1;", 1)
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
		code := strings.Replace(control, "= adamic_retain(argument_0);", "= argument_0;", 1)
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
	begin := strings.Index(code, " if (root->rejected) {")
	if begin < 0 {
		t.Fatal("missing root cleanup hook")
	}
	end := strings.Index(code[begin:], " adamic_release(root); return 0;") + begin
	if end < begin {
		t.Fatal("missing root cleanup hook")
	}
	code = code[:begin] + " if (!root->rejected) return 2;\n" + code[end:]
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(t *testing.T) {
			source := code
			if mutant {
				start := strings.Index(source, "drop(frame->local_")
				end := strings.Index(source[start:], ";") + start
				if start < 0 || end < start {
					t.Fatal("mutant did not change input")
				}
				source = source[:start] + "(void)frame" + source[end:]
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
