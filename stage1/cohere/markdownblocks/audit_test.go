package markdownblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

type auditInput struct {
	Corpus bool   `json:"-"`
	Name   string `json:"name"`
	Text   string `json:"text"`
}
type auditOutput struct {
	Name      string   `json:"name"`
	Auto      string   `json:"auto"`
	Off       string   `json:"off"`
	AutoError string   `json:"autoError"`
	OffError  string   `json:"offError"`
	Embeds    []string `json:"embeds"`
}

// census reports whether the corpus covers every Markdown file in the repository and its submodules. It is
// opt-in: that set moves with every commit and every cohere bump, and a gate must not turn red on a README this
// unit never owned. Without it the corpus is this unit's own files plus the generated documents.
func census() bool {
	return os.Getenv("ADAMIC_MARKDOWNBLOCKS_CENSUS") != ""
}

func auditCorpus(t *testing.T, root string) ([]auditInput, int) {
	t.Helper()
	var inputs []auditInput
	patterns := []string{"*.md", "*.markdown", "*.mdown", "*.mkd"}
	roots := []string{"stage1/cohere/markdownblocks"}
	if census() {
		roots = []string{"."}
	}
	paths := corpusfiles.Repository(t, root, roots, patterns)
	if census() {
		paths = append(paths, corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"CHANGELOG.md", "CONTRIBUTING.md", "README.md", "THIRD_PARTY_NOTICES.md", "TypeScript-shim", "editors", "internal", "schema", "swift"}, patterns)...)
		paths = append(paths, corpusfiles.Upstream(t, filepath.Join(root, "cohere/TypeScript"), corpusfiles.TypeScriptGoCommit, []string{".github", "CODE_OF_CONDUCT.md", "CONTRIBUTING.md", "README.md", "SECURITY.md", "SUPPORT.md", "packages", "tsc"}, patterns)...)
	}
	sort.Strings(paths)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, auditInput{Name: filepath.ToSlash(name), Text: string(content)})
	}
	files := len(inputs)
	for _, marker := range []string{"*", "-", "+", "1.", "1)", "10.", "999999999."} {
		for depth := 1; depth <= 5; depth++ {
			var lines []string
			for level := 0; level < depth; level++ {
				lines = append(lines, strings.Repeat("  ", level)+marker+" item *text* `x`")
			}
			inputs = append(inputs, auditInput{Name: "generated/list/" + marker + "/" + strings.Repeat("n", depth), Text: strings.Join(lines, "\n") + "\n"})
		}
	}
	for _, text := range []string{"# x #", "## x", "x\n=", "x\n---", "a\nb\n==="} {
		inputs = append(inputs, auditInput{Name: "generated/heading/" + text, Text: text + "\n"})
	}
	for _, prefix := range []string{"", "> ", "- ", "  "} {
		for _, block := range []string{"```\nx\n```", "~~~\nx\n~~~", "    x\n    y", "<div>\nx\n</div>", "<!-- x -->", "***", "---\na: b\n---", "|a|b|\n|-|-|\n|x|y|"} {
			inputs = append(inputs, auditInput{Name: "generated/block/" + prefix + block, Text: prefix + strings.ReplaceAll(block, "\n", "\n"+prefix) + "\n"})
		}
	}
	for _, code := range []string{`if (true) { console.log("x"); }`, `for (let i=0;i<2;i++) console.log(i);`, `const x = <div className="x">x</div>;`} {
		inputs = append(inputs, auditInput{Name: "generated/embedding/" + code, Text: "```tsx\n" + code + "\n```\n"})
	}
	for _, name := range []string{"embedded_jsonc.md", "embedded_flow.md"} {
		content, err := os.ReadFile(filepath.Join("gaps", name))
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, auditInput{Name: "generated/" + name, Text: string(content)})
	}
	return inputs, files
}

func auditResults(t *testing.T, name string, result run) []auditOutput {
	t.Helper()
	clean(t, name, result)
	var answers []auditOutput
	for _, line := range bytes.Split(bytes.TrimSuffix(result.stdout, []byte("\n")), []byte("\n")) {
		var answer auditOutput
		if err := json.Unmarshal(line, &answer); err != nil {
			t.Fatal(err)
		}
		if answer.AutoError != "" || answer.OffError != "" {
			t.Fatalf("%s formatting error at %s: %s %s", name, answer.Name, answer.AutoError, answer.OffError)
		}
		answers = append(answers, answer)
	}
	return answers
}

func TestWholeDocumentOraclePreflight(t *testing.T) {
	parallelMarkdown(t)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	inputs, files := auditCorpus(t, root)
	// This file sweep is small beside the mandatory generated and fixed checks.
	selection := selectMarkdownFiles(t, root, inputs, files, 1)
	if selection.Sample {
		t.Log(selection.Log(t.Name()))
	}
	inputs, files = selectedMarkdownInputs(inputs, files, selection)
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	driver, err := filepath.Abs("testdata/go_driver.go")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_markdown_blocks/main.go")
	build := func(replacements map[string]string, binary string) {
		t.Helper()
		replacements[mainPath] = driver
		overlay, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			t.Fatal(err)
		}
		path := binary + ".overlay.json"
		write(t, path, overlay)
		command := bounded(t, "go", "build", "-overlay="+path, "-o", binary, mainPath)
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			t.Fatalf("build: %v\n%s", err, output)
		}
	}
	binary := filepath.Join(dir, "go-printer")
	build(map[string]string{}, binary)
	goRun := execute(t, nil, binary, cases)
	goAnswers := auditResults(t, "Go", goRun)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	// The scratch installation must be the original fork bytes, not merely a package with its version.
	for _, name := range []string{"standalone.js", "plugins/estree.js", "plugins/typescript.js", "plugins/babel.js", "plugins/postcss.js", "plugins/markdown.js", "plugins/graphql.js", "plugins/yaml.js"} {
		original, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles", name))
		if err != nil {
			t.Fatal(err)
		}
		installed, err := os.ReadFile(filepath.Join(fork, name))
		if err != nil {
			t.Fatal(err)
		}
		equal(t, "pinned fork bundle "+name, installed, original)
	}
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	nodeRun := execute(t, nil, "node", script, fork, cases, "fork")
	nodeAnswers := auditResults(t, "pinned fork", nodeRun)
	if len(goAnswers) != len(inputs) || len(nodeAnswers) != len(inputs) {
		t.Fatal("oracle lost a document")
	}
	gaps := map[string]bool{
		"cohere/TypeScript/packages/vscode-typescript/README.md":                                                                                     false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_fn_type_mismatch_2.expect.md":                                 false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-hir_loc_diff_1.expect.md":                                         false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-pattern3_type_to_poly.expect.md":                                  false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_identifier_diff.expect.md":                                 false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-round2_severity_diff.expect.md":                                   false,
		"cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.todo-update-expression-context-variable-via-type-annotation.expect.md": false,
		"stage1/cohere/markdownblocks/gaps/embedded_jsonc.md":                                                                                        false,
		"stage1/cohere/markdownblocks/gaps/embedded_flow.md":                                                                                         false,
		"generated/embedded_jsonc.md": false,
		"generated/embedded_flow.md":  false,
	}
	expectedWitnesses := 7
	if !census() {
		for name := range gaps {
			if strings.HasPrefix(name, "cohere/") {
				delete(gaps, name)
			}
		}
		expectedWitnesses = 0
	}
	witnessFile, err := os.ReadFile("GAPS.md")
	if err != nil {
		t.Fatal(err)
	}
	var witnessStrings []string
	for _, line := range strings.Split(string(witnessFile), "\n") {
		if !strings.HasPrefix(line, `"`) {
			continue
		}
		var value string
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			t.Fatalf("witness JSON: %v", err)
		}
		witnessStrings = append(witnessStrings, value)
	}
	witnessCount := 0
	for index, actual := range goAnswers {
		expected := nodeAnswers[index]
		if actual.Name != inputs[index].Name || expected.Name != actual.Name {
			t.Fatal("oracle reordered documents")
		}
		equal(t, "embedding off: "+actual.Name, []byte(actual.Off), []byte(expected.Off))
		if actual.Auto != expected.Auto {
			if _, known := gaps[actual.Name]; !known {
				t.Fatalf("new auto-mode discrepancy at %s", actual.Name)
			}
			gaps[actual.Name] = true
			if strings.HasPrefix(actual.Name, "cohere/") {
				witnessCount++
				for _, output := range []string{actual.Auto, expected.Auto} {
					found := false
					for _, recorded := range witnessStrings {
						if recorded == output {
							found = true
						}
					}
					if !found {
						t.Fatalf("GAPS.md lost complete output for %s", actual.Name)
					}
				}
				digest := fmt.Sprintf("%x", sha256.Sum256([]byte(inputs[index].Text)))
				if !strings.Contains(string(witnessFile), "`"+digest+"`") {
					t.Fatalf("GAPS.md lost input digest for %s", actual.Name)
				}
			}
			offset := firstDifference(actual.Auto, expected.Auto)
			t.Logf("known auto gap %s at byte %d, embeds %v", actual.Name, offset, actual.Embeds)
		}
	}
	if witnessCount != expectedWitnesses || len(witnessStrings) != 14 {
		t.Fatalf("expected %d witnesses and fourteen complete outputs, got %d and %d", expectedWitnesses, witnessCount, len(witnessStrings))
	}
	for name, seen := range gaps {
		if !seen {
			t.Fatalf("auto gap closed or witness missing: %s; update GAPS.md", name)
		}
	}
	if keep := os.Getenv("ADAMIC_MARKDOWNBLOCKS_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(keep, "cases.jsonl"), batch.Bytes())
		write(t, filepath.Join(keep, "go.jsonl"), goRun.stdout)
		write(t, filepath.Join(keep, "fork.jsonl"), nodeRun.stdout)
	}
	for _, mutation := range []struct{ name, path, from, to string }{
		{"heading prefix", "heading.go", `strings.Repeat("#", currentNode(path).Depth)`, `strings.Repeat("#", currentNode(path).Depth+1)`},
		{"list marker", "list.go", `rawPrefix = "- "`, `rawPrefix = "+ "`},
		{"fence length", "print.go", `max(3, getMaxContinuousCount(value, "` + "`" + `")+1)`, `max(4, getMaxContinuousCount(value, "` + "`" + `")+1)`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			source := filepath.Join(cohere, "internal/format/markdown", mutation.path)
			content, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(content), mutation.from) != 1 {
				t.Fatal("mutant must alter exactly one site")
			}
			altered := filepath.Join(t.TempDir(), mutation.path)
			write(t, altered, []byte(strings.Replace(string(content), mutation.from, mutation.to, 1)))
			mutant := filepath.Join(t.TempDir(), "mutant")
			build(map[string]string{source: altered}, mutant)
			answers := auditResults(t, "Go mutant", execute(t, nil, mutant, cases))
			if len(answers) != len(inputs) {
				t.Fatal("mutant lost a document")
			}
			caught := false
			for index, actual := range answers {
				if actual.Name != inputs[index].Name {
					t.Fatal("mutant reordered documents")
				}
				if actual.Off != nodeAnswers[index].Off {
					t.Logf("output-only mutant caught by %s at byte %d", actual.Name, firstDifference(actual.Off, nodeAnswers[index].Off))
					caught = true
					break
				}
			}
			if !caught {
				t.Fatal("mutant survived byte comparison")
			}
		})
	}
	for _, side := range []struct {
		name, command string
		args          []string
	}{
		{"Go Markdown embedding off", binary, []string{cases, "off-only"}},
		{"Node pinned-fork Markdown embedding off", "node", []string{script, fork, cases, "fork", "off-only"}},
	} {
		var elapsed time.Duration
		for round := 0; round < 3; round++ {
			start := time.Now()
			answer := execute(t, nil, side.command, side.args...)
			elapsed += time.Since(start)
			answers := auditResults(t, side.name, answer)
			if len(answers) != len(inputs) {
				t.Fatal("measured baseline lost a document")
			}
			for index, actual := range answers {
				if actual.Name != inputs[index].Name {
					t.Fatal("measured baseline reordered documents")
				}
				equal(t, side.name+": "+actual.Name, []byte(actual.Off), []byte(goAnswers[index].Off))
			}
		}
		t.Logf("baseline throughput %s %.1f documents/s, 3 runs %.6fs; one off-mode format per document, startup/I/O included, no Adamic block formatter measured", side.name, float64(len(inputs)*3)/elapsed.Seconds(), elapsed.Seconds())
	}
	t.Logf("corpus %d physical repository/submodule Markdown files, %d generated documents, %d total; off byte-identical, auto %d/%d identical with %d named disagreements", files, len(inputs)-files, len(inputs), len(inputs)-len(gaps), len(inputs), len(gaps))
}
func firstDifference(a, b string) int {
	index := 0
	for index < len(a) && index < len(b) && a[index] == b[index] {
		index++
	}
	return index
}
