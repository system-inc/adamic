package decisions

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
)

func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s: %v %s", name, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func outputs(t *testing.T, root, probe string) [][]byte {
	t.Helper()
	program, err := load.Load([]string{probe})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "probe")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(directory, "probe.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	return [][]byte{run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, probe), run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted), run(t, "", binary)}
}

// Not parallel: native builds are sequential to bound compiler memory.
func TestResolvedImportPaths(t *testing.T) {
	root, e := filepath.Abs("../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	owned, e := filepath.Abs(".")
	if e != nil {
		t.Fatal(e)
	}
	cohere := filepath.Join(root, "cohere")
	virtual := filepath.Join(cohere, "adamic_wave12_paths.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "paths-oracle.go.txt")}})
	dir := t.TempDir()
	op := filepath.Join(dir, "overlay.json")
	os.WriteFile(op, overlay, 0644)
	oracle := filepath.Join(dir, "oracle")
	run(t, cohere, "go", "build", "-overlay="+op, "-o", oracle, virtual)
	want := run(t, "", oracle, filepath.Join(owned, "paths-cases.json"))
	for side, got := range outputs(t, root, filepath.Join(owned, "paths-probe.a")) {
		if !bytes.Equal(got, want) {
			os.WriteFile("/tmp/wave12-path-got.txt", got, 0644)
			os.WriteFile("/tmp/wave12-path-want.txt", want, 0644)
			t.Fatalf("path resolver differs on side %d", side)
		}
	}
	t.Logf("399 path resolutions agree across Go/Node/emitted JS/sanitized native, %d bytes", len(want))
	directory := t.TempDir()
	for _, name := range []string{"paths.a", "paths-probe.a"} {
		data, e := os.ReadFile(filepath.Join(owned, name))
		if e != nil {
			t.Fatal(e)
		}
		if name == "paths.a" {
			from := "reduced.pop();"
			if strings.Count(string(data), from) != 1 {
				t.Fatal("path mutant anchor")
			}
			data = []byte(strings.Replace(string(data), from, `reduced.push("..");`, 1))
		}
		os.WriteFile(filepath.Join(directory, name), data, 0644)
	}
	for side, got := range outputs(t, root, filepath.Join(directory, "paths-probe.a")) {
		if bytes.Equal(got, want) {
			t.Fatalf("path mutant survived side %d", side)
		}
		t.Logf("path mutant compiles and runs cleanly; caught only by comparison side %d", side)
	}
}
