package literalhelpers

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

func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	source, _ := filepath.Abs("testdata/oracle.go")
	exports, _ := filepath.Abs("testdata/exports.go.txt")
	virtual := filepath.Join(root, "adamic_literal_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source, filepath.Join(root, "internal/lint/ecmascript/literal/adamic_exports.go"): exports}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	write(t, path, overlay)
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func TestLiteralGoAgreement(t *testing.T) {
	oracle := oracle(t)
	built := build(t, ".")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.a")
	for _, name := range []string{"no-regex-spaces", "no-misleading-character-class", "controls"} {
		t.Run(name, func(t *testing.T) {
			path, _ := filepath.Abs("testdata/" + name + ".json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var rows []struct{ Kind, Want string }
			if err = json.Unmarshal(data, &rows); err != nil {
				t.Fatal(err)
			}
			expected := map[string]int{"no-regex-spaces": 460, "no-misleading-character-class": 1186, "controls": 82}
			if len(rows) != expected[name] {
				t.Fatalf("capture count drift: %d != %d", len(rows), expected[name])
			}
			kinds := map[string]int{}
			var recorded bytes.Buffer
			for _, r := range rows {
				kinds[r.Kind]++
				recorded.WriteString(r.Want + "\n")
			}
			if kinds["produced"] == 0 || kinds["tail"] == 0 {
				t.Fatal("missing captured helper", kinds)
			}

			want := run(t, "", oracle, path)
			compare(t, want, recorded.Bytes())
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
			compare(t, run(t, "", built.native, path), want)
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, path), want)
			t.Logf("%d Go/Node/emitted/sanitized-native helper calls agree", bytes.Count(want, []byte("\n")))
		})
	}
}
func TestLiteralMutants(t *testing.T) {
	oracle := oracle(t)
	path, _ := filepath.Abs("testdata/controls.json")
	want := run(t, "", oracle, path)
	for _, m := range []struct{ file, old, new string }{
		{"cooked_bytes_produced_by.a", "return width;", "return width + 1;"},
		{"produces_no_cooked_bytes.a", "index += 2;", "index += 3;"},
	} {
		t.Run(m.file, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "literal")
			if err := os.Mkdir(dir, 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"main.a", "cooked_bytes_produced_by.a", "produces_no_cooked_bytes.a"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == m.file {
					if strings.Count(string(data), m.old) != 1 {
						t.Fatalf("anchor %q changed", m.old)
					}
					data = []byte(strings.Replace(string(data), m.old, m.new, 1))
				}
				write(t, filepath.Join(dir, name), data)
			}
			options, err := os.ReadFile("../options_json.ts")
			if err != nil {
				t.Fatal(err)
			}
			write(t, filepath.Join(root, "options_json.ts"), options)
			for _, o := range observations(t, dir, path) {
				if bytes.Equal(o.output, want) {
					t.Fatalf("%s mutant survived", o.name)
				}
				t.Logf("%s compiling output mutant caught", o.name)
			}
		})
	}
}
