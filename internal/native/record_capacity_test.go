package native

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// Nonempty tables are fitted before the first store; an empty literal stays lazy.
// An eight-entry floor in map.c reds 0, 1 and 4; leaving sized maps lazy reds those
// and 16. The literal-0 row stays green for both mutants.
func TestRecordInitialCapacity(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	size_t keys = (size_t)strtoull(argv[1], NULL, 10);
	adamic_record *record = strtol(argv[2], NULL, 10) != 0 ? adamic_record_new(adamic_field_number) : adamic_record_new_capacity(adamic_field_number, keys);
	adamic_map *table = record->slots[0].reference;
	if (table->capacity == 0 && (table->entries != NULL || table->buckets != NULL)) abort();
	printf("%zu %zu\n", table->capacity, table->bucket_count);
	adamic_map *map = adamic_map_new(true, false);
	printf("%zu %zu\n", map->capacity, map->bucket_count);
	adamic_release(map);
	adamic_release(record);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "capacity")
	if err := Build(source, binary, native_runtime.Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name                                string
		keys, entries, buckets, defaultSize int
	}{{"0", 0, 4, 8, 1}, {"literal-0", 0, 0, 0, 0}, {"1", 1, 1, 2, 0}, {"4", 4, 4, 8, 0}, {"16", 16, 16, 32, 0}} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			stdout, stderr, err := recordRun(binary, fmt.Sprint(row.keys), fmt.Sprint(row.defaultSize))
			want := fmt.Sprintf("%d %d\n0 0\n", row.entries, row.buckets)
			if err != nil || stdout != want {
				t.Fatalf("initial table: want %q, got %q: %v\n%s", want, stdout, err, stderr)
			}
			recordCheckCounts(t, stderr)
			if strings.Contains(stderr, "Sanitizer") {
				t.Fatal(stderr)
			}
		})
	}
}
