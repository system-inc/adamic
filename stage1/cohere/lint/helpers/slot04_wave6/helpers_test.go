package stringhelpers

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
	virtual := filepath.Join(root, "adamic_slot04_wave6_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_slot04_wave6_exports.go"): exports}})
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

func TestStringsGoNodeNativeJavaScript(t *testing.T) {
	goOracle := oracle(t)
	built := build(t, ".")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.a")
	for _, fixture := range []string{"witnesses.json", "consumers.json", "calls.json", "--full"} {
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
func TestStringsCompilingMutants(t *testing.T) {
	goOracle := oracle(t)
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", goOracle, path)
	adapted := filepath.Join(t.TempDir(), "cases.json")
	write(t, adapted, run(t, "", goOracle, "--cases", path))
	mutations := []struct{ file, old, new string }{
		{"theme_prefix_key.a", "let index = 2", "let index = 1"},
		{"theme_prefix_key.a", "result.push(45);", "result.push(95);"},
		{"variant_registry_has.a", "registrations.has(root)", "registrations.get(root) === true"},
		{"design_system_prefix.a", "return system.theme.prefix;", "return system.theme.prefix.slice(0,0);"},
	}

	for _, mutation := range mutations {
		t.Run(mutation.file+mutation.old, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave6")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"main.a", "theme_prefix_key.a", "variant_registry_has.a", "design_system_prefix.a", "theme_view.a"} {
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
			if strings.HasSuffix(h, "Theme.PrefixKey") || strings.HasSuffix(h, "VariantRegistry.Has") || strings.HasSuffix(h, "DesignSystem.Prefix") {
				if !present[r.Rule] {
					missing = append(missing, r.Rule)
				}
				break
			}
		}
	}
	return missing
}
func TestStringsConsumerCoverage(t *testing.T) {
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

// Refusal is a separate oracle check: no fabricated value for Go's panicking slice.
func rejected(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	log, err := os.CreateTemp(t.TempDir(), "refusal-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	command.Stdout = log
	command.Stderr = log
	if err = command.Run(); err == nil {
		t.Fatalf("%s accepted Go's short-key panic input", name)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("refusal without diagnostic")
	}
	return data
}
func TestPrefixKeyShortKeyRefusalAndMutant(t *testing.T) {
	goOracle := oracle(t)
	directory, _ := filepath.Abs(".")
	built := build(t, directory)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry := filepath.Join(directory, "main.a")
	for _, arg := range []string{"--bad0", "--bad1"} {
		if got := rejected(t, goOracle, arg); !bytes.Contains(got, []byte("slice bounds")) {
			t.Fatalf("unexpected Go refusal: %s", got)
		}
		for _, mode := range []struct {
			name string
			args []string
		}{
			{"node", []string{"--disable-warning=ExperimentalWarning", runner, entry, arg}},
			{built.native, []string{arg}},
			{"node", []string{"--disable-warning=ExperimentalWarning", runner, built.script, arg}},
		} {
			if got := rejected(t, mode.name, mode.args...); !bytes.Contains(got, []byte("prefix key needs at least two bytes")) {
				t.Fatalf("unexpected Adamic refusal: %s", got)
			}
		}
		t.Logf("Go, Node source, sanitized native and emitted JavaScript reject %s", arg)
	}
	root := t.TempDir()
	mutant := filepath.Join(root, "slot04_wave6")
	if err := os.Mkdir(mutant, 0755); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"main.a", "theme_prefix_key.a", "variant_registry_has.a", "design_system_prefix.a", "theme_view.a"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if file == "theme_prefix_key.a" {
			if strings.Count(string(data), "key.length < 2") != 1 {
				t.Fatal("mutant anchor")
			}
			data = []byte(strings.Replace(string(data), "key.length < 2", "key.length < 0", 1))
		}
		write(t, filepath.Join(mutant, file), data)
	}
	options, err := os.ReadFile("../options_json.ts")
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "options_json.ts"), options)
	for _, arg := range []string{"--bad0", "--bad1"} {
		for _, got := range observations(t, mutant, arg) {
			if string(got.output) != "45,45,116,119,45\n" {
				t.Fatalf("unexpected guard mutant output: %q", got.output)
			}
			t.Logf("%s compiling guard mutant caught: fabricated short-key result for %s", got.name, arg)
		}
	}
}
