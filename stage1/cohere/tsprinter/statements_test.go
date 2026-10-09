package tsprinter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func statementCorpusBuild(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	root, _ := filepath.Abs(repository)
	files := printerCorpusFiles(t)
	// This count is fixed by the upstream commit, never by the live case total.
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	upstreamFiles, repositoryFiles := 0, 0
	for _, file := range files {
		path, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(upstream, path)
		if err != nil {
			t.Fatal(err)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			upstreamFiles++
		} else {
			repositoryFiles++
		}
	}
	if upstreamFiles != 77 {
		t.Fatalf("TypeScript 050880ce src/compiler: %d .ts files, want pinned 77", upstreamFiles)
	}
	if repositoryFiles == 0 {
		t.Fatal("repository statement corpus is empty")
	}
	oracleDir := printerGateGoOracle(t)
	flags := []string{"Go oracle sha256=" + statementFileHash(t, oracleDir+"/oracle"), "width=80"}
	for _, file := range files {
		flags = append(flags, "input="+file+" sha256="+statementFileHash(t, file))
	}
	product := statementProduct(t, buildcache.Inputs{
		Name:  "tsprinter-statements-oracle-output-v1",
		Files: []string{"stage1/cohere/tsprinter/statements_test.go", "stage1/cohere/tsprinter/statements_shards_test.go"},
		Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}, func(dir string) error {
		// Requests are scratch inputs. Publish only the oracle's four output files.
		request, err := json.Marshal(map[string]any{"Files": files, "Directory": dir})
		if err != nil {
			return err
		}
		if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
			return err
		}
		command := statementCommand(t, oracleDir+"/oracle", "-test.v", "-test.count=1", "-test.run=^TestAdamicStatementCorpus$", "-test.timeout=3h")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
		output, err := combinedOutput(command)
		if err != nil {
			return fmt.Errorf("Go statement corpus: %w\n%s", err, output)
		}
		t.Log(string(output))
		return nil
	})
	directory = product
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
// unset runs all shards. The mixed live repository/pinned TypeScript corpus uses
// a fixed 16 shards and SHA-256(relative path, file/mode-local index) modulo 16.
// Build products are prepared once and shared by leaves.
func statementPrepareCommon(t *testing.T) statementSharedProducts {
	t.Helper()
	setup := time.Now()
	cpu := statementCPU()
	t.Cleanup(func() { t.Logf("shared setup CPU including builds: %.3fs", statementCPU()-cpu) })
	cases, want, specs := statementCorpus(t)
	shards := statementShards(t, cases, want, specs)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	port, _ := filepath.Abs("statementsMain.ts")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	compilerHash := statementFileHash(t, executable)
	loweredDir := printerGateLowered(t, "statements")
	binary, release := printerGateNative(t, "statements", true)+"/port", printerGateNative(t, "statements", false)+"/port"
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
	libraryHash := statementDirectoryHash(t, library)
	elapsed := time.Since(setup).Seconds()
	t.Logf("shared setup: %.3fs, %d cases in %d shards", elapsed, strings.Count(want, "\n"), len(shards))
	return statementSharedProducts{Cases: cases, Want: want, Specs: specs, Port: port, CompilerHash: compilerHash, LoweredDir: loweredDir, Binary: binary, Release: release, Library: library, LibraryHash: libraryHash, Elapsed: elapsed}
}

func statementAgainstGoAndPrettierShard(t *testing.T, shardNumber int) {
	t.Helper()
	if !statementShardSelection(t)(shardNumber) {
		return
	}
	products := statementReadyShared(t)
	setup := time.Now()
	cpu := statementCPU()
	t.Cleanup(func() { t.Logf("shard CPU: %.3fs", statementCPU()-cpu) })
	cases, want, specs := products.Cases, products.Want, products.Specs
	shards := statementShards(t, cases, want, specs)
	port, compilerHash := products.Port, products.CompilerHash
	binary, release, library := products.Binary, products.Release, products.Library
	loweredDir := products.LoweredDir

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
	selected := func(number int) bool { return number == shardNumber }
	libraryHash := products.LibraryHash
	oracleAnswers := make([][2][]byte, len(shards))
	for number, shard := range shards {
		if !selected(number) {
			continue
		}
		for side, oracle := range []struct {
			name, script, library string
			files                 []string
			flags                 []string
		}{
			{"npm", script, library, []string{"stage1/cohere/tsprinter/testdata/expressions.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "stage1/cohere/tsprinter/testdata/prettier-differences.json"}, []string{"npm bytes sha256=" + libraryHash}},
			{"embedded", embedded, bundles, []string{"stage1/cohere/tsprinter/testdata/embedded.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "cohere/internal/format/prettier/bundles"}, nil},
		} {
			directory := statementLibraryOracle(t, number, side, shard, compilerHash, libraryHash, oracle.name, oracle.script, oracle.library, oracle.files, oracle.flags)
			answers, err := os.ReadFile(filepath.Join(directory, "answers.txt"))
			if err != nil {
				t.Fatal(err)
			}
			oracleAnswers[number][side] = answers
		}
	}
	t.Logf("shard-local input preparation: %.3fs; union: %d cases in %d shards", time.Since(setup).Seconds(), strings.Count(want, "\n"), len(shards))
	caseStart := time.Now()
	caseContext, cancelCases := context.WithTimeout(context.Background(), 60*time.Second)
	statementCaseContexts.Store(t, caseContext)
	t.Cleanup(func() {
		cancelCases()
		statementCaseContexts.Delete(t)
		t.Logf("case time after shared setup: %.3fs", time.Since(caseStart).Seconds())
	})
	for number, shard := range shards {
		if !selected(number) {
			continue
		}
		func() {
			compare := func(name string, result run) {
				if err := statementDisagreement(name, result, shard.want, shard.labels); err != nil {
					t.Fatal(err)
				}
			}
			compare("Node", statementOnNode(t, port, "--cases", shard.text, "80"))
			compare("native", statementExecute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary, "--cases", shard.text, "80"))
			compare("backend", statementOnNode(t, backend, "--cases", shard.text, "80"))
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node gaps", statementOnNode(t, port, "--cases", gapPath, "80")},
				{"native gaps", statementExecute(t, nil, binary, "--cases", gapPath, "80")},
				{"backend gaps", statementOnNode(t, backend, "--cases", gapPath, "80")},
			} {
				if err := statementDisagreement(side.name, side.result, gapWant, nil); err != nil {
					t.Fatal(err)
				}
			}
			if report := statementLeaks(t, binary, release, "--cases", shard.text, "80"); report != "" {
				t.Fatal(report)
			}
			statementPrinterLibrary(t, "npm Prettier", run{stdout: oracleAnswers[number][0]}, shard.specs, false)
			statementPrinterLibrary(t, "embedded Prettier", run{stdout: oracleAnswers[number][1]}, shard.specs, true)
			compare("release", statementExecute(t, nil, release, "--cases", shard.text, "80"))
			t.Logf("%d statement/program fragments byte-identical", len(shard.indices))
		}()
	}
}

func statementLibraryOracle(t *testing.T, number, side int, shard statementShard, compilerHash, libraryHash, name, script, library string, files, oracleFlags []string) string {
	t.Helper()
	return printerGateOnce(t, fmt.Sprintf("statement-library-%d-%d", number, side), func() string {
		flags := append(oracleFlags, "builder sha256="+compilerHash, "cases sha256="+statementFileHash(t, shard.specs), "library="+library, "script="+script, "width=80", "NODE_OPTIONS="+os.Getenv("NODE_OPTIONS"), "NODE_PATH="+os.Getenv("NODE_PATH"))
		return statementProduct(t, buildcache.Inputs{
			Name:  fmt.Sprintf("tsprinter-statements-%s-oracle-shard-%03d-v1", name, number),
			Files: files, Flags: flags,
			Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
		}, func(dir string) error {
			result := statementExecute(t, nil, "node", script, library, shard.specs)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				return fmt.Errorf("%s oracle exit %d: %s", name, result.exitCode, result.stderr)
			}
			return os.WriteFile(filepath.Join(dir, "answers.txt"), result.stdout, 0644)
		})
	})
}

func statementCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := printerGateOnce(t, "statement-corpus", func() string { cases, _, _ := statementCorpusBuild(t); return filepath.Dir(cases) })
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}
