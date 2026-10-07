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

func TestValueParserAgrees(t *testing.T) {
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
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(dir, "testdata/value_oracle.go"), export: filepath.Join(dir, "testdata/export.go")}})
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
		write(t, filepath.Join(evidence, "value-coverage.json"), coverage)
		data := read(t, cases)
		write(t, filepath.Join(evidence, "value-corpus.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(data))))
		write(t, filepath.Join(evidence, "value-go-output.sha256"), []byte(fmt.Sprintf("%x  want.txt\n", sha256.Sum256(want))))
	}
	runner := filepath.Join(root, "oracle/node.mjs")
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "value_main.a"), cases), want)
	binary, emitted := buildEntry(t, dir, "value_main.a")
	check(t, run(t, "", binary, cases), want)
	check(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted, cases), want)
	t.Logf("baseline Go/source Node/emitted JavaScript/sanitized native agree: %d output bytes", len(want))
	for _, mutant := range []struct{ name, old, replacement string }{
		{"CRLF normalization", "source.split('\\r\\n').join('\\n')", "source"},
		{"trailing backslash", "buffer += '\\\\undefined';", "buffer += '\\\\';"},
		{"slash word", "push('word', '/', false)", "push('separator', '/', false)"},
		{"unmatched close", "else { buffer = ''; }", "else { flush(); }"},
		{"function child presence", "stack.push(push('function', name, true));", "stack.push(push('function', name, false));"},
		{"unclosed final flush", "arena.push({kind: 'word', value: buffer, nodesPresent: false, nodes: []});\n        roots.push(index);", "arena.push({kind: 'word', value: buffer, nodesPresent: false, nodes: []});\n        const parent = stack[stack.length - 1]; if(parent === undefined) { roots.push(index); } else { (arena[parent] ?? panic('missing parent')).nodes.push(index); }"},
		{"unterminated quote", "const start = index;\n            for(let scan = index + 1;", "const start = index;\n            const scanStart = index + 1; index = input.length - 1;\n            for(let scan = scanStart;"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			mutantDir := t.TempDir()
			for _, file := range []string{"value_main.a", "parse_value.a"} {
				data := string(read(t, filepath.Join(dir, file)))
				if file == "value_main.a" {
					quoted, _ := json.Marshal(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts"))
					data = strings.Replace(data, "'../options_json.ts'", string(quoted), 1)
				}
				if file == "parse_value.a" {
					if strings.Count(data, mutant.old) != 1 {
						t.Fatal("mutant anchor drift")
					}
					data = strings.Replace(data, mutant.old, mutant.replacement, 1)
				}
				write(t, filepath.Join(mutantDir, file), []byte(data))
			}
			mutantBinary, mutantJS := buildEntry(t, mutantDir, "value_main.a")
			for _, command := range [][]string{
				{"node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(mutantDir, "value_main.a"), cases},
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
