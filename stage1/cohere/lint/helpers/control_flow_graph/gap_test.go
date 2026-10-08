package control_flow_graph

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This is a refusal test, not a helper parity or rule readiness certificate.
func TestAppendSuccessorOwnershipBoundary(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	temp := t.TempDir()
	run := func(dir, name string, args ...string) []byte {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil || stderr.Len() != 0 {
			t.Fatalf("%s: %v\n%s", name, err, stderr.String())
		}
		return out
	}
	adapter, _ := filepath.Abs("testdata/successor_exports.go.txt")
	oracle, _ := filepath.Abs("testdata/successor_oracle.go.txt")
	virtual := filepath.Join(cohere, "adamic_cfg_successor.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{
		virtual: oracle,
		filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph/adamic_successor.go"): adapter,
	}})
	overlayPath := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(cohere, "go", "run", "-overlay="+overlayPath, virtual)
	if string(want) != "1,7,true\n" {
		t.Fatalf("unexpected Go self-edge: %q", want)
	}
	for _, mutant := range []bool{false, true} {
		fixture := "append_successor.a.txt"
		if mutant {
			fixture = "append_successor_mutant.a.txt"
		}
		data, err := os.ReadFile(filepath.Join("testdata", fixture))
		if err != nil {
			t.Fatal(err)
		}
		entry := filepath.Join(temp, fixture+".a")
		if err := os.WriteFile(entry, data, 0644); err != nil {
			t.Fatal(err)
		}
		got := run(root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), entry)
		if mutant {
			if string(got) != "1,8,true\n" || bytes.Equal(got, want) {
				t.Fatalf("Node mutant not caught: %q", got)
			}
		} else if !bytes.Equal(got, want) {
			t.Fatalf("Go/Node disagreement: %q / %q", want, got)
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		_, err = lower.Lower(context.Background(), program)
		if err == nil || !strings.Contains(err.Error(), "adamic/cycle-capable") {
			t.Fatalf("expected current ownership refusal, got %v", err)
		}
		t.Logf("mutant=%t: %s", mutant, err)
	}
	t.Log("1 real private Go call agrees with source Node; 1 source Node mutant caught; emitted JavaScript and native both refused before emission")
}
