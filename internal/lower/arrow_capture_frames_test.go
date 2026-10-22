package lower

import (
	"os"
	"testing"
)

func TestArrowFieldCaptureFrames(t *testing.T) {
	t.Parallel()
	arrowCaptureFrames(t, "arrow_this_field.a", 1, 1)
}

func TestArrowSuperCaptureFrames(t *testing.T) {
	t.Parallel()
	arrowCaptureFrames(t, "arrow_this_super.a", 0, 1)
}

func TestArrowStoredCaptureFrames(t *testing.T) {
	t.Parallel()
	arrowCaptureFrames(t, "arrow_this_stored.a", 0, 1)
}

// A constructor discovered inside an unrelated arrow has its own receiver frame.
// Inspect the capture before emission so a malformed C build cannot kill the mutant.
func arrowCaptureFrames(t *testing.T, fixture string, empty, receivers int) {
	t.Helper()
	source, err := os.ReadFile("../oracle/testdata/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	gotEmpty, gotReceivers := 0, 0
	for _, function := range program.Functions {
		if !function.Closure {
			continue
		}
		if len(function.Environment) == 0 {
			gotEmpty++
			continue
		}
		if len(function.Environment) != 1 {
			t.Fatalf("arrow captured unrelated frame: %v", function.Environment)
		}
		local := program.Locals[function.Environment[0]]
		if local.Name != "this" || !local.Captured {
			t.Fatalf("arrow lost receiver capture: %+v", local)
		}
		gotReceivers++
	}
	if gotEmpty != empty || gotReceivers != receivers {
		t.Fatalf("receiver frames: got %d empty and %d receiver arrows, want %d and %d", gotEmpty, gotReceivers, empty, receivers)
	}
}
