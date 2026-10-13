package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The mutation witnesses are generated families, fixed by their generator's
// seeds and loops. Repository files cannot change their membership. Reuse the
// independent Go oracle while avoiding parsing/formatting unrelated inputs:
// mutantOracleBinary is that oracle with its add hook filtered to change's
// witness families, resolved once by the caller (binary) and run here on an
// empty request.
func mutantOracle(t *testing.T, change mutation, binary string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	gaps, err := filepath.Abs("testdata/notyet.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(map[string]any{"Files": []string{}, "Directory": directory, "Gaps": gaps})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
		t.Fatal(err)
	}
	kind := "Expression"
	if change.entry == "statementsMain.ts" {
		kind = "Statement"
	}
	// The test binary runs in its package's directory, as go test runs it.
	command := bounded(t, binary, "-test.count=1", "-test.timeout=0", "-test.run=^TestAdamic"+kind+"Corpus$")
	command.Dir = root + "/cohere/internal/format/javascript"
	command.Env = append(os.Environ(), "ADAMIC_TS_EXPRESSION_REQUEST="+directory+"/request.json", "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go mutation oracle: %v\n%s", err, output)
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	return directory + "/cases.txt", string(answers)
}

// mutantOracleAnchor is the side files' add hook, which mutantOracleBinary filters to a mutation's witness families.
const mutantOracleAnchor = "add := func(label, source string) { cases = append(cases, portExpressionCase{Label: label, Source: source}) }"

// mutantOracleBinary is change's Go oracle: cohere's internal/format/javascript test binary with the expression side
// file laid over it (and the statement one, for a statement mutation), each add hook filtered to change's witness
// families. It is a buildcache.GoTestMutated product, keyed by the files of the packages go compiles, the overlay's map
// and each planted filter, so it never walks cohere, whose TypeScript submodule a populated tree hashed for every
// lookup (#4ycgm2n). The filter was a scratch copy of the side file, which no product can key.
func mutantOracleBinary(t *testing.T, change mutation) string {
	t.Helper()
	arguments, filters := mutantOracleRequest(t, change)
	return buildcache.GoTestMutated(t, "cohere", "oracle", "./internal/format/javascript", arguments, filters)
}

// mutantOracleInputs is mutantOracleBinary's key, computed without building it.
func mutantOracleInputs(t *testing.T, change mutation) buildcache.Inputs {
	t.Helper()
	arguments, filters := mutantOracleRequest(t, change)
	inputs, err := buildcache.GoTestInputs("cohere", "oracle", "./internal/format/javascript", arguments, filters)
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

// mutantOracleRequest is how mutantOracleBinary builds change's oracle: the overlay argument and each planted filter.
func mutantOracleRequest(t *testing.T, change mutation) ([]string, []buildcache.Mutation) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	families, ok := mutantFamilies[change.name]
	if !ok && change.name != "hashbang loses its refusal" {
		t.Fatalf("missing witness families for %q", change.name)
	}
	allowed := "map[string]bool{"
	for _, family := range families {
		allowed += fmt.Sprintf("%q:true,", family)
	}
	allowed += "}"
	replace := map[string]string{}
	var filters []buildcache.Mutation
	for _, name := range []string{"expressions", "statements"} {
		if name == "statements" && change.entry != "statementsMain.ts" {
			continue
		}
		side := "stage1/cohere/tsprinter/testdata/" + name + "_side_test.go"
		replace[root+"/cohere/internal/format/javascript/adamic_"+name+"_test.go"] = filepath.Join(root, side)
		filters = append(filters, buildcache.Mutation{File: side, Before: mutantOracleAnchor, After: "allowed := " + allowed + "\nadd := func(label, source string) { if allowed[label] { cases = append(cases, portExpressionCase{Label: label, Source: source}) } }"})
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	// The overlay names this checkout's paths: GoTestMutated keys its map and the side files, never its own path.
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	return []string{"-overlay=" + path}, filters
}
