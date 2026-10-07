package wave12

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

func execute(t *testing.T, directory, command string, arguments ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	process := exec.CommandContext(ctx, command, arguments...)
	process.Dir = directory
	var output, errors bytes.Buffer
	process.Stdout = &output
	process.Stderr = &errors
	if err := process.Run(); err != nil || errors.Len() != 0 {
		t.Fatalf("%s %v: %v stderr %s", command, arguments, err, &errors)
	}
	return output.Bytes()
}
func upstream(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/oracle.go.txt")
	exports, _ := filepath.Abs("testdata/exports.go.txt")
	virtual := filepath.Join(root, "adamic_wave12_oracle.go")
	mapping, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_wave12_exports.go"): exports}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, mapping, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	execute(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func backends(t *testing.T, entry, corpus string) [][]byte {
	t.Helper()
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	source := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, corpus)
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	temporary := t.TempDir()
	binary := filepath.Join(temporary, "native")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	sanitized := execute(t, "", binary, corpus)
	emitted := filepath.Join(temporary, "emitted.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	javascriptOutput := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, corpus)
	return [][]byte{source, sanitized, javascriptOutput}
}
func difference(got, want []byte) int {
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return i + 1
		}
	}
	if len(a) != len(b) {
		return len(a)
	}
	return 0
}
func TestNormalizationMatchesGo(t *testing.T) {
	entry, _ := filepath.Abs("main.a")
	corpus, _ := filepath.Abs("testdata/cases.json")
	want := execute(t, "", upstream(t), corpus)
	for i, got := range backends(t, entry, corpus) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d differs at row %d", i, difference(got, want))
		}
	}
	t.Logf("%d cases and %d UTF-16 observation bytes match actual Go, source Node, ASan/UBSan native and emitted JavaScript", bytes.Count(want, []byte{'\n'}), len(want))
}
func TestNormalizationMutants(t *testing.T) {
	corpus, _ := filepath.Abs("testdata/cases.json")
	want := execute(t, "", upstream(t), corpus)
	for _, mutant := range []struct{ name, old, replacement string }{
		{"namespace-suffix", "argument += '-*';", "argument += '*';"},
		{"newline-crossing", "if(character === '\\n') { break; }", "if(false) { break; }"},
		{"nested-key-join", "joined += argument.slice(consumed, boundary) + '-*--';", "joined += argument.slice(consumed, boundary) + '--';"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			parent := t.TempDir()
			directory := filepath.Join(parent, "wave12")
			if err := os.Mkdir(directory, 0755); err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"main.a", "collapse_normalize_value_function_argument.a", "../options_json.ts"} {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(file, "argument.a") {
					if strings.Count(string(data), mutant.old) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), mutant.old, mutant.replacement, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, file), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range backends(t, filepath.Join(directory, "main.a"), corpus) {
				if bytes.Equal(got, want) {
					t.Fatalf("backend %d mutant survived", i)
				}
				t.Logf("backend %d compiled semantic mutant caught only by Go comparison at row %d", i, difference(got, want))
			}
		})
	}
}
