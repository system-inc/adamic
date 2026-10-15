package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestStringStartsWithPositionNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "startsWith", "library_string_affixes")
}
func TestStringEndsWithPositionNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "endsWith", "library_string_affixes")
}
func TestStringSubstrNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "substr", "library_tsc_census_substr")
}
func TestStringRepeatNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "repeat", "library_string_tail_repeat")
}
func TestStringLastIndexPositionNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "string_last_index_of_position", "string_lastindexof_position")
}
func TestStringRepeatSafeBoundNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "safe_repeat", "string_lastindexof_position")
}

func TestStringRawEmptyNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "raw_empty", "library_string_raw_empty")
}
func TestStringRawSnapshotNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "raw", "library_string_raw_empty")
}
func TestStringPrimitiveIdentityNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "identity", "library_string_primitive_identity")
}
func TestStringNullReceiverNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "null_receiver", "library_string_null_receiver")
}
func TestStringGenericReceiverNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "prototype_primitive", "library_string_generic_receiver")
}
func TestStringGenericPadDefaultNodeMutant(t *testing.T) {
	t.Parallel()
	stringIndicesMutant(t, "padEnd", "library_string_generic_pad_fill_call")
}

// All mutations operate on ordinary typed IR. The compiler runs in this test
// process; no Go builder or overlay subprocess is needed.
func stringIndicesMutant(t *testing.T, family, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_indices", fixture+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	constant := func(text string) ir.StringConstant {
		i := len(program.Strings)
		program.Strings = append(program.Strings, text)
		return ir.StringConstant{Index: i}
	}
	for i := range program.Functions {
		f := &program.Functions[i]
		match := f.Name == "library_string_"+family || f.Name == family
		if family == "padEnd" || family == "safe_repeat" {
			match = true
		}
		if !match || changed {
			continue
		}
		if family == "identity" || family == "raw_empty" {
			f.Body = []ir.Statement{ir.Return{Value: constant("mutant")}}
			changed = true
			continue
		}
		if family == "prototype_primitive" {
			for j, statement := range f.Body {
				if guard, ok := statement.(ir.If); ok {
					guard.Then = []ir.Statement{ir.Return{Value: constant("mutant")}}
					f.Body[j] = guard
					changed = true
					break
				}
			}
			continue
		}
		mutate := func(value ir.Expression) ir.Expression {
			if changed {
				return value
			}
			switch node := value.(type) {
			case ir.StringCall:
				if family == "safe_repeat" && node.Method == "repeat" {
					node.Arguments[0] = ir.NumberConstant{Value: 1}
					changed = true
					return node
				}
				if node.Method == "slice" && (family == "startsWith" || family == "endsWith" || family == "substr" || family == "string_last_index_of_position") {
					index := 0
					if family == "endsWith" || family == "string_last_index_of_position" {
						index = 1
					}
					node.Arguments[index] = ir.NumberConstant{Value: 0}
					changed = true
					return node
				}
			case ir.Concat:
				if family == "repeat" && len(node.Parts) == 2 {
					if read, ok := node.Parts[0].(ir.Read); ok && program.Locals[read.Local].Name == "power" {
						node.Parts = node.Parts[:1]
						changed = true
						return node
					}
				}
			case ir.StringConstant:
				if family == "null_receiver" && program.Strings[node.Index] == "String.prototype.slice called on null or undefined" {
					changed = true
					return constant("wrong receiver error")
				}
			case ir.Length:
				if family == "raw" {
					changed = true
					return ir.NumberConstant{Value: 1}
				}
			case ir.Coalesce:
				if family == "padEnd" || family == "safe_repeat" {
					if fallback, ok := node.Fallback.(ir.StringConstant); ok && program.Strings[fallback.Index] == " " {
						node.Fallback = constant("_")
						changed = true
						return node
					}
				}
			}
			return value
		}
		mutateStringExpressions(reflect.ValueOf(&f.Body).Elem(), mutate)
	}
	if family == "safe_repeat" && !changed {
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
			if node, ok := value.(ir.StringCall); ok && node.Method == "repeat" && !changed {
				node.Arguments[0] = ir.NumberConstant{Value: 1}
				changed = true
				return node
			}
			return value
		})
	}
	if !changed {
		t.Fatal("mutant changed no expression")
	}
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: %#v", got)
	}
	if delta := disagreement(onNode(t, path), got); delta != "stdout differs" {
		t.Fatalf("want only Node stdout to kill mutant, got %q", delta)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("caught only by Node stdout; clean exit and ASan, UBSan, LSan")
}
