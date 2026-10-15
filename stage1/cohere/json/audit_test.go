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

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
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
	cases, err := pinnedCorpusFiles(t, root)
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

// jsonUpstreamAllowed names the inputs Go and Prettier may disagree on: the separator witness and the generated
// separator and empty-input cases above. How each of them differs is known-upstream-differences.txt's, and upstream
// parity's key declares this file, so editing the list rebuilds the verdict.
func jsonUpstreamAllowed(name string) bool {
	return name == "stage1/cohere/json/gaps/numeric-separators.json" || strings.HasPrefix(name, "generated/12/") || strings.HasPrefix(name, "generated/13/") || strings.HasPrefix(name, "generated/14/") || strings.HasPrefix(name, "generated/16/")
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
	goPath := filepath.Join(scratch, "go.json")
	// Cohere with the harness laid over it is a product built ahead, and this unit only runs it in the package's
	// directory (#5qykzj5): unmutated, upstreamParityOracleProduct's test binary; with printer mutants, the same
	// package's test binary with them planted (printerMutantOracle, #ebc67r2).
	oracle := upstreamParityOracleProduct(t)
	if len(mutations) > 0 {
		oracle = printerMutantOracle(t, mutations...)
	}
	command := bounded(t, oracle, "-test.v", "-test.run=^TestAdamicJSONAudit$")
	command.Dir = filepath.Join(cohere, "internal/format/javascript")
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
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment, not PATH's. A
		// program a test built (the Go oracle, a printer mutant) runs as named, with this process's environment.
		command = toolchain.Command(name, arguments...)
	}
	return command
}

// These mutate the upstream Go printer, not an Adamic port. They prove that the external comparison
// notices real formatting changes while both sides still exit successfully.
type printerMutation struct{ name, file, from, to string }

// jsonPrinterMutations are the printer mutants the external comparison must catch, each a change to one file of
// cohere's javascript package.
var jsonPrinterMutations = []printerMutation{
	{"missing final newline", "print_json.go", `return concatIn(path, print("node", nil), hardline)`, `return print("node", nil)`},
	{"missing property space", "print_json.go", `print("key", nil), ": ", print("value", nil)`, `print("key", nil), ":", print("value", nil)`},
	{"wrong filename parser", "printer.go", `return "json-stringify"`, `return "json"`},
}

// printerMutantOracle is cohere's javascript package's test binary with the audit harness laid over it and mutations
// planted, a product built ahead (buildcache.GoTestMutated): the mutated sources live in the product, never in a
// directory of this unit's.
func printerMutantOracle(t *testing.T, mutations ...printerMutation) string {
	t.Helper()
	cohere, err := filepath.Abs("../../../cohere")
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	writeJSON(t, overlayPath, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/javascript/adamic_json_audit_test.go"): side}})
	var planted []buildcache.Mutation
	for _, mutation := range mutations {
		planted = append(planted, buildcache.Mutation{File: "cohere/internal/format/javascript/" + mutation.file, Before: mutation.from, After: mutation.to})
	}
	return buildcache.GoTestMutated(t, "cohere", "javascript.test", "./internal/format/javascript", []string{"-overlay=" + overlayPath}, planted)
}

func TestProduct_JSONPrinterMutant_0(t *testing.T) {
	t.Parallel()
	printerMutantOracle(t, jsonPrinterMutations[0])
}

func TestProduct_JSONPrinterMutant_1(t *testing.T) {
	t.Parallel()
	printerMutantOracle(t, jsonPrinterMutations[1])
}

func TestProduct_JSONPrinterMutant_2(t *testing.T) {
	t.Parallel()
	printerMutantOracle(t, jsonPrinterMutations[2])
}

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
	for _, mutation := range jsonPrinterMutations {
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
