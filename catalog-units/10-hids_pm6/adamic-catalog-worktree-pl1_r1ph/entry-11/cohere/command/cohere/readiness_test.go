package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/types/program"
)

// adamicOptionsOn is a tsconfig with every option Adamic sets.
func adamicOptionsOn() *core.CompilerOptions {
	return &core.CompilerOptions{
		Strict:                     core.TSTrue,
		NoUncheckedIndexedAccess:   core.TSTrue,
		ExactOptionalPropertyTypes: core.TSTrue,
	}
}

// TestReadinessCountsAFileReadyOnlyWhenNothingFailsIt: a file is ready with no set finding and no type
// error, and a file a set rule skipped, or one with no record (never measured), is never ready.
func TestReadinessCountsAFileReadyOnlyWhenNothingFailsIt(t *testing.T) {
	t.Parallel()
	clean := []program.AdamicCount{{Rule: "adamic/single-spread"}, {Rule: "adamic/nominal-class"}}
	records := map[string]*program.AdamicRecord{
		"/ready.ts":       {Counts: clean},
		"/typeError.ts":   {Counts: clean},
		"/finding.ts":     {Counts: []program.AdamicCount{{Rule: "adamic/single-spread", Findings: 2}, {Rule: "adamic/nominal-class"}}},
		"/skipped.ts":     {Counts: clean, Skipped: []string{"adamic/nominal-class"}},
		"/replayed.ts":    nil,
		"/alsoFinding.ts": {Counts: []program.AdamicCount{{Rule: "adamic/single-spread", Findings: 1}}},
	}
	summary := measureReadiness(records, map[string]bool{"/typeError.ts": true}, adamicOptionsOn())
	if summary.Files != 6 || summary.Ready != 1 || summary.Unmeasured != 2 {
		t.Errorf("files %d, ready %d, unmeasured %d; want 6, 1 and 2", summary.Files, summary.Ready, summary.Unmeasured)
	}
	if summary.FailingFilesByRule["adamic/single-spread"] != 2 || summary.FailingFilesByRule["adamic/nominal-class"] != 0 {
		t.Errorf("failing files by rule %v, want single-spread 2 and nominal-class none", summary.FailingFilesByRule)
	}
	if got, want := summary.segment(), "16% Adamic-ready (1 of 6; 2 unmeasured)"; got != want {
		t.Errorf("segment %q, want %q", got, want)
	}
}

// TestReadinessNamesTheOptionsAdamicSetsThatTheProjectLacks: measured under weaker options, the segment
// says which, so the share cannot read as Adamic's.
func TestReadinessNamesTheOptionsAdamicSetsThatTheProjectLacks(t *testing.T) {
	t.Parallel()
	records := map[string]*program.AdamicRecord{"/a.ts": {Counts: []program.AdamicCount{{Rule: "adamic/single-spread"}}}}
	for name, fixture := range map[string]struct {
		options *core.CompilerOptions
		segment string
	}{
		"every option": {adamicOptionsOn(), "100% Adamic-ready (1 of 1)"},
		"strict alone": {&core.CompilerOptions{Strict: core.TSTrue}, "100% Adamic-ready (1 of 1; tsconfig lacks noUncheckedIndexedAccess and exactOptionalPropertyTypes)"},
		"strict off":   {&core.CompilerOptions{Strict: core.TSFalse, NoUncheckedIndexedAccess: core.TSTrue, ExactOptionalPropertyTypes: core.TSTrue}, "100% Adamic-ready (1 of 1; tsconfig lacks strict)"},
		"one flag off": {&core.CompilerOptions{Strict: core.TSTrue, StrictNullChecks: core.TSFalse, NoUncheckedIndexedAccess: core.TSTrue, ExactOptionalPropertyTypes: core.TSTrue}, "100% Adamic-ready (1 of 1; tsconfig lacks strict)"},
		"none of them": {&core.CompilerOptions{Strict: core.TSFalse}, "100% Adamic-ready (1 of 1; tsconfig lacks strict, noUncheckedIndexedAccess, and exactOptionalPropertyTypes)"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := measureReadiness(records, nil, fixture.options).segment(); got != fixture.segment {
				t.Errorf("segment %q, want %q", got, fixture.segment)
			}
		})
	}
}

// TestTheFooterAndTheSummarySayReadiness: the segment ends the footer's line before its gaps, a run that
// could not measure says why, and --json carries the object exactly when the run asked for it.
func TestTheFooterAndTheSummarySayReadiness(t *testing.T) {
	t.Parallel()
	measured := runSummary{Total: 2400 * time.Millisecond, Rules: 480, FilesInScope: 4, FilesChecked: 4,
		Adamic: &readinessSummary{Measured: true, Files: 4, Ready: 3, OptionsMissing: []string{}, FailingFilesByRule: map[string]int{}}}
	if got, want := footer(measured, plain, footerOptions{}), "✓ 💎 2.4s (480 rules • 4 checked) • 75% Adamic-ready (3 of 4)"; got != want {
		t.Errorf("footer\n got %q\nwant %q", got, want)
	}
	bailed := runSummary{Total: 1100 * time.Millisecond, TypeErrors: 4, Adamic: notMeasured("types bailed")}
	if got := footer(bailed, plain, footerOptions{}); !strings.HasSuffix(got, " • Adamic readiness not measured: types bailed") {
		t.Errorf("a bailed run's footer does not say readiness went unmeasured: %q", got)
	}
	if got := footer(runSummary{Total: time.Second}, plain, footerOptions{}); strings.Contains(got, "Adamic") {
		t.Errorf("a run that did not lint names readiness in its footer: %q", got)
	}

	// --json carries adamic only under --adamic-readiness (#9tgm3dq), and then always: a run that asked and did
	// not lint says why.
	unasked, err := json.Marshal(summaryAsJSON(runSummary{Total: time.Second}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(unasked), `"adamic"`) {
		t.Errorf("a run that did not ask for readiness carries adamic in its summary: %s", unasked)
	}
	asked, err := json.Marshal(summaryAsJSON(runSummary{Total: time.Second, AdamicRequested: true}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(asked), `"adamic":{"measured":false,"reason":"lint did not run"`) {
		t.Errorf("a run that asked and did not lint has no not-measured adamic object in its summary: %s", asked)
	}
}
