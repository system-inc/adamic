package oracle

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// Rule C's source-slot certificate (#2msw6t8): a plain write of an object field through a checked view
// is admitted, and every read of the field's name whose type doesn't hold what the slot can (every view
// write and every declaration's type, undefined where one is optional) is checked, so a lie stops loudly
// at the read. Each program here is one road a slot is read by. Its outcome is one of three, and never a
// misread: notyet, lowering's stop; agrees, natively, sanitized and on JavaScript as Node runs it; or
// stops, each backend exiting 70 at a checked read where Node runs on, having printed only what Node
// printed before it. Checked and unchecked are the lines whose reads of field must be checked or must not.
// Mutants (Loom job J-7kv72wc): sourceSlotHeld answering true, or the check dropped for every read with a
// source position, and hazard.ts and the two destructuring stops misread natively (2.535e-321 printed
// for raw.child.other); the declared part of W dropped, and declared_narrower.ts's read goes unchecked.
// structuredClone has no pin: the library Adamic checks against doesn't declare it.
func TestSourceSlotCertificate(t *testing.T) {
	t.Parallel()
	const nullableWrite = "stage 0 can't lower writing a checked object field that may hold undefined or null"
	const uncheckedCast = "Adamic 0.1 refuses a cast the runtime can't check"
	directory := "internal/oracle/testdata/source_slot_certificate/"
	for _, pin := range []struct {
		file, outcome, field string
		checked, unchecked   []int
	}{
		{file: "hazard.ts", outcome: "stops: adamic: panic: field read failed: raw.child.other is not initialized", field: "child", checked: []int{9}},
		{file: "declared_narrower.ts", outcome: "agrees", field: "child", checked: []int{10}},
		{file: "flow_node.ts", outcome: "notyet: 10:1: " + nullableWrite, field: "flowNode"},
		{file: "flow_node_optional.ts", outcome: "notyet: 8:1: " + nullableWrite, field: "flowNode"},
		{file: "required_fresh.ts", outcome: "notyet: 14:13: " + uncheckedCast, field: "flowNode"},
		{file: "required_fresh_generic.ts", outcome: "notyet: 13:12: " + uncheckedCast, field: "flowNode"},
		{file: "road_computed_key.ts", outcome: "notyet: 9:20: Adamic 0.1 refuses view member read by element access", field: "child"},
		{file: "road_key_variable.ts", outcome: "notyet: 10:20: Adamic 0.1 refuses view member read by element access", field: "child"},
		{file: "road_destructure_declaration.ts", outcome: "stops: adamic: panic: field read failed: child.other is not initialized", field: "child", checked: []int{9}},
		{file: "road_destructure_parameter.ts", outcome: "stops: adamic: panic: field read failed: child.other is not initialized", field: "child", checked: []int{9}},
		{file: "road_destructure_for_of.ts", outcome: "notyet: 9:6: stage 0 can't lower a for...of destructuring an object", field: "child"},
		{file: "road_destructure_catch.ts", outcome: "notyet: 12:23: " + uncheckedCast, field: "child"},
		{file: "road_destructure_catch_any.ts", outcome: "notyet: 10:11: stage 0 can't lower throwing an object that isn't an Error", field: "child"},
		{file: "road_destructure_nested.ts", outcome: "notyet: 9:9: stage 0 can't lower a nested destructuring pattern", field: "child"},
		{file: "road_destructure_assignment.ts", outcome: "notyet: 11:20: stage 0 can't lower a read of field other through type { ready: boolean; other: number; }, which may hold a checked view's value", field: "child"},
		{file: "road_spread.ts", outcome: "notyet: 9:16: stage 0 can't lower a whole-object copy", field: "child"},
		{file: "road_object_assign.ts", outcome: "notyet: 9:32: stage 0 can't lower Object.assign adding a field to its target's fixed shape", field: "child"},
		{file: "road_object_values.ts", outcome: "notyet: 9:21: Adamic 0.1 refuses Object.values", field: "child"},
		{file: "road_object_entries.ts", outcome: "notyet: 9:28: Adamic 0.1 refuses Object.entries", field: "child"},
		{file: "road_for_in.ts", outcome: "agrees", field: "child"},
		{file: "road_json_stringify.ts", outcome: "notyet: 7:1: stage 0 can't lower writing a checked object field without its source-slot type certificate", field: "child"},
		{file: "road_reflect_get.ts", outcome: "notyet: 9:23: stage 0 can't lower Reflect.get other than", field: "child"},
		{file: "road_reflect_get_object.ts", outcome: "notyet: 9:42: stage 0 can't lower Reflect.get of a value not held as unknown", field: "child"},
		{file: "road_reflect_get_unknown.ts", outcome: "notyet: 10:42: stage 0 can't lower Reflect.get of a value that may be null or undefined", field: "child"},
		{file: "road_getter_dispatcher.ts", outcome: "notyet: 17:24: stage 0 can't lower a checked view read whose receiver may own an accessor of the same name", field: "child"},
	} {
		t.Run(pin.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, directory+pin.file))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := lowered(t, path)
			var outcome string
			var reads []string
			if err != nil {
				outcome = "notyet: " + strings.ReplaceAll(err.Error(), path+":", "")
				var notYet *lower.NotYet
				var refused *lower.Refused
				if !errors.As(err, &notYet) && !errors.As(err, &refused) {
					t.Errorf("want a stop, got %v", err)
				}
			} else {
				reads = sourceSlotReads(program, pin.field)
				outcome = sourceSlotOutcome(t, program, truth)
			}
			t.Logf("PIN %s: %s reads=%v", pin.file, outcome, reads)
			if strings.HasPrefix(outcome, "misread") {
				t.Errorf("%s misreads the slot: %s", pin.file, outcome)
			}
			if pin.outcome != "" && !strings.HasPrefix(outcome, pin.outcome) {
				t.Errorf("outcome: want %q, got %q", pin.outcome, outcome)
			}
			for _, want := range []struct {
				lines   []int
				checked bool
			}{{pin.checked, true}, {pin.unchecked, false}} {
				for _, line := range want.lines {
					if !slices.Contains(reads, strconv.Itoa(line)+":"+strconv.FormatBool(want.checked)) {
						t.Errorf("line %d: want a read of %s checked=%t, got %v", line, pin.field, want.checked, reads)
					}
				}
			}
		})
	}
}

// sourceSlotOutcome runs a lowered program natively, sanitized and on JavaScript and names what all
// three did against Node's truth: agrees, stops with the first line of the stop, or misread with the
// backend that did. Backends that differ in kind are listed each.
func sourceSlotOutcome(t *testing.T, program *ir.Program, truth run) string {
	t.Helper()
	sanitized, _ := nativelyUncached(t, program)
	runs := map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)}
	outcomes := map[string]string{}
	for backend, got := range runs {
		switch {
		case disagreement(truth, got) == "":
			outcomes[backend] = "agrees"
		case got.exitCode == 70 && bytes.HasPrefix(truth.stdout, got.stdout) && bytes.Contains(got.stderr, []byte("panic: ")):
			outcomes[backend] = "stops: " + strings.SplitN(string(got.stderr), "\n", 2)[0]
		default:
			outcomes[backend] = "misread on " + backend + ": exit " + strconv.Itoa(got.exitCode) + " stdout " + strconv.Quote(string(got.stdout)) + " stderr " + strconv.Quote(string(got.stderr))
		}
	}
	if outcomes["native"] == outcomes["sanitized"] && outcomes["native"] == outcomes["javascript"] {
		return outcomes["native"]
	}
	var each []string
	for _, backend := range []string{"native", "sanitized", "javascript"} {
		if strings.HasPrefix(outcomes[backend], "misread") {
			return outcomes[backend]
		}
		each = append(each, backend+" "+outcomes[backend])
	}
	return strings.Join(each, "; ")
}

// sourceSlotReads is every read of field in a lowered program, each as its source line and whether the
// program checks it (Program.CheckedRead), sorted: "9:true". A read with no source position is "0".
func sourceSlotReads(program *ir.Program, field string) []string {
	var reads []string
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				visit(value.Elem())
			}
		case reflect.Struct:
			if read, ok := value.Interface().(ir.Property); ok && read.Name == field {
				line := "0"
				if parts := strings.Split(read.ViewWhere, ":"); len(parts) >= 3 {
					line = parts[len(parts)-2]
				}
				reads = append(reads, line+":"+strconv.FormatBool(program.CheckedRead(read)))
			}
			for index := 0; index < value.NumField(); index++ {
				if value.Type().Field(index).IsExported() {
					visit(value.Field(index))
				}
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
		}
	}
	visit(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	sort.Strings(reads)
	return reads
}
