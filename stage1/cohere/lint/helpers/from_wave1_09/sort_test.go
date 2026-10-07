package fromwave109

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrivatePropertySortAgrees(t *testing.T) {
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
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata/sort_oracle.go"), export: filepath.Join(dir, "testdata/export.go")}})
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
		write(t, filepath.Join(evidence, "sort-coverage.json"), coverage)
		data := read(t, cases)
		write(t, filepath.Join(evidence, "sort-corpus.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(data))))
		write(t, filepath.Join(evidence, "sort-go-output.sha256"), []byte(fmt.Sprintf("%x  want.txt\n", sha256.Sum256(want))))
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "sort_main.a"), cases), want)
	binary, emitted := buildEntry(t, dir, "sort_main.a")
	check(t, run(t, "", binary, cases), want)
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases), want)
	t.Logf("baseline Go/source Node/emitted JavaScript/sanitized native agree: %d output bytes", len(want))
	for _, mutant := range []struct{ name, old, replacement string }{
		{"absent values", "if(!node.valuePresent)", "if(node.value.length === 0)"},
		{"count after latch", "count++;", "if(!seenTwSort) { count++; }"},
		{"unknown sort latch", "const position = propertyOrder.get(node.value);", "seenTwSort = true; const position = propertyOrder.get(node.value);"},
		{"context descent", "node.kind === 'rule' || node.kind === 'at-rule'", "node.kind === 'rule' || node.kind === 'at-rule' || node.kind === 'context'"},
		{"breadth first", "queue.push(child);", "queue.splice(head + 1, 0, child);"},
		{"position deduplication", "position !== undefined && !seen.has(position)", "position !== undefined"},
		{"ascending order", "left - right", "right - left"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			mutantDir := t.TempDir()
			for _, file := range []string{"sort_main.a", "property_sort.a"} {
				data := string(read(t, filepath.Join(dir, file)))
				if file == "sort_main.a" {
					quoted, _ := json.Marshal(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts"))
					data = strings.Replace(data, "'../options_json.ts'", string(quoted), 1)
				}
				if file == "property_sort.a" {
					if strings.Count(data, mutant.old) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = strings.Replace(data, mutant.old, mutant.replacement, 1)
				}
				write(t, filepath.Join(mutantDir, file), []byte(data))
			}
			mutantBinary, mutantJS := buildEntry(t, mutantDir, "sort_main.a")
			for _, command := range [][]string{
				{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(mutantDir, "sort_main.a"), cases},
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
