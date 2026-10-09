package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIntMapIndexRadixAgreesWithNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "parseint_map.a")
	if err := os.WriteFile(path, []byte("console.log(['10','10','10'].map(Number.parseInt).join(','));"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	if oracle.exitCode != 0 || string(oracle.stdout) != "10,NaN,2\n" || len(oracle.stderr) != 0 {
		t.Fatalf("Node: %#v", oracle)
	}
	backend := onJavaScriptBackend(t, program)
	if difference := disagreement(oracle, backend); difference != "" {
		t.Fatalf("JavaScript backend %s: %#v", difference, backend)
	}
	native, binary := natively(t, program)
	if difference := disagreement(oracle, native); difference != "" {
		t.Fatalf("native %s: %#v", difference, native)
	}
	if difference := disagreement(oracle, released(t, program)); difference != "" {
		t.Fatal("release: " + difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("Node, JavaScript backend, sanitized native and release native: 10,NaN,2")
}
