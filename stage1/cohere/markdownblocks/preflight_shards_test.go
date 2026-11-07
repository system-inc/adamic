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

	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

const testWholeDocumentOraclePreflightShards = 4

type preflightProducts struct {
	binary, fork, script string
	mutants              []string
	planted              string
}

func prepareWholeDocumentPreflight(test *testing.T, root string, budget *preflightBudget) *preflightProducts {
	cohere := filepath.Join(root, "cohere")
	binary, mutants := oracles.PreflightGoProducts(test)
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	for _, name := range []string{
		"standalone.js",
		"plugins/estree.js",
		"plugins/typescript.js",
		"plugins/babel.js",
		"plugins/postcss.js",
		"plugins/markdown.js",
		"plugins/graphql.js",
		"plugins/yaml.js",
	} {
		original, readError := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles", name))
		if readError != nil {
			test.Fatal(readError)
		}
		installed, readError := os.ReadFile(filepath.Join(fork, name))
		if readError != nil {
			test.Fatal(readError)
		}
		equal(test, "pinned fork bundle "+name, installed, original)
	}
	script, absolutePathError := filepath.Abs("testdata/library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return &preflightProducts{binary: binary, mutants: mutants, fork: fork, script: script}
}

func runWholeDocumentPreflightShard(
	test *testing.T,
	budget *preflightBudget,
	root string,
	inputs []auditInput,
	files, shard int,
	products *preflightProducts,
) {
	dir := test.TempDir()
	var batch bytes.Buffer
	encoder := json.NewEncoder(&batch)
	for _, input := range inputs {
		if encodeError := encoder.Encode(input); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(dir, "cases.jsonl")
	write(test, cases, batch.Bytes())
	binary := products.binary
	goRun := preflightExecute(test, budget, nil, binary, cases)
	goAnswers := auditResults(test, "Go", goRun)
	fork, script := products.fork, products.script
	nodeRun := preflightExecute(test, budget, nil, "node", script, fork, cases, "fork")
	nodeAnswers := auditResults(test, "pinned fork", nodeRun)
	if len(goAnswers) != len(inputs) || len(nodeAnswers) != len(inputs) {
		test.Fatal("oracle lost a document")
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
	witnessFile, readError := os.ReadFile("GAPS.md")
	if readError != nil {
		test.Fatal(readError)
	}
	var witnessStrings []string
	for _, line := range strings.Split(string(witnessFile), "\n") {
		if !strings.HasPrefix(line, `"`) {
			continue
		}
		var value string
		if decodeError := json.Unmarshal([]byte(line), &value); decodeError != nil {
			test.Fatalf("witness JSON: %v", decodeError)
		}
		witnessStrings = append(witnessStrings, value)
	}
	witnessCount := 0
	for index, actual := range goAnswers {
		expected := nodeAnswers[index]
		if actual.Name != inputs[index].Name || expected.Name != actual.Name {
			test.Fatal("oracle reordered documents")
		}
		if preflightOffComparisonError := preflightOffComparison(actual, expected); preflightOffComparisonError != nil {
			test.Fatal(preflightOffComparisonError)
		}
		if actual.Name == products.planted {
			planted := expected
			planted.Off += "planted disagreement"
			if preflightOffComparison(actual, planted) == nil {
				test.Fatal("planted failure escaped comparison")
			}
			test.Logf("planted failure caught by shard-%03d: %s", shard, actual.Name)
		}
		if actual.Auto != expected.Auto {
			if _, known := gaps[actual.Name]; !known {
				test.Fatalf("new auto-mode discrepancy at %s", actual.Name)
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
						test.Fatalf("GAPS.md lost complete output for %s", actual.Name)
					}
				}
				digest := fmt.Sprintf("%x", sha256.Sum256([]byte(inputs[index].Text)))
				if !strings.Contains(string(witnessFile), "`"+digest+"`") {
					test.Fatalf("GAPS.md lost input digest for %s", actual.Name)
				}
			}
			offset := firstDifference(actual.Auto, expected.Auto)
			test.Logf("known auto gap %s at byte %d, embeds %v", actual.Name, offset, actual.Embeds)
		}
	}
	if witnessCount != expectedWitnesses || len(witnessStrings) != 14 {
		test.Fatalf(
			"expected %d witnesses and fourteen complete outputs, got %d and %d",
			expectedWitnesses,
			witnessCount,
			len(witnessStrings),
		)
	}
	for name, seen := range gaps {
		if !seen {
			test.Fatalf("auto gap closed or witness missing: %s; update GAPS.md", name)
		}
	}
	if keep := os.Getenv("ADAMIC_MARKDOWNBLOCKS_KEEP"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		write(test, filepath.Join(keep, fmt.Sprintf("shard-%03d-cases.jsonl", shard)), batch.Bytes())
		write(test, filepath.Join(keep, fmt.Sprintf("shard-%03d-go.jsonl", shard)), goRun.stdout)
		write(test, filepath.Join(keep, fmt.Sprintf("shard-%03d-fork.jsonl", shard)), nodeRun.stdout)
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
			answer := preflightExecute(test, budget, nil, side.command, side.args...)
			elapsed += time.Since(start)
			answers := auditResults(test, side.name, answer)
			if len(answers) != len(inputs) {
				test.Fatal("measured baseline lost a document")
			}
			for index, actual := range answers {
				if actual.Name != inputs[index].Name {
					test.Fatal("measured baseline reordered documents")
				}
				equal(test, side.name+": "+actual.Name, []byte(actual.Off), []byte(goAnswers[index].Off))
			}
		}
		test.Logf(
			"baseline throughput %s %.1f documents/s, 3 runs %.6fs; one off-mode format per document, startup/I/O included, no Adamic block formatter measured",
			side.name,
			float64(len(inputs)*3)/elapsed.Seconds(),
			elapsed.Seconds(),
		)
	}
	test.Logf(
		"corpus %d physical repository/submodule Markdown files, %d generated documents, %d total; off byte-identical, auto %d/%d identical with %d named disagreements",
		files,
		len(inputs)-files,
		len(inputs),
		len(inputs)-len(gaps),
		len(inputs),
		len(gaps),
	)
}

func preflightOffComparison(actual, expected auditOutput) error {
	if !bytes.Equal([]byte(actual.Off), []byte(expected.Off)) {
		return fmt.Errorf(
			"embedding off: %s differs at byte %d",
			actual.Name,
			firstDifference(actual.Off, expected.Off),
		)
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
	return int(
		uint32(digest[0])<<24|uint32(digest[1])<<16|uint32(digest[2])<<8|uint32(digest[3]),
	) % testWholeDocumentOraclePreflightShards
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

func wholeDocumentPreflight(test *testing.T) *wholeDocumentPreflightFixture {
	test.Helper()
	wholeDocumentPreflightOnce.Do(func() {
		started := time.Now()
		budget := newPreflightBudget(test, "TestWholeDocumentOraclePreflight_Setup")
		defer budget.close()
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		inputs, files := auditCorpus(test, root)
		selection := selectMarkdownFiles(test, root, inputs, files, 1)
		if selection.Sample {
			test.Log(selection.Log(test.Name()))
		}
		inputs, files = selectedMarkdownInputs(inputs, files, selection)
		fixture := &wholeDocumentPreflightFixture{
			root:     root,
			inputs:   inputs,
			products: prepareWholeDocumentPreflight(test, root, budget),
		}
		for index, input := range inputs {
			shard := preflightShard(input.Name)
			fixture.shards[shard] = append(fixture.shards[shard], input)
			if index < files {
				fixture.physical[shard]++
			}
		}
		fixture.products.planted = inputs[0].Name
		wholeDocumentPreflightShared = fixture
		test.Logf("TestWholeDocumentOraclePreflight_Setup (setup) %.6fs", time.Since(started).Seconds())
	})
	if wholeDocumentPreflightShared == nil {
		test.Fatal("preflight setup failed")
	}
	return wholeDocumentPreflightShared
}

// TestProduct_WholeDocumentPreflightGoProducts builds the Go printer and its three mutants that every preflight unit
// reads (markdownblocks-whole-document-preflight-go-products, 75 s measured inside a unit on Oct 10), so Workshop
// builds them once and a unit only reads them. A build has a product's ceiling, not a unit's 90 s.

func TestWholeDocumentOraclePreflightUnion(test *testing.T) {
	test.Parallel()
	fixture := wholeDocumentPreflight(test)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(test, test.Name())
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
				test.Fatal("unstable shard assignment")
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
				test.Logf("planted failure caught by shard-%03d: %s", shard, input.Name)
			}
		}
	}
	if len(fixture.shards) != testWholeDocumentOraclePreflightShards ||
		count != len(fixture.inputs) ||
		len(actual) != len(expected) {
		test.Fatal("shard union lost a document")
	}
	for name, value := range expected {
		if value != 1 || actual[name] != 1 {
			test.Fatalf("shard union count for %s: %d", name, actual[name])
		}
	}
	for name := range preflightRequiredGaps() {
		if expected[name] != 1 {
			test.Fatalf("auto gap witness missing: %s", name)
		}
	}
	if caught != 1 {
		test.Fatal("planted disagreement did not reach exactly one shard")
	}
	test.Logf("union %d documents across %d shards", count, len(fixture.shards))
}
func wholeDocumentPreflightLeaf(test *testing.T, shard int) {
	test.Helper()
	fixture := wholeDocumentPreflight(test)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(test, test.Name())
	defer budget.close()
	runWholeDocumentPreflightShard(
		test,
		budget,
		fixture.root,
		fixture.shards[shard],
		fixture.physical[shard],
		shard,
		fixture.products,
	)
}
func TestWholeDocumentOraclePreflight_000(test *testing.T) {
	test.Parallel()
	wholeDocumentPreflightLeaf(test, 0)
}
func TestWholeDocumentOraclePreflight_001(test *testing.T) {
	test.Parallel()
	wholeDocumentPreflightLeaf(test, 1)
}
func TestWholeDocumentOraclePreflight_002(test *testing.T) {
	test.Parallel()
	wholeDocumentPreflightLeaf(test, 2)
}
func TestWholeDocumentOraclePreflight_003(test *testing.T) {
	test.Parallel()
	wholeDocumentPreflightLeaf(test, 3)
}

// Mutation witnesses are a property of the entire corpus, rather than every
// arbitrary hash bucket. This leaf retains the original full-corpus must-fail check.
func TestWholeDocumentOraclePreflightMutants(test *testing.T) {
	test.Parallel()
	fixture := wholeDocumentPreflight(test)
	// Shared preparation is complete before this leaf receives its full budget.
	budget := newPreflightBudget(test, test.Name())
	defer budget.close()
	products, inputs := fixture.products, fixture.inputs
	var batch bytes.Buffer
	for _, input := range inputs {
		if encodeError := json.NewEncoder(&batch).Encode(input); encodeError != nil {
			test.Fatal(encodeError)
		}
	}
	cases := filepath.Join(test.TempDir(), "cases.jsonl")
	write(test, cases, batch.Bytes())
	nodeAnswers := auditResults(
		test,
		"pinned fork",
		preflightExecute(test, budget, nil, "node", products.script, products.fork, cases, "fork"),
	)
	if len(nodeAnswers) != len(inputs) {
		test.Fatal("oracle lost a document")
	}
	for index, answer := range nodeAnswers {
		if answer.Name != inputs[index].Name {
			test.Fatal("oracle reordered documents")
		}
	}
	for _, mutant := range products.mutants {
		{
			answers := auditResults(test, "Go mutant", preflightExecute(test, budget, nil, mutant, cases))
			if len(answers) != len(inputs) {
				test.Fatal("mutant lost a document")
			}
			caught := false
			for index, actual := range answers {
				if actual.Name != inputs[index].Name {
					test.Fatal("mutant reordered documents")
				}
				if actual.Off != nodeAnswers[index].Off {
					test.Logf(
						"output-only mutant caught by %s at byte %d",
						actual.Name,
						firstDifference(actual.Off, nodeAnswers[index].Off),
					)
					caught = true
					break
				}
			}
			if !caught {
				test.Fatal("mutant survived byte comparison")
			}
		}
	}

}

// Setup owns an independent 90-second budget. Each case leaf starts a fresh
// budget only after wholeDocumentPreflight returns, including standalone runs.
type preflightBudget struct {
	label  string
	limit  time.Duration
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
	groups sync.Map
}

func newPreflightBudget(test *testing.T, label string) *preflightBudget {
	test.Helper()
	return newPreflightBudgetOf(test, label, 90*time.Second)
}

// newPreflightBudgetOf is a hard deadline of limit: 90 s for a unit, a product's 600 s (less a margin) for a build.
func newPreflightBudgetOf(test *testing.T, label string, limit time.Duration) *preflightBudget {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), limit)
	budget := &preflightBudget{ctx: timeoutContext, cancel: cancel, label: label, limit: limit}
	budget.timer = time.AfterFunc(limit, budget.cooked)
	return budget
}
func (budget *preflightBudget) cooked() {
	budget.cancel()
	budget.groups.Range(func(pid, _ any) bool { _ = syscall.Kill(-pid.(int), syscall.SIGKILL); return true })
	fmt.Fprintf(os.Stderr, "cooked: %s exceeded %.0fs hard deadline\n", budget.label, budget.limit.Seconds())
	os.Exit(124)
}
func (budget *preflightBudget) close() { budget.timer.Stop(); budget.cancel() }
func preflightCommand(test *testing.T, budget *preflightBudget, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	command := exec.CommandContext(budget.ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	// No WaitDelay: exec reads both streams to their end, where one closed them a second after the child exited and
	// failed a run that had finished but was slow to drain under load (#n3gk22h's shape). Not childguard.Capture, which
	// has no Started: cooked kills the groups preflightRun registers before it exits the process, and a group it
	// couldn't register would outlive the exit. The Go oracle and Node start no process that could hold a pipe open
	// after they exit, and the budget's hard deadline ends a wait on one that did.
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
	if startError := command.Start(); startError != nil {
		return startError
	}
	budget.groups.Store(command.Process.Pid, true)
	defer budget.groups.Delete(command.Process.Pid)
	return command.Wait()
}
func preflightExecute(
	test *testing.T,
	budget *preflightBudget,
	environment []string,
	name string,
	arguments ...string,
) run {
	test.Helper()
	command := preflightCommand(test, budget, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	preflightRunError := preflightRun(budget, command)
	var exitError *exec.ExitError
	if preflightRunError != nil && !errors.As(preflightRunError, &exitError) {
		test.Fatalf("running %s: %v", name, preflightRunError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}
