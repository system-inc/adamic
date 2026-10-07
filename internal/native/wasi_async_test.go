package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A WASI reactor supplies wake/wait hooks instead of a POSIX pipe. Exercise the
// ownership and completion path directly, including the unsupported default.
func TestWASIHostPromises(t *testing.T) {
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
	if err := Build(code, module, Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(repository, "oracle/wasi.mjs")
	control := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", runner, module, "hooks")
	if control.exit != 0 || string(control.stdout) != "host hooks clean\n" || len(control.stderr) != 0 {
		t.Fatalf("custom hooks: %+v", control)
	}
	refused := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", runner, module)
	if refused.exit != 70 || len(refused.stdout) != 0 || !strings.Contains(string(refused.stderr), "WASI host promises require wake and wait loop hooks") {
		t.Fatalf("default pipe must be refused clearly: %+v", refused)
	}
}
