package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func jsonRecordsRuntimeBinary(t *testing.T, mutant string) string {
	t.Helper()
	encoder, err := os.ReadFile("runtime/json_stringify.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(encoder)
	var old, replacement string
	switch mutant {
	case "prepare":
		old = "adamic_json_result result = prepare(w, *slot, schema->element, key, &owned);"
		replacement = "owned = false; adamic_json_result result = {*slot, schema->element};"
	case "ownership":
		old = " || schema->kind == adamic_json_record"
	case "throw":
		old = "adamic_release(ordered); return false;"
		replacement = "return false;"
	case "cycle":
		old = " || s.kind == adamic_json_record"
	}
	if old != "" {
		if !strings.Contains(source, old) {
			t.Fatal("record runtime mutant anchor missing")
		}
		source = strings.Replace(source, old, replacement, 1)
	}
	for _, name := range []string{"adamic_json_stringify_runtime", "adamic_json_stringify"} {
		source = strings.ReplaceAll(source, name+"(", "probe_"+name+"(")
	}
	harness, err := os.ReadFile("testdata/json-record-runtime/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	probe := strings.ReplaceAll(string(harness), "adamic_json_stringify_runtime(", "probe_adamic_json_stringify_runtime(")
	binary := filepath.Join(t.TempDir(), "records")
	if err := Build(source+"\n"+probe, binary, Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func jsonRecordsRuntimeCheck(t *testing.T, mutant string) {
	t.Helper()
	binary := jsonRecordsRuntimeBinary(t, mutant)
	path, err := filepath.Abs("testdata/json-record-runtime/oracle.a")
	if err != nil {
		t.Fatal(err)
	}
	want, nodeError, err := recordNodeRun("--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	if err != nil || nodeError != "" {
		t.Fatalf("source Node: %v\n%s", err, nodeError)
	}
	got, stderr, err := recordRun(binary)
	if mutant == "ownership" || mutant == "throw" {
		counts := recordCounts.FindStringSubmatch(stderr)
		if got != want || strings.Contains(stderr, "ERROR: AddressSanitizer:") || strings.Contains(stderr, "runtime error:") || (counts == nil || counts[1] == counts[2]) && !strings.Contains(stderr, "LeakSanitizer") {
			t.Fatalf("ownership mutant must preserve Node output and fail counts or leaks: %v\n%s\n%s", err, got, stderr)
		}
		t.Log("ownership mutant caught by allocation/free balance or LeakSanitizer")
		return
	}
	if err != nil {
		t.Fatalf("native: %v\n%s", err, stderr)
	}
	recordCheckCounts(t, stderr)
	if mutant == "prepare" {
		if got == want || !strings.Contains(got, `"nested":{}`) {
			t.Fatalf("Node must catch skipped preparation: native %q, Node %q", got, want)
		}
		t.Log("skipped preparation: clean sanitizers and counts; caught by source Node")
	} else if got != want {
		t.Fatalf("record runtime differs: native %q, Node %q", got, want)
	}
}

func TestJSONRecordsRuntimePreparedValues(t *testing.T) {
	t.Parallel()
	jsonRecordsRuntimeCheck(t, "")
}

func TestJSONRecordsRuntimePrepareMutant(t *testing.T) {
	t.Parallel()
	jsonRecordsRuntimeCheck(t, "prepare")
}

func TestJSONRecordsRuntimeOwnershipMutant(t *testing.T) {
	t.Parallel()
	jsonRecordsRuntimeCheck(t, "ownership")
}

func TestJSONRecordsRuntimeThrowCleanupMutant(t *testing.T) {
	t.Parallel()
	jsonRecordsRuntimeCheck(t, "throw")
}

func TestJSONRecordsRuntimeCycleGuard(t *testing.T) {
	t.Parallel()
	for _, mutant := range []string{"", "cycle"} {
		binary := jsonRecordsRuntimeBinary(t, mutant)
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		command := exec.CommandContext(ctx, binary, "cycle")
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := command.CombinedOutput()
		cancel()
		if ctx.Err() == context.DeadlineExceeded || err == nil || strings.Contains(string(output), "NotYet: JSON.stringify cyclic container") != (mutant == "") {
			t.Fatalf("cycle mutant=%q: want cycle guard only for control: %v\n%s", mutant, err, output)
		}
	}
}
