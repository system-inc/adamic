package slot03

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func batch3Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch3/testdata", []string{"/text.hexValue"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch3/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch3.go")
	overlay := filepath.Join(directory, "overlay.json")

	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "hex_oracle.go"), filepath.Join(root, "internal/lint/ecmascript/text/adamic_slot03_batch3.go"): filepath.Join(here, "text_export.go")}})
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
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch3Mutants(t *testing.T) {
	cases, want := batch3Oracle(t)
	for _, mutant := range []struct{ file, old, new string }{{"hex_value.a", "character <= 70", "character <= 69"}} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"hex_value.a", "hex_main.a"} {
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
