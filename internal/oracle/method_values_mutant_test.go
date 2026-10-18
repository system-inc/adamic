package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Each behavioral mutant compiles and finishes without sanitizer findings. Node's source result,
// rather than an expected string copied out of the implementation, must kill it.
// libraryMethodValueMutants is the mutants of library_method_values.a, by name, the key TestLibraryMethodValueMutants'
// family (families_test.go) owns each by.
func libraryMethodValueMutants() []string {
	return []string{"search value", "apply bound", "last default", "NaN includes", "trim", "substring bounds", "charAt missing", "object tag", "own field", "map String", "parseInt radix", "parseFloat", "fixed digits", "numeric radix", "Math receiver", "character code", "bound receiver", "null tag", "ignored this evaluation", "map optional String"}
}

// TestLibraryMethodValueMutants is a family (families_test.go): TestLibraryMethodValueMutants_000 and on each run the
// mutants their shard owns, each built and run, and caught only by stdout against the source on Node.
func runLibraryMethodValueMutantsShard(t *testing.T, shard int) {
	t.Helper()
	planted, probing := familyProbe("TestLibraryMethodValueMutants")
	names := libraryMethodValueMutants()
	for _, index := range familyCases(t, "TestLibraryMethodValueMutants", shard, names) {
		name := names[index]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_method_values.a"))
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
					if name == "ignored this evaluation" && strings.HasPrefix(program.Functions[node.Function].Name, "library_method_sequence") && len(node.Arguments) == 3 {
						if node.Arguments[1].Type() == ir.Array {
							node.Arguments[1] = ir.ArrayLiteral{Element: ir.Number}
							changed = true
							return node
						}
					}
					if name == "bound receiver" && strings.HasPrefix(program.Functions[node.Function].Name, "library_bind_factory") {
						node.Arguments[0] = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, "changed")
						changed = true
						return node
					}
				case ir.ArraySearch:
					switch {
					case name == "search value" && !node.Last && !node.Includes && node.From == nil:
						node.Value = ir.NumberConstant{Value: 42}
						changed = true
						return node
					case name == "apply bound" && !node.Last && !node.Includes && node.From != nil:
						node.From = ir.NumberConstant{Value: 0}
						changed = true
						return node
					case name == "last default" && node.Last && node.From == nil:
						node.From = ir.NumberConstant{Value: 0}
						changed = true
						return node
					case name == "NaN includes" && node.Includes:
						node.Value = ir.NumberConstant{Value: 0}
						changed = true
						return node
					}
				case ir.Trim:
					if name == "trim" {
						changed = true
						return ir.StringCall{Method: "trimStart", Value: node.Value}
					}
				case ir.MathCall:
					if name == "substring bounds" && node.Function == "max" {
						node.Function = "min"
						changed = true
						return node
					}
					if name == "Math receiver" && node.Function == "abs" {
						node.Arguments = []ir.Expression{ir.NumberConstant{Value: 0}}
						changed = true
						return node
					}
				case ir.Coalesce:
					if name == "map optional String" && node.Type() == ir.String {
						if _, read := node.Value.(ir.Read); read {
							changed = true
							return node.Value
						}
					}
					if name == "charAt missing" {
						if _, yes := node.Value.(ir.StringIndex); yes {
							node.Fallback = ir.StringConstant{Index: len(program.Strings)}
							program.Strings = append(program.Strings, "undefined")
							changed = true
							return node
						}
					}
				case ir.StringConstant:
					if name == "null tag" && program.Strings[node.Index] == "[object Null]" {
						node.Index = len(program.Strings)
						program.Strings = append(program.Strings, "[object Object]")
						changed = true
						return node
					}
					if name == "object tag" && program.Strings[node.Index] == "[object Array]" {
						node.Index = len(program.Strings)
						program.Strings = append(program.Strings, "[object Object]")
						changed = true
						return node
					}
				case ir.HasOwn:
					if name == "own field" {
						changed = true
						// Records specialize some known own fields before this walk. Invert
						// the remaining query, whether its source answer is true or false.
						return ir.Unary{Operator: ir.Not, Operand: node}
					}
				case ir.NumberToString:
					if _, read := node.Value.(ir.Read); name == "map String" && read {
						node.Value = ir.Unary{Operator: ir.Negate, Operand: node.Value}
						changed = true
						return node
					}
				case ir.NumberCall:
					if name == "parseInt radix" && node.Function == "parseInt" && len(node.Arguments) == 2 {
						if _, read := node.Arguments[1].(ir.Read); read {
							node.Arguments[1] = ir.NumberConstant{Value: 10}
							changed = true
							return node
						}
					}
					if name == "parseFloat" && node.Function == "parseFloat" {
						node.Function = "parseInt"
						changed = true
						return node
					}
				case ir.ToFixed:
					if name == "fixed digits" {
						node.Digits = ir.NumberConstant{Value: 0}
						changed = true
						return node
					}
				case ir.NumberFormat:
					if name == "numeric radix" && node.Method == "toString" {
						node.Argument = ir.NumberConstant{Value: 10}
						changed = true
						return node
					}
				case ir.StringFromCodes:
					if name == "character code" && len(node.Codes) > 0 {
						node.Codes[0] = ir.NumberConstant{Value: 90}
						changed = true
						return node
					}
				}
				return value
			}
			switch {
			case probing && name == planted:
				// TestFamiliesCatchAPlantedCase's mutant: this mutant withheld, so it changes nothing.
			case name == "map String" || name == "map optional String":
				for index := range program.Functions {
					if program.Functions[index].Name == "library_map_String" {
						mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
					}
				}
			default:
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 {
				t.Fatalf("mutant must finish, not fail compilation or sanitizers: exit %d stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node stdout alone to kill mutant, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("behavior mutant leaked: %s", report)
			}
			t.Log("caught only by stdout comparison with source on Node")
		})
	}
}

// Restoring an own-field load for a supported intrinsic must recreate the original missing-field
// panic, never slip past the source oracle as a callable value.
func TestLibraryMethodOwnLoadMutants(t *testing.T) {
	t.Parallel()
	for _, localName := range []string{"search", "trim", "own", "parse", "abs", "fixed", "character", "convert"} {
		t.Run(localName, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_method_values.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index, statement := range program.Main {
				if declaration, yes := statement.(ir.Declare); yes && program.Locals[declaration.Local].Name == localName {
					declaration.Value = ir.Property{Object: ir.ObjectLiteral{}, Name: localName, Of: ir.Closure}
					program.Main[index] = declaration
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			result, _ := natively(t, program)
			if result.exitCode != 70 || !strings.Contains(string(result.stderr), "compiler bug: a field the checker proved is there is missing") {
				t.Fatalf("want the own-field panic, got exit %d stderr %s", result.exitCode, result.stderr)
			}
			if disagreement(onNode(t, path), result) == "" {
				t.Fatal("source oracle accepted an own-field load of an intrinsic")
			}
			t.Log("caught by the source oracle's exit and stderr comparison")
		})
	}
}
