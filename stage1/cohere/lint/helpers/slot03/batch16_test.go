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

func batch16Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch16/testdata", []string{"/control_flow_graph.*Builder[E].makeReturn", "/control_flow_graph.*Builder[E].makeThrow", "/control_flow_graph.*Builder[E].makeYield"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch16/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch16.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/control_flow_graph/adamic_slot03_batch16.go"): filepath.Join(here, "cfg_export.go")}
	for _, item := range []struct {
		file    string
		anchors map[string]string
	}{
		{"statements.go", map[string]string{"func (b *Builder[E]) makeReturn(": "func (b *Builder[E]) adamicRawMakeReturn(", "func (b *Builder[E]) makeThrow(": "func (b *Builder[E]) adamicRawMakeThrow(", "func (b *Builder[E]) makeYield(": "func (b *Builder[E]) adamicRawMakeYield("}},
		{"cfg.go", map[string]string{"func (b *Builder[E]) returnFrame(": "func (b *Builder[E]) adamicRawReturnFrame(", "func (b *Builder[E]) throwFrame(": "func (b *Builder[E]) adamicRawThrowFrame(", "func (b *Builder[E]) link(": "func (b *Builder[E]) adamicRawLink(", "func (b *Builder[E]) markFinal(": "func (b *Builder[E]) adamicRawMarkFinal(", "func (b *Builder[E]) markThrown(": "func (b *Builder[E]) adamicRawMarkThrown(", "func (b *Builder[E]) makeUnreachable(": "func (b *Builder[E]) adamicRawMakeUnreachable(", "func (b *Builder[E]) newBlock(": "func (b *Builder[E]) adamicRawNewBlock(", "func (b *Builder[E]) enter(": "func (b *Builder[E]) adamicRawEnter(", "func throwTarget[": "func adamicRawThrowTarget["}},
	} {
		original := filepath.Join(root, "internal/lint/ecmascript/control_flow_graph", item.file)
		raw, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		for old, new := range item.anchors {
			if strings.Count(text, old) != 1 {
				t.Fatal("Go anchor changed")
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
	return cases, want
}

// Not parallel: native sanitizer builds and large fixture observations bound memory.
func TestBatch16Helpers(t *testing.T) {
	cases, want := batch16Oracle(t)
	entry, _ := filepath.Abs("batch16/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch16JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch16Mutants(t *testing.T) {
	cases, want := batch16Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"make_return.a", "if (!state.cur.reachable)", "if (false)", ""},
		{"make_return.a", "dependencies.makeUnreachable();", "", ""},
		{"make_return.a", "frame.returnedAny = true;", "frame.returnedAny = false;", ""},
		{"make_return.a", "frame.finallyEntry", "state.cur.index", ""},
		{"make_return.a", "dependencies.markFinal(state.cur.index);", "dependencies.markFinal(state.cur.index + 1);", ""},
		{"make_throw.a", "if (!state.cur.reachable)", "if (false)", ""},
		{"make_throw.a", "dependencies.makeUnreachable();", "", ""},
		{"make_throw.a", "frame.thrownAny = true;", "frame.thrownAny = false;", ""},
		{"make_throw.a", "dependencies.throwTarget(frame)", "frame.finallyEntry", ""},
		{"make_throw.a", "dependencies.markThrown(state.cur.index);", "dependencies.markThrown(state.cur.index + 1);", ""},
		{"make_yield.a", "if (!state.cur.reachable)", "if (false)", ""},
		{"make_yield.a", "frame.returnedAny = true;", "frame.returnedAny = false;", ""},
		{"make_yield.a", "frame.finallyEntry", "state.cur.index", ""},
		{"make_yield.a", "dependencies.markFinal(state.cur.index);", "dependencies.markFinal(state.cur.index + 1);", ""},
		{"make_yield.a", "frame.thrownAny = true;", "frame.thrownAny = false;", ""},
		{"make_yield.a", "dependencies.throwTarget(frame)", "frame.finallyEntry", ""},
		{"make_yield.a", "dependencies.markThrown(state.cur.index);", "dependencies.markThrown(state.cur.index + 1);", ""},
		{"make_yield.a", "dependencies.link(state.cur.index, next);", "", ""},
		{"make_yield.a", "dependencies.enter(next);", "", ""},
		{"make_yield.a", "dependencies.enter(next);", "dependencies.enter(state.cur.index);", ""},
		{"make_return.a", "if (i >= 0)", "if (i > 0)", ""},
		{"make_throw.a", "if (i >= 0)", "if (i > 0)", ""},
		{"make_yield.a", "if (returning >= 0)", "if (returning > 0)", ""},
		{"make_yield.a", "if (throwing >= 0)", "if (throwing > 0)", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"make_return.a", "make_throw.a", "make_yield.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch16", file))
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

func batch16JavaScript(t *testing.T, entry string) string {
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
