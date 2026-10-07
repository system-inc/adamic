package native

import (
	"path/filepath"
	"testing"
)

// Exercise all release cases with runtime-built strings, shared owners and a deep container chain.
// Live counts catch a skipped drain even when its pending values remain reachable to LeakSanitizer;
// sanitizer builds also check ownership and nonrecursive destruction.
func TestRuntimeReleasePaths(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include "count.h"
#include <stdio.h>
#include <string.h>

int main(void) {
 static adamic_string literal = ADAMIC_STRING("literal");
 adamic_release(NULL);
 adamic_release(&literal);
 adamic_string *text = adamic_string_concat(2, (adamic_string *const[]){&literal, &literal});
 adamic_retain(text);
 adamic_release(text);
 if(adamic_counted.live != 1) {fprintf(stderr,"shared release left %zu values\n",adamic_counted.live);return 1;}
 adamic_write_line(adamic_stdout, text);
 adamic_release(text);
 if(adamic_counted.live != 0) {fprintf(stderr,"last release left %zu values\n",adamic_counted.live);return 1;}
 static const char *const names[] = {"next", "label"};
 static const bool references[] = {true, true};
 static const adamic_shape shape = {2, names, references, NULL};
 adamic_object *chain = NULL;
 for (size_t i = 0; i < 100000; i++) {
  adamic_object *next = adamic_object_new(&shape);
  next->slots[0].reference = chain;
  next->slots[1].reference = adamic_string_concat(2, (adamic_string *const[]){&literal, &literal});
  chain = next;
 }
 adamic_release(chain);
 if(adamic_counted.live != 0) {fprintf(stderr,"chain release left %zu values\n",adamic_counted.live);return 1;}
 static adamic_string done = ADAMIC_STRING("released");
 adamic_write_line(adamic_stdout, &done);
 return 0;
}
`
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "release")
		if err := Build(source, binary, Options{Sanitize: sanitize, Count: true}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != "literalliteral\nreleased\n" {
			t.Fatalf("sanitize %v: %q", sanitize, got)
		}
	}
}
