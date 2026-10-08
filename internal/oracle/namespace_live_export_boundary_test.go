package oracle

import (
	"path/filepath"
	"testing"
)

func TestNamespaceLiveExportBoundary(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "stage3/namespace-live-export/live.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "false:false\ntrue:true\nfalse:false\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node observation: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	observed, binary := nativelyUncached(t, program)
	if difference := disagreement(truth, observed); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
