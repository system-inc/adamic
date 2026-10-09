package native

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A null payload still represents a pending throw. The first throwing index
// decides the parallel result, including when a later index throws an Error.
func TestParallelUndefinedThrowAndPendingMutant(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
static adamic_value work(adamic_closure *self, adamic_value *arguments) {
 (void)self;
 if (arguments[1].number == 3 || arguments[1].number == 9) {
  static adamic_string message = ADAMIC_STRING("later");
  adamic_thrown = arguments[1].number == 3 ? NULL : &adamic_error_new(&message)->heap;
  adamic_exception_pending = true;
 }
 return (adamic_value){.number = 1};
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 adamic_array *items = adamic_array_new(64, false);
 for (size_t i = 0; i < 64; i++) adamic_array_push(items, (adamic_value){.number = (double)i});
 adamic_closure *callback = adamic_closure_new(work, 0);
 adamic_array *results = adamic_parallel_map(items, callback, false);
 printf("%s %s\n", results == NULL ? "true" : "false", adamic_exception_pending && adamic_thrown == NULL ? "true" : "false");
 adamic_release(results); adamic_release(callback); adamic_release(items);
 adamic_release(adamic_thrown); adamic_thrown = NULL; adamic_exception_pending = false;
 return 0;
}
`
	expected, err := exec.Command("node", "-e", `let caught = false, value; try {Array.from({length:64}, (_, i) => {if(i===3) throw undefined; if(i===9) throw new Error('later'); return 1;})} catch(e) {caught=true;value=e} console.log(caught, caught && value===undefined)`).CombinedOutput()
	if err != nil || string(expected) != "true true\n" {
		t.Fatalf("Node: %v %s", err, expected)
	}
	for _, mutant := range []bool{false, true} {
		input := source
		if mutant {
			input = strings.Replace(input, "adamic_exception_pending = true;", "adamic_exception_pending = adamic_thrown != NULL;", 1)
		}
		binary := filepath.Join(t.TempDir(), "program")
		if err := Build(input, binary, Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		for _, threads := range []string{"1", "4"} {
			command := exec.Command(binary)
			command.Env = parallelEnvironment(threads, true)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("mutant=%t threads=%s: %v %s", mutant, threads, err, output)
			}
			if !mutant && string(output) != string(expected) {
				t.Fatalf("threads=%s: %s", threads, output)
			}
			if mutant && string(output) == string(expected) {
				t.Fatalf("pending mutant survived at %s threads", threads)
			}
		}
	}
	t.Log("Node catches null-payload pending erasure with one and four workers; both builds finish sanitizer and leak clean")
}
