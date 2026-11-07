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

func jsonRecordsRuntimeBinary(test *testing.T, mutant string) string {
	test.Helper()
	encoder, readError := os.ReadFile("runtime/json_stringify.c")
	if readError != nil {
		test.Fatal(readError)
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
			test.Fatal("record runtime mutant anchor missing")
		}
		source = strings.Replace(source, old, replacement, 1)
	}
	for _, name := range []string{"adamic_json_stringify_runtime", "adamic_json_stringify"} {
		source = strings.ReplaceAll(source, name+"(", "probe_"+name+"(")
	}
	harness, readError := os.ReadFile("testdata/json-record-runtime/harness.c")
	if readError != nil {
		test.Fatal(readError)
	}
	probe := strings.ReplaceAll(
		string(harness),
		"adamic_json_stringify_runtime(",
		"probe_adamic_json_stringify_runtime(",
	)
	binary := filepath.Join(test.TempDir(), "records")
	if buildError := Build(source+"\n"+probe, binary, Options{Sanitize: true, Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	return binary
}

func jsonRecordsRuntimeCheck(test *testing.T, mutant string) {
	test.Helper()
	binary := jsonRecordsRuntimeBinary(test, mutant)
	path, absolutePathError := filepath.Abs("testdata/json-record-runtime/oracle.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want, nodeError, recordNodeRunError := recordNodeRun(
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		path,
	)
	if recordNodeRunError != nil || nodeError != "" {
		test.Fatalf("source Node: %v\n%s", recordNodeRunError, nodeError)
	}
	got, stderr, recordRunError := recordRun(binary)
	if mutant == "ownership" || mutant == "throw" {
		counts := recordCounts.FindStringSubmatch(stderr)
		if got != want ||
			strings.Contains(stderr, "ERROR: AddressSanitizer:") ||
			strings.Contains(stderr, "runtime error:") ||
			(counts == nil || counts[1] == counts[2]) && !strings.Contains(stderr, "LeakSanitizer") {
			test.Fatalf(
				"ownership mutant must preserve Node output and fail counts or leaks: %v\n%s\n%s",
				recordRunError,
				got,
				stderr,
			)
		}
		test.Log("ownership mutant caught by allocation/free balance or LeakSanitizer")
		return
	}
	if recordRunError != nil {
		test.Fatalf("native: %v\n%s", recordRunError, stderr)
	}
	recordCheckCounts(test, stderr)
	if mutant == "prepare" {
		if got == want || !strings.Contains(got, `"nested":{}`) {
			test.Fatalf("Node must catch skipped preparation: native %q, Node %q", got, want)
		}
		test.Log("skipped preparation: clean sanitizers and counts; caught by source Node")
	} else if got != want {
		test.Fatalf("record runtime differs: native %q, Node %q", got, want)
	}
}

func TestJSONRecordsRuntimePreparedValues(test *testing.T) {
	test.Parallel()
	jsonRecordsRuntimeCheck(test, "")
}

func TestJSONRecordsRuntimePrepareMutant(test *testing.T) {
	test.Parallel()
	jsonRecordsRuntimeCheck(test, "prepare")
}

func TestJSONRecordsRuntimeOwnershipMutant(test *testing.T) {
	test.Parallel()
	jsonRecordsRuntimeCheck(test, "ownership")
}

func TestJSONRecordsRuntimeThrowCleanupMutant(test *testing.T) {
	test.Parallel()
	jsonRecordsRuntimeCheck(test, "throw")
}

func TestJSONRecordsRuntimeCycleGuard(test *testing.T) {
	test.Parallel()
	for _, mutant := range []string{"", "cycle"} {
		binary := jsonRecordsRuntimeBinary(test, mutant)
		timeoutContext, cancel := context.WithTimeout(test.Context(), 10*time.Second)
		command := exec.CommandContext(timeoutContext, binary, "cycle")
		command.Env = append(
			os.Environ(),
			"ASAN_OPTIONS=detect_leaks=0:halt_on_error=1",
			"UBSAN_OPTIONS=halt_on_error=1",
		)
		output, combinedOutputError := command.CombinedOutput()
		cancel()
		if timeoutContext.Err() == context.DeadlineExceeded ||
			combinedOutputError == nil ||
			strings.Contains(string(output), "NotYet: JSON.stringify cyclic container") != (mutant == "") {
			test.Fatalf(
				"cycle mutant=%q: want cycle guard only for control: %v\n%s",
				mutant,
				combinedOutputError,
				output,
			)
		}
	}
}
