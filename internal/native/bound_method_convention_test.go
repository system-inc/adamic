package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
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

// Not parallel: uses the WASI toolchain. Only closure.c changes conventions;
// class lookup and method tables retain the generated program's convention.
func TestWASICountedBoundMethodConvention(t *testing.T) {
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1")
	}
	checked, err := load.Load([]string{filepath.Join("testdata", "bound_counted_method.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	counted := false
	for index, function := range program.Functions {
		if function.MethodName == "m" && program.PackedCountNeeded(index) && !function.ReadsArguments && function.RestElement == 0 {
			counted = true
		}
	}
	if !counted {
		t.Fatal("the bound method must require a count without reading arguments or collecting rest")
	}
	code := C(program)
	const expected = "number:7\nboolean:false\n"
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	truth := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(repository, "internal/native/testdata/bound_counted_method.a"))
	if truth.exit != 0 || string(truth.stdout) != expected || len(truth.stderr) != 0 {
		t.Fatalf("source Node: exit=%d stdout=%q stderr=%q", truth.exit, truth.stdout, truth.stderr)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := Build(code, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := observeWASI(t, repository, binary)
	if actual.exit != 0 || string(actual.stdout) != expected || len(actual.stderr) != 0 {
		t.Fatalf("native bound method: exit=%d stdout=%q stderr=%q", actual.exit, actual.stdout, actual.stderr)
	}
	options := Options{Target: "wasm32-wasi"}
	library, err := RuntimeLibraryForSource("", code, options)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Dir(library)
	closure, err := os.ReadFile(filepath.Join(directory, "closure.c"))
	if err != nil {
		t.Fatal(err)
	}
	const dispatch = "  if (self->bound_method_storage.counted) {\n   return ((adamic_counted_method)self->bound_method_storage.code)(self->cells[1]->value.reference, arguments, argument_count);\n  }\n"
	if strings.Count(string(closure), dispatch) != 1 {
		t.Fatal("counted dispatch mutation site moved")
	}
	for _, mutant := range []bool{false, true} {
		name := "fixed"
		runtimeSource := string(closure)
		if mutant {
			name = "ignored-tag"
			runtimeSource = strings.Replace(runtimeSource, dispatch, "", 1)
		}
		t.Run(name, func(t *testing.T) {
			scratch := t.TempDir()
			// Cached headers stamp the archive's features before their include guard.
			// Copy them together so a nested include cannot turn the convention back on.
			headers, err := filepath.Glob(filepath.Join(directory, "*.h"))
			if err != nil {
				t.Fatal(err)
			}
			for _, header := range headers {
				contents, err := os.ReadFile(header)
				if err != nil {
					t.Fatal(err)
				}
				clean := strings.ReplaceAll(string(contents), "#define ADAMIC_CLOSURE_CONVENTION 1\n", "")
				writeWASIFile(t, filepath.Join(scratch, filepath.Base(header)), clean)
			}
			plain := filepath.Join(scratch, "closure.c")
			object := filepath.Join(scratch, "closure.o")
			writeWASIFile(t, plain, runtimeSource)
			flags := append(wasiFlags(os.Getenv("WASI_SYSROOT")), featureFlags(code)...)
			flags = append(flags, "-UADAMIC_CLOSURE_CONVENTION", "-I", scratch, "-c", plain, "-o", object)
			if output, err := toolchain.Command(wasiCompiler(), flags...).CombinedOutput(); err != nil {
				t.Fatalf("plain runtime compile: %v\n%s", err, output)
			}
			source, module := filepath.Join(scratch, "main.c"), filepath.Join(scratch, "main.wasm")
			writeWASIFile(t, source, code)
			arguments := append(Flags(options), "-I", directory, "-Wl,--export=__stack_low", source, object, library, "-lm", "-o", module)
			if output, err := toolchain.Command(wasiCompiler(), arguments...).CombinedOutput(); err != nil {
				t.Fatalf("WASI link: %v\n%s", err, output)
			}
			actual := observeWASI(t, repository, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "internal/native/wasm/run.mjs"), module)
			if mutant {
				if actual.exit == 0 || !strings.Contains(string(actual.stderr), "null function or function signature mismatch") {
					t.Fatalf("ignored tag escaped the wasm signature check: exit=%d stdout=%q stderr=%q", actual.exit, actual.stdout, actual.stderr)
				}
				t.Log("ignored tag compiled, then trapped on the counted method signature")
			} else if actual.exit != 0 || string(actual.stdout) != expected || len(actual.stderr) != 0 {
				t.Fatalf("mixed counted bound method: exit=%d stdout=%q stderr=%q", actual.exit, actual.stdout, actual.stderr)
			}
		})
	}
}
