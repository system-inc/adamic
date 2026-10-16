package tsprinter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func TestDocumentsAgainstGoAndPrettier(t *testing.T) {
	t.Parallel()
	cases, want, specs := oracles.DocumentCorpus(t)
	port, _ := filepath.Abs("docMain.ts")
	compare := func(name string, result oracles.Run) {
		if result.ExitCode != 0 || len(result.Stderr) > 0 || string(result.Stdout) != want {
			t.Fatalf("%s exit %d stderr %s diff %s", name, result.ExitCode, result.Stderr, corpusDifference(t, cases, string(result.Stdout), want))
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
