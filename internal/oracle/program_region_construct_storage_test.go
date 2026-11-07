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
var programConstructSite = regexp.MustCompile(
	`adamic_object \*\s*(\w+) = adamic_object_construct\([^)]*\);\s*(\w+) = adamic_program_adopt_owned\((\w+), adamic_object_allocation_size\((\w+)\)\);`,
)

// programConstructed is the fixture's Node run, and its C with a check, right after the adoption, that
// the constructed literal really is a Program member, so no run can pass without adopting it. It
// returns the constructed object's name too, for the size mutant.
func programConstructed(test *testing.T) (run, string, string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, programConstructFixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle := onNode(test, path)
	want := "size\nsize,weight,label\nbranch2\n"
	if oracle.exitCode != 0 || string(oracle.stdout) != want || len(oracle.stderr) != 0 {
		test.Fatalf("Node: %d %q %s", oracle.exitCode, oracle.stdout, oracle.stderr)
	}
	program := programRegionLowered(test, path, true)
	if diff := disagreement(oracle, onJavaScriptBackend(test, program)); diff != "" {
		test.Fatalf("JavaScript backend: %s", diff)
	}
	code := native.C(program)
	if !strings.Contains(code, "adamic_object_construct(") || !strings.Contains(code, "adamic_program_adopt_owned(") {
		test.Fatal("the fixture must construct an object and adopt Program members")
	}
	matches := programConstructSite.FindAllStringSubmatchIndex(code, -1)
	if len(matches) != 1 {
		test.Fatalf("expected exactly one adopted construction, got %d", len(matches))
	}
	m := matches[0]
	object := code[m[2]:m[3]]
	if code[m[4]:m[5]] != object || code[m[6]:m[7]] != object || code[m[8]:m[9]] != object {
		test.Fatalf("the adoption after the construction is of another value: %s", code[m[0]:m[1]])
	}
	message := "constructed literal is not a Program member"
	assertion := fmt.Sprintf(
		"\n if (!adamic_program_is(%s)) adamic_panic(\"%s\", sizeof \"%s\" - 1);\n",
		object,
		message,
		message,
	)
	return oracle, code[:m[1]] + assertion + code[m[1]:], object
}

func TestProgramRegionConstructedMemberStorage(test *testing.T) {
	test.Parallel()
	oracle, checked, _ := programConstructed(test)
	binary := filepath.Join(test.TempDir(), "construct")
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		binary,
		native.Options{ProgramRegion: true, Sanitize: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if difference := disagreement(oracle, result); difference != "" {
		test.Fatalf("constructed member: %s %s", difference, result.stderr)
	}
	report, checkError := leakcheck.Check(leakcheck.Program{
		C: checked, Sanitized: binary, Counted: filepath.Join(test.TempDir(), "counted"),
		BuildCounted: func(code, output string) error {
			return programRegionBuild(test, code, output, native.Options{ProgramRegion: true, Count: true})
		},
		Execute: func(environment []string, name string, args ...string) leakcheck.Run {
			return leakRun(executeWith(test, environment, name, args...))
		},
	})
	if checkError != nil || report != "" {
		test.Fatalf("leak check: %v %s", checkError, report)
	}
	// LeakSanitizer is the Linux leak proof; counters independently account for every member.
	countedBinary := filepath.Join(test.TempDir(), "construct-counts")
	if programRegionBuildError := programRegionBuild(
		test,
		checked,
		countedBinary,
		native.Options{ProgramRegion: true, Count: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	command, args := pinnedStack(countedBinary)
	counted := execute(test, command, args...)
	observed := counted
	observed.stderr = leakcheck.CountsLine.ReplaceAll(counted.stderr, nil)
	if difference := disagreement(oracle, observed); difference != "" {
		test.Fatalf("counted constructed member: %s %s", difference, counted.stderr)
	}
	if report := leakcheck.Unbalanced(leakRun(counted)); report != "" {
		test.Fatal(report)
	}
	test.Logf("constructed Program member agrees with Node under ASan/UBSan and the leak check: %s", counted.stderr)
}

// Without the re-aim, the adopted object's write order still names the freed original, and the first
// publication after adoption writes into it.
func TestProgramRegionConstructedOrderMutant(test *testing.T) {
	test.Parallel()
	oracle, checked, _ := programConstructed(test)
	directory := test.TempDir()
	sources := filepath.Join(repository, "internal/native/runtime")
	// Reads: internal/native/runtime/*
	entries, readDirectoryError := os.ReadDir(sources)
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	before := "\t\t((adamic_object *)heap)->write_order = (size_t *)(void *)((unsigned char *)heap + order_offset);\n"
	changed := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, readError := os.ReadFile(filepath.Join(sources, entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		if entry.Name() == "heap.c" {
			if strings.Count(string(data), before) != 1 {
				test.Fatal("the write order re-aim must match exactly once in heap.c")
			}
			data = []byte(strings.Replace(string(data), before, "\t\t(void)order_offset;\n", 1))
			changed = true
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if !changed {
		test.Fatal("runtime mutation changed nothing")
	}
	options := native.Options{ProgramRegion: true, Sanitize: true}
	archive, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(directory, checked, options)
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	binary := filepath.Join(test.TempDir(), "order-mutant")
	if buildWithRuntimeError := native.BuildWithRuntime(
		checked,
		binary,
		options,
		archive,
	); buildWithRuntimeError != nil {
		test.Fatalf("mutant must compile: %v", buildWithRuntimeError)
	}
	result := executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
	)
	if disagreement(oracle, result) == "" || !strings.Contains(string(result.stderr), "heap-use-after-free") {
		test.Fatalf("write order re-aim mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	test.Log("dropped write order re-aim caught by ASan heap-use-after-free")
}

// Sized as a plain object, the adoption would copy short of the write order. Leaf's three fields make
// the two sizes differ; the runtime must refuse the copy rather than make it.
func TestProgramRegionConstructedSizeMutant(test *testing.T) {
	test.Parallel()
	oracle, checked, object := programConstructed(test)
	before := "adamic_object_allocation_size(" + object + ")"
	if strings.Count(checked, before) != 1 {
		test.Fatal("the constructed object's adoption size must match exactly once")
	}
	mutant := strings.Replace(checked, before, "adamic_object_size("+object+"->shape->count)", 1)
	binary := filepath.Join(test.TempDir(), "size-mutant")
	if programRegionBuildError := programRegionBuild(
		test,
		mutant,
		binary,
		native.Options{ProgramRegion: true, Sanitize: true},
	); programRegionBuildError != nil {
		test.Fatal(programRegionBuildError)
	}
	result := executeWith(
		test,
		[]string{"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"},
		binary,
	)
	if result.exitCode != 70 ||
		!strings.Contains(string(result.stderr), "compiler bug: adoption copies short of an object's storage") ||
		disagreement(oracle, result) == "" {
		test.Fatalf("short adoption mutant escaped: %d %s", result.exitCode, result.stderr)
	}
	test.Log("plain-object adoption size refused by the runtime before the copy")
}
