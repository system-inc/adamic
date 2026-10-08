package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// These certify only the primitive/nullish selector component of the original
// rich dictionary contract. Object and array alternatives belong to other lanes.
func TestCheckedViewOriginalDictionaryPrimitiveComponents(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, group := range []string{"compileroptions", "optionsbase", "buildoptions"} {
		for _, variant := range []string{"string", "number", "boolean", "null", "undefined", "missing", "wrong", "wrong-array", "wrong-object"} {
			t.Run(group+"/"+variant, func(t *testing.T) {
				input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/dictionary-" + group + "-" + variant + ".a"))
				if err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(t.TempDir(), "primitive.a")
				bound := strings.ReplaceAll(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.Join(directory, "compiler/types.d.ts")))
				for _, module := range []string{"commandLineParser", "tsbuildPublic"} {
					bound = strings.ReplaceAll(bound, "'original-tsc-"+module+"'", fmt.Sprintf("%q", filepath.Join(directory, "compiler", module+".d.ts")))
				}
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, file)
				if truth.exitCode != 0 {
					t.Fatalf("Node: %#v", truth)
				}
				text := map[string]string{"string": "word-built\n", "number": "42\n", "boolean": "true\n", "null": "null\n", "undefined": "undefined\n", "missing": "undefined\n", "wrong": "function\n", "wrong-array": "object\n", "wrong-object": "object\n"}[variant]
				if string(truth.stdout) != text {
					t.Fatalf("Node: %#v", truth)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				root := map[string]string{"compileroptions": "CompilerOptions", "optionsbase": "OptionsBase", "buildoptions": "BuildOptions"}[group]
				assertOriginalDictionaryFields(t, program, root)
				originalUnsupported := false
				for _, contract := range program.ViewContracts {
					originalUnsupported = originalUnsupported || strings.Contains(contract.Name, "CompilerOptionsValue") && contract.Unsupported != ""
				}
				if !originalUnsupported {
					t.Fatal("original unsupported reference descriptor was erased")
				}

				declared := "CompilerOptionsValue"
				if group == "compileroptions" || group == "optionsbase" {
					declared = "TsConfigSourceFile | CompilerOptionsValue"
				}
				want := truth
				if strings.HasPrefix(variant, "wrong") {
					found := map[string]string{"wrong": "function", "wrong-array": "array", "wrong-object": "object"}[variant]
					want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: value[key]; expected " + declared + ", found " + found + "\n")}
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				}
				if want.exitCode == 0 {
					got, binary := nativelyUncached(t, program)
					if diff := disagreement(want, got); diff != "" {
						t.Fatal(diff)
					}
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
				if variant == "wrong" {
					assertDictionaryFunctionAdmissionMutant(t, program)
				}
				if variant == "number" {
					assertDictionaryFirstMemberMutant(t, program)
				}
				if variant == "null" {
					assertDictionaryNullErasureMutant(t, program)
				}

			})
		}
	}
}

func TestCheckedViewPrimitiveDictionarySelector(t *testing.T) {
	for _, variant := range []string{"string", "number", "boolean", "null", "undefined", "missing", "record", "wrong"} {
		t.Run(variant, func(t *testing.T) {
			program, file := interfaceFixture(t, "lane4/primitive-original/dictionary-selector-"+variant)
			truth := onNode(t, file)
			text := map[string]string{"string": "word-built\n", "number": "42\n", "boolean": "true\n", "null": "null\n", "undefined": "undefined\n", "missing": "undefined\n", "wrong": "function\n", "record": "null\nword-built\n"}[variant]
			if truth.exitCode != 0 || string(truth.stdout) != text {
				t.Fatalf("Node: %#v", truth)
			}
			want := truth
			if variant == "wrong" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: cast failed: field read failed: value[key]; expected string | number | boolean | null | undefined, found function\n")}
			}
			for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("%s: %#v", diff, got)
				}
			}
			if variant != "wrong" {
				got, binary := nativelyUncached(t, program)
				if diff := disagreement(want, got); diff != "" {
					t.Fatal(diff)
				}
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
				if variant == "number" {
					assertDictionaryFirstMemberMutant(t, program)
				}
				if variant == "null" {
					assertDictionaryNullErasureMutant(t, program)
				}
				return
			}
			assertDictionaryFunctionAdmissionMutant(t, program)
		})
	}
}

func assertDictionaryNullErasureMutant(t *testing.T, program *ir.Program) {
	t.Helper()
	c := native.C(program)
	if !strings.Contains(c, "adamic_view_dictionary_box(") {
		t.Fatal("missing native box to mutate")
	}
	wrapper := `#include "view_dictionaries.h"
static adamic_heap *erase_null(adamic_view_union_value value) {
 if(value.kind==adamic_view_union_null)return NULL;
 return adamic_view_dictionary_box(value);
}
`
	binary := filepath.Join(t.TempDir(), "null-mutant")
	if err := native.Build(wrapper+strings.ReplaceAll(c, "adamic_view_dictionary_box(", "erase_null("), binary, native.Options{}); err != nil {
		t.Fatal("semantic null mutant must compile", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "adamicViewDictionaryRead(") {
		t.Fatal("missing JS dictionary read to mutate")
	}
	code = "const eraseNull = (...args) => { const result = adamicViewDictionaryRead(...args); if (result.value === null) result.value = undefined; return result; };\n" + strings.ReplaceAll(code, "adamicViewDictionaryRead(", "eraseNull(")
	js := filepath.Join(t.TempDir(), "null-mutant.mjs")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 0 || string(got.stdout) != "undefined\n" {
			t.Fatalf("null mutant must execute valid code: %#v", got)
		}
		t.Log("null/undefined erasure mutant caught by Node null control")
	}
}

func assertDictionaryFunctionAdmissionMutant(t *testing.T, program *ir.Program) {
	t.Helper()

	c := native.C(program)
	if !strings.Contains(c, "adamic_view_dictionary_source_read(") {
		t.Fatal("missing actual read to mutate")
	}
	wrapper := `#include "view_dictionaries.h"
static adamic_view_dictionary_result admit_function(const adamic_object *object,const adamic_string *key,unsigned int kinds,size_t contract,const char *expression,const char *declared) {
 return adamic_view_dictionary_source_read(object,key,kinds | (1u << adamic_view_union_function),contract,expression,declared);
}
`
	binary := filepath.Join(t.TempDir(), "member-mutant")
	if err := native.Build(wrapper+strings.ReplaceAll(c, "adamic_view_dictionary_source_read(", "admit_function("), binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "adamicViewDictionaryRead(") {
		t.Fatal("missing JS read to mutate")
	}
	code = "const admitFunction = (record,key,kinds,contract,expression,declared) => adamicViewDictionaryRead(record,key,[...kinds,'function'],contract,expression,declared);\n" + strings.ReplaceAll(code, "adamicViewDictionaryRead(", "admitFunction(")
	js := filepath.Join(t.TempDir(), "member-mutant.mjs")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 0 || string(got.stdout) != "function\n" {
			t.Fatalf("mutant must execute original wrong member: %#v", got)
		}
		t.Log("actual function-member admission mutant caught by named read pin")
	}
}

func assertOriginalDictionaryFields(t *testing.T, program *ir.Program, root string) {
	t.Helper()
	data, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/primitive-original/dictionary-candidates.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Pairs []struct {
			Type   string   `json:"type"`
			Fields []string `json:"original_fields"`
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, pair := range manifest.Pairs {
		if pair.Type != root {
			continue
		}
		for _, contract := range program.ViewContracts {
			if contract.Name != root || contract.Kind != ir.ViewDictionary {
				continue
			}
			fields := []string{}
			for _, field := range contract.Fields {
				fields = append(fields, field.Name)
			}
			slices.Sort(fields)
			if !slices.Equal(fields, pair.Fields) {
				t.Fatalf("original %s fields changed", root)
			}
			return
		}
	}
	t.Fatalf("missing complete original %s dictionary descriptor", root)
}

func assertDictionaryFirstMemberMutant(t *testing.T, program *ir.Program) {
	t.Helper()
	count := 0
	rewrite := func(value any) any {
		read, ok := value.(ir.Property)
		if call, record := value.(ir.RecordCall); record && call.Method == "get" && call.DictionaryRead != nil {
			read, ok = *call.DictionaryRead, true
		}
		if !ok || read.DictionaryKey == nil {
			return value
		}
		count++
		return ir.Box{Value: ir.NumberConstant{Value: 0}}
	}
	for i := range program.Functions {
		program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
	}
	if count != 1 {
		t.Fatalf("want one first-member mutation, got %d", count)
	}
	for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "0\n" {
			t.Fatalf("first-member mutant must execute valid code: %#v", got)
		}
		t.Log("untested dictionary first member caught by Node control")
	}
}
