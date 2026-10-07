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

func TestSlot02Batch5(t *testing.T) {
	// Not parallel: baseline and five compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch5"
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
	symbols := []string{"tailwind/collapse.DesignSystem.HasUtility", "tailwind/collapse.findRoots", "tailwind/collapse.PropertySort"}
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
	cohere, _ := filepath.Abs("../../../../cohere")
	config := smallFixture(t, map[string]any{"Inputs": inputs, "Paths": []string{filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/theme_test.go"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/framework_variants_test.go")}})
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	collapseExport, _ := filepath.Abs(base + "/testdata/collapse_export.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch5.go"): oracleSource, filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_slot02_batch5.go"): collapseExport}
	// Wrap only the public PropertySort dependency call site. The actual private
	// traversal remains unchanged; this records list identity and call count.
	designPath, _ := filepath.Abs("../../../../cohere/internal/lint/rules/tailwind/collapse/walk.go")
	design, err := os.ReadFile(designPath)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "sort, _ := propertySort(nodes)"
	if strings.Count(string(design), anchor) != 1 {
		t.Fatal("sort dependency anchor drift")
	}
	instrumented := filepath.Join(t.TempDir(), "walk.go")
	if err = os.WriteFile(instrumented, []byte(strings.Replace(string(design), anchor, "sort, _ := AdamicSlot02Batch5PropertySort(nodes)", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[designPath] = instrumented
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch5.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want             string
		Texts            []string
		ParsedCandidates int
		Stores, Sorts    []json.RawMessage
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
	t.Logf("%d captured inputs; %d distinct texts; %d parsed candidates; %d stores; %d sort trees", len(inputs), len(corpus.Texts), corpus.ParsedCandidates, len(corpus.Stores), len(corpus.Sorts))
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
		{"repository precedence", "design_system_has_utility.a", "if (system.repository.has(root))", "if (false)"},
		{"repository kind truth", "design_system_has_utility.a", "declared.get(kind) ?? false", "declared.has(kind)"},
		{"static presence", "design_system_has_utility.a", "system.frameworkStatic.has(root)", "system.frameworkStatic.get(root) ?? false"},
		{"functional truth", "design_system_has_utility.a", "system.frameworkFunctional.get(root) ?? false", "system.frameworkFunctional.has(root)"},
		{"exact root first", "find_roots.a", "if (exists(input))", "if (false)"},
		{"empty value stop", "find_roots.a", "if (value === '') { break; }", "if (value === '') { index = input.slice(0, index).lastIndexOf('-'); continue; }"},
		{"at fallback after dash stop", "find_roots.a", "input.charCodeAt(0) === 64", "input.charCodeAt(0) === 65"},
		{"at callback short circuit", "find_roots.a", "root === '@' && exists('@')", "exists('@') && root === '@'"},
		{"sort single delegation", "property_sort.a", "const result = compute(nodes);", "compute(nodes); const result = compute(nodes);"},
		{"sort count", "property_sort.a", "count: result.sort.count", "count: result.sort.count + 1"},
		{"sort order alias", "property_sort.a", "order: result.sort.order", "order: result.sort.order.slice()"},
		{"sort count value semantics", "property_sort.a", "return { order: result.sort.order, count: result.sort.count };", "return result.sort;"},
		{"sort node identity", "property_sort.a", "compute(nodes)", "compute(nodes.slice())"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/design_system_has_utility.a", base + "/find_roots.a", base + "/property_sort.a"}
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
