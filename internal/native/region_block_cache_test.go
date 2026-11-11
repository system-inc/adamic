package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"
)

// A thread keeps the blocks its regions free, up to ADAMIC_REGION_BLOCK_CACHE, each a mapping of its
// own, and its next region takes them back before mapping new ones (#yecmp1d, Adamic's ruling). These
// drive region.c alone:
//   - "reuse": the same region opened and ended over and over maps its blocks once, so the blocks
//     made are the same at 10 rounds and at 1000.
//   - "cap": one region that grew past 1 MB of blocks leaves the thread at most the cap.
//   - "threads": four threads each run regions and end (adamic_heap_thread_end), then the main thread
//     ends too, and every block made has been let go.
//   - "idle": after each thread's regions peak and end, every thread idles, and the process holds at
//     most the cache per thread above where it started, under glibc as shipped.
//   - "sanitized": under AddressSanitizer the blocks are malloc's and nothing is kept, so a read of a
//     region's object after the region ends is still a use after free.
//
// Mutants, each red here: block_give lets every block go (reuse grows with the rounds); the cap test
// removed from block_give (cap keeps more than 1 MB); adamic_heap_thread_end not calling
// adamic_region_thread_end (threads leaves blocks unfreed); the AddressSanitizer guard removed
// (sanitized keeps blocks, and the late read lands in a kept block, which ASan doesn't see); standard
// blocks from malloc instead of their own mappings (idle holds about 50 MB a thread).
const regionBlockCacheHarness = `#define _GNU_SOURCE 1
#include "adamic.h"
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

static long rss_kb(void) {
 FILE *file = fopen("/proc/self/status", "r");
 char line[256];
 long value = -1;
 while (file != NULL && fgets(line, sizeof line, file) != NULL) {
  if (strncmp(line, "VmRSS:", 6) == 0) { value = strtol(line + 6, NULL, 10); }
 }
 if (file != NULL) { fclose(file); }
 return value;
}

static pthread_barrier_t idle;
static void *peak_then_idle(void *unused) {
 (void)unused;
 for (int round = 0; round < 20; round++) { fill(400000); }
 pthread_barrier_wait(&idle);
 pthread_barrier_wait(&idle);
 adamic_heap_thread_end();
 return NULL;
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
 } else if (strcmp(mode, "idle") == 0) {
  // Each thread's regions peak at about 20 MB, then every thread idles and the process's resident
  // set is read against what it was before any of them ran.
  long threads = rounds;
  long baseline = rss_kb();
  pthread_barrier_init(&idle, NULL, (unsigned)threads + 1);
  pthread_t ids[16];
  for (long index = 0; index < threads; index++) { pthread_create(&ids[index], NULL, peak_then_idle, NULL); }
  pthread_barrier_wait(&idle);
  long idle_kb = rss_kb();
  pthread_barrier_wait(&idle);
  for (long index = 0; index < threads; index++) { pthread_join(ids[index], NULL); }
  printf("above %ld\n", idle_kb - baseline);
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
	// The ruling's bound (#yecmp1d): after its last region, a thread holds at most the cache above what
	// the process held before, under glibc as shipped. 2 MB of slack covers the threads' stacks and
	// malloc's own bookkeeping: with the cache off, 1 and 4 threads read 320 and 932 KB above
	// (Codex 6af450b5). Blocks kept in malloc's arena read 50 MB a thread here.
	t.Run("idle", func(t *testing.T) {
		if goruntime.GOOS != "linux" {
			t.Skip("reads /proc/self/status")
		}
		for _, threads := range []int{1, 4} {
			stdout, stderr, exit := run(counted, "idle", strconv.Itoa(threads))
			if exit != 0 {
				t.Fatalf("%d threads: exit %d: %s%s", threads, exit, stdout, stderr)
			}
			if above, limit := number(`above (-?\d+)`, stdout), threads*1024+2048; above > limit {
				t.Errorf("%d threads hold %d KB above baseline while idle, want at most %d (the 1 MB cache each, plus 2 MB)", threads, above, limit)
			} else {
				t.Logf("%d threads: %d KB above baseline while idle", threads, above)
			}
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
