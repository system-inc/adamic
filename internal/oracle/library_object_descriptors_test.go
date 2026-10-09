package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var objectDescriptorFixtures = []string{"library_object_descriptors", "library_object_definitions", "library_object_enumerable", "library_object_descriptor_shape"}

func init() {
	for _, name := range objectDescriptorFixtures {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name + ".a", true, false})
	}
}
func TestObjectDescriptorFixtures(t *testing.T) {
	t.Parallel()
	for _, name := range objectDescriptorFixtures {
		t.Run(name, func(t *testing.T) {
			fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, fixture)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, fixture)
			actual, binary := natively(t, program)
			if d := disagreement(want, actual); d != "" {
				t.Fatalf("native: %s\nNode %s\nnative %s\nstderr %s", d, want.stdout, actual.stdout, actual.stderr)
			}
			if d := disagreement(want, onJavaScriptBackend(t, program)); d != "" {
				t.Fatal("JavaScript: " + d)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if d := disagreement(want, onWASI(t, native.C(program))); d != "" {
					t.Fatal("WASI: " + d)
				}
			}
		})
	}
}

func TestObjectDescriptorNodeMutants(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/library_object_descriptors.c"))
	if err != nil {
		t.Fatal(err)
	}
	runtimeSource := string(data)
	mutations := []struct{ name, fixture, old, replacement string }{
		{"descriptor_value", "library_object_descriptors", "default:return adamic_retain(value.reference);", "default:return (adamic_heap *)adamic_string_repeat(value.reference,2);"},
		{"frozen_writable", "library_object_descriptors", "result->slots[1].reference=object->frozen && !captured ? &adamic_box_false : &adamic_box_true;", "result->slots[1].reference=&adamic_box_true;"},
		{"definition_value", "library_object_definitions", "if(next!=NULL) {", "if(next!=NULL && kind!=1) {"},
		{"definitions_order", "library_object_definitions", "adamic_public_index(descriptors->shape,at)", "adamic_public_index(descriptors->shape,descriptors->shape->count-1-at)"},
		{"same_value_zero", "library_object_definitions", "left.number==right.number && (left.number!=0 || signbit(left.number)==signbit(right.number))", "left.number==right.number"},
		{"enumerable_stack", "library_object_enumerable", "if(object->has_captured_stack && key_is(key,\"stack\")) return false;", "if(object->has_captured_stack && key_is(key,\"stack\")) return true;"},
		{"complete_shape", "library_object_descriptor_shape", "", ""},
		{"descriptor_order", "library_object_descriptor_shape", "", ""},
	}
	names := regexp.MustCompile(`\badamic_(?:object_descriptor|object_descriptors|object_property_enumerable|descriptor_flag|descriptor_value|descriptor_number|descriptor_boolean|object_define_property|object_define_properties)\b`)
	rename := func(s string) string {
		return names.ReplaceAllStringFunc(s, func(name string) string { return "probe_" + name })
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+mutation.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := runtimeSource
			if mutation.old != "" {
				if strings.Count(source, mutation.old) != 1 {
					t.Fatal("mutant must change exactly one real operation")
				}
				source = strings.Replace(source, mutation.old, mutation.replacement, 1)
			} else {
				changed := false
				for i, statement := range program.Main {
					declare, ok := statement.(ir.Declare)
					if !ok {
						continue
					}
					call, ok := declare.Value.(ir.ObjectCall)
					if !ok || call.Method != "getOwnPropertyDescriptors" {
						continue
					}
					if mutation.name == "complete_shape" {
						call.DescriptorFields = call.DescriptorFields[:len(call.DescriptorFields)-1]
					} else {
						n := len(call.DescriptorFields)
						call.DescriptorFields[n-1], call.DescriptorFields[n-2] = call.DescriptorFields[n-2], call.DescriptorFields[n-1]
					}
					declare.Value = call
					program.Main[i] = declare
					changed = true
				}
				if !changed {
					t.Fatal("mutant changed no descriptor")
				}
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(rename(source)+"\n"+rename(native.C(program)), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("must finish and leak nothing: exit %d stderr %s", actual.exitCode, actual.stderr)
			}
			if d := disagreement(onNode(t, path), actual); d != "stdout differs" {
				t.Fatalf("Node alone must catch mutant: %q", d)
			}
			t.Log("compiled, exited zero, passed sanitizers/leaks; Node alone caught stdout differs")
		})
	}
}
