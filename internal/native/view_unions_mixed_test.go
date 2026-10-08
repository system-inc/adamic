package native

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestViewMixedUnionUnknownAndUnavailable(t *testing.T) {
	t.Parallel()
	source := `
#include "view_unions_mixed.h"
#include <stdlib.h>
static bool forged(void *context,const adamic_view_union_member *member,const adamic_view_union_value *value){(void)context;(void)member;(void)value;return true;}
int main(int argc,char **argv){
 if(argc!=2)return 2;
 int scenario=atoi(argv[1]);int object=0;
 adamic_view_union_value value={adamic_view_union_unknown,{.number=42}};
 adamic_view_union_member member={adamic_view_union_number,false,{.number=0},0};
 adamic_view_union_match match=NULL;const char *declared="number | string";
 if(scenario==1||scenario==2){value=(adamic_view_union_value){adamic_view_union_object,{.reference=&object}};member.kind=adamic_view_union_object;member.contract=scenario==1?1:0;match=scenario==2?forged:NULL;declared="Left | Right";}
 if(scenario==3){value=(adamic_view_union_value){adamic_view_union_null,{.reference=NULL}};member.kind=adamic_view_union_undefined;declared="string | undefined";}
 if(scenario==4){value=(adamic_view_union_value){adamic_view_union_boolean,{.boolean=true}};member=(adamic_view_union_member){adamic_view_union_boolean,true,{.boolean=false},0};declared="false | string";}
 (void)adamic_view_mixed_union_select(&value,&member,1,match,NULL,"view.value",declared);
 return 0;
}`
	binary := filepath.Join(t.TempDir(), "mixed-union")
	if err := Build(source+"\n", binary, Options{}); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct{ argument, declared, found string }{
		{"0", "number | string", "unsupported representation"},
		{"1", "Left | Right", "object"},
		{"2", "Left | Right", "object"},
		{"3", "string | undefined", "null"},
		{"4", "false | string", "boolean"},
	} {
		t.Run(sample.argument, func(t *testing.T) {
			command := exec.Command(binary, sample.argument)
			var out, errout bytes.Buffer
			command.Stdout = &out
			command.Stderr = &errout
			err := command.Run()
			failure, ok := err.(*exec.ExitError)
			want := "adamic: panic: field read failed: view.value matches no member of " + sample.declared + "; expected " + sample.declared + ", found " + sample.found + "\n"
			if !ok || failure.ExitCode() != 70 || out.Len() != 0 || errout.String() != want {
				t.Fatalf("got %v, stdout %q, stderr %q; want exit 70 and %q", err, out.String(), errout.String(), want)
			}
		})
	}
}
