package native

import (
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
)

// Not parallel: the RSS comparison and recorded CPU bound are performance gates.
func TestLibraryMapSetIteratorResources(t *testing.T) {
	const harness = `#include "adamic.h"
#include <stdio.h>
#include <string.h>
#include <sys/resource.h>
#include <time.h>

int main(int count, char **arguments) {
	(void)count;
	adamic_map *map = adamic_map_new(false, false);
	adamic_value key, value;
	if (strcmp(arguments[1], "speed") == 0) {
		for (int i = 0; i < 1000; i++) {
			adamic_map_set(map, (adamic_value){.number = i}, (adamic_value){.number = i});
		}
		clock_t start = clock();
		double total = 0;
		for (int round = 0; round < 2000; round++) {
			for (int i = 0; i < 1000; i++) {
				total += adamic_map_get(map, (adamic_value){.number = i})->number;
			}
		}
		if (total != 999000000) { return 2; }
		printf("%.9f\n", (double)(clock() - start) / CLOCKS_PER_SEC);
	} else {
		bool check = strcmp(arguments[1], "count") == 0;
		bool held = check || strcmp(arguments[1], "held") == 0;
		adamic_map_iterator *iterator = held ? adamic_map_iterate(map) : NULL;
		if (held) {
			if (adamic_map_iterator_next(iterator, &key, &value)) { return 3; }
			// Calling next again must not end another iterator's active iteration.
			adamic_map_iterator *active = adamic_map_iterate(map);
			if (adamic_map_iterator_next(iterator, &key, &value)) { return 4; }
			if (check && map->iterating != 1) { return 5; }
			adamic_release(active);
		}
		for (int i = 0; i < 4000000; i++) {
			adamic_map_set(map, (adamic_value){.number = i}, (adamic_value){.number = i});
			if (!adamic_map_delete(map, (adamic_value){.number = i})) { return 6; }
		}
		if (held) {
			if (adamic_map_iterator_next(iterator, &key, &value)) { return 7; }
			adamic_release(iterator);
			if (check && map->iterating != 0) { return 8; }
		}
		struct rusage usage;
		getrusage(RUSAGE_SELF, &usage);
		printf("%ld\n", usage.ru_maxrss);
	}
	adamic_release(map);
	return 0;
}
`
	binary := filepath.Join(t.TempDir(), "resources")
	if err := Build(harness, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	run := func(t *testing.T, mode string) string {
		t.Helper()
		// Fork after exec to avoid inheriting the Go process's resident high-water mark.
		output, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "resources", binary, mode).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", mode, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	t.Run("peak", func(t *testing.T) {
		peak := func(mode string) int {
			value, err := strconv.Atoi(run(t, mode))
			if err != nil {
				t.Fatal(err)
			}
			return residentKibibytes(value, goruntime.GOOS)
		}
		control, held := peak("control"), peak("held")
		// Allow 4 MiB of startup noise, far below the 200+ MiB tombstone regression.
		if held > control+4096 {
			t.Errorf("held exhausted iterator peak %d KiB, control %d KiB, allowance 4096 KiB", held, control)
		}
		t.Logf("held exhausted iterator %d KiB, control %d KiB", held, control)
	})
	t.Run("count", func(t *testing.T) { run(t, "count") })
	t.Run("speed", func(t *testing.T) {
		seconds, err := strconv.ParseFloat(run(t, "speed"), 64)
		if err != nil {
			t.Fatal(err)
		}
		// Recorded gate: 2 million successful 0..999 lookups take at most 0.35 CPU seconds.
		// CPU time excludes scheduling delays. The all-zero hash mutant must cross this bound.
		if seconds > 0.35 {
			t.Errorf("numeric lookups took %.6f CPU seconds, bound 0.35", seconds)
		}
		t.Logf("2 million numeric lookups %.6f CPU seconds, bound 0.35", seconds)
	})
}
