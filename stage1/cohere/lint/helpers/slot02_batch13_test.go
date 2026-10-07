package helpers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlot02Batch13(t *testing.T) {
	// Not parallel: baseline and mutants share the actual Go consumer corpus.
	base := "slot02/batch13"
	wanted := map[string]bool{"array-callback-return": true, "consistent-return": true, "no-unreachable-loop": true, "react-hooks/rules-of-hooks": true, "better-tailwindcss/enforce-consistent-class-order": true, "better-tailwindcss/enforce-shorthand-classes": true, "better-tailwindcss/no-conflicting-classes": true, "better-tailwindcss/no-unknown-classes": true}
	observed := map[string]bool{}
	inputs := []map[string]string{}
	for _, name := range []string{"slot02/batch6/testdata/sources.jsonl.gz", "slot02/batch10/testdata/sources.jsonl.gz"} {
		f, e := os.Open(name)
		if e != nil {
			t.Fatal(e)
		}
		z, e := gzip.NewReader(f)
		if e != nil {
			t.Fatal(e)
		}
		s := bufio.NewScanner(z)
		s.Buffer(make([]byte, 65536), 16<<20)
		for s.Scan() {
			var r map[string]string
			if e = json.Unmarshal(s.Bytes(), &r); e != nil {
				t.Fatal(e)
			}
			if wanted[r["rule"]] {
				observed[r["rule"]] = true
				inputs = append(inputs, r)
			}
		}
		if e = s.Err(); e != nil {
			t.Fatal(e)
		}
		z.Close()
		f.Close()
	}
	for r := range wanted {
		if !observed[r] {
			t.Fatalf("consumer missing: %s", r)
		}
	}
	cohere, _ := filepath.Abs("../../../../cohere")
	paths, err := filepath.Glob(filepath.Join(cohere, "internal/lint/rules/tailwind/collapse", "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	config := smallFixture(t, map[string]any{"Inputs": inputs, "Paths": paths})
	// Dependency substitutions retain the complete actual methods and their recursive calls.
	replacements := map[string]string{}
	for _, pair := range [][2]string{{"adamic_slot02_batch13.go", "oracle.go"}, {"internal/lint/ecmascript/control_flow_graph/adamic_slot02_batch13.go", "cfg_export.go"}, {"internal/lint/rules/tailwind/collapse/adamic_slot02_batch13.go", "collapse_export.go"}} {
		p, _ := filepath.Abs(base + "/testdata/" + pair[1])
		replacements[filepath.Join(cohere, pair[0])] = p
	}
	for _, m := range []struct {
		file, start, end string
		pairs            [][2]string
	}{
		{"internal/lint/ecmascript/control_flow_graph/expressions.go", "func (b *Builder[E]) condition(", "// visitUnknown", [][2]string{{"b.expr(", "Slot13Call(\"expr\",b,"}, {"b.link(", "Slot13Link(b,"}, {"b.newBlock()", "Slot13New(b)"}, {"b.enter(", "Slot13Enter(b,"}}},
		{"internal/lint/ecmascript/control_flow_graph/statements.go", "func (b *Builder[E]) statement(", "func (b *Builder[E]) variableDeclarationList", [][2]string{{"b.reachedStatement(", "Slot13Call(\"reached\",b,"}, {"b.statements(", "Slot13List(b,"}, {"b.variableDeclarationList(", "Slot13Call(\"declarations\",b,"}, {"b.expr(", "Slot13Call(\"expr\",b,"}, {"b.ifStatement(", "Slot13Call(\"if\",b,"}, {"b.whileStatement(", "Slot13Call(\"while\",b,"}, {"b.doStatement(", "Slot13Call(\"do\",b,"}, {"b.forStatement(", "Slot13Call(\"for\",b,"}, {"b.forInOfStatement(", "Slot13Call(\"forInOf\",b,"}, {"b.switchStatement(", "Slot13Call(\"switch\",b,"}, {"b.tryStatement(", "Slot13Call(\"try\",b,"}, {"b.labeledStatement(", "Slot13Call(\"label\",b,"}, {"b.makeReturn()", "Slot13Call(\"return\",b,nil)"}, {"b.makeThrow()", "Slot13Call(\"throw\",b,nil)"}, {"b.makeBreak(", "Slot13Call(\"break\",b,"}, {"b.makeContinue(", "Slot13Call(\"continue\",b,"}, {"b.classLike(", "Slot13Call(\"class\",b,"}, {"b.visitUnknown(", "Slot13Call(\"unknown\",b,"}}},
		{"internal/lint/rules/tailwind/collapse/utility.go", "func (state *utilityEvaluation) resolveValueFunction(", "// resolveThemeArgument", [][2]string{{"ParseValue(", "Slot13Parse("}, {"state.resolveThemeArgument(", "Slot13Theme(state,"}, {"state.resolveBareArgument(", "Slot13Bare(state,"}, {"state.resolveArbitraryArgument(", "Slot13Arbitrary(state,"}}},
	} {
		p := filepath.Join(cohere, m.file)
		data, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		s := string(data)
		a := strings.Index(s, m.start)
		if a < 0 {
			t.Fatal("method missing")
		}
		b := a + strings.Index(s[a:], m.end)
		if b <= a {
			t.Fatal("end missing")
		}
		body := s[a:b]
		for _, pair := range m.pairs {
			if !strings.Contains(body, pair[0]) {
				t.Fatalf("dependency anchor missing: %s", pair[0])
			}
			body = strings.ReplaceAll(body, pair[0], pair[1])
		}
		target := filepath.Join(t.TempDir(), filepath.Base(p))
		if e = os.WriteFile(target, []byte(s[:a]+body+s[b:]), 0644); e != nil {
			t.Fatal(e)
		}
		replacements[p] = target
	}
	overlay := smallFixture(t, map[string]any{"Replace": replacements})
	oracle := filepath.Join(t.TempDir(), "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlay, "-o", oracle, filepath.Join(cohere, "adamic_slot02_batch13.go"))
	data := run(t, "", oracle, config)
	var corpus struct {
		Want                 string
		Nodes, Cases, Values []json.RawMessage
		CFGWant, ValueWant   []string
	}
	if e := json.Unmarshal(data, &corpus); e != nil {
		t.Fatal(e)
	}
	var input map[string]json.RawMessage
	if e := json.Unmarshal(data, &input); e != nil {
		t.Fatal(e)
	}
	delete(input, "Want")
	delete(input, "CFGWant")
	delete(input, "ValueWant")
	path := smallFixture(t, input)
	want := []byte(corpus.Want)
	t.Logf("%d consumer inputs across all eight rules; %d actual nodes; %d CFG queries; %d value cases", len(inputs), len(corpus.Nodes), len(corpus.Cases), len(corpus.Values))
	entry, _ := filepath.Abs(base + "/main.a")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(t *testing.T, entry string, inputPath string) []byte {
		t.Helper()
		node := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, inputPath)
		p, e := load.Load([]string{entry})
		if e != nil {
			t.Fatal(e)
		}
		ir, e := lower.Lower(context.Background(), p)
		if e != nil {
			t.Fatal(e)
		}
		bin := filepath.Join(t.TempDir(), "program")
		if e = native.Build(native.C(ir), bin, native.Options{Sanitize: true}); e != nil {
			t.Fatal(e)
		}
		got := run(t, "", bin, inputPath)
		js := filepath.Join(t.TempDir(), "program.mjs")
		if e = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
			t.Fatal(e)
		}
		emitted := run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, js, inputPath)
		compare(t, got, node)
		compare(t, emitted, node)
		return got
	}
	compare(t, check(t, entry, path), want)
	t.Log("Actual Go, source Node, emitted JavaScript and sanitized native agree")
	// Keep full corpus parity above. Sample mutant queries by kind/operator/mode/builder,
	// plus all final control nodes; every expected row is independently recorded by Go.
	if len(corpus.CFGWant) != len(corpus.Cases) || len(corpus.ValueWant) != len(corpus.Values) {
		t.Fatal("Go row observations missing")
	}
	type view struct{ Kind, Operator string }
	views := []view{}
	for _, row := range corpus.Nodes {
		var v view
		if err := json.Unmarshal(row, &v); err != nil {
			t.Fatal(err)
		}
		views = append(views, v)
	}
	sample := func(op string) (string, []byte) {
		rows := []json.RawMessage{}
		truth := []string{}
		counts := map[string]int{}
		if op == "value" {
			truth = corpus.ValueWant
		} else {
			for i, row := range corpus.Cases {
				var c struct {
					Op                  string
					Node, Mode, Builder int
				}
				if err := json.Unmarshal(row, &c); err != nil {
					t.Fatal(err)
				}
				if c.Op != op {
					continue
				}
				kind, operator := "nil", ""
				if c.Node >= 0 {
					kind = views[c.Node].Kind
					operator = views[c.Node].Operator
				}
				key := c.Op + ":" + kind + ":" + operator + ":" + string(rune('0'+c.Mode)) + ":" + string(rune('0'+c.Builder))
				if counts[key] >= 2 && c.Node < len(views)-200 {
					continue
				}
				counts[key]++
				rows = append(rows, row)
				truth = append(truth, corpus.CFGWant[i])
			}
		}
		values := []json.RawMessage{}
		nodeRows := corpus.Nodes
		if op == "value" {
			values = corpus.Values
			nodeRows = []json.RawMessage{}
		}
		t.Logf("%s mutant corpus: %d CFG queries, %d value cases", op, len(rows), len(values))
		return smallFixture(t, map[string]any{"Nodes": nodeRows, "Cases": rows, "Values": values}), []byte(strings.Join(truth, "\n") + "\n")
	}
	conditionPath, conditionWant := sample("condition")
	statementPath, statementWant := sample("statement")
	valuePath, valueWant := sample("value")

	mutants := []struct{ name, file, anchor, replacement string }{
		{"condition negation targets", "condition.a", "condition(builder,view.a,whenFalse,whenTrue,nodes,builders,deps)", "condition(builder,view.a,whenTrue,whenFalse,nodes,builders,deps)"},
		{"condition and continuation", "condition.a", "condition(builder,view.a,right,whenFalse,nodes,builders,deps)", "condition(builder,view.a,whenTrue,whenFalse,nodes,builders,deps)"},
		{"condition nullish false continuation", "condition.a", "deps.link(builder,state.current,whenFalse); deps.link(builder,state.current,right);", "deps.link(builder,state.current,right);"},
		{"condition hook presence", "condition.a", "if (deps.hookPresent)", "if (true)"},
		{"statement notification", "statement.a", "deps.call('reached',builder,node);", ""},
		{"statement return order", "statement.a", "deps.call('expr',builder,view.a); deps.call('return',builder,-1);", "deps.call('return',builder,-1); deps.call('expr',builder,view.a);"},
		{"statement default traversal", "statement.a", "for (const child of view.children)", "for (const child of view.children.slice().reverse())"},
		{"statement with body", "statement.a", "statement(builder,view.b,nodes,deps);", "statement(builder,-1,nodes,deps);"},
		{"value default name", "resolve_value_function.a", "argument.value === '--default'", "argument.value === '--value'"},
		{"value matching quotes", "resolve_value_function.a", "word[word.length-1] === quote", "true"},
		{"value bare ratio", "resolve_value_function.a", "if (result.found) { return result; }", "if (result.found) { return {nodes:result.nodes,ratio:false,found:true}; }"},
		{"value theme fallthrough", "resolve_value_function.a", "} else if (view.kind === 'named'", "} if (view.kind === 'named'"},
	}
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, n := range []string{"options_json.ts", base + "/main.a", base + "/condition.a", base + "/statement.a", base + "/resolve_value_function.a"} {
				d, e := os.ReadFile(n)
				if e != nil {
					t.Fatal(e)
				}
				if n == base+"/"+m.file {
					count := strings.Count(string(d), m.anchor)
					if count == 0 {
						t.Fatal("mutant drift")
					}
					d = []byte(strings.ReplaceAll(string(d), m.anchor, m.replacement))
				}
				target := n
				if n == "options_json.ts" {
					target = "options_json.a"
				}
				if n == base+"/main.a" {
					d = []byte(strings.Replace(string(d), "../../options_json.ts", "../../options_json.a", 1))
				}
				p := filepath.Join(dir, target)
				if e = os.MkdirAll(filepath.Dir(p), 0755); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(p, d, 0644); e != nil {
					t.Fatal(e)
				}
			}
			mutantPath, mutantWant := conditionPath, conditionWant
			if m.file == "statement.a" {
				mutantPath, mutantWant = statementPath, statementWant
			} else if m.file == "resolve_value_function.a" {
				mutantPath, mutantWant = valuePath, valueWant
			}
			got := check(t, filepath.Join(dir, base, "main.a"), mutantPath)
			if bytes.Equal(got, mutantWant) {
				t.Fatal("compiled semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(mutantWant), "\n")
			for i, line := range a {
				if i < len(b) && line != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q; Go %q", i+1, line, b[i])
					break
				}
			}
		})
	}
}
