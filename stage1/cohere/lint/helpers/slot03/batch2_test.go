package slot03

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func batch2Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch2/testdata", []string{"/jsx.IsIntrinsicElementNamed", "/tailwind.holeEdges", "/tailwind.*ClassLiteralReader.readClassValues"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch2/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch2.go")
	overlay := filepath.Join(directory, "overlay.json")
	original := filepath.Join(root, "internal/lint/rules/tailwind/class_literals.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	modified := string(source)
	for _, surface := range []struct{ method, route string }{{"attributeValues", "Attribute"}, {"calleeValues", "Callee"}, {"variableValues", "Variable"}} {
		anchor := "return r." + surface.method + "(node)"
		if strings.Count(modified, anchor) != 1 {
			t.Fatal("Go dispatcher trace anchor changed")
		}
		modified = strings.Replace(modified, anchor, "adamicRoute += \""+surface.route+"\"; "+anchor, 1)
	}
	traced := filepath.Join(directory, "class_literals.go")
	if err = os.WriteFile(traced, []byte(modified), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: traced, virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot03_batch2.go"): filepath.Join(here, "tailwind_export.go")}})
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
func TestBatch2Helpers(t *testing.T) {
	cases, want := batch2Oracle(t)
	entry, _ := filepath.Abs("batch2/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch2Mutants(t *testing.T) {
	cases, want := batch2Oracle(t)
	for _, mutant := range []struct{ file, old, new string }{{"intrinsic_element_named.a", "kind === 'Identifier'", "kind !== 'StringLiteral'"}, {"hole_edges.a", "leading = template.leading;", "leading = false;"}, {"read_class_values.a", "return calleeValues(node);", "return attributeValues(node);"}} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"intrinsic_element_named.a", "hole_edges.a", "read_class_values.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch2", file))
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
				if file == "hole_edges.a" {
					dependency, _ := filepath.Abs("tailwind_space.a")
					text = strings.ReplaceAll(text, "../tailwind_space.a", filepath.ToSlash(dependency))
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
