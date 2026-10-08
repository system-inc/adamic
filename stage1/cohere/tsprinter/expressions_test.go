package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func expressionCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	files := printerCorpusFiles(t)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	request, _ := json.Marshal(map[string]any{"Files": files, "Directory": directory, "Gaps": gaps})
	if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/expressions_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/adamic_expressions_test.go": side}})
	path := directory + "/overlay.json"
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+path, "-run=^TestAdamicExpressionCorpus$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_TS_EXPRESSION_REQUEST="+directory+"/request.json")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go expression corpus %v\n%s", err, output)
	} else {
		t.Log(string(output))
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_TS_PRINTER_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"coverage.json", "cases.json", "cases.txt", "answers.txt", "gaps.txt", "gap-answers.txt", "gaps.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

// Not parallel: corpus artifacts use a caller-selected fixed output directory.
func TestExpressionsAgainstGoAndPrettier(t *testing.T) {
	cases, want, specs := expressionCorpus(t)
	port, _ := filepath.Abs("main.ts")
	compare := func(name string, result run) {
		if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != want {
			t.Fatalf("%s exit %d stderr %s diff %s", name, result.exitCode, result.stderr, firstDifference(string(result.stdout), want))
		}
	}
	compare("Node", onNode(t, port, "--cases", cases, "80"))
	program := lowered(t, port)
	result, binary := natively(t, program, "--cases", cases, "80")
	compare("native", result)
	compare("backend", onJavaScriptBackend(t, program, "--cases", cases, "80"))
	if report := leaks(t, program, binary, "--cases", cases, "80"); report != "" {
		t.Fatal(report)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	script, _ := filepath.Abs("testdata/expressions.mjs")
	comparePrinterLibrary(t, "Prettier", execute(t, nil, "node", script, library, specs), specs, "expressions", false)
	embeddedScript, _ := filepath.Abs("testdata/embedded.mjs")
	bundles, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	comparePrinterLibrary(t, "embedded Prettier", execute(t, nil, "node", embeddedScript, bundles, specs), specs, "expressions", true)
	gapDirectory := filepath.Dir(cases)
	gapWant, err := os.ReadFile(gapDirectory + "/gap-answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", onNode(t, port, "--cases", gapDirectory+"/gaps.txt", "80")},
		{"native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, "--cases", gapDirectory+"/gaps.txt", "80")},
		{"backend", onJavaScriptBackend(t, program, "--cases", gapDirectory+"/gaps.txt", "80")},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != string(gapWant) {
			t.Fatalf("%s loud gaps: exit %d stderr %s diff %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), string(gapWant)))
		}
	}
	gapPrettier := execute(t, nil, "node", script, library, gapDirectory+"/gaps.json")
	if gapPrettier.exitCode != 0 || len(gapPrettier.stderr) != 0 {
		t.Fatalf("Go/Prettier gap proof: %s", gapPrettier.stderr)
	}
	release := filepath.Join(t.TempDir(), "release")
	if artifacts := os.Getenv("ADAMIC_TS_PRINTER_ARTIFACTS"); artifacts != "" {
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(artifacts+"/port.c", []byte(native.C(program)), 0644); err != nil {
			t.Fatal(err)
		}
		release = artifacts + "/port"
	}
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	compare("native release", execute(t, nil, release, "--cases", cases, "80"))
	t.Logf("%d unported shapes return NotYet on native, Node and backend; Go and Prettier format every proving input", strings.Count(string(gapWant), "\n"))
	t.Logf("%d expression fragments byte-identical", strings.Count(want, "\n"))
}
