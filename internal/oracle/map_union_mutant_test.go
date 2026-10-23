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
func checkMapUnionLayout(t *testing.T, name string) {
	t.Helper()
	fixture := "map_union_value_" + name + ".a"
	checkRuntimeWitness(t, fixture)
	path := "internal/oracle/testdata/" + fixture
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, absolute)
	if want.exitCode != 0 {
		t.Fatalf("Node did not finish: %s", want.stderr)
	}
	program, err := lowered(t, absolute)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	old := "adamic_map_new(true, true)"
	if !strings.Contains(source, old) {
		t.Fatal("layout has no reference-valued Map constructor")
	}
	mutant := strings.ReplaceAll(source, old, "adamic_map_new(true, false)")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatalf("layout mutant must compile: %v", err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	difference := disagreement(want, got)
	report := ""
	if got.exitCode == 0 {
		report = leakcheck.Report(t, mutant, binary)
	}
	if difference == "" && report == "" {
		// Conservative leak scans may keep the mutant's stale pointers reachable.
		// The counted ownership ledger must still see the references the Map lost.
		countedBinary := filepath.Join(t.TempDir(), "mutant-counted")
		if err := native.Build(mutant, countedBinary, native.Options{Count: true}); err != nil {
			t.Fatalf("counted layout mutant must compile: %v", err)
		}
		command, arguments := pinnedStack(countedBinary)
		counted := execute(t, command, arguments...)
		if counted.exitCode != 0 {
			t.Fatalf("counted mutant failed before ownership check: %s", counted.stderr)
		}
		report = leakcheck.Unbalanced(leakRun(counted))
		if report == "" {
			t.Fatal("reference-slot mutant survived Node and ownership checks")
		}
	}
	if got.exitCode != 0 && !strings.Contains(string(got.stderr), "ERROR: AddressSanitizer") && !strings.Contains(string(got.stderr), "runtime error") {
		t.Fatalf("layout mutant failed outside the checked behavior: exit %d stderr %q", got.exitCode, got.stderr)
	}
	t.Logf("%s reference-slot mutant caught: Node difference %q; leak report %s", name, difference, report)
	_, measure := counted(t, path, false, nil, false, false)
	row := measure()
	if *updateCounts {
		measuredCounts.Store(path, row)
	} else {
		rows, err := recordedCounts()
		if err != nil {
			t.Fatal(err)
		}
		if row != rows[path] {
			t.Fatalf("counts changed: got %s; recorded %s", row, rows[path])
		}
	}
	t.Log(row)
}

func TestMapUnionValueStringFalse(t *testing.T) { t.Parallel(); checkMapUnionLayout(t, "string_false") }
func TestMapUnionValueStringNumber(t *testing.T) {
	t.Parallel()
	checkMapUnionLayout(t, "string_number")
}
func TestMapUnionValueSourceFile(t *testing.T) { t.Parallel(); checkMapUnionLayout(t, "source_file") }
func TestMapUnionValueResolvedProject(t *testing.T) {
	t.Parallel()
	checkMapUnionLayout(t, "resolved_project")
}
func TestMapUnionValueFilesystemEntries(t *testing.T) {
	t.Parallel()
	checkMapUnionLayout(t, "filesystem_entries")
}
func TestMapUnionValueHostFileInfo(t *testing.T) {
	t.Parallel()
	checkMapUnionLayout(t, "host_file_info")
}
func TestMapUnionValueVisitResult(t *testing.T) { t.Parallel(); checkMapUnionLayout(t, "visit_result") }
func TestMapUnionValuePaths(t *testing.T)       { t.Parallel(); checkMapUnionLayout(t, "paths") }

// The constructor's number must be boxed before entering its declared reference
// slot. The mutant keeps C types valid while storing raw numeric bits instead.
func TestMapUnionConstructorBoxingMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/map_union_value_paths.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		t.Fatalf("want one boxed constructor entry, got %d", constructors)
	}
	source := native.C(program)
	number := "(" + strconv.FormatFloat(18, 'x', -1, 64) + ")"
	needle := "adamic_box_number(" + number + ")"
	if strings.Count(source, needle) != 1 {
		t.Fatalf("missing constructor boxing call %s", needle)
	}
	mutant := insertCollectionMutant(strings.Replace(source, needle, "map_constructor_unboxed_number("+number+")", 1), `static adamic_heap *map_constructor_unboxed_number(double number) {
 adamic_value unboxed = {.number = number};
 return unboxed.reference;
}`)
	for _, options := range []native.Options{{}, {Sanitize: true}} {
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(mutant, binary, options); err != nil {
			t.Fatalf("mutant must compile: %v", err)
		}
		got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
		if got.exitCode == 0 || disagreement(want, got) == "" {
			t.Fatal("constructor boxing mutant survived")
		}
		if options.Sanitize && !strings.Contains(string(got.stderr), "Sanitizer") && !strings.Contains(string(got.stderr), "runtime error") {
			t.Fatalf("missing sanitizer finding: %s", got.stderr)
		}
		t.Logf("missing constructor boxing caught at runtime, sanitize=%t exit=%d", options.Sanitize, got.exitCode)
	}
}

func TestMapUnionEntriesRecordAlias(t *testing.T) {
	t.Parallel()
	checkRuntimeWitness(t, "entries_record_alias.a")
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/entries_record_alias.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "split")
	if err := native.Build(native.C(program), binary, native.Options{Split: true, Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if diff := disagreement(want, got); diff != "" {
		t.Fatalf("split R2 witness: %s\n%s", diff, got.stderr)
	}
}

func TestMapUnionValueContextualArray(t *testing.T) {
	t.Parallel()
	checkMapUnionLayout(t, "contextual_array")
}
