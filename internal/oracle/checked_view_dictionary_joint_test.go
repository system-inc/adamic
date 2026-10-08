package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryJoint(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set ADAMIC_BRAND_ORIGINAL_DECLS to the pinned original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	for _, group := range []string{"buildoptions", "compileroptions"} {
		for _, name := range []string{"good", "good-record", "missing", "wrong-element", "wrong-representation", "wrong-value"} {
			t.Run(group+"/"+name, func(t *testing.T) {
				program, file := dictionaryJointFixture(t, directory, group, name)
				root := "BuildOptions"
				if group == "compileroptions" {
					root = "CompilerOptions"
				}
				assertOriginalDictionaryFields(t, program, root)
				truth := onNode(t, file)
				if truth.exitCode != 0 {
					t.Fatalf("Node: %#v", truth)
				}
				want := truth
				switch name {
				case "wrong-element":
					want = run{exitCode: 70, stderr: []byte("adamic: panic: element read failed: values[0] expected string, found number\n")}
				case "wrong-representation":
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: dictionary['name']; expected string[] | undefined, found array\n")}
				case "wrong-value":
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: dictionary['name']; expected string[] | undefined, found number\n")}
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if diff := disagreement(want, got); diff != "" {
						t.Fatalf("%s: %#v", diff, got)
					}
				}
				if name == "good" {
					data, err := os.ReadFile(file)
					if err != nil {
						t.Fatal(err)
					}
					mutable := strings.ReplaceAll(string(data), "readonly [key:string]", "[key:string]")
					path := filepath.Join(t.TempDir(), "mutable.a")
					if err := os.WriteFile(path, []byte(mutable), 0600); err != nil {
						t.Fatal(err)
					}
					_, err = lowered(t, path)
					if err == nil || !strings.Contains(err.Error(), "fixed objects and records have different storage") {
						t.Fatalf("mutable conversion must remain refused: %v", err)
					}
				}
				if strings.HasPrefix(name, "wrong-") {
					assertDictionaryJointMutant(t, program, name)
				}
				if want.exitCode == 0 {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

// Mutants preserve safe physical extraction so a lost diagnostic, rather than
// a compiler error or invalid dereference, kills each mutation.
func assertDictionaryJointMutant(t *testing.T, program *ir.Program, name string) {
	t.Helper()
	c, js := native.C(program), javascript.JavaScript(program)
	if name == "wrong-element" {
		anchor := "adamic_view_array_at("
		if !strings.Contains(c, anchor) {
			t.Fatal("missing transitive array check")
		}
		wrapper := `#include "adamic.h"
#include "view_arrays.h"
#include <string.h>
static adamic_string joint_wrong_text = ADAMIC_STRING("42");
static adamic_value *joint_array_mutant(const adamic_array *a,double i,bool relative,bool allowed,unsigned char wanted,const char *expected,const char *expression,adamic_value *snapshot,adamic_array *owner) {
 if(strcmp(expression,"values[0]")==0) { snapshot->reference=&joint_wrong_text; return snapshot; }
 return adamic_view_array_at(a,i,relative,allowed,wanted,expected,expression,snapshot,owner);
}
`
		c = wrapper + strings.ReplaceAll(c, anchor, "joint_array_mutant(")
		if !strings.Contains(js, "adamicViewArrayElement(") {
			t.Fatal("missing JS transitive check")
		}
		js = "const jointElementMutant=(value,expression,...args)=>expression==='values[0]'?String(value):adamicViewArrayElement(value,expression,...args);\n" + strings.ReplaceAll(js, "adamicViewArrayElement(", "jointElementMutant(")
	} else {
		anchor := "adamic_view_dictionary_source_read("
		if !strings.Contains(c, anchor) {
			t.Fatal("missing representation read")
		}
		action := `adamic_view_dictionary_result r=adamic_view_dictionary_source_read(object,key,kinds | (1u<<adamic_view_union_number),child,expression,declared); if(r.value.kind==adamic_view_union_number) return (adamic_view_dictionary_result){{adamic_view_union_undefined,{.reference=NULL}},0}; return r;`
		jsAction := `const r=adamicViewDictionaryRead(record,key,[...kinds,'number'],child,expression,declared);return typeof r.value==='number'?{value:undefined,contract:0}:r;`
		if name == "wrong-representation" {
			action = `if(object && object->heap.kind==adamic_kind_array)return (adamic_view_dictionary_result){{adamic_view_union_undefined,{.reference=NULL}},0}; return adamic_view_dictionary_source_read(object,key,kinds,child,expression,declared);`
			jsAction = `if(Array.isArray(record))return {value:undefined,contract:0};return adamicViewDictionaryRead(record,key,kinds,child,expression,declared);`
		}
		wrapper := `#include "view_dictionaries.h"
#include <string.h>
static adamic_view_dictionary_result joint_dictionary_mutant(const adamic_object *object,const adamic_string *key,unsigned int kinds,size_t child,const char *expression,const char *declared) {
 if(strcmp(expression,"dictionary['name']")==0){` + action + `}
 return adamic_view_dictionary_source_read(object,key,kinds,child,expression,declared);
}
`
		c = wrapper + strings.ReplaceAll(c, anchor, "joint_dictionary_mutant(")
		if !strings.Contains(js, "adamicViewDictionaryRead(") {
			t.Fatal("missing JS dictionary selector")
		}
		js = "const jointDictionaryMutant=(record,key,kinds,child,expression,declared)=>{if(expression===\"dictionary['name']\"){ " + jsAction + " }return adamicViewDictionaryRead(record,key,kinds,child,expression,declared);};\n" + strings.ReplaceAll(js, "adamicViewDictionaryRead(", "jointDictionaryMutant(")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(c, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal("mutant must compile", err)
	}
	file := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(file, []byte(js), 0600); err != nil {
		t.Fatal(err)
	}
	output := "missing\n"
	if name == "wrong-element" {
		output = "42\n"
	}
	for _, got := range []run{execute(t, binary), onNode(t, file)} {
		if got.exitCode != 0 || string(got.stdout) != output || len(got.stderr) != 0 {
			t.Fatalf("mutant must lose pinned refusal safely: %#v", got)
		}
		t.Logf("%s mutant caught: exit 0, stdout %q", name, got.stdout)
	}
}

func dictionaryJointFixture(t *testing.T, directory, group, name string) (*ir.Program, string) {
	t.Helper()
	input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/dictionaries/joint/" + group + "-" + name + ".a"))
	if err != nil {
		t.Fatal(err)
	}
	bound := string(input)
	for _, module := range []string{"types", "tsbuildPublic"} {
		bound = strings.ReplaceAll(bound, "'original-tsc-"+module+"'", fmt.Sprintf("%q", filepath.Join(directory, "compiler", module+".d.ts")))
	}
	file := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	return program, file
}

func TestCheckedViewDictionaryJointCounts(t *testing.T) {
	directory := os.Getenv("ADAMIC_BRAND_ORIGINAL_DECLS")
	if directory == "" {
		t.Skip("set original declarations")
	}
	verifyUnionTargetDeclarations(t, directory)
	table := "| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |\n| --- | --- | --- | --- | --- | --- | --- |\n"
	for _, group := range []string{"buildoptions", "compileroptions"} {
		for _, name := range []string{"good", "good-record", "missing", "wrong-element", "wrong-representation", "wrong-value"} {
			program, _ := dictionaryJointFixture(t, directory, group, name)
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			got := execute(t, binary)
			match := countsLine.FindSubmatch(got.stderr)
			if match == nil {
				t.Fatalf("missing counts: %#v", got)
			}
			table += "| " + group + "-" + name + " | " + strings.Join([]string{string(match[1]), string(match[2]), string(match[3]), string(match[4]), string(match[5]), string(match[6])}, " | ") + " |\n"
		}
	}
	path := checkedViewFixturePath("../../stage3/interface-downcasts/dictionaries/joint/counts.md")
	if *updateCounts {
		if err := os.WriteFile(path, []byte(table), 0644); err != nil {
			t.Fatal(err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != table {
			t.Fatal("refresh joint counts with -args -update-counts")
		}
	}
}
