package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

// adamic_object_spread_with (object_spread.c, #5c0z219) is { ...source, name: value } for a literal whose
// result can't take its source's shape. No program lowers to it until Emit's half of #7nhzq7h, so
// testdata/object-spread/harness.c drives it directly, one witness a mode: Seats' five
// (internal/oracle/testdata/spread_overwrite), the width pin, p01, moduleNameResolver.ts 245:77 over both
// of its sources (:140's, which lacks originalPath, and :2774's, which holds it written as undefined or
// never written) and 279:17's cached resolution. Each mode is held to Node running the same program
// (testdata/object-spread/oracle.js): the result's keys in order, each key's value read by a run-time
// key, an in test, and the source unchanged after, byte for byte. Each is also held to its layout, the
// result's allocation size and every slot's kind, representation byte, initialized byte and write order,
// under ASan and UBSan, and counted, where its allocations must be its frees.
//
// Memory's two musts are rows too: absent_kept, a constructed source's unwritten field kept absent with a
// name appended after its keys, by Object.keys and JSON.stringify, and graph_result_absent, that result
// adopted into a graph region and read after the move; p01 and p01_reverse, a scalar over a scalar and a
// number over a reference, balanced. Runtime's four conditions are rows too: a NULL source and &adamic_null both an empty one; an overwrite
// whose kind and byte don't describe its value stopping in the primitive; a source in a graph region and
// one in the Program region read and never written, their references retained into a heap result that
// outlives them; and a thousand spreads over one layout interning one shape, by the counted build's
// entry count. So are the refused sources, each a NotYet.
//
// Mutants, each a rewrite of object_spread.c, each turning its rows red behind the green baseline
// (TestObjectSpreadWithMutants):
//   - an overwrite written into the source's slot too: number_source's source changes;
//   - shapes interned by their names' addresses, which are left in the source's storage: freed_source
//     reads freed names under ASan, and intern makes a second shape for equal content;
//   - an overwrite's write order set to its slot's index, not published last: resolver_2774_absent puts
//     originalPath before resolvedUsingTsExtension;
//   - an overwritten slot's representation byte taken from the source: p01 reads a MaybeBoolean as a number;
//   - an overwritten slot's kind taken from the source: width reads a number as a pointer, and
//     number_source stops reading undefined as a scalar;
//   - the kind and byte invariant dropped: wrong_byte stops later, in other words;
//   - every accessor on the chain refused again, not only one named as a public field (#vjvcnnj): prototype_getter
//     and prototype_getter_past stop where Node copies past the getter;
//   - an interned shape's reference bit set against its kind, TestShapeProducerInventory's bitmap mutant for
//     a shape made at run time: width and p01 stop in adamic_shape_check;
//   - the result always made by adamic_object_new, never constructed (Memory's must 1): absent_kept shows
//     the field its source never wrote, and resolver_2774_absent orders originalPath by its slot;
//   - a slot an overwrite takes copied from the source anyway, retained by the result's reference bit
//     (Memory's must 2): number_source retains a number's bits as a pointer;
//   - that copy retained by the source's bit instead: p01_reverse retains the string the number replaces,
//     and the counts go unbalanced. With the skip in place the two bits agree on every slot copied, so
//     this mutant drops the skip too, or it changes nothing.

// objectSpreadMode is one mode of the harness: its oracle.js mode, or the words it stops with, and its
// layout lines, the counted build's alone marked by an "interned" line.
type objectSpreadMode struct {
	name, stop string
	layout     []string
	program    bool
}

// objectSpreadSeat is a Seats row's layout: { ...source, k: undefined } holds k as a reference whose byte
// is undefined's, complete over a complete source and published first over one that never wrote k.
var (
	objectSpreadSeat        = []string{"layout: size 66 complete", "layout: k kind 2 byte 13 initialized 1 order -"}
	objectSpreadSeatWritten = []string{"layout: size 72 written", "layout: k kind 2 byte 13 initialized 1 order 0"}
	objectSpreadKept        = []string{
		"layout: size 104 written",
		"layout: a kind 2 byte 3 initialized 1 order 0",
		"layout: b kind 2 byte 0 initialized 0 order absent",
		"layout: c kind 0 byte 1 initialized 1 order 1",
	}
	objectSpreadWritten2774 = []string{
		"layout: size 144 written",
		"layout: path kind 2 byte 3 initialized 1 order 0",
		"layout: extension kind 2 byte 3 initialized 1 order 1",
		"layout: packageId kind 2 byte 13 initialized 1 order 2",
		"layout: originalPath kind 2 byte 3 initialized 1 order 4",
		"layout: resolvedUsingTsExtension kind 0 byte 9 initialized 1 order 3",
	}
)

var objectSpreadModes = []objectSpreadMode{
	{name: "number_source", layout: objectSpreadSeat},
	{name: "number_source_undefined", layout: objectSpreadSeat},
	{name: "number_source_null", layout: objectSpreadSeat},
	{name: "number_present", layout: objectSpreadSeat},
	{name: "optional_string_present", layout: objectSpreadSeat},
	{name: "optional_string_absent", layout: objectSpreadSeatWritten},
	{name: "optional_number_present", layout: objectSpreadSeat},
	{name: "optional_number_absent", layout: objectSpreadSeatWritten},
	{name: "string_source", layout: objectSpreadSeat},
	{name: "string_source_undefined", layout: objectSpreadSeat},
	{name: "width", layout: []string{"layout: size 84 complete", "layout: a kind 0 byte 1 initialized 1 order -", "layout: k kind 0 byte 1 initialized 1 order -"}},
	{name: "width_method", layout: []string{"layout: size 84 complete", "layout: a kind 0 byte 1 initialized 1 order -", "layout: k kind 2 byte 8 initialized 1 order -"}},
	{name: "p01", layout: []string{"layout: size 84 complete", "layout: kind kind 2 byte 3 initialized 1 order -", "layout: value kind 0 byte 9 initialized 1 order -"}},
	{name: "p01_true", layout: []string{"layout: size 84 complete", "layout: kind kind 2 byte 3 initialized 1 order -", "layout: value kind 0 byte 9 initialized 1 order -"}},
	{name: "p01_reverse", layout: []string{"layout: size 84 complete", "layout: kind kind 2 byte 3 initialized 1 order -", "layout: value kind 0 byte 1 initialized 1 order -"}},
	{name: "absent_kept", layout: objectSpreadKept},
	{name: "prototype_getter", layout: []string{"layout: size 84 complete", "layout: k kind 0 byte 1 initialized 1 order -", "layout: g kind 0 byte 1 initialized 1 order -"}},
	{name: "prototype_getter_past", layout: []string{"layout: size 84 complete", "layout: k kind 0 byte 1 initialized 1 order -", "layout: w kind 0 byte 1 initialized 1 order -"}},
	{name: "resolver_140", layout: []string{
		"layout: size 138 complete",
		"layout: path kind 2 byte 3 initialized 1 order -",
		"layout: extension kind 2 byte 3 initialized 1 order -",
		"layout: packageId kind 2 byte 13 initialized 1 order -",
		"layout: resolvedUsingTsExtension kind 0 byte 9 initialized 1 order -",
		"layout: originalPath kind 2 byte 3 initialized 1 order -",
	}},
	{name: "resolver_2774_undefined", layout: []string{
		"layout: size 138 complete",
		"layout: path kind 2 byte 3 initialized 1 order -",
		"layout: extension kind 2 byte 3 initialized 1 order -",
		"layout: packageId kind 2 byte 13 initialized 1 order -",
		"layout: originalPath kind 2 byte 3 initialized 1 order -",
		"layout: resolvedUsingTsExtension kind 0 byte 9 initialized 1 order -",
	}},
	{name: "resolver_2774_absent", layout: objectSpreadWritten2774},
	{name: "cache_279", layout: []string{
		"layout: size 120 written",
		"layout: resolvedModule kind 2 byte 4 initialized 1 order 0",
		"layout: failedLookupLocations kind 2 byte 5 initialized 1 order 2",
		"layout: affectingLocations kind 2 byte 5 initialized 1 order 1",
		"layout: resolutionDiagnostics kind 2 byte 5 initialized 1 order 3",
	}},
	{name: "intern", layout: []string{"layout: one shape", "layout: interned 1"}},
	{name: "loop", layout: []string{"layout: interned 1"}},
	{name: "freed_source", layout: []string{}},
	{name: "graph_source", layout: []string{"layout: size 102 complete", "layout: held kind 2 byte 4 initialized 1 order -", "layout: label kind 2 byte 3 initialized 1 order -", "layout: count kind 0 byte 1 initialized 1 order -"}},
	{name: "program_source", program: true, layout: []string{"layout: size 102 complete", "layout: held kind 2 byte 4 initialized 1 order -", "layout: label kind 2 byte 3 initialized 1 order -", "layout: count kind 0 byte 1 initialized 1 order -"}},
	{name: "graph_result_complete", layout: objectSpreadSeat},
	{name: "graph_result_written", layout: objectSpreadWritten2774},
	{name: "graph_result_absent", layout: objectSpreadKept},
	{name: "refuse_record", stop: "NotYet: spread of record storage"},
	{name: "refuse_host", stop: "NotYet: a spread with overwrites of a runtime library value (a host or groups object)"},
	{name: "refuse_groups", stop: "NotYet: a spread with overwrites of a runtime library value (a host or groups object)"},
	{name: "refuse_error", stop: "NotYet: a spread with overwrites of an Error"},
	{name: "refuse_tuple", stop: "NotYet: a spread with overwrites of a tuple"},
	{name: "refuse_accessor_field", stop: "NotYet: a spread with overwrites of a class instance with an accessor named as a public field"},
	{name: "refuse_value", stop: "NotYet: a spread with overwrites of a value that isn't an object"},
	{name: "wrong_byte", stop: "compiler bug: a spread's overwrite of k states kind 0 and representation 3, which don't describe its value"},
	{name: "wrong_value", stop: "compiler bug: a spread's overwrite of k states kind 2 and representation 3, which don't describe its value"},
	{name: "duplicate", stop: "compiler bug: a spread overwrites k twice"},
}

// objectSpreadMutations are the mutants named at the top of this file: each before becomes its after in
// object_spread.c, and every row in reds goes red.
var objectSpreadMutations = []struct {
	name    string
	changes [][2]string
	reds    []string
}{
	{"source-slot-written", [][2]string{{"\t\tresult->slots[index] = overwrites[at].value;\n", "\t\tif (index < held) ((adamic_object *)source)->slots[index] = overwrites[at].value;\n\t\tresult->slots[index] = overwrites[at].value;\n"}}, []string{"number_source"}},
	{"interned-by-address", [][2]string{
		{"strcmp(entry->shape.names[index], names[index]) == 0;", "entry->shape.names[index] == names[index];"},
		{"\t\t\tmemcpy(text, names[index], length);\n\t\t\tcopied[index] = text;\n", "\t\t\tcopied[index] = names[index];\n"},
	}, []string{"freed_source", "intern"}},
	{"published-in-slot-order", [][2]string{{"\t\tadamic_object_publish(result, index);\n", "\t\tif (result->write_order != NULL) result->write_order[index] = index;\n"}}, []string{"resolver_2774_absent"}},
	{"byte-from-source", [][2]string{{"\t\tadamic_object_field_types(result)[index] = overwrites[at].representation;\n", "\t\tadamic_object_field_types(result)[index] = index < held && adamic_object_present(source, index) ? adamic_object_field_type(source, index) : overwrites[at].representation;\n"}}, []string{"p01"}},
	{"kind-from-source", [][2]string{{"\t\tkinds[index] = overwrites[at].kind;\n", "\t\tif (index >= held) kinds[index] = overwrites[at].kind;\n"}}, []string{"width", "number_source"}},
	{"always-object-new", [][2]string{{"\tadamic_object *result = ordered ? adamic_object_construct(shape) : adamic_object_new(shape);\n\tif (ordered) memcpy(result->write_order, source->write_order, held * sizeof(size_t));\n", "\tadamic_object *result = adamic_object_new(shape);\n"}}, []string{"absent_kept", "resolver_2774_absent"}},
	{"replaced-slot-copied", [][2]string{{"if (taken[index] || !adamic_object_present(source, index)) continue;", "if (!adamic_object_present(source, index)) continue;"}}, []string{"number_source"}},
	{"retain-by-source-bit", [][2]string{
		{"if (taken[index] || !adamic_object_present(source, index)) continue;", "if (!adamic_object_present(source, index)) continue;"},
		{"if (shape->references[index]) adamic_retain(found->reference);", "if (base->references[index]) adamic_retain(found->reference);"},
	}, []string{"p01_reverse"}},
	{"invariant-dropped", [][2]string{{"\t\tif (agrees(&overwrites[at])) continue;\n", "\t\tcontinue;\n"}}, []string{"wrong_byte"}},
	{"accessor-refusal-wide", [][2]string{{"if (strcmp(public_shape->names[index], class->accessors[at].name) == 0) refuse(", "refuse("}}, []string{"prototype_getter", "prototype_getter_past"}},
	{"reference-bit-against-kind", [][2]string{{"\t\t\treferences[index] = kinds[index] == adamic_field_reference;\n", "\t\t\treferences[index] = kinds[index] != adamic_field_reference;\n"}}, []string{"width", "p01"}},
}

func objectSpreadHarness(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "object-spread", "harness.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

type objectSpreadRun struct {
	stdout, stderr string
	exit           int
}

func objectSpreadExecute(t *testing.T, binary, mode string) objectSpreadRun {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, mode)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "LSAN_OPTIONS=exitcode=23", "UBSAN_OPTIONS=halt_on_error=1:print_stacktrace=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	// A binary that never ran says nothing about its row, and must not read as a red one.
	if command.ProcessState == nil {
		t.Fatalf("%s %s didn't run: %v", filepath.Base(binary), mode, err)
	}
	return objectSpreadRun{stdout.String(), stderr.String(), command.ProcessState.ExitCode()}
}

// objectSpreadNode is Node's stdout for a mode.
func objectSpreadNode(t *testing.T, mode string) string {
	t.Helper()
	command := toolchain.Command("node", filepath.Join("testdata", "object-spread", "oracle.js"), mode)
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("Node %s: %v\n%s", mode, err, stderr.String())
	}
	return stdout.String()
}

// objectSpreadRed says why a run of mode is red, or "" when it's green: a stop must stop in its words, and
// any other mode must print Node's lines, then its layout, and a counted run its counts balanced.
func objectSpreadRed(mode objectSpreadMode, node string, counted bool, run objectSpreadRun) string {
	if mode.stop != "" {
		if run.exit != 70 || !strings.Contains(run.stderr, "adamic: panic: "+mode.stop+"\n") {
			return fmt.Sprintf("want exit 70 with %q, got exit %d", mode.stop, run.exit)
		}
		return ""
	}
	if run.exit != 0 {
		return fmt.Sprintf("exit %d", run.exit)
	}
	var shown strings.Builder
	layout := []string{}
	for _, line := range strings.SplitAfter(run.stdout, "\n") {
		if strings.HasPrefix(line, "layout: ") {
			layout = append(layout, strings.TrimSuffix(line, "\n"))
		} else {
			shown.WriteString(line)
		}
	}
	if shown.String() != node {
		return fmt.Sprintf("stdout %q, Node's %q", shown.String(), node)
	}
	want := slices.DeleteFunc(slices.Clone(mode.layout), func(line string) bool {
		return !counted && strings.HasPrefix(line, "layout: interned ")
	})
	if !slices.Equal(layout, want) {
		return fmt.Sprintf("layout %q, want %q", layout, want)
	}
	if counted {
		return objectSpreadUnbalanced(run.stderr)
	}
	return ""
}

// objectSpreadUnbalanced reads a counted run's last line, as leakcheck.Unbalanced does (which this package
// can't import): its allocations must be its frees and its values let go of with their regions.
func objectSpreadUnbalanced(stderr string) string {
	counts := recordGraphCounts.FindStringSubmatch(stderr)
	if counts == nil {
		return "no counts line"
	}
	var values [3]int
	for index := range values {
		value, err := strconv.Atoi(counts[index+1])
		if err != nil {
			return fmt.Sprintf("unreadable counts: %v", err)
		}
		values[index] = value
	}
	if allocations, frees, regions := values[0], values[1], values[2]; allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d, in regions %d)", allocations-frees-regions, allocations, frees, regions)
	}
	return ""
}

func objectSpreadModeNamed(t *testing.T, name string) objectSpreadMode {
	t.Helper()
	index := slices.IndexFunc(objectSpreadModes, func(mode objectSpreadMode) bool { return mode.name == name })
	if index < 0 {
		t.Fatalf("no mode named %s", name)
	}
	return objectSpreadModes[index]
}

func TestObjectSpreadWith(t *testing.T) {
	t.Parallel()
	harness := objectSpreadHarness(t)
	directory := t.TempDir()
	binaries := map[string]native_runtime.Options{
		"sanitized": {Sanitize: true},
		"counted":   {Sanitize: true, Count: true},
		"program":   {Sanitize: true, Count: true, ProgramRegion: true},
	}
	for name, options := range binaries {
		if err := Build(harness, filepath.Join(directory, name), options); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	for _, mode := range objectSpreadModes {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			node := ""
			if mode.stop == "" {
				node = objectSpreadNode(t, mode.name)
			}
			builds := []string{"sanitized", "counted"}
			if mode.program {
				builds = []string{"program"}
			}
			for _, build := range builds {
				run := objectSpreadExecute(t, filepath.Join(directory, build), mode.name)
				if red := objectSpreadRed(mode, node, build != "sanitized", run); red != "" {
					t.Fatalf("%s: %s\nstdout:\n%s\nstderr:\n%s", build, red, run.stdout, run.stderr)
				}
			}
		})
	}
}

// objectSpreadArchive is the sanitized, counted runtime archive with object_spread.c rewritten by changes,
// each of whose text must be there once, or the archive unchanged when there are none.
func objectSpreadArchive(t *testing.T, compiler, library string, flags []string, changes [][2]string) string {
	t.Helper()
	if len(changes) == 0 {
		return library
	}
	source, err := native_runtime.Sources.ReadFile("c/object_spread.c")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, change := range changes {
		if strings.Count(text, change[0]) != 1 {
			t.Fatalf("object_spread.c mutation site moved: %q", change[0])
		}
		text = strings.Replace(text, change[0], change[1], 1)
	}
	directory := t.TempDir()
	path, object := filepath.Join(directory, "object_spread.c"), filepath.Join(directory, "object_spread.o")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-c", path, "-o", object)
	if output, err := productCompilerCommand(t, t.Context(), compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("object_spread.c mutant must compile: %v\n%s", err, output)
	}
	data, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(directory, "runtime.a")
	if err := os.WriteFile(archive, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(native_runtime.ArchiverName(compiler), "rcs", archive, object).CombinedOutput(); err != nil {
		t.Fatalf("object_spread.c mutant archive: %v\n%s", err, output)
	}
	return archive
}

func TestObjectSpreadWithMutants(t *testing.T) {
	t.Parallel()
	harness := objectSpreadHarness(t)
	flags := native_runtime.Flags(native_runtime.Options{Sanitize: true, Count: true})
	compiler, library := runtimeWithFlags(t, flags)
	flags = append(flags, "-I", filepath.Dir(library))
	// The baseline is the same harness linked the same way, so each mutant's red is its change alone.
	baseline := buildHarnessWithLibrary(t, harness, flags, compiler, library)
	for _, mutant := range objectSpreadMutations {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			binary := buildHarnessWithLibrary(t, harness, flags, compiler, objectSpreadArchive(t, compiler, library, flags, mutant.changes))
			for _, name := range mutant.reds {
				mode := objectSpreadModeNamed(t, name)
				node := ""
				if mode.stop == "" {
					node = objectSpreadNode(t, name)
				}
				if red := objectSpreadRed(mode, node, true, objectSpreadExecute(t, baseline, name)); red != "" {
					t.Fatalf("baseline %s is red: %s", name, red)
				}
				red := objectSpreadRed(mode, node, true, objectSpreadExecute(t, binary, name))
				if red == "" {
					t.Fatalf("%s left %s green", mutant.name, name)
				}
				t.Logf("%s reds %s: %s", mutant.name, name, red)
			}
		})
	}
}
