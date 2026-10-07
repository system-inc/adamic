package wave08

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	log, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v %s", name, args, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr: %s", stderr.String())
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
func cases(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile("consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	var consumers []struct {
		Rule  string
		Tests []string
	}
	if err = json.Unmarshal(data, &consumers); err != nil {
		t.Fatal(err)
	}
	values := []string{"", "--color-a", "\x00", "é", "😀", "--*", "--text-sm--line-height", "var(--x)", "VAR(--x)", "xvar(--x)", ""}
	for _, consumer := range consumers {
		count := 0
		for _, path := range consumer.Tests {
			tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(tree, func(n ast.Node) bool {
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					values = append(values, value)
					count++
				}
				return true
			})
		}
		t.Logf("%s: %d fixture literals", consumer.Rule, count)
	}
	output := filepath.Join(t.TempDir(), "cases.json")
	data, err = json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	write(t, output, data)
	t.Logf("%d constructor probes", len(values))
	return output
}
func TestNewTheme(t *testing.T) {
	compareHelper(t, "theme_new.a", "main.a", "deadKeys: 0", "deadKeys: 1", "new")
}
func TestClearNamespace(t *testing.T) {
	compareHelper(t, "theme_clear_namespace.a", "clear_main.a", "!key.startsWith(namespace)", "!key.startsWith(namespace + \"-\")", "clear")
}
func compareHelper(t *testing.T, helperName, mainName, anchor, replacement, mode string) {
	// Not parallel: baseline and mutants compile large sanitizer drivers on the same worker.
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	owned, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "wave08_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "oracle.go.txt"), filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/wave08_probe.go"): filepath.Join(owned, "theme_oracle.go.txt")}})
	overlayPath := filepath.Join(scratch, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", goBinary, virtual)
	inputs := cases(t, root)
	want := run(t, "", goBinary, inputs, mode)
	for _, mutant := range []bool{false, true} {
		t.Run(strconv.FormatBool(mutant), func(t *testing.T) {
			dir := t.TempDir()
			helper, err := os.ReadFile(helperName)
			if err != nil {
				t.Fatal(err)
			}
			if mutant {
				old := anchor
				if strings.Count(string(helper), old) != 1 {
					t.Fatal("mutant anchor")
				}
				helper = []byte(strings.Replace(string(helper), old, replacement, 1))
			}
			write(t, filepath.Join(dir, helperName), helper)
			if helperName != "theme_new.a" {
				dependency, err := os.ReadFile("theme_new.a")
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(dir, "theme_new.a"), dependency)
			}
			main, err := os.ReadFile(mainName)
			if err != nil {
				t.Fatal(err)
			}
			main = []byte(strings.Replace(string(main), "../options_json.ts", filepath.ToSlash(filepath.Join(owned, "../options_json.ts")), 1))
			entry := filepath.Join(dir, "main.a")
			write(t, entry, main)
			program, err := load.Load([]string{entry})
			if err != nil {
				t.Fatal(err)
			}
			ir, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(dir, "native")
			if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			js := filepath.Join(dir, "emitted.mjs")
			write(t, js, []byte(javascript.JavaScript(ir)))
			node := filepath.Join(root, "oracle/node.mjs")
			for _, backend := range []struct {
				name, command string
				args          []string
			}{{"Node", "node", []string{"--disable-warning=ExperimentalWarning", node, entry, inputs}}, {"JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", node, js, inputs}}, {"native", binary, []string{inputs}}} {
				got := run(t, "", backend.command, backend.args...)
				if bytes.Equal(got, want) == mutant {
					t.Fatalf("mutant=%t %s wrong comparison result", mutant, backend.name)
				}
				if mutant {
					t.Logf("compiled semantic mutant caught on %s", backend.name)
				} else {
					t.Logf("%s matched Go: %d bytes", backend.name, len(want))
				}
			}
		})
	}
}

func TestBreakpointWidth(t *testing.T) {
	compareHelper(t, "breakpoint_width.a", "width_main.a", "variant.value.includes('var(')", "false", "width")
}
