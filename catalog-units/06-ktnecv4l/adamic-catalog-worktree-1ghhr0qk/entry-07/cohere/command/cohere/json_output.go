package main

import (
	"encoding/json"
	"io"
	"sort"
)

// `--json` is the run for a program to read: newline-delimited JSON, one object per line, each with a
// `kind`. A finding is one line, each file the run fixed or formatted is one line, each crash is one
// line, and the summary is the last line, so a reader can stream the findings and still know the run
// ended by seeing it.
//
//	{"kind":"finding","path":"/repo/a.ts","line":3,"column":7,"severity":"error","rule":"nexus/...","messageId":"...","message":"..."}
//	{"kind":"fixed","path":"/repo/b.ts","rules":["prefer-const"]}
//	{"kind":"formatted","path":"/repo/c.ts"}
//	{"kind":"crash","path":"/repo/d.ts","rule":"react-hooks/purity","cause":"..."}
//	{"kind":"summary","schemaVersion":1,"verdict":"pass",...}
//
// It is the second view of runSummary, beside the footer, and the one a consumer such as Structure's
// parity check reads. A consumer that parsed the human text read a layout cohere is free to change; this
// is a contract, versioned by schemaVersion and described in schema/CohereOutput.schema.json.

// outputSchemaVersion is the version of the `--json` contract. A change a reader could notice, such as
// a field removed or renamed or a meaning changed, raises it; a new optional field does not.
const outputSchemaVersion = 1

// findingJSON is one finding's line: the finding's own fields beside its kind.
type findingJSON struct {
	Kind string `json:"kind"`
	runFinding
}

func findingAsJSON(finding runFinding) findingJSON {
	return findingJSON{Kind: "finding", runFinding: finding}
}

// changedFileJSON is one file the run rewrote, by fixing or by formatting.
type changedFileJSON struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
	// Rules are the rules whose fixes were applied, for a fixed file, each once.
	Rules []string `json:"rules,omitempty"`
}

// changedFilesAsJSON is a line per kind of rewrite: a file both fixed and formatted is a `fixed` line and
// a `formatted` line, so a reader counting either kind counts it.
func changedFilesAsJSON(files []changedFile) []changedFileJSON {
	lines := make([]changedFileJSON, 0, len(files))
	for _, file := range files {
		if file.Fixed {
			rules := make([]string, 0, len(file.FixedBy))
			for rule := range file.FixedBy {
				rules = append(rules, rule)
			}
			sort.Strings(rules)
			lines = append(lines, changedFileJSON{Kind: "fixed", Path: file.Path, Rules: rules})
		}
		if file.Formatted {
			lines = append(lines, changedFileJSON{Kind: "formatted", Path: file.Path})
		}
	}
	return lines
}

// summaryJSON is runSummary as the last line. Durations are seconds, as a reader would compare them.
type summaryJSON struct {
	Kind          string      `json:"kind"`
	SchemaVersion int         `json:"schemaVersion"`
	Label         string      `json:"label,omitempty"`
	Verdict       string      `json:"verdict"`
	Seconds       float64     `json:"seconds"`
	GraphSeconds  float64     `json:"graphSeconds"`
	Rules         int         `json:"rules"`
	Phases        []phaseJSON `json:"phases"`
	// FormattingSeconds is the part of the fix phase's seconds with a format in flight, absent when no
	// formatter ran.
	FormattingSeconds float64  `json:"formattingSeconds,omitempty"`
	Cache             cacheUse `json:"cache"`
	TypeErrors        int      `json:"typeErrors"`
	Findings          int      `json:"findings"`
	FilesFixed        int      `json:"filesFixed"`
	FilesFormatted    int      `json:"filesFormatted"`
	WouldChange       int      `json:"wouldChange"`
	// Cohered is the files rewritten; Checked and Cached split FilesInScope between fresh and cached.
	Cohered      int `json:"cohered"`
	FilesInScope int `json:"filesInScope"`
	Checked      int `json:"checked"`
	Cached       int `json:"cached"`
	Nodes        int `json:"nodes"`
	// Gaps is every way the run fell short of checking everything. A reader deciding whether a green run
	// can be trusted reads this, not the verdict alone.
	Gaps runGaps `json:"gaps"`
	// Adamic is the run's readiness, present only under --adamic-readiness (#9tgm3dq). A run that asked and
	// did not lint says not measured, and why, so a reader who asked never takes its absence for a pass.
	Adamic *readinessSummary `json:"adamic,omitempty"`
}

// phaseJSON is one phase's outcome.
type phaseJSON struct {
	Name    string  `json:"name"`
	Outcome string  `json:"outcome"`
	Seconds float64 `json:"seconds"`
	Detail  string  `json:"detail,omitempty"`
}

// summaryAsJSON is the summary's last line.
func summaryAsJSON(summary runSummary) summaryJSON {
	verdict := "pass"
	if summary.failed() {
		verdict = "fail"
	}
	adamic := summary.Adamic
	if adamic == nil && summary.AdamicRequested {
		adamic = notMeasured("lint did not run")
	}
	phases := make([]phaseJSON, 0, len(summary.Phases))
	for _, record := range summary.Phases {
		phases = append(phases, phaseJSON{
			Name:    string(record.Name),
			Outcome: string(record.Outcome),
			Seconds: record.Elapsed.Seconds(),
			Detail:  record.Detail,
		})
	}
	return summaryJSON{
		Kind:              "summary",
		SchemaVersion:     outputSchemaVersion,
		Label:             summary.Label,
		Verdict:           verdict,
		Seconds:           summary.Total.Seconds(),
		GraphSeconds:      summary.Graph.Seconds(),
		Rules:             summary.Rules,
		Phases:            phases,
		FormattingSeconds: summary.Formatting.Seconds(),
		Cache:             summary.Cache,
		TypeErrors:        summary.TypeErrors,
		Findings:          summary.Findings,
		FilesFixed:        summary.filesFixed(),
		FilesFormatted:    summary.filesFormatted(),
		WouldChange:       summary.WouldChange,
		Cohered:           summary.cohered(),
		FilesInScope:      summary.FilesInScope,
		Checked:           summary.FilesChecked,
		Cached:            summary.FilesCached,
		Nodes:             summary.Nodes,
		Gaps:              summary.Gaps,
		Adamic:            adamic,
	}
}

// writeJSONLine writes one object as one line.
func writeJSONLine(out io.Writer, value any) error {
	line, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = out.Write(append(line, '\n'))
	return err
}
