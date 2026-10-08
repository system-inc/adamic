package oracle

import (
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Normalized component snapshots exercise the family hook before shared source
// admission exists. They are not evidence that any census pair has been unlocked.
func TestCheckedViewUntaggedSelection(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "untagged")
	if err := native.Build(untaggedSelectionC, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	sanitized := filepath.Join(t.TempDir(), "untagged-sanitized")
	if err := native.Build(untaggedSelectionC, sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ name, stdout, nodeWrong string }{
		{"binding-name", "1:name\n2:name\n3:name\n4:name\n5:name\n6:name\n7:name\n8:name\n9:name\n10:name\n11:name\n12:name\n", "99:name\n"},
		{"option-element", "1:name\n2:name\n3:name\n4:name\n5:name\n", "99:name\n"},
		{"structural", "2:name\n", "99:true\n"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path, err := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/" + sample.name + "-good.a")
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(sample.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("source Node: " + difference)
			}
			for _, got := range []run{execute(t, binary, sample.name, "good"), execute(t, sanitized, sample.name, "good"), untaggedSelectionJS(t, sample.name, "good")} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
			for _, variant := range []string{"wrong", "nested", "missing", "uninitialized", "no-adapter"} {
				source := sample.nodeWrong
				if variant == "nested" {
					source = "2:true\n"
				}
				if variant == "missing" || variant == "uninitialized" {
					source = "2:undefined\n"
				}
				if variant == "no-adapter" {
					source = "2:name\n"
				}
				path, err := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/" + sample.name + "-" + variant + ".a")
				if err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte(source)}, onNode(t, path)); difference != "" {
					t.Fatal("negative source Node: " + difference)
				}
				want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of " + sample.name + "; expected " + sample.name + ", found object\n")}
				for _, got := range []run{execute(t, binary, sample.name, variant), executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized, sample.name, variant), untaggedSelectionJS(t, sample.name, variant)} {
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s: %s; got %#v", variant, difference, got)
					}
				}
			}
		})
	}
}

func untaggedSelectionJS(t *testing.T, family, variant string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "selection.mjs")
	source := "import {panic} from 'adamic';\n" + javascript.MixedUnionRuntime() + javascript.UntaggedUnionRuntime() + untaggedSelectionJSCases + "\nprobeCase(" + strconv.Quote(family) + "," + strconv.Quote(variant) + ");\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path)
}

const untaggedSelectionJSCases = `
const snapshot=value=>({kind:typeof value,value});
const slotProbe=(object,field)=>{
 const descriptor=Object.getOwnPropertyDescriptor(object.value,field);
 if(descriptor===undefined || !Object.hasOwn(descriptor,'value')) return undefined;
 return {present:true,initialized:descriptor.value!==undefined,snapshot:snapshot(descriptor.value)};
};
const matches=(member,object)=>{
 const text=slotProbe(object,'text');
 if(text!==undefined) return member.contract===1 && text.initialized && text.snapshot.kind==='string';
 const child=slotProbe(object,'child');
 if(child===undefined || !child.initialized || child.snapshot.kind!=='object' || child.snapshot.value===null) return false;
 const label=slotProbe(child.snapshot,'label');
 return member.contract!==1 && label!==undefined && label.initialized && label.snapshot.kind==='string';
};
const probeCase=(family,variant)=>{
 const count=family==='binding-name'?12:family==='option-element'?5:2;
 const members=Array.from({length:count},(_,index)=>({contract:index+1,tags:family==='structural'?[]:[{field:index===count-1?'flavor':'kind',allowed:[{kind:'number',literal:true,value:index+1}]}]}));
 let values=family==='structural'?[{child:{label:'name'}}]:Array.from({length:count},(_,index)=>index===0?{kind:1,text:'name'}:index===count-1?{flavor:count,child:{label:'name'}}:{kind:index+1,child:{label:'name'}});
 if(variant==='wrong') values=[{kind:99,flavor:99,child:{label:family==='structural'?true:'name'}}];
 if(variant==='nested') values=[{kind:2,child:{label:true}}];
 if(variant==='missing') values=[{kind:2,child:{}}];
 if(variant==='uninitialized') values=[{kind:2,child:{label:undefined}}];
 if(variant==='no-adapter') values=[{kind:2,child:{label:'name'}}];
 for(const value of values){
  const contract=adamicViewUntaggedUnionSelect(snapshot(value),members,slotProbe,variant==='no-adapter'?undefined:matches,'view.value',family);
  console.log(contract+':'+(value.text===undefined?value.child.label:value.text));
 }
};
`

const untaggedSelectionC = `
#include "view_unions_untagged.h"
#include <stdio.h>
#include <string.h>
static adamic_string name=ADAMIC_STRING("name");
typedef struct sample { double kind; double flavor; bool text; bool missing; bool initialized; adamic_view_union_value label; } sample;
static bool probe_slot(void *context,const adamic_view_union_value *value,const char *field,adamic_view_union_value *slot){
 (void)context;const sample *object=value->payload.reference;
 if(strcmp(field,"kind")==0 && object->kind!=0){*slot=(adamic_view_union_value){adamic_view_union_number,{.number=object->kind}};return true;}
 if(strcmp(field,"flavor")==0 && object->flavor!=0){*slot=(adamic_view_union_value){adamic_view_union_number,{.number=object->flavor}};return true;}
 return false;
}
static bool matches(void *context,const adamic_view_union_member *member,const adamic_view_union_value *value){
 (void)context;const sample *object=value->payload.reference;
 if(object->text) return member->contract==1 && object->label.kind==adamic_view_union_string;
 return member->contract!=1 && !object->missing && object->initialized && object->label.kind==adamic_view_union_string;
}
int main(int argc,char **argv){
 if(argc!=3)return 2;
 const char *family=argv[1],*variant=argv[2];
 bool structural=strcmp(family,"structural")==0;
 size_t count=strcmp(family,"binding-name")==0?12:strcmp(family,"option-element")==0?5:2;
 adamic_view_union_member literals[12]={0};adamic_view_untagged_tag tags[12]={0};adamic_view_untagged_member members[12]={0};
 for(size_t i=0;i<count;i++){
  literals[i]=(adamic_view_union_member){.kind=adamic_view_union_number,.literal=true,.value={.number=(double)i+1}};
  tags[i]=(adamic_view_untagged_tag){i==count-1?"flavor":"kind",&literals[i],1};
  members[i]=(adamic_view_untagged_member){i+1,&tags[i],structural?0:1};
 }
 size_t amount=structural?1:count;
 if(strcmp(variant,"good")!=0)amount=1;
 for(size_t i=0;i<amount;i++){
  sample object={.kind=(double)i+1,.text=i==0,.initialized=true,.label={adamic_view_union_string,{.reference=&name}}};
  if(i==count-1){object.flavor=(double)count;object.kind=0;}
  if(structural){object.kind=2;object.text=false;}
  if(strcmp(variant,"good")!=0){object.kind=2;object.text=false;object.flavor=0;}
  if(strcmp(variant,"wrong")==0){object.kind=99;object.flavor=99;if(structural)object.label=(adamic_view_union_value){adamic_view_union_boolean,{.boolean=true}};}
  if(strcmp(variant,"nested")==0)object.label=(adamic_view_union_value){adamic_view_union_boolean,{.boolean=true}};
  if(strcmp(variant,"missing")==0)object.missing=true;
  if(strcmp(variant,"uninitialized")==0)object.initialized=false;
  adamic_view_union_value value={adamic_view_union_object,{.reference=&object}};
  size_t contract=adamic_view_untagged_union_select(&value,members,count,probe_slot,strcmp(variant,"no-adapter")==0?NULL:matches,NULL,"view.value",family);
  const char *label=object.label.kind==adamic_view_union_string?"name":"true";
  printf("%zu:%s\n",contract,label);
 }
 return 0;
}
`

// This frontier test deliberately pins the missing shared hook rather than
// pretending the component selector is source compiler admission.
func TestCheckedViewUntaggedSourceFrontier(t *testing.T) {
	for _, family := range []string{"binding-name", "option-element", "structural"} {
		t.Run(family, func(t *testing.T) {
			path, err := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/" + family + "-view-read.a")
			if err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
				t.Fatal(difference)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), loaded)
			if err == nil {
				t.Skip("owner hooks installed; tested by SourceDispatch")
			}
			if !strings.Contains(err.Error(), "checked view read of field value with unsupported untagged object union contract") {
				t.Fatalf("shared source frontier changed; remeasure admission: %v", err)
			}
			t.Logf("shared frontend refusal (both backends): %v", err)
			program, absentPath := interfaceFixture(t, "untagged/fixtures/"+family+"-absent")
			want := run{stdout: []byte("true\n")}
			t.Log(counted(t, "stage3/interface-downcasts/untagged/fixtures/"+family+"-absent.a", false, nil, false, false))
			for _, got := range []run{onNode(t, absentPath), releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("ordinary absent control: %s", difference)
				}
			}
		})
	}
}

// Run with source-overlay.py's three owner hooks until the integrator applies
// them. These tests compile the actual .a program, rather than adapter snapshots.
func TestCheckedViewUntaggedSourceDispatch(t *testing.T) {
	path, _ := filepath.Abs("../../stage3/interface-downcasts/untagged/fixtures/binding-name-source-good.a")
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), loaded)
	if err != nil {
		if os.Getenv("VIEW_UNTAGGED_SOURCE_REQUIRED") == "1" || !strings.Contains(err.Error(), "unsupported untagged object union") {
			t.Fatal(err)
		}
		t.Skip("three owner hooks pending; use source-overlay.py for source validation")
	}
	data, err := os.ReadFile("../../stage3/interface-downcasts/untagged/source-refusals.json")
	if err != nil {
		t.Fatal(err)
	}
	pins := map[string]string{}
	if err = json.Unmarshal(data, &pins); err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"binding-name", "option-element", "structural"} {
		for _, variant := range []string{"good", "wrong", "nested", "absent"} {
			t.Run(family+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "untagged/fixtures/"+family+"-source-"+variant)
				// untagged source mutation anchor
				node := onNode(t, path)
				t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
				for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if variant == "good" || variant == "absent" || variant == "empty" {
						if difference := disagreement(node, got); difference != "" {
							t.Fatal(backend + ": " + difference)
						}
					} else {
						expected, ok := pins[family+"/"+variant]
						if !ok {
							t.Fatal("missing refusal pin")
						}
						if difference := disagreement(run{exitCode: 70, stderr: []byte(expected)}, got); difference != "" {
							t.Errorf("%s: %s; got %#v", backend, difference, got)
						}
					}
				}
			})
		}
	}
}

// Fault injection at lowered source reads, using the same IR consumed by both backends.
func dropUntaggedNestedSourceReads(program *ir.Program) {
	mutate := func(expression ir.Expression) ir.Expression {
		field, ok := expression.(ir.Property)
		if ok && (field.Name == "payload" || field.Name == "label") {
			field.View = ""
			field.ViewContract = 0
			return field
		}
		return expression
	}
	// Rewriting an outer expression visits its child on the next pass.
	for i := 0; i < 3; i++ {
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	}
}

func TestCheckedViewUntaggedSourceFlows(t *testing.T) {
	for _, flow := range []string{"helper", "generic", "callback", "stored"} {
		for _, variant := range []string{"good", "wrong"} {
			t.Run(flow+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "untagged/fixtures/flow-"+flow+"-"+variant)
				want := run{stdout: []byte("true\n")}
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				if variant == "wrong" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: holder.value matches no member of Target; expected Target, found object\n")}
				}
				for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
			})
		}
	}
}

func TestCheckedViewUntaggedCandidatePairs(t *testing.T) {
	var pairs []struct {
		Rank        int
		SourceAlias string `json:"source_alias"`
		GoodStdout  string `json:"good_stdout"`
		Variants    []string
	}
	data, err := os.ReadFile("../../stage3/interface-downcasts/untagged/candidate-fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &pairs); err != nil {
		t.Fatal(err)
	}
	pinData, err := os.ReadFile("../../stage3/interface-downcasts/untagged/candidate-refusals.json")
	if err != nil {
		t.Fatal(err)
	}
	pins := map[string]string{}
	if err = json.Unmarshal(pinData, &pins); err != nil {
		t.Fatal(err)
	}
	for _, pair := range pairs {
		for _, variant := range pair.Variants {
			t.Run(strconv.Itoa(pair.Rank)+"/"+variant, func(t *testing.T) {
				name := "untagged/candidates/pair-" + strconv.Itoa(pair.Rank) + "-" + variant
				program, path := interfaceFixture(t, name)
				node := onNode(t, path)
				expectedNode := "true\n"
				if variant == "good" {
					expectedNode = pair.GoodStdout
				}
				if difference := disagreement(run{stdout: []byte(expectedNode)}, node); difference != "" {
					t.Fatal("Node control: " + difference)
				}
				t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
				// untagged candidate mutation anchor
				for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
					t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
					if variant == "good" || variant == "absent" || variant == "empty" {
						if difference := disagreement(node, got); difference != "" {
							t.Errorf("%s: %s", backend, difference)
						}
					} else {
						expected, ok := pins[strconv.Itoa(pair.Rank)+"/"+variant]
						if !ok {
							t.Fatal("missing candidate refusal pin")
						}
						if difference := disagreement(run{exitCode: 70, stderr: []byte(expected)}, got); difference != "" {
							t.Errorf("%s: %s; got %#v", backend, difference, got)
						}
					}
				}
			})
		}
	}
}

func TestCheckedViewUntaggedOwnClassData(t *testing.T) {
	for _, variant := range []string{"good", "wrong", "getter"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "untagged/fixtures/class-data-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			want := node
			if variant != "good" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n")}
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

func dropUntaggedIndexedSourceReads(program *ir.Program) {
	mutate := func(expression ir.Expression) ir.Expression {
		index, ok := expression.(ir.ArrayIndex)
		if ok {
			index.ViewContract = 0
			return index
		}
		return expression
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
}

// Recursive field-only selection must inspect complete descendants rather than
// treating a cycle in the schema as a certificate for the input.
func TestCheckedViewUntaggedRecursive(t *testing.T) {
	for _, variant := range []string{"good", "absent", "wrong", "nested", "cycle", "cycle-wrong"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "untagged/fixtures/recursive-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			want := node
			if variant == "wrong" || variant == "nested" || variant == "cycle-wrong" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of Target; expected Target, found object\n")}
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
		})
	}
}

func TestCheckedViewUntaggedArrayUnion(t *testing.T) {
	for _, variant := range []string{"good", "wrong", "nested", "empty", "mixed", "unread"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "untagged/fixtures/array-union-"+variant)
			node := onNode(t, path)
			// array transitive mutation anchor
			t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if variant == "good" || variant == "empty" || variant == "unread" {
					if difference := disagreement(node, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				} else {
					message := "adamic: panic: field read failed: view.elements is not a readonly ElementA[] | readonly ElementB[]; expected readonly ElementA[] | readonly ElementB[], found boolean\n"
					if variant == "mixed" {
						message = "adamic: panic: field read failed: view.elements matches no member of readonly ElementA[] | readonly ElementB[]; expected readonly ElementA[] | readonly ElementB[], found array\n"
					}
					if variant == "nested" {
						message = "adamic: panic: field read failed: item.payload matches no member of { readonly label: string; } | { readonly label: string; }; expected { readonly label: string; } | { readonly label: string; }, found object\n"
					}
					if difference := disagreement(run{exitCode: 70, stderr: []byte(message)}, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
			}
		})
	}
}

func TestCheckedViewUntaggedCallableUnion(t *testing.T) {
	for _, variant := range []string{"good-number", "good-string", "wrong", "nested"} {
		t.Run(variant, func(t *testing.T) {
			program, path := interfaceFixture(t, "untagged/fixtures/callable-union-"+variant)
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
				t.Fatal(difference)
			}
			for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if variant != "wrong" && variant != "nested" {
					if difference := disagreement(node, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				} else {
					field := "view.value"
					if variant == "nested" {
						field = "view.holder.value"
					}
					want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + field + " expected Target, found function with incompatible parameter representations\n")}
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s: %s", backend, difference)
					}
				}
			}
		})
	}
}

// Lane 5's aggregate callable contracts close these reduced required-checker
// boundaries. Keep the original source and require all backends and leak checks.
func TestCheckedViewUntaggedCompletedBoundaries(t *testing.T) {
	for _, fixture := range []string{"type-contract-boundary", "base-type-boundary"} {
		t.Run(fixture, func(t *testing.T) {
			program, path := interfaceFixture(t, "untagged/fixtures/"+fixture)
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte("true\n")}, truth); difference != "" {
				t.Fatal(difference)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

func TestCheckedViewUntaggedCompletedBoundaryMutants(t *testing.T) {
	for _, fixture := range []string{"type-contract-boundary", "base-type-boundary"} {
		source, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/untagged/fixtures", fixture+".a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, probe := range []struct{ name, before, after string }{
			{"callable-kind", "check:(value:Type):boolean=>true", "check:7"},
			{"callable-parameter", "check:(value:Type):boolean=>true", "check:(value:number):boolean=>true"},
			{"nested-field", "symbol:{name:'s'}", "symbol:{name:7}"},
		} {
			t.Run(fixture+"/"+probe.name, func(t *testing.T) {
				if strings.Count(string(source), probe.before) != 1 {
					t.Fatal("mutation must match once")
				}
				path := filepath.Join(t.TempDir(), "boundary.a")
				if err := os.WriteFile(path, []byte(strings.Replace(string(source), probe.before, probe.after, 1)), 0600); err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if difference := disagreement(run{stdout: []byte("true\n")}, truth); difference != "" {
					t.Fatal(difference)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				sanitized, _ := nativelyUncached(t, program)
				for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "field read failed:") || !strings.Contains(string(got.stderr), "matches no member of Type") {
						t.Fatalf("malformed boundary exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
					}
				}
				changed := 0
				mutate := func(expression ir.Expression) ir.Expression {
					switch read := expression.(type) {
					case ir.Property:
						if read.Name == "value" && read.ViewContract != 0 {
							read.ViewContract = 0
							read.View = ""
							changed++
							return read
						}
					case ir.ArrayIndex:
						if read.ViewContract != 0 {
							read.ViewContract = 0
							read.View = ""
							changed++
							return read
						}
					}
					return expression
				}
				for pass := 0; pass < 3; pass++ {
					mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				}
				if changed != 1 {
					t.Fatalf("expected one guarded Type read, changed %d", changed)
				}
				for _, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatalf("bypass must execute valid counterfactual: %s stderr %q", difference, got.stderr)
					}
					t.Log("Type read bypass caught: malformed callable or descendant executed with exit 0")
				}
			})
		}
	}
}

// Optional represented parameters now use lane 5's immutable producer contracts.
func TestCheckedViewUntaggedOptionalCallableControl(t *testing.T) {
	program, path := interfaceFixture(t, "untagged/fixtures/callable-union-optional-boundary")
	node := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("true\n")}, node); difference != "" {
		t.Fatal(difference)
	}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(node, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}
