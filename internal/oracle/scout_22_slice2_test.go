package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var scout22Slice2Fixtures = []string{"scout22_conversion_hints", "scout22_conversion_errors", "scout22_range_errors", "scout22_property_errors", "scout22_tsc_to_string", "scout22_spread_errors"}

func init() {
	for _, name := range scout22Slice2Fixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}
func TestScout22Slice2Fixtures(t *testing.T) {
	for _, name := range scout22Slice2Fixtures {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			result, binary := natively(t, program)
			if d := disagreement(expected, result); d != "" {
				t.Fatalf("native: %s\nNode %+v\nnative %+v", d, expected, result)
			}
			if d := disagreement(expected, onJavaScriptBackend(t, program)); d != "" {
				t.Fatalf("JavaScript: %s", d)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if d := disagreement(expected, onWASI(t, native.C(program))); d != "" {
					t.Fatalf("WASI: %s", d)
				}
			}
		})
	}
}

// Mutations remain well typed and terminate without diagnostics or leaks. The
// source Node execution is the independent witness of the semantic difference.
func TestScout22Slice2Mutants(t *testing.T) {
	for _, family := range []string{"conversion_hints", "conversion_errors", "property_errors", "range_errors", "padStart", "normalize", "toFixed", "toExponential", "toPrecision", "toString", "error_identity", "spread_errors"} {
		t.Run(family, func(t *testing.T) {
			fixture := family
			switch family {
			case "padStart", "normalize", "toFixed", "toExponential", "toPrecision", "toString":
				fixture = "range_errors"
			case "error_identity":
				fixture = "conversion_errors"
			}
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_"+fixture+".a"))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			stringHint := -1
			for i, function := range program.Functions {
				if strings.Contains(function.Name, "ordinary_string") {
					stringHint = i
					break
				}
			}
			wrong := len(program.Strings)
			program.Strings = append(program.Strings, "Wrong primitive conversion")
			mutate := func(value ir.Expression) ir.Expression {
				if changed != 0 {
					return value
				}
				switch family {
				case "conversion_hints":
					if call, ok := value.(ir.Call); ok && strings.Contains(program.Functions[call.Function].Name, "ordinary_default") && stringHint >= 0 {
						call.Function = stringHint
						changed++
						return call
					}
				case "conversion_errors":
					if call, ok := value.(ir.MakeError); ok {
						call.Message = ir.StringConstant{Index: wrong}
						changed++
						return call
					}
				case "property_errors":
					if call, ok := value.(ir.ObjectCall); ok && call.Method == "freeze" {
						call.Method = "preventExtensions"
						changed++
						return call
					}
				case "error_identity":
					if call, ok := value.(ir.ObjectCall); ok && call.Method == "errorIsType" {
						changed++
						return ir.BooleanConstant{Value: false}
					}
				case "padStart", "normalize":
					if call, ok := value.(ir.StringCall); ok && call.Method == family {
						if family == "normalize" {
							index := len(program.Strings)
							program.Strings = append(program.Strings, "NFC")
							call.Arguments[0] = ir.StringConstant{Index: index}
						} else {
							call.Arguments[0] = ir.NumberConstant{Value: 0}
						}
						changed++
						return call
					}
				case "toFixed":
					if call, ok := value.(ir.ToFixed); ok {
						call.Digits = ir.NumberConstant{Value: 0}
						changed++
						return call
					}
				case "toExponential", "toPrecision", "toString":
					if call, ok := value.(ir.NumberFormat); ok && call.Method == family {
						argument := 1.0
						if family == "toString" {
							argument = 10
						}
						call.Argument = ir.NumberConstant{Value: argument}
						changed++
						return call
					}
				case "range_errors", "spread_errors":
					if call, ok := value.(ir.StringCall); ok && call.Method == "repeat" && len(call.Arguments) > 0 {
						// Skipping the negative-count error exercises the catch continuation.
						if count, constant := call.Arguments[0].(ir.NumberConstant); !constant || count.Value < 0 {
							call.Arguments[0] = ir.NumberConstant{Value: 0}
							changed++
							return call
						}
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if changed != 1 {
				t.Fatalf("want one mutation, got %d", changed)
			}
			scout22MutantMatchesOnlyNode(t, path, program)
		})
	}
}
