package regex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShapeFixtures(t *testing.T) {
	// Not parallel: every fixture is compiled and compared, including the native transition control.
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("table.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ ID string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "shapes-gate")
	run(t, repository, "go", "build", "-o", binary, "stage1/cohere/lint/regex/testdata/shapes/gate.go")
	invoke := func(arguments ...string) ([]byte, error) {
		command := exec.Command(binary, arguments...)
		command.Dir = repository
		var output bytes.Buffer
		command.Stdout = &output
		command.Stderr = &output
		err := command.Run()
		return output.Bytes(), err
	}
	var arguments []string
	if os.Getenv("ADAMIC_REGEX_TRANSLATION_MUTANT") == "1" {
		arguments = append(arguments, "-mutant")
	}
	output, err := invoke(arguments...)
	if err != nil {
		t.Fatalf("per-shape fixture comparison failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), fmt.Sprintf("PASS fixtures=%d ", len(rows))) {
		t.Fatalf("fixture coverage count missing: %s", output)
	}
	t.Log(strings.TrimSpace(string(output)))
	t.Run("one-character-translation-mutant", func(t *testing.T) {
		output, err := invoke("-mutant")
		if err == nil || !strings.Contains(string(output), rows[0].ID+": source Node fixture mismatch") {
			t.Fatalf("mutant escaped or failed outside comparison: %v\n%s", err, output)
		}
		t.Logf("one-character mutant caught by fixture comparison: %s", rows[0].ID)
	})
	t.Run("forced-native-transition", func(t *testing.T) {
		forced, err := invoke("-force-native")
		if strings.Contains(string(output), fmt.Sprintf("awaits codex/regex-runtime-compiler=%d", len(rows))) {
			if err == nil || !strings.Contains(string(forced), "awaits codex/regex-runtime-compiler: forced native requirement caught refusal") {
				t.Fatalf("native refusal control failed: %v\n%s", err, forced)
			}
			t.Log("awaits codex/regex-runtime-compiler: forced native requirement caught named refusal")
		} else if err != nil {
			t.Fatalf("native is now required: %v\n%s", err, forced)
		}
	})
}
