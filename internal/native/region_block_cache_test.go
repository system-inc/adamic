package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// A thread keeps the blocks its regions free, up to ADAMIC_REGION_BLOCK_CACHE, and its next region
// takes them back before asking malloc (#yecmp1d, Adamic's ruling). These drive region.c alone:
//   - "reuse": the same region opened and ended over and over asks malloc for its blocks once, so the
//     blocks malloc gave are the same at 10 rounds and at 1000.
//   - "cap": one region that grew past 1 MB of blocks leaves the thread at most the cap.
//   - "threads": four threads each run regions and end (adamic_heap_thread_end), then the main thread
//     ends too, and every block malloc gave has been freed.
//   - "sanitized": under AddressSanitizer the cache is off, so nothing is kept and a read of a region's
//     object after the region ends is still a use after free.
//
// Mutants, each red here: block_give frees every block (reuse grows with the rounds); the cap test
// removed from block_give (cap keeps more than 1 MB); adamic_heap_thread_end not calling
// adamic_region_thread_end (threads leaves blocks unfreed); the AddressSanitizer guard removed
// (sanitized keeps blocks, and the late read lands in a kept block, which ASan doesn't see).
const regionBlockCacheHarness = `#include "adamic.h"
#include "count.h"
#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

static const char *const names[] = {"a", "b", "c"};
static const bool references[] = {false, false, false};
static const adamic_field_kind kinds[] = {adamic_field_number, adamic_field_number, adamic_field_number};
static const adamic_shape shape = {3, names, references, NULL, kinds};

// fill makes count objects in a region and ends it: enough objects take the blocks up their sizes.
static void fill(size_t count) {
 adamic_region region = ADAMIC_REGION;
 for (size_t index = 0; index < count; index++) {
  adamic_object *object = adamic_object_new_filled_in(&region, &shape);
  object->slots[0].number = (double)index;
 }
 adamic_region_end(&region);
}

static void *worker(void *unused) {
 (void)unused;
 for (int round = 0; round < 50; round++) { fill(2000); }
 adamic_heap_thread_end();
 return NULL;
}

int main(int count, char **arguments) {
 if (count != 3) { return 2; }
 const char *mode = arguments[1];
 long rounds = strtol(arguments[2], NULL, 10);
 if (strcmp(mode, "reuse") == 0) {
  for (long round = 0; round < rounds; round++) { fill(2000); }
  printf("blocks %zu\n", (size_t)adamic_counted.region_blocks);
 } else if (strcmp(mode, "cap") == 0) {
  fill(100000);
  size_t cached = adamic_region_cached();
  printf("cached %zu\n", cached);
  if (cached == 0 || cached > 1048576) { return 3; }
 } else if (strcmp(mode, "threads") == 0) {
  pthread_t threads[4];
  for (int index = 0; index < 4; index++) { pthread_create(&threads[index], NULL, worker, NULL); }
  for (int index = 0; index < 4; index++) { pthread_join(threads[index], NULL); }
  fill(2000);
  adamic_heap_thread_end();
  printf("blocks %zu freed %zu\n", (size_t)adamic_counted.region_blocks, (size_t)adamic_counted.region_blocks_freed);
  if (adamic_counted.region_blocks != adamic_counted.region_blocks_freed) { return 4; }
 } else if (strcmp(mode, "sanitized") == 0) {
  adamic_region region = ADAMIC_REGION;
  adamic_object *object = adamic_object_new_filled_in(&region, &shape);
  object->slots[0].number = 7;
  adamic_region_end(&region);
  printf("cached %zu\n", adamic_region_cached());
  fflush(stdout);
  // A use after the region ended: ASan reports it, since the block went back to malloc.
  volatile double late = object->slots[0].number;
  (void)late;
 } else {
  return 5;
 }
 return 0;
}
`

func TestRegionBlockCache(t *testing.T) {
	t.Parallel()
	counted := filepath.Join(t.TempDir(), "counted")
	if err := Build(regionBlockCacheHarness, counted, Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	sanitized := filepath.Join(t.TempDir(), "sanitized")
	if err := Build(regionBlockCacheHarness, sanitized, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	run := func(binary, mode, rounds string) (string, string, int) {
		command := exec.Command(binary, mode, rounds)
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
		return stdout.String(), stderr.String(), exit
	}
	number := func(pattern, text string) int {
		match := regexp.MustCompile(pattern).FindStringSubmatch(text)
		if match == nil {
			t.Fatalf("no %q in %q", pattern, text)
		}
		value, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	t.Run("reuse", func(t *testing.T) {
		few, _, exit := run(counted, "reuse", "10")
		many, _, exit2 := run(counted, "reuse", "1000")
		if exit != 0 || exit2 != 0 {
			t.Fatalf("exit %d and %d", exit, exit2)
		}
		if one, two := number(`blocks (\d+)`, few), number(`blocks (\d+)`, many); one != two {
			t.Fatalf("malloc gave %d blocks over 10 rounds and %d over 1000: the cache doesn't reuse them", one, two)
		}
	})
	t.Run("cap", func(t *testing.T) {
		stdout, stderr, exit := run(counted, "cap", "1")
		if exit != 0 {
			t.Fatalf("exit %d: %s%s", exit, stdout, stderr)
		}
	})
	t.Run("threads", func(t *testing.T) {
		stdout, stderr, exit := run(counted, "threads", "1")
		if exit != 0 {
			t.Fatalf("exit %d: %s%s", exit, stdout, stderr)
		}
	})
	t.Run("sanitized", func(t *testing.T) {
		stdout, stderr, _ := run(sanitized, "sanitized", "1")
		if number(`cached (\d+)`, stdout) != 0 {
			t.Fatalf("the cache kept blocks under AddressSanitizer: %s", stdout)
		}
		if !strings.Contains(stderr, "heap-use-after-free") {
			t.Fatalf("a read after the region ended wasn't reported: %s", stderr)
		}
	})
}
