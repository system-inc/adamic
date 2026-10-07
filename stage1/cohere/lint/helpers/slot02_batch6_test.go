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

func TestSlot02Batch6(t *testing.T) {
	// Not parallel: baseline and compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch6"
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
	symbols := []string{"tailwind/collapse.*stylesheetCollector.loadFile", "tailwind/collapse.*stylesheetCollector.resolveImport", "tailwind/collapse.*stylesheetCollector.ingestCustomVariant"}
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
	config := smallFixture(t, map[string]any{"Inputs": inputs, "Paths": []string{filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/theme_test.go"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/framework_variants_test.go"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/utility_test.go")}})
	oracleSource, _ := filepath.Abs(base + "/testdata/oracle.go")
	collapseExport, _ := filepath.Abs(base + "/testdata/collapse_export.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch6.go"): oracleSource, filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_slot02_batch6.go"): collapseExport}
	original := filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/design_system.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(source), "\"os\"\n", "", 1)
	start := strings.Index(changed, "func (collector *stylesheetCollector) loadFile")
	end := strings.Index(changed[start:], "// ingest walks") + start
	section := changed[start:end]
	for _, pair := range [][2]string{{"filepath.Abs(path)", "AdamicSlot02Batch6Abs(path)"}, {"os.ReadFile(absolutePath)", "AdamicSlot02Batch6Read(absolutePath)"}, {"ParseCSS(string(content))", "AdamicSlot02Batch6Parse(string(content))"}, {"collector.ingest(nodes, absolutePath)", "AdamicSlot02Batch6Ingest(collector,nodes,absolutePath)"}} {
		if strings.Count(section, pair[0]) != 1 {
			t.Fatal("load dependency anchor drift")
		}
		section = strings.Replace(section, pair[0], pair[1], 1)
	}
	changed = changed[:start] + section + changed[end:]
	for _, function := range []string{"resolveImport", "ingestCustomVariant"} {
		start := strings.Index(changed, "func (collector *stylesheetCollector) "+function)
		tail := changed[start:]
		anchor := "parts := segment(strings.TrimSpace(node.Params), ' ')"
		if !strings.Contains(tail, anchor) {
			t.Fatal("segment dependency anchor drift")
		}
		changed = changed[:start] + strings.Replace(tail, anchor, "parts := AdamicSlot02Batch6Segment(strings.TrimSpace(node.Params), ' ')", 1)
	}
	destination := filepath.Join(t.TempDir(), "design_system.go")
	if err = os.WriteFile(destination, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[original] = destination
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch6.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                     string
		Variants, Imports, Loads []json.RawMessage
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
	t.Logf("%d captured inputs; %d variant params; %d import params; %d file loads", len(inputs), len(corpus.Variants), len(corpus.Imports), len(corpus.Loads))
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
		{"absolute error unwrap cause", "stylesheet_load_file.a", "deps.describe(absolute.error)}`, absolute.error", "deps.describe(absolute.error)}`, -1"},
		{"parse error cleanup", "stylesheet_load_file.a", "const error = deps.makeError(`parse ${resolved}: ${deps.describe(nodes.error)}`, nodes.error);\n        collector.visiting.delete(resolved);\n        return error;", "return deps.makeError(`parse ${resolved}: ${deps.describe(nodes.error)}`, nodes.error);"},
		{"empty specifier refused", "stylesheet_resolve_import.a", "if (parts.length === 0 || parts[0] === '')", "if (false)"},

		{"visiting is a stack", "stylesheet_load_file.a", "collector.visiting.delete(resolved);\n    return result;", "return result;"},
		{"ingest sees visiting", "stylesheet_load_file.a", "collector.visiting.set(resolved, true);", "collector.visiting.set(resolved, false);"},
		{"stylesheet append precedes ingest", "stylesheet_load_file.a", "collector.stylesheets.push(resolved);\n    const result = deps.ingest(nodes.value, resolved);", "const result = deps.ingest(nodes.value, resolved);\n    collector.stylesheets.push(resolved);"},
		{"read error unwrap cause", "stylesheet_load_file.a", "deps.describe(content.error)}`, content.error", "deps.describe(content.error)}`, -1"},
		{"read error cleanup", "stylesheet_load_file.a", "const error = deps.makeError(`read ${resolved}: ${deps.describe(content.error)}`, content.error);\n        collector.visiting.delete(resolved);\n        return error;", "return deps.makeError(`read ${resolved}: ${deps.describe(content.error)}`, content.error);"},
		{"cycle guard", "stylesheet_load_file.a", "collector.visiting.get(resolved) ?? false", "false"},
		{"strip all edge quotes", "stylesheet_resolve_import.a", "first.slice(start, end)", "first"},
		{"source modifier accepted", "stylesheet_resolve_import.a", "if (modifier.startsWith('source('))", "if (false)"},
		{"unsupported modifier refuses", "stylesheet_resolve_import.a", "if (modifier === '')", "if (true)"},
		{"resolver error unwrap cause", "stylesheet_resolve_import.a", "deps.describe(resolved.error)}`, resolved.error", "deps.describe(resolved.error)}`, -1"},
		{"resolver argument order", "stylesheet_resolve_import.a", "deps.resolve(specifier, deps.directory(path))", "deps.resolve(deps.directory(path), specifier)"},
		{"go whitespace includes NEL", "ingest_custom_variant.a", "value === 133", "value === 134"},
		{"go whitespace excludes BOM", "ingest_custom_variant.a", "value === 12288", "value === 12288 || value === 65279"},
		{"functional suffix", "ingest_custom_variant.a", "if (name.endsWith('-*'))", "if (false)"},
		{"empty root skipped", "ingest_custom_variant.a", "if (name === '') { return; }", ""},
	}

	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/stylesheet_load_file.a", base + "/stylesheet_resolve_import.a", base + "/ingest_custom_variant.a"}
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
