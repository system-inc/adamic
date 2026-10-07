package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: this is a small pin check on the external formatters, not a port benchmark.
func TestStatementUpstreamDifferences(t *testing.T) {
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	side, _ := filepath.Abs("testdata/statements_side_test.go")
	expressions, _ := filepath.Abs("testdata/expressions_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/adamic_expressions_test.go": expressions, root + "/cohere/internal/format/javascript/adamic_statements_test.go": side}})
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	recordsPath, _ := filepath.Abs("testdata/statement-upstream-differences.json")
	command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+path, "-run=^TestAdamicStatementUpstreamDifferences$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_UPSTREAM="+recordsPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go: %v %s", err, output)
	} else {
		t.Log(string(output))
	}
	data, err := os.ReadFile(recordsPath)
	if err != nil {
		t.Fatal(err)
	}
	var records []struct{ Source, Go, Prettier string }
	if err = json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	cases := []map[string]string{}
	var want strings.Builder
	for _, record := range records {
		cases = append(cases, map[string]string{"Source": record.Source, "Want": record.Go})
		want.WriteString("ok\t" + strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(record.Go) + "\n")
	}
	specs, _ := json.Marshal(cases)
	specsPath := filepath.Join(directory, "cases.json")
	if err = os.WriteFile(specsPath, specs, 0644); err != nil {
		t.Fatal(err)
	}
	script, _ := filepath.Abs("testdata/embedded.mjs")
	bundles := root + "/cohere/internal/format/prettier/bundles"
	result := execute(t, nil, "node", script, bundles, specsPath)
	if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want.String() {
		t.Fatalf("embedded: exit %d stderr %s diff %s", result.exitCode, result.stderr, firstDifference(string(result.stdout), want.String()))
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until cloud/setup.sh installs it")
	}
	script, _ = filepath.Abs("testdata/statementDifferences.mjs")
	result = execute(t, nil, "node", script, library, recordsPath)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("npm: exit %d stderr %s", result.exitCode, result.stderr)
	}
	t.Log(string(result.stdout))
}
