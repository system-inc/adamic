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
	}{"internal/oracle/testdata/object_is_null_undefined.a", true, false})
}

// Restore comparison of the erased null pointers. It compiles and finishes leak-clean; only
// Node's original source can expose the wrong SameValue result.
func TestObjectIsNullUndefinedMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_is_null_undefined.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if conditional, ok := value.(ir.Conditional); ok {
			if call, ok := conditional.Condition.(ir.ObjectCall); ok && call.Method == "is" {
				// The area now boxes null with a distinct sentinel. Erase that
				// tag too, so this still restores the erroneous null/undefined
				// pointer comparison rather than a correct tagged SameValue.
				call.Arguments = append([]ir.Expression(nil), call.Arguments...)
				for index, argument := range call.Arguments {
					if box, ok := argument.(ir.Box); ok {
						if null, ok := box.Value.(ir.Null); ok {
							call.Arguments[index] = ir.Box{Value: ir.Undefined{Of: null.Of}}
						}
					}
				}
				changed++
				return call
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	if changed == 0 {
		t.Fatal("no exact-null comparisons mutated")
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
	t.Logf("Node alone caught %d erased null comparisons", changed)
}
