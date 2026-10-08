package oracle

import (
	"os"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Rewrites executable index reads, leaving operand and contract metadata intact.
func changeTupleArrayRead(program *ir.Program, change func(ir.ArrayIndex) ir.ArrayIndex) int {
	count := 0
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			out := reflect.New(value.Type()).Elem()
			out.Set(rewrite(value.Elem()))
			return out
		case reflect.Struct:
			out := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				out.Field(index).Set(rewrite(value.Field(index)))
			}
			if read, ok := out.Interface().(ir.ArrayIndex); ok && read.TupleUnion {
				count++
				out.Set(reflect.ValueOf(change(read)))
			}
			return out
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			out := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for index := 0; index < value.Len(); index++ {
				out.Index(index).Set(rewrite(value.Index(index)))
			}
			return out
		default:
			return value
		}
	}
	for index := range program.Functions {
		program.Functions[index].Body = rewrite(reflect.ValueOf(program.Functions[index].Body)).Interface().([]ir.Statement)
	}
	program.Main = rewrite(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	return count
}

func TestCheckedViewTupleOriginalRootIndex(t *testing.T) {
	root, _ := tupleOriginalInputs(t)
	for _, sample := range []struct {
		name, stdout string
		required     bool
	}{
		{"root-index-tuple", "object\n", false},
		{"root-index-scalar", "number\n", false},
		{"root-index-bounds", "undefined\n", false},
		{"root-index-required-missing", "undefined\n", true},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path, program := tupleOriginalProgram(t, root, sample.name)
			if difference := disagreement(run{stdout: []byte(sample.stdout)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			// The frontend's ordinary index is optional. Exercise the required
			// consumer promise on that same generated read, without a second path.
			if count := changeTupleArrayRead(program, func(read ir.ArrayIndex) ir.ArrayIndex {
				read.Required = sample.required && os.Getenv("ADAMIC_TUPLE_INDEX_MUTANT") != "presence"
				return read
			}); count != 1 {
				t.Fatalf("wanted one tuple-union index, found %d", count)
			}
			if os.Getenv("ADAMIC_TUPLE_INDEX_MUTANT") == "scalar" && sample.name == "root-index-scalar" {
				changed := false
				for index := range program.ViewContracts {
					c := &program.ViewContracts[index]
					if c.Kind == ir.ViewScalar && c.Of == ir.Number && c.Name == "IncrementalBuildInfoFileId" {
						c.Kind, c.Of, c.FixedTuple = ir.ViewObject, ir.Object, true
						c.Tuple = []ir.ViewContractID{1, 1}
						changed = true
					}
				}
				if !changed {
					t.Fatal("scalar member mutant changed no contract")
				}
			}
			want := run{stdout: []byte(sample.stdout)}
			if sample.required {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: element read failed: values[0] expected IncrementalBuildInfoRoot, found undefined\n")}
			}
			if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
				t.Error("JavaScript: " + difference)
			}
			actual, binary := nativelyUncached(t, program)
			if difference := disagreement(want, actual); difference != "" {
				t.Errorf("sanitized: %s; %#v", difference, actual)
			}
			if difference := disagreement(want, releasedUncached(t, program)); difference != "" {
				t.Error("release: " + difference)
			}
			if !sample.required && !t.Failed() {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
