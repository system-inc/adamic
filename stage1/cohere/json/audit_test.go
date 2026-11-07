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

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

type textCase oracles.ComparisonCase

func oracleCases(cases []textCase) []oracles.ComparisonCase {
	result := make([]oracles.ComparisonCase, len(cases))
	for index, item := range cases {
		result[index] = oracles.ComparisonCase(item)
	}
	return result
}
func localCases(cases []oracles.ComparisonCase) []textCase {
	result := make([]textCase, len(cases))
	for index, item := range cases {
		result[index] = textCase(item)
	}
	return result
}

type answer = oracles.ComparisonAnswer

// Error messages have different APIs. Compare acceptance and successful output, never erase an error
// into a successful empty output. This is a necessary condition for four-way formatter parity.
func sameAnswer(left, right answer) bool { return oracles.SameAnswer(left, right) }

// Go cohere refuses a numeric separator out of place in Babel's words since cohere 1ee5f944 (#w6vtsrn), so the
// separator gap Prettier showed is closed: both refuse each of these, under both filenames.
func TestUpstreamNumericSeparatorsAgree(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		test.Skip("set ADAMIC_JSON_PRETTIER for the external oracle")
	}
	fixture, readError := os.ReadFile("gaps/numeric-separators.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var cases []textCase
	for _, text := range []string{strings.TrimSpace(string(fixture)), "[1_]", "[0x_1]"} {
		for _, name := range []string{"probe.json", "package.json"} {
			cases = append(cases, textCase{name, text})
		}
	}
	goAnswers, prettierAnswers := oracleAnswers(test, cases)
	for index, item := range cases {
		goAnswer, prettierAnswer := goAnswers[index], prettierAnswers[index]
		if !strings.Contains(goAnswer.Error, "numeric separator") {
			test.Fatalf("Go no longer refuses %s %q: %+v; update GAPS.md", item.Name, item.Text, goAnswer)
		}
		if !strings.Contains(prettierAnswer.Error, "numeric separator") || !sameAnswer(goAnswer, prettierAnswer) {
			test.Fatalf(
				"Prettier no longer refuses %s %q as Go does: %+v; update GAPS.md",
				item.Name,
				item.Text,
				prettierAnswer,
			)
		}
		test.Logf(
			"%s %s: Go refuses: %s; Prettier refuses: %s",
			item.Name,
			item.Text,
			goAnswer.Error,
			prettierAnswer.Error,
		)
	}
}

// fileCorpus walks every JSON input, including provisioned files and submodules.
func fileCorpus(root string) ([]textCase, error) {
	var cases []textCase
	walkError := filepath.WalkDir(root, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		encoded, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		if !utf8.Valid(encoded) {
			return fmt.Errorf("non-UTF-8 corpus file %s: decoding needs its own oracle", path)
		}
		relative, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			return relativePathError
		}
		relative = filepath.ToSlash(relative)
		cases = append(cases, textCase{relative, string(encoded)})
		return nil
	})
	return cases, walkError
}

// corpusCases validates the whole corpus before landing-gate sampling.
func corpusCases(test *testing.T) []textCase { return localCases(oracles.CorpusCases(test)) }

// jsonUpstreamAllowed names the inputs Go and Prettier may disagree on: the generated empty-input cases above. How
// each of them differs is known-upstream-differences.txt's, and upstream parity's key declares this file, so editing
// the list rebuilds the verdict.

func oracleAnswers(test *testing.T, cases []textCase, mutations ...printerMutation) ([]answer, []answer) {
	return cohereAnswers(test, cases, true, mutations...)
}
func cohereAnswers(
	test *testing.T,
	cases []textCase,
	external bool,
	mutations ...printerMutation,
) ([]answer, []answer) {
	test.Helper()
	library := os.Getenv("ADAMIC_JSON_PRETTIER")
	if !external {
		library = ""
	}
	scratch := test.TempDir()
	casesPath := filepath.Join(scratch, "cases.json")
	writeJSON(test, casesPath, cases)
	cohere, absolutePathError := filepath.Abs("../../../cohere")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	goPath := filepath.Join(scratch, "go.json")
	// Cohere with the harness laid over it is a product built ahead, and this unit only runs it in the package's
	// directory (#5qykzj5): unmutated, upstreamParityOracleProduct's test binary; with printer mutants, the same
	// package's test binary with them planted (printerMutantOracle, #ebc67r2).
	oracle := upstreamParityOracleProduct(test)
	if len(mutations) > 0 {
		oracle = printerMutantOracle(test, mutations...)
	}
	command := bounded(test, oracle, "-test.v", "-test.run=^TestAdamicJSONAudit$")
	command.Dir = filepath.Join(cohere, "internal/format/javascript")
	command.Env = append(os.Environ(), "ADAMIC_JSON_CASES="+casesPath, "ADAMIC_JSON_ANSWERS="+goPath)
	output, combinedOutputError := childguard.CombinedOutput(command, jsonGuard)
	if combinedOutputError != nil {
		test.Fatalf("Go cohere: %v\n%s", combinedOutputError, output)
	}
	test.Logf("Go cohere: %s", output)
	readAnswers := func(path string) []answer {
		encoded, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		var answers []answer
		if decodeError := json.Unmarshal(encoded, &answers); decodeError != nil {
			test.Fatal(decodeError)
		}
		if len(answers) != len(cases) {
			test.Fatalf("%s answered %d of %d cases", path, len(answers), len(cases))
		}
		return answers
	}
	if library == "" {
		return readAnswers(goPath), nil
	}
	prettierPath := filepath.Join(scratch, "prettier.json")
	command = bounded(test, "node", "testdata/library.mjs", library, casesPath, prettierPath)
	output, combinedOutputError = childguard.CombinedOutput(command, jsonGuard)
	if combinedOutputError != nil {
		test.Fatalf("Prettier: %v\n%s", combinedOutputError, output)
	}
	test.Logf("Prettier: %s", output)
	return readAnswers(goPath), readAnswers(prettierPath)
}

func writeJSON(test *testing.T, path string, value any) {
	test.Helper()
	encoded, encodeError := json.Marshal(value)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(path, encoded, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
}

func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
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

// jsonPrinterMutations are the printer mutants the external comparison must catch, each a change to one file of
// cohere's javascript package.
type printerMutation struct{ name, file, from, to string }

var jsonPrinterMutations = func() []printerMutation {
	var result []printerMutation
	for _, mutation := range oracles.PrinterMutations() {
		result = append(result, printerMutation{mutation.Name, mutation.File, mutation.From, mutation.To})
	}
	return result
}()

// printerMutantOracle is cohere's javascript package's test binary with the audit harness laid over it and mutations
// planted, a product built ahead (buildcache.GoTestMutated): the mutated sources live in the product, never in a
// directory of this unit's.
func printerMutantOracle(test *testing.T, mutations ...printerMutation) string {
	var planted []buildcache.Mutation
	for _, mutation := range mutations {
		planted = append(
			planted,
			buildcache.Mutation{
				File:   "cohere/internal/format/javascript/" + mutation.file,
				Before: mutation.from,
				After:  mutation.to,
			},
		)
	}
	return oracles.PrinterMutant(test, planted)
}

func TestProduct_JsonFindsPrinterMutant_0(test *testing.T) {
	test.Parallel()
	printerMutantOracle(test, jsonPrinterMutations[0])
}

func TestProduct_JsonFindsPrinterMutant_1(test *testing.T) {
	test.Parallel()
	printerMutantOracle(test, jsonPrinterMutations[1])
}

func TestProduct_JsonFindsPrinterMutant_2(test *testing.T) {
	test.Parallel()
	printerMutantOracle(test, jsonPrinterMutations[2])
}

func TestExternalComparisonCatchesThreePrinterMutants(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
		test.Skip("set ADAMIC_JSON_PRETTIER for the external oracle")
	}
	cases := []textCase{
		{"probe.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"a":1,"b":[2,3]}`},
		{"package.json", `{"text":"é😀","escaped":"\u0041","empty":[]}`},
		{"probe.json", "// leading\n{\"a\":1}"},
	}
	goAnswers, prettierAnswers := oracleAnswers(test, cases)
	for index, item := range cases {
		if goAnswers[index].Error != "" ||
			prettierAnswers[index].Error != "" ||
			!sameAnswer(goAnswers[index], prettierAnswers[index]) {
			test.Fatalf("control %s: Go %+v, Prettier %+v", item.Name, goAnswers[index], prettierAnswers[index])
		}
	}
	for _, mutation := range jsonPrinterMutations {
		test.Run(mutation.name, func(test *testing.T) {
			test.Parallel()
			changed, external := oracleAnswers(test, cases, mutation)
			caught := 0
			for index, item := range cases {
				if changed[index].Error != "" {
					test.Fatalf("mutant must format successfully, %s: %s", item.Name, changed[index].Error)
				}
				if !sameAnswer(changed[index], external[index]) {
					caught++
					test.Logf(
						"caught on %s: Go %q, Prettier %q",
						item.Name,
						changed[index].Output,
						external[index].Output,
					)
				}
			}
			if caught == 0 {
				test.Fatal("external comparison did not catch the mutant")
			}
		})
	}
}
