package batch4

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

func TestSlot05LiveKeys(t *testing.T) {
	verify(t, "live", "theme_live_keys.a", "if(!visit(key, value))", "if(visit(key, value))")
}
func TestSlot05ThemeEntries(t *testing.T) {
	verify(t, "entries", "theme_entries.a", "options: value.options", "options: 0")
}
func TestSlot05KeysInNamespaces(t *testing.T) {
	verify(t, "namespaces", "theme_keys_in_namespaces.a", "if(nested)", "if(false)")
}
func verify(t *testing.T, mode, target, old, replacement string) {
	root, _ := filepath.Abs("../../../../../..")
	cohere := filepath.Join(root, "cohere")
	dir, _ := filepath.Abs(".")
	if got := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); got != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("Go pin drift", got)
	}

	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_slot05_batch4.go")
	replacements := map[string]string{virtual: filepath.Join(dir, "testdata", mode+"_oracle.go")}
	replacements[filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_slot05_batch4.go")] = filepath.Join(dir, "testdata/collapse_export.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": replacements})
	overlayPath := filepath.Join(scratch, "overlay.json")
	if e := os.WriteFile(overlayPath, overlay, 0644); e != nil {
		t.Fatal(e)
	}
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	t.Log(strings.TrimSpace(string(run(t, "", oracle, root, scratch))))
	cases := filepath.Join(scratch, "cases.json")
	want, e := os.ReadFile(filepath.Join(scratch, "want.txt"))
	if e != nil {
		t.Fatal(e)
	}
	coverage, e := os.ReadFile(filepath.Join(scratch, "coverage.json"))
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("consumer coverage: %s", coverage)
	if evidence := os.Getenv("ADAMIC_SLOT05_BATCH4_EVIDENCE"); evidence != "" {
		if e = os.WriteFile(filepath.Join(evidence, mode+"-coverage.json"), coverage, 0644); e != nil {
			t.Fatal(e)
		}
		data, e := os.ReadFile(cases)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(evidence, mode+"-corpus.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(data))), 0644); e != nil {
			t.Fatal(e)
		}
	}
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "main.a"), cases), want)
	emitted := slot05JavaScript(t, dir)
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), emitted, cases), want)
	binary := slot05Build(t, dir)
	compare(t, run(t, "", binary, cases), want)
	if mode == "namespaces" {
		refusal := filepath.Join(scratch, "refusal.json")
		for _, args := range [][]string{{binary, refusal}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), emitted, refusal}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(dir, "main.a"), refusal}} {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			output, e := os.CreateTemp(t.TempDir(), "refusal-")
			if e != nil {
				t.Fatal(e)
			}
			cmd.Stdout = output
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			e = cmd.Run()
			cancel()
			output.Close()
			exit, ok := e.(*exec.ExitError)
			if !ok || exit.ExitCode() != 70 || !strings.Contains(stderr.String(), "theme key shorter than two UTF-8 bytes") {
				t.Fatalf("short key refusal: %v %q", e, stderr.String())
			}
		}
		t.Log("Go, Node source, emitted JavaScript and native agree on matched too-short-key refusal")
	}
	mutant := t.TempDir()
	entries, e := os.ReadDir(dir)
	if e != nil {
		t.Fatal(e)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		data, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			t.Fatal(e)
		}
		if name == target {
			if strings.Count(string(data), old) != 1 {
				t.Fatal("mutant anchor drift")
			}
			data = []byte(strings.Replace(string(data), old, replacement, 1))
		}
		if name == "main.a" {
			data = []byte(strings.Replace(string(data), "'../../options_json.ts'", strconvQuote(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")), 1))
		}
		if e = os.WriteFile(filepath.Join(mutant, name), data, 0644); e != nil {
			t.Fatal(e)
		}
	}
	got := run(t, "", slot05Build(t, mutant), cases)
	if bytes.Equal(got, want) {
		t.Fatal("compiled semantic mutant survived")
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := range a {
		if i < len(b) && a[i] != b[i] {
			t.Logf("%s mutant caught at verdict %d: mutant %s, Go %s", target, i+1, a[i], b[i])
			break
		}
	}
}
func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	f, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("output line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size: got %d Go %d", len(got), len(want))
}

func slot05Build(t *testing.T, dir string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "slot05")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func strconvQuote(path string) string { data, _ := json.Marshal(path); return string(data) }

func slot05JavaScript(t *testing.T, dir string) string {
	t.Helper()
	program, e := load.Load([]string{filepath.Join(dir, "main.a")})
	if e != nil {
		t.Fatal(e)
	}
	lowered, e := lower.Lower(context.Background(), program)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "emitted.mjs")
	if e = os.WriteFile(path, []byte(javascript.JavaScript(lowered)), 0644); e != nil {
		t.Fatal(e)
	}
	return path
}
