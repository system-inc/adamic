package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

// This tests actual source storage normalization before enabling read dispatch.
// Reference callbacks remain mandatory; a heap category is not conformance.
func TestCheckedViewPrimitiveHeap(t *testing.T) {
	path, err := filepath.Abs("../../stage3/interface-downcasts/lane4/selection-fixtures/heap-primitives-good.a")
	if err != nil {
		t.Fatal(err)
	}
	source := run{stdout: []byte("string:word\nnumber:42\nboolean:false\nundefined:undefined\n")}
	if difference := disagreement(source, onNode(t, path)); difference != "" {
		t.Fatal("source Node: " + difference)
	}

	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "primitive-heap")
		if err := native.Build(primitiveHeapC, binary, native.Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		for _, sample := range []struct{ variant, text, found string }{
			{"string", "string:word", ""}, {"number", "number:42", ""}, {"false", "boolean:false", ""}, {"undefined", "undefined:undefined", ""},
			{"true", "", "boolean"}, {"object", "", "object"}, {"unknown", "", "unsupported representation"},
		} {
			want := run{stdout: []byte(sample.text + "\n")}
			if sample.found != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.value matches no member of string | number | false | undefined; expected string | number | false | undefined, found " + sample.found + "\n")}
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

const primitiveHeapC = `
#include "view_unions_mixed.h"
#include <stdio.h>
#include <string.h>
static adamic_string word = ADAMIC_STRING("word");
int main(int argc,char **argv) {
 if(argc!=2)return 2;
 adamic_start(argc,argv);
 const char *variant=argv[1];
 adamic_heap *number=adamic_box_number(42);
 adamic_heap object={0,adamic_kind_object,0};
 adamic_heap unknown={0,adamic_kind_cell,0};
 const adamic_heap *value=strcmp(variant,"string")==0?&word.heap:strcmp(variant,"number")==0?number:strcmp(variant,"false")==0?&adamic_box_false.heap:strcmp(variant,"true")==0?&adamic_box_true.heap:strcmp(variant,"object")==0?&object:strcmp(variant,"unknown")==0?&unknown:NULL;
 const adamic_view_union_member members[]={
  {adamic_view_union_string,false,{.reference=NULL},0},
  {adamic_view_union_number,false,{.number=0},0},
  {adamic_view_union_boolean,true,{.boolean=false},0},
  {adamic_view_union_undefined,false,{.reference=NULL},0}
 };
 adamic_view_union_value snapshot=adamic_view_union_heap(value);
 size_t selected=adamic_view_mixed_union_select(&snapshot,members,4,NULL,NULL,"view.value","string | number | false | undefined");
 adamic_string *text=adamic_union_to_string((adamic_heap *)value);
 printf("%s:%.*s\n",adamic_view_union_kind_name(members[selected].kind),(int)text->length,text->bytes);
 adamic_release(text);
 adamic_release(number);
 return adamic_process_status();
}
`
