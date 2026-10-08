package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestJsxHelpersAndMutants(t *testing.T) {
	scratch := t.TempDir()
	t.Log(string(run(t, "", "python3", "jsx/testdata/capture.py", scratch)))
	root, _ := filepath.Abs("../../../../cohere")
	side, _ := filepath.Abs("jsx/testdata/oracle.go")
	virtual := filepath.Join(root, "adamic_jsx_helpers.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	path := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(scratch, "oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", oracle, virtual)
	cases := filepath.Join(scratch, "cases.json")
	expected := filepath.Join(scratch, "want.txt")
	t.Log(string(run(t, "", oracle, filepath.Join(scratch, "sources.json"), cases, expected)))
	want, err := os.ReadFile(expected)
	if err != nil {
		t.Fatal(err)
	}
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	check := func(t *testing.T, entry string, mutant bool) {
		t.Helper()
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(t.TempDir(), "program")
		if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		emitted := filepath.Join(t.TempDir(), "program.mjs")
		if err = os.WriteFile(emitted, []byte(javascript.JavaScript(ir)), 0644); err != nil {
			t.Fatal(err)
		}
		for _, backend := range []struct {
			name, exe string
			args      []string
		}{
			{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, cases}},
			{"emitted JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", runner, emitted, cases}},
			{"sanitized native", binary, []string{cases}},
		} {
			got := run(t, "", backend.exe, backend.args...)
			if mutant {
				if bytes.Equal(got, want) {
					t.Fatalf("mutant survived %s", backend.name)
				}
				t.Logf("output mutant caught on %s", backend.name)
			} else {
				compare(t, got, want)
			}
		}
	}
	entry, _ := filepath.Abs("jsx/main.a")
	check(t, entry, false)
	t.Logf("%d Go output rows byte-identical on Node, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
	mutants := []struct{ file, old, new string }{
		{"attribute_name.a", "return { name: name.text, named: true };", "return { name: name.text, named: false };"},
		{"element_parts.a", "node.kind === 'JsxSelfClosingElement'", "node.kind === 'JsxClosingElement'"},
		{"has_attribute_named.a", "name.named && matches", "!name.named && matches"},
		{"intrinsic_element_named.a", "kind === 'Identifier'", "kind !== 'StringLiteral'"},
		{"match_exactly.a", "candidate === wanted", "candidate.toLowerCase() === wanted.toLowerCase()"},
		{"match_ignoring_case.a", "8490, 75,", "8490, 8490,"},
	}
	for _, m := range mutants {
		t.Run(m.file, func(t *testing.T) {
			// The parent corpus and want are read-only; sources and outputs belong to this case.
			t.Parallel()
			dir := t.TempDir()
			files, err := filepath.Glob("jsx/*.a")
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				s := string(data)
				if filepath.Base(file) == m.file {
					if strings.Count(s, m.old) != 1 {
						t.Fatal("mutant anchor drift")
					}
					s = strings.Replace(s, m.old, m.new, 1)
				}
				if filepath.Base(file) == "main.a" {
					options, _ := filepath.Abs("options_json.ts")
					s = strings.Replace(s, "../options_json.ts", filepath.ToSlash(options), 1)
				}
				if err = os.WriteFile(filepath.Join(dir, filepath.Base(file)), []byte(s), 0644); err != nil {
					t.Fatal(err)
				}
			}
			check(t, filepath.Join(dir, "main.a"), true)
		})
	}
}

// Regenerate Go's complete Unicode simple-fold table, not just the consumer names.
func TestJsxFoldingTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "folds.json")
	generator, _ := filepath.Abs("jsx/testdata/folds.go")
	run(t, "", "go", "run", generator, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var folds struct {
		Version string
		Pairs   []int
	}
	if err = json.Unmarshal(data, &folds); err != nil {
		t.Fatal(err)
	}
	var table strings.Builder
	fmt.Fprintf(&table, "// BEGIN GENERATED SIMPLE FOLD TABLE\n// Go Unicode %s. Regenerate with testdata/regenerate_folds.py.\nconst foldPairs: readonly number[] = [\n", folds.Version)
	for i := 0; i < len(folds.Pairs); i += 2 {
		fmt.Fprintf(&table, "    %d, %d,\n", folds.Pairs[i], folds.Pairs[i+1])
	}
	table.WriteString("];\n")
	source, err := os.ReadFile("jsx/match_ignoring_case.a")
	if err != nil {
		t.Fatal(err)
	}
	marker := strings.Index(string(source), "// BEGIN GENERATED SIMPLE FOLD TABLE")
	if marker < 0 {
		t.Fatal("table marker missing")
	}
	compare(t, source[marker:], []byte(table.String()))
	t.Logf("all Unicode scalars verified against Go %s; %d noncanonical mappings", folds.Version, len(folds.Pairs)/2)
}
