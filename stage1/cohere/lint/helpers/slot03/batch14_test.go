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

func batch14Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch14/testdata", []string{"/collapse.isLineWidth", "/collapse.isImage", "/collapse.isBackgroundPosition"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch14/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch14.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch14.go"): filepath.Join(here, "collapse_export.go")}
	for _, item := range []struct {
		file    string
		anchors map[string]string
	}{
		{"data_type.go", map[string]string{"func IsLength(": "func adamicRawLength(", "func isNumber(": "func adamicRawNumber(", "func isPercentage(": "func adamicRawPercentage(", "func isURL(": "func adamicRawURL("}},
		{"segment.go", map[string]string{"func segment(": "func adamicRawSegment("}},
	} {
		original := filepath.Join(root, "internal/lint/rules/tailwind/collapse", item.file)
		raw, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(raw)
		for old, new := range item.anchors {
			if strings.Count(text, old) != 1 {
				t.Fatal("Go dependency anchor changed")
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
func TestBatch14Helpers(t *testing.T) {
	cases, want := batch14Oracle(t)
	entry, _ := filepath.Abs("batch14/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch14JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch14Mutants(t *testing.T) {
	cases, want := batch14Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"is_line_width.a", "'medium', 'thick'", "'medium'", ""},
		{"is_line_width.a", "if (dependencies.isLength(part)", "dependencies.isNumber(part); if (dependencies.isLength(part)", ""},
		{"is_line_width.a", "dependencies.isLength(part)", "dependencies.isLength(part + 'x')", ""},
		{"is_line_width.a", "dependencies.segment(value, ' ')", "dependencies.segment(value + 'x', ' ')", ""},
		{"is_line_width.a", "dependencies.segment(value, ' ')", "dependencies.segment(value, ',')", ""},
		{"is_line_width.a", "return false;", "return true;", ""},
		{"is_image.a", "/^var\\(/u", "/^var\\(/iu", ""},
		{"is_image.a", "return count > 0;", "return count >= 0;", ""},
		{"is_image.a", "(?:repeating-)?", "", ""},
		{"is_image.a", "conic|linear|radial", "conic|linear", ""},
		{"is_image.a", "element|image|cross-fade|image-set", "element|image|image-set", ""},
		{"is_image.a", "/^(?:element|image|cross-fade|image-set)\\(/u", "/^(?:element|image|cross-fade|image-set)\\(/iu", ""},
		{"is_image.a", "dependencies.isURL(part)", "dependencies.isURL(part + 'x')", ""},
		{"is_image.a", "dependencies.segment(value, ',')", "dependencies.segment(value + 'x', ',')", ""},
		{"is_image.a", "return false;", "continue;", ""},
		{"is_image.a", "if (dependencies.isURL(part)) { count++; continue; }", "if (dependencies.isURL(part)) { continue; }", ""},
		{"is_background_position.a", "'right', 'bottom', 'left'", "'right', 'left'", ""},
		{"is_background_position.a", "/^var\\(/u.test(part)", "false", ""},
		{"is_background_position.a", "return count > 0;", "return count >= 0;", ""},
		{"is_background_position.a", " || dependencies.isPercentage(part)", " && dependencies.isPercentage(part)", ""},
		{"is_background_position.a", "if (dependencies.isLength(part)", "dependencies.isPercentage(part); if (dependencies.isLength(part)", ""},
		{"is_background_position.a", "dependencies.isLength(part)", "dependencies.isLength(part + 'x')", ""},
		{"is_background_position.a", "dependencies.segment(value, ' ')", "dependencies.segment(value, ',')", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"is_line_width.a", "is_image.a", "is_background_position.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch14", file))
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

func batch14JavaScript(t *testing.T, entry string) string {
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
