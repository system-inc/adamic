package atrule

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

// Not parallel: compile baseline and semantic mutants sequentially to bound memory.
func TestParseAtRule(t *testing.T) {
	root, _ := filepath.Abs("../../../../../..")
	dir, _ := filepath.Abs(".")
	cohere := filepath.Join(root, "cohere")
	if pin := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); pin != "715ba94f3608a6500086b1076ce5cb7e51b836db" {
		t.Fatal("Go pin drift", pin)
	}
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_wave13_at_rule.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata", "oracle.go")}})
	overlayPath := filepath.Join(scratch, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(scratch, "go-oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	cases := filepath.Join(scratch, "cases.json")
	t.Log(strings.TrimSpace(string(run(t, "", oracle, "--capture", cohere, cases))))
	want := run(t, "", oracle, cases)
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, variant := range []string{"baseline", "byte-offset-mutant", "NEL-trim-mutant", "child-copy-mutant"} {
		sourceDir := dir
		if variant != "baseline" {
			sourceDir = t.TempDir()
			for _, name := range []string{"main.a", "parse_at_rule.a"} {
				data, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				text := string(data)
				if name == "main.a" {
					text = strings.ReplaceAll(text, "'../../options_json.ts'", strconvQuote(filepath.Join(dir, "../../options_json.ts")))
				}
				if name == "parse_at_rule.a" {
					switch variant {
					case "byte-offset-mutant":
						text = strings.Replace(text, "byteOffset >= 5", "byteOffset >= 1", 1)
					case "NEL-trim-mutant":
						text = strings.Replace(text, `\u0085`, "", 1)
					case "child-copy-mutant":
						text = strings.Replace(text, "trimGo(params), nodes)", "trimGo(params), [...nodes])", 1)
					}
				}
				if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
		}
		emitted := wave13JavaScript(t, sourceDir)
		binary := wave13Build(t, sourceDir)
		for _, command := range [][]string{
			{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(sourceDir, "main.a"), cases},
			{"node", "--disable-warning=ExperimentalWarning", runner, emitted, cases},
			{binary, cases},
		} {
			got := run(t, "", command[0], command[1:]...)
			if variant == "baseline" {
				compare(t, got, want)
			} else if bytes.Equal(got, want) {
				t.Fatal(variant, "survived")
			}
		}
		t.Logf("%s compiled and finished cleanly on source Node, emitted JavaScript and sanitized native; compared %d Go bytes", variant, len(want))
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

func wave13Build(t *testing.T, dir string) string {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, "main.a")})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "wave13")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

func strconvQuote(path string) string { data, _ := json.Marshal(path); return string(data) }

func wave13JavaScript(t *testing.T, dir string) string {
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
