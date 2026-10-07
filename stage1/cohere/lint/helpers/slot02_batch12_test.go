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

func TestSlot02Batch12(t *testing.T) {
	// Not parallel: one consumer corpus and native sanitizer builds are shared by the baseline and mutants.
	base := "slot02/batch12"
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
	for _, symbol := range []string{"tailwind/collapse.*Theme.Resolve", "control_flow_graph.*Builder[E].variableDeclaration", "tailwind.declineListeners"} {
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
	for _, pair := range [][2]string{{"adamic_slot02_batch12.go", "oracle.go"}, {"internal/lint/rules/tailwind/collapse/adamic_slot02_batch12.go", "collapse_export.go"}, {"internal/lint/ecmascript/control_flow_graph/adamic_slot02_batch12.go", "cfg_export.go"}, {"internal/lint/rules/tailwind/adamic_slot02_batch12.go", "tailwind_export.go"}} {
		path, _ := filepath.Abs(base + "/testdata/" + pair[1])
		replacements[filepath.Join(cohere, pair[0])] = path
	}
	for _, method := range []struct {
		file, anchor, end string
		pairs             [][2]string
	}{
		{"internal/lint/rules/tailwind/collapse/theme.go", "func (theme *Theme) Resolve(", "// ResolveValue returns", [][2]string{{"theme.resolveKey(candidateValue, candidateValuePresent, themeKeys)", "Slot12Key(theme,candidateValue,candidateValuePresent,themeKeys)"}, {"theme.variableReference(themeKey)", "Slot12Reference(theme,themeKey)"}}},
		{"internal/lint/ecmascript/control_flow_graph/statements.go", "func (b *Builder[E]) variableDeclaration(", "func (b *Builder[E]) ifStatement", [][2]string{{"{\n", "{\nslot12Capture(b,node)\n"}, {"b.expr(", "Slot12Expression(b,"}, {"b.patternBind(", "Slot12Bind(b,"}, {"b.patternReads(", "Slot12Reads(b,"}, {"b.patternWrites(", "Slot12Writes(b,"}}},
		{"internal/lint/rules/tailwind/enforce_consistent_class_order.go", "func declineListeners(", "// reportClassOrder", [][2]string{{"ctx.ReportNode(node, rule.Message{", "Slot12Report(ctx,node, rule.Message{"}, {"DesignSystemDeclineMessage(ruleName, designSystem)", "Slot12Describe(ruleName,designSystem)"}}},
	} {
		original := filepath.Join(cohere, method.file)
		data, err = os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		start := strings.Index(source, method.anchor)
		end := strings.Index(source[start:], method.end) + start
		if start < 0 || end <= start {
			t.Fatal("method boundary drift")
		}
		section := source[start:end]
		for _, pair := range method.pairs {
			if !strings.Contains(section, pair[0]) {
				t.Fatal("dependency anchor drift")
			}
			if pair[0] == "{\n" {
				section = strings.Replace(section, pair[0], pair[1], 1)
			} else {
				section = strings.ReplaceAll(section, pair[0], pair[1])
			}
		}
		target := filepath.Join(t.TempDir(), filepath.Base(method.file))
		if err = os.WriteFile(target, []byte(source[:start]+section+source[end:]), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[original] = target
	}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch12.go"))
	data = run(t, "", oracle, config)
	var corpus struct {
		Want                string
		Theme, CFG, Decline []json.RawMessage
		Calls               int
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
	t.Logf("%d captured inputs, %d theme replays, %d CFG replays, %d actual CFG helper entries", len(inputs), len(corpus.Theme), len(corpus.CFG), corpus.Calls)
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
		{"theme miss guard", "theme_resolve.a", "if (!key.found)", "if (false)"},
		{"theme caller inline", "theme_resolve.a", "(options | value.options)", "value.options"},
		{"theme entry inline", "theme_resolve.a", "(options | value.options)", "options"},
		{"theme literal empty presence", "theme_resolve.a", "value: value.value, found: true", "value: value.value, found: value.value !== ''"},
		{"theme key identity", "theme_resolve.a", "theme, key.value", "theme, candidate"},
		{"theme identity", "theme_resolve.a", "deps.key(theme, candidate", "deps.key(0, candidate"},
		{"theme candidate presence", "theme_resolve.a", "candidate, present, keys", "candidate, !present, keys"},
		{"theme key slice identity", "theme_resolve.a", "present, keys);", "present, keys.slice());"},
		{"declaration nil guard", "variable_declaration.a", "if (node === -1) { return; }", "if (node === -1) { deps.reads(builder,node); return; }"},
		{"declaration annotation", "variable_declaration.a", "deps.expression(builder, declaration.type);", ""},
		{"declaration destructuring order", "variable_declaration.a", "deps.expression(builder, declaration.initializer);\n        deps.bind(builder, declaration.name);", "deps.bind(builder, declaration.name);\n        deps.expression(builder, declaration.initializer);"},
		{"declaration destructuring route", "variable_declaration.a", "if (declaration.pattern)", "if (false)"},
		{"declaration uninitialized write", "variable_declaration.a", "deps.reads(builder, declaration.name); return;", "deps.writes(builder, declaration.name); return;"},
		{"declaration initialized order", "variable_declaration.a", "deps.reads(builder, declaration.name);\n    deps.expression(builder, declaration.initializer);", "deps.expression(builder, declaration.initializer);\n    deps.reads(builder, declaration.name);"},
		{"declaration write identity", "variable_declaration.a", "deps.writes(builder, declaration.name);", "deps.writes(builder, -1);"},
		{"decline repeated guard", "decline_listeners.a", "if (this.reported)", "if (false)"},
		{"decline guard before description", "decline_listeners.a", "this.reported = true;\n        this.deps.report(this.context, node, 'designSystemUnavailable', this.deps.describe(this.ruleName, this.system));", "this.deps.report(this.context, node, 'designSystemUnavailable', this.deps.describe(this.ruleName, this.system));\n        this.reported = true;"},
		{"decline message id", "decline_listeners.a", "'designSystemUnavailable'", "'unavailable'"},
		{"decline context identity", "decline_listeners.a", "this.deps.report(this.context, node", "this.deps.report(0, node"},
		{"decline node identity", "decline_listeners.a", "this.deps.report(this.context, node", "this.deps.report(this.context, 0"},
		{"decline system identity", "decline_listeners.a", "this.deps.describe(this.ruleName, this.system)", "this.deps.describe(this.ruleName, 0)"},
	}

	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			files := []string{"options_json.ts", base + "/main.a", base + "/theme_resolve.a", base + "/variable_declaration.a", base + "/decline_listeners.a"}
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
