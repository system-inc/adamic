package high_level_intermediate_representation

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func command(t *testing.T, dir string, env []string, args ...string) []byte {
	t.Helper()
	c := exec.Command(args[0], args[1:]...)
	c.Dir = dir
	c.Env = append(os.Environ(), env...)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args, err, out)
	}
	return out
}
func TestStraightLineOracle(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/high_level_intermediate_representation")
	temp := t.TempDir()
	output := filepath.Join(temp, "go.txt")
	target := filepath.Join(root, "cohere/internal/lint/ecmascript/high_level_intermediate_representation/stage1_hir_oracle_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{target: filepath.Join(lane, "testdata/oracle_test.go")}})
	overlayPath := filepath.Join(temp, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	corpus := filepath.Join(lane, "testdata/corpus.txt")
	command(t, filepath.Join(root, "cohere"), []string{"GOWORK=" + filepath.Join(root, "cohere/go.work"), "HIR_CORPUS=" + corpus, "HIR_OUTPUT=" + output}, "go", "test", "-overlay", overlayPath, "-tags=lintoracle", "-count=1", "-run=^TestStage1HIRDump$", "./internal/lint/ecmascript/high_level_intermediate_representation")
	want, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	node := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(lane, "main.ts"), corpus)
	if !bytes.Equal(node, want) {
		t.Fatalf("Node differs from Go\n%s", firstDifference(node, want))
	}
	binary := filepath.Join(temp, "hir")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(lane, "main.ts"), "-o", binary)
	native := command(t, root, nil, binary, corpus)
	if !bytes.Equal(native, want) {
		t.Fatalf("native differs from Go\n%s", firstDifference(native, want))
	}
	t.Logf("%d functions match Go on Node and natively", strings.Count(string(want), "hir-v1\n"))
	// Mutate one semantic operation, leave the dump alone, and require a successful but wrong result.
	original, err := os.ReadFile(filepath.Join(lane, "lower.ts"))
	if err != nil {
		t.Fatal(err)
	}
	needle := "fn.instructions.push(new Instruction(fn.instructions.length, fn.returns, value, statement.pos, statement.end));"
	if strings.Count(string(original), needle) != 1 {
		t.Fatal("mutant anchor moved")
	}
	mutant := strings.Replace(string(original), needle, "fn.instructions.push(new Instruction(fn.instructions.length, fn.returns, { kind: 'Primitive', literal: 'nil' }, statement.pos, statement.end));", 1)
	// Copy only the lane into a temporary sibling, retaining relative dependency imports.
	mutantLane, err := os.MkdirTemp(filepath.Dir(lane), "hir-mutant-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mutantLane)
	for _, file := range []string{"core.ts", "graph.ts", "lower.ts", "dump.ts", "main.ts"} {
		data, err := os.ReadFile(filepath.Join(lane, file))
		if err != nil {
			t.Fatal(err)
		}
		if file == "lower.ts" {
			data = []byte(mutant)
		}
		if err := os.WriteFile(filepath.Join(mutantLane, file), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	badNode := command(t, root, nil, "node", "--no-warnings", "oracle/node.mjs", filepath.Join(mutantLane, "main.ts"), corpus)
	if bytes.Equal(badNode, want) {
		t.Fatal("return-store mutant survived on Node")
	}
	mutantBinary := filepath.Join(temp, "mutant")
	command(t, root, nil, "go", "run", "./cmd/adamic", "build", filepath.Join(mutantLane, "main.ts"), "-o", mutantBinary)
	badNative := command(t, root, nil, mutantBinary, corpus)
	if bytes.Equal(badNative, want) {
		t.Fatal("return-store mutant survived natively")
	}
	t.Log("return-store-to-nil mutant caught on Node and natively")
}
func firstDifference(got, want []byte) string {
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return "got: " + a[i] + "\nwant: " + b[i]
		}
	}
	return "different lengths"
}
