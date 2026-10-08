package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"proven_record", "proven_index", "mixed", "assign_record", "proven_names"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/object_reflection_ruling/" + name + ".a", true, false})
	}
}

var objectReflectionFixtures = []struct {
	name    string
	checked bool
	node    string
}{
	{"scanner", false, "1\n"}, {"hidden", true, "before\n2\n"}, {"hidden_control", false, "before\n1\n"},
	{"integers", false, "0:0\n2:2\n10:10\n4294967294:4\nfirst:1\n01:3\n4294967295:5\n-0:6\n"},
	{"assign_extra", true, "before\n2\n"}, {"assign_hidden", true, "before\n2\n"}, {"assign_control", false, "before\n2\n"},
	{"proven_record", false, "2:2\n10:10\nfirst:1\n"},
	{"proven_index", false, "2:2\n10:10\nfirst:1\n"},
	{"mixed", false, "text\nnumber\nboolean\n3\n"},
	{"assign_record", false, "2\n"},
	{"proven_names", false, "1:3\n#field:1\n1\x00key:2\n"},
	{"nul_keys", false, "1:3\nx:1\n1\x00notIndex:2\n\x00key:4\n"},
	{"hidden_boolean", true, "before\n2\n"},
	{"boolean_layout", true, "2\nbefore\n2\n"},
	{"hidden_literal", true, "before\n2\n"},
	{"nested_entries", false, "2\n"},
}

// Fixtures stay .a in the repository. A temporary .ts copy exercises the source policy.
func reflectionFixture(t *testing.T, name string) (string, *ir.Program) {
	t.Helper()
	text, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/object_reflection_ruling", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+".ts")
	if err = os.WriteFile(path, text, 0644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	if program.ReflectionChecks < 1 {
		t.Fatal("missing visible reflection check count")
	}
	return path, program
}

func TestObjectReflectionRuling(t *testing.T) {
	for _, fixture := range objectReflectionFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			path, program := reflectionFixture(t, fixture.name)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(fixture.node)}, node); difference != "" {
				t.Fatal("Node source: " + difference)
			}
			backend := onJavaScriptBackend(t, program)
			native, binary := nativelyUncached(t, program)
			if difference := disagreement(backend, native); difference != "" {
				t.Fatalf("backend comparison: %s; native %#v JS %#v", difference, native, backend)
			}
			if fixture.checked {
				if native.exitCode != 70 || !strings.Contains(string(native.stderr), "slot check failed") || string(native.stdout) != strings.TrimSuffix(fixture.node, "2\n") {
					t.Fatalf("check failed to stop: %#v", native)
				}
			} else {
				if difference := disagreement(node, native); difference != "" {
					t.Fatalf("Node comparison: %s; native %#v", difference, native)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func mutateReflection(value reflect.Value, mutate func(ir.ObjectCall) ir.ObjectCall) {
	if !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return
		}
		if call, ok := value.Interface().(ir.ObjectCall); ok {
			value.Set(reflect.ValueOf(mutate(call)))
			return
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		mutateReflection(copy, mutate)
		value.Set(copy)
		return
	}
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			mutateReflection(value.Field(i), mutate)
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			mutateReflection(value.Index(i), mutate)
		}
	}
}

func TestObjectReflectionRulingMutants(t *testing.T) {
	for _, rule := range []string{"fold check", "fold assign check", "layout order", "literal member"} {
		t.Run(rule, func(t *testing.T) {
			name := "hidden"
			if rule == "fold assign check" {
				name = "assign_hidden"
			}
			if rule == "layout order" {
				name = "integers"
			}
			if rule == "literal member" {
				name = "hidden_literal"
			}
			path, program := reflectionFixture(t, name)
			changed := 0
			mutate := func(call ir.ObjectCall) ir.ObjectCall {
				if call.Reflection != nil && ((rule == "fold assign check" && call.Method == "assign") || (rule != "fold assign check" && call.Method == "entries")) {
					changed++
					if rule == "fold check" || rule == "fold assign check" {
						call.Reflection = nil
					} else if rule == "literal member" {
						for i := range call.Reflection.Members {
							call.Reflection.Members[i].Literal = false
						}
					} else {
						call.Reflection.OwnPropertyOrder = false
					}
				}
				return call
			}
			for i := range program.Functions {
				mutateReflection(reflect.ValueOf(&program.Functions[i].Body).Elem(), mutate)
			}
			mutateReflection(reflect.ValueOf(&program.Main).Elem(), mutate)
			if changed == 0 {
				t.Fatal("mutant changed no check")
			}
			native, binary := nativelyUncached(t, program)
			if native.exitCode != 0 {
				t.Fatalf("mutant failed outside intended comparison: %#v", native)
			}
			node := onNode(t, path)
			if rule != "layout order" {
				if difference := disagreement(node, native); difference != "" {
					t.Fatalf("unchecked control must match Node: %s", difference)
				}
				if native.exitCode == 70 {
					t.Fatal("folded check survived")
				}
				t.Logf("%s mutant caught: wanted panic, mutant follows Node and exits zero", rule)
			} else {
				if difference := disagreement(node, native); difference != "stdout differs" {
					t.Fatalf("want key order mismatch, got %q", difference)
				}
				t.Log("integer-key fixture catches layout order by Node stdout mismatch")
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestObjectReflectionAdamicPolicy(t *testing.T) {
	for _, name := range []string{"scanner", "hidden", "assign_extra", "assign_control"} {
		path := filepath.Join(repository, "internal/oracle/testdata/object_reflection_ruling", name+".a")
		_, err := lowered(t, path)
		if err == nil || !strings.Contains(err.Error(), "unproven shape") {
			t.Fatalf("%s must retain .a refusal, got %v", name, err)
		}
	}
}

func TestObjectReflectionSplitBackend(t *testing.T) {
	path, program := reflectionFixture(t, "integers")
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{Split: true, Sanitize: true, Jobs: 2}); err != nil {
		t.Fatal(err)
	}
	result := execute(t, binary)
	if difference := disagreement(onNode(t, path), result); difference != "" {
		t.Fatalf("split backend: %s %#v", difference, result)
	}
}
