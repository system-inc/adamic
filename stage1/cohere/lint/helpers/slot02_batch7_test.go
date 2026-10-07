package helpers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestSlot02Batch7(t *testing.T) {
	// Not parallel: baseline and compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch7"
	// Pin both Go's full simple-fold mapping and the generated production table.
	foldsFile := filepath.Join(t.TempDir(), "folds.json")
	generator, _ := filepath.Abs(base + "/testdata/folds.go")
	run(t, "", "go", "run", generator, foldsFile)
	fresh, err := os.ReadFile(foldsFile)
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(base + "/testdata/folds.json")
	if err != nil {
		t.Fatal(err)
	}
	compare(t, bytes.TrimSpace(fresh), bytes.TrimSpace(pinned))
	var folding struct {
		Version string
		Pairs   []int
	}
	if err = json.Unmarshal(fresh, &folding); err != nil {
		t.Fatal(err)
	}
	var table strings.Builder
	fmt.Fprintf(&table, "// BEGIN GENERATED SIMPLE FOLD TABLE\n// Go Unicode %s. Regenerate with testdata/regenerate_folds.py.\nconst foldPairs: readonly number[] = [\n", folding.Version)
	for i := 0; i < len(folding.Pairs); i += 2 {
		fmt.Fprintf(&table, "    %d, %d,\n", folding.Pairs[i], folding.Pairs[i+1])
	}
	table.WriteString("];\n")
	helper, err := os.ReadFile(base + "/match_ignoring_case.a")
	if err != nil {
		t.Fatal(err)
	}
	position := strings.Index(string(helper), "// BEGIN GENERATED SIMPLE FOLD TABLE")
	if position < 0 {
		t.Fatal("fold table marker drift")
	}
	compare(t, helper[position:], []byte(table.String()))
	t.Logf("all Unicode scalars verified: %s, %d noncanonical mappings", folding.Version, len(folding.Pairs)/2)
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
	symbols := []string{"imports.CallExpressionSource", "jsx.MatchIgnoringCase", "jsx.HasAttributeNamed"}
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
		expected := 4
		if symbol == "imports.CallExpressionSource" {
			expected = 5
		}
		if count != expected {
			t.Fatalf("consumer ledger drift: %s: %d", symbol, count)
		}
		t.Logf("%s: %d consumers with actual runtime capture", symbol, count)
	}
	config := smallFixture(t, inputs)
	cohere, _ := filepath.Abs("../../../../cohere")
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch7.go"): oracleSource}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch7.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                     string
		Sources                  int
		Nodes, Pairs, Attributes []json.RawMessage
		UnicodeVersion           string
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	var input map[string]json.RawMessage
	if err = json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	delete(input, "Want")
	path := smallFixture(t, input)
	want := []byte(corpus.Want)
	t.Logf("%d captured inputs; %d distinct parsed sources; %d nodes; %d Unicode/name pairs; %d attribute targets; Unicode %s", len(inputs), corpus.Sources, len(corpus.Nodes), len(corpus.Pairs), len(corpus.Attributes), corpus.UnicodeVersion)
	entry, _ := filepath.Abs(base + "/main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(t *testing.T, entry string) []byte {
		t.Helper()
		gotNode := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path)
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
		got := run(t, "", nativePath, path)
		jsPath := filepath.Join(t.TempDir(), "program.mjs")
		if err = os.WriteFile(jsPath, []byte(javascript.JavaScript(ir)), 0644); err != nil {
			t.Fatal(err)
		}
		gotJS := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, jsPath, path)
		compare(t, got, gotNode)
		compare(t, gotJS, gotNode)
		return got
	}
	compare(t, check(t, entry), want)
	t.Log("Actual Go, Node source, emitted JavaScript and sanitized native agree")
	mutants := []struct{ name, file, anchor, replacement string }{
		{"require argument count", "call_expression_source.a", "node.arguments.length === 1", "node.arguments.length >= 1"},
		{"deferred imports", "call_expression_source.a", "expression.text === 'defer'", "expression.text === 'source'"},
		{"template literal source", "call_expression_source.a", "first.kind !== 'NoSubstitutionTemplateLiteral'", "true"},
		{"empty source present", "call_expression_source.a", "found: true", "found: first.text !== ''"},
		{"Kelvin fold", "match_ignoring_case.a", "8490, 75,", "8490, 8490,"},
		{"long s fold", "match_ignoring_case.a", "383, 83,", "383, 383,"},
		{"supplementary rune fold", "match_ignoring_case.a", "66600, 66560,", "66600, 66600,"},
		{"unequal string length", "match_ignoring_case.a", "return left === candidate.length && right === wanted.length;", "return true;"},
		{"attribute matcher arguments", "has_attribute_named.a", "matches(name.name, wanted)", "matches(wanted, name.name)"},
		{"attribute order", "has_attribute_named.a", "for (const property of attributes.properties)", "for (const property of attributes.properties.slice().reverse())"},
		{"unnamed attributes", "has_attribute_named.a", "name.named && matches", "matches"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", "slot02_ast.a", "jsx_attribute_name.a", base + "/main.a", base + "/call_expression_source.a", base + "/has_attribute_named.a", base + "/match_ignoring_case.a"}
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
				target := filepath.Join(dir, targetName)
				if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(target, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := check(t, filepath.Join(dir, base, "main.a"))
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
