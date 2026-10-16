package tsprinter

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func TestDocumentsAgainstGoAndPrettier(t *testing.T) {
	t.Parallel()
	cases, want, specs := oracles.DocumentCorpus(t)
	// The port's lowered program and natives are products, as expressions' and statements' are (#g7a9yd3): the test is
	// only its runs.
	products := printerGateFamilyProducts(t, "documents")
	compare := func(name string, result oracles.Run) {
		if result.ExitCode != 0 || len(result.Stderr) > 0 || string(result.Stdout) != want {
			t.Fatalf("%s exit %d stderr %s diff %s", name, result.ExitCode, result.Stderr, corpusDifference(t, cases, string(result.Stdout), want))
		}
	}
	compare("Node", onNode(t, products.source, cases))
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	compare("native", execute(t, environment, products.sanitized, cases))
	compare("backend", onNode(t, products.backend, cases))
	switch runtime.GOOS {
	case "darwin":
		if report := execute(t, nil, "leaks", "--atExit", "--", products.release, cases); report.ExitCode != 0 {
			t.Fatal(string(report.Stdout))
		}
	case "linux":
		if report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, cases); report.ExitCode != 0 {
			t.Fatalf("exit %d\n%s", report.ExitCode, report.Stderr)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	script, _ := filepath.Abs("testdata/doc.mjs")
	compare("Prettier", execute(t, nil, "node", script, library, specs))
	t.Log("5072 docs identical to Go and Prettier on native, Node and backend; sanitizer and leak checks pass")
}
