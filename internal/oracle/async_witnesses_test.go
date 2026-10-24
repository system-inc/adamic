package oracle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Refused sources still run independently on Node. The frame ownership cycle is
// forbidden; thenable adoption remains a semantic gap.
func assertAsyncWitnessRefusal(t *testing.T, key, path string, diagnostic error) bool {
	t.Helper()
	var want, output string
	cycle := false
	switch key {
	case "internal/oracle/testdata/async_refused/async_refuse_frame_capture_cycle.a":
		want, output, cycle = "async frame capture cycle", "seed3:held:7\n", true
	case "internal/oracle/testdata/async_refused/async_refuse_return_thenable.a", "internal/oracle/testdata/async_refused/async_refuse_arrow_thenable.a":
		want, output = "return of thenables", "7\n"
	default:
		return false
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || len(observed.stderr) != 0 || string(observed.stdout) != output {
		t.Fatalf("Node refusal witness: %+v", observed)
	}
	if diagnostic == nil || !strings.Contains(diagnostic.Error(), want) {
		t.Fatalf("want named refusal %q: %v", want, diagnostic)
	}
	var refused *lower.Refused
	var notYet *lower.NotYet
	if cycle {
		if !errors.As(diagnostic, &refused) {
			t.Fatalf("cycle must be Refused: %v", diagnostic)
		}
	} else if !errors.As(diagnostic, &notYet) {
		t.Fatalf("unproven semantics must be NotYet: %v", diagnostic)
	}
	t.Logf("Node stdout %q; compiler: %v", observed.stdout, diagnostic)
	return true
}

// An escaped closure keeps a captured cell alive after the async frame finishes.
// Ordinary suspension and callable argument tests have no such escaping reader.
func TestAsyncCapturedBindingSurvivesFinishMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/async_escaped_capture.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	control := native.C(program)
	code := control
	for index, function := range program.Functions {
		if function.Name != "make" {
			continue
		}
		for slot, local := range function.FrameEnvironment {
			if program.Locals[local].Name != "held" || !program.Locals[local].Captured {
				continue
			}
			marker := fmt.Sprintf("static void adamic_async_finish_%d(adamic_generated_frame_%d *frame) {\n", index, index)
			detach := fmt.Sprintf(" void *adamic_test_captured = frame->cells[%d].value.reference; frame->cells[%d].value.reference = NULL; adamic_release(adamic_test_captured);\n", slot, slot)
			code = strings.Replace(code, marker, marker+detach, 1)
		}
	}
	if code == control {
		t.Fatal("no escaped captured binding to clear")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	if result.exitCode == 0 || (!strings.Contains(string(result.stderr), "runtime error:") && !strings.Contains(string(result.stderr), "AddressSanitizer")) {
		t.Fatalf("cleared captured binding survived sanitizer: %+v", result)
	}
}
