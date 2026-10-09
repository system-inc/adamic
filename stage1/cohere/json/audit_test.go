// Package json holds the Adamic JSON formatter to Go cohere and reports upstream Prettier differences.
package json

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/gatesample"
)

type textCase struct {
	Name string `json:"name"`
	Text string `json:"text"`
}
type answer struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

// Error messages have different APIs. Compare acceptance and successful output, never erase an error
// into a successful empty output. This is a necessary condition for four-way formatter parity.
func sameAnswer(left, right answer) bool {
	if (left.Error == "") != (right.Error == "") {
		return false
	}
	return left.Error != "" || left.Output == right.Output
}

func TestUpstreamNumericSeparatorGap(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the external oracle")
	}
	fixture, err := os.ReadFile("gaps/numeric-separators.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []textCase
	for _, text := range []string{strings.TrimSpace(string(fixture)), "[1_]", "[0x_1]"} {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, textCase{name, text})
		}
	}
	goAnswers, prettierAnswers := oracleAnswers(t, cases)
	for index, item := range cases {
		goAnswer, prettierAnswer := goAnswers[index], prettierAnswers[index]
		if goAnswer.Error != "" || !strings.Contains(goAnswer.Output, item.Text[1:len(item.Text)-1]) {
			t.Fatalf("Go's gap changed for %s %q: %+v; update GAPS.md", item.Name, item.Text, goAnswer)
		}
		if !strings.Contains(prettierAnswer.Error, "numeric separator") || sameAnswer(goAnswer, prettierAnswer) {
			t.Fatalf("Prettier's gap changed for %s %q: %+v; update GAPS.md", item.Name, item.Text, prettierAnswer)
		}
		t.Logf("%s %s: Go formats %q; Prettier refuses: %s", item.Name, item.Text, goAnswer.Output, prettierAnswer.Error)
	}
}

// fileCorpus walks every JSON input, including provisioned files and submodules.
func fileCorpus(root string) ([]textCase, error) {
	var cases []textCase
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		encoded, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(encoded) {
			return fmt.Errorf("non-UTF-8 corpus file %s: decoding needs its own oracle", path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		cases = append(cases, textCase{relative, string(encoded)})
		return nil
	})
	return cases, err
}

// corpusCases validates the whole corpus before landing-gate sampling.
func corpusCases(t *testing.T) []textCase {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	cases, err := fileCorpus(root)
	if err != nil {
		t.Fatal(err)
	}
	// Generated boundary cases exercise both filename-selected printers, including the extension's
	// JavaScript literal forms that a JSON.parse/JSON.stringify substitute would lose.
	generated := []string{
		"[]", "{}", "null", "true", "false", `{"a":1,"b":[2,3]}`,
		strings.Repeat("[", 500) + "0" + strings.Repeat("]", 500),
		"[" + strings.Repeat("1,", 4999) + "1]",
		`["\u0000","\uD800","\uD83D\uDE00","é😀","\\","\""]`,
		`[0,-0,1,-1,1.50,1E+03,1e-003,1e309,1e-400,9007199254740993]`,
		`[+1,.5,1.,0x1f,0o17,0b101,1_000,Infinity,NaN,undefined]`,
		"// leading\n{a:1,/* inside */b:[1,2,],}",
		`[1__0]`, `[1_]`, `[0x_1]`, `{"a":`, ``,
	}
	for index, text := range generated {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, textCase{fmt.Sprintf("generated/%d/%s", index, name), text})
		}
	}
	t.Logf("full corpus: %d cases before gate sampling", len(cases))
	verifyCorpusPin(t, cases)
	return cases
}

// Prettier is a separate upstream report. Only the nine named disagreements are known;
// an added difference or a closed difference requires updating the report explicitly.
func jsonUpstreamTopShard(t *testing.T, target int) {
	if err := gatesample.Validate(); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the separate upstream report")
	}
	if !portMatchesShared.ready {
		t.Fatal("shared setup was not selected")
	}
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()

	cases, shards := jsonTopCorpus(t)
	if len(shards) != testUpstreamRepositoryCorpusParityShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testUpstreamRepositoryCorpusParityShards)
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	knownBytes, err := os.ReadFile("known-upstream-differences.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(knownBytes), "\n"), "\n")
	if len(lines) != 27 {
		t.Fatalf("known upstream report has %d lines, want 27", len(lines))
	}
	known := map[string]string{}
	for i := 0; i < len(lines); i += 3 {
		id := lines[i]
		if _, exists := known[id]; exists {
			t.Fatalf("known report repeats %s", id)
		}
		known[id] = strings.Join(lines[i:i+3], "\n") + "\n"
	}
	ids := map[string]bool{}
	for _, item := range cases {
		ids[item.Name] = true
	}
	for id := range known {
		if !ids[id] {
			t.Fatalf("known difference missing from corpus: %s", id)
		}
	}
	oracle := filepath.Join(jsonTopOracle(t), "go-cohere")
	reports := make([]string, len(shards))
	for ordinal, shard := range shards {
		if ordinal != target || ordinal%count != index {
			continue
		}
		func(t *testing.T) {
			items := cases[shard.start:shard.end]
			if len(items) == 0 {
				t.Log("empty hash bucket")
				return
			}
			goAnswers := jsonOracleAnswers(t, oracle, items)
			directory := t.TempDir()
			casesPath := filepath.Join(directory, "cases.json")
			answersPath := filepath.Join(directory, "prettier.json")
			writeJSON(t, casesPath, items)
			// A cold process must accept the same deep nesting as the warmed whole-corpus process.
			result := execute(t, nil, "node", "--stack-size=4096", "testdata/library.mjs", os.Getenv("ADAMIC_JSON_PRETTIER"), casesPath, answersPath)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("%s Prettier exit %d: %s", t.Name(), result.exitCode, result.stderr)
			}
			encoded, err := os.ReadFile(answersPath)
			if err != nil {
				t.Fatal(err)
			}
			var prettierAnswers []answer
			if err := json.Unmarshal(encoded, &prettierAnswers); err != nil {
				t.Fatal(err)
			}
			if len(prettierAnswers) != len(items) {
				t.Fatalf("%s Prettier answered %d of %d", t.Name(), len(prettierAnswers), len(items))
			}
			var report strings.Builder
			for i, item := range items {
				block := ""
				if !sameAnswer(goAnswers[i], prettierAnswers[i]) {
					allowed := item.Name == "stage1/cohere/json/gaps/numeric-separators.json" || strings.HasPrefix(item.Name, "generated/12/") || strings.HasPrefix(item.Name, "generated/13/") || strings.HasPrefix(item.Name, "generated/14/") || strings.HasPrefix(item.Name, "generated/16/")
					if !allowed {
						t.Errorf("%s unexpected upstream difference: %s", t.Name(), item.Name)
					}
					block = fmt.Sprintf("%s\nGo: %q error=%q\nPrettier: %q error=%q\n", item.Name, goAnswers[i].Output, goAnswers[i].Error, prettierAnswers[i].Output, prettierAnswers[i].Error)
				}
				if err := jsonUpstreamBlockCheck(t.Name(), item.Name, block, known[item.Name]); err != nil {
					t.Error(err)
				}
				report.WriteString(block)
			}
			reports[ordinal] = report.String()
			t.Logf("case range [%d:%d]; %d cases", shard.start, shard.end, len(items))
		}(t)
	}
	t.Cleanup(func() { jsonTopReport(t, target, reports[target]) })
	t.Logf("exact union: %d cases; all nine checked-in differences assigned", len(cases))
}

func oracleAnswers(t *testing.T, cases []textCase, mutations ...printerMutation) ([]answer, []answer) {
	return cohereAnswers(t, cases, true, mutations...)
}
func cohereAnswers(t *testing.T, cases []textCase, external bool, mutations ...printerMutation) ([]answer, []answer) {
	t.Helper()
	library := os.Getenv("ADAMIC_JSON_PRETTIER")
	if !external {
		library = ""
	}
	scratch := t.TempDir()
	casesPath := filepath.Join(scratch, "cases.json")
	writeJSON(t, casesPath, cases)
	cohere, err := filepath.Abs("../../../cohere")
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(scratch, "overlay.json")
	replacements := map[string]string{
		filepath.Join(cohere, "internal/format/javascript/adamic_json_audit_test.go"): side,
	}
	for _, mutation := range mutations {
		path := filepath.Join(cohere, "internal/format/javascript", mutation.file)
		encoded, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := string(encoded)
		if strings.Count(source, mutation.from) != 1 {
			t.Fatalf("mutant %s must change one place", mutation.name)
		}
		mutated := filepath.Join(scratch, mutation.file)
		if err := os.WriteFile(mutated, []byte(strings.Replace(source, mutation.from, mutation.to, 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		replacements[path] = mutated
	}
	writeJSON(t, overlayPath, map[string]any{"Replace": replacements})
	goPath := filepath.Join(scratch, "go.json")
	command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicJSONAudit$", "./internal/format/javascript")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_JSON_CASES="+casesPath, "ADAMIC_JSON_ANSWERS="+goPath)
	output, err := childguard.CombinedOutput(command, jsonGuard)
	if err != nil {
		t.Fatalf("Go cohere: %v\n%s", err, output)
	}
	t.Logf("Go cohere: %s", output)
	readAnswers := func(path string) []answer {
		encoded, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var answers []answer
		if err := json.Unmarshal(encoded, &answers); err != nil {
			t.Fatal(err)
		}
		if len(answers) != len(cases) {
			t.Fatalf("%s answered %d of %d cases", path, len(answers), len(cases))
		}
		return answers
	}
	if library == "" {
		return readAnswers(goPath), nil
	}
	prettierPath := filepath.Join(scratch, "prettier.json")
	command = bounded(t, "node", "testdata/library.mjs", library, casesPath, prettierPath)
	output, err = childguard.CombinedOutput(command, jsonGuard)
	if err != nil {
		t.Fatalf("Prettier: %v\n%s", err, output)
	}
	t.Logf("Prettier: %s", output)
	return readAnswers(goPath), readAnswers(prettierPath)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}

func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.Command(name, arguments...)
	return command
}

// These mutate the upstream Go printer, not an Adamic port. They prove that the external comparison
// notices real formatting changes while both sides still exit successfully.
type printerMutation struct{ name, file, from, to string }

func TestExternalComparisonCatchesThreePrinterMutants(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the external oracle")
	}
	cases := []textCase{
		{"probe.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"text":"é😀","escaped":"\u0041","empty":[]}`},
		{"probe.json", "// leading\n{\"a\":1}"},
	}
	goAnswers, prettierAnswers := oracleAnswers(t, cases)
	for index, item := range cases {
		if goAnswers[index].Error != "" || prettierAnswers[index].Error != "" || !sameAnswer(goAnswers[index], prettierAnswers[index]) {
			t.Fatalf("control %s: Go %+v, Prettier %+v", item.Name, goAnswers[index], prettierAnswers[index])
		}
	}
	for _, mutation := range []printerMutation{
		{"missing final newline", "print_json.go", `return concatIn(path, print("node", nil), hardline)`, `return print("node", nil)`},
		{"missing property space", "print_json.go", `print("key", nil), ": ", print("value", nil)`, `print("key", nil), ":", print("value", nil)`},
		{"wrong filename parser", "printer.go", `return "json-stringify"`, `return "json"`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()
			changed, external := oracleAnswers(t, cases, mutation)
			caught := 0
			for index, item := range cases {
				if changed[index].Error != "" {
					t.Fatalf("mutant must format successfully, %s: %s", item.Name, changed[index].Error)
				}
				if !sameAnswer(changed[index], external[index]) {
					caught++
					t.Logf("caught on %s: Go %q, Prettier %q", item.Name, changed[index].Output, external[index].Output)
				}
			}
			if caught == 0 {
				t.Fatal("external comparison did not catch the mutant")
			}
		})
	}
}
