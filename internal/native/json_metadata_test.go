package native

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func jsonMetadataSource(t *testing.T) string {
	t.Helper()
	e := &emitter{program: &ir.Program{Locals: []ir.Local{{Type: ir.Object}, {Type: ir.String}}, Functions: []ir.Function{
		{Name: "hook", Parameters: []int{0, 1}, Returns: ir.Union},
		{Name: "returned", Parameters: []int{0}, Returns: ir.Number},
		{Name: "outer", Parameters: []int{0}, Returns: ir.Object},
		{Name: "reentrant", Parameters: []int{0}, Returns: ir.Number},
	}}, reuse: &reusePlan{}}
	shapes := []struct {
		name    string
		names   []string
		types   []ir.Type
		methods []ir.Method
	}{
		{"PAIR", []string{"0", "1"}, []ir.Type{ir.Boolean, ir.String}, nil},
		{"ORDERED", []string{"10", "2", "tail", "#public"}, []ir.Type{ir.String, ir.String, ir.Boolean, ir.String}, nil},
		{"PLUGIN", []string{"name"}, []ir.Type{ir.String}, nil},
		{"OPTIONS", []string{"strict", "target", "plugins", "hidden"}, []ir.Type{ir.Boolean, ir.Number, ir.Array, ir.String}, nil},
		{"REF", []string{"path", "prepend", "circular"}, []ir.Type{ir.String, ir.Boolean, ir.Boolean}, nil},
		{"INFO", []string{"version", "root", "fileNames", "options", "hidden"}, []ir.Type{ir.String, ir.Array, ir.Array, ir.Object, ir.String}, nil},
		{"SIMPLE", []string{"value"}, []ir.Type{ir.Number}, nil},
		{"FLAGS", []string{"a", "b"}, []ir.Type{ir.Boolean, ir.Boolean}, nil},
		{"PROPERTY", []string{"p"}, []ir.Type{ir.Object}, nil},
		{"DERIVED", []string{"mode", "extra"}, []ir.Type{ir.Number, ir.String}, []ir.Method{{Name: "toJSON", Function: 0}}},
		{"REENTRANT", []string{"mode"}, []ir.Type{ir.Number}, []ir.Method{{Name: "toJSON", Function: 3}}},
		{"HOOK", []string{"mode"}, []ir.Type{ir.Number}, []ir.Method{{Name: "toJSON", Function: 0}}},
		{"RETURNED", []string{"value"}, []ir.Type{ir.Number}, []ir.Method{{Name: "toJSON", Function: 1}}},
		{"OUTER", nil, nil, []ir.Method{{Name: "toJSON", Function: 2}}},
	}
	prefix := `#include "json_metadata.h"
#include <math.h>
static adamic_heap *adamic_function_0_hook(adamic_object *, adamic_string *);
static double adamic_function_1_returned(adamic_object *);
static adamic_object *adamic_function_2_outer(adamic_object *);
static double adamic_function_3_reentrant(adamic_object *);
`
	for _, shape := range shapes {
		name := e.shapeWith(shape.names, shape.types, shape.methods, shape.name == "PAIR")
		prefix += fmt.Sprintf("#define %s_SHAPE %s\n", shape.name, name)
	}
	prefix += "#define PRIVATE_SHAPE " + e.shapeWithPrivate([]string{"#secret", "value"}, []ir.Type{ir.Number, ir.Number}, nil, []bool{true, false}, false) + "\n"
	consumer, err := os.ReadFile("testdata/json-metadata/consumer.c")
	if err != nil {
		t.Fatal(err)
	}
	return prefix + strings.Join(e.declarations, "\n") + "\n" + string(consumer)
}

func TestJSONMetadataContractMatchesNode(t *testing.T) {
	source := jsonMetadataSource(t)
	want, err := exec.Command("node", "testdata/json-metadata/expected.mjs").CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v\n%s", err, want)
	}
	for _, options := range []Options{{Count: true}, {Sanitize: true}} {
		binary := filepath.Join(t.TempDir(), "consumer")
		if err := Build(source, binary, options); err != nil {
			t.Fatal(err)
		}
		got, err := exec.Command(binary).Output()
		if err != nil {
			t.Fatalf("consumer: %v", err)
		}
		if string(got) != string(want) {
			t.Fatalf("native:\n%s\nNode:\n%s", got, want)
		}
	}
}

func TestJSONMetadataWrongFactoryDescriptorMutant(t *testing.T) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	changed := false
	for _, file := range files {
		contents := string(file.contents)
		if file.name == "array.c" {
			before := "adamic_array *array = adamic_array_new_typed((size_t)length, references, schema);"
			after := "adamic_array *array = adamic_array_new_typed((size_t)length, references, schema == &adamic_json_number_schema ? &adamic_json_boolean_schema : schema);"
			if strings.Count(contents, before) != 1 {
				t.Fatal("mutant target missing")
			}
			contents = strings.Replace(contents, before, after, 1)
			changed = true
		}
		if err := os.WriteFile(filepath.Join(directory, file.name), []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("mutant not applied")
	}
	binary := filepath.Join(t.TempDir(), "consumer")
	options := Options{Count: true}
	library, err := RuntimeLibrary(directory, options)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "main.c")
	if err := os.WriteFile(source, []byte(jsonMetadataSource(t)), 0644); err != nil {
		t.Fatal(err)
	}
	arguments := append(LinkFlags(options), "-I", filepath.Dir(library), "-o", binary, source)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("mutant must compile: %v\n%s", err, output)
	}
	want, err := exec.Command("node", "testdata/json-metadata/expected.mjs").Output()
	if err != nil {
		t.Fatal(err)
	}
	got, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatalf("mutant must compile and run, then fail Node agreement: %v", err)
	}
	if string(got) == string(want) {
		t.Fatal("wrong factory descriptor escaped Node agreement")
	}
	t.Log("wrong number-to-boolean factory descriptor caught by Node agreement")
}

func TestJSONSharedHeaderMatchesLibrary(t *testing.T) {
	contents, err := runtime.ReadFile("runtime/json_stringify.h")
	if err != nil {
		t.Fatal(err)
	}
	const expected = "7e75c5953e8d3f1cbd51066c850bf53b1779eac177f630030a4799dddcee9e9a"
	if got := fmt.Sprintf("%x", sha256.Sum256(contents)); got != expected {
		t.Fatalf("shared header differs from cf2cb58b: %s", got)
	}
}

func TestJSONMetadataMissingDescriptorsAndThrow(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "consumer")
	if err := Build(jsonMetadataSource(t), binary, Options{Count: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"empty", "poison"} {
		output, err := exec.Command(binary, mode).CombinedOutput()
		if err == nil || !strings.Contains(string(output), "JSON.stringify runtime array without complete element descriptors") {
			t.Fatalf("%s: %v\n%s", mode, err, output)
		}
		if strings.Contains(string(output), "AddressSanitizer") {
			t.Fatalf("descriptor refusal read a poisoned slot: %s", output)
		}
	}
	want, err := exec.Command("node", "-e", `let calls=0;class Hook{toJSON(){calls++;throw Error("stop");}};try{JSON.stringify(new Hook());}catch{console.log(calls===1?"hook threw once":"wrong");}`).Output()
	if err != nil {
		t.Fatal(err)
	}
	got, err := exec.Command(binary, "throw").Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("throw handoff: %q; Node %q", got, want)
	}
}
