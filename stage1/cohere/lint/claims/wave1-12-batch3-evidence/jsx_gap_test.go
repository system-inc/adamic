package jsxgap

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: building three execution artifacts deliberately bounds compiler memory.
func TestClaimedRulesNeedJSX(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	owned := filepath.Join(root, "stage1/cohere/lint/claims/wave1-12-batch3-evidence")
	probe := filepath.Join(owned, "parser-probe.a")
	if os.Getenv("ADAMIC_NEXT_PROBE_MUTANT") == "1" {
		source, err := os.ReadFile(probe)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Replace(string(source), "parser.file();", "console.log('mutant skipped parsing');", 1)
		text = strings.Replace(text, "../../../../typescript/parser/parser.ts", filepath.Join(root, "stage1/typescript/parser/parser.ts"), 1)
		probe = filepath.Join(scratch, "mutant.a")
		if err := os.WriteFile(probe, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{probe})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(scratch, "parser")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(scratch, "parser.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	virtual := filepath.Join(cohere, "adamic_wave12_next_probe.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "go-probe.go.txt")}})
	overlayPath := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(scratch, "go-probe")
	command := exec.Command("go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	command.Dir = cohere
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Go probe: %v %s", err, out)
	}
	var witnesses []string
	for _, slug := range []string{"next-no-before-interactive-script-outside-document", "next-no-css-tags"} {
		witnesses = append(witnesses, filepath.Join(root, "stage1/cohere/lint/claims/wave1-12-batch3-evidence", slug, "witness.ts.txt"))
	}
	out, err := exec.Command(oracle, witnesses...).CombinedOutput()
	if err != nil {
		t.Fatalf("Go witnesses: %v %s", err, out)
	}
	t.Logf("unmodified Go rules: %s", out)
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, witness := range witnesses {
		var expectedOutput, expectedError []byte
		for _, side := range []string{"Node", "emitted JavaScript", "sanitized native"} {
			var command *exec.Cmd
			switch side {
			case "Node":
				command = exec.Command("node", "--disable-warning=ExperimentalWarning", runner, probe, witness)
			case "emitted JavaScript":
				command = exec.Command("node", "--disable-warning=ExperimentalWarning", runner, emitted, witness)
			default:
				command = exec.Command(binary, witness)
			}
			var output, stderr bytes.Buffer
			command.Stdout = &output
			command.Stderr = &stderr
			err := command.Run()
			out := stderr.Bytes()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || output.Len() != 0 {
				t.Errorf("%s expected parser refusal exit 70 and empty stdout: %v stdout=%q stderr=%q", side, err, output.Bytes(), out)
				continue
			}
			if expectedError == nil {
				expectedOutput = append([]byte{}, output.Bytes()...)
				expectedError = append([]byte{}, out...)
			} else if !bytes.Equal(expectedOutput, output.Bytes()) || !bytes.Equal(expectedError, out) {
				t.Fatalf("%s parser refusal bytes differ", side)
			}
			if err == nil || !strings.Contains(string(out), "panic:") {
				t.Fatalf("%s unexpectedly accepted JSX: %v %s", side, err, out)
			}
			if strings.Contains(string(out), "ERROR: AddressSanitizer") || strings.Contains(string(out), "runtime error:") {
				t.Fatalf("sanitizer failure is not a parser refusal: %s", out)
			}
			first := strings.SplitN(string(out), "\n", 2)[0]
			t.Logf("%s %s: %v %s", filepath.Base(filepath.Dir(witness)), side, err, first)
		}
	}
}
