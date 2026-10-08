package arena

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConcreteIndices(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	lane := filepath.Join(root, "stage1/cohere/arena")
	run := func(args ...string) []byte {
		t.Helper()
		c := exec.Command(args[0], args[1:]...)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
		return out
	}
	entry := filepath.Join(lane, "testdata/main.a")
	want := []byte("0,1,2:0:0:true\n")
	node := run("node", "--no-warnings", "oracle/node.mjs", entry)
	if !bytes.Equal(node, want) {
		t.Fatalf("Node: %s", node)
	}
	binary := filepath.Join(t.TempDir(), "indices")
	run("go", "run", "./cmd/adamic", "build", entry, "-o", binary)
	if got := run(binary); !bytes.Equal(got, want) {
		t.Fatalf("native: %s", got)
	}
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join(lane, "arena_index.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, indexName := range []string{"FunctionIndex", "PatternIndex", "ScopeIndex", "ReactiveIndex"} {
		bad := strings.Replace(string(data), "new "+indexName+"(arena.length)", "new "+indexName+"(arena.length + 1)", 1)
		if bad == string(data) {
			t.Fatal("mutant anchor moved")
		}
		os.MkdirAll(filepath.Join(dir, "testdata"), 0755)
		if err := os.WriteFile(filepath.Join(dir, "arena_index.a"), []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		driver, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		mutant := filepath.Join(dir, "testdata/main.a")
		if err := os.WriteFile(mutant, driver, 0600); err != nil {
			t.Fatal(err)
		}
		badBinary := filepath.Join(dir, "mutant")
		run("go", "run", "./cmd/adamic", "build", mutant, "-o", badBinary)
		for _, args := range [][]string{{"node", "--no-warnings", filepath.Join(root, "oracle/node.mjs"), mutant}, {badBinary}} {
			c := exec.Command(args[0], args[1:]...)
			c.Dir = root
			out, err := c.CombinedOutput()
			if err == nil || !bytes.Contains(out, []byte(indexName+" out of range")) {
				t.Fatalf("mutant wasn't caught: %v %s", err, out)
			}
		}
	}

	for name, code := range map[string]string{
		"private":              "new FunctionIndex(2);",
		"scope-private":        "new ScopeIndex(0);",
		"reactive-private":     "new ReactiveIndex(0);",
		"scope-reactive-brand": "const scopes: ScopeIndex[] = []; const nodes: ReactiveIndex[] = []; ScopeIndex.read(scopes,ReactiveIndex.push(nodes));",
		"pattern-private":      "new PatternIndex(0);",
		"pattern-brand":        "const patterns: PatternIndex[] = []; const functions: FunctionIndex[] = []; PatternIndex.read(patterns,FunctionIndex.push(functions));",
		"brand":                "const functions: FunctionIndex[] = []; const blocks: BlockIndex[] = []; FunctionIndex.read(functions, BlockIndex.push(blocks));",
	} {
		p := filepath.Join(lane, "testdata", name+"-generated.a")
		text := "import { FunctionIndex, BlockIndex, PatternIndex, ScopeIndex, ReactiveIndex } from '../arena_index.a';\n" + code
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(p)
		c := exec.Command("go", "run", "./cmd/adamic", "build", p, "-o", filepath.Join(dir, name))
		c.Dir = root
		out, err := c.CombinedOutput()
		if err == nil {
			t.Fatalf("%s rejection failed: %s", name, out)
		}
		t.Logf("%s rejected: %s", name, out)
	}
}
