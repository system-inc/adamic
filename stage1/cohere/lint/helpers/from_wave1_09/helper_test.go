package fromwave109

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestStaticDeclarationNodesAgree(t *testing.T) {
	// Not parallel: the oracle owns one generated corpus and the evidence filenames.
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	if pin := strings.TrimSpace(string(run(t, cohere, "git", "rev-parse", "HEAD"))); pin != "715ba94f3608a6500086b1076ce5cb7e51b836db" {
		t.Fatal("Go pin drift", pin)
	}
	scratch := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_wave109_nodes.go")
	export := filepath.Join(cohere, "internal/lint/rules/tailwind/collapse/adamic_wave109_nodes.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata/oracle.go"), export: filepath.Join(dir, "testdata/export.go")}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(scratch, "overlay.json")
	write(t, overlayPath, overlay)
	oracle := filepath.Join(scratch, "oracle")
	run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	t.Log(strings.TrimSpace(string(run(t, "", oracle, root, scratch))))
	cases := filepath.Join(scratch, "cases.json")
	want := read(t, filepath.Join(scratch, "want.txt"))
	coverage := read(t, filepath.Join(scratch, "coverage.json"))
	t.Logf("coverage: %s", coverage)
	if evidence := os.Getenv("ADAMIC_WAVE109_HELPER_EVIDENCE"); evidence != "" {
		write(t, filepath.Join(evidence, "coverage.json"), coverage)
		data := read(t, cases)
		write(t, filepath.Join(evidence, "corpus.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(data))))
		write(t, filepath.Join(evidence, "go-output.sha256"), []byte(fmt.Sprintf("%x  want.txt\n", sha256.Sum256(want))))
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "main.a"), cases), want)
	binary, emitted := build(t, dir)
	check(t, run(t, "", binary, cases), want)
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases), want)
	t.Logf("baseline Go/source Node/emitted JavaScript/sanitized native agree: %d output bytes", len(want))
	for _, mutant := range []struct{ name, old, replacement string }{
		{"absent versus empty", "valuePresent: declaration.valuePresent", "valuePresent: declaration.value.length > 0"},
		{"important flag", "important: declaration.important", "important: false"},
		{"declaration kind", "kind: 'declaration'", "kind: 'comment'"},
		{"declaration order", "return nodes;", "return nodes.reverse();"},
		{"fresh nodes", "return nodes;", "const first = nodes[0]; if(first !== undefined && nodes.length > 1) { nodes[1] = first; } return nodes;"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			mutantDir := t.TempDir()
			for _, file := range []string{"main.a", "nodes_from_static_declarations.a"} {
				data := string(read(t, filepath.Join(dir, file)))
				if file == "main.a" {
					quoted, _ := json.Marshal(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts"))
					data = strings.Replace(data, "'../options_json.ts'", string(quoted), 1)
				}
				if file == "nodes_from_static_declarations.a" {
					if strings.Count(data, mutant.old) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = strings.Replace(data, mutant.old, mutant.replacement, 1)
				}
				write(t, filepath.Join(mutantDir, file), []byte(data))
			}
			mutantBinary, mutantJS := build(t, mutantDir)
			for _, command := range [][]string{
				{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(mutantDir, "main.a"), cases},
				{"node", "--disable-warning=ExperimentalWarning", runner, mutantJS, cases},
				{mutantBinary, cases},
			} {
				got := run(t, "", command[0], command[1:]...)
				if bytes.Equal(got, want) {
					t.Fatal("semantic mutant survived", command[0])
				}
				a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
				for i := 0; i < len(a) && i < len(b); i++ {
					if a[i] != b[i] {
						t.Logf("%s compiled and exited 0; comparison caught line %d: mutant %q Go %q", command[0], i+1, a[i], b[i])
						break
					}
				}
			}
		})
	}
}
func build(t *testing.T, dir string) (string, string) {
	return buildEntry(t, dir, "main.a")
}
func buildEntry(t *testing.T, dir, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(dir, entry)})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	binary := filepath.Join(output, "native")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(output, "emitted.mjs")
	write(t, emitted, []byte(javascript.JavaScript(ir)))
	return binary, emitted
}
func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	stdout, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	stderr, err := os.CreateTemp(t.TempDir(), "stderr-")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	diagnostic := read(t, stderr.Name())
	if err != nil || len(diagnostic) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, diagnostic)
	}
	return read(t, stdout.Name())
}
func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
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
func check(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output length: got %d Go %d", len(got), len(want))
}
