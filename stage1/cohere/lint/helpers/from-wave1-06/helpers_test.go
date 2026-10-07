package wave06

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
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

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	file, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cmd.Stdout = file
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr: %s", stderr.String())
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func cases(t *testing.T) string {
	t.Helper()
	f, err := os.Open("testdata/sources.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var coverage struct {
		Rules        []string
		Sources      int
		CohereCommit string `json:"cohere_commit"`
	}
	dataCoverage, err := os.ReadFile("testdata/coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(dataCoverage, &coverage); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	keys := []string{"control"}
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var row struct {
			Rule   string `json:"rule"`
			Source string `json:"source"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		seen[row.Rule] = true
		keys = append(keys, hex.EncodeToString([]byte(row.Source))+"00")
		count++
	}
	if len(coverage.Rules) != 6 || count != coverage.Sources {
		t.Fatal("consumer coverage drift")
	}
	for _, rule := range coverage.Rules {
		if !seen[rule] {
			t.Fatalf("missing consumer %s", rule)
		}
	}
	actual := strings.TrimSpace(string(run(t, "../../../../../cohere", "git", "rev-parse", "HEAD")))
	if actual != coverage.CohereCommit {
		t.Fatal("cohere pin drift")
	}
	t.Logf("%d actual captured sources from all %d consumers", count, len(seen))
	path := filepath.Join(t.TempDir(), "keys.txt")
	if err := os.WriteFile(path, []byte(strings.Join(keys, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	adapter, _ := filepath.Abs("testdata/export.go")
	main, _ := filepath.Abs("testdata/oracle.go")
	virtual := filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_wave06.go")
	data, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: adapter, filepath.Join(root, "adamic_wave06_oracle.go"): main}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, filepath.Join(root, "adamic_wave06_oracle.go"))
	return binary
}
func sides(t *testing.T, dir, cases string) [][]byte {
	entry, _ := filepath.Abs(filepath.Join(dir, "main.a"))
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(t.TempDir(), "emitted.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return [][]byte{run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, cases), run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases), run(t, "", binary, cases)}
}
func TestNewTheme(t *testing.T) {
	input := cases(t)
	want := run(t, "", oracle(t), input)
	for i, got := range sides(t, ".", input) {
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d mismatched Go", i)
		}
	}
	t.Logf("%d identical bytes on Go, source Node, emitted JavaScript and sanitized native", len(want))
	for _, mutant := range []struct{ name, old, new string }{
		{"dirty-prefix", "Prefix: ''", "Prefix: 'dirty'"},
		{"allocated-order", "nil: true", "nil: false"},
		{"shared-map", "values: new Map<string, ThemeValue>()", "values: sharedValues"},
		{"shared-record", "return { Prefix:", "return sharedTheme; // { Prefix:"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"main.a", "new_theme.a"} {
				data, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				if name == "new_theme.a" {
					s := string(data)
					if strings.Count(s, mutant.old) != 1 {
						t.Fatal("mutant anchor drift")
					}
					s = strings.Replace(s, mutant.old, mutant.new, 1)
					if mutant.name == "shared-map" {
						s += "\nconst sharedValues = new Map<string, ThemeValue>();\n"
					}
					if mutant.name == "shared-record" {
						s += "\nconst sharedTheme: Theme = { Prefix: '', values: new Map<string, ThemeValue>(), keyOrder: { nil: true, values: [] }, deadKeys: 0 };\n"
					}
					data = []byte(s)
				}
				if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range sides(t, dir, input) {
				if bytes.Equal(got, want) {
					t.Fatalf("mutant survived backend %d", i)
				}
				t.Logf("%s caught only by Go byte comparison on backend %d; execution clean", mutant.name, i)
			}
		})
	}
}
