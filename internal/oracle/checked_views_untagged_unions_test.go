package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
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
			if err == nil || !strings.Contains(err.Error(), "checked view read of field value with unsupported untagged object union contract") {
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
