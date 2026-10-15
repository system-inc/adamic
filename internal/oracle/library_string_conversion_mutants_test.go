package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStringArrayConversionNodeMutant(t *testing.T) {
	t.Parallel()
	stringConversionMutant(t, "collections", "array")
}
func TestStringMapConversionNodeMutant(t *testing.T) {
	t.Parallel()
	stringConversionMutant(t, "collections", "map")
}
func TestStringSetConversionNodeMutant(t *testing.T) {
	t.Parallel()
	stringConversionMutant(t, "collections", "set")
}
func TestStringTemplateConversionNodeMutant(t *testing.T) {
	t.Parallel()
	stringConversionMutant(t, "template", "template")
}

func TestStringCollectionEffectsNodeMutant(t *testing.T) {
	t.Parallel()
	stringConversionMutant(t, "collections", "effects")
}

// Each mutant stays valid, exits zero and remains sanitizer- and leak-clean.
func stringConversionMutant(t *testing.T, fixture, family string) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_conversion", fixture+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutate := func(value ir.Expression) ir.Expression {
		if changed {
			return value
		}
		switch node := value.(type) {
		case ir.Call:
			if family == "effects" && program.Functions[node.Function].Name == "library_string_map_primitive" && len(node.Arguments) == 1 {
				if _, called := node.Arguments[0].(ir.Call); called {
					argument := node.Arguments[0]
					node.Arguments[0] = ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: argument}}, Result: argument}
					changed = true
					return node
				}
			}
		case ir.Concat:
			if family == "template" && len(node.Parts) == 4 {
				if first, literal := node.Parts[0].(ir.StringConstant); literal && program.Strings[first.Index] == "before " {
					node.Parts[1], node.Parts[3] = node.Parts[3], node.Parts[1]
					changed = true
					return node
				}
			}
		case ir.ArrayJoin:
			if family == "array" {
				node.Separator = ir.StringConstant{Index: len(program.Strings)}
				program.Strings = append(program.Strings, ";")
				changed = true
				return node
			}
		case ir.StringConstant:
			if (family == "map" && program.Strings[node.Index] == "[object Map]") || (family == "set" && program.Strings[node.Index] == "[object Set]") {
				node.Index = len(program.Strings)
				program.Strings = append(program.Strings, "[object Wrong]")
				changed = true
				return node
			}
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
	if !changed {
		t.Fatal("mutant missed its operation")
	}
	result, binary := natively(t, program)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
		t.Fatalf("want only Node stdout comparison, got %q", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("caught only by stdout comparison with source on Node")
}
