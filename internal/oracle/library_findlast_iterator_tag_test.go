package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func init() {
	for _, name := range []string{"57f2d04_find_last_shrinks", "57f2d04_find_last_shrinks2", "903f25b_iterator_tag", "library_iterator_tags", "library_find_last_missing"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/" + name + ".a", lowers: true})
	}
}

// All mutated executions finish cleanly and agree with each other. Only the
// independent source-on-Node stdout comparison distinguishes the wrong behavior.
func TestLibraryFindLastIteratorTagMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, fixture, method string }{
		{"findLastIndex skips removed indices", "57f2d04_find_last_shrinks.a", "findLastIndex"},
		{"findLast skips removed indices", "57f2d04_find_last_shrinks2.a", "findLast"},
		{"iterator ignores inherited tag", "library_iterator_tags.a", ""},
	} {
		t.Run(one.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", one.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			if one.method != "" {
				for i := range program.Functions {
					function := &program.Functions[i]
					if function.Name != "array_"+one.method {
						continue
					}
					for j, statement := range function.Body {
						loop, ok := statement.(ir.Loop)
						if !ok {
							continue
						}
						index := loop.Condition.(ir.Binary).Left
						array := ir.Read{Local: function.Parameters[0], Of: ir.Array}
						loop.Body = []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Less, Left: index, Right: ir.Length{Array: array}}, Then: loop.Body}}
						function.Body[j] = loop
						changed++
					}
				}
			} else {
				object := len(program.Strings)
				program.Strings = append(program.Strings, "Object")
				replace := func(value ir.Expression) ir.Expression {
					if call, ok := value.(ir.CallClosure); ok {
						if property, ok := call.Closure.(ir.Property); ok && property.Name == "__adamic_iterator_tag" {
							changed++
							return ir.StringConstant{Index: object}
						}
					}
					return value
				}
				program.Main = rewriteTagMutant(reflect.ValueOf(program.Main), replace).Interface().([]ir.Statement)
				for i := range program.Functions {
					program.Functions[i].Body = rewriteTagMutant(reflect.ValueOf(program.Functions[i].Body), replace).Interface().([]ir.Statement)
				}
			}
			if changed == 0 {
				t.Fatal("mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1"}
			}
			got := executeWith(t, environment, binary)
			backend := onJavaScriptBackend(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 || backend.exitCode != 0 || len(backend.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: native %+v, backend %+v", got, backend)
			}
			if difference := disagreement(got, backend); difference != "" {
				t.Fatalf("mutant backends disagree: %s", difference)
			}
			truth := onNode(t, path)
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("Node caught mutant by %q", difference)
			}
			t.Log("both mutant backends agree; source Node alone catches stdout; clean exit, ASan, UBSan and LSan")
		})
	}
}

// Only rewrite the fixture's statement/expression trees, never analysis metadata.
func rewriteTagMutant(value reflect.Value, replace func(ir.Expression) ir.Expression) reflect.Value {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return value
		}
		result := reflect.New(value.Type()).Elem()
		if expression, ok := value.Interface().(ir.Expression); ok {
			changed := replace(expression)
			if !reflect.DeepEqual(expression, changed) {
				result.Set(reflect.ValueOf(changed))
				return result
			}
		}
		result.Set(rewriteTagMutant(value.Elem(), replace))
		return result
	}
	switch value.Kind() {
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		result.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if result.Field(i).CanSet() {
				result.Field(i).Set(rewriteTagMutant(value.Field(i), replace))
			}
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(rewriteTagMutant(value.Index(i), replace))
		}
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return value
		}
		result := reflect.New(value.Type().Elem())
		result.Elem().Set(rewriteTagMutant(value.Elem(), replace))
		return result
	}
	return value
}
