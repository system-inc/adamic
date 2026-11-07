package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/unknown_narrowing.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/unknown_narrowing_host.a", true, false})
}

// The source on Node remains unchanged. These mutants alter the lowered program, so an
// agreement between Adamic's two backends cannot hide a missing runtime guard.
func TestUnknownNarrowingMutants(test *testing.T) {
	test.Parallel()
	for _, mutant := range []string{"in_always_true", "skip_inner_typeof"} {
		test.Run(mutant, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/unknown_narrowing.a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			changed := 0
			mutateUnknown(reflect.ValueOf(program).Elem(), program, mutant, &changed)
			if changed == 0 {
				test.Fatal("the mutant changed no runtime guard")
			}
			truth := onNode(test, path)
			backend := onJavaScriptBackend(test, program)
			if difference := disagreement(truth, backend); difference != "stdout differs" {
				test.Fatalf("want Node stdout to catch %s, got %q", mutant, difference)
			}
			test.Logf("%s: Node stdout %q; mutated JavaScript stdout %q", mutant, truth.stdout, backend.stdout)
			if mutant == "in_always_true" {
				native, _ := natively(test, program)
				if difference := disagreement(truth, native); difference != "stdout differs" {
					test.Fatalf("want absent-key output to catch native in mutant, got %q", difference)
				}
				test.Logf("mutated native stdout %q", native.stdout)
			}
		})
	}
}

func mutateUnknown(value reflect.Value, program *ir.Program, mutant string, changed *int) {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return
		}
		if expression, ok := value.Interface().(ir.Expression); ok {
			replace := false
			if _, presence := expression.(ir.HasProperty); presence && mutant == "in_always_true" {
				replace = true
			}
			if binary, equal := expression.(ir.Binary); equal &&
				mutant == "skip_inner_typeof" &&
				binary.Operator == ir.Equal {
				if observed, typeof := binary.Left.(ir.TypeOf); typeof {
					if _, dynamic := observed.Value.(ir.DynamicProperty); dynamic {
						if text, constant := binary.Right.(ir.StringConstant); constant &&
							program.Strings[text.Index] == "string" {
							replace = true
						}
					}
				}
			}
			if replace {
				value.Set(reflect.ValueOf(ir.BooleanConstant{Value: true}))
				*changed++
				return
			}
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		mutateUnknown(copy, program, mutant, changed)
		value.Set(copy)
		return
	}
	switch value.Kind() {
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			mutateUnknown(value.Field(index), program, mutant, changed)
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			mutateUnknown(value.Index(index), program, mutant, changed)
		}
	}
}
