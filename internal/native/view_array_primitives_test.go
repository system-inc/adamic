package native

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrimitiveArraySnapshotStorageAndLifetime(t *testing.T) {
	source := `#include "view_array_primitives.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
int main(int argc, char **argv) {
 if(argc!=2)return 2;
 int scenario=atoi(argv[1]);
 bool references=scenario==1 || scenario==2 || scenario==7;
 adamic_array *array=adamic_array_new(1,references);array->length=1;
 array->element_kind=references?10:1;
 array->elements[0].number=42;
 if(scenario==1)array->elements[0].reference=adamic_box_number(42);
 if(scenario==2){adamic_string *s=adamic_string_allocate(4);memcpy((char *)s->bytes,"word",4);array->elements[0].reference=s;}
 if(scenario==3)array->element_kind=0;
 if(scenario==4)array->element_kind=10;
 if(scenario==5){array->element_kind=2;array->elements[0].boolean=true;}
 if(scenario==7)array->elements[0].reference=&adamic_null;
 adamic_view_union_value value=adamic_view_primitive_array_snapshot(array,scenario==6?99:0,false,"value[index]","string | number");
 adamic_view_union_member members[]={{adamic_view_union_number,false,{.number=0},0},{adamic_view_union_string,false,{.number=0},0}};
 (void)adamic_view_mixed_union_select(&value,members,2,NULL,NULL,"value[index]","string | number");
 adamic_heap *box=adamic_view_primitive_array_box(value);
 adamic_release(array);
 if(box->kind==adamic_kind_number)printf("%.0f\n",((adamic_number_box *)box)->number);
 else {adamic_string *text=(adamic_string *)box;fwrite(text->bytes,1,text->length,stdout);putchar('\n');}
 adamic_release(box);return 0;
}`
	binary := filepath.Join(t.TempDir(), "array-primitive")
	if err := Build(source+"\n", binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ argument, output, found string }{
		{"0", "42\n", ""}, {"1", "42\n", ""}, {"2", "word\n", ""},
		{"3", "", "unsupported representation"}, {"4", "", "unsupported representation"},
		{"5", "", "boolean"}, {"6", "", "undefined"}, {"7", "", "null"},
	} {
		t.Run(sample.argument, func(t *testing.T) {
			command := exec.Command(binary, sample.argument)
			var out, stderr bytes.Buffer
			command.Stdout = &out
			command.Stderr = &stderr
			err := command.Run()
			if sample.found == "" {
				if err != nil || out.String() != sample.output || stderr.Len() != 0 {
					t.Fatalf("storage/lifetime: %v stdout %q stderr %q", err, out.String(), stderr.String())
				}
				return
			}
			failure, ok := err.(*exec.ExitError)
			want := "adamic: panic: cast failed: field read failed: value[index] matches no member of string | number; expected string | number, found " + sample.found + "\n"
			if !ok || failure.ExitCode() != 70 || out.Len() != 0 || stderr.String() != want {
				t.Fatalf("guard: %v stdout %q stderr %q, want %q", err, out.String(), stderr.String(), want)
			}
		})
	}
	// A target-derived certificate must not substitute for missing producer metadata.
	wrapper := `#include "view_array_primitives.h"
static adamic_view_union_value forged_storage(const adamic_array *array, double index, bool relative, const char *expression, const char *declared) {
 unsigned char saved=array->element_kind;
 if(saved==0)((adamic_array *)array)->element_kind=1;
 adamic_view_union_value value=adamic_view_primitive_array_snapshot(array,index,relative,expression,declared);
 ((adamic_array *)array)->element_kind=saved;
 return value;
}
`
	mutant := filepath.Join(t.TempDir(), "forged-storage")
	if err := Build(wrapper+strings.ReplaceAll(source, "adamic_view_primitive_array_snapshot(", "forged_storage(")+"\n", mutant, Options{Sanitize: true}); err != nil {
		t.Fatal("metadata mutant must compile", err)
	}
	command := exec.Command(mutant, "3")
	var out, stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr
	if err := command.Run(); err != nil || out.String() != "42\n" || stderr.Len() != 0 {
		t.Fatalf("metadata mutant must execute actual scalar: %v stdout %q stderr %q", err, out.String(), stderr.String())
	}
	t.Log("forged producer metadata mutant caught by named unknown-storage refusal")

}
