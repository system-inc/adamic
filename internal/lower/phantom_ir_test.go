package lower

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Checker IDs and nominal assignability are metadata, not executable casts.
// Keep the original programs intact and verify their call/count facts separately.
func phantomExecutableIRMatches(t *testing.T, got, want *ir.Program) bool {
	t.Helper()
	for _, program := range []*ir.Program{got, want} {
		if len(program.FunctionTypeTargets) == 0 {
			t.Fatal("missing function-type target evidence")
		}
		covered := map[int]bool{}
		for id, targets := range program.FunctionTypeTargets {
			if id <= 0 || len(targets) == 0 {
				t.Fatalf("invalid function-type evidence %d: %v", id, targets)
			}
			seen := map[int]bool{}
			for _, target := range targets {
				if target < 0 || target >= len(program.Functions) || seen[target] {
					t.Fatalf("invalid target set %d: %v", id, targets)
				}
				seen[target], covered[target] = true, true
				function := program.Functions[target]
				if function.ReadsArguments || function.ArgumentsCount != 0 || program.PackedCountNeeded(target) {
					t.Fatalf("phantom erasure changed the no-count fact for %s", function.Name)
				}
			}
		}
		for index, function := range program.Functions {
			if !covered[index] {
				t.Fatalf("missing target evidence for %s", function.Name)
			}
		}
		var walk func(reflect.Value)
		walk = func(value reflect.Value) {
			switch value.Kind() {
			case reflect.Pointer, reflect.Interface:
				if !value.IsNil() {
					walk(value.Elem())
				}
			case reflect.Struct:
				if call, ok := value.Interface().(ir.Call); ok {
					targets := program.CallTargets(call)
					if call.Virtual != 0 || !reflect.DeepEqual(targets, []int{call.Function}) || len(call.Arguments) != len(program.Functions[call.Function].Parameters) {
						t.Fatalf("phantom erasure changed direct call facts: %#v targets %v", call, targets)
					}
				}
				for i := 0; i < value.NumField(); i++ {
					walk(value.Field(i))
				}
			case reflect.Slice, reflect.Array:
				for i := 0; i < value.Len(); i++ {
					walk(value.Index(i))
				}
			}
		}
		walk(reflect.ValueOf(program.Main))
		for _, function := range program.Functions {
			walk(reflect.ValueOf(function.Body))
		}
	}
	executableGot, executableWant := *got, *want
	executableGot.FunctionTypeTargets, executableWant.FunctionTypeTargets = nil, nil
	// These witnesses have no graph regions. Their allocation-site checker IDs
	// cannot affect executable ownership, but remain present in the original IR.
	if len(got.GraphTypes) != 0 || len(want.GraphTypes) != 0 {
		t.Fatal("phantom witness gained a graph region")
	}
	return reflect.DeepEqual(phantomExecutableNode(reflect.ValueOf(executableGot)), phantomExecutableNode(reflect.ValueOf(executableWant)))
}

func phantomExecutableNode(value reflect.Value) any {
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return phantomExecutableNode(value.Elem())
	case reflect.Struct:
		result := map[string]any{"type": value.Type().String()}
		for i := 0; i < value.NumField(); i++ {
			name := value.Type().Field(i).Name
			if name != "GraphTypes" {
				result[name] = phantomExecutableNode(value.Field(i))
			}
		}
		return result
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return nil
		}
		result := make([]any, value.Len())
		for i := range result {
			result[i] = phantomExecutableNode(value.Index(i))
		}
		return result
	default:
		return value.Interface()
	}
}
