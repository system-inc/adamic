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

func batch9Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch9/testdata", []string{"/collapse.ParseCSS", "/tailwind.DesignSystemDeclineMessage", "/collapse.*LoadedDesignSystem.Theme"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch9/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch9.go")
	overlay := filepath.Join(directory, "overlay.json")

	original := filepath.Join(root, "internal/lint/rules/tailwind/collapse/css_parser.go")
	parser, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	modified := string(parser)
	for old, replacement := range map[string]string{"func ParseAtRule(": "func adamicOriginalAtRule(", "func parseCSSDeclaration(": "func adamicOriginalDeclaration(", "func parseCSSString(": "func adamicOriginalString("} {
		if strings.Count(modified, old) != 1 {
			t.Fatal("Go observation anchor changed")
		}
		modified = strings.Replace(modified, old, replacement, 1)
	}
	modified = strings.ReplaceAll(modified, "strings.TrimSpace(", "adamicTrim(")
	instrumented := filepath.Join(directory, "css_parser.go")
	if err = os.WriteFile(instrumented, []byte(modified), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: instrumented, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_integer.go"): filepath.Join(here, "integer.go"), virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch9.go"): filepath.Join(here, "collapse_export.go")}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases, filepath.Join(root, "internal/lint/rules/tailwind/collapse/testdata/cssparser_fixtures.json"))
	return cases, want
}

// Not parallel: native sanitizer builds and large fixture observations bound memory.
func TestBatch9Helpers(t *testing.T) {
	cases, want := batch9Oracle(t)
	entry, _ := filepath.Abs("batch9/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch9JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch9Mutants(t *testing.T) {
	cases, want := batch9Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"parse_css.a", "for(const id of licenses) {output.push(id);}", "for(const id of licenses) {if(id<0) {output.push(id);}}", ""},
		{"parse_css.a", "if(input.startsWith('\\u00ef\\u00bb\\u00bf'))", "if(false)", ""},
		{"parse_css.a", "c===45 && peek(i+1)===45 && state.buffer===''", "false", ""},
		{"parse_css.a", "c===59 && top()!==41", "c===59", ""},
		{"parse_css.a", "if(closing.length>0 && state.parent>=0)", "if(false)", ""},
		{"parse_css.a", "failed: true, message, offset};", "failed: true, message, offset: offset+1};", ""},
		{"design_system_decline_message.a", "+ where +", "+ '' +", ""},
		{"design_system_decline_message.a", "errorPresent ? errorText : '<nil>'", "errorText", ""},
		{"loaded_theme.a", "return system.theme;", "return -1;", ""},
		{"loaded_theme.a", "return system.theme;", "const theme=system.theme;system.theme=-1;return theme;", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"parse_css.a", "design_system_decline_message.a", "loaded_theme.a", "css_model.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch9", file))
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
					whitespace, _ := filepath.Abs("batch4/followed_by_whitespace.a")
					text = strings.ReplaceAll(text, "../batch4/followed_by_whitespace.a", filepath.ToSlash(whitespace))
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

func batch9JavaScript(t *testing.T, entry string) string {
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
