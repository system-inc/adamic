package native

// Probe, never to land (#yecmp1d): resident memory after regions end on several threads, then idle.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const regionRSSProbe = `#include "adamic.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static const char *const names[] = {"a", "b", "c", "d"};
static const bool references[] = {false, false, false, false};
static const adamic_field_kind kinds[] = {adamic_field_number, adamic_field_number, adamic_field_number, adamic_field_number};
static const adamic_shape shape = {4, names, references, NULL, kinds};

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
static void *worker(void *unused) {
 (void)unused;
 for (int round = 0; round < 20; round++) {
  adamic_region region = ADAMIC_REGION;
  for (size_t index = 0; index < 400000; index++) {
   adamic_object *object = adamic_object_new_filled_in(&region, &shape);
   object->slots[0].number = (double)index;
  }
  adamic_region_end(&region);
 }
 printf("thread cached %zu\n", adamic_region_cached());
 pthread_barrier_wait(&idle);
 pthread_barrier_wait(&idle);
 adamic_heap_thread_end();
 return NULL;
}

int main(int count, char **arguments) {
 int threads = count > 1 ? atoi(arguments[1]) : 4;
 long baseline = rss_kb();
 pthread_barrier_init(&idle, NULL, (unsigned)threads + 1);
 pthread_t ids[64];
 for (int index = 0; index < threads; index++) { pthread_create(&ids[index], NULL, worker, NULL); }
 pthread_barrier_wait(&idle);
 sleep(1);
 long idle_kb = rss_kb();
 pthread_barrier_wait(&idle);
 for (int index = 0; index < threads; index++) { pthread_join(ids[index], NULL); }
 long ended_kb = rss_kb();
 printf("threads %d baseline_kb %ld idle_kb %ld ended_kb %ld idle_above_kb %ld\n", threads, baseline, idle_kb, ended_kb, idle_kb - baseline);
 return 0;
}
`

func TestZzRegionRSSProbe(t *testing.T) {
	if os.Getenv("RSS_PROBE") == "" {
		t.Skip("RSS_PROBE unset")
	}
	binary := filepath.Join(t.TempDir(), "rss")
	if err := Build(regionRSSProbe, binary, Options{}); err != nil {
		t.Fatal(err)
	}
	for _, threads := range []string{"1", "4"} {
		for _, environment := range [][]string{nil, {"MALLOC_MMAP_THRESHOLD_=131072", "MALLOC_TRIM_THRESHOLD_=131072"}} {
			command := exec.Command(binary, threads)
			command.Env = append(os.Environ(), environment...)
			output, err := command.CombinedOutput()
			fmt.Printf("== threads %s env %v err %v\n%s", threads, environment, err, output)
		}
	}
}
