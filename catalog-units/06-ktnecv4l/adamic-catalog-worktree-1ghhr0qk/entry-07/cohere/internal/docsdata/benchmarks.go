package docsdata

import (
	"fmt"
	"sort"

	"github.com/system-inc/cohere/internal/benchresults"
)

// Benchmarks is benchmarks.json: every recorded bench/quiet_machine.sh run, newest first, each exactly as
// it was recorded. Nothing here summarizes again or picks a number; the records already say which numbers
// were earned.
type Benchmarks struct {
	SourceNote string                `json:"sourceNote"`
	Records    []benchresults.Record `json:"records"`
}

// benchmarksSourceNote is what the site reads before it shows a number.
const benchmarksSourceNote = "Each record is one bench/quiet_machine.sh --record run, read and validated by " +
	"internal/benchresults. A mode's quiet summary is present only when at least one of its runs was quiet, and " +
	"it is taken from those runs alone: in a schema 2 record, a run whose cores were idle at or over the idle " +
	"floor just before and just after it; in a schema 1 record, one that started and ended at or under the load " +
	"ceiling. A mode without one has no number to report, only its loaded runs, which are shown and never " +
	"reported as the number."

// buildBenchmarks orders the records newest first and validates each again, so a caller handing records in
// directly cannot publish one the reader would refuse.
func buildBenchmarks(records []benchresults.Record) (Benchmarks, error) {
	ordered := append([]benchresults.Record{}, records...)
	for _, record := range ordered {
		if err := benchresults.Validate(record); err != nil {
			return Benchmarks{}, fmt.Errorf("docsdata: benchmark record %s: %w", benchresults.FileName(record), err)
		}
	}
	sort.SliceStable(ordered, func(left, right int) bool {
		if ordered[left].RecordedAt != ordered[right].RecordedAt {
			return ordered[left].RecordedAt > ordered[right].RecordedAt
		}
		return benchresults.FileName(ordered[left]) < benchresults.FileName(ordered[right])
	})
	return Benchmarks{SourceNote: benchmarksSourceNote, Records: ordered}, nil
}
