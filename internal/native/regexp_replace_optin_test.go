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

func TestRegExpReplacementScoutNotLinked(t *testing.T) {
	library, err := RuntimeLibraryForSource("", "int main(void) { return 0; }", Options{Sanitize: true})
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("nm", library).CombinedOutput()
	if err != nil {
		t.Fatalf("nm: %v: %s", err, output)
	}
	if bytes.Contains(output, []byte("adamic_regex_replace_callback")) {
		t.Fatal("replacement callback linked without a callback program")
	}
}

func TestRegExpReplacementScoutOptInMutant(t *testing.T) {
	source, err := os.ReadFile("library.go")
	if err != nil {
		t.Fatal(err)
	}
	old := `strings.Contains(source, "#define "+feature+" 1\n")`
	if strings.Count(string(source), old) != 1 {
		t.Fatal("callback opt-in site moved")
	}
	directory := t.TempDir()
	replacement := filepath.Join(directory, "library.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(source), old, `feature == "ADAMIC_REGEXP_REPLACE_CALLBACK" || `+old, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	original, _ := filepath.Abs("library.go")
	data, err := json.Marshal(map[string]map[string]string{"Replace": {original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", overlay, ".", "-run", "^TestRegExpReplacementScoutNotLinked$", "-count=1", "-v")
	output, err := command.CombinedOutput()
	if err == nil || bytes.Contains(output, []byte("build failed")) || !bytes.Contains(output, []byte("replacement callback linked without a callback program")) {
		t.Fatalf("always-linked mutant did not reach the symbol assertion: %v\n%s", err, output)
	}
	t.Log("always-linked mutant caught by the callback symbol assertion")
}
