package program

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// Readiness is the cohere:adamic set a readiness run measures every file against (#drbrp8c).
//
// A file is Adamic-ready when the type phase reports nothing in it and no cohere:adamic rule finds anything
// in it, at the set's own options, whatever the project's chain says: Adamic will not read a disable
// comment or a departure, so readiness cannot either. So a rule the chain enables is counted as it runs, its
// findings counted before suppression, and a rule the chain leaves off or never names runs measure-only:
// its findings are counted and never reported, never fail the run, and never carry a fix to the edit engine.
//
// A rule the chain enables with options of its own is counted at the chain's options, not the set's. Today
// only ban-ts-comment can differ, and only in a chain that sets it; running it twice to measure the set's
// options is the refinement, if a project shows the difference matters.
type Readiness struct {
	// Options is each set rule's raw options, as the set writes them.
	Options map[string][]json.RawMessage
}

// Measures is whether ruleName is one of the set's.
func (r *Readiness) Measures(ruleName string) bool {
	if r == nil {
		return false
	}
	_, measured := r.Options[ruleName]
	return measured
}

// Findings is the record's total. AdamicRecord itself is the findings cache's, in lint_cache.go.
func (r *AdamicRecord) Findings() int {
	total := 0
	for _, count := range r.Counts {
		total += int(count.Findings)
	}
	return total
}

// Measured is whether the record can say ready or not: a file has one, and no set rule skipped it. A file
// with none was replayed from an entry its run never measured, which the findings cache refuses to a run that
// measures (FindingsReuse.MeasureReadiness), so it reads unmeasured rather than as clean.
func (r *AdamicRecord) Measured() bool {
	return r != nil && len(r.Skipped) == 0
}

// fileReadiness accumulates one file's record while the dispatcher walks it.
type fileReadiness struct {
	counts      map[string]int32
	skipped     []string
	measureOnly map[string]bool
}

// newFileReadiness starts a file's record over the set rules among the rules about to walk.
func newFileReadiness(readiness *Readiness, rules []rule.Rule, measureOnly map[string]bool) *fileReadiness {
	if readiness == nil {
		return nil
	}
	record := &fileReadiness{counts: map[string]int32{}, measureOnly: measureOnly}
	for _, subject := range rules {
		if readiness.Measures(subject.Name) {
			record.counts[subject.Name] = 0
		}
	}
	return record
}

// count adds a finding of ruleName, when it is a set rule.
func (f *fileReadiness) count(ruleName string) {
	if f == nil {
		return
	}
	if current, measured := f.counts[ruleName]; measured {
		f.counts[ruleName] = current + 1
	}
}

// hidden is whether ruleName's findings and notes stay out of the run: a rule running measure-only.
func (f *fileReadiness) hidden(ruleName string) bool {
	return f != nil && f.measureOnly[ruleName]
}

// note keeps a set rule's skip. A skip covered by another check that ran leaves nothing unchecked, so it
// is no gap here either.
func (f *fileReadiness) note(ruleName string, key string) {
	if f == nil {
		return
	}
	if _, measured := f.counts[ruleName]; measured && strings.HasPrefix(key, rule.SkippedNotePrefix) {
		f.skipped = append(f.skipped, ruleName)
	}
}

// record is the file's measurement, sorted, with a pair for every set rule that ran, zeros included.
func (f *fileReadiness) record() *AdamicRecord {
	if f == nil {
		return nil
	}
	record := &AdamicRecord{Counts: make([]AdamicCount, 0, len(f.counts))}
	for name, findings := range f.counts {
		record.Counts = append(record.Counts, AdamicCount{Rule: name, Findings: findings})
	}
	sort.Slice(record.Counts, func(left, right int) bool { return record.Counts[left].Rule < record.Counts[right].Rule })
	record.Skipped = sortedUnique(f.skipped)
	return record
}

func sortedUnique(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	unique := names[:1]
	for _, name := range names[1:] {
		if name != unique[len(unique)-1] {
			unique = append(unique, name)
		}
	}
	return unique
}
