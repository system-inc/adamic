package lower

import (
	"os"
	"testing"
)

func TestArrowFieldCaptureFrames(test *testing.T) {
	test.Parallel()
	arrowCaptureFrames(test, "arrow_this_field.a", 1, 1)
}

func TestArrowSuperCaptureFrames(test *testing.T) {
	test.Parallel()
	arrowCaptureFrames(test, "arrow_this_super.a", 0, 1)
}

func TestArrowStoredCaptureFrames(test *testing.T) {
	test.Parallel()
	arrowCaptureFrames(test, "arrow_this_stored.a", 0, 1)
}

// A constructor discovered inside an unrelated arrow has its own receiver frame.
// Inspect the capture before emission so a malformed C build cannot kill the mutant.
func arrowCaptureFrames(test *testing.T, fixture string, empty, receivers int) {
	test.Helper()
	source, err := os.ReadFile("../oracle/testdata/" + fixture)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowerSource(test, string(source))
	if err != nil {
		test.Fatal(err)
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
			test.Fatalf("arrow captured unrelated frame: %v", function.Environment)
		}
		local := program.Locals[function.Environment[0]]
		if local.Name != "this" || !local.Captured {
			test.Fatalf("arrow lost receiver capture: %+v", local)
		}
		gotReceivers++
	}
	if gotEmpty != empty || gotReceivers != receivers {
		test.Fatalf(
			"receiver frames: got %d empty and %d receiver arrows, want %d and %d",
			gotEmpty,
			gotReceivers,
			empty,
			receivers,
		)
	}
}
