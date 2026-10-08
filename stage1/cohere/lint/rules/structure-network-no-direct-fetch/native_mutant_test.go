package ruleproof

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Use the shared mutant gate's sanitized native canary for this rule through an
// oracle-only overlay; the shared harness and its ordinary canary stay intact.
func TestNativeMutant(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "stage1/cohere/lint/lint_test.go")
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	from := `const nativeCanaryRule = "react/jsx-no-comment-textnodes"`
	if strings.Count(string(data), from) != 1 {
		t.Fatal("native canary anchor drift")
	}
	scratch := t.TempDir()
	changed := filepath.Join(scratch, "lint_test.go")
	if err = os.WriteFile(changed, []byte(strings.Replace(string(data), from, `const nativeCanaryRule = "structure/network-no-direct-fetch"`, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(scratch, "overlay.json")
	encoded, err := json.Marshal(map[string]any{"Replace": map[string]string{original: changed}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-overlay="+overlay, "./stage1/cohere/lint", "-run", "^TestMutants$/^bare-fetch-omitted$", "-count=1", "-v", "-timeout=30m")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native mutant gate: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "sanitized native") {
		t.Fatalf("missing native mutant observation:\n%s", output)
	}
	t.Log(string(output))
}
