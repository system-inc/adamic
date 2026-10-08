package unit9

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A stopped-boundary regression, not a pass certificate.
func TestSharedScopeBoundary(t *testing.T) {
	t.Parallel()
	before, err := os.ReadFile("testdata/before.hir.txt")
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile("testdata/after.hir.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("Go loop flattening unexpectedly rewrote its graph")
	}
	result, err := os.ReadFile("testdata/flattened.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(result, []byte("{\"7\":true}\n")) {
		t.Fatalf("Go flattened identities: %s", result)
	}
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "--no-warnings", "oracle/node.mjs", "stage1/cohere/high_level_intermediate_representation/passes/unit-9/scope_replay.a")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("adamic: panic: unknown terminal Scope")) {
		t.Fatalf("Scope blocker changed; resume unit 9: err=%v output=%s", err, output)
	}
}
