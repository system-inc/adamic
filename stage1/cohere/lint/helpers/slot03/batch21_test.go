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

func batch21Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch21/testdata", []string{"/nextjs.splitPath", "/imports.HasPathSegment", "/collapse.gapRootAcceptsModifierOnArbitrary"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch21/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch21.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/nextjs/adamic_slot03_batch21.go"): filepath.Join(here, "next_export.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch21.go"): filepath.Join(here, "collapse_export.go")}
	original := filepath.Join(root, "internal/lint/ecmascript/nextjs/paths.go")
	raw, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Count(text, "func lastSeparator(") != 1 {
		t.Fatal("Go separator anchor changed")
	}
	text = strings.Replace(text, "func lastSeparator(", "func adamicRawLastSeparator(", 1)
	modified := filepath.Join(directory, "paths.go")
	if err = os.WriteFile(modified, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	replacements[original] = modified

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
func TestBatch21Helpers(t *testing.T) {
	cases, want := batch21Oracle(t)
	entry, _ := filepath.Abs("batch21/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch21JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch21Mutants(t *testing.T) {
	cases, want := batch21Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"split_path.a", "index >= 0", "index > 0", ""},
		{"split_path.a", "filePath.slice(index + 1)", "filePath.slice(index)", ""},
		{"split_path.a", "parentIndex >= 0", "parentIndex > 0", ""},
		{"split_path.a", "parent.slice(parentIndex + 1)", "parent.slice(parentIndex)", ""},
		{"split_path.a", "lastSeparator(parent)", "lastSeparator(filePath)", ""},
		{"has_path_segment.a", "segment.length === 0", "false", ""},
		{"has_path_segment.a", "byte === 92", "false", ""},
		{"has_path_segment.a", "byte === 47", "false", ""},
		{"has_path_segment.a", "index <= path.length", "index < path.length", ""},
		{"has_path_segment.a", "equal = false;", "equal = true;", ""},
		{"has_path_segment.a", "start = index + 1", "start = index", ""},
		{"gap_root_accepts_modifier.a", "if (isColor)", "if (!isColor)", ""},
		{"gap_root_accepts_modifier.a", "return false;", "return true;", ""},
		{"gap_root_accepts_modifier.a", "for (const isColor of arms)", "for (const isColor of arms.slice(0,1))", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"split_path.a", "has_path_segment.a", "gap_root_accepts_modifier.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch21", file))
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
					separator, _ := filepath.Abs("batch20/last_separator.a")
					text = strings.ReplaceAll(text, "../batch20/last_separator.a", filepath.ToSlash(separator))
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
					t.Logf("compiled semantic mutant caught at line %d byte %d: got %q Go %q", i+1, offset, batch21Shorten(a[i][start:]), batch21Shorten(b[i][start:]))
					return
				}
			}
			t.Fatal("no changed output line")
		})
	}
}

func batch21JavaScript(t *testing.T, entry string) string {
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

func batch21Shorten(value string) string {
	if len(value) > 240 {
		return value[:240] + "..."
	}
	return value
}
