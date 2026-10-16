package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
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
		code := strings.Replace(control, "= (double)resumed.number;", "= (double)resumed.number + 1;", 1)
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only the Node stdout check must catch the mutant: %+v", result)
		}
		if report := leakcheck.Report(t, code, binary); report != "" {
			t.Fatal(report)
		}
		t.Log("wrong resumed value caught only by Node stdout comparison")
	})
	t.Run("missing-frame-parameter-retain", func(t *testing.T) {
		parameter := regexp.MustCompile(`(frame->cells\[[0-9]+\]\.value\.reference = )adamic_retain\(([^)]+)\);`)
		location := parameter.FindStringSubmatchIndex(control)
		if location == nil {
			t.Fatal("missing frame parameter store")
		}
		code := control[:location[0]] + parameter.ReplaceAllString(control[location[0]:location[1]], "${1}${2};") + control[location[1]:]
		if code == control {
			t.Fatal("mutant did not change input")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		result := execute(t, binary)
		if result.exitCode == 0 || !strings.Contains(string(result.stderr), "heap-use-after-free") {
			t.Fatalf("ASan did not catch a borrowed dynamic parameter across await: %+v", result)
		}
		t.Log("missing frame parameter retain caught by ASan heap-use-after-free")
	})
}

// The source oracle covers the uncaught throw. This internal root harness consumes that rejection
// instead of panicking, so the shared leak checker can check frame cleanup on Linux and Darwin.
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
	begin := strings.Index(code, "if (root->rejected) {")
	if begin < 0 {
		t.Fatal("missing root cleanup hook")
	}
	end := strings.Index(code[begin:], "\n") + begin
	if end < begin {
		t.Fatal("missing root cleanup hook")
	}
	code = code[:begin] + "if (!root->rejected) return 2;" + code[end:]
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "missing-local-release"}[mutant], func(t *testing.T) {
			source := code
			if mutant {
				slot := regexp.MustCompile(`adamic_release\((adamic_slot_[0-9]+)\);`)
				location := slot.FindStringSubmatchIndex(source)
				if location == nil {
					t.Fatal("missing private frame slot release")
				}
				source = source[:location[0]] + slot.ReplaceAllString(source[location[0]:location[1]], "(void)$1;") + source[location[1]:]
			}
			binary := filepath.Join(t.TempDir(), "throw")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			report := leakcheck.Report(t, source, binary)
			// Exercise Darwin's counted rule on Linux too, using this actual frame mutant.
			counted := filepath.Join(t.TempDir(), "counted-proof")
			if err := native.Build(source, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			countedReport := leakcheck.Unbalanced(leakRun(execute(t, counted)))
			if mutant {
				if !strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "heap values leaked:") {
					t.Fatalf("leak mutant survived: %s", report)
				}
				if !strings.Contains(countedReport, "heap values leaked:") {
					t.Fatalf("counted rule missed local leak: %s", countedReport)
				}
				t.Logf("throw frame local release mutant caught: %s; counted rule: %s", report, countedReport)
			} else if report != "" || countedReport != "" {
				t.Fatalf("throw root cleanup leaked: %s; counted rule: %s", report, countedReport)
			}
		})
	}
}

// Mutate the shared lowered typeof operand so both backends can agree with each other and still be
// wrong. The source on Node must independently catch the mistaken async function classification.
func TestAsyncTypeOfWrongKindMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_typeof.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	want := "function\nfunction\nnested: function\nstring\nnumber\nshadow after await: number\ninside: function\nfunction\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		t.Fatalf("Node witness: %+v", oracle)
	}
	changed := false
	wrongKind := len(program.Strings)
	program.Strings = append(program.Strings, "object")
	var mutate func(reflect.Value) reflect.Value
	mutate = func(value reflect.Value) reflect.Value {
		if value.Kind() == reflect.Interface {
			result := reflect.New(value.Type()).Elem()
			if !value.IsNil() {
				result.Set(mutate(value.Elem()))
			}
			return result
		}
		if observation, ok := value.Interface().(ir.TypeOf); ok && observation.Value.Type() == ir.Closure {
			changed = true
			return reflect.ValueOf(ir.StringConstant{Index: wrongKind})
		}
		switch value.Kind() {
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				result.Field(index).Set(mutate(value.Field(index)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for index := 0; index < value.Len(); index++ {
				result.Index(index).Set(mutate(value.Index(index)))
			}
			return result
		}
		return value
	}
	program.Main = mutate(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = mutate(reflect.ValueOf(program.Functions[index].Body)).Interface().([]ir.Statement)
	}
	if !changed {
		t.Fatal("mutant did not change an async typeof operand")
	}
	binary := filepath.Join(t.TempDir(), "wrong-kind")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Report(t, native.C(program), binary); report != "" {
		t.Fatal(report)
	}
	for name, result := range map[string]run{
		"native":     execute(t, binary),
		"JavaScript": onJavaScriptBackend(t, program),
	} {
		if disagreement(oracle, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("only Node stdout must catch %s's wrong-kind mutant: %+v", name, result)
		}
		t.Logf("Node stdout catches %s's async function classified as object", name)
	}
}

// Keep the inherited fixtures in their registration order; new witnesses are
// explicit entries in oracle_test.go and run in the same bounded families.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/async_coverage_discard.a",
		"internal/oracle/testdata/async_coverage_parameters.a",
		"internal/oracle/testdata/async_coverage_reject_eager.a",
		"internal/oracle/testdata/async_coverage_reject_empty.a",
		"internal/oracle/testdata/async_coverage_reject_nested.a",
		"internal/oracle/testdata/async_coverage_typeof.a",
		"internal/oracle/testdata/async_coverage_unions.a",
		"internal/oracle/testdata/async_coverage_values.a",
		"internal/oracle/testdata/async_nested.a",
		"internal/oracle/testdata/async_plain.a",
		"internal/oracle/testdata/async_three.a",
		"internal/oracle/testdata/async_throw.a",
		"internal/oracle/testdata/async_typeof.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}
