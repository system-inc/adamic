package wave06

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
	"regexp"
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
func capturedNames(t *testing.T) []string {
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
	keys := []string{"🌊", "é", "a\x00b", "DISPLAY", "", "__wave06_absent__", "__wave06_empty_nil__", "__wave06_empty_non_nil__", "__wave06_undefined__"}
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
		keys = append(keys, regexp.MustCompile(`[A-Za-z0-9_@/-]+`).FindAllString(row.Source, -1)...)
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
	return keys
}

func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	adapter, _ := filepath.Abs("testdata/repository_export.go")
	main, _ := filepath.Abs("testdata/repository_oracle.go")
	virtual := filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_wave06.go")
	original := filepath.Join(root, "internal/lint/rules/tailwind/collapse/descriptor_live.go")
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "sorted := PropertySort(nodes)"
	if strings.Count(string(source), anchor) != 1 {
		t.Fatal("Go observation hook drift")
	}
	observation := filepath.Join(t.TempDir(), "descriptor_live.go")
	if err := os.WriteFile(observation, []byte(strings.Replace(string(source), anchor, anchor+"\n adamicRepoSorted[root] = Reading{Order: sorted.Order, Count: sorted.Count}", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: adapter, filepath.Join(root, "adamic_wave06_oracle.go"): main, original: observation, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_wave06_defs.go"): filepath.Join(filepath.Dir(adapter), "static_export.go")}})
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlay, data, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, filepath.Join(root, "adamic_wave06_oracle.go"))
	return binary
}
func sides(t *testing.T, dir, cases string) [][]byte {
	entry, _ := filepath.Abs(filepath.Join(dir, "repository_main.a"))
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
func TestRepositoryStatics(t *testing.T) {
	names := capturedNames(t)
	goOracle := oracle(t)
	definitions := run(t, "", goOracle, "definitions")
	var values []struct{ Name string }
	if err := json.Unmarshal(definitions, &values); err != nil {
		t.Fatal(err)
	}
	type row struct {
		Roots []string
		Ids   []int
	}
	cases := []row{{Roots: []string{}, Ids: []int{}}, {Roots: []string{"duplicate", "duplicate"}, Ids: []int{0, len(values) - 1}}, {Roots: []string{"🌊", "\ue000", "é", "a\x00b", "<&>\u2028\u2029"}, Ids: []int{0, 1, 2, 3, 4}}, {Roots: []string{"untouched"}, Ids: []int{len(values) - 1}}}
	for id, value := range values {
		cases = append(cases, row{Roots: []string{value.Name}, Ids: []int{id}})
	}
	for id, name := range names {
		cases = append(cases, row{Roots: []string{name}, Ids: []int{id % len(values)}})
	}
	data, err := json.Marshal(map[string]any{"Definitions": json.RawMessage(definitions), "Cases": cases})
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "cases.json")
	if err := os.WriteFile(input, data, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", goOracle, input)
	for i, got := range sides(t, ".", input) {
		if !bytes.Equal(got, want) {
			a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
			for j := 0; j < len(a) && j < len(b); j++ {
				if a[j] != b[j] {
					t.Fatalf("backend %d line %d: got %q Go %q", i, j, a[j], b[j])
				}
			}
			t.Fatalf("backend %d output size", i)
		}
	}
	t.Logf("%d cases, %d definition bodies, %d identical bytes on Go, source Node, emitted JavaScript and sanitized native", len(cases), len(values), len(want))
	for _, m := range []struct{ name, old, new string }{
		{"skip-overwrite", "statics.set(root,", "if (!statics.has(root)) { statics.set(root,"},
		{"wrong-count", "Count: sorted.Count", "Count: sorted.Count + 1"},
		{"copy-order", "Order: sorted.Order", "Order: { nil: sorted.Order.nil, values: sorted.Order.values.slice() }"},
		{"reuse-sort-record", "{ Order: sorted.Order, Count: sorted.Count }", "sorted"},
	} {
		t.Run(m.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"repository_main.a", "repository_statics.a"} {
				content, err := os.ReadFile(name)
				if err != nil {
					t.Fatal(err)
				}
				text := string(content)
				if name == "repository_statics.a" {
					if strings.Count(text, m.old) != 1 {
						t.Fatal("anchor drift")
					}
					text = strings.Replace(text, m.old, m.new, 1)
					if m.name == "skip-overwrite" {
						text = strings.Replace(text, "Count: sorted.Count });", "Count: sorted.Count }); }", 1)
					}
				} else {
					options, _ := filepath.Abs("../options_json.ts")
					text = strings.Replace(text, "../options_json.ts", filepath.ToSlash(options), 1)
				}
				if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			for i, got := range sides(t, dir, input) {
				if bytes.Equal(got, want) {
					t.Fatalf("mutant survived backend %d", i)
				}
				t.Logf("%s caught by byte comparison on backend %d after clean execution", m.name, i)
			}
		})
	}
}
