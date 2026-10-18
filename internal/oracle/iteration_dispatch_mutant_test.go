package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Mutate the actual lowered protocol, then require Node to catch each change
// independently in both backends. Compilation and sanitizer failures do not count.
func TestIterationDispatchMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, fixture string }{
		{"next reread", "cached_next"}, {"completion value enters body", "class"},
		{"close skipped on break", "break"}, {"close on next throw", "next_throw"},
		{"close throw replaces original", "close_throw_body"}, {"close skipped on outer continue", "labeled_continue"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, pathErr := filepath.Abs(filepath.Join(repository, "stage3/fixtures/iteration-dispatch", probe.fixture+".a"))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			active, iterator := -1, -1
			for index, local := range program.Locals {
				if local.Name == "iterator_close_needed" {
					active = index
				}
				if local.Name == "iterator" {
					iterator = index
				}
			}
			if active < 0 || iterator < 0 {
				t.Fatal("protocol locals missing")
			}
			changed := 0
			var visit func(reflect.Value)
			visit = func(value reflect.Value) {
				if value.Kind() == reflect.Interface && !value.IsNil() {
					node := value.Interface()
					switch node := node.(type) {
					case ir.CallClosure:
						if probe.name == "next reread" {
							if read, ok := node.Closure.(ir.Read); ok && program.Locals[read.Local].Name == "iterator_next" {
								node.Closure = ir.IteratorMethod{Object: ir.Read{Local: iterator, Of: ir.Object}, Name: "next"}
								value.Set(reflect.ValueOf(node))
								changed++
							}
						}
					case ir.If:
						if probe.name == "close skipped on break" {
							if read, ok := node.Condition.(ir.Read); ok && read.Local == active {
								node.Condition = ir.BooleanConstant{Value: false}
								value.Set(reflect.ValueOf(node))
								changed++
							}
						}
						if probe.name == "completion value enters body" {
							if truth, ok := node.Condition.(ir.Truthy); ok {
								if field, ok := truth.Value.(ir.IteratorField); ok && field.Name == "done" {
									node.Then = append([]ir.Statement{ir.WriteLine{Value: ir.IteratorField{Object: field.Object, Name: "value", Of: ir.String}, Stream: ir.Stdout}}, node.Then...)
									value.Set(reflect.ValueOf(node))
									changed++
								}
							}
						}
					case ir.Assign:
						if probe.name == "close on next throw" && node.Local == active {
							if b, ok := node.Value.(ir.BooleanConstant); ok && !b.Value {
								node.Value = ir.BooleanConstant{Value: true}
								value.Set(reflect.ValueOf(node))
								changed++
							}
						}
					case ir.Try:
						if probe.name == "close throw replaces original" && node.HasCatch && node.CatchLocal == -1 && !node.HasFinally {
							value.Set(reflect.ValueOf(ir.Block{Body: node.Body}))
							changed++
						}
					case ir.Continue:
						if probe.name == "close skipped on outer continue" && node.Label == "outer" {
							value.Set(reflect.ValueOf(ir.Block{Body: []ir.Statement{ir.Assign{Local: active, Value: ir.BooleanConstant{Value: false}}, node}}))
							changed++
							return
						}
					}
					copy := reflect.New(value.Elem().Type()).Elem()
					copy.Set(value.Elem())
					visit(copy)
					value.Set(copy)
					return
				}
				switch value.Kind() {
				case reflect.Struct:
					for i := 0; i < value.NumField(); i++ {
						visit(value.Field(i))
					}
				case reflect.Slice:
					for i := 0; i < value.Len(); i++ {
						visit(value.Index(i))
					}
				}
			}
			visit(reflect.ValueOf(&program.Main).Elem())
			visit(reflect.ValueOf(&program.Functions).Elem())
			if changed == 0 {
				t.Fatal("mutant changed nothing")
			}
			expected := onNode(t, path)
			t.Run("javascript", func(t *testing.T) {
				actual := onJavaScriptBackend(t, program)
				if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(expected, actual) != "stdout differs" {
					t.Fatalf("want only Node stdout to catch mutant: %+v", actual)
				}
				t.Log("caught by Node stdout")
			})
			t.Run("native", func(t *testing.T) {
				actual, binary := natively(t, program)
				if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(expected, actual) != "stdout differs" {
					t.Fatalf("want only Node stdout to catch mutant: %+v", actual)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				t.Log("caught by Node stdout")
			})
		})
	}
}
