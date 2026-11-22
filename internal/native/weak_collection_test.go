package native_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// WeakMap and WeakSet's runtime half (weak_collection.c, weak.c): a collection keyed by its keys' weak
// handles, whose entries go when a key is freed and whose values go with them. These drive the C
// contract Library builds on directly, judged by the counted build (allocations equal frees, and its
// own line of weak entries, "adamic: weak entries: live N peak P") and by ASan, with the size-class
// allocator on and off.
//
// Mutants, each red here:
//   - adamic_weak_map_store retains the key's target as well as its handle: "freed" stops at its first
//     turn, the freed key's handle still naming it, and the counted build's live entries and heap
//     values grow with the loop. It's the mode that frees a key under a live collection every turn; the
//     set and the free orders free a key too and go red with it, by their own checks.
//   - the collection keys its entries by the target's address, not the handle (entry_of hashing and
//     comparing reveal(key->target)): "reuse" goes red. Forget clears the target before it removes the
//     entries, so it no longer finds the entry it was told is there and stops on its invariant ("a weak
//     handle named a collection without its entry"); "freed" stops the same way, but only "reuse" puts a
//     new object at the freed key's address and asks the map about it.
//   - adamic_weak_map_free_children skips taking its names off its keys' handles: "map first" frees the
//     collection while plain Weak handles keep its keys' handles alive, so the keys' later frees reach the
//     freed collection through the names, which ASan reports as a use after free. "keys first" can't
//     catch it: the keys' frees remove every entry and name before the collection goes.
//   - adamic_weak_key leaves the named bit clear while free_one stays gated on it: "freed" goes red at
//     its first turn, since the freed key is never forgotten and its handle still names it.
//   - deallocate masks the slab without ADAMIC_WEAK_NAMED: "freed" goes red at its first freed key in
//     every build, the bit read as part of the chunk number; with slabs on, ASan reports the read past
//     the chunk table's pages.
//   - adamic_weak_map_store takes a region's value: "region value" exits 0 instead of panicking.
//   - share.c's walk passes a collection, or a key a collection names: "share map" or "share key" exits
//     0. adamic_weak_named answering true for any handle: "share removed", a key whose entry was
//     removed, panics where it must not.
const weakCollectionHarness = `#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const adamic_shape shape = {0, NULL, NULL, NULL, NULL};

static adamic_object *object(void) { return adamic_object_new(&shape); }

// live_is checks the entries every collection holds, where the counted build keeps them; elsewhere
// each mode asks a handle it holds instead.
static bool live_is(size_t expected) {
#ifdef ADAMIC_COUNT
 return adamic_weak_entries_live() == expected;
#else
 (void)expected;
 return true;
#endif
}

static int fail(const char *what, long at) {
 printf("%s at %ld\n", what, at);
 return 4;
}

static adamic_value reference(void *value) { return (adamic_value){.reference = value}; }

int main(int count, char **arguments) {
 if (count != 3) { return 2; }
 const char *mode = arguments[1];
 long loops = strtol(arguments[2], NULL, 10);
 if (strcmp(mode, "freed") == 0) {
  // One long-lived map, and each turn a key and a value stored into it and the key let go of.
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  for (long i = 0; i < loops; i++) {
   adamic_object *key = object(), *value = object();
   adamic_weak *handle = adamic_weak_key(key);
   // Held past the key, so the handle can be asked what's left once the key is freed.
   adamic_retain(handle);
   adamic_weak_map_store(map, handle, reference(value));
   adamic_value found = {0};
   if (adamic_weak_key_find(key) != handle || !adamic_weak_map_lookup(map, handle, &found) || found.reference != value) { return fail("the stored entry wasn't found", i); }
   if (!live_is(1)) { return fail("one entry stored, another count live", i); }
   adamic_release(key);
   if (adamic_weak_target(handle) != NULL) { return fail("the freed key's handle still names it", i); }
   if (adamic_weak_map_lookup(map, handle, NULL)) { return fail("the freed key's entry is still in the map", i); }
   if (!live_is(0)) { return fail("entries live after the key was freed", i); }
   adamic_release(handle);
  }
  adamic_release(map);
  puts("freed");
 } else if (strcmp(mode, "replace") == 0) {
  // A second store for one key replaces the value and lets go of the old one and the extra handle
  // count; a remove lets go of the rest, and a second remove finds nothing.
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_object *key = object(), *first = object(), *second = object();
  adamic_weak_map_store(map, adamic_weak_key(key), reference(first));
  adamic_weak_map_store(map, adamic_weak_key(key), reference(second));
  adamic_value found = {0};
  if (!adamic_weak_map_lookup(map, adamic_weak_key_find(key), &found) || found.reference != second) { return fail("the replaced value was read", 0); }
  if (!live_is(1)) { return fail("a replaced entry counted twice", 0); }
  if (!adamic_weak_map_remove(map, adamic_weak_key_find(key))) { return fail("the entry wasn't removed", 0); }
  // The entry held the handle's last count, so the key has none now.
  if (adamic_weak_key_find(key) != NULL || adamic_weak_map_remove(map, adamic_weak_key_find(key))) { return fail("a removed entry was found again", 0); }
  if (!live_is(0)) { return fail("entries live after the remove", 0); }
  adamic_release(key);
  adamic_release(map);
  puts("replaced");
 } else if (strcmp(mode, "reuse") == 0) {
  // A new object made at a freed key's address is a new key: it gets a new handle, and the map
  // holds nothing for it.
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_object *key = object();
  adamic_weak *handle = adamic_weak_key(key);
  adamic_retain(handle);
  adamic_weak_map_store(map, handle, reference(object()));
  uintptr_t address = (uintptr_t)key;
  adamic_release(key);
  // Every object made is kept until one lands at the freed address, so none goes back to be taken
  // again. The sanitizers' quarantines may hold the address past the cap.
  enum { attempts = 4096 };
  adamic_object **made = malloc(attempts * sizeof *made);
  if (made == NULL) { return 2; }
  size_t made_count = 0;
  adamic_object *reused = NULL;
  while (made_count < attempts && reused == NULL) {
   adamic_object *each = object();
   made[made_count++] = each;
   if ((uintptr_t)each == address) { reused = each; }
  }
  if (reused == NULL) {
   printf("no reuse in %d objects\n", (int)attempts);
  } else {
   adamic_weak *found = adamic_weak_key_find(reused);
   if (found != NULL && adamic_weak_map_lookup(map, found, NULL)) { return fail("a new object at the freed key's address found its entry", (long)made_count); }
   adamic_weak *fresh = adamic_weak_key(reused);
   if (fresh == handle) { return fail("a new object at the freed key's address took its handle", (long)made_count); }
   if (adamic_weak_map_lookup(map, fresh, NULL)) { return fail("a new object's new handle found an entry", (long)made_count); }
   adamic_weak_map_store(map, fresh, reference(object()));
   if (!adamic_weak_map_lookup(map, fresh, NULL) || adamic_weak_map_lookup(map, handle, NULL)) { return fail("the new key's entry and the old one's disagree", (long)made_count); }
   if (!live_is(1)) { return fail("the new key's entry counted wrong", (long)made_count); }
   printf("reused after %zu objects\n", made_count);
  }
  for (size_t i = 0; i < made_count; i++) { adamic_release(made[i]); }
  free(made);
  if (!live_is(0)) { return fail("entries live after every key was freed", 0); }
  adamic_release(handle);
  adamic_release(map);
 } else if (strcmp(mode, "map first") == 0 || strcmp(mode, "keys first") == 0) {
  // Either order frees cleanly. Plain Weak handles to the keys keep each key's handle alive past the
  // map, as a Weak<T> held elsewhere does, so a name left on a handle is still there to be followed.
  enum { keys = 3 };
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_object *key[keys];
  adamic_weak *weak[keys];
  for (size_t i = 0; i < keys; i++) {
   key[i] = object();
   weak[i] = adamic_weak_of(key[i]);
   adamic_weak_map_store(map, adamic_weak_key(key[i]), reference(object()));
  }
  if (!live_is(keys)) { return fail("entries stored, another count live", 0); }
  bool map_first = strcmp(mode, "map first") == 0;
  if (map_first) {
   adamic_release(map);
   if (!live_is(0)) { return fail("entries live after the map was freed", 0); }
  }
  for (size_t i = 0; i < keys; i++) {
   if (adamic_weak_target(weak[i]) != key[i]) { return fail("a live key's handle lost it", (long)i); }
   adamic_release(key[i]);
   if (adamic_weak_target(weak[i]) != NULL) { return fail("a freed key's handle still names it", (long)i); }
  }
  if (!map_first) {
   if (!live_is(0)) { return fail("entries live after every key was freed", 0); }
   adamic_release(map);
  }
  for (size_t i = 0; i < keys; i++) { adamic_release(weak[i]); }
  puts(mode);
 } else if (strcmp(mode, "set") == 0) {
  // A WeakSet's add, has and delete go through the same contract, with no values.
  adamic_weak_map *set = adamic_weak_map_new(true, false);
  adamic_object *key = object(), *other = object();
  adamic_weak_map_store(set, adamic_weak_key(key), (adamic_value){0});
  if (!adamic_weak_map_lookup(set, adamic_weak_key_find(key), NULL)) { return fail("an added key isn't in the set", 0); }
  if (adamic_weak_map_lookup(set, adamic_weak_key_find(other), NULL)) { return fail("a key never added is in the set", 0); }
  if (!adamic_weak_map_remove(set, adamic_weak_key_find(key)) || adamic_weak_map_lookup(set, adamic_weak_key_find(key), NULL)) { return fail("a deleted key is still in the set", 0); }
  adamic_weak_map_store(set, adamic_weak_key(key), (adamic_value){0});
  adamic_weak *handle = adamic_weak_of(key);
  adamic_release(key);
  if (adamic_weak_map_lookup(set, handle, NULL) || !live_is(0)) { return fail("a freed key is still in the set", 0); }
  adamic_release(handle);
  adamic_release(other);
  adamic_release(set);
  puts("set");
 } else if (strcmp(mode, "region value") == 0) {
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_object *key = object();
  adamic_region region = ADAMIC_REGION;
  adamic_object *inside = adamic_object_new_filled_in(&region, &shape);
  adamic_weak_map_store(map, adamic_weak_key(key), reference(inside));
  puts("stored a region's value");
 } else if (strcmp(mode, "region key") == 0) {
  adamic_region region = ADAMIC_REGION;
  adamic_object *inside = adamic_object_new_filled_in(&region, &shape);
  adamic_weak_key(inside);
  puts("keyed a region's value");
 } else if (strcmp(mode, "share map") == 0) {
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_share(map);
  puts("shared a map");
 } else if (strcmp(mode, "share key") == 0 || strcmp(mode, "share removed") == 0) {
  adamic_weak_map *map = adamic_weak_map_new(false, true);
  adamic_object *key = object();
  // A plain Weak's handle, the same one the collection keys by, outlives the entry.
  adamic_weak *plain = adamic_weak_of(key);
  adamic_weak_map_store(map, adamic_weak_key(key), reference(object()));
  if (strcmp(mode, "share removed") == 0) {
   // A key whose entry is gone is no collection's: it still carries the named bit and has a handle,
   // and it shares, as a value a plain Weak points at always has.
   adamic_weak_map_remove(map, adamic_weak_key_find(key));
  }
  adamic_share(key);
  puts("shared a key");
  adamic_release(key);
  if (adamic_weak_target(plain) != NULL) { return fail("a freed shared key's handle still names it", 0); }
  adamic_release(plain);
  adamic_release(map);
 } else {
  return 3;
 }
 return 0;
}
`

var weakEntriesLine = regexp.MustCompile(`(?m)^adamic: weak entries: live (\d+) peak (\d+)$`)

func TestWeakCollection(t *testing.T) {
	t.Parallel()
	builds := []struct {
		name    string
		options native_runtime.Options
	}{
		{"sanitized", native_runtime.Options{Sanitize: true}},
		{"slabs", native_runtime.Options{Sanitize: true, Slabs: true}},
		{"counted", native_runtime.Options{Sanitize: true, Count: true}},
		// No sanitizer's quarantine holds a freed slot here, so "reuse" finds the freed key's address.
		{"plain", native_runtime.Options{Count: true}},
	}
	binaries := map[string]string{}
	for _, build := range builds {
		binary := filepath.Join(t.TempDir(), build.name)
		if err := native.Build(weakCollectionHarness, binary, build.options); err != nil {
			t.Fatalf("%s: %v", build.name, err)
		}
		binaries[build.name] = binary
	}
	run := func(binary, mode, loops string) leakcheck.Run {
		command := exec.Command(binary, mode, loops)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		exit := 0
		if err != nil {
			exit = -1
			if command.ProcessState != nil {
				exit = command.ProcessState.ExitCode()
			}
		}
		return leakcheck.Run{Stdout: []byte(stdout.String()), Stderr: []byte(stderr.String()), ExitCode: exit}
	}
	for _, mode := range []string{"freed", "replace", "map first", "keys first", "set", "share removed"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			for _, build := range builds {
				result := run(binaries[build.name], mode, "10")
				if result.ExitCode != 0 {
					t.Fatalf("%s exit %d\n%s%s", build.name, result.ExitCode, result.Stdout, result.Stderr)
				}
				if !build.options.Count {
					continue
				}
				if report := leakcheck.Unbalanced(result); report != "" {
					t.Fatalf("%s: %s", build.name, report)
				}
				if match := weakEntriesLine.FindSubmatch(result.Stderr); match == nil || string(match[1]) != "0" {
					t.Fatalf("%s: weak entries still live at exit, or no weak entries line\n%s", build.name, result.Stderr)
				}
			}
		})
	}
	// Address reuse runs in the plain build alone: under a sanitizer, ASan's quarantine and the slab
	// quarantine both hold a freed slot past any cap a test can afford (J-7232pdp: no reuse in 4096
	// objects in all three), so the row can't happen there, and a skip would read as a test that ran.
	// The plain build's size classes hand a freed slot out again first, so there a miss is a failure.
	t.Run("reuse", func(t *testing.T) {
		t.Parallel()
		result := run(binaries["plain"], "reuse", "1")
		if result.ExitCode != 0 {
			t.Fatalf("exit %d\n%s%s", result.ExitCode, result.Stdout, result.Stderr)
		}
		if !strings.Contains(string(result.Stdout), "reused after") {
			t.Fatalf("the freed key's address wasn't taken again, where the plain build's size classes took it after 2 objects on J-7232pdp\n%s", result.Stdout)
		}
		if report := leakcheck.Unbalanced(result); report != "" {
			t.Fatal(report)
		}
		t.Log(strings.TrimSpace(string(result.Stdout)))
	})
	// Each refusal is a panic with its own message, exit 70, in every build.
	for _, refusal := range []struct{ mode, message string }{
		{"region value", "a WeakMap can't hold a statement region's value"},
		{"region key", "a statement region's value can't be a WeakMap or WeakSet key"},
		{"share map", "a WeakMap or WeakSet can't cross into parallel work"},
		{"share key", "a WeakMap or WeakSet key can't cross into parallel work"},
	} {
		t.Run(refusal.mode, func(t *testing.T) {
			t.Parallel()
			for _, build := range builds {
				result := run(binaries[build.name], refusal.mode, "1")
				if result.ExitCode != 70 || !strings.Contains(string(result.Stderr), "adamic: panic: "+refusal.message) {
					t.Fatalf("%s: want exit 70 with %q, got exit %d\n%s%s", build.name, refusal.message, result.ExitCode, result.Stdout, result.Stderr)
				}
			}
		})
	}
	// The entries a long-lived map holds stay one at a time, whatever the loop count: a freed key's
	// entry goes with it. Held as a key would be, the peak grows with the loop.
	t.Run("freed peak", func(t *testing.T) {
		t.Parallel()
		peak := func(loops string) int {
			result := run(binaries["counted"], "freed", loops)
			match := weakEntriesLine.FindSubmatch(result.Stderr)
			if result.ExitCode != 0 || match == nil {
				t.Fatalf("loops %s: exit %d\n%s%s", loops, result.ExitCode, result.Stdout, result.Stderr)
			}
			if string(match[1]) != "0" {
				t.Fatalf("loops %s: %s entries live at exit", loops, match[1])
			}
			value, err := strconv.Atoi(string(match[2]))
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		if few, many := peak("10"), peak("1000"); few != 1 || many != 1 {
			t.Fatalf("peak entries %d at 10 loops, %d at 1000, want 1 at both: a freed key's entry outlives it", few, many)
		}
	})
}
