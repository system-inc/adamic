package property

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

func build(t *testing.T, directory string) artifacts {
	return buildEntry(t, directory, "testdata/tagged/main.a")
}
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
		{"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "testdata/tagged/main.a"), path)},
		{"sanitized native", run(t, "", built.native, path)},
		{"emitted JavaScript", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, path)},
	}
}

func TestNameTaggedConsumerAgreementAndMutant(t *testing.T) {
	capture, _ := filepath.Abs("testdata/tagged/capture.py")
	t.Logf("live Go capture: %s", run(t, "", "python3", capture))
	data, err := os.ReadFile("testdata/tagged/calls.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ Result string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	for _, row := range rows {
		want.WriteString(row.Result)
		want.WriteByte('\n')
	}
	path, _ := filepath.Abs("testdata/tagged/calls.json")
	directory, _ := filepath.Abs(".")
	for _, side := range observations(t, directory, path) {
		compare(t, side.output, []byte(want.String()))
		t.Logf("%s identical to Go: %d calls, %d bytes", side.name, len(rows), want.Len())
	}
	root := t.TempDir()
	mutated := filepath.Join(root, "property")
	for _, file := range []string{"property/name_tagged.a", "property/testdata/tagged/main.a", "property_name.a", "slot02_ast.a", "options_json.ts"} {
		contents, err := os.ReadFile(filepath.Join("..", file))
		if err != nil {
			t.Fatal(err)
		}
		if file == "property/name_tagged.a" {
			const anchor = "? 'number:' : 'string:'"
			if strings.Count(string(contents), anchor) != 1 {
				t.Fatal("mutant anchor changed")
			}
			contents = []byte(strings.Replace(string(contents), anchor, "? 'string:' : 'string:'", 1))
		}
		target := filepath.Join(root, file)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		write(t, target, contents)
	}
	for _, side := range observations(t, mutated, path) {
		if bytes.Equal(side.output, []byte(want.String())) {
			t.Fatalf("NameTagged mutant survived on %s", side.name)
		}
		t.Logf("NameTagged number-tag mutant caught on %s: compiled, ran and disagreed with Go", side.name)
	}
}
