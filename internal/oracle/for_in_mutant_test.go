package oracle

import (
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Each mutant builds cleanly, exits normally under sanitizers, and is caught by
// Node's stdout. None relies on a compiler warning or sanitizer failure.
func TestForInMutants(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"synthetic_keys", "shared_presence", "missing_write", "numeric_order", "removed_key", "receiver_twice"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_in_runtime.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if name == "removed_key" || name == "receiver_twice" {
				initializers := map[int]ir.Expression{}
				for _, statement := range program.Main {
					if declared, ok := statement.(ir.Declare); ok {
						initializers[declared.Local] = declared.Value
					}
				}
				mutate := func(value ir.Expression) ir.Expression {
					if name == "removed_key" {
						if _, ok := value.(ir.ForInOwn); ok {
							changed = true
							return ir.BooleanConstant{Value: true}
						}
					}
					if name == "receiver_twice" {
						if keys, ok := value.(ir.ObjectKeys); ok && keys.Enumeration {
							if read, ok := keys.Object.(ir.Read); ok {
								if call, ok := initializers[read.Local].(ir.Call); ok && program.Functions[call.Function].Name == "receiver" {
									changed = true
									keys.Object = call
									return keys
								}
							}
						}
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			source := native.C(program)
			switch name {
			case "shared_presence":
				after := strings.ReplaceAll(source, "adamic_for_in_copy(", "adamic_object_copy(")
				changed = after != source
				source = after
			case "missing_write":
				after := regexp.MustCompile(`(?m)^\s+adamic_for_in_write\([^\n]*\);\n`).ReplaceAllString(source, "")
				changed = after != source
				source = after
			case "synthetic_keys", "numeric_order":
				after := strings.ReplaceAll(source, "adamic_for_in_keys((const adamic_heap *)", "adamic_mutant_keys((const adamic_heap *)")
				changed = after != source
				source = after
				body := "if (value != NULL && value->kind == adamic_kind_object) return adamic_class_object_keys((const adamic_object *)value);"
				if name == "numeric_order" {
					body = "if (value != NULL && value->kind == adamic_kind_object) { const adamic_object *object = (const adamic_object *)value; if (object->shape->methods == &adamic_for_in_metadata) { const adamic_array *keys = object->slots[object->shape->count - 1].reference; return adamic_array_slice(keys, 0, (double)keys->length, true); } }"
				}
				source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\nextern const adamic_methods adamic_for_in_metadata;\nadamic_array *adamic_for_in_keys(const adamic_heap *);\nstatic adamic_array *adamic_mutant_keys(const adamic_heap *value) { "+body+" return adamic_for_in_keys(value); }\n", 1)
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			truth := onNode(t, path)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant killed outside stdout: %+v", got)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("mutant catcher: %q", difference)
			}
			if name == "removed_key" || name == "receiver_twice" {
				if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
					t.Fatalf("backend mutant catcher: %q", difference)
				}
			}
			t.Logf("%s caught only by stdout: Node %q; native %q", name, truth.stdout, got.stdout)
		})
	}
}

func TestForInInheritedMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_in_static.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	source = strings.ReplaceAll(source, "adamic_for_in_keys((const adamic_heap *)", "adamic_mutant_keys((const adamic_heap *)")
	source = strings.Replace(source, `#include "adamic.h"`, `#include "adamic.h"`+"\nadamic_array *adamic_for_in_keys(const adamic_heap *);\nstatic adamic_array *adamic_mutant_keys(const adamic_heap *value) { if (value != NULL && value->kind == adamic_kind_object) return adamic_class_object_keys((const adamic_object *)value); return adamic_for_in_keys(value); }\n", 1)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	truth := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("inherited key mutant was not caught only by stdout: %+v", got)
	}
	t.Logf("Node %q; own-only mutant %q", truth.stdout, got.stdout)
}

func TestForInScalarBoxMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/for_in_primitives.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	after := strings.ReplaceAll(source, "adamic_box_number(", "adamic_string_from_number(")
	if after == source {
		t.Fatal("mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(after, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	truth := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("scalar box mutant catcher: %+v", got)
	}
	t.Logf("Node %q; string box mutant %q", truth.stdout, got.stdout)
}
