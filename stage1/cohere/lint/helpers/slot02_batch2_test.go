package helpers

import (
	"bufio"
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

func TestSlot02Batch2(t *testing.T) {
	// Not parallel: baseline and four semantic mutants deliberately run the same
	// exhaustive Unicode probe; sequential execution bounds sanitizer memory.
	base := "slot02/batch2"
	file, err := os.Open(base + "/testdata/sources.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	zipped, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer zipped.Close()
	var inputs []map[string]string
	observed := map[string]bool{}
	scan := bufio.NewScanner(zipped)
	scan.Buffer(make([]byte, 65536), 16<<20)
	for scan.Scan() {
		var row map[string]string
		if err = json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, row)
		observed[row["rule"]] = true
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	symbols := []string{"tailwind.*ClassLiteralReader.ClassLiteralsIn", "tailwind.SplitClasses", "react.IsNamespacedMember"}
	for _, symbol := range symbols {
		count := 0
		for _, row := range ledger.Remaining {
			for _, h := range row.Helpers {
				if strings.HasSuffix(h, "/"+symbol) {
					count++
					if !observed[row.Rule] {
						t.Fatalf("missing captured consumer %s", row.Rule)
					}
				}
			}
		}
		t.Logf("%s: %d consumers with actual runtime capture", symbol, count)
	}
	config := smallFixture(t, inputs)
	cohere, _ := filepath.Abs("../../../../cohere")
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	exportSource, _ := filepath.Abs(base + "/testdata/tailwind_export.go")
	overlay := smallFixture(t, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "adamic_slot02_batch2.go"): oracleSource, filepath.Join(cohere, "internal/lint/rules/tailwind/adamic_slot02_batch2.go"): exportSource}})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch2.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                   string
		Sources                int
		Nodes, Texts, Readings []json.RawMessage
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err = json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "Want")
	corpusPath := smallFixture(t, input)
	want := []byte(corpus.Want)
	t.Logf("%d captured inputs; %d distinct parsed sources; %d node projections; %d split texts; %d literal projections; 1114112 separator candidates", len(inputs), corpus.Sources, len(corpus.Nodes), len(corpus.Texts), len(corpus.Readings))
	entry, _ := filepath.Abs(base + "/main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(entry string) []byte {
		t.Helper()
		gotNode := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpusPath)
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			t.Fatal(err)
		}
		nativePath := filepath.Join(t.TempDir(), "program")
		if err = native.Build(native.C(ir), nativePath, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		gotNative := run(t, "", nativePath, corpusPath)
		jsPath := filepath.Join(t.TempDir(), "program.mjs")
		if err = os.WriteFile(jsPath, []byte(javascript.JavaScript(ir)), 0644); err != nil {
			t.Fatal(err)
		}
		gotJS := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, jsPath, corpusPath)
		compare(t, gotNative, gotNode)
		compare(t, gotJS, gotNode)
		return gotNative
	}
	compare(t, check(entry), want)
	t.Log("Actual Go, Node source, emitted JavaScript and ASan/UBSan native agree")
	mutants := []struct{ name, file, anchor, replacement string }{
		{"namespace guard", "namespaced_member.a", "base.text !== 'React'", "false"},
		{"drop partial fields", "split_classes.a", "!field.includes('${')", "true"},
		{"Unicode whitespace", "split_classes.a", "code === 160", "false"},
		{"shared slice", "class_literals_in.a", "classValuesIn(node).literals", "classValuesIn(node).literals.slice()"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			files := []string{"options_json.ts", "slot02_ast.a", base + "/main.a", base + "/class_literals_in.a", base + "/split_classes.a", base + "/namespaced_member.a"}
			for _, name := range files {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == base+"/"+mutant.file {
					if strings.Count(string(data), mutant.anchor) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = []byte(strings.Replace(string(data), mutant.anchor, mutant.replacement, 1))
				}
				targetName := name
				if name == "options_json.ts" {
					targetName = "options_json.a"
				}
				if name == base+"/main.a" {
					data = []byte(strings.Replace(string(data), "../../options_json.ts", "../../options_json.a", 1))
				}
				target := filepath.Join(directory, targetName)
				if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(target, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := check(filepath.Join(directory, base, "main.a"))
			if bytes.Equal(got, want) {
				t.Fatal("compiled semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i, line := range a {
				if i < len(b) && line != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, line, b[i])
					break
				}
			}
		})
	}
}
