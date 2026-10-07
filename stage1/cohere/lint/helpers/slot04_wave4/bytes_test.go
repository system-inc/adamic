package bytehelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
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
	virtual := filepath.Join(root, "adamic_slot04_wave4_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot04_wave4_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
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

type artifacts struct{ native, script string }

func build(t *testing.T, directory string) artifacts { return buildEntry(t, directory, "main.a") }
func buildEntry(t *testing.T, directory, entry string) artifacts {
	t.Helper()
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
	script := filepath.Join(t.TempDir(), "emitted.mjs")
	write(t, script, []byte(javascript.JavaScript(ir)))
	return artifacts{binary, script}
}
func observations(t *testing.T, directory, path string) []struct {
	name   string
	output []byte
} {
	t.Helper()
	built := build(t, directory)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	return []struct {
		name   string
		output []byte
	}{
		{"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.a"), path)},
		{"sanitized native", run(t, "", built.native, path)},
		{"emitted JavaScript", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, path)},
	}
}

func TestBytesGoNodeNativeJavaScript(t *testing.T) {
	goOracle := oracle(t)
	built := build(t, ".")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.a")
	for _, fixture := range []string{"witnesses.json", "consumers.json", "--full"} {
		t.Run(fixture, func(t *testing.T) {
			path := fixture
			var adapted string
			if fixture == "--full" {
				adapted = "--full"
			} else {
				path, _ = filepath.Abs("testdata/" + fixture)
				adapted = filepath.Join(t.TempDir(), "cases.json")
				write(t, adapted, run(t, "", goOracle, "--cases", path))
			}
			want := run(t, "", goOracle, path)
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, adapted), want)
			compare(t, run(t, "", built.native, adapted), want)
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, adapted), want)
			t.Logf("Go, source Node, sanitized native and emitted JavaScript match %d output lines", bytes.Count(want, []byte("\n")))
		})
	}
}
func TestBytesCompilingMutants(t *testing.T) {
	goOracle := oracle(t)
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", goOracle, path)
	adapted := filepath.Join(t.TempDir(), "cases.json")
	write(t, adapted, run(t, "", goOracle, "--cases", path))
	mutations := []struct{ file, old, new string }{
		{"top_of_stack.a", "stack[stack.length - 1]", "stack[0]"},
		{"peek_byte.a", "return input[index]", "return input[0]"},
		{"peek_byte.a", "return 0;", "return 1;"},
	}
	for _, mutation := range mutations {
		t.Run(mutation.file+mutation.old, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave4")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"main.a", "top_of_stack.a", "peek_byte.a"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == mutation.file {
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
			for _, observation := range observations(t, directory, adapted) {
				if bytes.Equal(observation.output, want) {
					t.Fatalf("%s compiling mutant survived", observation.name)
				}
				a, b := strings.Split(string(observation.output), "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Logf("%s compiling mutant caught at line %d: got %q, Go %q", observation.name, i+1, a[i], b[i])
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
			if strings.HasSuffix(h, ".sortedKeys") || strings.HasSuffix(h, ".topOfStack") || strings.HasSuffix(h, ".peekByte") {
				if !present[r.Rule] {
					missing = append(missing, r.Rule)
				}
				break
			}
		}
	}
	return missing
}
func TestBytesConsumerCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	if missing := missingConsumers(t, data); len(missing) > 0 {
		t.Fatalf("missing %v", missing)
	}
	var rows []struct{ Name, Source string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	omittedConsumer, _, _ := strings.Cut(rows[0].Name, ":")
	filtered := []struct{ Name, Source string }{}
	for _, row := range rows {
		if !strings.HasPrefix(row.Name, omittedConsumer+":") {
			filtered = append(filtered, row)
		}
	}
	omitted, _ := json.Marshal(filtered)
	if len(missingConsumers(t, omitted)) == 0 {
		t.Fatal("consumer omission survived")
	}
	t.Logf("all six consumers covered; omission caught: %s", omittedConsumer)
}

func sortOracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	source, _ := filepath.Abs("testdata/sort_oracle.go")
	exports, _ := filepath.Abs("testdata/sort_exports.go")
	virtual := filepath.Join(root, "adamic_slot04_wave4_sort_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot04_sort_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func sortObservations(t *testing.T, directory, path string) []struct {
	name   string
	output []byte
} {
	t.Helper()
	built := buildEntry(t, directory, "sort_main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	return []struct {
		name   string
		output []byte
	}{
		{"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "sort_main.a"), path)},
		{"sanitized native", run(t, "", built.native, path)},
		{"emitted JavaScript", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, path)},
	}
}
func TestSortedKeys(t *testing.T) {
	oracle := sortOracle(t)
	for _, fixture := range []string{"sort-witnesses.json", "sort-consumers.json"} {
		t.Run(fixture, func(t *testing.T) {
			path, _ := filepath.Abs("testdata/" + fixture)
			want := run(t, "", oracle, path)
			for _, got := range sortObservations(t, ".", path) {
				compare(t, got.output, want)
			}
			t.Logf("Go, source Node, sanitized native and emitted JavaScript match %d map-state lines", bytes.Count(want, []byte("\n")))
		})
	}
}
func TestSortedKeysCompilingMutants(t *testing.T) {
	path, _ := filepath.Abs("testdata/sort-witnesses.json")
	want := run(t, "", sortOracle(t), path)
	for _, name := range []string{"omit false values", "reverse UTF8", "UTF16 comparator"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave4")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"sort_main.a", "sorted_keys.a"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				s := string(data)
				if file == "sorted_keys.a" {
					switch name {
					case "omit false values":
						s = strings.Replace(s, "keys.push(key);", "if(set.get(key) === true) { keys.push(key); }", 1)
					case "reverse UTF8":
						s = strings.Replace(s, "utf8At(left,index) - utf8At(right,index)", "utf8At(right,index) - utf8At(left,index)", 1)
					case "UTF16 comparator":
						s = strings.ReplaceAll(s, "utf8Length(left)", "left.length")
						s = strings.ReplaceAll(s, "utf8Length(right)", "right.length")
						s = strings.ReplaceAll(s, "utf8At(left,index)", "left.charCodeAt(index)")
						s = strings.ReplaceAll(s, "utf8At(right,index)", "right.charCodeAt(index)")
					}
					if s == string(data) {
						t.Fatal("mutant anchor changed")
					}
				}
				write(t, filepath.Join(directory, file), []byte(s))
			}
			options, err := os.ReadFile("../options_json.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "options_json.ts"), options)
			for _, got := range sortObservations(t, directory, path) {
				if bytes.Equal(got.output, want) {
					t.Fatalf("%s compiling mutant survived", got.name)
				}
				a, b := strings.Split(string(got.output), "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Logf("%s compiling mutant caught at line %d: got %q, Go %q", got.name, i+1, a[i], b[i])
						break
					}
				}
			}
		})
	}
}
