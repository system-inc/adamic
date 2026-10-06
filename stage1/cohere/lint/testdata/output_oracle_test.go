package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/report"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type outputCase struct {
	Source  string
	Driver  string
	Answers []outputAnswer
}

func TestAdamicAuxiliaryCapture(t *testing.T) {
	module := os.Getenv("ADAMIC_OUTPUT_MODULE")
	captured := []outputCase{}
	for index := 0; index < 6; index++ {
		elapsed := []time.Duration{0, 750 * time.Millisecond, 1125 * time.Millisecond, 1250 * time.Millisecond, 2 * time.Second, -62500 * time.Microsecond}[index]
		facts := overallFacts{Total: elapsed, ProjectsByEngine: map[string]int{"TypeScript": 2, "Swift": 1}, Root: "repo"}
		if index == 1 {
			facts.Failed = []string{"web (exit 1)"}
			facts.Unfinished = 1
		}
		if index == 2 {
			facts.ProjectsByEngine = map[string]int{}
			facts.NotChecked = []string{"package not run"}
		}
		if index == 3 {
			facts.ProjectsByEngine = map[string]int{"TypeScript": 1001}
		}
		text := "é\r\nx\u2028y\n😀;"
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/b.ts", Path: tspath.Path("/b.ts")}, text, core.ScriptKindTS)
		diagnostics := []rule.Diagnostic{}
		findings := []any{}
		for ordinal, start := range []int{len(text), 4, 5, -1, len(text) + 1} {
			path, source := "/b.ts", any(text)
			diagnostic := rule.Diagnostic{SourceFile: file, RuleName: fmt.Sprintf("r%d", ordinal), Range: core.NewTextRange(start, start), Message: rule.Message{Description: "raw\nmessage <&>"}}
			if ordinal == 4 {
				diagnostic.SourceFile = nil
				source = nil
				path = "ignored"
			}
			diagnostics = append(diagnostics, diagnostic)
			findings = append(findings, map[string]any{"path": path, "source": source, "start": start, "rule": diagnostic.RuleName, "message": diagnostic.Message.Description})
		}
		if index == 0 {
			diagnostics = nil
			findings = []any{}
		}
		coverage := report.Coverage{FilesChecked: index, RulesRun: index, GraphWarm: index%2 == 0, Elapsed: elapsed}
		// Maps are reconstructed explicitly because projects is not a summary DTO field.
		engines := []string{}
		for engine := range facts.ProjectsByEngine {
			engines = append(engines, engine)
		}
		sort.Strings(engines)
		entries := []string{}
		for _, engine := range engines {
			entries = append(entries, "["+outputTS(engine, "")+","+fmt.Sprint(facts.ProjectsByEngine[engine])+"]")
		}
		source := "import { overallFooter, crashLine, legacyReport, nothingChecked } from " + outputTS(filepath.Join(module, "other.ts"), "") + ";\nimport type { Overall, LegacyFinding, Coverage } from " + outputTS(filepath.Join(module, "other.ts"), "") + ";\nimport { crashJSON } from " + outputTS(filepath.Join(module, "render.ts"), "") + ";\n"
		source += "const overall: Overall = { seconds: " + fmt.Sprint(elapsed.Seconds()) + ", projects: new Map<string, number>([" + strings.Join(entries, ",") + "]), failed: " + outputTS(facts.Failed, "") + ", unfinished: " + fmt.Sprint(facts.Unfinished) + ", notChecked: " + outputTS(facts.NotChecked, "") + ", root: 'repo' };\n"
		// nil Go slices are empty collections here, not absent TypeScript fields.
		source = strings.ReplaceAll(source, "failed: null", "failed: []")
		source = strings.ReplaceAll(source, "notChecked: null", "notChecked: []")
		source += "const legacy: LegacyFinding[] = " + outputTS(findings, "") + ";\nconst coverage: Coverage = { filesChecked: " + fmt.Sprint(index) + ", rulesRun: " + fmt.Sprint(index) + ", graphWarm: " + fmt.Sprint(coverage.GraphWarm) + ", seconds: " + fmt.Sprint(elapsed.Seconds()) + " };\n"
		driver := "const args = programArguments();\nconst color = args[1] === 'true';\nconsole.log(overallFooter(overall, color));\nconsole.log(crashLine('a.ts', '', 'lost <&>', color));\nconsole.log(crashLine('b.ts', 'no-var', 'broken', color));\nconsole.log(crashJSON('a.ts', '', 'lost <&>'));\nconsole.log(crashJSON('b.ts', 'no-var', 'broken'));\nconsole.log(legacyReport(legacy, coverage).slice(0, -1));\nconsole.log(nothingChecked('no files').slice(0, -1));\nprocess.exitCode = 1;\n"
		record := outputCase{Source: source, Driver: driver}
		for _, color := range []bool{false, true} {
			var out bytes.Buffer
			fmt.Fprintln(&out, overallFooter(facts, textStyle{color: color}))
			for _, crash := range []runCrash{{Path: "a.ts", Cause: "lost <&>"}, {Path: "b.ts", Rule: "no-var", Cause: "broken"}} {
				fmt.Fprintln(&out, crashLine(crash, textStyle{color: color}))
			}
			for _, crash := range []runCrash{{Path: "a.ts", Cause: "lost <&>"}, {Path: "b.ts", Rule: "no-var", Cause: "broken"}} {
				writeJSONLine(&out, crashJSON{Kind: "crash", runCrash: crash})
			}
			report.Write(&out, diagnostics, coverage)
			report.WriteNothingChecked(&out, "no files")
			record.Answers = append(record.Answers, outputAnswer{Mode: "auxiliary", Color: color, Output: out.String(), Exit: 1})
		}
		captured = append(captured, record)
	}
	encoded, err := json.Marshal(captured)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("ADAMIC_OUTPUT_CAPTURE"), "cases.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}

type outputAnswer struct {
	Mode          string
	Color, Phases bool
	Output        string
	Exit          int
}

// This file is overlaid into the pinned command package. It calls the actual renderers,
// with the very same captured run facts exported as TypeScript input, and changes no Go renderer.
func TestAdamicOutputCapture(t *testing.T) {
	destination := os.Getenv("ADAMIC_OUTPUT_CAPTURE")
	if destination == "" {
		t.Fatal("missing capture destination")
	}
	var captured []outputCase
	for index := 0; index < 12; index++ {
		summary := runSummary{
			Total: 250 * time.Millisecond, Graph: 50 * time.Millisecond, Rules: 30,
			FilesInScope: 3926, FilesChecked: 3, FilesCached: 3923, Nodes: 1234567,
			Phases: []phaseRecord{{Name: phaseFix, Outcome: outcomeChecked, Elapsed: 125 * time.Millisecond}, {Name: phaseTypes, Outcome: outcomeRan, Elapsed: 25 * time.Millisecond}, {Name: phaseLint, Outcome: outcomeReused, Detail: "fix's walk"}, {Name: phaseUnused, Outcome: outcomeSkipped, Detail: "not requested"}},
		}
		switch index {
		case 0:
			summary.Total = 0
		case 1:
			summary.Total = 4 * time.Millisecond
			summary.Label = "project <&>"
		case 2:
			summary.Total = 75 * time.Millisecond
			summary.TypeErrors = 1
			summary.Findings = 2
		case 3:
			summary.Total = 10500 * time.Millisecond
			summary.Cache = cacheUse{Replayed: true, FilesReplayed: 3926}
			summary.FilesChecked = 0
			summary.FilesCached = 3926
		case 4:
			summary.Total = 11500 * time.Millisecond
			summary.WouldChange = 1
			summary.Gaps = runGaps{ProgramFiles: 5000, FormattingNotChecked: true, RulesSkippingEverything: 1, Unread: "unread <&>\u2028"}
		case 5:
			summary.Gaps = runGaps{CrashedFiles: 1, RuleCrashes: 2, ModifiedBuild: true}
			summary.Phases[1].Outcome = outcomeNotReached
			summary.Phases[2].Outcome = outcomeSkipped
		case 6:
			summary.Gaps.NothingToCheck = true
			summary.FilesInScope = 0
			summary.FilesChecked = 0
			summary.FilesCached = 0
			summary.Rules = 0
			summary.Nodes = 0
			summary.Phases = nil
		case 7:
			summary.Adamic = notMeasured("types bailed")
			summary.Cache = cacheUse{Off: true, MissReason: "key moved", FindingsMissReason: "changed bytes"}
		case 8:
			summary.Adamic = &readinessSummary{Measured: true, Files: 4387, Ready: 3814, Unmeasured: 2, OptionsMissing: []string{"strict", "noUncheckedIndexedAccess", "exactOptionalPropertyTypes"}, FailingFilesByRule: map[string]int{"z": 2, "<&>": 1}}
		case 9:
			summary.Adamic = &readinessSummary{Measured: true, OptionsMissing: []string{}, FailingFilesByRule: map[string]int{}}
		case 10:
			summary.Formatting = 30 * time.Millisecond
			summary.Phases[0].Outcome = outcomeRan
			summary.Findings = 1
		case 11:
			summary.Total = 1 * time.Nanosecond
			summary.Phases = nil
		}
		if index == 2 || index == 10 {
			for i := 22; i >= 0; i-- {
				summary.Changed = append(summary.Changed, changedFile{Path: fmt.Sprintf("é/😀-%02d.ts", i), Fixed: i%2 == 0, Formatted: i%3 == 0, FixedBy: map[string]int{"z": 1, "a": 2, "b": 2}})
			}
		}
		findings := []runFinding{
			{Path: "\U00010000.ts", Line: 2, Column: 7, Severity: "warn", Rule: "no-var", MessageID: "unexpectedVar", Message: "a warning\nwith advice <&>\u2028\u2029"},
			{Path: "\ue000.ts", Line: 1, Column: 3, Severity: "error", Rule: "eqeqeq", MessageID: "unexpected", Message: "expected ==="},
			{Path: "a.ts", Line: 1, Column: 1, Severity: "error", Rule: "TS2322", Message: "bad type"},
			{Severity: "error", Rule: "formatter", Message: "no location"},
		}
		sort.SliceStable(findings, func(i, j int) bool { return findings[i].Path < findings[j].Path })
		// Model fields include ranges for the upstream ordering contract. These synthetic ranges
		// tie; the corpus test supplies the real diagnostic ranges.
		inputFindings := []any{}
		for _, finding := range findings {
			inputFindings = append(inputFindings, map[string]any{"path": finding.Path, "line": finding.Line, "column": finding.Column, "severity": finding.Severity, "rule": finding.Rule, "messageId": finding.MessageID, "message": finding.Message, "description": finding.Message, "start": 0, "end": 0})
		}
		phases := []any{}
		for _, phase := range summary.Phases {
			phases = append(phases, map[string]any{"name": string(phase.Name), "outcome": string(phase.Outcome), "seconds": phase.Elapsed.Seconds(), "detail": phase.Detail, "findings": phase.Findings, "narrowed": phase.Narrowed})
		}
		changed := []any{}
		for _, file := range summary.Changed {
			changed = append(changed, map[string]any{"path": file.Path, "fixed": file.Fixed, "formatted": file.Formatted, "fixedBy": file.FixedBy})
		}
		adamic := any(nil)
		if summary.Adamic != nil {
			adamic = summary.Adamic
		}
		data := map[string]any{"label": summary.Label, "seconds": summary.Total.Seconds(), "graphSeconds": summary.Graph.Seconds(), "rules": summary.Rules, "phases": phases, "formattingSeconds": summary.Formatting.Seconds(), "cache": map[string]any{"replayed": summary.Cache.Replayed, "filesReplayed": summary.Cache.FilesReplayed, "off": summary.Cache.Off, "missReason": summary.Cache.MissReason, "findingsMissReason": summary.Cache.FindingsMissReason}, "typeErrors": summary.TypeErrors, "findings": summary.Findings, "changed": changed, "wouldChange": summary.WouldChange, "filesInScope": summary.FilesInScope, "checked": summary.FilesChecked, "cached": summary.FilesCached, "nodes": summary.Nodes, "gaps": map[string]any{"programFiles": summary.Gaps.ProgramFiles, "crashedFiles": summary.Gaps.CrashedFiles, "ruleCrashes": summary.Gaps.RuleCrashes, "rulesSkippingEverything": summary.Gaps.RulesSkippingEverything, "formattingNotChecked": summary.Gaps.FormattingNotChecked, "nothingToCheck": summary.Gaps.NothingToCheck, "modifiedBuild": summary.Gaps.ModifiedBuild, "unread": summary.Gaps.Unread}, "adamic": adamic}
		// Fill omitted optional DTO fields for the TypeScript model, without changing Go's facts.
		encoded, _ := json.Marshal(data)
		var normalized map[string]any
		json.Unmarshal(encoded, &normalized)
		if ready, ok := normalized["adamic"].(map[string]any); ok {
			if _, present := ready["reason"]; !present {
				ready["reason"] = ""
			}
		}
		source := "const summary: RunSummary = " + outputTS(normalized, "") + ";\nconst findings: RunFinding[] = " + outputTS(inputFindings, "") + ";\n"
		record := outputCase{Source: source}
		for _, mode := range []outputMode{outputHuman, outputVerbose, outputJSON} {
			for _, color := range []bool{false, true} {
				for _, phases := range []bool{false, true} {
					activeOutput = outputSettings{Mode: mode, Style: textStyle{color: color}, Phases: phases}
					var out bytes.Buffer
					printChangedFiles(&out, summary.Changed)
					for _, finding := range findings {
						printFinding(&out, finding, fmt.Sprintf("%s:%d:%d - %s [%s/%s]\n", finding.Path, finding.Line, finding.Column, singleLineDescription(finding.Message), finding.Rule, finding.MessageID))
					}
					// The no-file rule path has its own verbose spelling in printRuleDiagnostic.
					if mode == outputVerbose {
						out.Reset()
						for _, finding := range findings {
							if finding.Path == "" {
								fmt.Fprintf(&out, "error %s: %s\n", finding.Rule, finding.Message)
							} else {
								fmt.Fprintf(&out, "%s:%d:%d - %s [%s/%s]\n", finding.Path, finding.Line, finding.Column, singleLineDescription(finding.Message), finding.Rule, finding.MessageID)
							}
						}
					}
					if mode == outputJSON {
						writeJSONLine(&out, summaryAsJSON(summary))
					} else {
						fmt.Fprintln(&out, footer(summary, textStyle{color: color}, footerOptions{Phases: phases, Verbose: mode == outputVerbose}))
					}
					name := "human"
					if mode == outputVerbose {
						name = "verbose"
					}
					if mode == outputJSON {
						name = "json"
					}
					status := 0
					if summary.failed() {
						status = 1
					}
					record.Answers = append(record.Answers, outputAnswer{name, color, phases, out.String(), status})
				}
			}
		}
		captured = append(captured, record)
	}
	encoded, err := json.Marshal(captured)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "cases.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}

func outputTS(value any, key string) string {
	if value == nil {
		return "undefined"
	}
	if key == "fixedBy" || key == "failingFilesByRule" {
		encoded, _ := json.Marshal(value)
		var values map[string]any
		json.Unmarshal(encoded, &values)
		keys := []string{}
		for name := range values {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		entries := []string{}
		for _, name := range keys {
			entries = append(entries, "["+outputTS(name, "")+","+outputTS(values[name], "")+"]")
		}
		return "new Map<string, number>([" + strings.Join(entries, ",") + "])"
	}
	switch v := value.(type) {
	case map[string]any:
		keys := []string{}
		for name := range v {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		fields := []string{}
		for _, name := range keys {
			fields = append(fields, outputTS(name, "")+":"+outputTS(v[name], name))
		}
		return "{" + strings.Join(fields, ",") + "}"
	case []any:
		parts := []string{}
		for _, entry := range v {
			parts = append(parts, outputTS(entry, ""))
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		encoded, _ := json.Marshal(value)
		return string(encoded)
	}
}
