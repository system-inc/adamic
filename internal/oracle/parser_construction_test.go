package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// These are source programs, not hand-built IR or stub passes. The base remains
// a rehearsal: publishing them does not certify the placeholder dependency.
func TestParserConstructionSource(t *testing.T) {
	for _, name := range []string{"factory-unset", "own-key-order", "node-array-keys", "node-array-json", "node-array-length", "node-array-unset"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
				if d := disagreement(truth, got); d != "" {
					t.Fatalf("%s: %s", backend, d)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("source Node, native sanitizers, leak check and JavaScript agree; pending codex/placeholder-nonnull integration")
		})
	}
}

func TestParserConstructionUnsetUse(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings/factory-use.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 70 || string(truth.stdout) != "true\n" || !strings.Contains(string(truth.stderr), "TypeError") {
		t.Fatalf("Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	message := "unset use failed: field 'text', factory 'createBaseNode', use 'node.text' at " + path + ":5:13; expected string"
	want := run{stdout: []byte("true\n"), stderr: []byte("adamic: panic: " + message + "\n"), exitCode: 70}
	actual, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if d := disagreement(want, got); d != "" {
			t.Fatalf("%s: %s", backend, d)
		}
	}
	t.Log("undefined presence observation is accepted; declared string use stops naming field, factory and use")
}

// The rehearsal retains readiness but lacks null-placeholder observation. Keep
// the dependency visible rather than manufacturing undefined from a null! slot.
func TestParserConstructionNullPlaceholderPending(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings/null-placeholder-pending.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "null\nundefined\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node %+v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "observable null placeholder awaiting codex/placeholder-nonnull") {
		t.Fatalf("pending boundary: %v", err)
	}
	t.Log("pending codex/placeholder-nonnull; no backend acceptance claimed")
}
