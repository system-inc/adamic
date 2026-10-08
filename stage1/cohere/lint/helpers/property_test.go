package helpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const propertySymbol = "github.com/system-inc/cohere/internal/lint/ecmascript/property.Name"

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
	inputFile, err := os.Open("testdata/slot02_inputs.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer inputFile.Close()
	zipped, err := gzip.NewReader(inputFile)
	if err != nil {
		t.Fatal(err)
	}
	defer zipped.Close()
	var inputs []struct{ Rule, File, Source string }
	if err = json.NewDecoder(zipped).Decode(&inputs); err != nil {
		t.Fatal(err)
	}
	selected := []any{}
	observed := map[string]int{}
	for _, input := range inputs {
		if consumers[input.Rule] {
			selected = append(selected, input)
			observed[input.Rule]++
		}
	}
	for name := range consumers {
		if observed[name] == 0 {
			t.Fatalf("no actual captured input for %s", name)
		}
	}
	config := smallFixture(t, map[string]any{"Paths": paths, "Inputs": selected})
	cohere := filepath.Join(root, "cohere")
	source, _ := filepath.Abs("testdata/slot02_oracle.go")
	overlay := smallFixture(t, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "adamic_slot02.go"): source}})
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", binary, filepath.Join(cohere, "adamic_slot02.go"))
	mode := "attribute"
	if symbol == propertySymbol {
		mode = "property"
	}
	data = run(t, "", binary, config, mode)
	var corpus struct {
		Want                     string
		Files, Sources, Captured int
		Nodes, Queries           []json.RawMessage
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %d consumers, %d test files, %d actual captured inputs, %d distinct parsed sources with controls, %d nodes, %d node queries", symbol, len(consumers), corpus.Files, corpus.Captured, corpus.Sources, len(corpus.Nodes), len(corpus.Queries))
	path := filepath.Join(t.TempDir(), "corpus.json")
	var runtimeCorpus map[string]json.RawMessage
	if err = json.Unmarshal(data, &runtimeCorpus); err != nil {
		t.Fatal(err)
	}
	delete(runtimeCorpus, "Want")
	data, err = json.Marshal(runtimeCorpus)
	if err != nil {
		t.Fatal(err)
	}
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
	if err = os.WriteFile(binary+".mjs", []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestPropertyName(t *testing.T) {
	slot02Check(t, propertySymbol, "property_name.a", "inner.kind === 'Identifier' || inner.kind === 'PrivateIdentifier'", "false")
}
func slot02Check(t *testing.T, symbol, file, anchor, replacement string) {
	corpus, want := slot02Corpus(t, symbol)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("property/main.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus), want)
	binary := slot02Build(t, entry)
	compare(t, run(t, "", binary, corpus), want)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, binary+".mjs", corpus), want)
	t.Log("Go, Node source, emitted JavaScript and sanitized native agree")
	directory := t.TempDir()
	for _, name := range []string{"options_json.ts", "slot02_ast.a", "property_name.a", "property/main.a"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == file {
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), anchor, replacement, 1))
		}
		targetName := name
		if name == "options_json.ts" {
			targetName = "options_json.a"
		}
		if name == "property/main.a" {
			data = []byte(strings.Replace(string(data), "../options_json.ts", "../options_json.a", 1))
		}
		target := filepath.Join(directory, targetName)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutantEntry := filepath.Join(directory, "property/main.a")
	nodeMutant := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, mutantEntry, corpus)
	mutantBinary := slot02Build(t, mutantEntry)
	got := run(t, "", mutantBinary, corpus)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, mutantBinary+".mjs", corpus), got)
	if !bytes.Equal(got, nodeMutant) {
		t.Fatal("mutant Node and native disagree before comparison with Go")
	}
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := range a {
		if i < len(b) && a[i] != b[i] {
			t.Logf("compiled semantic mutant caught at line %d: got %s, Go %s", i+1, a[i], b[i])
			break
		}
	}
}
