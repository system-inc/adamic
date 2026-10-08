package native

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Count dispatches independently of the ownership ledger. A restored slow path must
// fail this performance control even though it preserves the program's answer.
func TestScout36FastPathDispatch(t *testing.T) {
	const source = `#include "adamic.h"
#include "count.h"
#include <stdio.h>
size_t scout_slow_retains, scout_slow_releases;
int main(void) {
 static adamic_string keyword = ADAMIC_STRING("return");
 adamic_retain(NULL); adamic_release(NULL);
 for(size_t i = 0; i < 32; i++) {adamic_retain(&keyword); adamic_release(&keyword);}
 if(adamic_counted.retains != 33 || adamic_counted.releases != 33) {return 2;}
 if(scout_slow_retains != 0 || scout_slow_releases != 0) {
  fprintf(stderr,"immortal/null slow dispatch: %zu/%zu\n",scout_slow_retains,scout_slow_releases); return 7;
 }
 puts("zero slow dispatch; all calls counted");
 return 0;
}
`
	for _, mutant := range []bool{false, true} {
		edit := func(name, data string) string {
			if name == "heap.c" {
				data = "#include <stddef.h>\nextern size_t scout_slow_retains, scout_slow_releases;\n" + data
				for _, a := range []struct{ from, to string }{
					{"void *adamic_retain_slow(void *value) {", "void *adamic_retain_slow(void *value) {\n scout_slow_retains++;"},
					{"void adamic_release_slow(void *value) {", "void adamic_release_slow(void *value) {\n scout_slow_releases++;"},
				} {
					if strings.Count(data, a.from) != 1 {
						t.Fatal("dispatch probe anchor changed")
					}
					data = strings.Replace(data, a.from, a.to, 1)
				}
			}
			if mutant && name == "adamic.h" {
				for _, line := range []string{"if (heap->kind != adamic_kind_cell) { return value; }", "if (count == 0 && heap->kind != adamic_kind_cell) { return; }"} {
					if strings.Count(data, line) != 1 {
						t.Fatal("slow-path mutant anchor changed")
					}
					data = strings.Replace(data, line, "/* mutant: restore zero-count slow dispatch */", 1)
				}
			}
			return data
		}
		binary := buildEdited(t, source, Options{Sanitize: true, Count: true}, edit)
		cmd := exec.Command(binary)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := cmd.CombinedOutput()
		if !mutant && (err != nil || !bytes.Contains(output, []byte("zero slow dispatch; all calls counted"))) {
			t.Fatalf("baseline: %v\n%s", err, output)
		}
		if mutant {
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 7 || !bytes.Contains(output, []byte("immortal/null slow dispatch: 32/32")) {
				t.Fatalf("restored slow path survived: %v\n%s", err, output)
			}
			t.Log("restored zero-count slow dispatch rejected (32 retains/32 releases)")
		}
	}
}

// The first shape must not turn an environment's zero-count interior cell immortal.
func TestScout36OwnedCellMutant(t *testing.T) {
	const source = `#include "adamic.h"
int main(void) {
 adamic_environment *environment = adamic_environment_new(1);
 adamic_cell *cell = &environment->cells[0];
 adamic_retain(cell);
 adamic_release(environment);
 adamic_release(cell);
 return 0;
}
`
	for _, mutant := range []bool{false, true} {
		var edit func(string, string) string
		if mutant {
			edit = func(name, data string) string {
				if name != "adamic.h" {
					return data
				}
				if strings.Count(data, "heap->kind != adamic_kind_cell") != 2 {
					t.Fatal("owned-cell mutant anchor changed")
				}
				return strings.ReplaceAll(data, "heap->kind != adamic_kind_cell", "true")
			}
		}
		binary := buildEdited(t, source, Options{Sanitize: true}, edit)
		cmd := exec.Command(binary)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := cmd.CombinedOutput()
		if !mutant && err != nil {
			t.Fatalf("owned cell: %v\n%s", err, output)
		}
		if mutant && (err == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-use-after-free"))) {
			t.Fatalf("owned-cell mutant survived: %v\n%s", err, output)
		}
	}
	t.Log("counting the environment preserves its interior cell; skipping it fails ASan")
}
