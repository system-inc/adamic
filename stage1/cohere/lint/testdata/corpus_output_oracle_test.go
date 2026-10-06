package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func TestAdamicCorpusOutputCapture(t *testing.T) {
	destination := os.Getenv("ADAMIC_OUTPUT_CAPTURE")
	manifest := os.Getenv("ADAMIC_OUTPUT_MANIFEST")
	root := os.Getenv("ADAMIC_OUTPUT_ROOT")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(string(data), "\n")
	captured := []outputCase{}
	for _, fix := range []bool{false, true} {
		started := time.Now()
		summary := runSummary{Rules: 30}
		diagnostics := []rule.Diagnostic{}
		for _, row := range rows {
			if row == "" {
				continue
			}
			summary.FilesInScope++
			summary.FilesChecked++
			fields := strings.Split(row, "\t")
			for len(fields) < 7 {
				fields = append(fields, "")
			}
			path := fields[0]
			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := string(bytes)
			result, err := edit.FixText(path, source, func(name, text string) ([]edit.Proposal, error) {
				return edit.ProposalsFrom(collect(name, text, fields)), nil
			}, 10)
			if err != nil || !result.Converged {
				t.Fatalf("fix %s: %v", path, err)
			}
			if result.Text != source {
				if fix {
					fixedBy := map[string]int{}
					for _, proposal := range result.Applied {
						fixedBy[proposal.RuleName]++
					}
					relative := path
					if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
						relative = rel
					}
					summary.Changed = append(summary.Changed, changedFile{Path: relative, Fixed: true, FixedBy: fixedBy})
					source = result.Text
				} else {
					summary.WouldChange++
				}
			}
			diagnostics = append(diagnostics, collect(path, source, fields)...)
		}
		sort.SliceStable(diagnostics, func(i, j int) bool {
			left, right := diagnostics[i], diagnostics[j]
			if left.SourceFile.FileName() != right.SourceFile.FileName() {
				return left.SourceFile.FileName() < right.SourceFile.FileName()
			}
			if left.Range.Pos() != right.Range.Pos() {
				return left.Range.Pos() < right.Range.Pos()
			}
			if left.Range.End() != right.Range.End() {
				return left.Range.End() < right.Range.End()
			}
			if left.RuleName != right.RuleName {
				return left.RuleName < right.RuleName
			}
			if left.Message.Id != right.Message.Id {
				return left.Message.Id < right.Message.Id
			}
			return left.Message.Description < right.Message.Description
		})
		summary.Findings = len(diagnostics)
		summary.Total = time.Since(started)
		encoded, _ := json.Marshal(summaryAsJSON(summary))
		var facts map[string]any
		json.Unmarshal(encoded, &facts)
		for _, key := range []string{"kind", "schemaVersion", "verdict", "filesFixed", "filesFormatted", "cohered"} {
			delete(facts, key)
		}
		facts["label"] = ""
		facts["formattingSeconds"] = 0
		facts["changed"] = []any{}
		facts["adamic"] = nil
		cache := facts["cache"].(map[string]any)
		cache["missReason"] = ""
		cache["findingsMissReason"] = ""
		facts["gaps"].(map[string]any)["unread"] = ""
		// The captured Go time is a supplied renderer fact, preserved exactly, not normalized.
		module := os.Getenv("ADAMIC_OUTPUT_MODULE")
		source := "import { panic, readTextFile } from 'adamic';\nimport { lintFiles } from " + outputTS(filepath.Join(module, "corpus.ts"), "") + ";\nconst facts: RunSummary = " + outputTS(facts, "") + ";\n"
		source += "const listed = readTextFile(" + outputTS(manifest, "") + ");\nif (listed.kind === 'Error') { panic(listed.message); }\nconst corpus = lintFiles(listed.text.split('\\n'), " + fmt.Sprint(fix) + ", " + outputTS(root, "") + ");\nconst findings: RunFinding[] = corpus.findings;\nconst summary: RunSummary = { ...facts, findings: findings.length, changed: corpus.changed, wouldChange: corpus.wouldChange, filesInScope: corpus.files, checked: corpus.files, cached: 0 };\n"
		record := outputCase{Source: source}
		for _, mode := range []outputMode{outputHuman, outputVerbose, outputJSON} {
			for _, color := range []bool{false, true} {
				for _, phases := range []bool{false, true} {
					activeOutput = outputSettings{Mode: mode, Style: textStyle{color: color}, Phases: phases}
					var out bytes.Buffer
					printChangedFiles(&out, summary.Changed)
					for _, diagnostic := range diagnostics {
						printRuleDiagnostic(&out, diagnostic, nil)
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
		t.Logf("%d files, fix=%t, %d findings, %d changed files, %d would change", summary.FilesInScope, fix, summary.Findings, len(summary.Changed), summary.WouldChange)
	}
	encoded, err := json.Marshal(captured)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "cases.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}
