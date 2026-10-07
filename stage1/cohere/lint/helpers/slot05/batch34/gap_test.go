package batch34

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDynamicCompilationBoundary(t *testing.T) {
	root, err := filepath.Abs("../../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	virtual := filepath.Join(root, "cohere/adamic_slot05_batch34.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go")}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(scratch, "overlay.json")
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	commands := []*exec.Cmd{
		exec.Command("go", "run", "-overlay="+overlay, virtual, "a"),
		exec.Command("node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "dynamic_pattern.a"), "a"),
	}
	commands[0].Dir = filepath.Join(root, "cohere")
	for _, cmd := range commands {
		output, err := cmd.CombinedOutput()
		if err != nil || string(output) != "true\n" {
			t.Fatalf("%v: %v %q", cmd.Args, err, output)
		}
		t.Logf("%s: true", cmd.Args[0])
	}
	program, err := load.Load([]string{filepath.Join(dir, "dynamic_pattern.a")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "RegExp with a nonconstant pattern") {
		t.Fatalf("expected the documented dynamic pattern gap, got %v", err)
	}
	t.Log(err)
}
