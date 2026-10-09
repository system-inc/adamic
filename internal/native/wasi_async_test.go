package native

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A WASI reactor supplies wake/wait hooks instead of a POSIX pipe. Exercise the
// ownership and completion path directly, including the unsupported default.
func TestWASIHostPromises(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" || os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("WASI integration requires ADAMIC_TEST_WASI=1 and WASI_SYSROOT")
	}
	const code = `#include "async.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
static adamic_host_request *request;
static unsigned wakes, waits;
static void wake(void) { wakes++; }
static void wait(void) {
 waits++;
 assert(adamic_host_resolve(request, "ok", 2, 201));
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 if (argc > 1) assert(adamic_host_set_loop_hooks(wake, wait));
 adamic_promise *promise;
 request = adamic_host_promise_new(&promise);
 adamic_async_run();
 assert(wakes == 1 && waits == 1 && promise->settled && !promise->rejected);
 adamic_object *result = promise->value.reference;
 adamic_string *body = result->slots[0].reference;
 assert(body->length == 2 && memcmp(body->bytes, "ok", 2) == 0);
 assert(result->slots[1].number == 201);
 adamic_release(promise);
 puts("host hooks clean");
 return 0;
}
`
	directory := t.TempDir()
	module := filepath.Join(directory, "host.wasm")
	if err := Build(code, module, Options{Target: "wasm32-wasi", Count: true}); err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(repository, "oracle/wasi.mjs")
	control := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", runner, module, "hooks")
	counts := regexp.MustCompile(`^adamic: counts: allocations ([0-9]+) frees ([0-9]+) retains [0-9]+ releases [0-9]+ peak [0-9]+ regions 0\n$`).FindSubmatch(control.stderr)
	if control.exit != 0 || string(control.stdout) != "host hooks clean\n" || len(counts) != 3 || string(counts[1]) == "0" || string(counts[1]) != string(counts[2]) {
		t.Fatalf("custom hooks: %+v", control)
	}
	// Leave the owned promise behind: successful output still matches, while
	// the independent counted build must expose its retained result graph.
	mutant := filepath.Join(directory, "leak.wasm")
	if err := Build(strings.Replace(code, "adamic_release(promise);", "/* missing release */", 1), mutant, Options{Target: "wasm32-wasi", Count: true}); err != nil {
		t.Fatal(err)
	}
	leaked := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", runner, mutant, "hooks")
	leakCounts := regexp.MustCompile(`allocations ([0-9]+) frees ([0-9]+)`).FindSubmatch(leaked.stderr)
	if leaked.exit != 0 || string(leaked.stdout) != string(control.stdout) || len(leakCounts) != 3 || string(leakCounts[1]) == string(leakCounts[2]) {
		t.Fatalf("missing release mutant must fail only the leak count: %+v", leaked)
	}
	t.Logf("WASI host leak control: %s; missing release mutant: %s", control.stderr, leaked.stderr)
	refused := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", runner, module)
	if refused.exit != 70 || len(refused.stdout) != 0 || !strings.Contains(string(refused.stderr), "WASI host promises require wake and wait loop hooks") {
		t.Fatalf("default pipe must be refused clearly: %+v", refused)
	}
}
