package native

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// The table is fitted before the first store. Restoring Map's eight-entry start in
// record.c reds the one-key and four-key rows; semantics still grow through map.c.
func TestRecordInitialCapacity(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	size_t keys = (size_t)strtoull(argv[1], NULL, 10);
	adamic_record *record = keys == 0 ? adamic_record_new(adamic_field_number) : adamic_record_new_capacity(adamic_field_number, keys);
	adamic_map *table = record->slots[0].reference;
	printf("%zu %zu\n", table->capacity, table->bucket_count);
	adamic_map *map = adamic_map_new(true, false);
	printf("%zu %zu\n", map->capacity, map->bucket_count);
	adamic_release(map);
	adamic_release(record);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "capacity")
	if err := Build(source, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ keys, entries, buckets int }{{0, 4, 8}, {1, 1, 2}, {4, 4, 8}, {16, 16, 32}} {
		t.Run(fmt.Sprint(row.keys), func(t *testing.T) {
			t.Parallel()
			stdout, stderr, err := recordRun(binary, fmt.Sprint(row.keys))
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
