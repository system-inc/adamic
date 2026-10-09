package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// These source programs exercise the delivered placeholder machinery through both backends.
func TestParserConstructionSourceFactoryUnset(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "factory-unset")
}

func TestParserConstructionSourceOwnKeyOrder(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "own-key-order")
}

func TestParserConstructionSourceNodeArrayKeys(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "node-array-keys")
}

func TestParserConstructionSourceNodeArrayJson(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "node-array-json")
}

func TestParserConstructionSourceNodeArrayLength(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "node-array-length")
}

func TestParserConstructionSourceNodeArrayUnset(t *testing.T) {
	t.Parallel()
	assertParserConstructionSource(t, "node-array-unset")
}

func assertParserConstructionSource(t *testing.T, name string) {
	t.Helper()
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
	t.Log("source Node, native sanitizers, leak check and JavaScript agree; delivered placeholder machinery")
}

func TestParserConstructionUnsetUse(t *testing.T) {
	t.Parallel()
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

// The delivered placeholder machinery retains null independently of factory unset.
func TestParserConstructionNullPlaceholderDelivered(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/rulings/null-placeholder-pending.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "true\ntrue\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node %+v", truth)
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
}

func TestParserProgramRegionLifetime(t *testing.T) {
	t.Parallel()
	t.Skip("awaits compiler/program-region-lowering: parser trees in the Program region")
}
