package markdownblocks

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testWholeDocumentOraclePreflightShards = 4

type preflightProducts struct {
	binary, fork, script string
	mutants              []string
	planted              string
}

func prepareWholeDocumentPreflight(t *testing.T, root string, budget *preflightBudget) *preflightProducts {
	driver, err := filepath.Abs("testdata/go_driver.go")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_markdown_blocks/main.go")
	mutations := []struct{ name, path, from, to string }{
		{"heading prefix", "heading.go", `strings.Repeat("#", currentNode(path).Depth)`, `strings.Repeat("#", currentNode(path).Depth+1)`},
		{"list marker", "list.go", `rawPrefix = "- "`, `rawPrefix = "+ "`},
		{"fence length", "print.go", `max(3, getMaxContinuousCount(value, "` + "`" + `")+1)`, `max(4, getMaxContinuousCount(value, "` + "`" + `")+1)`},
	}
	alteredSources := make([][]byte, len(mutations))
	for index, mutation := range mutations {
		content, err := os.ReadFile(filepath.Join(cohere, "internal/format/markdown", mutation.path))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(content), mutation.from) != 1 {
			t.Fatal("mutant must alter exactly one site")
		}
		alteredSources[index] = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
	}
	buildFiles := []string{"go.mod", "go.work", "cohere", "stage1/cohere/markdownblocks/testdata/go_driver.go", "stage1/cohere/markdownblocks/preflight_shards_test.go"}
	for _, name := range []string{"go.sum", "go.work.sum"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			buildFiles = append(buildFiles, name)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	product := buildcache.Product(t, buildcache.Inputs{
		Name:      "markdownblocks-whole-document-preflight-go-products",
		Files:     buildFiles,
		Flags:     []string{"go build -overlay -o cmd/adamic_markdown_blocks/main.go"},
		Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "GOFLAGS", "GOTOOLCHAIN", "GOWORK")},
	}, func(dir string) error {
		build := func(replacements map[string]string, binary string) error {
			replacements[mainPath] = driver
			overlay, err := json.Marshal(map[string]any{"Replace": replacements})
			if err != nil {
				return err
			}
			path := binary + ".overlay.json"
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			command := preflightCommand(t, budget, "go", "build", "-overlay="+path, "-o", binary, mainPath)
			command.Dir = cohere
			var output bytes.Buffer
			command.Stdout, command.Stderr = &output, &output
			if err := preflightRun(budget, command); err != nil {
				return fmt.Errorf("build: %w\n%s", err, output.Bytes())
			}
			return nil
		}
		if err := build(map[string]string{}, filepath.Join(dir, "go-printer")); err != nil {
			return err
		}
		for index, mutation := range mutations {
			altered := filepath.Join(dir, mutation.path)
			if err := os.WriteFile(altered, alteredSources[index], 0644); err != nil {
				return err
			}
			source := filepath.Join(cohere, "internal/format/markdown", mutation.path)
			if err := build(map[string]string{source: altered}, filepath.Join(dir, fmt.Sprintf("mutant-%d", index))); err != nil {
				return err
			}
		}
		return nil
	})
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
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
	products := &preflightProducts{binary: filepath.Join(product, "go-printer"), fork: fork, script: script}
	for index := range mutations {
		products.mutants = append(products.mutants, filepath.Join(product, fmt.Sprintf("mutant-%d", index)))
	}
	return products
}

func runWholeDocumentPreflightShard(t *testing.T, budget *preflightBudget, root string, inputs []auditInput, files, shard int, products *preflightProducts) {
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
	goRun := preflightExecute(t, budget, nil, binary, cases)
	goAnswers := auditResults(t, "Go", goRun)
	fork, script := products.fork, products.script
	nodeRun := preflightExecute(t, budget, nil, "node", script, fork, cases, "fork")
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
			answer := preflightExecute(t, budget, nil, side.command, side.args...)
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

// Products and the live corpus are prepared once, with an independent setup
// deadline. Cached binaries outlive any leaf that uses them.
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
		started := time.Now()
		budget := newPreflightBudget(t, "TestWholeDocumentOraclePreflight_Setup")
		defer budget.close()
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
		fixture := &wholeDocumentPreflightFixture{root: root, inputs: inputs, products: prepareWholeDocumentPreflight(t, root, budget)}
		for index, input := range inputs {
			shard := preflightShard(input.Name)
			fixture.shards[shard] = append(fixture.shards[shard], input)
			if index < files {
				fixture.physical[shard]++
			}
		}
		fixture.products.planted = inputs[0].Name
		wholeDocumentPreflightShared = fixture
		t.Logf("TestWholeDocumentOraclePreflight_Setup (setup) %.6fs", time.Since(started).Seconds())
	})
	if wholeDocumentPreflightShared == nil {
		t.Fatal("preflight setup failed")
	}
	return wholeDocumentPreflightShared
}
func TestWholeDocumentOraclePreflightUnion(t *testing.T) {
	t.Parallel()
	fixture := wholeDocumentPreflight(t)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(t, t.Name())
	defer budget.close()
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
	fixture := wholeDocumentPreflight(t)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(t, t.Name())
	defer budget.close()
	runWholeDocumentPreflightShard(t, budget, fixture.root, fixture.shards[shard], fixture.physical[shard], shard, fixture.products)
}
func TestWholeDocumentOraclePreflight_000(t *testing.T) {
	t.Parallel()
	wholeDocumentPreflightLeaf(t, 0)
}
func TestWholeDocumentOraclePreflight_001(t *testing.T) {
	t.Parallel()
	wholeDocumentPreflightLeaf(t, 1)
}
func TestWholeDocumentOraclePreflight_002(t *testing.T) {
	t.Parallel()
	wholeDocumentPreflightLeaf(t, 2)
}
func TestWholeDocumentOraclePreflight_003(t *testing.T) {
	t.Parallel()
	wholeDocumentPreflightLeaf(t, 3)
}

// Mutation witnesses are a property of the entire corpus, rather than every
// arbitrary hash bucket. This leaf retains the original full-corpus must-fail check.
func TestWholeDocumentOraclePreflightMutants(t *testing.T) {
	t.Parallel()
	fixture := wholeDocumentPreflight(t)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(t, t.Name())
	defer budget.close()
	products, inputs := fixture.products, fixture.inputs
	var batch bytes.Buffer
	for _, input := range inputs {
		if err := json.NewEncoder(&batch).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	cases := filepath.Join(t.TempDir(), "cases.jsonl")
	write(t, cases, batch.Bytes())
	nodeAnswers := auditResults(t, "pinned fork", preflightExecute(t, budget, nil, "node", products.script, products.fork, cases, "fork"))
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
			answers := auditResults(t, "Go mutant", preflightExecute(t, budget, nil, mutant, cases))
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

// Setup owns an independent 90-second budget. Each case leaf starts a fresh
// budget only after wholeDocumentPreflight returns, including standalone runs.
type preflightBudget struct {
	label  string
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
	groups sync.Map
}

func newPreflightBudget(t *testing.T, label string) *preflightBudget {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	budget := &preflightBudget{ctx: ctx, cancel: cancel, label: label}
	budget.timer = time.AfterFunc(90*time.Second, budget.cooked)
	return budget
}
func (budget *preflightBudget) cooked() {
	budget.cancel()
	budget.groups.Range(func(pid, _ any) bool { _ = syscall.Kill(-pid.(int), syscall.SIGKILL); return true })
	fmt.Fprintf(os.Stderr, "cooked: %s exceeded 90s hard deadline\n", budget.label)
	os.Exit(124)
}
func (budget *preflightBudget) close() { budget.timer.Stop(); budget.cancel() }
func preflightCommand(t *testing.T, budget *preflightBudget, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.CommandContext(budget.ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}
func preflightRun(budget *preflightBudget, command *exec.Cmd) error {
	// CommandContext can return just before the hard timer fires. Preserve the
	// cooked classification in that race rather than reporting an oracle failure.
	defer func() {
		if errors.Is(budget.ctx.Err(), context.DeadlineExceeded) {
			budget.cooked()
		}
	}()
	if err := command.Start(); err != nil {
		return err
	}
	budget.groups.Store(command.Process.Pid, true)
	defer budget.groups.Delete(command.Process.Pid)
	return command.Wait()
}
func preflightExecute(t *testing.T, budget *preflightBudget, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := preflightCommand(t, budget, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := preflightRun(budget, command)
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}
