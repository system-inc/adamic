package oracle

import (
	"os"
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
	}{"internal/oracle/testdata/library_string_tail_repeat.a", true, false})
}
func TestStringTailRepeatMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_tail_repeat.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
		if operation, ok := value.(ir.Binary); ok && operation.Operator == ir.Remainder {
			if n, ok := operation.Right.(ir.NumberConstant); ok && n.Value == 2 {
				changed++
				operation.Right = ir.NumberConstant{Value: 3}
				return operation
			}
		}
		return value
	})
	if changed == 0 {
		t.Fatal("repeat doubling mutant missed its target")
	}
	expected := onNode(t, path)
	actual, binary := natively(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", actual)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, result := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(expected, result); difference != "stdout differs" {
			t.Fatalf("%s: %q", name, difference)
		}
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "stdout differs" {
			t.Fatalf("WASI: %q", difference)
		}
	}
	t.Log("Node alone caught the wrong doubling arm; clean exit and no native leaks")
}
