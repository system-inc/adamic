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

var scout30Families = []string{"predicates", "parse", "number", "math", "string", "object", "json"}

func init() {
	for _, family := range scout30Families {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/scout_30_tsc_" + family + ".a", true, false})
	}
}

func TestScout30FamilyMutants(t *testing.T) {
	t.Parallel()
	for _, family := range scout30Families {
		t.Run(family, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_30_tsc_"+family+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			mutate := func(value ir.Expression) ir.Expression {
				switch v := value.(type) {
				case ir.NumberCall:
					if family == "predicates" && (v.Function == "isFinite" || v.Function == "isNaN") {
						changed++
						return ir.Unary{Operator: ir.Not, Operand: v}
					}
					if (family == "parse" && v.Function == "parseInt") || (family == "number" && v.Function == "convert") {
						changed++
						return ir.Binary{Operator: ir.Add, Left: v, Right: ir.NumberConstant{Value: 1}}
					}
				case ir.MathCall:
					if family == "math" && v.Function == "max" {
						changed++
						v.Function = "min"
						return v
					}
				case ir.NumberToString:
					if family == "string" {
						changed++
						v.Value = ir.Binary{Operator: ir.Add, Left: v.Value, Right: ir.NumberConstant{Value: 1}}
						return v
					}
				case ir.ObjectCall:
					if family == "object" && v.Method == "entries" {
						changed++
						return ir.ArrayReverse{Array: v}
					}
				case ir.JSONStringify:
					if family == "json" {
						changed++
						index := len(program.Strings)
						program.Strings = append(program.Strings, "wrong JSON")
						return ir.StringConstant{Index: index}
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if changed == 0 {
				t.Fatal("mutant missed its family")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant did not finish cleanly: %+v", result)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			expected := onNode(t, path)
			if difference := disagreement(expected, result); difference != "stdout differs" {
				t.Fatalf("native: %q", difference)
			}
			if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "stdout differs" {
				t.Fatalf("JavaScript: %q", difference)
			}
			t.Log("Node alone caught wrong output; native and JavaScript exited cleanly")
		})
	}
}

func TestScout30Refusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, reason string }{
		{"json_specs", "never"},
		{"json_parse", "any"},
		{"object_entries_references", "entries"},
		{"string_diagnostic", "String"},
		{"scanner_parse_concat", "BinaryExpression"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_30_refused/"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: %+v", truth)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want named refusal %q, got %v", probe.reason, err)
			}
			t.Log(err)
		})
	}
}

// This old gap fixture closes when the proven prototype-query branch is merged.
func TestScout30PrototypeObservationMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_30_refused/object_prototype.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	actual, binary := natively(t, program)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
			t.Fatal(difference)
		}
	}
	wrong := len(program.Strings)
	program.Strings = append(program.Strings, "wrong prototype kind")
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.ObjectCall); ok && call.Method == "getPrototypeOf" {
			changed = true
			return ir.StringConstant{Index: wrong}
		}
		return value
	})
	if !changed {
		t.Fatal("no prototype observation changed")
	}
	scout22MutantMatchesOnlyNode(t, path, program)
}
