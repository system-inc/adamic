package native

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: uses the WASI toolchain and the strict runtime product.
// The producer uses the counted entry representation while closure.c uses the
// plain convention. The ordinary oracle builds both sides with matching flags.
func TestWASIBoundMethodConvention(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1")
	}
	directory, objects := strictWASIRuntimeVariant(t, strictWASIRuntime(t), "plain")
	scratch := t.TempDir()
	source := filepath.Join(scratch, "main.c")
	writeWASIFile(t, source, `#define ADAMIC_CLOSURE_CONVENTION 1
#include "adamic.h"
static adamic_value method(adamic_object *self, adamic_value *arguments) {
 (void)arguments;
 return (adamic_value){.number = 41 + (double)self->heap.references};
}
static adamic_value adapter(adamic_closure *self, adamic_value *arguments, size_t count) {
 return adamic_bound_closure_call(self, arguments, count, 0);
}
int main(void) {
 adamic_object object = {.heap = {.references = 1}};
 adamic_cell target = {0}, receiver = {.value = {.reference = &object}};
 adamic_closure *bound = adamic_counted_bound_closure_new(adapter, 2);
 bound->cells[0] = &target;
 bound->cells[1] = &receiver;
 bound->bound_method = (adamic_method_entry){.code = method};
 printf("%.0f\n", adamic_closure_call(bound, NULL, 0).number);
 adamic_release(bound);
 adamic_heap_end();
 return 0;
}
`)
	module := filepath.Join(scratch, "main.wasm")
	arguments := append(wasiFlags(os.Getenv("WASI_SYSROOT")), "-I", directory, "-Wl,--export=__stack_low", source, "-o", module)
	arguments = append(arguments, objects...)
	arguments = append(arguments, "-lm")
	if output, err := toolchain.Command(wasiCompiler(), arguments...).CombinedOutput(); err != nil {
		t.Fatalf("WASI link: %v\n%s", err, output)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	actual := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "internal/native/wasm/run.mjs"), module)
	if actual.exit != 0 || string(actual.stdout) != "42\n" || len(actual.stderr) != 0 {
		t.Fatalf("mixed convention bound method: exit=%d stdout=%q stderr=%q", actual.exit, actual.stdout, actual.stderr)
	}
}
