package unit3

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// This checks the Go adapter only. It is not the census or an Adamic certificate.
func TestGoCheckpointAdapter(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation/passes/unit-3")
	target := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation")
	replacements := map[string]string{}
	for _, name := range []string{"checkpoint_adapter_test.go", "checkpoint_adapter_checks_test.go"} {
		replacements[filepath.Join(target, "stage1_unit3_"+name)] = filepath.Join(lane, "testdata", name)
	}
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-tags=lintoracle", "-overlay", overlay, "-count=1", "-timeout=3h", "-run=^TestUnit3Adapter", "./internal/lint/ecmascript/high_level_intermediate_representation")
	command.Dir = filepath.Join(root, "cohere")
	command.Env = append(os.Environ(), "GOWORK="+filepath.Join(root, "cohere/go.work"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Go adapter: %v\n%s", err, output)
	}
	t.Logf("%s", output)
}
