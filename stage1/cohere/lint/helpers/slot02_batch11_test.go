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

func TestSlot02Batch11(t *testing.T) {
	// Not parallel: one consumer corpus and native sanitizer builds are shared by the baseline and mutants.
	base := "slot02/batch11"
	var inputs []map[string]string
	observed := map[string]bool{}
	for _, name := range []string{"slot02/batch6/testdata/sources.jsonl.gz", "slot02/batch10/testdata/sources.jsonl.gz"} {
		f, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		scan := bufio.NewScanner(z)
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
		z.Close()
		f.Close()
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
	for _, symbol := range []string{"tailwind/collapse.*stylesheetCollector.ingest", "control_flow_graph.*Builder[E].nestedFunction", "control_flow_graph.*Builder[E].variableDeclarationList"} {
		count := 0
		for _, row := range ledger.Remaining {
			for _, h := range row.Helpers {
				if strings.HasSuffix(h, "/"+symbol) {
					count++
					if !observed[row.Rule] {
						t.Fatalf("missing consumer %s", row.Rule)
					}
				}
			}
		}
		t.Logf("%s: %d consumers captured", symbol, count)
	}
	cohere, _ := filepath.Abs("../../../../cohere")
	config := smallFixture(t, map[string]any{"Inputs": inputs, "Paths": []string{filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/theme_test.go"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/framework_variants_test.go"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/utility_test.go")}})
	replacements := map[string]string{}
	for _, pair := range [][2]string{{"adamic_slot02_batch11.go", "oracle.go"}, {"internal/lint/rules/tailwind/collapse/adamic_slot02_batch11.go", "collapse_export.go"}, {"internal/lint/ecmascript/control_flow_graph/adamic_slot02_batch11.go", "cfg_export.go"}} {
		path, _ := filepath.Abs(base + "/testdata/" + pair[1])
		replacements[filepath.Join(cohere, pair[0])] = path
	}
	original := filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/design_system.go")
	data, err = os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	start := strings.Index(s, "func (collector *stylesheetCollector) ingest(")
	end := strings.Index(s[start:], "// resolveImport turns") + start
	section := s[start:end]
	for _, pair := range [][2]string{{"collector.resolveImport(node, path)", "Slot11Resolve(collector,node,path)"}, {"collector.loadFile(resolved)", "Slot11Load(collector,resolved)"}, {"collector.ingestThemeBlock(node, path)", "Slot11Theme(collector,node,path)"}, {"collector.ingestUtilityBlock(node, path)", "Slot11Utility(collector,node,path)"}, {"collector.ingestCustomVariant(node)", "Slot11Custom(collector,node)"}} {
		if strings.Count(section, pair[0]) != 1 {
			t.Fatal("dependency anchor drift")
		}
		section = strings.Replace(section, pair[0], pair[1], 1)
	}
	target := filepath.Join(t.TempDir(), "design_system.go")
	if err = os.WriteFile(target, []byte(s[:start]+section+s[end:]), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[original] = target
	for _, method := range []struct{ file, name, op, anchor, replacement string }{
		{"statements.go", "variableDeclarationList", "list", "b.variableDeclaration(decl)", "Slot11Declaration(b,decl)"},
		{"expressions.go", "nestedFunction", "nested", "b.memberHeader(node)", "Slot11Header(b,node)"},
	} {
		original = filepath.Join(cohere, "internal/lint/ecmascript/control_flow_graph/"+method.file)
		data, err = os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		s = string(data)
		start = strings.Index(s, "func (b *Builder[E]) "+method.name+"(")
		end = strings.Index(s[start+1:], "\nfunc ") + start + 1
		if start < 0 || end <= start {
			t.Fatal("method boundary")
		}
		section = s[start:end]
		node := "node"
		if method.op == "list" {
			node = "list"
		}
		section = strings.Replace(section, "{\n", "{\nslot11Capture(\""+method.op+"\",b,"+node+")\n", 1)
		if strings.Count(section, method.anchor) != 1 {
			t.Fatal("dependency anchor drift")
		}
		section = strings.Replace(section, method.anchor, method.replacement, 1)
		target = filepath.Join(t.TempDir(), method.file)
		if err = os.WriteFile(target, []byte(s[:start]+section+s[end:]), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[original] = target
	}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch11.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want     string
		CSS, CFG []json.RawMessage
		Calls    int
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
	t.Logf("%d captured inputs, %d CSS replays, %d CFG replays, %d actual CFG helper entries", len(inputs), len(corpus.CSS), len(corpus.CFG), corpus.Calls)
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
		{"import resolved path", "stylesheet_ingest.a", "deps.loadFile(resolved.value)", "deps.loadFile(path)"},
		{"resolve error propagation", "stylesheet_ingest.a", "if (resolved.error !== -1)", "if (false)"},
		{"theme error propagation", "stylesheet_ingest.a", "deps.theme(id, path);\n            if (error !== -1)", "deps.theme(id, path);\n            if (false)"},
		{"load error propagation", "stylesheet_ingest.a", "deps.loadFile(resolved.value);\n            if (error !== -1)", "deps.loadFile(resolved.value);\n            if (false)"},
		{"utility error propagation", "stylesheet_ingest.a", "deps.utility(id, path);\n            if (error !== -1)", "deps.utility(id, path);\n            if (false)"},
		{"utility dispatch", "stylesheet_ingest.a", "deps.utility(id, path)", "deps.theme(id, path)"},
		{"custom variant dispatch", "stylesheet_ingest.a", "deps.customVariant(id);", "deps.customVariant(0);"},
		{"skipped path", "stylesheet_ingest.a", "params: node.params, path", "params: node.params, path: ''"},
		{"unknown at-rule recursion", "stylesheet_ingest.a", "} else if (node.container)", "} else if (false)"},
		{"non-at-rule recursion", "stylesheet_ingest.a", "if (node.container) {\n                const", "if (false) {\n                const"},
		{"nil list guard", "variable_declaration_list.a", "if (node === -1) { return; }", "if (node === -1) { visit(builder,node); return; }"},
		{"nil declarations guard", "variable_declaration_list.a", "if (!hasDeclarations) { return; }", "if (!hasDeclarations) { visit(builder,node); return; }"},
		{"nested forwarding identity", "nested_function.a", "memberHeader(builder, node);", "memberHeader(builder, -1);"},
		{"nested builder identity", "nested_function.a", "memberHeader(builder, node);", "memberHeader(0, node);"},
		{"nested exactly once", "nested_function.a", "memberHeader(builder, node);", "memberHeader(builder, node); memberHeader(builder, node);"},
		{"declaration order", "variable_declaration_list.a", "for (const declaration of declarations)", "for (const declaration of declarations.slice().reverse())"},
		{"declaration identity", "variable_declaration_list.a", "visit(builder, declaration);", "visit(builder, -1);"},
		{"declaration builder identity", "variable_declaration_list.a", "visit(builder, declaration);", "visit(0, declaration);"},
		{"declaration exactly once", "variable_declaration_list.a", "visit(builder, declaration);", "visit(builder, declaration); visit(builder, declaration);"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/stylesheet_ingest.a", base + "/nested_function.a", base + "/variable_declaration_list.a"}
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
