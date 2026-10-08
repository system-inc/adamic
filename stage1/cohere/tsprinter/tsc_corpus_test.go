package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type printerCase struct {
	Label, Source, Want string
}

// Not parallel: the audit report has a caller-selected output directory.
func TestTSCCorpusAgreement(t *testing.T) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files, err := trackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
	if err != nil {
		t.Fatal(err)
	}
	for _, statements := range []bool{false, true} {
		family, entry := "expressions", "main.ts"
		if statements {
			family, entry = "statements", "statementsMain.ts"
		}
		t.Run(family, func(t *testing.T) {
			directory := t.TempDir()
			gaps, _ := filepath.Abs("testdata/notyet.json")
			request, _ := json.Marshal(map[string]any{"Files": files, "Directory": directory, "Gaps": gaps})
			requestPath := filepath.Join(directory, "request.json")
			if err := os.WriteFile(requestPath, request, 0644); err != nil {
				t.Fatal(err)
			}
			expressionSide, _ := filepath.Abs("testdata/expressions_side_test.go")
			replacements := map[string]string{root + "/cohere/internal/format/javascript/adamic_expressions_test.go": expressionSide}
			oracle, variable := "TestAdamicExpressionCorpus", "ADAMIC_TS_EXPRESSION_REQUEST"
			if statements {
				statementSide, _ := filepath.Abs("testdata/statements_side_test.go")
				replacements[root+"/cohere/internal/format/javascript/adamic_statements_test.go"] = statementSide
				oracle, variable = "TestAdamicStatementCorpus", "ADAMIC_TS_STATEMENT_REQUEST"
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": replacements})
			overlayPath := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				t.Fatal(err)
			}
			command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+overlayPath, "-run=^"+oracle+"$", "./internal/format/javascript")
			command.Dir = root + "/cohere"
			command.Env = append(os.Environ(), variable+"="+requestPath)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("Go %s corpus: %v\n%s", family, err, output)
			} else {
				t.Log(string(output))
			}
			data, err := os.ReadFile(filepath.Join(directory, "cases.json"))
			if err != nil {
				t.Fatal(err)
			}
			var selected, cases []printerCase
			if err := json.Unmarshal(data, &selected); err != nil {
				t.Fatal(err)
			}
			escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
			var batch strings.Builder
			for _, item := range selected {
				// This focused gate holds repository fragments, not generated cases.
				if strings.HasPrefix(item.Label, root+"/stage3/drivers/tsc/corpus/") {
					cases = append(cases, item)
					batch.WriteString(">" + escape.Replace(item.Source) + "\n")
				}
			}
			path := filepath.Join(directory, "tsc-cases.txt")
			if err := os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
				t.Fatal(err)
			}
			port, _ := filepath.Abs(entry)
			program := lowered(t, port)
			native, binary := natively(t, program, "--cases", path, "80")
			report := make([]map[string]string, 0)
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node", onNode(t, port, "--cases", path, "80")},
				{"native", native},
				{"backend", onJavaScriptBackend(t, program, "--cases", path, "80")},
			} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Errorf("%s %s: exit %d stderr %s", side.name, family, side.result.exitCode, side.result.stderr)
				}
				lines := strings.Split(strings.TrimSuffix(string(side.result.stdout), "\n"), "\n")
				if len(lines) != len(cases) {
					t.Fatalf("%s %s answers: got %d, want %d", side.name, family, len(lines), len(cases))
				}
				for index, item := range cases {
					want := "ok\t" + escape.Replace(item.Want)
					if lines[index] != want {
						label := strings.TrimPrefix(item.Label, root+"/")
						report = append(report, map[string]string{"build": side.name, "file": label, "source": item.Source, "port": lines[index], "go": want})
						t.Errorf("%s %s: port %q; Go cohere %q", side.name, label, lines[index], want)
					}
				}
			}
			if report := leaks(t, program, binary, "--cases", path, "80"); report != "" {
				t.Fatal(report)
			}
			if keep := os.Getenv("ADAMIC_TSC_PRINTER_AUDIT"); keep != "" {
				if err := os.MkdirAll(keep, 0755); err != nil {
					t.Fatal(err)
				}
				data, _ := json.MarshalIndent(report, "", "  ")
				if err := os.WriteFile(filepath.Join(keep, family+".json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				data, _ = json.Marshal(cases)
				if err := os.WriteFile(filepath.Join(keep, family+"-cases.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				coverage, err := os.ReadFile(filepath.Join(directory, "coverage.json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(keep, family+"-coverage.json"), coverage, 0644); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("%s: %d tsc corpus fragments, %d disagreements", family, len(cases), len(report))
		})
	}
}
