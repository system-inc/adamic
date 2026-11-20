package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// A typed array owner adds length times its width to the backing count memoryUsage() reports as
// arrayBuffers, and its free drops the same (#3wfes07). heap.c's free once kept a width table of its
// own that gave a Uint16Array 8 bytes an element, so every one freed took 6 bytes too many from a
// size_t, which wrapped. The harness makes and lets go of an owner of every kind, and a view of one,
// and holds the count to where it began, under AddressSanitizer and the counted build.
//
// Mutant, by hand on Loom: heap.c's old table (uint8 1, int32 4, else 8) reds "uint16", the count
// short by 6 bytes an element, and only that row.
const typedArrayBackingHarness = `#include "adamic.h"
#include "count.h"
#include <stdio.h>

static double backing(void) {
 adamic_host_memory usage;
 if (!adamic_host_memory_usage(&usage)) { return -1; }
 return usage.arrayBuffers;
}

static void row(const char *name, enum adamic_typed_array_kind kind) {
 double before = backing();
 adamic_typed_array *owner = adamic_typed_array_new(kind, 1000);
 double held = backing() - before;
 adamic_typed_array *view = adamic_typed_array_subarray(owner, 10, 20, true);
 adamic_release(owner);
 adamic_release(view);
 printf("%s held %.0f after %.0f\n", name, held, backing() - before);
}

int main(void) {
 row("uint8", adamic_typed_array_uint8);
 row("uint16", adamic_typed_array_uint16);
 row("int32", adamic_typed_array_int32);
 row("float64", adamic_typed_array_float64);
 printf("live %zu\n", (size_t)adamic_counted.live);
 return 0;
}
`

func TestTypedArrayBackingBalances(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "typed-array-backing")
	if err := Build(typedArrayBackingHarness, binary, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	want := "uint8 held 1000 after 0\nuint16 held 2000 after 0\nint32 held 4000 after 0\nfloat64 held 8000 after 0\nlive 0\n"
	if err != nil || stdout.String() != want || strings.Contains(stderr.String(), "Sanitizer") || strings.Contains(stderr.String(), "runtime error") {
		t.Fatalf("typed array backing count: %v\nstdout %q, want %q\n%s", err, stdout.String(), want, stderr.String())
	}
}
