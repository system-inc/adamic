package native

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A remote free must make a slot available both in the current giving chunk and
// in a full chunk that has fallen off the giving list. Probe allocator state in
// heap.c so a skipped drain cannot hide behind allocating another chunk.
func TestRemoteSlabDrainAndMutant(t *testing.T) {
	const probe = `
#include <pthread.h>
#include <stdio.h>
static void *remote_probe_free(void *value) { adamic_release(value); return NULL; }
int adamic_remote_probe(void) {
 for (int full = 0; full < 2; full++) {
  adamic_heap *first = adamic_allocate(256, adamic_kind_number);
  chunk *each = find_chunk(first->slab - 1);
  adamic_heap *held[CHUNK / 256];
  size_t count = 0;
  if (full) {
   while (each->listed) { held[count++] = adamic_allocate(256, adamic_kind_number); }
  }
  pthread_t worker;
  if (pthread_create(&worker, NULL, remote_probe_free, first) != 0) { return 1; }
  if (pthread_join(worker, NULL) != 0) { return 2; }
  if (atomic_load_explicit(&each->remote, memory_order_relaxed) == NULL) { return 3; }
  adamic_heap *reused = adamic_allocate(256, adamic_kind_number);
  if (reused != first || atomic_load_explicit(&each->remote, memory_order_relaxed) != NULL) {
   fprintf(stderr, "remote slot was not recycled: full=%d\n", full); return 4;
  }
  adamic_release(reused);
  for (size_t i = 0; i < count; i++) { adamic_release(held[i]); }
 }
 return 0;
}
`
	for _, mutant := range []bool{false, true} {
		t.Run(map[bool]string{false: "control", true: "skip_nonempty"}[mutant], func(t *testing.T) {
			directory := t.TempDir()
			files, err := readRuntime(runtime, "runtime")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data := string(file.contents)
				if file.name == "heap.c" {
					if mutant {
						old := "if (atomic_load_explicit(&each->remote, memory_order_relaxed) == NULL) { return; }"
						if strings.Count(data, old) != 1 {
							t.Fatal("mutant lost its unique anchor")
						}
						data = strings.Replace(data, old, "if (true) { return; }", 1)
					}
					data += probe
				}
				if err := os.WriteFile(filepath.Join(directory, file.name), []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			options := Options{Release: true}
			library, err := RuntimeLibrary(directory, options)
			if err != nil {
				t.Fatal(err)
			}
			main := filepath.Join(directory, "main.c")
			if err := os.WriteFile(main, []byte("int adamic_remote_probe(void); int main(void) { return adamic_remote_probe(); }\n"), 0644); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "probe")
			args := append(LinkFlags(options), "-I", filepath.Dir(library), main, "-o", binary)
			args = append(args, RuntimeLinkFlags(library)...)
			args = append(args, "-lm")
			if out, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
				t.Fatalf("compile: %v %s", err, out)
			}
			out, err := exec.Command(binary).CombinedOutput()
			if mutant {
				failure, ok := err.(*exec.ExitError)
				if !ok || failure.ExitCode() != 4 || !strings.Contains(string(out), "remote slot was not recycled") {
					t.Fatalf("mutant survived or failed incorrectly: %v %s", err, out)
				}
				t.Logf("skip-nonempty mutant caught: %s", out)
			} else if err != nil {
				t.Fatalf("control: %v %s", err, out)
			}
		})
	}
}
