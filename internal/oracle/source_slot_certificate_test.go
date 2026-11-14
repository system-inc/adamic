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
func TestSourceSlotCertificate(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/source_slot_certificate/"
	for _, pin := range []struct {
		file, outcome, field string
		checked, unchecked   []int
	}{
		{file: "hazard.ts", outcome: "stops: adamic: panic: field read failed: raw.child.other", field: "child", checked: []int{9}},
		{file: "flow_node.ts", outcome: "agrees", field: "flowNode", unchecked: []int{9, 11}},
		{file: "flow_node_optional.ts", outcome: "notyet: a write to optional field flowNode through a checked view", field: "flowNode"},
		{file: "required_fresh.ts", field: "flowNode"},
		{file: "required_fresh_generic.ts", field: "flowNode"},
		{file: "road_computed_key.ts", field: "child"},
		{file: "road_key_variable.ts", field: "child"},
		{file: "road_destructure_declaration.ts", field: "child"},
		{file: "road_destructure_parameter.ts", field: "child"},
		{file: "road_destructure_for_of.ts", field: "child"},
		{file: "road_destructure_catch.ts", field: "child"},
		{file: "road_destructure_nested.ts", field: "child"},
		{file: "road_destructure_assignment.ts", field: "child"},
		{file: "road_spread.ts", field: "child"},
		{file: "road_object_assign.ts", field: "child"},
		{file: "road_structured_clone.ts", field: "child"},
		{file: "road_object_values.ts", field: "child"},
		{file: "road_object_entries.ts", field: "child"},
		{file: "road_for_in.ts", field: "child"},
		{file: "road_json_stringify.ts", field: "child"},
		{file: "road_reflect_get.ts", field: "child"},
		{file: "road_getter_dispatcher.ts", field: "child"},
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
