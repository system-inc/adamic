package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestFieldStoresMatchNode(t *testing.T) {
	t.Parallel()
	const source = `function store(box: { payload: string }, value: string): void { box.payload = value; }
const box = { payload: 'heap'.repeat(3) };
store(box, box.payload);
console.log(box.payload);
class Counter {
    #value = 0;
    increment(): number { this.#value += 1; return this.#value; }
}
const counter = new Counter();
console.log(counter.increment().toString());
class Base { static inherited = 1; }
class Derived extends Base {}
Base.inherited = 2;
Derived.inherited = 7;
Base.inherited = 3;
console.log(Base.inherited.toString() + ':' + Derived.inherited.toString());
`
	directory := t.TempDir()
	path := filepath.Join(directory, "stores.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	if !strings.Contains(generated, "->frozen)") {
		t.Fatal("store lost its frozen guard")
	}
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, path)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "stores")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("sanitize %t: native %q, Node %q", sanitize, got, want)
		}
	}
	// A borrowed value may have its sole strong owner in the destination slot.
	// Exercise that ABI directly: source argument evaluation normally takes an
	// extra temporary count, which could mask a release-before-retain mutant.
	store := ""
	for index, function := range program.Functions {
		if function.Name == "store" {
			store = (&emitter{program: program}).functionName(index)
		}
	}
	if store == "" {
		t.Fatal("store function missing")
	}
	harness := "#include <stdio.h>\n#define main adamic_original_main\n" + generated + "\n#undef main\n" + `
int main(void) {
    static const char *const names[] = {"payload", "own"};
    static const bool references[] = {true, false};
    static const adamic_shape shape = {2, names, references, NULL};
    static const size_t flags[] = {2, 0};
    static const adamic_class static_class = {.count = 2, .is_static = true, .static_flags = flags};
    adamic_object *box = adamic_object_new(&shape);
    adamic_string seed = {{0, adamic_kind_string, 0}, 4, "heap", 0, NULL, NULL, 0};
    box->slots[0].reference = adamic_string_repeat(&seed, 3);
    ` + store + `(box, box->slots[0].reference);
    box->class = &static_class;
    ` + store + `(box, box->slots[0].reference);
    if (box->slots[1].number != 1) {
        adamic_release(box);
        fputs("static store failed to mark an own property\n", stderr);
        return 1;
    }
    adamic_write_line(adamic_stdout, box->slots[0].reference);
    adamic_release(box);
    return 0;
}
`
	binary := filepath.Join(directory, "borrowed-store")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != "heapheapheap\n" {
		t.Fatalf("sole-owner self store: %q", got)
	}
}

func TestPrimitiveArrayStoresMatchNode(t *testing.T) {
	t.Parallel()
	const source = `function store(array: boolean[], index: number, value: boolean): void { array[index] = value; }
const flags = [false, false];
store(flags, -0, true);
console.log(flags.join(','));
for (const index of [0, 1, 2, -1, 0.5]) {
console.log(String(flags[index] === true) + ':' + String(flags[index] !== true) + ':' + String(flags[index] === false) + ':' + String(flags[index] !== false));
}
const numbers: number[] = [1, 2];
numbers[1] = 3.5;
console.log(numbers.join(','));
const maybe: (number | undefined)[] = [1, undefined];
maybe[0] = undefined;
maybe[1] = 0 / 0;
console.log(maybe.join(','));
const strings = ['old'.repeat(3)];
strings[0] = 'new'.repeat(3);
console.log(strings.join(','));
`
	directory := t.TempDir()
	path := filepath.Join(directory, "arrays.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	generated := C(program)
	oracle, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", "--disable-warning=ExperimentalWarning", oracle, path)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "arrays")
		if err := Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("native %q, Node %q", got, want)
		}
	}
	store := ""
	for index, function := range program.Functions {
		if function.Name == "store" {
			store = (&emitter{program: program}).functionName(index)
		}
	}
	// The fallback preserves Adamic's checked-write contract, including fractional,
	// negative, NaN and infinite indices. Compare it with the existing runtime path.
	for _, index := range []string{"-1.0", "1.0", "0.5", "NAN", "INFINITY"} {
		var results []ran
		for _, call := range []string{store + "(array, " + index + ", true)", "adamic_array_set(array, " + index + ", (adamic_value){.boolean=true})"} {
			harness := "#define main adamic_original_main\n" + generated + "\n#undef main\nint main(void) { adamic_array *array=adamic_array_new(1,false); adamic_array_push(array,(adamic_value){.boolean=false}); " + call + "; adamic_release(array); return 0; }\n"
			binary := filepath.Join(directory, "bounds")
			if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			results = append(results, outcome(t, binary))
		}
		if results[0] != results[1] || results[0].exitCode != 70 {
			t.Fatalf("index %s: emitted %+v, runtime %+v", index, results[0], results[1])
		}
	}
}

func TestColdAllocationFailure(t *testing.T) {
	t.Parallel()
	// Fail a real allocation after output has been buffered. The cold path must
	// flush that output and preserve the runtime's message and exit code.
	const source = `#include "adamic.h"
#include <stdlib.h>
void *malloc(size_t size) { (void)size; return NULL; }
int main(void) {
    static adamic_string before = ADAMIC_STRING("before");
    adamic_write_line(adamic_stdout, &before);
    (void)adamic_allocate(300, adamic_kind_object);
    return 0;
}
`
	binary := filepath.Join(t.TempDir(), "allocation-failure")
	if err := Build(source, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	got := outcome(t, binary)
	if got.stdout != "before\n" || got.stderr != "adamic: panic: out of memory\n" || got.exitCode != 70 {
		t.Fatalf("allocation failure: %+v", got)
	}
}
