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

func batch3Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch3/testdata", []string{"/text.hexValue", "/text.UnescapeStringLiteralText", "/structure.parameterNodes"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch3/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch3.go")
	overlay := filepath.Join(directory, "overlay.json")

	original := filepath.Join(root, "internal/lint/ecmascript/text/jsx_entities.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "if replacement, ok := decodeEntity(entity[1 : len(entity)-1]); ok {"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("Go entity observation anchor changed")
	}
	modified := strings.Replace(string(source), anchor, "adamicEntityBodies = append(adamicEntityBodies,entity[1 : len(entity)-1]); "+anchor, 1)
	traced := filepath.Join(directory, "jsx_entities.go")
	if err = os.WriteFile(traced, []byte(modified), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: traced, virtual: filepath.Join(here, "hex_oracle.go"), filepath.Join(root, "internal/lint/ecmascript/text/adamic_slot03_batch3.go"): filepath.Join(here, "text_export.go"), filepath.Join(root, "internal/lint/rules/structure/adamic_slot03_batch3.go"): filepath.Join(here, "structure_export.go")}})
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
func TestBatch3Helpers(t *testing.T) {
	cases, want := batch3Oracle(t)
	entry, _ := filepath.Abs("batch3/hex_main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch3JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch3Mutants(t *testing.T) {
	cases, want := batch3Oracle(t)
	for _, mutant := range []struct{ file, old, new string }{{"hex_value.a", "character <= 70", "character <= 69"}, {"unescape_string_literal_text.a", "result += replacement.text;", "result += body;"}, {"parameter_nodes.a", "return nodes;", "return nodes.slice(0);"}, {"parameter_nodes.a", "if(!parametersPresent) { return []; }", "if(!parametersPresent) { return nodes; }"}} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"hex_value.a", "unescape_string_literal_text.a", "parameter_nodes.a", "hex_main.a"} {
				data, err := os.ReadFile(filepath.Join("batch3", file))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if file == mutant.file {
					if strings.Count(text, mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					text = strings.Replace(text, mutant.old, mutant.new, 1)
				}
				if file == "hex_main.a" {
					reader, _ := filepath.Abs("../options_json.ts")
					text = strings.ReplaceAll(text, "../../options_json.ts", filepath.ToSlash(reader))
				}
				if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := command(t, "", build(t, filepath.Join(scratch, "hex_main.a")), cases)
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

func batch3JavaScript(t *testing.T, entry string) string {
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
