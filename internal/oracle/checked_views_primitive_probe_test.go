package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

func TestCheckedViewPrimitiveProbe(t *testing.T) {
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "primitive-probe")
		if err := native.Build(primitiveProbeC, binary, native.Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		for _, sample := range []struct{ variant, text, found string }{
			{"number", "number:42", ""}, {"string", "string:word", ""}, {"false", "boolean:false", ""},
			{"boxed-number", "number:42", ""}, {"boxed-false", "boolean:false", ""}, {"maybe-number", "number:42", ""},
			{"maybe-undefined", "undefined:undefined", ""}, {"maybe-false", "boolean:false", ""}, {"undefined", "undefined:undefined", ""},
			{"optional-missing", "undefined:undefined", ""}, {"inherited", "number:42", ""},
			{"true", "", "boolean"}, {"null", "", "null"}, {"object", "", "object"}, {"unknown", "", "unsupported representation"},
			{"mismatched-string", "", "unsupported representation"}, {"uninitialized", "", "uninitialized"},
			{"optional-uninitialized", "", "uninitialized"}, {"required-missing", "", "missing"},
		} {
			want := run{stdout: []byte(sample.text + "\n")}
			if sample.found != "" {
				message := "cast failed: field read failed: view.value matches no member of string | number | false | undefined; expected string | number | false | undefined, found " + sample.found
				if sample.found == "missing" || sample.found == "uninitialized" {
					message = "cast failed: field read failed: view.value is not initialized; expected string | number | false | undefined, found " + sample.found
				}
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + message + "\n")}
			}
			var got run
			if sanitized && sample.found != "" {
				got = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, sample.variant)
			} else {
				got = execute(t, binary, sample.variant)
			}
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("sanitized=%t variant=%s: %s; got %#v", sanitized, sample.variant, difference, got)
			}
		}
	}
}

const primitiveProbeC = `
#include "view_unions_mixed.h"
#include <stdio.h>
#include <string.h>
static adamic_string word=ADAMIC_STRING("word");
static const char *names[]={"value"};
static const bool references[]={false};
static const adamic_shape shape={.count=1,.names=names,.references=references};
static const char *child_names[]={"value","flag","parent"};
static const bool child_references[]={false,false,true};
static const adamic_shape child_shape={.count=3,.names=child_names,.references=child_references};
static const size_t flags[]={2,0,0};
static const adamic_class child_class={.count=3,.is_static=true,.static_parent=3,.static_flags=flags};
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 adamic_start(argc,argv);
 const char *variant=argv[1];
 adamic_object *object=adamic_object_new(&shape);
 adamic_heap *number=adamic_box_number(42);
 adamic_heap fake_object={0,adamic_kind_object,0};
 object->slots[0].number=42;
 unsigned char storage=1;
 bool optional=false;
 const char *field="value";
 if(strcmp(variant,"string")==0){storage=3;object->slots[0].reference=&word;}
 else if(strcmp(variant,"false")==0 || strcmp(variant,"true")==0){storage=2;object->slots[0].boolean=strcmp(variant,"true")==0;}
 else if(strcmp(variant,"boxed-number")==0){storage=10;object->slots[0].reference=number;}
 else if(strcmp(variant,"boxed-false")==0 || strcmp(variant,"maybe-false")==0){storage=strcmp(variant,"maybe-false")==0?9:10;object->slots[0].reference=&adamic_box_false;}
 else if(strcmp(variant,"maybe-number")==0 || strcmp(variant,"maybe-undefined")==0){storage=7;object->slots[0].number=adamic_maybe_number_pack((adamic_maybe_number){strcmp(variant,"maybe-number")==0,42});}
 else if(strcmp(variant,"undefined")==0 || strcmp(variant,"null")==0){storage=strcmp(variant,"null")==0?12:13;object->slots[0].reference=NULL;}
 else if(strcmp(variant,"object")==0){storage=4;object->slots[0].reference=&fake_object;}
 else if(strcmp(variant,"unknown")==0){storage=0;}
 else if(strcmp(variant,"mismatched-string")==0){storage=3;object->slots[0].reference=&adamic_box_false;}
 else if(strcmp(variant,"uninitialized")==0 || strcmp(variant,"optional-uninitialized")==0){adamic_object_initialized(object)[0]=0;optional=strcmp(variant,"optional-uninitialized")==0;}
 else if(strcmp(variant,"required-missing")==0 || strcmp(variant,"optional-missing")==0){field="absent";optional=strcmp(variant,"optional-missing")==0;}
 adamic_object_field_types(object)[0]=storage;
 if(strcmp(variant,"inherited")==0){
  adamic_object *child=adamic_object_new(&child_shape);
  child->class=&child_class;
  child->slots[1].number=0;
  child->slots[2].reference=object;
  object=child;
 }
 adamic_slot_cache cache={NULL,0};
 adamic_view_union_value value=adamic_object_view_union_snapshot(object,field,&cache,"view.value","string | number | false | undefined",optional);
 const adamic_view_union_member members[]={
  {adamic_view_union_string,false,{.reference=NULL},0},
  {adamic_view_union_number,false,{.number=0},0},
  {adamic_view_union_boolean,true,{.boolean=false},0},
  {adamic_view_union_undefined,false,{.reference=NULL},0}
 };
 size_t selected=adamic_view_mixed_union_select(&value,members,4,NULL,NULL,"view.value","string | number | false | undefined");
 printf("%s:",adamic_view_union_kind_name(members[selected].kind));
 if(value.kind==adamic_view_union_string){const adamic_string *s=value.payload.reference;printf("%.*s",(int)s->length,s->bytes);}
 else if(value.kind==adamic_view_union_number){printf("%g",value.payload.number);}
 else if(value.kind==adamic_view_union_boolean){printf("%s",value.payload.boolean?"true":"false");}
 else {printf("undefined");}
 printf("\n");
 adamic_release(object);
 adamic_release(number);
 return adamic_process_status();
}
`
