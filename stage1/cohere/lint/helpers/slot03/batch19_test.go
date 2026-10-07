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

func batch19Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch19/testdata", []string{"/control_flow_graph.*Builder[E].expr", "/control_flow_graph.*Builder[E].patternBind", "/control_flow_graph.*Builder[E].switchStatement"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch19/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch19.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_slot03_batch19.go"): filepath.Join(here, "cfg_export.go")}
	replacements[filepath.Join(root, "TypeScript/tsc/internal/ast/adamic_slot03_batch19.go")] = filepath.Join(here, "ast_controls.go")
	for _, item := range []struct {
		file  string
		names []string
	}{
		{"expressions.go", []string{"expr", "binaryExpression", "conditionalExpression", "updateExpression", "accessOrCall", "classLike", "nestedFunction", "visitUnknown"}},
		{"patterns.go", []string{"patternBind", "bindWithDefault"}},
		{"statements.go", []string{"switchStatement", "statements", "pushJump", "popJump", "makeYield"}},
		{"cfg.go", []string{"read", "write", "firstThrowableFork", "newBlock", "link", "enter", "linkWithCycleBarrier", "setCycleBarrier"}},
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
		modified := filepath.Join(directory, item.file)
		if err = os.WriteFile(modified, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		replacements[original] = modified
	}
	for _, item := range []struct{ file, name string }{{"helpers.go", "isThrowableIdentifier"}, {"roots.go", "IsRoot"}} {
		original := filepath.Join(root, "internal/lint/ecmascript/control_flow_graph", item.file)
		raw, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		old := "func " + item.name + "("
		if strings.Count(text, old) != 1 {
			t.Fatal("Go free anchor changed", item.name)
		}
		text = strings.Replace(text, old, "func adamicRaw"+item.name+"(", 1)
		modified := filepath.Join(directory, item.file)
		if err = os.WriteFile(modified, []byte(text), 0644); err != nil {
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
func TestBatch19Helpers(t *testing.T) {
	cases, want := batch19Oracle(t)
	entry, _ := filepath.Abs("batch19/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch19JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch19Mutants(t *testing.T) {
	cases, want := batch19Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"expression.a", "if (!view.present)", "if (false)", ""},
		{"expression.a", "if (view.hookPresent)", "if (false)", ""},
		{"expression.a", "if (view.hookPresent)", "if (true)", ""},
		{"expression.a", "dependencies.read(view.node);", "", ""},
		{"expression.a", "if (dependencies.throwable(view.node))", "if (false)", ""},
		{"expression.a", "dependencies.binaryExpression(view.node);", "dependencies.expr(view.node);", ""},
		{"expression.a", "if (view.update)", "if (false)", ""},
		{"expression.a", "dependencies.expr(view.operand); dependencies.makeYield();", "dependencies.makeYield(); dependencies.expr(view.operand);", ""},
		{"expression.a", "if (dependencies.isRoot(view.node))", "if (false)", ""},
		{"expression.a", "dependencies.visitUnknown(child);", "dependencies.visitUnknown(view.node);", ""},
		{"pattern_bind.a", "if (!view.present)", "if (false)", ""},
		{"pattern_bind.a", "dependencies.write(view.node);", "", ""},
		{"pattern_bind.a", "if (!view.elementsPresent)", "if (false)", ""},
		{"pattern_bind.a", "if (!element.present)", "if (false)", ""},
		{"pattern_bind.a", "if (element.computed)", "if (false)", ""},
		{"pattern_bind.a", "dependencies.bindWithDefault(element.target, element.fallback);", "dependencies.bindWithDefault(element.fallback, element.target);", ""},
		{"pattern_bind.a", "dependencies.patternBind(property.target); break;\n          case 'shorthand':", "dependencies.patternBind(property.fallback); break;\n          case 'shorthand':", ""},
		{"pattern_bind.a", "if (view.equals)", "if (false)", ""},
		{"switch_statement.a", "dependencies.expr(view.expression);", "", ""},
		{"switch_statement.a", "if (clause.isDefault) { defaultIndex = i; continue; }", "if (false) { defaultIndex = i; continue; }", ""},
		{"switch_statement.a", "dependencies.linkWithCycleBarrier(testCur, bodies[defaultIndex] ?? -1, false)", "dependencies.linkWithCycleBarrier(testCur, after, false)", ""},
		{"switch_statement.a", "const barrier = hadSwitchBreak || beforeFirstCase;", "const barrier = false;", ""},
		{"switch_statement.a", "hadSwitchBreak = state.jumps[switchJump] ?? false;", "hadSwitchBreak = false;", ""},
		{"switch_statement.a", "dependencies.statements(clause.statements);", "", ""},
		{"switch_statement.a", "dependencies.popJump();", "", ""},
		{"switch_statement.a", "dependencies.enter(after); return;", "return;", ""},
		{"switch_statement.a", "dependencies.linkWithCycleBarrier(fallthroughFrom, bodies[i] ?? -1, barrier);", "dependencies.linkWithCycleBarrier(-1, bodies[i] ?? -1, barrier);", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"expression.a", "pattern_bind.a", "switch_statement.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch19", file))
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

func batch19JavaScript(t *testing.T, entry string) string {
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
