package oracle

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func TestProcessExitSafeInteger_000(t *testing.T)         { t.Parallel(); processExitSafeInteger(t, 0) }
func TestProcessExitSafeInteger_001(t *testing.T)         { t.Parallel(); processExitSafeInteger(t, 1) }
func TestProcessExitSafeInteger_002(t *testing.T)         { t.Parallel(); processExitSafeInteger(t, 2) }
func TestProcessExitSafeIntegerRangeMutant(t *testing.T)  { t.Parallel(); processExitSafeInteger(t, 3) }
func TestProcessExitSafeIntegerFormatMutant(t *testing.T) { t.Parallel(); processExitSafeInteger(t, 4) }
func TestProcessExitSafeIntegerScientificMutant(t *testing.T) {
	t.Parallel()
	processExitSafeInteger(t, 5)
}

func processExitSafeInteger(t *testing.T, shard int) {
	t.Helper()
	values := []float64{math.Pow(2, 53), -math.Pow(2, 53), 1e21, -1e21, 1.2e21, 1e100, -1e100, math.MaxFloat64}
	if shard >= 3 {
		values = values[:1]
	}
	if shard == 5 {
		values = []float64{1e100}
	}
	for _, value := range values {
		source := fmt.Sprintf("process.exitCode=7;try{process.exitCode=%g;}catch(error){console.log(error.code);console.log(error.message);}process.exitCode=0;", value)
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		truth := onNodeWith(t, inputRun{}, path)
		if truth.exitCode != 0 || len(truth.stderr) != 0 || !strings.HasPrefix(string(truth.stdout), "ERR_OUT_OF_RANGE\n") {
			t.Fatalf("unexpected Node safe-integer contract: %+v", truth)
		}
		setter := func(number float64) ir.Statement {
			return ir.Evaluate{Value: ir.ProcessCall{Operation: "setExitCode", Of: ir.MaybeNumber, Arguments: []ir.Expression{ir.MaybeOf{Of: ir.MaybeNumber, Value: ir.NumberConstant{Value: number}}}}}
		}
		program := &ir.Program{Source: "process-exit-safe-integer.a", Locals: []ir.Local{{Name: "error", Type: ir.Object, Function: -1}}}
		program.Main = []ir.Statement{setter(7), ir.Try{HasCatch: true, CatchLocal: 0, Body: []ir.Statement{setter(value)}, Catch: []ir.Statement{
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "code", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "message", Of: ir.String}},
		}}, setter(0)}
		if shard < 3 {
			processExitSafeIntegerAgreement(t, program, truth, shard, value)
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/process.c"))
		if err != nil {
			t.Fatal(err)
		}
		before, after := "bool range = integer && fabs(code.number) > 9007199254740991.0;", "bool range = false;"
		if shard == 4 {
			before, after = "exit_code_text(code.number, range)", "exit_code_text(code.number, false)"
		}
		if shard == 5 {
			before = "if (!grouped) { return text; }"
			after = before + "\nfor (size_t index = 0; index < text->length; index++) { if (text->bytes[index] == 'e') { return text; } }"
		}
		if strings.Count(string(data), before) != 1 {
			t.Fatal("safe-integer mutation anchor changed")
		}
		implementation := strings.Replace(string(data), before, after, 1)
		code := native.C(program)
		for _, name := range []string{"adamic_process_exit_code", "adamic_process_set_exit_code", "adamic_process_exit", "adamic_process_is_tty", "adamic_process_environment"} {
			implementation = strings.ReplaceAll(implementation, name+"(", "safe_mutant_"+name+"(")
			code = strings.ReplaceAll(code, name+"(", "safe_mutant_"+name+"(")
		}
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(implementation+"\n"+code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
			t.Fatalf("safe-integer mutant failed outside Node stdout: %+v", got)
		}
		t.Log("mutant compiled and ran cleanly; only Node stdout caught it")
	}
}

func processExitSafeIntegerAgreement(t *testing.T, program *ir.Program, truth run, shard int, value float64) {
	t.Helper()
	check := func(backend string, got run) {
		t.Helper()
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%g %s: %s; Node stdout=%q stderr=%q status=%d; Adamic stdout=%q stderr=%q status=%d", value, backend, difference, truth.stdout, truth.stderr, truth.exitCode, got.stdout, got.stderr, got.exitCode)
		}
	}
	switch shard {
	case 0:
		got, binary := nativelyUncached(t, program)
		check("native", got)
		check("JavaScript", nodeProcessJavaScript(t, program))
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	case 1:
		check("release", releasedUncached(t, program))
	case 2:
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			t.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		check("wasm32", onWASI(t, native.C(program)))
	default:
		t.Fatalf("unknown safe-integer shard %d", shard)
	}
}
