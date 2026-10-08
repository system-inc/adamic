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
// Keep the descriptor used by the C ABI harness within the proven IR layouts.
class Payload { static payload = "unused"; }
console.log(Payload.payload);
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
	if !strings.Contains(generated, "adamic_object_check_data_write(") {
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
    static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_number};
    static const adamic_shape shape = {2, names, references, NULL, kinds};
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
