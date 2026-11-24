package native

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// runtimeSentinelsHarness reads the runtime's four sentinel methods tables (adamic.h) through volatile pointers, so
// each comparison is made at run time on the addresses the link gave them, never folded by the compiler, and then
// asks each predicate that reads a sentinel about an object of each sentinel's shape: adamic_host_shape,
// adamic_regexp_groups_is and adamic_error_is.
const runtimeSentinelsHarness = `#include "adamic.h"
#include <stdio.h>
static const adamic_shape host_shape = {0, NULL, NULL, &adamic_host_methods, NULL};
static const adamic_shape groups_shape = {0, NULL, NULL, &adamic_groups_methods, NULL};
static const adamic_shape error_shape = {0, NULL, NULL, &adamic_error_methods, NULL};
static const adamic_shape coded_error_shape = {0, NULL, NULL, &adamic_coded_error_methods, NULL};
int main(void) {
 adamic_stack_thread_start();
 const adamic_methods *volatile tables[] = {&adamic_host_methods, &adamic_groups_methods, &adamic_error_methods, &adamic_coded_error_methods};
 int shared = 0;
 for (int first = 0; first < 4; first++) {
  for (int second = first + 1; second < 4; second++) {
   if (tables[first] == tables[second]) shared++;
  }
 }
 printf("shared %d\n", shared);
 const adamic_shape *const shapes[] = {&host_shape, &groups_shape, &error_shape, &coded_error_shape};
 for (int index = 0; index < 4; index++) {
  adamic_object *object = adamic_object_new(shapes[index]);
  printf("%d%d%d\n", adamic_host_shape(object->shape), adamic_regexp_groups_is(object), adamic_error_is(object));
  adamic_release(object);
 }
 return 0;
}
`

// The runtime tells a host object, a RegExp match's groups object and an Error apart by which empty methods table
// their shape names, adamic_host_methods, adamic_groups_methods, adamic_error_methods or adamic_coded_error_methods,
// and the four have the same content, which is what a linker's or ThinLTO's identical folding merges. Merged, a
// groups object would read as an Error, or every host object as a groups object. So a ThinLTO release binary, the
// one a program ships as, is built and run, and each pair of addresses must differ there and each predicate answer
// for its own sentinel alone (Runtime's condition 1 on #eksbz4c, head G): by running it, never by reasoning.
func TestRuntimeSentinelsStayDistinctInRelease(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "sentinels")
	if err := Build(runtimeSentinelsHarness, binary, native_runtime.Options{Release: true}); err != nil {
		t.Fatal(err)
	}
	// After the count of shared addresses, one line a shape, host, groups, Error and coded Error in that order, each
	// digit one predicate's answer: adamic_host_shape, adamic_regexp_groups_is, adamic_error_is.
	want := "shared 0\n100\n110\n001\n001\n"
	if got := runWithInput(t, "", binary); got != want {
		t.Fatalf("the runtime's sentinel tables in a ThinLTO release binary:\ngot  %q\nwant %q", got, want)
	}
}

// groupsSentinelHomes are the only files that may take adamic_groups_methods' address: adamic.h's and exceptions.c's
// tests of it, and the two writers of a RegExp match's groups shape, a literal's and a run-time pattern's. Any other
// shape naming it would throw Node's groups TypeError for an object that isn't one.
var groupsSentinelHomes = []string{
	"internal/native_runtime/c/adamic.h",
	"internal/native_runtime/c/exceptions.c",
	"internal/native_runtime/c/regexp_compile_runtime.c",
	"internal/regexp/native.go",
}

// adamic.h says a groups object's sentinel is written into a groups shape and no other; this holds every C file under
// internal, the runtime's, and every Go file that isn't a test, to that.
func TestGroupsSentinelNamesOnlyGroupsShapes(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..")
	var found []string
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || !(strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".h")) {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(source), "&adamic_groups_methods") {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			found = append(found, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(found)
	if !slices.Equal(found, groupsSentinelHomes) {
		t.Fatalf("&adamic_groups_methods is taken in %q, want exactly %q", found, groupsSentinelHomes)
	}
}
