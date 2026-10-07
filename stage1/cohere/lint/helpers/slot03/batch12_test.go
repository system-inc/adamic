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

func batch12Oracle(t *testing.T) (string, []byte) {
	t.Helper()
	verifyCoverageAt(t, "batch12/testdata", []string{"/collapse.isAngle", "/collapse.isNumber", "/collapse.isPercentage"})
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("batch12/testdata")
	directory := t.TempDir()
	virtual := filepath.Join(root, "adamic_slot03_batch12.go")
	overlay := filepath.Join(directory, "overlay.json")

	replacements := map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot03_batch12.go"): filepath.Join(here, "collapse_export.go")}
	for _, item := range []struct {
		file    string
		anchors map[string]string
	}{
		{"data_type.go", map[string]string{"func scanNumber(": "func adamicRawScanNumber(", "func numberWithSuffix(": "func adamicRawSuffix("}},
		{"segment.go", map[string]string{"func hasMathFunction(": "func adamicRawMath("}},
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
func TestBatch12Helpers(t *testing.T) {
	cases, want := batch12Oracle(t)
	entry, _ := filepath.Abs("batch12/main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases), command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, batch12JavaScript(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d lines match real Go, Node source, emitted JavaScript and sanitized native", bytes.Count(want, []byte{'\n'}))
}

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch12Mutants(t *testing.T) {
	cases, want := batch12Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"is_angle.a", "'grad', 'turn'", "'grad'", ""},
		{"is_angle.a", "return numberWithSuffix(value,", "return !numberWithSuffix(value,", ""},
		{"is_number.a", " && consumed > 0", "", ""},
		{"is_number.a", "consumed === value.length && consumed > 0", "consumed > 0", ""},
		{"is_number.a", "const consumed = dependencies.scanNumber(value);", "dependencies.hasMathFunction(value); const consumed = dependencies.scanNumber(value);", ""},
		{"is_percentage.a", "['%']", "['percent']", ""},
		{"is_percentage.a", " || dependencies.hasMathFunction(value)", " && dependencies.hasMathFunction(value)", ""},
		{"is_percentage.a", "return dependencies.numberWithSuffix(value,", "dependencies.hasMathFunction(value); return dependencies.numberWithSuffix(value,", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"is_angle.a", "is_number.a", "is_percentage.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch12", file))
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

// Not parallel: compiling semantic variants run serially with sanitizer observations.
func TestBatch12ArgumentMutants(t *testing.T) {
	cases, want := batch12Oracle(t)
	for _, mutant := range []struct{ file, old, new, prefix string }{
		{"is_angle.a", "numberWithSuffix(value,", "numberWithSuffix(value + 'x',", ""},
		{"is_number.a", "dependencies.scanNumber(value)", "dependencies.scanNumber(value + 'x')", ""},
		{"is_number.a", "dependencies.hasMathFunction(value)", "dependencies.hasMathFunction(value + 'x')", ""},
	} {
		t.Run(mutant.file, func(t *testing.T) {
			scratch := t.TempDir()
			for _, file := range []string{"is_angle.a", "is_number.a", "is_percentage.a", "main.a"} {
				data, err := os.ReadFile(filepath.Join("batch12", file))
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

func batch12JavaScript(t *testing.T, entry string) string {
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
