package markdownblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const testWholeDocumentOraclePreflightShards = 4

type preflightProducts struct {
	binary, fork, script string
	mutants              []string
	planted              string
}

func prepareWholeDocumentPreflight(t *testing.T, root string) *preflightProducts {
	dir, err := os.MkdirTemp(artifactDirectory, "preflight-")
	if err != nil {
		t.Fatal(err)
	}
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
	var mutants []string
	for _, mutation := range []struct{ name, path, from, to string }{
		{"heading prefix", "heading.go", `strings.Repeat("#", currentNode(path).Depth)`, `strings.Repeat("#", currentNode(path).Depth+1)`},
		{"list marker", "list.go", `rawPrefix = "- "`, `rawPrefix = "+ "`},
		{"fence length", "print.go", `max(3, getMaxContinuousCount(value, "` + "`" + `")+1)`, `max(4, getMaxContinuousCount(value, "` + "`" + `")+1)`},
	} {
		{
			source := filepath.Join(cohere, "internal/format/markdown", mutation.path)
			content, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(content), mutation.from) != 1 {
				t.Fatal("mutant must alter exactly one site")
			}
			altered := filepath.Join(dir, mutation.path)
			write(t, altered, []byte(strings.Replace(string(content), mutation.from, mutation.to, 1)))
			mutant := filepath.Join(dir, fmt.Sprintf("mutant-%d", len(mutants)))
			build(map[string]string{source: altered}, mutant)
			mutants = append(mutants, mutant)
		}
	}
	return &preflightProducts{binary: binary, fork: fork, script: script, mutants: mutants}
}
func runWholeDocumentPreflightShard(t *testing.T, root string, inputs []auditInput, files, shard int, products *preflightProducts) {
	dir := t.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if err := encoder.Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(t, cases, batch.Bytes())
	binary := products.binary
	goRun := execute(t, nil, binary, cases)
	goAnswers := auditResults(t, "Go", goRun)
	fork, script := products.fork, products.script
	nodeRun := execute(t, nil, "node", script, fork, cases, "fork")
	nodeAnswers := auditResults(t, "pinned fork", nodeRun)
	if len(goAnswers) != len(inputs) || len(nodeAnswers) != len(inputs) {
		t.Fatal("oracle lost a document")
	}
	gaps := preflightRequiredGaps()
	expectedWitnesses := 0
	for name := range gaps {
		if preflightShard(name) != shard {
			delete(gaps, name)
		} else if strings.HasPrefix(name, "cohere/") {
			expectedWitnesses++
		}
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
		if err := preflightOffComparison(actual, expected); err != nil {
			t.Fatal(err)
		}
		if actual.Name == products.planted {
			planted := expected
			planted.Off += "planted disagreement"
			if preflightOffComparison(actual, planted) == nil {
				t.Fatal("planted failure escaped comparison")
			}
			t.Logf("planted failure caught by shard-%03d: %s", shard, actual.Name)
		}
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
		write(t, filepath.Join(keep, fmt.Sprintf("shard-%03d-cases.jsonl", shard)), batch.Bytes())
		write(t, filepath.Join(keep, fmt.Sprintf("shard-%03d-go.jsonl", shard)), goRun.stdout)
		write(t, filepath.Join(keep, fmt.Sprintf("shard-%03d-fork.jsonl", shard)), nodeRun.stdout)
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

func preflightOffComparison(actual, expected auditOutput) error {
	if !bytes.Equal([]byte(actual.Off), []byte(expected.Off)) {
		return fmt.Errorf("embedding off: %s differs at byte %d", actual.Name, firstDifference(actual.Off, expected.Off))
	}
	return nil
}

func preflightRequiredGaps() map[string]bool {
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

	if !census() {
		for name := range gaps {
			if strings.HasPrefix(name, "cohere/") {
				delete(gaps, name)
			}
		}
	}
	return gaps
}
func preflightShard(name string) int {
	digest := sha256.Sum256([]byte(name))
	return int(uint32(digest[0])<<24|uint32(digest[1])<<16|uint32(digest[2])<<8|uint32(digest[3])) % testWholeDocumentOraclePreflightShards
}

// Products and the live corpus are prepared once for all top-level leaves. TestMain
// owns their directory so one leaf's cleanup cannot invalidate another's products.
type wholeDocumentPreflightFixture struct {
	root     string
	inputs   []auditInput
	shards   [testWholeDocumentOraclePreflightShards][]auditInput
	physical [testWholeDocumentOraclePreflightShards]int
	products *preflightProducts
}

var wholeDocumentPreflightOnce sync.Once
var wholeDocumentPreflightShared *wholeDocumentPreflightFixture

func wholeDocumentPreflight(t *testing.T) *wholeDocumentPreflightFixture {
	t.Helper()
	wholeDocumentPreflightOnce.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		inputs, files := auditCorpus(t, root)
		selection := selectMarkdownFiles(t, root, inputs, files, 1)
		if selection.Sample {
			t.Log(selection.Log(t.Name()))
		}
		inputs, files = selectedMarkdownInputs(inputs, files, selection)
		fixture := &wholeDocumentPreflightFixture{root: root, inputs: inputs, products: prepareWholeDocumentPreflight(t, root)}
		for index, input := range inputs {
			shard := preflightShard(input.Name)
			fixture.shards[shard] = append(fixture.shards[shard], input)
			if index < files {
				fixture.physical[shard]++
			}
		}
		fixture.products.planted = inputs[0].Name
		wholeDocumentPreflightShared = fixture
	})
	if wholeDocumentPreflightShared == nil {
		t.Fatal("preflight setup failed")
	}
	return wholeDocumentPreflightShared
}
func TestWholeDocumentOraclePreflightUnion(t *testing.T) {
	t.Parallel()
	fixture := wholeDocumentPreflight(t)
	expected := make(map[string]int)
	actual := make(map[string]int)
	for _, input := range fixture.inputs {
		expected[input.Name]++
	}
	count, caught := 0, 0
	for shard, cases := range fixture.shards {
		for _, input := range cases {
			if preflightShard(input.Name) != shard {
				t.Fatal("unstable shard assignment")
			}
			actual[input.Name]++
			count++
			oracle := auditOutput{Name: input.Name, Off: input.Text}
			planted := oracle
			if input.Name == fixture.products.planted {
				planted.Off += "planted disagreement"
			}
			if preflightOffComparison(oracle, planted) != nil {
				caught++
				t.Logf("planted failure caught by shard-%03d: %s", shard, input.Name)
			}
		}
	}
	if len(fixture.shards) != testWholeDocumentOraclePreflightShards || count != len(fixture.inputs) || len(actual) != len(expected) {
		t.Fatal("shard union lost a document")
	}
	for name, n := range expected {
		if n != 1 || actual[name] != 1 {
			t.Fatalf("shard union count for %s: %d", name, actual[name])
		}
	}
	for name := range preflightRequiredGaps() {
		if expected[name] != 1 {
			t.Fatalf("auto gap witness missing: %s", name)
		}
	}
	if caught != 1 {
		t.Fatal("planted disagreement did not reach exactly one shard")
	}
	t.Logf("union %d documents across %d shards", count, len(fixture.shards))
}
func wholeDocumentPreflightLeaf(t *testing.T, shard int) {
	t.Helper()
	t.Parallel()
	fixture := wholeDocumentPreflight(t)
	runWholeDocumentPreflightShard(t, fixture.root, fixture.shards[shard], fixture.physical[shard], shard, fixture.products)
}
func TestWholeDocumentOraclePreflight_000(t *testing.T) { wholeDocumentPreflightLeaf(t, 0) }
func TestWholeDocumentOraclePreflight_001(t *testing.T) { wholeDocumentPreflightLeaf(t, 1) }
func TestWholeDocumentOraclePreflight_002(t *testing.T) { wholeDocumentPreflightLeaf(t, 2) }
func TestWholeDocumentOraclePreflight_003(t *testing.T) { wholeDocumentPreflightLeaf(t, 3) }

// Mutation witnesses are a property of the entire corpus, rather than every
// arbitrary hash bucket. This leaf retains the original full-corpus must-fail check.
func TestWholeDocumentOraclePreflightMutants(t *testing.T) {
	t.Parallel()
	fixture := wholeDocumentPreflight(t)
	products, inputs := fixture.products, fixture.inputs
	var batch bytes.Buffer
	for _, input := range inputs {
		if err := json.NewEncoder(&batch).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(t.TempDir(), "cases.jsonl")
	write(t, cases, batch.Bytes())
	nodeAnswers := auditResults(t, "pinned fork", execute(t, nil, "node", products.script, products.fork, cases, "fork"))
	if len(nodeAnswers) != len(inputs) {
		t.Fatal("oracle lost a document")
	}
	for index, answer := range nodeAnswers {
		if answer.Name != inputs[index].Name {
			t.Fatal("oracle reordered documents")
		}
	}
	for _, mutant := range products.mutants {
		{
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
		}
	}

}
