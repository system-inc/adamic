package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// A construction cast's object is longer than adamic_object_size says, and its write order points into
// its own allocation (construction.c). Adopted as a Program member it must move whole, with its write
// order re-aimed at the new allocation, or every later presence or key-order read reaches the freed
// original. The fixture reads both, and Node holds the answer.
const programConstructFixture = "internal/oracle/testdata/program_region/program_region_construct_storage.a"

// programConstructSite is the constructed literal and its Program adoption, which the emitter writes
// back to back (emit_objects.go). The adoption must size the object by its whole allocation.
var programConstructSite = regexp.MustCompile(`adamic_object \*\s*(\w+) = adamic_object_construct\([^)]*\);\s*(\w+) = adamic_program_adopt_owned\((\w+), adamic_object_allocation_size\((\w+)\)\);`)

// programConstructed is the fixture's Node run, and its C with a check, right after the adoption, that
// the constructed literal really is a Program member, so no run can pass without adopting it. It
// returns the constructed object's name too, for the size mutant.
func programConstructed(t *testing.T) (run, string, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, programConstructFixture))
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	want := "size\nsize,weight,label\nbranch2\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		t.Fatalf("Node: %d %q %s", oracle.exitCode, oracle.stdout, oracle.stderr)
	}
	program := programRegionLowered(t, path, true)
	if diff := disagreement(oracle, onJavaScriptBackend(t, program)); diff != "" {
		t.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	if !strings.Contains(code, "adamic_object_construct(") || !strings.Contains(code, "adamic_program_adopt_owned(") {
		t.Fatal("the fixture must construct an object and adopt Program members")
	}
	matches := programConstructSite.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 {
		t.Fatalf("expected exactly one adopted construction, got %d", len(matches))
	}
	m := matches[0]
	object := code[m[2]:m[3]]
	if code[m[4]:m[5]] != object || code[m[6]:m[7]] != object || code[m[8]:m[9]] != object {
		t.Fatalf("the adoption after the construction is of another value: %s", code[m[0]:m[1]])
	}
	message := "constructed literal is not a Program member"
	assertion := fmt.Sprintf("\n if (!adamic_program_is(%s)) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n", object, message, message)
	return oracle, code[:m[1]] + assertion + code[m[1]:], object
}

func TestProgramRegionConstructedMemberStorage(t *testing.T) {
	t.Parallel()
	oracle, checked, _ := programConstructed(t)
	binary := filepath.Join(t.TempDir(), "construct")
	if err := programRegionBuild(t, checked, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		t.Fatalf("constructed member: %s %s", difference, result.stderr)
	}
	report, err := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(t.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(t, code, output, native.Options{ProgramRegion: true, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(t, environment, name, args...))
		},
	})
	if err != nil || report != "" {
		t.Fatalf("leak check: %v %s", err, report)
	}
	// LeakSanitizer is the Linux leak proof; counters independently account for every member.
	countedBinary := filepath.Join(t.TempDir(), "construct-counts")
	if err := programRegionBuild(t, checked, countedBinary, native.Options{ProgramRegion: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	command, args := pinnedStack(countedBinary)
	counted := execute(t, command, args...)
	observed := counted
	observed.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(oracle, observed); difference != "" {
		t.Fatalf("counted constructed member: %s %s", difference, counted.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		t.Fatal(report)
	}
	t.Logf("constructed Program member agrees with Node under ASan/UBSan and the leak check: %s", counted.stderr)
}

// Without the re-aim, the adopted object's write order still names the freed original, and the first
// publication after adoption writes into it.
func TestProgramRegionConstructedOrderMutant(t *testing.T) {
	t.Parallel()
	oracle, checked, _ := programConstructed(t)
	directory := t.TempDir()
	sources := filepath.Join(repository, "internal/native/runtime")
	// Reads: internal/native/runtime/*
	entries, err := os.ReadDir(sources)
	if err != nil {
		t.Fatal(err)
	}
	before := "\t\t((adamic_object *)heap)->write_order = (size_t *)(void *)((unsigned char *)heap + order_offset);\n"
	changed := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sources, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == "heap.c" {
			if strings.Count(string(data), before) != 1 {
				t.Fatal("the write order re-aim must match exactly once in heap.c")
			}
			data = []byte(strings.Replace(string(data), before, "\t\t(void)order_offset;\n", 1))
			changed = true
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("runtime mutation changed nothing")
	}
	options := native.Options{ProgramRegion: true, Sanitize: true}
	archive, err := native.RuntimeLibraryForSource(directory, checked, options)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "order-mutant")
	if err := native.BuildWithRuntime(checked, binary, options, archive); err != nil {
		t.Fatalf("mutant must compile: %v", err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if disagreement(oracle, result) == "" || !strings.Contains(string(result.stderr), "heap-use-after-free") {
		t.Fatalf("write order re-aim mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	t.Log("dropped write order re-aim caught by ASan heap-use-after-free")
}

// Sized as a plain object, the adoption would copy short of the write order. Leaf's three fields make
// the two sizes differ; the runtime must refuse the copy rather than make it.
func TestProgramRegionConstructedSizeMutant(t *testing.T) {
	t.Parallel()
	oracle, checked, object := programConstructed(t)
	before := "adamic_object_allocation_size(" + object + ")"
	if strings.Count(checked, before) != 1 {
		t.Fatal("the constructed object's adoption size must match exactly once")
	}
	mutant := strings.Replace(checked, before, "adamic_object_size("+object+"->shape->count)", 1)
	binary := filepath.Join(t.TempDir(), "size-mutant")
	if err := programRegionBuild(t, mutant, binary, native.Options{ProgramRegion: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode != 70 || !strings.Contains(string(result.stderr), "compiler bug: adoption copies short of an object's storage") || disagreement(oracle, result) == "" {
		t.Fatalf("short adoption mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	t.Log("plain-object adoption size refused by the runtime before the copy")
}
