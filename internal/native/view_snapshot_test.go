package native

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestViewSnapshotPayloadOwnership(t *testing.T) {
	t.Parallel()
	source := `
#include "view_unions_mixed.h"
#include <string.h>
#include <stdio.h>
static const char *const names[]={"value"};
static const bool references[]={false};
static const adamic_shape shape={1,names,references,NULL};
int main(void){
 adamic_object *object=adamic_object_new(&shape);
 adamic_slot_cache cache={0};
 const unsigned char reps[]={adamic_rep_number,adamic_rep_boolean,adamic_rep_maybe_number,adamic_rep_maybe_boolean,adamic_rep_null,adamic_rep_undefined,0};
 for(size_t i=0;i<sizeof reps;i++){
  for(int present=0;present<2;present++){
   memset(&object->slots[0],0x5a,sizeof object->slots[0]);
   adamic_value expected={0};
   unsigned char rep=reps[i];
   if(rep==adamic_rep_number){object->slots[0].number=42;expected.number=42;}
   if(rep==adamic_rep_boolean){object->slots[0].boolean=present;expected.boolean=present;}
   if(rep==adamic_rep_maybe_number){object->slots[0].number=adamic_maybe_number_pack((adamic_maybe_number){present,7});if(present)expected.number=7;}
   if(rep==adamic_rep_maybe_boolean){object->slots[0].maybe_boolean=adamic_maybe_boolean_pack((adamic_maybe_boolean){present,true});if(present)expected.boolean=true;}
   if(rep==0)expected=object->slots[0];
   adamic_object_field_types(object)[0]=rep;
   adamic_view_union_value value=adamic_object_view_union_snapshot(object,"value",&cache,"object.value","unknown",false);
   if(memcmp(&value.payload,&expected,sizeof expected)!=0){fprintf(stderr,"rep=%u present=%d got=%a expected=%a\n",rep,present,value.payload.number,expected.number);adamic_release(object);return 3;}
  }
 }
 adamic_slot_cache absent_cache={0};
 adamic_view_union_value absent=adamic_object_view_union_snapshot(object,"missing",&absent_cache,"object.missing","undefined",true);
 adamic_value zero={0};
 if(absent.kind!=adamic_view_union_undefined || memcmp(&absent.payload,&zero,sizeof zero)!=0)return 4;
 adamic_release(object);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "snapshot")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
}
