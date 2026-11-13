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

// A thread's spare heap chunks past ADAMIC_HEAP_SPARE_CAP go back to the kernel only at a quiet point
// (#vea9pph, Adamic's ruling), adamic_heap_quiet, never as they empty. These drive heap.c alone, as a
// quiet point would:
//   - "drop": a million small values made and let go of release nothing until the thread goes quiet,
//     and then the chunks past the cap go back, leaving none pending.
//   - "reuse": the released chunks are carved again and every value made in them reads back as it was
//     written, so no slot is taken to be zero or stale.
//   - "idle": on 1 and 4 threads, each having made and dropped a million values and gone quiet, the
//     process holds at most the cap per thread above where it started, plus each released chunk's
//     header pages and a stated slack (Runtime's bound on #vea9pph).
//   - "remote": values one thread made and another let go of come home at their maker's quiet point,
//     so a parked worker's chunks empty and go back too.
//   - "stale" and "sanitized-reuse": with the size classes on under AddressSanitizer (ADAMIC_SLABS), a
//     read of a value in a released chunk is still reported, and a released chunk carved again reads
//     clean (Language's road on #vea9pph).
//
// Mutants, each red here: adamic_heap_quiet keeping every spare (drop and idle); release skipping its
// madvise (idle); a release in give_local as a chunk empties, at the drop (drop's released before
// quiet); adamic_heap_quiet not draining remote frees (remote).
const heapQuietHarness = `#define _GNU_SOURCE 1
#include "adamic.h"
#include "count.h"
#include <pthread.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#define VALUES 1000000

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

// make fills values with boxed numbers, each a small heap value from the size classes.
static void make(adamic_heap **values, size_t count, double offset) {
 for (size_t index = 0; index < count; index++) { values[index] = adamic_box_number(offset + (double)index); }
}
static void drop(adamic_heap **values, size_t count) {
 for (size_t index = 0; index < count; index++) { adamic_release(values[index]); }
}
static int same(adamic_heap **values, size_t count, double offset) {
 for (size_t index = 0; index < count; index++) {
  if (((adamic_number_box *)values[index])->number != offset + (double)index) { return 0; }
 }
 return 1;
}

static pthread_barrier_t idle;
static void *peak_then_quiet(void *unused) {
 (void)unused;
 adamic_heap **values = malloc(VALUES * sizeof *values);
 make(values, VALUES, 0);
 drop(values, VALUES);
 free(values);
 adamic_heap_quiet();
 pthread_barrier_wait(&idle);
 pthread_barrier_wait(&idle);
 return NULL;
}

static adamic_heap **handed;
static pthread_barrier_t handoff;
static void *maker(void *unused) {
 (void)unused;
 make(handed, VALUES, 0);
 pthread_barrier_wait(&handoff);
 pthread_barrier_wait(&handoff);
 printf("pending %d\n", adamic_heap_quiet_pending() ? 1 : 0);
 adamic_heap_quiet();
 printf("released %zu\n", (size_t)adamic_counted.heap_chunks_released);
 return NULL;
}

int main(int count, char **arguments) {
 if (count != 3) { return 2; }
 const char *mode = arguments[1];
 long threads = strtol(arguments[2], NULL, 10);
 printf("page %ld\n", sysconf(_SC_PAGESIZE));
 adamic_heap **values = malloc(VALUES * sizeof *values);
 if (strcmp(mode, "drop") == 0) {
  make(values, VALUES, 0);
  drop(values, VALUES);
  printf("before %zu pending %d\n", (size_t)adamic_counted.heap_chunks_released, adamic_heap_quiet_pending() ? 1 : 0);
  adamic_heap_quiet();
  printf("after %zu pending %d\n", (size_t)adamic_counted.heap_chunks_released, adamic_heap_quiet_pending() ? 1 : 0);
 } else if (strcmp(mode, "reuse") == 0 || strcmp(mode, "sanitized-reuse") == 0) {
  make(values, VALUES, 0);
  drop(values, VALUES);
  adamic_heap_quiet();
  make(values, VALUES, 7);
  printf("same %d released %zu\n", same(values, VALUES, 7), (size_t)adamic_counted.heap_chunks_released);
  drop(values, VALUES);
 } else if (strcmp(mode, "idle") == 0) {
  long baseline = rss_kb();
  pthread_barrier_init(&idle, NULL, (unsigned)threads + 1);
  pthread_t ids[16];
  for (long index = 0; index < threads; index++) { pthread_create(&ids[index], NULL, peak_then_quiet, NULL); }
  pthread_barrier_wait(&idle);
  long idle_kb = rss_kb();
  pthread_barrier_wait(&idle);
  for (long index = 0; index < threads; index++) { pthread_join(ids[index], NULL); }
  printf("above %ld released %zu\n", idle_kb - baseline, (size_t)adamic_counted.heap_chunks_released);
 } else if (strcmp(mode, "remote") == 0) {
  handed = values;
  pthread_barrier_init(&handoff, NULL, 2);
  pthread_t id;
  pthread_create(&id, NULL, maker, NULL);
  pthread_barrier_wait(&handoff);
  drop(values, VALUES);
  pthread_barrier_wait(&handoff);
  pthread_join(id, NULL);
 } else if (strcmp(mode, "stale") == 0) {
  make(values, VALUES, 0);
  adamic_heap *late = values[VALUES / 2];
  drop(values, VALUES);
  adamic_heap_quiet();
  printf("released %zu\n", (size_t)adamic_counted.heap_chunks_released);
  fflush(stdout);
  // A read of a value let go of, its chunk since given back: ASan reports it, since its slot stays
  // poisoned through the release.
  volatile double read = ((adamic_number_box *)late)->number;
  (void)read;
 } else {
  return 5;
 }
 free(values);
 return 0;
}
`

func TestHeapQuietRelease(t *testing.T) {
	t.Parallel()
	counted := filepath.Join(t.TempDir(), "counted")
	if err := Build(heapQuietHarness, counted, Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	slabs := filepath.Join(t.TempDir(), "slabs")
	if err := Build(heapQuietHarness, slabs, Options{Sanitize: true, Count: true, slabs: true}); err != nil {
		t.Fatal(err)
	}
	run := func(binary, mode, threads string) (string, string, int) {
		command := exec.Command(binary, mode, threads)
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
	t.Run("drop", func(t *testing.T) {
		stdout, stderr, exit := run(counted, "drop", "1")
		if exit != 0 {
			t.Fatalf("exit %d: %s%s", exit, stdout, stderr)
		}
		if before := number(`before (\d+)`, stdout); before != 0 {
			t.Fatalf("%d chunks went back at the drop, before any quiet point", before)
		}
		if number(`before \d+ pending (\d)`, stdout) != 1 {
			t.Fatalf("a million values dropped left nothing pending: %s", stdout)
		}
		if after, pending := number(`after (\d+)`, stdout), number(`after \d+ pending (\d)`, stdout); after == 0 || pending != 0 {
			t.Fatalf("the quiet point released %d chunks and left pending %d, want some and none", after, pending)
		}
	})
	t.Run("reuse", func(t *testing.T) {
		for _, binary := range []string{counted, slabs} {
			mode := "reuse"
			if binary == slabs {
				mode = "sanitized-reuse"
			}
			stdout, stderr, exit := run(binary, mode, "1")
			if exit != 0 || strings.Contains(stderr, "Sanitizer") {
				t.Fatalf("%s: exit %d: %s%s", mode, exit, stdout, stderr)
			}
			if number(`same (\d)`, stdout) != 1 || number(`released (\d+)`, stdout) == 0 {
				t.Fatalf("%s: values made in released chunks don't read back, or nothing was released: %s", mode, stdout)
			}
		}
	})
	// The ruled bound (#vea9pph): once each thread has gone quiet, the process holds at most the 1 MB cap
	// per thread above what it held before, plus each released chunk's header page and the page before
	// its aligned interior (two pages a chunk, Runtime's formula), plus 2 MB for the threads' stacks and
	// malloc's own bookkeeping, as the region cache's pin allows.
	t.Run("idle", func(t *testing.T) {
		if goruntime.GOOS != "linux" {
			t.Skip("reads /proc/self/status")
		}
		for _, threads := range []int{1, 4} {
			stdout, stderr, exit := run(counted, "idle", strconv.Itoa(threads))
			if exit != 0 {
				t.Fatalf("%d threads: exit %d: %s%s", threads, exit, stdout, stderr)
			}
			page := number(`page (\d+)`, stdout) / 1024
			released := number(`released (\d+)`, stdout)
			limit := threads*1024 + released*2*page + 2048
			if above := number(`above (-?\d+)`, stdout); above > limit {
				t.Errorf("%d threads hold %d KB above baseline after going quiet, want at most %d (1 MB each, 2 pages for each of %d released chunks, and 2 MB)", threads, above, limit, released)
			} else {
				t.Logf("%d threads: %d KB above baseline after going quiet, %d chunks released, limit %d", threads, above, released, limit)
			}
		}
	})
	t.Run("remote", func(t *testing.T) {
		stdout, stderr, exit := run(counted, "remote", "1")
		if exit != 0 {
			t.Fatalf("exit %d: %s%s", exit, stdout, stderr)
		}
		if number(`pending (\d)`, stdout) != 1 || number(`released (\d+)`, stdout) == 0 {
			t.Fatalf("values made on a thread and let go of on another didn't come home at its quiet point: %s", stdout)
		}
	})
	t.Run("stale", func(t *testing.T) {
		stdout, stderr, _ := run(slabs, "stale", "1")
		if number(`released (\d+)`, stdout) == 0 {
			t.Fatalf("nothing was released, so the read pins nothing: %s", stdout)
		}
		if !strings.Contains(stderr, "use-after-poison") {
			t.Fatalf("a read of a value in a released chunk wasn't reported: %s", stderr)
		}
	})
}
