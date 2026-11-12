package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Adoption moves an allocation and frees the original, so it must never take memory the heap didn't
// allocate (#yzpahn3, Runtime's guard on #jawqwv7). The stdout handle (node_process.c) is a host
// object built in static storage with count zero. Graph adoption refuses it by its count, and
// adamic_heap_graph_storage, which both adoptions move through, refuses it itself, so a later caller
// can't free static storage. Mutant: the storage guard removed sends "storage" to AddressSanitizer's
// bad free instead of the named stop.
func TestGraphAdoptionRefusesStaticSource(t *testing.T) {
	t.Parallel()
	const main = `#include "node_process.h"
int main(int count, char **arguments) {
 (void)count;
 adamic_object *handle = adamic_node_stdout_handle();
 if (handle == NULL) { puts("no handle"); return 2; }
 if (strcmp(arguments[1], "graph") == 0) {
  adamic_graph_adopt(handle, adamic_object_size(1));
  return 1;
 }
 adamic_heap_graph_storage(handle, adamic_object_size(1));
 return 1;
}
`
	source := graphHarness[:strings.Index(graphHarness, "__attribute__")] + main
	binary := filepath.Join(t.TempDir(), "static-source")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for mode, want := range map[string]string{
		"graph":   "compiler bug: graph adoption needs a new allocation",
		"storage": "runtime invariant broken: adoption would move an immortal or static value",
	} {
		command := exec.Command(binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := command.CombinedOutput()
		if err == nil || !strings.Contains(string(output), want) {
			t.Fatalf("%s: want %q: %v\n%s", mode, want, err, output)
		}
		if strings.Contains(string(output), "Sanitizer") {
			t.Fatalf("%s stopped by a sanitizer, not by name:\n%s", mode, output)
		}
	}
}
