package ruleproof

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuleProof(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	replacements := map[string]string{}
	replacements[filepath.Join(root, "stage1/cohere/lint/adamic_structure_owned_test.go")] = filepath.Join(root, "stage1/cohere/lint/rules/structure-next-require-api-parameter-name/testdata/selected_test.go")
	for _, change := range []struct{ file, from, to string }{
		{"shared_test.go", "\tvar names []string\n\tfor name := range packages", "\tpackages[\"structure\"] = append(packages[\"structure\"], \"TestNoRuleCrashesOnAbsentOptionalNodes\")\n\tvar names []string\n\tfor name := range packages"},
		{"lint_test.go", `const nativeCanaryRule = "react/jsx-no-comment-textnodes"`, `const nativeCanaryRule = "structure/next-require-api-parameter-name"`},
	} {
		original := filepath.Join(root, "stage1/cohere/lint", change.file)
		data, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), change.from) != 1 {
			t.Fatal("overlay anchor drift", change.file)
		}
		path := filepath.Join(scratch, change.file)
		if err = os.WriteFile(path, []byte(strings.Replace(string(data), change.from, change.to, 1)), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[original] = path
	}
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(scratch, "overlay.json")
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-overlay="+overlay, "./stage1/cohere/lint", "-run", "^TestOwnedRuleGoAgreement$|^TestMutants$/^next-search-parameters-repair-wrong$", "-count=1", "-v", "-timeout=30m")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rule proof: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "sanitized native canary") {
		t.Fatalf("missing sanitized native mutant observation:\n%s", output)
	}
	t.Log(string(output))
}
