package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func schemaCases(t *testing.T) (string, int, int) {
	path, files, count := lexCases(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var extra strings.Builder
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	values := []string{"", "~", "null", "NULL", "Null", "NuLl", "true", "True", "TRUE", "false", "yes", "Yes", "YES", "on", "OFF", "n", "00", "07", "08", "0b101", "0o17", "0xFF", "1_000", "+1", "-0", ".0", "1.", ".", "1e3", "1_e+3", ".Inf", "-.inf", ".nan", ".NAN", "12:34", "12:34:56.70", "2000-01-01", "2000-1-1T0:0:0Z", "0000-99-99 24:59:59.123+2:30", "<<", "2000-01-01x"}
	for _, value := range values {
		for _, before := range []string{"", "+", "-", "x"} {
			for _, after := range []string{"", "x", "_", ".0"} {
				extra.WriteString("0\t" + escape.Replace(before+value+after) + "\n")
				count++
			}
		}
	}
	target := filepath.Join(t.TempDir(), "schema-cases.txt")
	if err := os.WriteFile(target, append(data, []byte(extra.String())...), 0644); err != nil {
		t.Fatal(err)
	}
	return target, files, count
}

func goSchema(t *testing.T, cases string) []byte {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/schema_go.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	goBinary := filepath.Join(t.TempDir(), "go-schema")
	run(t, filepath.Join(root, "cohere"), nil, "go", "build", "-overlay", path, "-o", goBinary, "./command/formatter_comparison")
	return run(t, "", nil, goBinary, cases)
}

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestSchemaMatchesGo(t *testing.T) {
	cases, files, count := schemaCases(t)
	expected := goSchema(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("schema_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "schema")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	nativeOut := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, binary, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
	emitted := filepath.Join(t.TempDir(), "schema.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	backend := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases)
	for _, side := range []struct {
		name string
		out  []byte
	}{{"native ASan/UBSan/LSan", nativeOut}, {"Node", node}, {"emitted JavaScript", backend}} {
		if !bytes.Equal(side.out, expected) {
			t.Fatalf("%s: %s", side.name, firstDifference(side.out, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to an npm install of yaml@2.9.0 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	external := run(t, "", nil, "node", "testdata/schema_library.mjs", library, cases)
	if !bytes.Equal(external, expected) {
		t.Fatal(firstDifference(external, expected))
	}
	t.Logf("%d repository files, %d cases, %d Schema answer bytes identical on all five sides", files, count, len(expected))
}

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestSchemaMutants(t *testing.T) {
	cases, _, _ := schemaCases(t)
	expected := goSchema(t, cases)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutant := range []struct{ name, from, to string }{
		{"class range loses last character", "code <= last.charCodeAt(0)", "code < last.charCodeAt(0)"},
		{"plus accepts empty", "minimum = quantifier === '+' ? 1 : 0;", "minimum = quantifier === '+' ? 0 : 0;"},
		{"optional accepts twice", "maximum = quantifier === '?' ? 1 : -1;", "maximum = quantifier === '?' ? 2 : -1;"},
		{"nullable prefix loses successors", "node.kind === 'concat' && !this.match(child, '', 0).includes(0)", "node.kind === 'concat'"},
	} {
		// Not parallel: native.Build writes the shared adamic/runtime or adamic/units cache.
		t.Run(mutant.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, file := range []string{"patternNode.ts", "schemaPattern.ts", "schemaTag.ts", "schemaTags.ts", "schema_main.ts"} {
				source, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == "schemaPattern.ts" {
					if strings.Count(string(source), mutant.from) != 1 {
						t.Fatal("mutation site must occur once")
					}
					source = []byte(strings.Replace(string(source), mutant.from, mutant.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), source, 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "schema_main.ts")
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(lowered), binary, native.Options{}); err != nil {
				t.Fatal(err)
			}
			nativeOut := run(t, "", nil, binary, cases)
			node := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, cases)
			for _, side := range []struct {
				name string
				out  []byte
			}{{"native", nativeOut}, {"Node", node}} {
				if bytes.Equal(side.out, expected) {
					t.Fatalf("%s missed mutant", side.name)
				}
				t.Logf("%s successful execution, wrong bytes caught: %s", side.name, firstDifference(side.out, expected))
			}
		})
	}
}
