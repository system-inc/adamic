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

func TestSlot02Batch3(t *testing.T) {
	// Not parallel: baseline and five compiling semantic mutants share one large
	// consumer corpus; sequential execution bounds native sanitizer memory.
	base := "slot02/batch3"
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
	symbols := []string{"react.IsEs5ComponentCall", "react.isCreateClassName", "tailwind/collapse.isMathFunctionName"}
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
	reactExport, _ := filepath.Abs(base + "/testdata/react_export.go")
	collapseExport, _ := filepath.Abs(base + "/testdata/collapse_export.go")
	replacements := map[string]string{filepath.Join(cohere, "adamic_slot02_batch3.go"): oracleSource, filepath.Join(cohere, "internal/lint/ecmascript/react/adamic_slot02_batch3.go"): reactExport, filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_slot02_batch3.go"): collapseExport}
	// Instrument only dependency call sites, preserving actual Go helper bodies
	// and actual private predicate answers. Worktree files are untouched.
	componentPath := filepath.Join(cohere, "internal/lint/ecmascript/react/component.go")
	component, err := os.ReadFile(componentPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(component)
	start := strings.Index(text, "func IsEs5ComponentCall(")
	end := strings.Index(text[start:], "func isCreateClassName(") + start
	if start < 0 || end < start {
		t.Fatal("factory oracle anchor drift")
	}
	segment := text[start:end]
	anchor := "isIdentifierNamed(access.Expression, reactPragma)"
	if strings.Count(segment, anchor) != 1 {
		t.Fatal("factory predicate anchor drift")
	}
	component = []byte(text[:start] + strings.Replace(segment, anchor, "AdamicSlot02IdentifierNamed(access.Expression, reactPragma)", 1) + text[end:])
	instrumented := filepath.Join(t.TempDir(), "component.go")
	if err = os.WriteFile(instrumented, component, 0644); err != nil {
		t.Fatal(err)
	}
	replacements[componentPath] = instrumented
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch3.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                    string
		Sources                 int
		Nodes, Names, MathNames []json.RawMessage
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
	t.Logf("%d captured inputs; %d distinct parsed sources; %d nodes; %d factory names; %d math names", len(inputs), corpus.Sources, len(corpus.Nodes), len(corpus.Names), len(corpus.MathNames))
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
		{"standalone factory", "create_class_name.a", "name === 'createReactClass'", "false"},
		{"bare factory calls", "es5_component_call.a", "return isCreateClassName(expression.text);", "return false;"},
		{"React-only namespace", "es5_component_call.a", "!isIdentifierNamed(expression.expression, 'React')", "false"},
		{"math exact membership", "math_function_name.a", "functionName === name", "name.includes(functionName)"},
		{"math calc table entry", "math_function_name.a", "'calc',", "'Calc',"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", "slot02_ast.a", base + "/main.a", base + "/create_class_name.a", base + "/es5_component_call.a", base + "/math_function_name.a"}
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
