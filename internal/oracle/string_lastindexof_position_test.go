package oracle

import (
	"bytes"
	"math"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/string_lastindexof_position.a", true, false})
}

func TestStringLastIndexOfPositionMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_lastindexof_position.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %+v", truth)
	}
	control, sanitized := natively(t, program)
	if difference := disagreement(truth, control); difference != "" {
		t.Fatalf("unmutated native: %s (exit %d, stderr %q)", difference, control.exitCode, control.stderr)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("unmutated JavaScript: %s", difference)
	}
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
	t.Logf("unmutated native and JavaScript agree with Node on %d output lines; exit 0, no stderr or leaks", bytes.Count(truth.stdout, []byte{'\n'}))
	code := native.C(program)
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "string_last_index_of_position" {
			continue
		}
		// Position normalization moved from the runtime call into this lowered
		// helper. Change its NaN arm, preserving operand evaluation and clamping.
		mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
			conditional, ok := value.(ir.Conditional)
			if !ok {
				return value
			}
			condition, ok := conditional.Condition.(ir.NumberCall)
			if !ok || condition.Function != "isNaN" {
				return value
			}
			infinity, ok := conditional.WhenTrue.(ir.NumberConstant)
			if !ok || !math.IsInf(infinity.Value, 1) {
				return value
			}
			conditional.WhenTrue = ir.NumberConstant{Value: 0}
			changed++
			return conditional
		})
	}
	mutant := native.C(program)
	if changed == 0 || mutant == code {
		t.Fatal("mutant changed no code")
	}
	t.Logf("mutated %d lastIndexOf NaN normalization arms; generated C changed", changed)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must execute cleanly: exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(truth, result); difference != "stdout differs" {
		t.Fatalf("Node must catch mutant: %q", difference)
	}
	t.Log("NaN treated as zero: clean exit 0, no sanitizer finding; only Node stdout comparison caught it")
}
