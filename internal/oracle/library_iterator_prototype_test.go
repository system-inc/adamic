package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Language-level reflection remains refused. This probe holds the runtime's inherited
// method layout to Node directly, without admitting reflection through lowering.
func TestLibraryIteratorPrototypeMatchesNode(t *testing.T) {
	t.Parallel()
	const source = `const a = [].values(), b = [].keys(), m = new Map().keys(), s = new Set().values(), t = ''[Symbol.iterator]();
console.log([Object.hasOwn(a, 'next'), Object.getPrototypeOf(a) === Object.getPrototypeOf(b), a.next === b.next, a.next === m.next, m.next === s.next, a[Symbol.iterator] === m[Symbol.iterator], a[Symbol.iterator]() === a, t[Symbol.iterator]() === t].map(Number).join(' '));`
	const harness = `#include "adamic.h"
#include <stdio.h>
int main(void) {
    adamic_array *array = adamic_array_new(0, false);
    adamic_map *map = adamic_map_new(false, false);
    adamic_map *set = adamic_map_new(false, false);
    static adamic_string empty = ADAMIC_STRING("");
    static adamic_string next = ADAMIC_STRING("next");
    adamic_object *a = adamic_collection_iterator(array, 2, 1, 1, false);
    adamic_object *b = adamic_collection_iterator(array, 1, 1, 1, false);
    adamic_object *m = adamic_collection_iterator(map, 1, 1, 1, false);
    adamic_object *s = adamic_collection_iterator(set, 2, 1, 1, true);
    adamic_object *text = adamic_collection_iterator(&empty, 1, 1, 3, false);
    adamic_value identity = a->shape->methods->code[1](a, NULL);
    adamic_value string_identity = text->shape->methods->code[1](text, NULL);
    printf("%d %d %d %d %d %d %d %d\n", adamic_object_has(a, &next),
        a->shape->methods == b->shape->methods,
        a->shape->methods->code[0] == b->shape->methods->code[0],
        a->shape->methods->code[0] == m->shape->methods->code[0],
        m->shape->methods->code[0] == s->shape->methods->code[0],
        a->shape->methods->code[1] == m->shape->methods->code[1],
        identity.reference == a, string_identity.reference == text);
    adamic_release(identity.reference); adamic_release(string_identity.reference);
    adamic_release(a); adamic_release(b); adamic_release(m); adamic_release(s); adamic_release(text);
    adamic_release(array); adamic_release(map); adamic_release(set);
    return 0;
}`
	binary := filepath.Join(t.TempDir(), "prototype")
	if err := native.Build(harness+"\n", binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("runtime must run cleanly: %d %s", actual.exitCode, actual.stderr)
	}
	expected := executeWith(t, nil, "node", "--input-type=module", "-e", source)
	if difference := disagreement(expected, actual); difference != "" {
		t.Fatalf("Node: %s; native stdout %s", difference, actual.stdout)
	}
}
