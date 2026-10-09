package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/from_code_point_catch.a", true, false})
}

// ToUint16 in place of code-point validation finishes cleanly, but loses the catch and finally
// observations that only the original source on Node can determine.
func TestFromCodePointCatchMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/from_code_point_catch.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if codes, ok := value.(ir.StringFromCodes); ok && codes.CodePoints {
			codes.CodePoints = false
			changed++
			return codes
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
	if changed == 0 {
		t.Fatal("no code-point calls mutated")
	}
	result, binary := natively(t, program)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %+v", result)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
		t.Fatalf("want stdout alone, got %q", difference)
	}
	t.Logf("Node alone caught %d calls using ToUint16 instead of RangeError validation", changed)
}
