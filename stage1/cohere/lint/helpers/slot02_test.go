package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const attributeSymbol = "github.com/system-inc/cohere/internal/lint/ecmascript/jsx.AttributeName"

func slot02Corpus(t *testing.T, symbol string) (string, []byte) {
	t.Helper()
	root, _ := filepath.Abs("../../../..")
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	consumers := map[string]bool{}
	for _, row := range ledger.Remaining {
		for _, h := range row.Helpers {
			if h == symbol {
				consumers[row.Rule] = true
			}
		}
	}
	data, err = os.ReadFile("../inventory/inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	var inventory struct {
		Rules []struct {
			Name  string
			Tests struct{ Files []string }
		}
	}
	if err = json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	paths := []string{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		if len(r.Tests.Files) == 0 {
			t.Fatalf("no test file for %s", r.Name)
		}
		for _, p := range r.Tests.Files {
			if !seen[p] {
				paths = append(paths, filepath.Join(root, p))
				seen[p] = true
			}
		}
	}
	config := smallFixture(t, paths)
	cohere := filepath.Join(root, "cohere")
	source, _ := filepath.Abs("testdata/slot02_oracle.go")
	overlay := smallFixture(t, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "adamic_slot02.go"): source}})
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", binary, filepath.Join(cohere, "adamic_slot02.go"))
	data = run(t, "", binary, config)
	var corpus struct {
		Want           string
		Files, Sources int
		Nodes, Queries []json.RawMessage
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %d consumers, %d test files, %d unique string-literal sources, %d nodes, %d queries", symbol, len(consumers), corpus.Files, corpus.Sources, len(corpus.Nodes), len(corpus.Queries))
	path := filepath.Join(t.TempDir(), "corpus.json")
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path, []byte(corpus.Want)
}
func slot02Build(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "helper")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestSlot02AttributeName(t *testing.T) {
	corpus, want := slot02Corpus(t, attributeSymbol)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("slot02/main.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus), want)
	compare(t, run(t, "", slot02Build(t, entry), corpus), want)
	t.Log("Go, Node source and sanitized native agree")
	directory := t.TempDir()
	for _, name := range []string{"options_json.ts", "slot02_ast.a", "jsx_attribute_name.a", "slot02/main.a"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "jsx_attribute_name.a" {
			anchor := "name.kind !== 'Identifier'"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), anchor, "false", 1))
		}
		target := filepath.Join(directory, name)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	got := run(t, "", slot02Build(t, filepath.Join(directory, "slot02/main.a")), corpus)
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := range a {
		if i < len(b) && a[i] != b[i] {
			t.Logf("compiled namespaced-name mutant caught at line %d: got %s, Go %s", i+1, a[i], b[i])
			break
		}
	}
}
