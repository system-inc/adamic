package native

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Runtime controls bypass the language's prohibition on owning reference cycles.
func TestFX7ViewCycleAndStackGuard(t *testing.T) {
	t.Parallel()
	source := `
#include "view_unions_untagged.h"
#include <stdint.h>
int main(int argc, char **argv) {
 (void)argv;
 const char *names[]={"next"}; const bool references[]={true};
 adamic_shape shape={1,names,references,NULL};
 adamic_object *object=adamic_object_new(&shape);
 object->slots[0].reference=object;
 adamic_object_field_types(object)[0]=adamic_rep_object;
 const adamic_view_untagged_field fields[]={{"next",2,false}};
 const size_t members[]={1,3};
 const adamic_view_untagged_contract contracts[]={
 {.kind=2,.of=adamic_rep_object,.fields=fields,.field_count=1},
 {.kind=4,.of=adamic_rep_object,.members=members,.member_count=2},
 {.kind=7,.of=adamic_rep_undefined}};
 if(argc>2) {
  object->slots[0].reference=NULL;
  adamic_object_field_types(object)[0]=adamic_rep_undefined;
  for(size_t i=0;i<20000;i++) {
   adamic_object *next=adamic_object_new(&shape);
   next->slots[0].reference=object;
   adamic_object_field_types(next)[0]=adamic_rep_object;
   object=next;
  }
 }
 adamic_view_union_value value=adamic_view_union_heap((const adamic_heap *)object);
 if(argc==2) adamic_stack_limit=UINTPTR_MAX;
 bool matches=adamic_view_untagged_plain_matches(contracts,3,1,&value);
 object->slots[0].reference=NULL;
 adamic_release(object);
 return matches ? 0 : 1;
}
`
	for _, sanitized := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "view-runtime")
		if err := Build(source, binary, Options{Sanitize: sanitized}); err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"cycle", "guard", "deep"} {
			guard := mode != "cycle"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			args := []string{}
			if guard {
				args = append(args, "guard")
			}
			command := exec.CommandContext(ctx, binary, args...)
			if mode == "deep" {
				command = exec.CommandContext(ctx, "sh", "-c", "ulimit -s 512; exec \"$1\" deep walk", "fx7", binary)
			}
			var out, errout bytes.Buffer
			command.Stdout = &out
			command.Stderr = &errout
			err := command.Run()
			cancel()
			if guard {
				failure, ok := err.(*exec.ExitError)
				if !ok || failure.ExitCode() != 70 || errout.String() != "adamic: panic: RangeError: Maximum call stack size exceeded\n" {
					t.Fatalf("guard: %v %q", err, errout.String())
				}
			} else if err != nil || errout.Len() != 0 {
				t.Fatalf("cycle: %v %q", err, errout.String())
			}
		}
	}
}
