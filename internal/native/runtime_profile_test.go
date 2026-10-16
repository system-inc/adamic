package native

import (
	"path/filepath"
	"testing"
)

// Equal heap strings may have either the same header or separate headers. Undefined is its own
// value, and equal-length strings still need a byte comparison when their headers differ.
func TestRuntimeStringEquality(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
int main(void) {
 static adamic_string a = ADAMIC_STRING("made");
 static adamic_string b = ADAMIC_STRING("here");
 adamic_string *left = adamic_string_concat(2, (adamic_string *const[]){&a,&b});
 adamic_string *copy = adamic_string_concat(2, (adamic_string *const[]){&a,&b});
 adamic_string *other = adamic_string_concat(2, (adamic_string *const[]){&b,&a});
 printf("%d %d %d %d %d %d\n", adamic_string_equal(left,left), adamic_string_equal(left,copy), adamic_string_equal(left,other), adamic_string_equal(NULL,NULL), adamic_string_equal(left,NULL), adamic_string_equal(NULL,left));
 adamic_release(left);adamic_release(copy);adamic_release(other);
 return 0;
}
`
	want := runWithInput(t, "", "node", "--eval", `const a=['made','here'].join(''),b=['made','here'].join(''),c=['here','made'].join('');console.log([a===a,a===b,a===c,undefined===undefined,a===undefined,undefined===a].map(Number).join(' '));`)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "equal")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %v: %q; Node %q", sanitize, got, want)
		}
	}
}
