package oracle

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Each layout owns its Node comparison, ownership check and reference-slot
// mutant. The mutant remains valid C; only runtime behavior may kill it.
func checkMapUnionLayout(test *testing.T, name string) {
	test.Helper()
	fixture := "map_union_value_" + name + ".a"
	checkRuntimeWitness(test, fixture)
	path := "internal/oracle/testdata/" + fixture
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, absolute)
	if want.exitCode != 0 {
		test.Fatalf("Node did not finish: %s", want.stderr)
	}
	program, loweredError := lowered(test, absolute)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := native.C(program)
	old := "adamic_map_new(true, true)"
	if !strings.Contains(source, old) {
		test.Fatal("layout has no reference-valued Map constructor")
	}
	mutant := strings.ReplaceAll(source, old, "adamic_map_new(true, false)")
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildError := native.Build(mutant, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatalf("layout mutant must compile: %v", buildError)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	difference := disagreement(want, got)
	report := ""
	if got.exitCode == 0 {
		report = leakcheck.Report(test, mutant, binary)
	}
	if difference == "" && report == "" {
		// Conservative leak scans may keep the mutant's stale pointers reachable.
		// The counted ownership ledger must still see the references the Map lost.
		countedBinary := filepath.Join(test.TempDir(), "mutant-counted")
		if buildError := native.Build(mutant, countedBinary, native.Options{Count: true}); buildError != nil {
			test.Fatalf("counted layout mutant must compile: %v", buildError)
		}
		command, arguments := pinnedStack(countedBinary)
		counted := execute(test, command, arguments...)
		if counted.exitCode != 0 {
			test.Fatalf("counted mutant failed before ownership check: %s", counted.stderr)
		}
		report = leakcheck.Unbalanced(leakRun(counted))
		if report == "" {
			test.Fatal("reference-slot mutant survived Node and ownership checks")
		}
	}
	if got.exitCode != 0 &&
		!strings.Contains(string(got.stderr), "ERROR: AddressSanitizer") &&
		!strings.Contains(string(got.stderr), "runtime error") {
		test.Fatalf("layout mutant failed outside the checked behavior: exit %d stderr %q", got.exitCode, got.stderr)
	}
	test.Logf("%s reference-slot mutant caught: Node difference %q; leak report %s", name, difference, report)
	_, measure := counted(test, path, false, nil, false, false)
	row := measure()
	if *updateCounts {
		measuredCounts.Store(path, row)
	} else {
		rows, recordedCountsError := recordedCounts()
		if recordedCountsError != nil {
			test.Fatal(recordedCountsError)
		}
		if row != rows[path] {
			test.Fatalf("counts changed: got %s; recorded %s", row, rows[path])
		}
	}
	test.Log(row)
}

func TestMapUnionValueStringFalse(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "string_false")
}
func TestMapUnionValueStringNumber(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "string_number")
}
func TestMapUnionValueSourceFile(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "source_file")
}
func TestMapUnionValueResolvedProject(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "resolved_project")
}
func TestMapUnionValueFilesystemEntries(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "filesystem_entries")
}
func TestMapUnionValueHostFileInfo(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "host_file_info")
}
func TestMapUnionValueVisitResult(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "visit_result")
}
func TestMapUnionValuePaths(test *testing.T) { test.Parallel(); checkMapUnionLayout(test, "paths") }

// The constructor's number must be boxed before entering its declared reference
// slot. The mutant keeps C types valid while storing raw numeric bits instead.
func TestMapUnionConstructorBoxingMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/map_union_value_paths.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	constructors := 0
	for _, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		constructor, ok := declaration.Value.(ir.MapNew)
		if !ok {
			continue
		}
		for _, entry := range constructor.Entries {
			box, ok := entry[1].(ir.Box)
			if !ok {
				continue
			}
			number, ok := box.Value.(ir.NumberConstant)
			if ok && number.Value == 18 {
				constructors++
			}
		}
	}
	if constructors != 1 {
		test.Fatalf("want one boxed constructor entry, got %d", constructors)
	}
	source := native.C(program)
	number := "(" + strconv.FormatFloat(18, 'x', -1, 64) + ")"
	needle := "adamic_box_number(" + number + ")"
	if strings.Count(source, needle) != 1 {
		test.Fatalf("missing constructor boxing call %s", needle)
	}
	mutant := insertCollectionMutant(
		strings.Replace(source, needle, "map_constructor_unboxed_number("+number+")", 1),
		`static adamic_heap *map_constructor_unboxed_number(double number) {
 adamic_value unboxed = {.number = number};
 return unboxed.reference;
}`,
	)
	for _, options := range []native.Options{{}, {Sanitize: true}} {
		binary := filepath.Join(test.TempDir(), "mutant")
		if buildError := native.Build(mutant, binary, options); buildError != nil {
			test.Fatalf("mutant must compile: %v", buildError)
		}
		got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
		if got.exitCode == 0 || disagreement(want, got) == "" {
			test.Fatal("constructor boxing mutant survived")
		}
		if options.Sanitize &&
			!strings.Contains(string(got.stderr), "Sanitizer") &&
			!strings.Contains(string(got.stderr), "runtime error") {
			test.Fatalf("missing sanitizer finding: %s", got.stderr)
		}
		test.Logf("missing constructor boxing caught at runtime, sanitize=%t exit=%d", options.Sanitize, got.exitCode)
	}
}

func TestMapUnionEntriesRecordAlias(test *testing.T) {
	test.Parallel()
	checkRuntimeWitness(test, "entries_record_alias.a")
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/entries_record_alias.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(test.TempDir(), "split")
	if buildError := native.Build(
		native.C(program),
		binary,
		native.Options{Split: true, Sanitize: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	got := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if diff := disagreement(want, got); diff != "" {
		test.Fatalf("split R2 witness: %s\n%s", diff, got.stderr)
	}
}

func TestMapUnionValueContextualArray(test *testing.T) {
	test.Parallel()
	checkMapUnionLayout(test, "contextual_array")
}
