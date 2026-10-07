package wave2

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: these comparisons compile whole profiles and hold large output streams.
func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = directory
	log, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	source, _ := filepath.Abs("testdata/oracle.go")
	exports, _ := filepath.Abs("testdata/exports.go")
	virtual := filepath.Join(root, "adamic_slot04_wave2_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot04_wave2_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func build(t *testing.T, directory string) string { return buildEntry(t, directory, "main.a") }
func buildEntry(t *testing.T, directory, entry string) string {
	program, err := load.Load([]string{filepath.Join(directory, entry)})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size %d != %d", len(got), len(want))
}
func TestWave2GoNodeNative(t *testing.T) {
	goOracle := oracle(t)
	binary := build(t, ".")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.a")
	for _, fixture := range []string{"witnesses.json", "consumers.json"} {
		t.Run(fixture, func(t *testing.T) {
			path, _ := filepath.Abs("testdata/" + fixture)
			want := run(t, "", goOracle, path)
			adapted := filepath.Join(t.TempDir(), "ast.json")
			write(t, adapted, run(t, "", goOracle, "--ast", path))
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, adapted), want)
			compare(t, run(t, "", binary, adapted), want)
			t.Logf("Go, Node source and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
		})
	}
}
func TestWave2CompilingMutants(t *testing.T) {
	goOracle := oracle(t)
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", goOracle, path)
	adapted := filepath.Join(t.TempDir(), "ast.json")
	write(t, adapted, run(t, "", goOracle, "--ast", path))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	mutations := []struct{ file, old, new string }{
		{"class_values_under.a", "leading: false, trailing: false", "leading: true, trailing: false"},
		{"collect_class_values.a", "collectClassValues(arena, item.whenFalse, origin, edges, values, dependencies);", ""},
		{"collect_class_values.a", "collectClassValues(arena, item.right, origin, edges, values, dependencies);", "collectClassValues(arena, item.left, origin, edges, values, dependencies);"},
	}
	for _, mutation := range mutations {
		t.Run(mutation.file+mutation.old, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave2")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"model.a", "main.a", "class_values_under.a", "collect_class_values.a"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == mutation.file {
					if !strings.Contains(string(data), mutation.old) {
						t.Fatal("mutant anchor missing")
					}
					data = []byte(strings.Replace(string(data), mutation.old, mutation.new, 1))
				}
				write(t, filepath.Join(directory, name), data)
			}
			options, err := os.ReadFile("../options_json.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "options_json.ts"), options)
			binary := build(t, directory)
			outputs := []struct {
				name string
				data []byte
			}{{"sanitized native", run(t, "", binary, adapted)}, {"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.a"), adapted)}}
			for _, output := range outputs {
				if bytes.Equal(output.data, want) {
					t.Fatalf("%s compiling mutant survived", output.name)
				}
				a, b := strings.Split(string(output.data), "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Logf("%s compiling mutant caught at line %d: got %q, Go %q", output.name, i+1, a[i], b[i])
						break
					}
				}
			}
		})
	}
}
func missingConsumers(t *testing.T, data []byte) []string {
	t.Helper()
	var rows []struct{ Name string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, r := range rows {
		name, _, _ := strings.Cut(r.Name, ":")
		present[name] = true
	}
	ledger, err := os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(ledger, &d); err != nil {
		t.Fatal(err)
	}
	missing := []string{}
	for _, r := range d.Remaining {
		for _, h := range r.Helpers {
			if strings.HasSuffix(h, ".attributeValues") || strings.HasSuffix(h, ".classValuesUnder") || strings.HasSuffix(h, ".collectClassValues") {
				if !present[r.Rule] {
					missing = append(missing, r.Rule)
				}
				break
			}
		}
	}
	return missing
}
func TestWave2ConsumerCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	if missing := missingConsumers(t, data); len(missing) > 0 {
		t.Fatalf("missing consumers %v", missing)
	}
	var rows []struct{ Name, Source string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	filtered := rows[:0]
	for _, row := range rows {
		if !strings.HasPrefix(row.Name, "better-tailwindcss/enforce-canonical-classes:") {
			filtered = append(filtered, row)
		}
	}
	omitted, _ := json.Marshal(filtered)
	if len(missingConsumers(t, omitted)) == 0 {
		t.Fatal("omission mutant survived")
	}
	t.Log("all 11 consumers covered; canonical-class consumer omission caught")
}

func unescapeOracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	source, _ := filepath.Abs("testdata/unescape_oracle.go")
	exports, _ := filepath.Abs("testdata/text_exports.go")
	virtual := filepath.Join(root, "adamic_slot04_wave2_text_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/ecmascript/text/adamic_slot04_wave2_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func TestWave2Unescape(t *testing.T) {
	goOracle := unescapeOracle(t)
	binary := buildEntry(t, ".", "unescape_main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("unescape_main.a")
	for _, fixture := range []string{"text-witnesses.json", "text-consumers.json"} {
		t.Run(fixture, func(t *testing.T) {
			path, _ := filepath.Abs("testdata/" + fixture)
			want := run(t, "", goOracle, path)
			adapted := filepath.Join(t.TempDir(), "ast.json")
			write(t, adapted, run(t, "", goOracle, "--ast", path))
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, adapted), want)
			compare(t, run(t, "", binary, adapted), want)
			t.Logf("Go, Node source and sanitized native match %d text values", bytes.Count(want, []byte("\n")))
		})
	}
}
func TestWave2UnescapeMutants(t *testing.T) {
	goOracle := unescapeOracle(t)
	path, _ := filepath.Abs("testdata/text-witnesses.json")
	want := run(t, "", goOracle, path)
	adapted := filepath.Join(t.TempDir(), "ast.json")
	write(t, adapted, run(t, "", goOracle, "--ast", path))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, mutation := range []struct{ old, new string }{{"if(replacement.valid)", "if(true)"}, {"index = semicolon + 1;", "index = semicolon + 2;"}} {
		t.Run(mutation.old, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave2")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"unescape_main.a", "unescape_string_literal_text.a"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "unescape_string_literal_text.a" {
					if strings.Count(string(data), mutation.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), mutation.old, mutation.new, 1))
				}
				write(t, filepath.Join(directory, name), data)
			}
			options, err := os.ReadFile("../options_json.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "options_json.ts"), options)
			binary := buildEntry(t, directory, "unescape_main.a")
			outputs := []struct {
				name string
				data []byte
			}{{"sanitized native", run(t, "", binary, adapted)}, {"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "unescape_main.a"), adapted)}}
			for _, output := range outputs {
				if bytes.Equal(output.data, want) {
					t.Fatalf("%s compiling mutant survived", output.name)
				}
				a, b := strings.Split(string(output.data), "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Logf("%s compiling mutant caught at line %d: got %q, Go %q", output.name, i+1, a[i], b[i])
						break
					}
				}
			}
		})
	}
}
func TestWave2TextConsumerCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/text-consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Name, Source string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	ledger, err := os.ReadFile("../readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(ledger, &d); err != nil {
		t.Fatal(err)
	}
	missing := func(rows []struct{ Name, Source string }) []string {
		present := map[string]bool{}
		for _, row := range rows {
			name, _, _ := strings.Cut(row.Name, ":")
			present[name] = true
		}
		result := []string{}
		for _, row := range d.Remaining {
			for _, h := range row.Helpers {
				if strings.HasSuffix(h, ".UnescapeStringLiteralText") && !present[row.Rule] {
					result = append(result, row.Rule)
				}
			}
		}
		return result
	}
	if absent := missing(rows); len(absent) > 0 {
		t.Fatalf("missing text consumers %v", absent)
	}
	filtered := rows[:0]
	for _, row := range rows {
		if !strings.HasPrefix(row.Name, "@next/next/google-font-display:") {
			filtered = append(filtered, row)
		}
	}
	if len(missing(filtered)) == 0 {
		t.Fatal("text consumer omission mutant survived")
	}
	t.Log("all nine text consumers covered; google-font-display omission caught")
}
