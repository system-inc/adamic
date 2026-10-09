package regex

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
	"strconv"
	"strings"
	"testing"
)

func run(t *testing.T, cwd, command string, args ...string) []byte {
	t.Helper()
	c := exec.Command(command, args...)
	c.Dir = cwd
	var out, err bytes.Buffer
	c.Stdout = &out
	c.Stderr = &err
	if e := c.Run(); e != nil || err.Len() != 0 {
		t.Fatalf("%s: %v stderr=%s", command, e, err.String())
	}
	return out.Bytes()
}
func backends(t *testing.T, root string) [][]byte {
	t.Helper()
	entry := filepath.Join(root, "testdata/runner.a")
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	temp := t.TempDir()
	bin := filepath.Join(temp, "runner")
	if e := native.Build(native.C(ir), bin, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	js := filepath.Join(temp, "runner.js")
	if e := os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
		t.Fatal(e)
	}
	repository, _ := filepath.Abs("../../../..")
	return [][]byte{run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry), run(t, root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), js), run(t, root, bin)}
}
func firstRow(a, b []byte) int {
	left := strings.Split(string(a), "\n")
	right := strings.Split(string(b), "\n")
	for i, s := range left {
		if i >= len(right) || s != right[i] {
			n, _ := strconv.Atoi(strings.Split(s, ":")[0])
			return n
		}
	}
	return 0
}
// Not parallel: native.Build writes the shared adamic/runtime and adamic/units caches and native.runtimeBuilds map.
func TestFixedPatterns(t *testing.T) {
	// Not parallel: recompiles the same sizable generated corpus for a semantic mutant.
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	expected := run(t, cohere, "go", "run", filepath.Join(root, "testdata/oracle.go"), filepath.Join(root, "table.json"), filepath.Join(root, "testdata/corpus.json"))
	var rows []struct {
		ID      string  `json:"id"`
		Pattern *string `json:"go_pattern"`
	}
	data, _ := os.ReadFile("table.json")
	if e := json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	ids := []string{}
	for _, r := range rows {
		if r.Pattern != nil {
			ids = append(ids, r.ID)
		}
	}
	for backend, actual := range backends(t, root) {
		if !bytes.Equal(expected, actual) {
			t.Fatalf("backend %d differs: %s", backend, ids[firstRow(expected, actual)])
		}
	}
	t.Logf("%d fixed patterns, %d match records, %d identical bytes on Go, Node, emitted JS and sanitized native", len(ids), bytes.Count(expected, []byte("\n")), len(expected))
	temp := t.TempDir()
	if e := os.MkdirAll(filepath.Join(temp, "testdata"), 0755); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"patterns.a", "testdata/runner.a", "testdata/corpus.a"} {
		content, e := os.ReadFile(name)
		if e != nil {
			t.Fatal(e)
		}
		if name == "patterns.a" {
			content = bytes.Replace(content, []byte("/PaginationInput("), []byte("/PaginationInputBROKEN("), 1)
		}
		if e := os.WriteFile(filepath.Join(temp, name), content, 0644); e != nil {
			t.Fatal(e)
		}
	}
	for backend, actual := range backends(t, temp) {
		if bytes.Equal(expected, actual) {
			t.Fatalf("translation mutant escaped backend %d", backend)
		}
		t.Logf("translation mutant caught by byte comparison on backend %d: %s", backend, ids[firstRow(expected, actual)])
	}
}
func TestDynamicPatternGap(t *testing.T) {
	t.Parallel()
	entry, _ := filepath.Abs("testdata/dynamic_gap.a")
	repository, _ := filepath.Abs("../../../..")
	out := run(t, ".", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), entry, "TODO")
	if string(out) != "true\n" {
		t.Fatalf("Node %q", out)
	}
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	_, e = lower.Lower(context.Background(), p)
	if e == nil || !strings.Contains(e.Error(), "RegExp with a nonconstant pattern") {
		t.Fatalf("expected named native gap, got %v", e)
	}
	t.Log(e)
}

func TestOptionDialectGap(t *testing.T) {
	t.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	goAnswer := run(t, cohere, "go", "run", filepath.Join(root, "testdata/option_dialects.go"))
	nodeAnswer := run(t, root, "node", filepath.Join(root, "testdata/option_dialects.mjs"))
	expectedGo := "\\s\tfalse\na$\tfalse\n(?i)todo\ttrue\n\\p{Greek}\ttrue\na\\z\ttrue\n(?P<word>a)\ttrue\n"
	expectedNode := "\\s\ttrue\na$\tfalse\n(?i)todo\tSyntaxError\n\\p{Greek}\tSyntaxError\na\\z\tSyntaxError\n(?P<word>a)\tSyntaxError\n"
	if string(goAnswer) != expectedGo || string(nodeAnswer) != expectedNode {
		t.Fatalf("dialect observations changed: Go %q Node %q", goAnswer, nodeAnswer)
	}
	t.Logf("raw new RegExp(pattern, u) differs from accepted Go options on five of six named controls; Go=%q Node=%q", goAnswer, nodeAnswer)
}

func TestInventoryMatchesPinnedSource(t *testing.T) {
	t.Parallel()
	root, _ := filepath.Abs(".")
	cohere, _ := filepath.Abs("../../../../cohere")
	actual := run(t, cohere, "go", "run", filepath.Join(root, "testdata/inventory.go"), filepath.Join(cohere, "internal/lint/rules"))
	expected, err := os.ReadFile("sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, actual) {
		t.Fatal("regexp census drifted from pinned cohere AST; regenerate and review the table")
	}
}
