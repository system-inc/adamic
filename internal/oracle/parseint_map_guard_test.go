package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIntMapIndexRadixAgreesWithNode(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "parseint_map.a")
	if writeError := os.WriteFile(
		path,
		[]byte("console.log(['10','10','10'].map(Number.parseInt).join(','));"),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	oracle := onNode(test, path)
	if oracle.exitCode != 0 || string(oracle.stdout) != "10,NaN,2\n" || len(oracle.stderr) != 0 {
		test.Fatalf("Node: %#v", oracle)
	}
	backend := onJavaScriptBackend(test, program)
	if difference := disagreement(oracle, backend); difference != "" {
		test.Fatalf("JavaScript backend %s: %#v", difference, backend)
	}
	native, binary := natively(test, program)
	if difference := disagreement(oracle, native); difference != "" {
		test.Fatalf("native %s: %#v", difference, native)
	}
	if difference := disagreement(oracle, released(test, program)); difference != "" {
		test.Fatal("release: " + difference)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("Node, JavaScript backend, sanitized native and release native: 10,NaN,2")
}
