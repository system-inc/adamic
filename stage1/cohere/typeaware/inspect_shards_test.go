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
func TestInspectRequestRefusals(t *testing.T) {
	started := time.Now()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SIX_REQUEST_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory, sixBuilds: true, setupStarted: started}
	traceGroup(t)
	stage0 := h.stage0()
	normal := h.archive(false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts")
	binary := h.build(stage0, "request", entry, normal, false)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	probe := h.write("probe.ts", "-1;\n")
	var shards []sixShard
	expected := []string{"request-valid", "released-name", "wrong-kind/refused", "wrong-kind/mutant", "unknown-question/refused", "unknown-question/mutant"}
	add := func(name string, run func(*harness)) {
		shards = append(shards, sixShard{name: name, ids: []string{name}, run: run})
	}
	add("request-valid", func(h *harness) {
		h.must("request-valid", exec.Command(binary, config, probe, "0", "2", "PrefixUnaryExpression", "raw-type", "2"))
	})
	// A testdata program, so its native is keyed and built ahead (TestProduct_released_name_native), where a scratch
	// copy was linked inside this check every run (#5gf3w8q).
	releasedEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released_name.ts")
	released := h.build(stage0, "released-name", releasedEntry, normal, false)
	add("released-name", func(h *harness) {
		stale := h.run("released-name-run", exec.Command(released, config, probe))
		if code, ok := stale.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(stale.stderr, []byte("invalid or released checker handle")) {
			h.t.Fatalf("released name escaped: %v %s", stale.err, stale.stderr)
		}
		h.t.Log("released type-name query: panic 70, invalid or released checker handle")
	})
	for _, change := range []struct{ name, kind, question, message string }{
		{"wrong-kind", "Identifier", "raw-type", "no exact Identifier node"},
		{"unknown-question", "PrefixUnaryExpression", "unknown", "unsupported checker question"},
	} {
		args := []string{config, probe, "0", "2", change.kind, change.question, "2"}
		add(change.name+"/refused", func(h *harness) {
			got := h.run(change.name+"-refused", exec.Command(binary, args...))
			if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(change.message)) {
				h.t.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
			}
		})
		// The guard mutant's native is a product over its checker mutant archive (TestProduct_inspect_<name>_native),
		// where h.build linked it here, keyed by nothing, since its archive isn't the unchanged checker.
		mutant := h.mutantNative("inspect-"+change.name, entry, checkerMutantArchive(t, change.name), false)
		add(change.name+"/mutant", func(h *harness) {
			h.must(change.name+"-mutant-run", exec.Command(mutant, args...))
			h.t.Logf("%s guard mutant exits 0", change.name)
		})
	}
	sixRunShards(t, h, expected, shards, os.Getenv("ADAMIC_TEST_SHARD"), testInspectRequestRefusalsShards)
}
