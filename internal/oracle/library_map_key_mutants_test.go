package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/testsupport"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mapKeyMutation struct{ name, kind, before, after string }

var mapKeyMutations = []mapKeyMutation{

	{"object shape equality", "objects", "default:\n\t\treturn left == right;", "default:\n\t\treturn true;"},
	{"class shape equality", "classes", "default:\n\t\treturn left == right;", "default:\n\t\treturn true;"},
	{"number NaN unequal", "numbers", "return left.number == right.number || (isnan(left.number) && isnan(right.number));", "return left.number == right.number;"},
	{"string identity equality", "strings", "return adamic_string_equal(left.reference, right.reference);", "return left.reference == right.reference;"},
	{"mixed NaN unequal", "mixed", "return a == b || (isnan(a) && isnan(b));", "return a == b;"},
	{"mixed numeric identity hash", "mixed", "if (reference != NULL && reference->kind == adamic_kind_number) {", "if (false && reference != NULL && reference->kind == adamic_kind_number) {"},
	{"mixed string identity hash", "mixed", "if (reference != NULL && reference->kind == adamic_kind_string) {", "if (false && reference != NULL && reference->kind == adamic_kind_string) {"},
	{"mixed zero not normalized", "mixed", "adamic_heap *zero = adamic_box_number(0);", "adamic_heap *zero = adamic_box_number(number);"},
	{"readded entry order", "strings", "map->used++;\n\tmap->count++;", "map->used++;\n\tmap->count++;\n\tif (map->used > map->count) {\n\t\tsize_t first = 0, last = map->used - 1;\n\t\twhile (map->entries[first].deleted) first++;\n\t\tadamic_map_entry swap = map->entries[first];\n\t\tmap->entries[first] = map->entries[last];\n\t\tmap->entries[last] = swap;\n\t\tfor (size_t bucket = 0; bucket < map->bucket_count; bucket++) {\n\t\t\tif (map->buckets[bucket] == first + 1) map->buckets[bucket] = last + 1;\n\t\t\telse if (map->buckets[bucket] == last + 1) map->buckets[bucket] = first + 1;\n\t\t}\n\t}"},
}

func TestMapKeyObjectIdentityNodeMutant(t *testing.T)  { t.Parallel(); mapKeyNodeMutant(t, 0) }
func TestMapKeyClassIdentityNodeMutant(t *testing.T)   { t.Parallel(); mapKeyNodeMutant(t, 1) }
func TestMapKeyNumberNaNNodeMutant(t *testing.T)       { t.Parallel(); mapKeyNodeMutant(t, 2) }
func TestMapKeyStringEqualityNodeMutant(t *testing.T)  { t.Parallel(); mapKeyNodeMutant(t, 3) }
func TestMapKeyMixedNaNNodeMutant(t *testing.T)        { t.Parallel(); mapKeyNodeMutant(t, 4) }
func TestMapKeyMixedNumberHashNodeMutant(t *testing.T) { t.Parallel(); mapKeyNodeMutant(t, 5) }
func TestMapKeyMixedStringHashNodeMutant(t *testing.T) { t.Parallel(); mapKeyNodeMutant(t, 6) }
func TestMapKeyMixedZeroNodeMutant(t *testing.T)       { t.Parallel(); mapKeyNodeMutant(t, 7) }
func TestMapKeyReinsertOrderNodeMutant(t *testing.T)   { t.Parallel(); mapKeyNodeMutant(t, 8) }

func mapKeyNodeMutant(t *testing.T, index int) {
	t.Helper()
	mutation := mapKeyMutations[index]
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/map_set_keys/library_map_set_keys_"+mutation.kind+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/map.c"))
	if err != nil {
		t.Fatal(err)
	}
	before := string(contents)
	after := strings.Replace(before, mutation.before, mutation.after, 1)
	if before == after {
		t.Fatal("mutant changed no runtime code")
	}
	mapRuntimeNodeOnly(t, native.C(program), after, path)
}
func TestMapSmallNaNNodeMutant(t *testing.T)  { t.Parallel(); mapSmallNodeMutant(t, "NaN") }
func TestMapSmallZeroNodeMutant(t *testing.T) { t.Parallel(); mapSmallNodeMutant(t, "Zero") }
func TestMapSmallPromotionOrderNodeMutant(t *testing.T) {
	t.Parallel()
	mapSmallNodeMutant(t, "PromotionOrder")
}
func TestMapSmallPromotionIteratorNodeMutant(t *testing.T) {
	t.Parallel()
	mapSmallNodeMutant(t, "PromotionIterator")
}

func mapSmallNodeMutant(t *testing.T, name string) {
	t.Helper()
	fixture := "numbers"
	if name == "PromotionOrder" {
		fixture = "order"
	}
	if name == "PromotionIterator" {
		fixture = "live"
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/map_set_keys/library_map_small_"+fixture+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/map.c"))
	if err != nil {
		t.Fatal(err)
	}
	before := string(contents)
	var after string
	switch name {
	case "NaN":
		after = strings.Replace(before, "same_key(map, entry->key, key)", "(!map->reference_keys && !map->boolean_keys && !map->maybe_number_keys ? entry->key.number == key.number : same_key(map, entry->key, key))", 1)
	case "Zero":
		after = strings.Replace(before, "same_key(map, entry->key, key)", "(!map->reference_keys && !map->boolean_keys && !map->maybe_number_keys ? memcmp(&entry->key.number, &key.number, sizeof key.number) == 0 : same_key(map, entry->key, key))", 1)
	case "PromotionOrder":
		after = strings.Replace(before, "memcpy(grown, map->entries, map->used * sizeof *grown);", `memcpy(grown, map->entries, map->used * sizeof *grown);
 if (map->used > 1) { adamic_map_entry swap = grown[0]; grown[0] = grown[1]; grown[1] = swap; }`, 1)
	case "PromotionIterator":
		after = strings.Replace(before, "// used is read each time", "if (iterator->map->bucket_count != 0 && iterator->next == 1) { iterator->next++; }\n\t// used is read each time", 1)
	}
	if before == after {
		t.Fatal("mutant changed no runtime code")
	}
	mapRuntimeNodeOnly(t, native.C(program), after, path)
}

// Rebuild only the changed translation unit against the same headers and flags.
// Every unrelated object comes from the cached, correctly featured runtime archive.
func mapRuntimeNodeOnly(t *testing.T, source, changed, path string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := testsupport.BuildMapMutant(source, changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	truth := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || truth.exitCode != 0 || len(truth.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("mutant failed outside Node stdout: truth %+v got %+v", truth, got)
	}
	t.Log("Node stdout alone catches mutation; exit 0, empty stderr, ASan/UBSan/LSan clean")
}
