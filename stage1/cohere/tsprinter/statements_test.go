package tsprinter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
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
	oracleDir := statementProduct(t, statementInputs{
		Name: "Go statement oracle", Files: []string{expressionSide, statementSide, root + "/cohere"},
		Flags: []string{"test", "-c", "-overlay"}, Toolchain: "go",
	}, func(dir string) error {
		command := bounded(t, "go", "test", "-c", "-o="+dir+"/oracle", "-overlay="+directory+"/overlay.json", "./internal/format/javascript")
		command.Dir = root + "/cohere"
		output, err := combinedOutput(command)
		if err != nil {
			return fmt.Errorf("%w: %s", err, output)
		}
		return nil
	})
	command := bounded(t, oracleDir+"/oracle", "-test.v", "-test.count=1", "-test.run=^TestAdamicStatementCorpus$", "-test.timeout=3h")
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

const testStatementsAgainstGoAndPrettierShards = 16

// ADAMIC_TEST_SHARD=i/n runs the shards whose number modulo n is i;
// unset runs all shards. Build products are prepared once and shared by leaves.
func TestStatementsAgainstGoAndPrettier(t *testing.T) {
	setup := time.Now()
	cpu := statementCPU()
	t.Cleanup(func() { t.Logf("CPU including builds: %.3fs", statementCPU()-cpu) })
	cases, want, specs := statementCorpus(t)
	shards := statementShards(t, cases, want, specs)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	port, _ := filepath.Abs("statementsMain.ts")
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	parserFiles, err := filepath.Glob("../../typescript/parser/*.ts")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, parserFiles...)
	var program *ir.Program
	var code string
	loweredDir := statementProduct(t, statementInputs{Name: "lowered program", Files: files, Toolchain: "Adamic Go checker/lowerer"}, func(dir string) error {
		program = lowered(t, port)
		code = native.C(program)
		if err := os.WriteFile(filepath.Join(dir, "port.c"), []byte(code), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	nativeDir := statementProduct(t, statementInputs{Name: "sanitized native", Files: []string{filepath.Join(loweredDir, "port.c")}, Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: "clang"}, func(dir string) error {
		return native.Build(code, filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	releaseDir := statementProduct(t, statementInputs{Name: "release native", Files: []string{filepath.Join(loweredDir, "port.c")}, Flags: native.Flags(native.Options{}), Toolchain: "clang"}, func(dir string) error {
		return native.Build(code, filepath.Join(dir, "port"), native.Options{})
	})
	binary, release := filepath.Join(nativeDir, "port"), filepath.Join(releaseDir, "port")
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_ARTIFACTS"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for name, source := range map[string]string{"port.c": filepath.Join(loweredDir, "port.c"), "port": release} {
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			mode := os.FileMode(0644)
			if name == "port" {
				mode = 0755
			}
			if err := os.WriteFile(filepath.Join(keep, name), data, mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	gapInput := ">const x:number=1;\n>const {x}=value;\n>export const x=1;\n>if(x)f();\n>#!/usr/bin/env node\\nf();\n"
	gapWant := "notyet\tvariable-types\nnotyet\tvariable-pattern\nnotyet\tvariable-modifiers\nnotyet\tIfStatement\nnotyet\tcomment-attachment\n"
	gapPath := filepath.Join(t.TempDir(), "statement-gaps.txt")
	if err := os.WriteFile(gapPath, []byte(gapInput), 0644); err != nil {
		t.Fatal(err)
	}
	script, _ := filepath.Abs("testdata/expressions.mjs")
	embedded, _ := filepath.Abs("testdata/embedded.mjs")
	bundles, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	backend := filepath.Join(loweredDir, "program.mjs")
	selected := statementShardSelection(t)
	t.Logf("setup including builds: %.3fs; union: %d cases in %d shards", time.Since(setup).Seconds(), strings.Count(want, "\n"), len(shards))
	for number, shard := range shards {
		if !selected(number) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			compare := func(name string, result run) {
				if err := statementDisagreement(name, result, shard.want, shard.labels); err != nil {
					t.Fatal(err)
				}
			}
			compare("Node", onNode(t, port, "--cases", shard.text, "80"))
			compare("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, "--cases", shard.text, "80"))
			compare("backend", onNode(t, backend, "--cases", shard.text, "80"))
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node gaps", onNode(t, port, "--cases", gapPath, "80")},
				{"native gaps", execute(t, nil, binary, "--cases", gapPath, "80")},
				{"backend gaps", onNode(t, backend, "--cases", gapPath, "80")},
			} {
				if err := statementDisagreement(side.name, side.result, gapWant, nil); err != nil {
					t.Fatal(err)
				}
			}
			if report := leaks(t, program, binary, "--cases", shard.text, "80"); report != "" {
				t.Fatal(report)
			}
			statementPrinterLibrary(t, "npm Prettier", execute(t, nil, "node", script, library, shard.specs), shard.specs, false)
			statementPrinterLibrary(t, "embedded Prettier", execute(t, nil, "node", embedded, bundles, shard.specs), shard.specs, true)
			compare("release", execute(t, nil, release, "--cases", shard.text, "80"))
			t.Logf("%d statement/program fragments byte-identical", len(shard.indices))
		})
	}
}
