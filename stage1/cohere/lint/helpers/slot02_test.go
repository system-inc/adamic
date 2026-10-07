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

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const propertySymbol = "github.com/system-inc/cohere/internal/lint/ecmascript/property.Name"
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
	return binary
}
func TestSlot02AttributeName(t *testing.T) {
	slot02Check(t, attributeSymbol, "jsx_attribute_name.a", "name.kind !== 'Identifier'", "false")
}
func TestSlot02PropertyName(t *testing.T) {
	slot02Check(t, propertySymbol, "property_name.a", "inner.kind === 'Identifier' || inner.kind === 'PrivateIdentifier'", "false")
}
func slot02Check(t *testing.T, symbol, file, anchor, replacement string) {
	corpus, want := slot02Corpus(t, symbol)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("slot02/main.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus), want)
	compare(t, run(t, "", slot02Build(t, entry), corpus), want)
	t.Log("Go, Node source and sanitized native agree")
	directory := t.TempDir()
	for _, name := range []string{"options_json.ts", "slot02_ast.a", "jsx_attribute_name.a", "property_name.a", "slot02/main.a"} {
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
		if name == "slot02/main.a" {
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
	mutantEntry := filepath.Join(directory, "slot02/main.a")
	nodeMutant := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, mutantEntry, corpus)
	got := run(t, "", slot02Build(t, mutantEntry), corpus)
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

func TestSlot02ReaderFor(t *testing.T) {
	root, _ := filepath.Abs("../../../../cohere")
	source, _ := filepath.Abs("testdata/slot02_reader_oracle.go")
	wrapper, _ := filepath.Abs("testdata/slot02_reader_go.go")
	overlay := smallFixture(t, map[string]any{"Replace": map[string]string{filepath.Join(root, "adamic_slot02_reader.go"): source, filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot02.go"): wrapper}})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(root, "adamic_slot02_reader.go"))
	data := run(t, "", oracle)
	var corpus struct{ Want string }
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "Want")
	path := smallFixture(t, input)
	want := []byte(corpus.Want)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("slot02/reader.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", slot02Build(t, entry), path), want)
	t.Logf("%d cache/binding observations match actual Go on Node and sanitized native", len(strings.Split(strings.TrimSpace(corpus.Want), "\n")))
	dir := t.TempDir()
	for _, name := range []string{"options_json.ts", "tailwind_reader_for.a", "slot02/reader.a"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "tailwind_reader_for.a" {
			anchor := "'tailwind.classValues:' + key"
			if strings.Count(string(data), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			data = []byte(strings.Replace(string(data), anchor, "key", 1))
		}
		targetName := name
		if name == "options_json.ts" {
			targetName = "options_json.a"
		}
		if name == "slot02/reader.a" {
			data = []byte(strings.Replace(string(data), "../options_json.ts", "../options_json.a", 1))
		}
		target := filepath.Join(dir, targetName)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutantEntry := filepath.Join(dir, "slot02/reader.a")
	gotNode := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, mutantEntry, path)
	got := run(t, "", slot02Build(t, mutantEntry), path)
	if !bytes.Equal(got, gotNode) {
		t.Fatal("mutant Node/native disagree")
	}
	if bytes.Equal(got, want) {
		t.Fatal("compiled cache-key mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := range a {
		if i < len(b) && a[i] != b[i] {
			t.Logf("compiled cache-key mutant caught at line %d: got %s, Go %s", i+1, a[i], b[i])
			break
		}
	}
}
