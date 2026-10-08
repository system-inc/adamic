package native

import (
	"bytes"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTupleHeapUnionSelectionProbe(t *testing.T) {
	p := &ir.Program{ViewContracts: []ir.ViewContract{
		{Kind: ir.ViewScalar, Of: ir.Number, Name: "ID"},
		{Kind: ir.ViewObject, Of: ir.Object, Name: "Tuple", FixedTuple: true, Tuple: []ir.ViewContractID{1, 1}},
		{Kind: ir.ViewUnion, Of: ir.Union, Name: "ID | Tuple", Members: []ir.ViewContractID{1, 2}},
	}}
	e := emitter{program: p}
	if !e.viewTupleHeapUnion(3, "value", "item") {
		t.Fatal("tuple union adapter unavailable")
	}
	source := strings.Join(e.declarations, "\n") + `
#include <stdio.h>
#include <stdlib.h>
int main(int argc,char **argv) {
 if(argc!=2) return 2;
 int scenario=atoi(argv[1]);
 static const char *const names[]={"0","1"};
 static const bool references[]={false,false};
 static const adamic_shape shape={2,names,references,NULL};
 static const adamic_shape short_shape={1,names,references,NULL};
 adamic_heap *value;
 if(scenario==0) value=adamic_box_number(7);
 else if(scenario==3) value=(adamic_heap *)&adamic_box_false;
 else {adamic_object *object=adamic_object_new(scenario==4?&short_shape:&shape);object->tuple=scenario!=2;object->slots[0].number=7;value=(adamic_heap *)object;}
` + e.out.String() + `
 adamic_view_union_value answer=adamic_view_union_heap(value);
 printf("%.0f\n",answer.kind==adamic_view_union_number?answer.payload.number:((adamic_object *)answer.payload.reference)->slots[0].number);
 adamic_release(value);
 return adamic_process_status();
}
`
	if os.Getenv("ADAMIC_TUPLE_SELECTOR_MUTANT") == "shape" {
		source = strings.ReplaceAll(source, "return adamic_tuple_matches(value->payload.reference, 2);", "return true;")
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "selector")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		for _, sample := range []struct{ arg, found string }{{"0", ""}, {"1", ""}, {"2", "object"}, {"3", "boolean"}, {"4", "object"}} {
			command := exec.Command(binary, sample.arg)
			var out, errout bytes.Buffer
			command.Stdout = &out
			command.Stderr = &errout
			err := command.Run()
			if sample.found == "" {
				if err != nil || out.String() != "7\n" || errout.Len() != 0 {
					t.Fatalf("sanitize %v: %v %q %q", sanitize, err, out.String(), errout.String())
				}
				continue
			}
			failure, ok := err.(*exec.ExitError)
			want := "adamic: panic: cast failed: field read failed: item matches no member of ID | Tuple; expected ID | Tuple, found " + sample.found + "\n"
			if !ok || failure.ExitCode() != 70 || out.Len() != 0 || errout.String() != want {
				t.Fatalf("sanitize %v: %v %q %q", sanitize, err, out.String(), errout.String())
			}
		}
	}
}
