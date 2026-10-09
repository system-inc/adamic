package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// The five readers are copied byte-for-byte from integration/review-lane-fxspptb
// e23d33d1460cfe48e45d46e5d4399b97be92024e, review/fxspptb/4e1649c_generic_*.a.
func init() {
	for _, name := range []string{"receiver", "receiver_charat", "receiver_number", "pad_fill", "pad_fill_call", "errors"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_string_generic_" + name + ".a", true, false})
	}
}

// Each reader and backend is an independent shard. Mutants keep valid IR, finish cleanly, and
// have no sanitizer findings; only comparison against Node catches them.
func TestStringGenericMutants(t *testing.T) {
	for _, name := range []string{"receiver", "receiver_charat", "receiver_number", "pad_fill", "pad_fill_call", "errors"} {
		t.Run(name, func(t *testing.T) {
			for _, backend := range []string{"native", "JavaScript", "WASI"} {
				t.Run(backend, func(t *testing.T) {
					t.Parallel()
					if backend == "WASI" && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
						t.Skip("set ADAMIC_ORACLE_WASI=1 to run the WASI mutant")
					}
					path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_generic_"+name+".a"))
					if err != nil {
						t.Fatal(err)
					}
					program, err := lowered(t, path)
					if err != nil {
						t.Fatal(err)
					}
					changed := 0
					if strings.HasPrefix(name, "receiver") {
						for i := range program.Functions {
							function := &program.Functions[i]
							if function.Name != "library_string_prototype_primitive" {
								continue
							}
							for j, statement := range function.Body {
								if guard, ok := statement.(ir.If); ok {
									guard.Condition = ir.BooleanConstant{Value: false}
									function.Body[j] = guard
									changed++
								}
							}
						}
					} else if name == "errors" {
						for index, text := range program.Strings {
							if text == "String.prototype.trimLeft called on null or undefined" {
								program.Strings[index] = "String.prototype.trimStart called on null or undefined"
								changed++
							}
						}
					} else {
						program.Strings = append(program.Strings, "_")
						replacement := ir.StringConstant{Index: len(program.Strings) - 1}
						mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
							if coalesce, ok := value.(ir.Coalesce); ok {
								if fallback, ok := coalesce.Fallback.(ir.StringConstant); ok && program.Strings[fallback.Index] == " " {
									coalesce.Fallback = replacement
									changed++
									return coalesce
								}
							}
							return value
						})
					}
					if changed == 0 {
						t.Fatal("mutant missed its target")
					}
					expected := onNode(t, path)
					var actual run
					switch backend {
					case "native":
						var binary string
						actual, binary = natively(t, program)
						if actual.exitCode != 0 || len(actual.stderr) != 0 {
							t.Fatalf("mutant must finish cleanly: %+v", actual)
						}
						if report := leaks(t, program, binary); report != "" {
							t.Fatal(report)
						}
					case "JavaScript":
						actual = onJavaScriptBackend(t, program)
					case "WASI":
						actual = onWASI(t, native.C(program))
					}
					if difference := disagreement(expected, actual); difference != "stdout differs" {
						t.Fatalf("%s: %q", backend, difference)
					}
					t.Logf("Node alone caught %d mutated operations on %s", changed, backend)
				})
			}
		})
	}
}
