package classmembers

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

func TestConsumerCallsAgree(t *testing.T) {
	script, _ := filepath.Abs("testdata/capture.py")
	t.Logf("fresh Go capture: %s", run(t, "", "python3", script))
	data, err := os.ReadFile("testdata/consumer-calls.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Symbol string
		Result any
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.Symbol]++
		encoded, err := json.Marshal(row.Result)
		if text, ok := row.Result.(string); ok {
			encoded = []byte(text)
		}
		if err != nil {
			t.Fatal(err)
		}
		want.Write(encoded)
		want.WriteByte('\n')
	}
	path, _ := filepath.Abs("testdata/consumer-calls.json")
	directory, _ := filepath.Abs(".")
	for _, side := range observations(t, directory, path) {
		compare(t, side.output, []byte(want.String()))
		t.Logf("%s identical to Go: %d calls, %d bytes", side.name, len(rows), want.Len())
	}
	t.Logf("per-helper calls: %v", counts)
}
func TestHelperMutants(t *testing.T) {
	data, err := os.ReadFile("testdata/consumer-calls.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Result any }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	for _, row := range rows {
		encoded, _ := json.Marshal(row.Result)
		if text, ok := row.Result.(string); ok {
			encoded = []byte(text)
		}
		want.Write(encoded)
		want.WriteByte('\n')
	}
	path, _ := filepath.Abs("testdata/consumer-calls.json")
	mutations := []struct{ file, old, new string }{
		{"member_name.a", "case 'MethodDeclaration':", "case 'Constructor':"},
		{"is_overload_signature.a", "return !hasBody;", "return hasBody;"},
		{"is_accessor_kind.a", "kind === 'SetAccessor'", "kind === 'MethodDeclaration'"},
		{"key_of.a", "isStatic: member.isStatic", "isStatic: false"},
		{"for_each_duplicate.a", "member.kind !== previous", "member.kind === previous"},
	}
	for _, m := range mutations {
		t.Run(m.file, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "classmembers")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"main.a", "member_name.a", "is_overload_signature.a", "is_accessor_kind.a", "member.a", "key_of.a", "for_each_duplicate.a"} {
				contents, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if file == m.file {
					if strings.Count(string(contents), m.old) != 1 {
						t.Fatal("mutation anchor changed")
					}
					contents = []byte(strings.Replace(string(contents), m.old, m.new, 1))
				}
				write(t, filepath.Join(directory, file), contents)
			}
			for _, file := range []string{"options_json.ts", "slot02_ast.a", "property_name.a", "property/name_tagged.a"} {
				data, err := os.ReadFile(filepath.Join("..", file))
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(root, file)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				write(t, path, data)
			}
			for _, side := range observations(t, directory, path) {
				if bytes.Equal(side.output, []byte(want.String())) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("%s caught on %s: compiled and ran, output differs from Go", m.file, side.name)
			}
		})
	}
}
