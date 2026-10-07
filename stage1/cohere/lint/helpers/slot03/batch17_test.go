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

func batch17Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch17/testdata", []string{"/control_flow_graph.*Builder[E].conditionalExpression", "/control_flow_graph.*Builder[E].bindWithDefault", "/control_flow_graph.*Builder[E].memberHeader"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch17/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch17.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_slot03_batch17.go"): filepath.Join(here, "cfg_export.go")}
	for _, item := range []struct {
		file  string
		names []string
	}{
		{"expressions.go", []string{"conditionalExpression", "memberHeader", "expr", "decorators"}},
		{"patterns.go", []string{"bindWithDefault", "patternBind"}},
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
func TestBatch17Helpers(t *testing.T) {
	cases, want := batch17Oracle(t)
	entry, _ := filepath.Abs("batch17/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch17JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch17Mutants(t *testing.T) {
	cases, want := batch17Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"conditional_expression.a", "dependencies.expr(condition);", "", ""},
		{"conditional_expression.a", "dependencies.link(testEnd, whenFalse);", "dependencies.link(state.cur, whenFalse);", ""},
		{"conditional_expression.a", "dependencies.expr(whenTrueNode);", "dependencies.expr(whenFalseNode);", ""},
		{"conditional_expression.a", "dependencies.enter(join);", "", ""},
		{"bind_with_default.a", "if (fallback >= 0)", "if (false)", ""},
		{"bind_with_default.a", "dependencies.patternBind(target);", "dependencies.patternBind(fallback);", ""},
		{"bind_with_default.a", "dependencies.enter(join);", "", ""},
		{"member_header.a", "dependencies.decorators(member.node);", "", ""},
		{"member_header.a", "dependencies.decorators(parameter);", "dependencies.decorators(member.node);", ""},
		{"member_header.a", "if (member.computed)", "if (false)", ""},
		{"member_header.a", "if (member.property)", "if (false)", ""},
		{"member_header.a", "dependencies.expr(typeNode);", "dependencies.expr(member.typeNode);", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"conditional_expression.a", "bind_with_default.a", "member_header.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch17", file))
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

func batch17JavaScript(t *testing.T, entry string) string {
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
