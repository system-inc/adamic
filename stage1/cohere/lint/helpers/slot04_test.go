package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func slot04Oracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	source, _ := filepath.Abs("testdata/slot04/oracle.go")
	virtual := filepath.Join(root, "adamic_slot04_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func slot04Build(t *testing.T, directory string) string {
	return slot04BuildEntry(t, directory, "slot04_main.a")
}
func slot04BuildEntry(t *testing.T, directory, entry string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, entry)})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slot04")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
func slot04Cases(t *testing.T, fixture string) (string, []byte) {
	t.Helper()
	path, _ := filepath.Abs("testdata/slot04/" + fixture)
	oracle := slot04Oracle(t)
	adapted := run(t, "", oracle, "--ast", path)
	out := filepath.Join(t.TempDir(), "adapted.json")
	if err := os.WriteFile(out, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	return out, run(t, "", oracle, path)
}
func TestSlot04ElementParts(t *testing.T) {
	path, want := slot04Cases(t, "consumers.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("slot04_main.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", slot04Build(t, "."), path), want)
	t.Logf("Go, Node source and sanitized native match %d output lines", bytes.Count(want, []byte("\n")))
}

// Not parallel: limit simultaneous compiler and clang memory use.
func TestSlot04ElementPartsMutant(t *testing.T) {
	path, want := slot04Cases(t, "witnesses.json")
	directory := t.TempDir()
	for _, file := range []string{"slot04_main.a", "jsx_ast.a", "jsx_element_parts.a", "options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == "jsx_element_parts.a" {
			anchor := " || node.kind === 'JsxSelfClosingElement'"
			if strings.Count(text, anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, anchor, "", 1)
		}
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	slot04RunMutant(t, directory, "slot04_main.a", path, want)
}

func TestSlot04ElementShapes(t *testing.T) {
	path, want := slot04Cases(t, "witnesses.json")
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("slot04_main.a")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", slot04Build(t, "."), path), want)
	t.Logf("Go, Node and sanitized native match %d shape output lines", bytes.Count(want, []byte("\n")))
}

func slot04SettingsOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	source, _ := filepath.Abs("testdata/slot04/settings_oracle.go")
	exports, _ := filepath.Abs("testdata/slot04/tailwind_exports.go")
	virtual := filepath.Join(root, "adamic_slot04_settings_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot04_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func TestSlot04SettingsKey(t *testing.T) {
	oracle := slot04SettingsOracle(t)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("tailwind04_main.a")
	binary := slot04BuildEntry(t, ".", "tailwind04_main.a")
	for _, fixture := range []string{"settings-consumers.json", "settings-witnesses.json"} {
		path, _ := filepath.Abs("testdata/slot04/" + fixture)
		want := run(t, "", oracle, path)
		compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
		compare(t, run(t, "", binary, path), want)
		t.Logf("%s: Go, Node source and sanitized native match %d settings keys", fixture, bytes.Count(want, []byte("\n")))
	}
}

// Not parallel: limit simultaneous compiler and clang memory use.
func TestSlot04SettingsKeyMutant(t *testing.T) {
	path, _ := filepath.Abs("testdata/slot04/settings-witnesses.json")
	want := run(t, "", slot04SettingsOracle(t), path)
	directory := t.TempDir()
	for _, file := range []string{"tailwind04_main.a", "tailwind_settings04.a", "tailwind_settings_key.a", "options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == "tailwind_settings_key.a" {
			anchor := "key += '\\u0001';"
			if strings.Count(text, anchor) != 2 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, anchor, "key += '';", 1)
		}
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	slot04RunMutant(t, directory, "tailwind04_main.a", path, want)
}

func slot04ReaderOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	source, _ := filepath.Abs("testdata/slot04/reader_oracle.go")
	exports, _ := filepath.Abs("testdata/slot04/tailwind_exports.go")
	virtual := filepath.Join(root, "adamic_slot04_reader_oracle.go")
	factory := filepath.Join(root, "internal/lint/rules/tailwind/class_literals.go")
	data, err := os.ReadFile(factory)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "return reader"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("reader factory capture anchor changed")
	}
	changed := strings.Replace(string(data), anchor, "AdamicRecordCreated(reader,settings)\n"+anchor, 1)
	side := filepath.Join(t.TempDir(), "class_literals.go")
	if err := os.WriteFile(side, []byte(changed), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/adamic_slot04_exports.go"): exports, factory: side}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func TestSlot04CompiledReader(t *testing.T) {
	oracle := slot04ReaderOracle(t)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("reader04_main.a")
	binary := slot04BuildEntry(t, ".", "reader04_main.a")
	for _, fixture := range []string{"reader-consumers.json", "reader-witnesses.json"} {
		path, _ := filepath.Abs("testdata/slot04/" + fixture)
		want := run(t, "", oracle, path)
		compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
		compare(t, run(t, "", binary, path), want)
		t.Logf("%s: Go, Node source and sanitized native match %d cache calls", fixture, bytes.Count(want, []byte("\n")))
	}
}

// Not parallel: limit simultaneous compiler and clang memory use.
func TestSlot04CompiledReaderMutant(t *testing.T) {
	path, _ := filepath.Abs("testdata/slot04/reader-witnesses.json")
	want := run(t, "", slot04ReaderOracle(t), path)
	directory := t.TempDir()
	for _, file := range []string{"reader04_main.a", "tailwind_compiled_reader.a", "tailwind_settings04.a", "tailwind_settings_key.a", "options_json.ts"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == "tailwind_compiled_reader.a" {
			anchor := "return previous;"
			if strings.Count(text, anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, anchor, "return create(settings);", 1)
		}
		if err := os.WriteFile(filepath.Join(directory, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	slot04RunMutant(t, directory, "reader04_main.a", path, want)
}

func slot04RunMutant(t *testing.T, directory, entry, path string, want []byte) {
	t.Helper()
	binary := slot04BuildEntry(t, directory, entry)
	runner, _ := filepath.Abs("../../../../oracle/node.mjs")
	source := filepath.Join(directory, entry)
	for _, mode := range []string{"sanitized native", "Node source"} {
		var got []byte
		if mode == "sanitized native" {
			got = run(t, "", binary, path)
		} else {
			got = run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, source, path)
		}
		if bytes.Equal(got, want) {
			t.Fatalf("%s semantic mutant survived", mode)
		}
		expected := strings.Split(string(want), "\n")
		for i, line := range strings.Split(string(got), "\n") {
			if i < len(expected) && line != expected[i] {
				t.Logf("%s compiling mutant caught at line %d: got %q, Go %q", mode, i+1, line, expected[i])
				break
			}
		}
	}
}
func slot04MissingConsumers(t *testing.T, fixture []byte, symbols []string) []string {
	t.Helper()
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Name string }
	if err := json.Unmarshal(fixture, &rows); err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, row := range rows {
		name, _, _ := strings.Cut(row.Name, ":")
		present[name] = true
	}
	missing := []string{}
	for _, row := range ledger.Remaining {
		for _, helper := range row.Helpers {
			for _, symbol := range symbols {
				if helper == symbol && !present[row.Rule] {
					missing = append(missing, row.Rule)
				}
			}
		}
	}
	return missing
}
func TestSlot04ConsumerCoverage(t *testing.T) {
	for _, check := range []struct {
		file    string
		symbols []string
		count   int
	}{
		{"consumers.json", []string{"github.com/system-inc/cohere/internal/lint/ecmascript/jsx.ElementParts"}, 24},
		{"tailwind-consumers.json", []string{"github.com/system-inc/cohere/internal/lint/rules/tailwind.ClassLiteralSettings.key", "github.com/system-inc/cohere/internal/lint/rules/tailwind.compiledClassLiteralReader"}, 12},
	} {
		data, err := os.ReadFile("testdata/slot04/" + check.file)
		if err != nil {
			t.Fatal(err)
		}
		if missing := slot04MissingConsumers(t, data, check.symbols); len(missing) != 0 {
			t.Fatalf("missing consumers: %v", missing)
		}
		var rows []map[string]any
		if err := json.Unmarshal(data, &rows); err != nil {
			t.Fatal(err)
		}
		first := rows[0]["name"].(string)
		removed, _, _ := strings.Cut(first, ":")
		mutant := []map[string]any{}
		for _, row := range rows {
			name, _, _ := strings.Cut(row["name"].(string), ":")
			if name != removed {
				mutant = append(mutant, row)
			}
		}
		changed, err := json.Marshal(mutant)
		if err != nil {
			t.Fatal(err)
		}
		if len(slot04MissingConsumers(t, changed, check.symbols)) == 0 {
			t.Fatal("consumer omission mutant survived")
		}
		t.Logf("%s covers all %d consumers; removing %s is caught", check.file, check.count, removed)
	}
}
