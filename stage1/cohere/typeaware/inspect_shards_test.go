package typeaware

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Prepare products once before parallel check units. ADAMIC_TEST_SHARD=i/n
// selects zero-based shards; unset runs every refusal and mutant check.
const testInspectRequestRefusalsShards = 6

// Not parallel: refusal controls share checker setup and build resources.
func TestInspectRequestRefusals(test *testing.T) {
	started := time.Now()
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := test.TempDir()
	if path := os.Getenv("ADAMIC_SIX_REQUEST_ARTIFACTS"); path != "" {
		directory, absolutePathError = filepath.Abs(path)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	h := &harness{t: test, repository: repository, directory: directory, sixBuilds: true, setupStarted: started}
	traceGroup(test)
	stage0 := h.stage0()
	normal := h.archive(false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts")
	binary := h.build(stage0, "request", entry, normal, false)
	configuration := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	probe := h.write("probe.ts", "-1;\n")
	var shards []sixShard
	expected := []string{
		"request-valid",
		"released-name",
		"wrong-kind/refused",
		"wrong-kind/mutant",
		"unknown-question/refused",
		"unknown-question/mutant",
	}
	add := func(name string, run func(*harness)) {
		shards = append(shards, sixShard{name: name, ids: []string{name}, run: run})
	}
	add("request-valid", func(harness *harness) {
		harness.must(
			"request-valid",
			exec.Command(binary, configuration, probe, "0", "2", "PrefixUnaryExpression", "raw-type", "2"),
		)
	})
	releasedEntry := h.write(
		"released-name.ts",
		"import {panic,programArguments,tsgoProgram,tsgoInspect,tsgoRelease} from 'adamic'; const a=programArguments(); const config=a[0]??panic('config');const file=a[1]??panic('file');const p=tsgoProgram(config,[file]); console.log(tsgoInspect(p,file,0,2,'PrefixUnaryExpression','raw-shape')); tsgoRelease(p); console.log(tsgoInspect(p,file,0,2,'PrefixUnaryExpression','name\\n1'));\n",
	)
	released := h.build(stage0, "released-name", releasedEntry, normal, false)
	add("released-name", func(harness *harness) {
		stale := harness.run("released-name-run", exec.Command(released, configuration, probe))
		if code, ok := stale.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			!bytes.Contains(stale.stderr, []byte("invalid or released checker handle")) {
			harness.t.Fatalf("released name escaped: %v %s", stale.err, stale.stderr)
		}
		harness.t.Log("released type-name query: panic 70, invalid or released checker handle")
	})
	for _, change := range []struct{ name, kind, question, message string }{
		{"wrong-kind", "Identifier", "raw-type", "no exact Identifier node"},
		{"unknown-question", "PrefixUnaryExpression", "unknown", "unsupported checker question"},
	} {
		args := []string{configuration, probe, "0", "2", change.kind, change.question, "2"}
		add(change.name+"/refused", func(harness *harness) {
			got := harness.run(change.name+"-refused", exec.Command(binary, args...))
			if code, ok := got.err.(*exec.ExitError); !ok ||
				code.ExitCode() != 70 ||
				!bytes.Contains(got.stderr, []byte(change.message)) {
				harness.t.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
			}
		})
		archive := checkerMutantArchive(test, change.name)
		mutant := h.build(stage0, change.name, entry, archive, false)
		add(change.name+"/mutant", func(harness *harness) {
			harness.must(change.name+"-mutant-run", exec.Command(mutant, args...))
			harness.t.Logf("%s guard mutant exits 0", change.name)
		})
	}
	sixRunShards(test, h, expected, shards, os.Getenv("ADAMIC_TEST_SHARD"), testInspectRequestRefusalsShards)
}
