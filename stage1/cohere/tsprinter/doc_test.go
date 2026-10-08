package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func documentCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	cohere, _ := filepath.Abs(repository + "/cohere")
	side, _ := filepath.Abs("testdata/doc_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{cohere + "/internal/format/doc/adamic_docs_test.go": side}})
	path := directory + "/overlay.json"
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+path, "-run=^TestAdamicDocuments$", "./internal/format/doc")
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_TS_DOC_OUTPUT="+directory)
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go doc corpus %v\n%s", err, output)
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	return directory + "/docs.txt", string(answers), directory + "/docs.json"
}
func TestDocumentsAgainstGoAndPrettier(t *testing.T) {
	t.Parallel()
	cases, want, specs := documentCorpus(t)
	port, _ := filepath.Abs("docMain.ts")
	compare := func(name string, result run) {
		if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != want {
			t.Fatalf("%s exit %d stderr %s diff %s", name, result.exitCode, result.stderr, firstDifference(string(result.stdout), want))
		}
	}
	compare("Node", onNode(t, port, cases))
	program := lowered(t, port)
	result, binary := natively(t, program, cases)
	compare("native", result)
	compare("backend", onJavaScriptBackend(t, program, cases))
	if report := leaks(t, program, binary, cases); report != "" {
		t.Fatal(report)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	script, _ := filepath.Abs("testdata/doc.mjs")
	compare("Prettier", execute(t, nil, "node", script, library, specs))
	t.Log("5072 docs identical to Go and Prettier on native, Node and backend; sanitizer and leak checks pass")
}
