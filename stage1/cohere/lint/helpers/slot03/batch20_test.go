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

func batch20Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch20/testdata", []string{"/control_flow_graph.*Builder[E].tryStatement", "/nextjs.IsDocumentFile", "/nextjs.lastSeparator"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch20/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch20.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_slot03_batch20.go"): filepath.Join(here, "cfg_export.go")}
	replacements[filepath.Join(root, "internal/lint/ecmascript/nextjs/adamic_slot03_batch20.go")] = filepath.Join(here, "next_export.go")
	for _, item := range []struct {
		file  string
		names []string
	}{
		{"patterns.go", []string{"patternBind"}},
		{"statements.go", []string{"tryStatement", "statement"}},
		{"cfg.go", []string{"newBlock", "link", "enter", "snapshotForks", "restoreForks", "returnFrame", "throwFrame", "markFinal", "markThrown"}},
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
		if item.file == "cfg.go" {
			if strings.Count(text, "func throwTarget[E any](") != 1 {
				t.Fatal("generic free anchor changed")
			}
			text = strings.Replace(text, "func throwTarget[E any](", "func adamicRawThrowTarget[E any](", 1)
		}
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
func TestBatch20Helpers(t *testing.T) {
	cases, want := batch20Oracle(t)
	entry, _ := filepath.Abs("batch20/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch20JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch20Mutants(t *testing.T) {
	cases, want := batch20Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"try_statement.a", "if (view.hasCatch) {", "if (false) {", ""},
		{"try_statement.a", "if (view.bindingPresent)", "if (false)", ""},
		{"try_statement.a", "frame.position = 1", "frame.position = 0", ""},
		{"try_statement.a", "frame.thrownForked = false", "frame.thrownForked = true", ""},
		{"try_statement.a", "state.stack.pop();", "", ""},
		{"try_statement.a", "if (!view.hasFinally)", "if (false)", ""},
		{"try_statement.a", "for (const source of frame.implicit)", "for (const source of normalEnds)", ""},
		{"try_statement.a", "dependencies.restoreForks(snapshot);", "", ""},
		{"try_statement.a", "if (state.incoming.includes(frame.finallyEntry))", "if (false)", ""},
		{"try_statement.a", "if (state.reachable)", "if (true)", ""},
		{"try_statement.a", "if (frame.returnedAny)", "if (false)", ""},
		{"try_statement.a", "if (frame.thrownAny)", "if (false)", ""},
		{"try_statement.a", "outer.returnedAny = true;", "", ""},
		{"try_statement.a", "outer.thrownAny = true;", "", ""},
		{"try_statement.a", "dependencies.markFinal(state.cur);", "", ""},
		{"try_statement.a", "dependencies.markThrown(state.cur);", "", ""},
		{"try_statement.a", "dependencies.enter(after);\n}", "\n}", ""},
		{"is_document_file.a", "startsWith('_document.')", "startsWith('_document')", ""},
		{"is_document_file.a", "parts.baseName.startsWith('index') && parts.parentName.startsWith('_document')", "parts.baseName === 'index' && parts.parentName === '_document'", ""},
		{"is_document_file.a", "&& parts.parentName.startsWith('_document')", "", ""},
		{"last_separator.a", "if (byte === 92)", "if (false)", ""},
		{"last_separator.a", "if (byte === 47)", "if (false)", ""},
		{"last_separator.a", "if (windows > posix)", "if (windows < posix)", ""},
		{"last_separator.a", "posix = i;", "posix = 0;", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"try_statement.a", "is_document_file.a", "last_separator.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch20", file))
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
					offset := 0
					for offset < len(a[i]) && offset < len(b[i]) && a[i][offset] == b[i][offset] {
						offset++
					}
					start := offset - 60
					if start < 0 {
						start = 0
					}
					t.Logf("compiled semantic mutant caught at line %d byte %d: got %q Go %q", i+1, offset, shorten(a[i][start:]), shorten(b[i][start:]))
					return
				}
			}
			t.Fatal("no changed output line")
		})
	}
}

func batch20JavaScript(t *testing.T, entry string) string {
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

func shorten(value string) string {
	if len(value) > 240 {
		return value[:240] + "..."
	}
	return value
}
