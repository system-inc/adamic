package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Graph regions meet concurrency's counts in adamic.h: the inline retain and release must never count
// a graph member (its region counts too) or an environment's interior cell (its environment counts it),
// and a graph value must never become shared, since region counts are plain.

// buildEdited builds source against the runtime with edit applied to each runtime file, as the
// release fast-path mutant does, so a mutant changes the runtime the program links.
func buildEdited(t *testing.T, source string, options Options, edit func(name string, data string) string) string {
	t.Helper()
	directory := t.TempDir()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data := string(file.contents)
		if edit != nil {
			data = edit(file.name, data)
		}
		if err := os.WriteFile(filepath.Join(directory, file.name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	library, err := RuntimeLibrary(directory, options)
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "program")
	arguments := append(Flags(options), "-I", filepath.Dir(library), "-o", binary, main)
	arguments = append(arguments, RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("clang: %v %s", err, output)
	}
	return binary
}

var graphCounts = regexp.MustCompile(`adamic: counts: allocations (\d+) frees (\d+) `)

// A member holding two counts when it joins a region adds both to the region. If the inline path
// handled its first release, the region would keep that count forever and never be freed.
func TestGraphMembersCountOnTheirRegion(t *testing.T) {
	source := `#include "adamic.h"
#include "graph_regions.h"
int main(void) {
	adamic_cell *a = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 1}, false), sizeof(adamic_cell));
	adamic_retain(a);
	adamic_cell *b = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 2}, false), sizeof(adamic_cell));
	adamic_graph_merge(a, b);
	adamic_release(a);
	adamic_release(a);
	adamic_release(b);
	return 0;
}
`
	inline := func(name, data string) string {
		if name != "adamic.h" {
			return data
		}
		old := "(heap->slab & ADAMIC_SLOW_COUNT) == 0"
		if strings.Count(data, old) != 2 {
			t.Fatal("mutant lost its anchors in adamic.h")
		}
		return strings.ReplaceAll(data, old, "(heap->slab & ADAMIC_SHARED_HEADER) == 0")
	}
	for _, mutant := range []bool{false, true} {
		var edit func(string, string) string
		if mutant {
			edit = inline
		}
		binary := buildEdited(t, source, Options{Count: true}, edit)
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("mutant %v: %v %s", mutant, err, output)
		}
		match := graphCounts.FindSubmatch(output)
		if match == nil {
			t.Fatalf("mutant %v: no counts line: %s", mutant, output)
		}
		balanced := bytes.Equal(match[1], match[2])
		if !mutant && !balanced {
			t.Fatalf("region not freed: %s", output)
		}
		if mutant && balanced {
			t.Fatalf("inline-counting mutant freed everything, so the test can't see it: %s", output)
		}
	}
	t.Log("the region is freed; counting a graph member inline leaves it allocated")
}

// An environment's interior cell has count zero and is counted by its environment, on the slow path.
// Without that redirect the environment is freed while a reference through its cell is still held.
func TestEnvironmentCellsCountTheirEnvironment(t *testing.T) {
	source := `#include "adamic.h"
int main(void) {
	adamic_environment *environment = adamic_environment_new(1);
	adamic_cell *cell = &environment->cells[0];
	adamic_retain(cell);
	adamic_release(environment);
	adamic_release(cell);
	return 0;
}
`
	redirect := func(name, data string) string {
		if name != "heap.c" {
			return data
		}
		old := "if (heap != NULL && heap->kind == adamic_kind_cell && ((adamic_cell *)heap)->owner != NULL) {\n\t\theap = ((adamic_cell *)heap)->owner;\n\t}\n\treturn heap;"
		if strings.Count(data, old) != 1 {
			t.Fatal("mutant lost its anchor in heap.c")
		}
		return strings.Replace(data, old, "return heap;", 1)
	}
	for _, mutant := range []bool{false, true} {
		var edit func(string, string) string
		if mutant {
			edit = redirect
		}
		binary := buildEdited(t, source, Options{Sanitize: true}, edit)
		command := exec.Command(binary)
		command.Env = parallelEnvironment("1", true)
		output, err := command.CombinedOutput()
		if !mutant && err != nil {
			t.Fatalf("environment cell: %v %s", err, output)
		}
		if mutant && (err == nil || !bytes.Contains(output, []byte("AddressSanitizer: heap-use-after-free"))) {
			t.Fatalf("mutant without the cell redirect survived: %v %s", err, output)
		}
	}
	t.Log("a retained interior cell keeps its environment; without the redirect ASan sees the free")
}

// Region counts are plain and single-threaded, so a graph value reaching adamic_share stops by name.
// The compiler refuses such a program first (internal/oracle/testdata/concurrency/refused/graph_region.a).
func TestGraphValueIsNeverShared(t *testing.T) {
	source := `#include "adamic.h"
#include "graph_regions.h"
int main(void) {
	adamic_cell *cell = adamic_graph_adopt(adamic_cell_new((adamic_value){.number = 1}, false), sizeof(adamic_cell));
	adamic_share(cell);
	adamic_release(cell);
	return 0;
}
`
	binary := buildEdited(t, source, Options{}, nil)
	output, err := exec.Command(binary).CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("adamic: panic: a graph region can't cross into parallel work yet")) {
		t.Fatalf("want the graph sharing panic, got %v %s", err, output)
	}
}
