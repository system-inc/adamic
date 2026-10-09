package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// The original scout's base factory preserves one allocation through its checked
// Identifier view; reservation permits later writes without synthesizing own keys.
func TestParserAheadFactoryBoundariesComplete(t *testing.T) {
	t.Parallel()
	assertParserAheadFactoryBoundary(t, "complete", "ready\n")
}

func TestParserAheadFactoryBoundariesEscaped(t *testing.T) {
	t.Parallel()
	assertParserAheadFactoryBoundary(t, "escaped", "undefined\n")
}

func assertParserAheadFactoryBoundary(t *testing.T, name, stdout string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
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
	t.Log("base factory source matches both backends; delivered placeholder machinery")
}

func TestParserAheadFactoryCompleteControl(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories/already-complete.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatal(d)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	changed := 0
	for i, s := range program.Strings {
		if s == "ready" {
			program.Strings[i] = "wrong"
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("payload mutant changed %d constants", changed)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "wrong\n" {
			t.Fatalf("mutant must execute normally: %+v", got)
		}
		if disagreement(truth, got) == "" {
			t.Fatal("factory payload mutant survived")
		}
	}
	t.Log("factory payload mutant caught in native and JavaScript by source Node stdout")
}

func TestParserAheadNodeArraySourceFields(t *testing.T) {
	t.Parallel()
	assertParserAheadNodeArraySource(t, "fields", "-1|7|true|1|1|1\ntrue\n", "a cast the runtime can't check")
}

func TestParserAheadNodeArraySourcePresence(t *testing.T) {
	t.Parallel()
	assertParserAheadNodeArraySource(t, "presence", "false|true\ntrue|true\ntrue|ready\n", "an unproven relation")
}

func assertParserAheadNodeArraySource(t *testing.T, name, stdout, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/node-array", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != stdout || len(truth.stderr) != 0 {
		t.Fatalf("source Node %+v", truth)
	}
	program, err := lowered(t, path)
	if name == "fields" && err == nil {
		actual, binary := nativelyUncached(t, program)
		for _, got := range []run{actual, onJavaScriptBackend(t, program)} {
			if d := disagreement(truth, got); d != "" {
				t.Fatal(d)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
		t.Log("all four NodeArray metadata fields match source Node; delivered placeholder machinery")
		return
	}
	if err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("pinned optional-only array boundary changed: %v", err)
	}
	t.Log("optional-only array relation remains refused; this boundary test does not certify acceptance")
}
