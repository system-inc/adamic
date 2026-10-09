package report

import (
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// The property this whole package exists for: a run that checked nothing must not be able to
// produce the same output as a run that checked everything and found it clean.
func TestNothingCheckedIsNotCleanish(t *testing.T) {
	t.Parallel()
	var checkedNothing strings.Builder
	WriteNothingChecked(&checkedNothing, "no binary for darwin-arm64")

	var cleanRun strings.Builder
	Write(&cleanRun, nil, Coverage{FilesChecked: 3416, RulesRun: 182, Elapsed: 231 * time.Millisecond})

	if strings.Contains(checkedNothing.String(), "✓") {
		t.Fatalf("a run that checked nothing must never render a pass mark: %q", checkedNothing.String())
	}
	if !strings.Contains(cleanRun.String(), "✓") {
		t.Fatalf("a genuinely clean run must render a pass mark: %q", cleanRun.String())
	}
	if checkedNothing.String() == cleanRun.String() {
		t.Fatal("checked-nothing and clean rendered identically, which is the defect this guards")
	}
}

// The coverage line has to carry the numbers, not just the verdict. A reader who cannot see how
// many files were checked cannot tell a narrow run from a full one.
func TestCoverageStatesWhatWasChecked(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	Write(&out, nil, Coverage{
		FilesInProgram: 9530,
		FilesChecked:   3416,
		RulesRun:       182,
		GraphWarm:      true,
		Elapsed:        231 * time.Millisecond,
	})

	rendered := out.String()
	for _, want := range []string{"3416 files", "182 rules", "graph warm", "231ms"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("coverage line missing %q: %s", want, rendered)
		}
	}
}

func TestWriteReturnsWhetherItPassed(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	if passed := Write(&out, nil, Coverage{FilesChecked: 1, RulesRun: 1}); !passed {
		t.Error("no findings should report as passed")
	}

	out.Reset()
	withFinding := []rule.Diagnostic{{
		RuleName: "probe",
		Message:  rule.Message{Id: "probe", Description: "a finding"},
	}}
	if passed := Write(&out, withFinding, Coverage{FilesChecked: 1, RulesRun: 1}); passed {
		t.Error("a finding should report as failed")
	}
	if !strings.Contains(out.String(), "1 finding") {
		t.Errorf("expected a singular finding count, got: %s", out.String())
	}
}

func TestDurationReadsAtAGlance(t *testing.T) {
	t.Parallel()
	cases := map[time.Duration]string{
		231 * time.Millisecond:               "231ms",
		4*time.Second + 500*time.Millisecond: "4.50s",
	}
	for elapsed, want := range cases {
		if got := formatDuration(elapsed); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", elapsed, got, want)
		}
	}
}
