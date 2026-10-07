package slot03

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func batch18Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch18/testdata", []string{"/control_flow_graph.*Builder[E].binaryExpression", "/control_flow_graph.*Builder[E].classLike", "/control_flow_graph.*Builder[E].ifStatement"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch18/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch18.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_slot03_batch18.go"): filepath.Join(here, "cfg_export.go")}
	for _, item := range []struct {
		file  string
		names []string
	}{
		{"expressions.go", []string{"binaryExpression", "classLike", "expr", "decorators", "typeParameters", "memberHeader", "condition"}},
		{"statements.go", []string{"ifStatement", "statement"}},
		{"patterns.go", []string{"patternBind", "patternReads", "patternWrites"}},
		{"cfg.go", []string{"newBlock", "link", "enter"}},
	} {
		original := filepath.Join(root, "internal/lint/ecmascript/control_flow_graph", item.file)
		raw, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		for _, name := range item.names {
			old := "func (b *Builder[E]) " + name + "("
			new := "func (b *Builder[E]) adamicRaw" + strings.ToUpper(name[:1]) + name[1:] + "("
			if strings.Count(text, old) != 1 {
				t.Fatal("Go anchor changed", name)
			}
			text = strings.Replace(text, old, new, 1)
		}
		if item.file == "patterns.go" {
			old := "func isDestructuringTarget("
			if strings.Count(text, old) != 1 {
				t.Fatal("Go free helper anchor changed")
			}
			text = strings.Replace(text, old, "func adamicRawIsDestructuringTarget(", 1)
		}
		modified := filepath.Join(directory, item.file)
		if err := os.WriteFile(modified, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[original] = modified
	}

	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases)
	stats, err := os.ReadFile(cases + ".stats.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("real helper observations: %s", stats)
	return cases, want
}

// Not parallel: native sanitizer builds and large fixture observations bound memory.
func TestBatch18Helpers(t *testing.T) {
	cases, want := batch18Oracle(t)
	entry, _ := filepath.Abs("batch18/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch18JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch18Mutants(t *testing.T) {
	cases, want := batch18Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"binary_expression.a", "if (binary.shortCircuit)", "if (false)", ""},
		{"binary_expression.a", "else if (binary.logicalAssignment)", "else if (false)", ""},
		{"binary_expression.a", "else if (binary.equals)", "else if (false)", ""},
		{"binary_expression.a", "else if (binary.assignment)", "else if (false)", ""},
		{"binary_expression.a", "if (dependencies.isDestructuringTarget(binary.left))", "if (false)", ""},
		{"binary_expression.a", "dependencies.patternBind(binary.left);", "dependencies.patternBind(binary.right);", ""},
		{"binary_expression.a", "dependencies.expr(binary.left);\n    dependencies.expr(binary.right);", "dependencies.expr(binary.right);\n    dependencies.expr(binary.left);", ""},
		{"class_like.a", "if (!view.present)", "if (false)", ""},
		{"class_like.a", "dependencies.decorators(view.node);", "", ""},
		{"class_like.a", "dependencies.typeParameters(view.node);", "", ""},
		{"class_like.a", "if (view.heritagePresent)", "if (true)", ""},
		{"class_like.a", "!heritage.present || !heritage.typesPresent", "false", ""},
		{"class_like.a", "if (view.membersPresent)", "if (true)", ""},
		{"class_like.a", "dependencies.memberHeader(member);", "dependencies.memberHeader(view.node);", ""},
		{"if_statement.a", "let elseEntry = after;", "let elseEntry = thenEntry;", ""},
		{"if_statement.a", "dependencies.condition(statement.expression, thenEntry, elseEntry);", "dependencies.condition(statement.expression, elseEntry, thenEntry);", ""},
		{"if_statement.a", "dependencies.statement(statement.thenStatement);", "dependencies.statement(statement.elseStatement);", ""},
		{"if_statement.a", "dependencies.enter(after);", "", ""},
		{"if_statement.a", "if (statement.elseStatement >= 0) { elseEntry = dependencies.newBlock(); }", "if (false) { elseEntry = dependencies.newBlock(); }", ""},
		{"if_statement.a", "if (statement.elseStatement >= 0) {\n    dependencies.enter(elseEntry);", "if (false) {\n    dependencies.enter(elseEntry);", ""},
		{"if_statement.a", "if (statement.elseStatement >= 0) {\n    dependencies.enter(elseEntry);", "if (true) {\n    dependencies.enter(elseEntry);", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"binary_expression.a", "class_like.a", "if_statement.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch18", file))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if file == mutant.file {
					if strings.Count(text, mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					text = mutant.prefix + strings.Replace(text, mutant.old, mutant.new, 1)
				}
				if file == "main.a" {
					reader, _ := filepath.Abs("../options_json.ts")
					text = strings.ReplaceAll(text, "../../options_json.ts", filepath.ToSlash(reader))
				}
				if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := command(t, "", build(t, filepath.Join(scratch, "main.a")), cases)
			if bytes.Equal(got, want) {
				t.Fatal("semantic mutant survived")
			}
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for i := 0; i < len(a) && i < len(b); i++ {
				if a[i] != b[i] {
					t.Logf("compiled semantic mutant caught at line %d: got %q Go %q", i+1, a[i], b[i])
					return
				}
			}
			t.Fatal("no changed output line")
		})
	}
}

func batch18JavaScript(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "helper.mjs")
	if err = os.WriteFile(path, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
