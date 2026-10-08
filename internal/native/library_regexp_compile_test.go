package native

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegExpRuntimeOwnershipOptInMutant(t *testing.T) {
	source, err := os.ReadFile("library_regexp_compile.go")
	if err != nil {
		t.Fatal(err)
	}
	old := `strings.Contains(source, "\n#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n")`
	if strings.Count(string(source), old) != 1 {
		t.Fatal("ownership opt-in site moved")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "library_regexp_compile.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(source), old, "true", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay := map[string]map[string]string{"Replace": {}}
	// Carry the caller's explicit test overlay, including an approved or proposed
	// Build hook and any separately documented integration-test field rename.
	if path := os.Getenv("ADAMIC_REGEX_TEST_OVERLAY"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &overlay); err != nil {
			t.Fatal(err)
		}
	}
	original, _ := filepath.Abs("library_regexp_compile.go")
	overlay["Replace"][original] = replacement
	data, _ := json.Marshal(overlay)
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", path, ".", "-run", "^TestRegExpRuntimeCompilerNotLinked$", "-count=1", "-v")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("ownership opt-in mutant survived")
	}
	if !bytes.Contains(output, []byte("compiler ownership linked")) || bytes.Contains(output, []byte("build failed")) {
		t.Fatalf("mutant did not reach symbol assertion: %v %s", err, output)
	}
	t.Log("caught always-on ownership variant at the symbol assertion")
}
