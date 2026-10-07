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

func batch8Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch8/testdata", []string{"/collapse.decodeArbitraryValue", "/collapse.recursivelyDecodeArbitraryValues", "/collapse.registerThemeBreakpointVariants"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch8/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch8.go")
	overlay := filepath.Join(directory, "overlay.json")

	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch8.go"): filepath.Join(here, "collapse_export.go")}})
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
func TestBatch8Helpers(t *testing.T) {
	cases, want := batch8Oracle(t)
	entry, _ := filepath.Abs("batch8/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch8JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch8Mutants(t *testing.T) {
	cases, want := batch8Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{{"recursively_decode_arbitrary_values.a", "node.value === 'url' || node.value.endsWith('_url')", "false", ""}, {"recursively_decode_arbitrary_values.a", "child.value = convert(child.value, true);", "child.value = convert(child.value, false);", ""}, {"recursively_decode_arbitrary_values.a", "node.kind === 'separator' || node.kind === 'word'", "true", ""}, {"decode_arbitrary_value.a", "if(!input.includes('('))", "if(true)", ""}, {"decode_arbitrary_value.a", "return dependencies.math(css);", "return css;", ""}, {"decode_arbitrary_value.a", "const arena = dependencies.parse(input);", "const arena = dependencies.parse(input + '_');", ""}, {"register_theme_breakpoint_variants.a", "if(!themePresent)", "if(false)", ""}, {"register_theme_breakpoint_variants.a", "if(!group.present)", "if(false)", ""}, {"register_theme_breakpoint_variants.a", "name === '' || dependencies.has(name)", "name === ''", ""}, {"register_theme_breakpoint_variants.a", "group.order, 'static'", "group.order + 1, 'static'", ""}} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"recursively_decode_arbitrary_values.a", "decode_arbitrary_value.a", "register_theme_breakpoint_variants.a", "value_arena.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch8", file))
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

func batch8JavaScript(t *testing.T, entry string) string {
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
