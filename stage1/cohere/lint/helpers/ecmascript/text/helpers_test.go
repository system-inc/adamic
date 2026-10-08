package text

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func command(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	output, e := os.CreateTemp(t.TempDir(), "stdout-")
	if e != nil {
		t.Fatal(e)
	}
	defer output.Close()
	c.Stdout = output
	e = c.Run()
	out, readErr := os.ReadFile(output.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if e != nil || stderr.Len() > 0 {
		t.Fatalf("%s: %v\n%s", name, e, stderr.String())
	}
	return out
}
func compile(t *testing.T, entry string, buildNative bool) (string, string) {
	t.Helper()
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "text")
	if buildNative {
		if e = native.Build(native.C(ir), bin, native.Options{Sanitize: true}); e != nil {
			t.Fatal(e)
		}
	}
	js := filepath.Join(dir, "text.mjs")
	if e = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
		t.Fatal(e)
	}
	return bin, js
}
func corpus(t *testing.T) (string, []byte) {
	t.Helper()
	compressed, e := os.Open("testdata/cases.json.gz")
	if e != nil {
		t.Fatal(e)
	}
	defer compressed.Close()
	z, e := gzip.NewReader(compressed)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	raw, e := io.ReadAll(z)
	path := filepath.Join(t.TempDir(), "cases.json")
	if e == nil {
		e = os.WriteFile(path, raw, 0644)
	}
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct{ Symbol, Want string }
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	seen := map[string]int{}
	var want strings.Builder
	for _, r := range rows {
		seen[r.Symbol]++
		want.WriteString(r.Want + "\n")
	}
	if len(seen) != 18 {
		t.Fatalf("captured %d/18 helpers", len(seen))
	}
	t.Logf("Go agreement calls: %d; by helper %v", len(rows), seen)
	return path, []byte(want.String())
}
func outputs(t *testing.T, entry, corpus string, buildNative bool) [][]byte {
	t.Helper()
	runner, _ := filepath.Abs("../../../../../../oracle/node.mjs")
	bin, js := compile(t, entry, buildNative)
	got := [][]byte{command(t, "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus), command(t, "node", "--disable-warning=ExperimentalWarning", runner, js, corpus)}
	if buildNative {
		got = append(got, command(t, bin, corpus))
	}
	return got
}
func equal(t *testing.T, got, want []byte) {
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
	t.Fatalf("got %d bytes, Go %d", len(got), len(want))
}
func TestTextAgreement(t *testing.T) {
	path, want := corpus(t)
	entry, _ := filepath.Abs("main.a")
	for i, got := range outputs(t, entry, path, true) {
		t.Run([]string{"Node", "emitted-JavaScript", "sanitized-native"}[i], func(t *testing.T) { equal(t, got, want) })
	}
}
func TestTextMutants(t *testing.T) {
	path, want := corpus(t)
	raw, e := os.ReadFile("testdata/mutants.json")
	if e != nil {
		t.Fatal(e)
	}
	var mutants []struct{ Symbol, File, From, To string }
	if e = json.Unmarshal(raw, &mutants); e != nil {
		t.Fatal(e)
	}
	files, e := filepath.Glob("*.a")
	if e != nil {
		t.Fatal(e)
	}
	if len(mutants) != 18 {
		t.Fatal("one mutant per helper required")
	}
	for _, m := range mutants {
		t.Run(m.Symbol, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, file := range files {
				data, e := os.ReadFile(file)
				if e != nil {
					t.Fatal(e)
				}
				s := string(data)
				if file == m.File {
					if !strings.Contains(s, m.From) {
						t.Fatal("mutant anchor drift")
					}
					s = strings.ReplaceAll(s, m.From, m.To)
				}
				options, _ := filepath.Abs("../../options_json.ts")
				s = strings.ReplaceAll(s, "../../options_json.ts", filepath.ToSlash(options))
				if e = os.WriteFile(filepath.Join(dir, file), []byte(s), 0644); e != nil {
					t.Fatal(e)
				}
			}
			for i, got := range outputs(t, filepath.Join(dir, "main.a"), path, false) {
				if bytes.Equal(got, want) {
					t.Fatalf("mutant survived on backend %d", i)
				}
				if len(bytes.Split(got, []byte("\n"))) != len(bytes.Split(want, []byte("\n"))) {
					t.Fatal("mutant did not complete every input")
				}
			}
			t.Log("output mutant caught on Node and emitted JavaScript")
		})
	}
}

func TestTextCapturePin(t *testing.T) {
	raw, e := os.ReadFile("testdata/coverage.json")
	if e != nil {
		t.Fatal(e)
	}
	var coverage struct {
		Pin   string
		Calls map[string]int
	}
	if e = json.Unmarshal(raw, &coverage); e != nil {
		t.Fatal(e)
	}
	root, _ := filepath.Abs("../../../../../../cohere")
	if strings.TrimSpace(string(command(t, "git", "-C", root, "rev-parse", "HEAD"))) != coverage.Pin {
		t.Fatal("Go capture pin changed; regenerate with actual Go")
	}
	got := command(t, "go", "run", "testdata/tables.go")
	want, e := os.ReadFile("unicode_tables.a")
	if e != nil {
		t.Fatal(e)
	}
	equal(t, got, want)
}
