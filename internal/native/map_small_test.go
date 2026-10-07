package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
)

// Not parallel: RSS and timing observations must not compete with another performance gate.
func TestMapSmallProfile(t *testing.T) {
	source, err := os.ReadFile("testdata/map_small_profile.c")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "profile")
	if err := Build(string(source), binary, Options{}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		// The extra fork after exec avoids inheriting the Go process's RSS high-water mark.
		output, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "profile", binary).CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", err, output)
		}
		peak := mapSmallMeasurements(t, string(output))
		if peak > mapSmallResidentBound {
			t.Fatalf("profile peak %d KiB exceeds inline storage gate %d KiB", peak, mapSmallResidentBound)
		}
		t.Logf("trial %d: %s; normalized RSS %d KiB", trial+1, strings.TrimSpace(string(output)), peak)
	}
}

const mapSmallResidentBound = 48 * 1024

// Not parallel: independently proves the RSS gate can reject disabling inline storage.
func TestMapSmallResidentMutant(t *testing.T) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	changed := false
	for _, file := range files {
		source := string(file.contents)
		if file.name == "map.c" {
			before := source
			source = strings.Replace(source, "map->capacity = 4;\n\tmap->entries = map->small;", "map->capacity = 0;\n\tmap->entries = NULL;", 1)
			changed = source != before
		}
		if err := os.WriteFile(filepath.Join(directory, file.name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("mutant changed no code")
	}
	library, err := RuntimeLibrary(directory, Options{})
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/map_small_profile.c")
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, source, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "profile")
	arguments := append(Flags(Options{}), "-I", filepath.Dir(library), "-o", binary, main)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	output, err := exec.Command("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "profile", binary).CombinedOutput()
	if err != nil {
		t.Fatalf("mutant must finish: %v: %s", err, output)
	}
	peak := mapSmallMeasurements(t, string(output))
	if peak <= mapSmallResidentBound {
		t.Fatalf("RSS gate failed to catch mutant: %d KiB <= %d KiB", peak, mapSmallResidentBound)
	}
	t.Logf("disabled inline storage: clean exit 0, RSS %d KiB exceeds %d KiB gate", peak, mapSmallResidentBound)
}

func mapSmallMeasurements(t *testing.T, output string) int {
	t.Helper()
	var maps, live, sets, gets, has, rss int
	var checksum, cpu, wall float64
	if _, err := fmt.Sscanf(output, "maps=%d live=%d set=%d get=%d has=%d checksum=%f cpu=%f wall=%f rss=%d", &maps, &live, &sets, &gets, &has, &checksum, &cpu, &wall, &rss); err != nil {
		t.Fatal(err)
	}
	if maps != 271022 || live != 143047 || sets != 1201980 || gets != 4429860 || has != 318504 || checksum != 5269033 {
		t.Fatalf("profile workload changed: %s", output)
	}
	return residentKibibytes(rss, goruntime.GOOS)
}

func TestMapSmallStorage(t *testing.T) {
	const source = `#include "adamic.h"
#include <stdio.h>

static bool inline_map(const adamic_map *map) {
    return map->entries == map->small && map->capacity == 4 && map->buckets == NULL && map->bucket_count == 0;
}
int main(void) {
    adamic_map *map = adamic_map_new(false, false);
    if (!inline_map(map) || map->count != 0) { return 2; }
    for (int i = 0; i < 4; i++) {
        adamic_map_set(map, (adamic_value){.number = i}, (adamic_value){.number = i});
        if (!inline_map(map)) { return 3; }
    }
    adamic_map_set(map, (adamic_value){.number = 0}, (adamic_value){.number = 9});
    if (!inline_map(map) || map->count != 4) { return 4; }
    adamic_map_set(map, (adamic_value){.number = 4}, (adamic_value){.number = 4});
    if (inline_map(map) || map->count != 5 || map->bucket_count == 0) { return 5; }
    adamic_map_iterator *first = adamic_map_iterate(map);
    adamic_map_iterator *second = adamic_map_iterate(map);
    adamic_map_delete(map, (adamic_value){.number = 0});
    adamic_map_delete(map, (adamic_value){.number = 1});
    if (inline_map(map) || map->iterating != 2) { return 6; }
    adamic_value key, value;
    while (adamic_map_iterator_next(first, &key, &value)) {}
    if (inline_map(map) || map->iterating != 1) { return 7; }
    adamic_release(second);
    if (!inline_map(map) || map->iterating != 0 || map->used != 3) { return 8; }
    if (adamic_map_iterator_next(first, &key, &value) || map->iterating != 0) { return 9; }
    adamic_release(first);
    if (map->iterating != 0) { return 10; }
    adamic_map_set(map, (adamic_value){.number = 5}, (adamic_value){.number = 5});
    adamic_map_set(map, (adamic_value){.number = 6}, (adamic_value){.number = 6});
    adamic_map_delete(map, (adamic_value){.number = 2});
    if (!inline_map(map) || map->count != 4) { return 11; }
    adamic_map_clear(map);
    if (!inline_map(map) || map->count != 0 || map->used != 0) { return 12; }
    // A tiny live population can need more historical slots while an iterator is open.
    adamic_map_set(map, (adamic_value){.number = 0}, (adamic_value){.number = 0});
    first = adamic_map_iterate(map);
    for (int i = 1; i < 10; i++) {
        adamic_map_delete(map, (adamic_value){.number = i - 1});
        adamic_map_set(map, (adamic_value){.number = i}, (adamic_value){.number = i});
    }
    if (inline_map(map) || map->count != 1) { return 13; }
    adamic_release(first);
    if (!inline_map(map) || map->used != 1 || map->iterating != 0) { return 14; }
    adamic_release(map);
    puts("small storage and iteration counts hold");
    return 0;
}
`
	for _, sanitize := range []bool{true, false} {
		binary := filepath.Join(t.TempDir(), "storage")
		if err := Build(source, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("sanitized %v: %v: %s", sanitize, err, output)
		}
		if string(output) != "small storage and iteration counts hold\n" {
			t.Fatalf("unexpected output: %s", output)
		}
	}
}
