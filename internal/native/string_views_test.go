package native

import (
	"path/filepath"
	"testing"
)

// Counts establish both ends of the view lifetime; Node establishes its bytes.
// The source and stack pieces are built at runtime so immortal literals cannot mask a missing hold.
func TestRuntimeStringViews(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <string.h>

static adamic_string *made(size_t capacity, size_t length) {
 adamic_string *text = adamic_string_allocate(capacity);
 for(size_t i=0;i<capacity;i++) ((char *)text->bytes)[i]=(char)('a'+i%26);
 text->length=length; text->units=length+1;
 return text;
}
static adamic_string *stack_piece(void) {
 char bytes[]="stack survives";
 adamic_string piece={{0,adamic_kind_string,0},sizeof bytes-1,bytes,sizeof bytes,NULL,NULL,0};
 adamic_string *copy=adamic_string_share(&piece,1,4);
 adamic_string *whole=adamic_string_share(&piece,0,piece.length);
 if(whole==&piece || whole->bytes==bytes) {fprintf(stderr,"stack piece was retained\n");return NULL;}
 adamic_release(whole);
 memset(bytes,'X',sizeof bytes-1);
 return copy;
}
int main(void) {
 adamic_string *parent=made(128,128);
 adamic_string *first=adamic_string_slice(parent,10,30,true);
 adamic_string *inner=adamic_string_slice(first,5,15,true);
 if(first->owner!=parent || first->bytes!=parent->bytes+10 || first->capacity!=0 ||
    inner->owner!=parent || inner->bytes!=parent->bytes+15 || inner->units!=11 || parent->heap.references!=3) {
  fprintf(stderr,"view offsets or holds differ\n");return 1;
 }
 adamic_release(first);adamic_release(parent);
 if(adamic_counted.live!=2 || inner->owner->heap.references!=1) {fprintf(stderr,"parent not held\n");return 1;}
 adamic_write_line(adamic_stdout,inner);adamic_release(inner);
 if(adamic_counted.live!=0) {fprintf(stderr,"last view did not free parent\n");return 1;}
 parent=made(sizeof(adamic_string)*7+64,sizeof(adamic_string)*7+64);
 adamic_string *below=adamic_string_slice(parent,1,8,true);
 adamic_string *boundary=adamic_string_slice(parent,1,9,true);
 if(below->owner!=NULL || boundary->owner!=parent) {fprintf(stderr,"eightfold storage boundary differs\n");return 1;}
 adamic_release(parent);adamic_release(below);adamic_release(boundary);
 parent=made(8192,64);
 adamic_string *tiny=adamic_string_slice(parent,1,33,true);
 if(tiny->owner!=NULL || tiny->capacity!=32) {fprintf(stderr,"spare capacity was pinned\n");return 1;}
 adamic_release(parent);adamic_write_line(adamic_stdout,tiny);adamic_release(tiny);
 adamic_string *copy=stack_piece();adamic_write_line(adamic_stdout,copy);adamic_release(copy);
 parent=made(128,128);
 for(size_t i=0;i<128;i++) ((char *)parent->bytes)[i]=(char)i;
 size_t allocations=adamic_counted.allocations, sum=0;
 for(size_t i=0;i<128;i++) {
  adamic_string *at=adamic_string_at(parent,(double)i);
  if(at==NULL || at->length!=1 || at->units!=2 || at->heap.references!=0) {fprintf(stderr,"ASCII character not immutable\n");return 1;}
  sum=(sum*33+(unsigned char)at->bytes[0])%1000003;adamic_release(at);
 }
 if(adamic_counted.allocations!=allocations) {fprintf(stderr,"ASCII indexing allocated\n");return 1;}
 adamic_output_flush();printf("%zu\n",sum);adamic_release(parent);
 if(adamic_counted.live!=0) {fprintf(stderr,"view test leaked\n");return 1;}
 return 0;
}
`
	want := runWithInput(t, "", "node", "--eval", `const made=n=>Array.from({length:n},(_,i)=>String.fromCharCode(97+i%26)).join('');console.log(made(128).slice(10,30).slice(5,15));console.log(made(64).slice(1,33));console.log('stack survives'.slice(1,5));console.log(Array.from({length:128},(_,i)=>String.fromCharCode(i)).join('').split('').reduce((s,c)=>(s*33+c.charCodeAt(0))%1000003,0));`)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "views")
		if err := Build(source, binary, Options{Sanitize: sanitize, Count: true}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: %q; Node %q", sanitize, got, want)
		}
	}
}
