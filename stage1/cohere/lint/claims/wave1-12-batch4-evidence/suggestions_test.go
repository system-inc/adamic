package suggestiongap

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: the independent Go build intentionally bounds compiler memory.
func TestSuggestionRangesExceedSharedContract(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	owned := filepath.Join(root, "stage1/cohere/lint/claims/wave1-12-batch4-evidence")
	cohere := filepath.Join(root, "cohere")
	scratch := t.TempDir()
	oracleFile := filepath.Join(cohere, "adamic_wave12_suggestion_oracle.go")
	selection := filepath.Join(cohere, "adamic_wave12_suggestion_selection.go")
	replacements := map[string]string{oracleFile: filepath.Join(root, "stage1/cohere/lint/testdata/oracle.go"), selection: filepath.Join(owned, "selection.go.txt")}
	if os.Getenv("ADAMIC_SUGGESTION_CONTROL") == "1" {
		source, err := os.ReadFile(replacements[oracleFile])
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Replace(string(source), "if len(d.Suggestions) > 0 {", "if false && len(d.Suggestions) > 0 {", 1)
		side := filepath.Join(scratch, "control.go.txt")
		if err := os.WriteFile(side, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[oracleFile] = side
	}
	data, _ := json.Marshal(map[string]any{"Replace": replacements})
	overlay := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(scratch, "oracle")
	command := exec.Command("go", "build", "-overlay="+overlay, "-o", binary, oracleFile, selection)
	command.Dir = cohere
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	for _, subject := range []struct{ file, name string }{{"optional.ts.txt", "@typescript-eslint/no-non-null-asserted-optional-chain"}, {"constraint.ts.txt", "@typescript-eslint/no-unnecessary-type-constraint"}} {
		path := filepath.Join(scratch, subject.file+".manifest")
		if err := os.WriteFile(path, []byte(filepath.Join(owned, subject.file)+"\t"+subject.name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(binary, "--manifest", path).CombinedOutput()
		if err == nil || !strings.Contains(string(out), "panic: unexpected suggestion shape") {
			t.Errorf("%s did not expose the range mismatch: %v %s", subject.name, err, out)
			continue
		}
		t.Logf("%s: compiled shared oracle rejects real suggestion: %s", subject.name, out)
	}
}
