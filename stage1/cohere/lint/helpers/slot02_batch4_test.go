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

func TestSlot02Batch4(t *testing.T) {
	// Not parallel: baseline and five compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch4"
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
	symbols := []string{"tailwind/collapse.isValidThemePrefix", "tailwind/collapse.namespaceForVariantRoot", "tailwind/collapse.convertUnderscoresToWhitespace"}
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
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch4.go"): oracleSource, filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_slot02_batch4.go"): collapseExport}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch4.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want           string
		Texts          []string
		ParsedVariants int
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
	t.Logf("%d captured inputs; %d distinct texts; %d parsed variants", len(inputs), len(corpus.Texts), corpus.ParsedVariants)
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
		{"empty prefix", "valid_theme_prefix.a", "if (prefix.length === 0) { return false; }", "if (prefix.length === 0) { return true; }"},
		{"prefix z boundary", "valid_theme_prefix.a", "code > 122", "code >= 122"},
		{"namespace root position", "namespace_for_variant_root.a", "startsWith('@')", "includes('@')"},
		{"underscore skip flag", "convert_underscores_to_whitespace.a", "code === 95 && !skip", "code === 95"},
		{"escaped underscore", "convert_underscores_to_whitespace.a", "input.slice(start, index) + '_'", "input.slice(start, index) + ' '"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/valid_theme_prefix.a", base + "/namespace_for_variant_root.a", base + "/convert_underscores_to_whitespace.a"}
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
