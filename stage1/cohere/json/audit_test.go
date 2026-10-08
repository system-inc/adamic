// Package json holds the Adamic JSON formatter to Go cohere and reports upstream Prettier differences.
package json

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

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
		if !strings.Contains(goAnswer.Error, "numeric separator") || !strings.Contains(prettierAnswer.Error, "numeric separator") || !sameAnswer(goAnswer, prettierAnswer) {
			t.Fatalf("closed numeric-separator gap regressed for %s %q: Go %+v; Prettier %+v", item.Name, item.Text, goAnswer, prettierAnswer)
		}
		t.Logf("%s %s: Go and Prettier both refuse (closed by cohere 1ee5f944): Go %s; Prettier %s", item.Name, item.Text, goAnswer.Error, prettierAnswer.Error)
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

// Prettier is a separate upstream report. Only the two empty-input disagreements are known;
// an added difference or a closed difference requires updating the report explicitly.
func TestUpstreamRepositoryCorpusParity(t *testing.T) {
	if err := gatesample.Validate(); err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	t.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the separate upstream report")
	}
	cases := sampledCorpusCases(t, 8)
	goAnswers, prettierAnswers := oracleAnswers(t, cases)
	differences := 0
	var report strings.Builder
	for index, item := range cases {
		if sameAnswer(goAnswers[index], prettierAnswers[index]) {
			continue
		}
		differences++
		known := strings.HasPrefix(item.Name, "generated/16/")
		if !known {
			t.Errorf("unexpected upstream difference: %s", item.Name)
		}
		fmt.Fprintf(&report, "%s\nGo: %q error=%q\nPrettier: %q error=%q\n", item.Name,
			goAnswers[index].Output, goAnswers[index].Error, prettierAnswers[index].Output, prettierAnswers[index].Error)
		if differences <= 20 {
			t.Logf("difference: %s (Go error=%q, Prettier error=%q)", item.Name, goAnswers[index].Error, prettierAnswers[index].Error)
		}
	}
	if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" {
		if err := os.WriteFile(path, []byte(report.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	knownReport, err := os.ReadFile("known-upstream-differences.txt")
	if err != nil {
		t.Fatal(err)
	}
	if report.String() != string(knownReport) {
		t.Error("upstream difference identities or answers changed; update the checked-in report")
	}
	if differences != 2 {
		t.Fatalf("known upstream report changed: %d differences, want exactly 2", differences)
	}
	t.Logf("exactly two known upstream differences in %d texts", len(cases))
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
	output, err := command.CombinedOutput()
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
	output, err = command.CombinedOutput()
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

func bounded(t *testing.T, name string, arguments ...string) *progressCommand {
	t.Helper()
	command := newProgressCommand(jsonProgressPolicy, name, arguments...)
	t.Cleanup(command.cancel)
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
