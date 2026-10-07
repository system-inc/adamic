package slot03

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

func command(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	log, err := os.CreateTemp(t.TempDir(), "command-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	data, readErr := os.ReadFile(log.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, data)
	}
	return data
}
func oracle(t *testing.T) (string, []byte) {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("testdata")
	virtual := filepath.Join(root, "adamic_slot03_oracle.go")
	raw, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(here, "oracle.go"), filepath.Join(root, "internal/lint/ecmascript/react/adamic_slot03.go"): filepath.Join(here, "react_export.go")}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	overlay := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlay, raw, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	command(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	cases := filepath.Join(directory, "cases.json")
	want := command(t, "", binary, filepath.Join(here, "sources.jsonl.gz"), cases)
	return cases, want
}
func build(t *testing.T, entry string) string {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "helper")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}
func TestComponentBaseNameMatchesCohere(t *testing.T) {
	cases, want := oracle(t)
	entry, _ := filepath.Abs("main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, got := range [][]byte{command(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), command(t, "", build(t, entry), cases)} {
		if !bytes.Equal(got, want) {
			mismatch(t, got, want)
		}
	}
	t.Logf("%d name verdict lines match Go, Node source and sanitized native", bytes.Count(want, []byte{'\n'}))
}
func mismatch(t *testing.T, got, want []byte) {
	t.Helper()
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output sizes: got %d Go %d", len(got), len(want))
}
func TestComponentBaseNameMutant(t *testing.T) {
	cases, want := oracle(t)
	scratch := t.TempDir()
	for _, file := range []string{"component_base_name.a", "main.a"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if file == "component_base_name.a" {
			old := "name === 'Component' || name === 'PureComponent'"
			if strings.Count(text, old) != 1 {
				t.Fatal("mutant anchor changed")
			}
			text = strings.Replace(text, old, "name === 'Component'", 1)
		}
		if file == "main.a" {
			reader, _ := filepath.Abs("../options_json.ts")
			text = strings.ReplaceAll(text, "../options_json.ts", filepath.ToSlash(reader))
		}
		if err = os.WriteFile(filepath.Join(scratch, file), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	got := command(t, "", build(t, filepath.Join(scratch, "main.a")), cases)
	if bytes.Equal(got, want) {
		t.Fatal("compiling semantic mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Logf("PureComponent mutant caught at line %d: got %q Go %q", i+1, a[i], b[i])
			return
		}
	}
	t.Fatal("mutant must change a semantic output line")
}
