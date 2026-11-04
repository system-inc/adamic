package native

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Observe only calls from the tested inline wrappers, then execute the real slow
// helper. Program operations are no-ops, so output alone cannot pin dispatch.
const retainReleaseSource = `#define _POSIX_C_SOURCE 200809L
#define adamic_retain_slow observed_retain_slow
#define adamic_release_slow observed_release_slow
#include "adamic.h"
#undef adamic_retain_slow
#undef adamic_release_slow
#include "graph_regions.h"
#include "async.h"
#include <assert.h>
#include <pthread.h>
#include <string.h>
void *adamic_retain_slow(void *);
void adamic_release_slow(void *);
static _Atomic size_t slow_retains, slow_releases;
void *observed_retain_slow(void *value) {
 atomic_fetch_add(&slow_retains, 1); return adamic_retain_slow(value);
}
void observed_release_slow(void *value) {
 atomic_fetch_add(&slow_releases, 1); adamic_release_slow(value);
}
static void redirected(void *value) {
 size_t retained = slow_retains, released = slow_releases;
 assert(adamic_retain(value) == value);
 assert(slow_retains == retained + 1 && "retain must redirect ownership");
 adamic_release(value);
 assert(slow_releases == released + 1 && "release must redirect ownership");
}
typedef struct { adamic_async_frame base; adamic_cell cell; } frame;
static void frame_children(adamic_async_frame *base, void (*drop)(void *)) {
 frame *self = (void *)base;
 adamic_environment_drop_cells(&self->cell, 1, drop);
}
static pthread_barrier_t barrier;
static void *worker(void *value) {
 pthread_barrier_wait(&barrier);
 for(size_t i=0;i<10000;i++) { adamic_retain(value); adamic_release(value); }
 adamic_heap_thread_end(); return NULL;
}
int main(int argc, char **argv) {
 adamic_start(argc,argv);
 assert(argc == 2);
 if(strcmp(argv[1],"ordinary")==0) {
  static adamic_string immortal = ADAMIC_STRING("literal");
  assert(adamic_retain(NULL)==NULL); adamic_release(NULL);
  assert(adamic_retain(&immortal)==&immortal); adamic_release(&immortal);
  assert(slow_retains==0 && slow_releases==0);
  adamic_cell *cell=adamic_cell_new((adamic_value){.number=7},false);
  adamic_retain(cell); assert(cell->heap.references==2);
  adamic_release(cell); assert(cell->heap.references==1);
  assert(slow_retains==0 && slow_releases==0);
  adamic_release(cell); assert(slow_releases==1);
  assert(adamic_counted.retains==3 && adamic_counted.releases==4);
 } else if(strcmp(argv[1],"graph")==0) {
  adamic_cell *a=adamic_graph_adopt(adamic_cell_new((adamic_value){.number=7},false),sizeof(adamic_cell));
  static const adamic_shape shape={0,NULL,NULL,NULL,NULL};
  adamic_object *b=adamic_graph_adopt(adamic_object_new(&shape),adamic_object_size(0));
  adamic_graph_merge(a,b); adamic_release(b);
  assert(b->heap.references==0);
  redirected(a); redirected(b);
  adamic_retain(b); adamic_release(a);
  assert(a->value.number==7 && b->shape==&shape);
  adamic_release(b);
 } else if(strcmp(argv[1],"program")==0) {
  adamic_environment *owner=adamic_program_adopt_owned(adamic_environment_new(1),sizeof(adamic_environment)+sizeof(adamic_cell));
  assert(adamic_program_is(owner) && owner->heap.references==0);
  redirected(owner); redirected(&owner->cells[0]);
  assert(owner->heap.references==0 && owner->cells[0].heap.references==0);
  adamic_program_region_end();
 } else if(strcmp(argv[1],"environment")==0) {
  adamic_environment *owner=adamic_environment_new(1);
  adamic_cell *cell=&owner->cells[0];
  cell->value.number=41; cell->ready=true;
  redirected(cell); assert(owner->heap.references==1 && cell->heap.references==0);
  adamic_retain(cell); adamic_release(owner);
  assert(cell->value.number==41); adamic_release(cell);
 } else if(strcmp(argv[1],"async")==0) {
  frame *owner=adamic_allocate(sizeof *owner,adamic_kind_async_frame);
  memset((char *)owner+sizeof(adamic_heap),0,sizeof *owner-sizeof(adamic_heap));
  owner->base.children=frame_children;
  adamic_environment_initialize_cells(&owner->base.heap,&owner->cell,1);
  adamic_cell *cell=&owner->cell; cell->value.number=42; cell->ready=true;
  redirected(cell); assert(owner->base.heap.references==1 && cell->heap.references==0);
  adamic_retain(cell); adamic_release(owner);
  assert(cell->value.number==42); adamic_release(cell);
 } else if(strcmp(argv[1],"shared")==0 || strcmp(argv[1],"shared-race")==0) {
  adamic_cell *cell=adamic_cell_new((adamic_value){.number=43},false);
  adamic_share(cell); if(strcmp(argv[1],"shared")==0) redirected(cell);
  assert(adamic_reference_count(&cell->heap)==1);
  assert(pthread_barrier_init(&barrier,NULL,3)==0);
  pthread_t threads[3];
  for(size_t i=0;i<3;i++) assert(pthread_create(&threads[i],NULL,worker,cell)==0);
  for(size_t i=0;i<3;i++) assert(pthread_join(threads[i],NULL)==0);
  pthread_barrier_destroy(&barrier);
  assert(adamic_reference_count(&cell->heap)==1 && cell->value.number==43);
  adamic_release(cell);
 } else { return 2; }
 assert(adamic_counted.allocations==adamic_counted.frees+adamic_counted.regions);
 puts("ownership clean"); return 0;
}
`

// One runtime build per sanitizer product. Header-only mutants compile just the
// calling translation unit against that archive, isolating the inline decision.
func retainReleaseProduct(t *testing.T, tsan bool) string {
	t.Helper()
	options := Options{ProgramRegion: true, Count: true, Sanitize: !tsan, ThreadSanitize: tsan}
	files, flags, compiler, version, err := runtimeInputs("", options, []string{"-DADAMIC_CANONICAL_CLOSURES=1"})
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: fmt.Sprintf("retain-release-inline-%t", tsan), Files: []string{"internal/native/runtime", "internal/native/retain_release_test.go", "internal/native/native.go", "internal/native/library.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(archiverName(compiler), "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		cache, err := cacheDirectory("retain-release-inline")
		if err != nil {
			return err
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			return err
		}
		scratch, err := os.MkdirTemp("", "adamic-retain-release-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		for _, file := range files {
			if strings.HasSuffix(file.name, ".h") {
				if err := os.WriteFile(filepath.Join(scratch, file.name), file.contents, 0644); err != nil {
					return err
				}
			}
		}
		header, err := os.ReadFile(filepath.Join(scratch, "adamic.h"))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(scratch, "main.c"), []byte(retainReleaseSource), 0644); err != nil {
			return err
		}
		names := []string{"control"}
		if tsan {
			names = append(names, "shared-retain", "shared-release")
		} else {
			for _, kind := range []string{"graph", "program", "environment", "async"} {
				for _, operation := range []string{"retain", "release"} {
					names = append(names, kind+"-"+operation)
				}
			}
		}
		for _, name := range names {
			changed := string(header)
			if name != "control" {
				kind, operation, _ := strings.Cut(name, "-")
				begin := strings.Index(changed, "ADAMIC_OWNERSHIP void *adamic_retain(")
				middle := strings.Index(changed, "ADAMIC_OWNERSHIP void adamic_release(")
				end := strings.Index(changed, "// Graph allocation and slot ownership")
				if begin < 0 || middle < begin || end < middle {
					return fmt.Errorf("missing wrapper mutation seam")
				}
				if operation == "retain" {
					end = middle
				} else {
					begin = middle
				}
				part := changed[begin:end]
				before, after := "", ""
				switch kind {
				case "graph":
					before = "ADAMIC_SLOW_COUNT"
					after = "(ADAMIC_SLOW_COUNT & ~ADAMIC_GRAPH_FLAG)"
				case "program":
					before = "ADAMIC_SLOW_COUNT"
					after = "(ADAMIC_SLOW_COUNT & ~ADAMIC_PROGRAM_FLAG)"
				case "shared":
					before = "ADAMIC_SLOW_COUNT"
					after = "(ADAMIC_SLOW_COUNT & ~ADAMIC_SHARED_HEADER)"
				case "environment", "async":
					before = "heap->kind != adamic_kind_cell"
					after = "true"
				}
				if before == "" || strings.Count(part, before) != 1 {
					return fmt.Errorf("ambiguous mutant %s", name)
				}
				changed = changed[:begin] + strings.Replace(part, before, after, 1) + changed[end:]
			}
			if err := os.WriteFile(filepath.Join(scratch, "adamic.h"), []byte(changed), 0644); err != nil {
				return err
			}
			args := append(slices.Clone(flags), prefixMaps(scratch, "/adamic-build")...)
			args = append(args, "-I", ".", "-o", filepath.Join(directory, name), "main.c")
			args = append(args, RuntimeLinkFlags(library)...)
			args = append(args, "-lm")
			if output, err := toolchain.CommandIn(scratch, compiler, args...).CombinedOutput(); err != nil {
				return fmt.Errorf("compile %s: %w\n%s", name, err, output)
			}
		}
		return nil
	})
}
func TestProduct_RetainReleaseASan(t *testing.T) { t.Parallel(); retainReleaseProduct(t, false) }
func TestProduct_RetainReleaseTSan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("ThreadSanitizer unavailable")
	}
	retainReleaseProduct(t, true)
}

func retainReleaseCheck(t *testing.T, kind string, tsan, mutant bool) {
	t.Helper()
	if tsan && !parallelTSan(t) {
		t.Skip("ThreadSanitizer unavailable")
	}
	directory := retainReleaseProduct(t, tsan)
	names := []string{"control"}
	if mutant {
		names = []string{kind + "-retain", kind + "-release"}
	}
	for _, name := range names {
		argument := kind
		if tsan && mutant {
			argument = "shared-race"
		}
		command, release := boundedrun.Command(10*time.Second, filepath.Join(directory, name), argument)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1", "TSAN_OPTIONS=halt_on_error=1")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		output := stderr.String()
		release()
		if mutant {
			if err == nil || (!strings.Contains(string(output), "must redirect ownership") && !strings.Contains(string(output), "ThreadSanitizer: data race")) {
				t.Fatalf("%s survived or failed differently: %v\n%s", name, err, output)
			}
			t.Logf("%s caught: %s", name, strings.SplitN(string(output), "\n", 2)[0])
		} else if err != nil || stdout.String() != "ownership clean\n" || strings.Contains(string(output), "ThreadSanitizer:") || strings.Contains(string(output), "ERROR:") {
			t.Fatalf("%s: %v\nstdout: %s\nstderr: %s", kind, err, stdout.String(), output)
		}
	}
}
func TestRetainReleaseOrdinaryASan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "ordinary", false, false)
}
func TestRetainReleaseGraphASan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "graph", false, false)
}
func TestRetainReleaseProgramASan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "program", false, false)
}
func TestRetainReleaseEnvironmentASan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "environment", false, false)
}
func TestRetainReleaseAsyncCellASan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "async", false, false)
}
func TestRetainReleaseSharedTSan(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "shared", true, false)
}
func TestRetainReleaseGraphMutants(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "graph", false, true)
}
func TestRetainReleaseProgramMutants(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "program", false, true)
}
func TestRetainReleaseEnvironmentMutants(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "environment", false, true)
}
func TestRetainReleaseAsyncCellMutants(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "async", false, true)
}
func TestRetainReleaseSharedMutants(t *testing.T) {
	t.Parallel()
	retainReleaseCheck(t, "shared", true, true)
}
