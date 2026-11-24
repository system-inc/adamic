package native

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// runtimeObjectTagsHarness makes a Map's, a Set's and a string's iterator, reads their shapes through volatile
// pointers, so each comparison is made at run time on the addresses the link gave them, and then asks
// adamic_object_tag and adamic_object_to_string of each, of a WeakMap, a WeakSet and a Promise, and of a program's
// own object and NULL, which have no tag.
const runtimeObjectTagsHarness = `#include "adamic.h"
#include "async.h"
#include <stdio.h>
static void show(const char *name, const void *value) {
 adamic_string *tag = adamic_object_tag(value);
 adamic_string *text = adamic_object_to_string(value);
 printf("%s %s %.*s\n", name, tag == NULL ? "none" : "tag", (int)text->length, text->bytes);
}
int main(void) {
 adamic_stack_thread_start();
 adamic_map *map = adamic_map_new(false, false);
 adamic_object *map_iterator = adamic_collection_iterator(map, 1, adamic_ir_Number, adamic_ir_Number, false);
 adamic_object *set_iterator = adamic_collection_iterator(map, 1, adamic_ir_Number, adamic_ir_Number, true);
 adamic_string *text = adamic_string_from_number(1);
 adamic_object *string_iterator = adamic_string_iterator(text);
 const adamic_shape *volatile shapes[] = {map_iterator->shape, set_iterator->shape, string_iterator->shape};
 int shared = 0;
 for (int first = 0; first < 3; first++) {
  for (int second = first + 1; second < 3; second++) {
   if (shapes[first] == shapes[second]) shared++;
  }
 }
 printf("shared %d\n", shared);
 printf("collection %d%d%d\n", adamic_collection_iterator_is(map_iterator), adamic_collection_iterator_is(set_iterator), adamic_collection_iterator_is(string_iterator));
 show("MapIterator", map_iterator);
 show("SetIterator", set_iterator);
 show("StringIterator", string_iterator);
 adamic_weak_map *weak_map = adamic_weak_map_new(false, false);
 adamic_weak_map *weak_set = adamic_weak_map_new(true, false);
 adamic_async_promise *promise = adamic_async_new();
 show("WeakMap", weak_map);
 show("WeakSet", weak_set);
 show("Promise", promise);
 static const char *const own_names[] = {"next"};
 static const bool own_references[] = {false};
 static const adamic_field_kind own_kinds[] = {adamic_field_number};
 static const adamic_shape own_shape = {1, own_names, own_references, NULL, own_kinds};
 adamic_object *own = adamic_object_new(&own_shape);
 show("own", own);
 show("undefined", NULL);
 adamic_release(own);
 adamic_release(promise);
 adamic_release(weak_set);
 adamic_release(weak_map);
 adamic_release(string_iterator);
 adamic_release(text);
 adamic_release(set_iterator);
 adamic_release(map_iterator);
 adamic_release(map);
 return 0;
}
`

// A Map's, a Set's and a string's iterator are told apart by which of map_set.c's three static shapes they take,
// and the three have the same layout, which is what a linker's or ThinLTO's identical folding merges: merged, a Set's
// iterator would answer a Map's tag. Each names itself past its count so their content differs, and a ThinLTO release
// binary, the one a program ships as, is built and run here, so the three addresses must differ there and each value
// answer its own tag (Runtime's condition 1 on #eksbz4c, head T): by running it, never by reasoning. The rows the
// release binary prints for the six library objects it can make are ir.LibraryObjectTags', so the native answers and
// the Go table can't drift (Runtime's condition 4); a matchAll iterator, which needs a compiled pattern, is held to
// the table by TestRuntimeObjectTagsAreIRs' reading of the source and to Node by the oracle's rows.
func TestRuntimeIteratorShapesStayDistinctInRelease(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "object-tags")
	if err := Build(runtimeObjectTagsHarness, binary, native_runtime.Options{Release: true}); err != nil {
		t.Fatal(err)
	}
	want := "shared 0\ncollection 110\n"
	for _, name := range []string{"MapIterator", "SetIterator", "StringIterator", "WeakMap", "WeakSet", "Promise"} {
		tag := ir.LibraryObjectTag(name)
		if tag == "" {
			t.Fatalf("ir.LibraryObjectTags has no row for %s", name)
		}
		want += fmt.Sprintf("%s tag [object %s]\n", name, tag)
	}
	want += "own none [object Object]\nundefined none [object Object]\n"
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("adamic_object_tag in a ThinLTO release binary:\ngot  %q\nwant %q", got, want)
	}
}

// runtimeObjectTagLiteral is a C string literal that spells an Object.prototype.toString answer.
var runtimeObjectTagLiteral = regexp.MustCompile(`"\[object ([^"\]]+)\]"`)

// adamic_object_tag is the only native spelling of a library tag (Runtime's condition 4): map_set.c's object_tags
// spells exactly ir.LibraryObjectTags' rows, in order, and no other runtime file spells any of them.
func TestRuntimeObjectTagsAreIRs(t *testing.T) {
	t.Parallel()
	directory := filepath.Join("..", "native_runtime", "c")
	source, err := os.ReadFile(filepath.Join(directory, "map_set.c"))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), "static adamic_string object_tags[] = {")
	end := strings.Index(string(source)[max(start, 0):], "};")
	if start < 0 || end < 0 {
		t.Fatal("map_set.c no longer holds object_tags")
	}
	var spelled, want []string
	for _, match := range runtimeObjectTagLiteral.FindAllStringSubmatch(string(source)[start:start+end], -1) {
		spelled = append(spelled, match[1])
	}
	tags := map[string]bool{}
	for _, row := range ir.LibraryObjectTags {
		want = append(want, row.Tag)
		tags[row.Tag] = true
	}
	if !slices.Equal(spelled, want) {
		t.Fatalf("map_set.c's object_tags spells %q, want ir.LibraryObjectTags' %q", spelled, want)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		text, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		found := 0
		for _, match := range runtimeObjectTagLiteral.FindAllStringSubmatch(string(text), -1) {
			if tags[match[1]] {
				found++
			}
		}
		if entry.Name() == "map_set.c" {
			found -= len(want)
		}
		if found != 0 {
			t.Errorf("%s spells a library tag outside adamic_object_tag's object_tags", entry.Name())
		}
	}
}
