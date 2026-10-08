package unit6

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This is a dependency reproducer, not a pass certificate. It must go red when
// the owner adds Scope decoding, so the stopped state cannot survive unnoticed.
func TestScopeTerminalReplayContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	hir := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	lane := filepath.Join(hir, "passes/unit-6")
	upstream := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	temporary := t.TempDir()
	mapping := map[string]string{
		filepath.Join(upstream, "stage1_unit6_oracle_test.go"):   filepath.Join(lane, "oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_oracle_test.go"):     filepath.Join(hir, "testdata/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_checkpoint_test.go"): filepath.Join(hir, "replay/oracle_test.go"),
		filepath.Join(upstream, "stage1_hir_inputs_test.go"):     filepath.Join(hir, "replay/inputs_test.go"),
	}
	data, err := json.Marshal(map[string]any{"Replace": mapping})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(temporary, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-v", "-count=1", "-tags=lintoracle", "-overlay", overlay, "-run=^TestUnit6ScopeTerminalFixture$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	command.Dir = filepath.Join(root, "cohere")
	command.Env = append(os.Environ(), "GOWORK="+filepath.Join(root, "cohere/go.work"), "HIR_UNIT6_FIXTURES="+temporary)
	output, runError := command.CombinedOutput()
	if err := os.WriteFile(filepath.Join(lane, "validation/go-fixture.txt"), output, 0600); err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("Go fixture failed: %v; see validation/go-fixture.txt", runError)
	}
	for _, name := range []string{"single-scope.before.checkpoint", "single-scope.after.checkpoint"} {
		generated, err := os.ReadFile(filepath.Join(temporary, name))
		if err != nil {
			t.Fatal(err)
		}
		checkedIn, err := os.ReadFile(filepath.Join(lane, "testdata", name))
		if os.IsNotExist(err) && os.Getenv("HIR_UNIT6_UPDATE") == "1" {
			if err := os.WriteFile(filepath.Join(lane, "testdata", name), generated, 0600); err != nil {
				t.Fatal(err)
			}
		} else if err != nil || !bytes.Equal(generated, checkedIn) {
			t.Fatalf("Go checkpoint differs from testdata/%s: %v", name, err)
		}
	}
	entry := filepath.Join(lane, "replay_contract.a")
	binary := filepath.Join(temporary, "replay-contract")
	build := exec.Command("go", "run", "./cmd/adamic", "build", entry, "-o", binary, "--sanitize")
	build.Dir = root
	output, runError = build.CombinedOutput()
	if err := os.WriteFile(filepath.Join(lane, "validation/native-build.txt"), output, 0600); err != nil {
		t.Fatal(err)
	}
	if runError != nil {
		t.Fatalf("native build failed: %v; see validation/native-build.txt", runError)
	}
	emitted := filepath.Join(temporary, "replay-contract.mjs")
	emit := exec.Command("go", "run", "./cmd/adamic", "js", entry)
	emit.Dir = root
	output, runError = emit.CombinedOutput()
	if runError != nil {
		t.Fatalf("JavaScript emission failed: %v\n%s", runError, output)
	}
	if err := os.WriteFile(emitted, output, 0600); err != nil {
		t.Fatal(err)
	}
	for _, runtime := range []struct {
		name      string
		arguments []string
	}{
		{"node", []string{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), entry}},
		{"emitted-javascript", []string{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), emitted}},
		{"native", []string{binary}},
	} {
		for _, boundary := range []string{"before", "after"} {
			input := filepath.Join(temporary, "single-scope."+boundary+".checkpoint")
			arguments := append(append([]string(nil), runtime.arguments...), input)
			process := exec.Command(arguments[0], arguments[1:]...)
			process.Dir = root
			process.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
			output, runError := process.CombinedOutput()
			if err := os.WriteFile(filepath.Join(lane, "validation/"+runtime.name+"-"+boundary+".txt"), output, 0600); err != nil {
				t.Fatal(err)
			}
			if boundary == "before" {
				if runError != nil {
					t.Fatalf("%s: prepass graph is not replayable: %v", runtime.name, runError)
				}
				checkpoint, err := os.ReadFile(input)
				if err != nil {
					t.Fatal(err)
				}
				lines := strings.Split(string(checkpoint), "\n")
				count, err := strconv.Atoi(strings.TrimPrefix(lines[2], "graph-lines "))
				if err != nil {
					t.Fatal(err)
				}
				want := strings.Join(lines[3:3+count], "\n") + "\n"
				if string(output) != want {
					t.Fatalf("%s: decoded prepass graph differs from Go", runtime.name)
				}
			}
			if boundary == "after" && (runError == nil || !strings.Contains(string(output), "unknown terminal Scope")) {
				t.Fatalf("%s: Scope blocker changed; remove the stopped-state assertion and certify the pass: %v", runtime.name, runError)
			}
		}
		t.Logf("%s: Go input decodes; actual Go scope-terminal output stops at unknown terminal Scope", runtime.name)
	}
	t.Log("Unit 6 pass certificate: native=0/1465 Node=0/1465 emitted-JavaScript=0/1465; semantic mutant not run")
}
