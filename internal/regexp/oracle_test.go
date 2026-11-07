package regexp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

type oracleCase struct {
	Pattern string
	Flags   string
	Source  string
}

// TestNodeAgreement compares every extracted regexp in the two test262 regexp
// suites with Node. testdata/test262.json was extracted from
// test262 commit 7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd. The clone used to
// generate it is scratch data and is deliberately not part of this repository.
func TestNodeAgreement(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("testdata/test262.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var encoded [][]string
	if decodeError := json.Unmarshal(data, &encoded); decodeError != nil {
		test.Fatal(decodeError)
	}
	cases := make([]oracleCase, 0, len(encoded)+1000)
	for _, fields := range encoded {
		if len(fields) != 3 {
			test.Fatalf("invalid corpus row: %q", fields)
		}
		cases = append(cases, oracleCase{fields[0], fields[1], fields[2]})
	}
	corpusCount := len(cases)
	cases = append(cases, generatedCases()...)

	input, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	command := toolchain.Command("node", "-e", `
const input = require("fs").readFileSync(0, "utf8");
const cases = JSON.parse(input);
process.stdout.write(JSON.stringify(cases.map(x => {
  try { new RegExp(x.Pattern, x.Flags); return true; }
  catch (error) {
    if (!(error instanceof SyntaxError)) throw error;
    return false;
  }
})));`)
	command.Stdin = bytes.NewReader(input)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node oracle: %v\n%s", combinedOutputError, output)
	}
	var valid []bool
	if decodeError := json.Unmarshal(output, &valid); decodeError != nil {
		test.Fatalf("decode Node output: %v\n%s", decodeError, output)
	}
	if len(valid) != len(cases) {
		test.Fatalf("Node returned %d answers for %d cases", len(valid), len(cases))
	}

	disagreements := make([]string, 0)
	for i, testCase := range cases {
		_, parseError := Parse(testCase.Pattern, testCase.Flags)
		ours := parseError == nil
		if ours != valid[i] {
			disagreements = append(disagreements, fmt.Sprintf(
				"%s: pattern=%q flags=%q parser=%v node=%v error=%v",
				testCase.Source,
				testCase.Pattern,
				testCase.Flags,
				ours,
				valid[i],
				parseError,
			))
		}
	}
	if len(disagreements) != 0 {
		test.Fatalf(
			"%d disagreements among %d test262 and %d generated cases:\n%s",
			len(disagreements),
			corpusCount,
			len(cases)-corpusCount,
			strings.Join(disagreements, "\n"),
		)
	}
	test.Logf("agreement: %d test262 patterns and %d generated cases", corpusCount, len(cases)-corpusCount)
}

func generatedCases() []oracleCase {
	atoms := []string{"a", ".", "[a-z]", "[^]", "(a)", "(?:a)", "(?=a)", "(?<=a)", "\\d", "\\p{Letter}"}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "{4,2}", "*?"}
	flags := []string{"", "u", "v", "i", "gimsy", "uv", "uu"}
	cases := make([]oracleCase, 0, len(atoms)*len(quantifiers)*len(flags))
	for _, atom := range atoms {
		for _, quantifier := range quantifiers {
			for _, flag := range flags {
				cases = append(cases, oracleCase{atom + quantifier, flag, "generated"})
			}
		}
	}
	return cases
}
