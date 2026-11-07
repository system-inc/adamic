package native_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// A suspended generator holds only what its resume will read (Memory, 04:42 on #qk8rztp): each
// yield kills the counted cells not live where it resumes. complete.a's prefix parameter is read
// only before its first yield, so both yields kill it; built is read after the last one, so no
// yield may. The drop walk frees a dead cell either way, so a kill set skipped never leaks and the
// oracle can't see it (M-b, a counts row on Linux); this reads the resume itself. Mutants, each red
// here: the kill set skipped (no kill of prefix), and the kill set taking every reference cell
// (built killed, and TestGeneratorsAgreeWithNode red under UBSan on every fixture).
func TestGeneratorKillSetReleasesOnlyDeadCells(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("../oracle/testdata/generators/complete.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	positions := map[string]int{}
	for _, function := range program.Functions {
		if function.Generator {
			for position, local := range function.FrameEnvironment {
				positions[program.Locals[local].Name] = position
			}
		}
	}
	prefix, hasPrefix := positions["prefix"]
	built, hasBuilt := positions["built"]
	if !hasPrefix || !hasBuilt {
		test.Fatalf("frame cells %v: want prefix and built", positions)
	}
	code := native.C(program)
	start := strings.Index(code, "static void adamic_generator_resume_")
	if start < 0 {
		test.Fatal("no resume function")
	}
	resume := code[start:]
	resume = resume[:strings.Index(resume, "\n}\n")]
	if got := strings.Count(resume, "frame->status = adamic_generator_status_suspended;"); got != 2 {
		test.Fatalf("%d yields in the resume, want 2", got)
	}
	if got := strings.Count(resume, fmt.Sprintf("adamic_generator_kill(frame, %d);", prefix)); got != 2 {
		test.Errorf("prefix, dead after the first yield, killed at %d yields, want 2", got)
	}
	if strings.Contains(resume, fmt.Sprintf("adamic_generator_kill(frame, %d);", built)) {
		test.Errorf("built, read after the last yield, is killed at one")
	}
}
