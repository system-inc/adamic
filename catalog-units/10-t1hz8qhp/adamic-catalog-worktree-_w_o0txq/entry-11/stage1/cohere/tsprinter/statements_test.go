package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func statementCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	files := printerCorpusFiles(t)
	request, _ := json.Marshal(map[string]any{"Files": files, "Directory": directory})
	if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
		t.Fatal(err)
	}
	expressionSide, _ := filepath.Abs("testdata/expressions_side_test.go")
	statementSide, _ := filepath.Abs("testdata/statements_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/adamic_expressions_test.go": expressionSide, root + "/cohere/internal/format/javascript/adamic_statements_test.go": statementSide}})
	if err := os.WriteFile(directory+"/overlay.json", overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-v", "-count=1", "-overlay="+directory+"/overlay.json", "-run=^TestAdamicStatementCorpus$", "./internal/format/javascript")
	command.Dir = root + "/cohere"
	command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go statement corpus: %v\n%s", err, output)
	} else {
		t.Log(string(output))
	}
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	planCorpus(t, directory+"/cases.txt")
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

// Not parallel: corpus and executable artifacts have caller-selected fixed directories.
func TestStatementsAgainstGoAndPrettier(t *testing.T) {
	cases, want, specs := statementCorpus(t)
	port, _ := filepath.Abs("statementsMain.ts")
	compare := func(name string, result run) {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != want {
			t.Fatalf("%s exit %d stderr %s diff %s", name, result.exitCode, result.stderr, corpusDifference(t, cases, string(result.stdout), want))
		}
	}
	compare("Node", onNode(t, port, "--cases", cases, "80"))
	program := lowered(t, port)
	result, binary := natively(t, program, "--cases", cases, "80")
	compare("native", result)
	compare("backend", onJavaScriptBackend(t, program, "--cases", cases, "80"))
	// These exact refusals are part of the file driver's contract.
	gapInput := ">const x:number=1;\n>const {x}=value;\n>export const x=1;\n>if(x)f();\n>#!/usr/bin/env node\\nf();\n"
	gapWant := "notyet\tvariable-types\nnotyet\tvariable-pattern\nnotyet\tvariable-modifiers\nnotyet\tIfStatement\nnotyet\tcomment-attachment\n"
	gapPath := filepath.Join(t.TempDir(), "statement-gaps.txt")
	if err := os.WriteFile(gapPath, []byte(gapInput), 0644); err != nil {
		t.Fatal(err)
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node gaps", onNode(t, port, "--cases", gapPath, "80")},
		{"native gaps", execute(t, nil, binary, "--cases", gapPath, "80")},
		{"backend gaps", onJavaScriptBackend(t, program, "--cases", gapPath, "80")},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != gapWant {
			t.Fatalf("%s: %s stderr %s", side.name, side.result.stdout, side.result.stderr)
		}
	}

	if report := leaks(t, program, binary, "--cases", cases, "80"); report != "" {
		t.Fatal(report)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	script, _ := filepath.Abs("testdata/expressions.mjs")
	comparePrinterLibrary(t, "npm Prettier", execute(t, nil, "node", script, library, specs), specs, "statements", false)
	embedded, _ := filepath.Abs("testdata/embedded.mjs")
	bundles, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	comparePrinterLibrary(t, "embedded Prettier", execute(t, nil, "node", embedded, bundles, specs), specs, "statements", true)
	release := filepath.Join(t.TempDir(), "release")
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_ARTIFACTS"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		release = filepath.Join(keep, "port")
		if err := os.WriteFile(filepath.Join(keep, "port.c"), []byte(native.C(program)), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := native.Build(native.C(program), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	compare("release", execute(t, nil, release, "--cases", cases, "80"))
	t.Logf("%d statement/program fragments byte-identical", strings.Count(want, "\n"))
}
