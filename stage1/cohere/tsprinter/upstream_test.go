package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// Not parallel: this is a small pin check on the external formatters, not a port benchmark.
func TestStatementUpstreamDifferences(test *testing.T) {
	directory := test.TempDir()
	root, _ := filepath.Abs(oracles.Repository)
	recordsPath, _ := filepath.Abs("testdata/statement-upstream-differences.json")
	// The same two harnesses over cohere's javascript package are tsprinter-go-oracle-v2, built ahead; the unit only
	// runs it, in the package's directory (#5qykzj5).
	command := bounded(
		test,
		oracles.PrinterGateGoOracle(test)+"/oracle",
		"-test.v",
		"-test.run=^TestAdamicStatementUpstreamDifferences$",
	)
	command.Dir = root + "/cohere/internal/format/javascript"
	command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_UPSTREAM="+recordsPath)
	if output, combinedOutputError := combinedOutput(command); combinedOutputError != nil {
		test.Fatalf("Go: %v %s", combinedOutputError, output)
	} else {
		test.Log(string(output))
	}
	data, readError := os.ReadFile(recordsPath)
	if readError != nil {
		test.Fatal(readError)
	}
	var records []struct{ Source, Go, Prettier string }
	if decodeError := json.Unmarshal(data, &records); decodeError != nil {
		test.Fatal(decodeError)
	}
	cases := []map[string]string{}
	var want strings.Builder
	for _, record := range records {
		cases = append(cases, map[string]string{"Source": record.Source, "Want": record.Go})
		want.WriteString(
			"ok\t" + strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(record.Go) + "\n",
		)
	}
	specs, _ := json.Marshal(cases)
	specsPath := filepath.Join(directory, "cases.json")
	if writeError := os.WriteFile(specsPath, specs, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	script, _ := filepath.Abs("testdata/embedded.mjs")
	bundles := root + "/cohere/internal/format/prettier/bundles"
	result := execute(test, nil, "node", script, bundles, specsPath)
	if result.ExitCode != 0 || len(result.Stderr) != 0 || string(result.Stdout) != want.String() {
		test.Fatalf(
			"embedded: exit %d stderr %s diff %s",
			result.ExitCode,
			result.Stderr,
			firstDifference(string(result.Stdout), want.String()),
		)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		test.Skip(
			"set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	script, _ = filepath.Abs("testdata/statementDifferences.mjs")
	result = execute(test, nil, "node", script, library, recordsPath)
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		test.Fatalf("npm: exit %d stderr %s", result.ExitCode, result.Stderr)
	}
	test.Log(string(result.Stdout))
}
