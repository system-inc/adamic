// Package oracles owns tsprinter's compiler-free oracle products and the helpers both test binaries share: the
// Go oracle builds, the pinned corpus, the Go-oracle output corpora, the npm and embedded Prettier shards, the TSC
// manifest and the mutant Go-oracle corpora (#2ba0c20). Relative paths resolve from stage1/cohere/tsprinter: the parent
// tests run there, and this package's TestMain changes its working directory to that parent directory.
package oracles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

var repositoryCorpusRoots = []string{
	"bench", "bridge", "cmd", "internal", "stage1", "stage3/drivers/tsc/corpus",
}

// repositoryCorpusCommit pins the repository's half of the printer corpus: every .ts file under repositoryCorpusRoots
// as it was at this commit on main, read from the checkout's objects, never from the working tree. The printers'
// agreement needs a large TypeScript corpus, not a current one, so a .ts edit anywhere in the repository moves no
// tsprinter product; walked live, one edit to lint's main.ts moved all 162 (#wy8gcke, Planner on #rybwgtz). The corpus
// moves only when this pin does. The checkout that builds the corpus product must hold the commit, as json's does.
const RepositoryCorpusCommit = "6f7d5d0ba4a7ccafb6ed69fd531d169f9effe443"

// tsPrinterPinnedCorpusInputs keys the pinned corpus by its commit and roots: a different pin, or another root, is
// another product.
func TSPrinterPinnedCorpusInputs(pin string) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "tsprinter pinned corpus",
		Files: []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
		Flags: []string{"repository TypeScript at " + pin, "roots " + strings.Join(repositoryCorpusRoots, " ")},
	}
}

// inRepositoryCorpus says whether a path of the pinned tree is a corpus file: a .ts file under one of the roots.
func inRepositoryCorpus(name string) bool {
	if !strings.HasSuffix(name, ".ts") {
		return false
	}
	for _, root := range repositoryCorpusRoots {
		if strings.HasPrefix(name, root+"/") {
			return true
		}
	}
	return false
}

// corpusListName is the product's list of its files, one repository path a line, in the pinned tree's order.
const corpusListName = "corpus.list"

// buildTSPrinterPinnedCorpus writes each corpus file at the pin under the product, at its repository path, and the list
// of them. Every byte comes through buildcache.PinnedFiles as key-time work, so the build's read set holds no file of
// the working tree.
func buildTSPrinterPinnedCorpus(inputs buildcache.Inputs, pin string) func(directory string) error {
	return func(directory string) error {
		root, absolutePathError := filepath.Abs(Repository)
		if absolutePathError != nil {
			return absolutePathError
		}
		files, pinnedFilesError := buildcache.PinnedFiles(root, inputs, pin, "TypeScript", inRepositoryCorpus)
		if pinnedFilesError != nil {
			return pinnedFilesError
		}
		var list strings.Builder
		for _, file := range files {
			path := filepath.Join(directory, filepath.FromSlash(file.Path))
			if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
				return createDirectoryError
			}
			if writeError := os.WriteFile(path, file.Data, 0o644); writeError != nil {
				return writeError
			}
			list.WriteString(file.Path + "\n")
		}
		return os.WriteFile(filepath.Join(directory, corpusListName), []byte(list.String()), 0o644)
	}
}

// pinnedCorpus is the pinned corpus product, fetched once per process: its directory, which case labels name as repo/
// as they name the checkout (tsPrinterLabelRoots, read by parallel tests that never fetched it, so atomic), and its
// files there, in order.
var pinnedCorpus struct {
	once      sync.Once
	directory atomic.Pointer[string]
	files     []string
	err       error
}

func PinnedCorpusFiles(test *testing.T) []string {
	test.Helper()
	pinnedCorpus.once.Do(func() {
		inputs := oracleInputs(TSPrinterPinnedCorpusInputs(RepositoryCorpusCommit))
		directory, getError := buildcache.Get(
			inputs,
			oracleBuild(inputs.Name, buildTSPrinterPinnedCorpus(inputs, RepositoryCorpusCommit)),
		)
		if getError != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", getError)
			return
		}
		list, readError := os.ReadFile(filepath.Join(directory, corpusListName))
		if readError != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", readError)
			return
		}
		for _, name := range strings.Split(strings.TrimSuffix(string(list), "\n"), "\n") {
			pinnedCorpus.files = append(pinnedCorpus.files, filepath.Join(directory, filepath.FromSlash(name)))
		}
		pinnedCorpus.directory.Store(&directory)
	})
	if pinnedCorpus.err != nil {
		test.Fatal(pinnedCorpus.err)
	}
	return append([]string(nil), pinnedCorpus.files...)
}

// tsPrinterLabelRoots are the directories a case label may name a file under, each with the name a stored product
// spells it by: the checkout and the pinned corpus (the repository at its pin) as repo/, the TypeScript source as
// typescript/. Read back (canonical false), repo/ is the checkout, which comes first.
func tsPrinterLabelRoots(root, external string, canonical bool) [][2]string {
	roots := readingLabelRoots(root, external)
	if directory := pinnedCorpus.directory.Load(); canonical && directory != nil {
		roots = slices.Insert(roots, 1, [2]string{*directory, "repo"})
	}
	return roots
}

// readingLabelRoots are tsPrinterLabelRoots read back (canonical false), which never names the pinned corpus: so a
// shards build (statementShardsBuild) keys its cases with no package state of this one.
func readingLabelRoots(root, external string) [][2]string {
	return [][2]string{{root, "repo"}, {external, "typescript"}}
}

func TrackedRootFiles(test *testing.T, checkout, root string) ([]string, error) {
	test.Helper()
	info, err := os.Stat(filepath.Join(checkout, root))
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("corpus root %s: missing directory (%v)", root, err)
	}
	// git ls-files where the checkout has .git, else the manifest a Loom runner's unpacked source carries. The listing
	// is key-time work (buildcache.Keying): every product that reads the corpus keys each listed .ts file itself
	// (statementCorpusBuild by an input= flag with its sha256, tsPrinterOracleProduct as Files, tscAgreementPrepare by
	// its declared corpus directory), so a .ts added or changed moves those keys. Walked under the trace, the listing
	// put every directory under bench, bridge, cmd, internal and stage1 into their read sets, and any file added there,
	// a css gap's or a Go file's, moved tsprinter's keys and every product downstream of them (#0rx19ha).
	var paths []string
	buildcache.Keying(func() { paths, err = tracked.PathsUnder(checkout, root) })
	if err != nil {
		return nil, fmt.Errorf("corpus root %s: tracked files: %w", root, err)
	}
	var files []string
	for _, name := range paths {
		if strings.HasSuffix(name, ".ts") {
			files = append(files, filepath.Join(checkout, name))
		}
	}
	// Without .git, a .ts file the manifest leaves out (and no .gitignore ignores) is a stale manifest, refused: a
	// check, never a byte a build reads, so it runs as key-time work too.
	var extra []string
	buildcache.Keying(func() { extra, err = tracked.Unrecorded(checkout, root) })
	if err != nil {
		return nil, fmt.Errorf("corpus root %s: %w", root, err)
	}
	for _, name := range extra {
		if strings.HasSuffix(name, ".ts") {
			return nil, fmt.Errorf("corpus root %s: %s is on disk but not in the tree's manifest", root, name)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("corpus root %s: no git-tracked .ts files", root)
	}
	test.Logf("corpus root %s: %d git-tracked .ts files", root, len(files))
	return files, nil
}

func printerCorpusFiles(test *testing.T) []string {
	test.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Skip(
			"set ADAMIC_TYPESCRIPT_SOURCE to a TypeScript 6.0.3 source checkout at 050880ce; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	// The pin is a check, never an input: tracked.Head asks git as key-time work, with the home's configuration out of
	// reach (home-6's dry run, Oct 11: this git read ~/.gitconfig inside the build, refusing it and the Go oracle
	// beside it).
	if head, headError := tracked.Head(source); headError != nil || head != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		test.Fatalf("TypeScript pin: %q %v", head, headError)
	}
	// The repository's half is the pinned corpus product's files, never the working tree's (repositoryCorpusCommit).
	files := PinnedCorpusFiles(test)
	listed, trackedRootFilesError := TrackedRootFiles(test, source, "src/compiler")
	if trackedRootFilesError != nil {
		test.Fatal(trackedRootFilesError)
	}
	return append(files, listed...)
}

func DocumentCorpus(test *testing.T) (string, string, string) {
	test.Helper()
	oracle := DocumentOracle(test)
	product := PrinterGateOnce(test, "document-corpus", func() string {
		return oracleProduct(test, buildcache.Inputs{
			Name:      "tsprinter-documents-oracle-output-v1",
			Files:     []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
			Flags:     []string{"Go oracle sha256=" + statementFileHash(test, oracle)},
			Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error {
			cohere, _ := filepath.Abs(Repository + "/cohere")
			// The oracle runs in its package's directory and writes docs.txt, answers.txt and docs.json into dir; the
			// documents are generated, so the product names no path.
			command := toolchain.Command(oracle, "-test.run=^TestAdamicDocuments$")
			command.Dir = cohere + "/internal/format/doc"
			command.Env = append(os.Environ(), "ADAMIC_TS_DOC_OUTPUT="+dir)
			if output, combinedOutputError := childguard.CombinedOutput(
				command,
				childguard.Options{},
			); combinedOutputError != nil {
				return fmt.Errorf("Go doc corpus %v\n%s", combinedOutputError, output)
			}
			return nil
		})
	})
	// A private copy per test: PlanCorpus keys its plan by path and forgets it when the test ends (#g7a9yd3).
	directory := test.TempDir()
	for _, name := range []string{"docs.txt", "answers.txt", "docs.json"} {
		data, readError := os.ReadFile(filepath.Join(product, name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), data, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	answers, readError := os.ReadFile(directory + "/answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	PlanCorpus(test, directory+"/docs.txt")
	return directory + "/docs.txt", string(answers), directory + "/docs.json"
}

// DocumentOracle is cohere's doc package with the harness laid over it, a product built ahead (#5qykzj5).
func DocumentOracle(test *testing.T) string {
	test.Helper()
	return goTestOracle(
		test,
		"TestProduct_TSPrinterDocumentOracle",
		"doc.test",
		"./internal/format/doc",
		documentOracleArguments(test),
		nil,
	)
}

// documentOracleInputs is DocumentOracle's key, computed without building it.
func documentOracleInputs(test *testing.T) buildcache.Inputs {
	test.Helper()
	inputs, goTestInputsError := buildcache.GoTestInputs(
		"cohere",
		"doc.test",
		"./internal/format/doc",
		documentOracleArguments(test),
		nil,
	)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	return inputs
}

// documentOracleArguments lay the harness over cohere's doc package.
func documentOracleArguments(test *testing.T) []string {
	test.Helper()
	cohere, _ := filepath.Abs(Repository + "/cohere")
	side, _ := filepath.Abs("testdata/doc_side_test.go")
	overlay, _ := json.Marshal(
		map[string]any{"Replace": map[string]string{cohere + "/internal/format/doc/adamic_docs_test.go": side}},
	)
	path := test.TempDir() + "/overlay.json"
	if writeError := os.WriteFile(path, overlay, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return []string{"-overlay=" + path}
}

func ExpressionGateKeys(test *testing.T, cases []PrinterCase) []string {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	upstream, absolutePathError := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keys, statementCaseKeysError := StatementCaseKeys(cases, root, upstream)
	if statementCaseKeysError != nil {
		test.Fatal(statementCaseKeysError)
	}
	return expressionCaseKeys(keys)
}

// expressionCaseKeys are statement case keys as expressions key them: generated expression cases are owned by their
// actual generator file.
func expressionCaseKeys(keys []string) []string {
	for i, key := range keys {
		keys[i] = strings.Replace(key, "testdata/statements_side_test.go:", "testdata/expressions_side_test.go:", 1)
	}
	return keys
}

func ExpressionGateOwners(keys []string, count int) [][]int {
	shards := make([][]int, count)
	for id, key := range keys {
		digest := sha256.Sum256([]byte(key))
		var value uint64
		for _, b := range digest[:8] {
			value = value<<8 | uint64(b)
		}
		owner := int(value % uint64(count))
		shards[owner] = append(shards[owner], id)
	}
	return shards
}

func expressionGatePartition(test *testing.T, cases []PrinterCase, count int) [][]int {
	test.Helper()
	partition, expressionPartitionError := expressionPartition(ExpressionGateKeys(test, cases), count)
	if expressionPartitionError != nil {
		test.Fatal(expressionPartitionError)
	}
	return partition
}

// expressionPartition is ExpressionGateOwners over keys, each stable key once.
func expressionPartition(keys []string, count int) ([][]int, error) {
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			return nil, fmt.Errorf("repeated stable expression key %q", key)
		}
		seen[key] = true
	}
	return ExpressionGateOwners(keys, count), nil
}

// The executable and every ordered input are inputs to corpus generation.
// Only reference answers are cached; every tested printer still runs each case.
func tsPrinterOracleProduct(t *testing.T, root string, files []string, gaps, family, oracle string) string {
	t.Helper()
	test, variable := "TestAdamicExpressionCorpus", "ADAMIC_TS_EXPRESSION_REQUEST"
	if family == "statements" {
		test, variable = "TestAdamicStatementCorpus", "ADAMIC_TS_STATEMENT_REQUEST"
	}
	external := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if external != "" {
		var absolutePathError error
		external, absolutePathError = filepath.Abs(external)
		if absolutePathError != nil {
			t.Fatal(absolutePathError)
		}
	}
	flags := []string{"test-run=" + test, "metadata=canonical-v1"}
	for _, file := range files {
		flags = append(flags, "input-id="+tsPrinterOutputLabel(file, root, external, true))
	}
	inputs := append([]string{oracle, gaps, root + "/stage1/cohere/tsprinter/oracles/oracles.go"}, files...)
	product := oracleExpressionBuild(t, PrinterBuildInputs{
		Name:      "TS " + family + " oracle outputs",
		Files:     inputs,
		Flags:     flags,
		Toolchain: runtime.Version(),
	}, func(directory string) error {
		request, encodeError := json.Marshal(map[string]any{"Files": files, "Directory": directory, "Gaps": gaps})
		if encodeError != nil {
			return encodeError
		}
		// Requests name this builder's paths and must never become shared outputs.
		requestPath := filepath.Join(t.TempDir(), "request.json")
		if writeError := os.WriteFile(requestPath, request, 0644); writeError != nil {
			return writeError
		}
		command := toolchain.Command(oracle, "-test.v", "-test.count=1", "-test.run=^"+test+"$", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), variable+"="+requestPath)
		output, combinedOutputError := childguard.CombinedOutput(command, childguard.Options{})
		if combinedOutputError != nil {
			return fmt.Errorf("Go %s corpus: %w\n%s", family, combinedOutputError, output)
		}
		t.Log(string(output))
		return TSPrinterRewriteMetadata(directory, root, external, true)
	})
	return product
}

func tsPrinterOracleOutputs(test *testing.T, root string, files []string, gaps, family, oracle string) string {
	test.Helper()
	external := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if external != "" {
		var absolutePathError error
		external, absolutePathError = filepath.Abs(external)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
	}
	product := tsPrinterOracleCorpus(test, root, files, gaps, family, oracle)
	// Shard preparation and caller-requested audits use a private writable copy.
	directory := test.TempDir()
	names := []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"}
	if family == "expressions" {
		names = append(names, "gaps.txt", "gap-answers.txt", "gaps.json")
	}
	for _, name := range names {
		data, readError := os.ReadFile(filepath.Join(product, name))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if tsPrinterRewriteMetadataError := TSPrinterRewriteMetadata(
		directory,
		root,
		external,
		false,
	); tsPrinterRewriteMetadataError != nil {
		test.Fatal(tsPrinterRewriteMetadataError)
	}
	return directory
}

// tsPrinterOracleCorpus is family's Go oracle outputs product over files, once per process: its labels are canonical,
// repo/ and typescript/ (TSPrinterRewriteMetadata), where tsPrinterOracleOutputs' copy names this checkout's paths.
func tsPrinterOracleCorpus(test *testing.T, root string, files []string, gaps, family, oracle string) string {
	test.Helper()
	return PrinterGateOnce(
		test,
		"oracle-output-"+family+"-"+StatementBytesHash([]byte(strings.Join(files, "\n"))),
		func() string { return tsPrinterOracleProduct(test, root, files, gaps, family, oracle) },
	)
}

func tsPrinterOutputLabel(label, root, external string, canonical bool) string {
	return outputLabel(label, tsPrinterLabelRoots(root, external, canonical), canonical)
}

// outputLabel is label spelled under the first of roots it names, as tsPrinterOutputLabel says.
func outputLabel(label string, roots [][2]string, canonical bool) string {
	for _, pair := range roots {
		if pair[0] == "" {
			continue
		}
		from, to := pair[0]+"/", pair[1]+"/"
		if !canonical {
			from, to = to, from
		}
		if strings.HasPrefix(label, from) {
			return to + strings.TrimPrefix(label, from)
		}
	}
	return label
}

// Rewrite path metadata alone. Source and Want may contain the same strings,
// and changing those would silently change the independent oracle's answers.
func TSPrinterRewriteMetadata(directory, root, external string, canonical bool) error {
	path := filepath.Join(directory, "cases.json")
	data, readError := os.ReadFile(path)
	if readError != nil {
		return readError
	}
	var cases []PrinterCase
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		return decodeError
	}
	for i := range cases {
		cases[i].Label = tsPrinterOutputLabel(cases[i].Label, root, external, canonical)
	}
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		return encodeError
	}
	if writeError := os.WriteFile(path, data, 0644); writeError != nil {
		return writeError
	}
	path = filepath.Join(directory, "coverage.json")
	data, readError = os.ReadFile(path)
	if readError != nil {
		return readError
	}
	var report map[string]json.RawMessage
	if decodeError := json.Unmarshal(data, &report); decodeError != nil {
		return decodeError
	}
	if raw, exists := report["file_refusals"]; exists {
		var refusals map[string]string
		if decodeError := json.Unmarshal(raw, &refusals); decodeError != nil {
			return decodeError
		}
		rewritten := make(map[string]string, len(refusals))
		var replacements []string
		for _, pair := range tsPrinterLabelRoots(root, external, canonical) {
			if pair[0] == "" {
				continue
			}
			from, to := pair[0]+"/", pair[1]+"/"
			if !canonical {
				from, to = to, from
			}
			replacements = append(replacements, from, to)
		}
		replace := strings.NewReplacer(replacements...)
		for label, diagnostic := range refusals {
			rewritten[tsPrinterOutputLabel(label, root, external, canonical)] = replace.Replace(diagnostic)
		}
		report["file_refusals"], encodeError = json.Marshal(rewritten)
		if encodeError != nil {
			return encodeError
		}
	}
	data, encodeError = json.MarshalIndent(report, "", "  ")
	if encodeError != nil {
		return encodeError
	}
	return os.WriteFile(path, data, 0644)
}

func ExpressionUnits(test *testing.T, path, want string, count int) (*CorpusPlan, [][]byte) {
	test.Helper()
	text, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	data, readError := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if readError != nil {
		test.Fatal(readError)
	}
	var cases []PrinterCase
	if decodeError := json.Unmarshal(data, &cases); decodeError != nil {
		test.Fatal(decodeError)
	}
	rows := bytes.Split(bytes.TrimSuffix(text, []byte("\n")), []byte("\n"))
	answers := bytes.Split(bytes.TrimSuffix([]byte(want), []byte("\n")), []byte("\n"))
	if len(rows) != len(cases) || len(answers) != len(cases) || len(cases) == 0 {
		test.Fatalf("expression enumeration: %d cases, %d rows, %d answers", len(cases), len(rows), len(answers))
	}
	// Every pinned upstream record must occur in the unsplit enumeration. Each
	// unit then checks the exact independently pinned outcome for its own cases.
	for _, record := range UpstreamOutcomes(test) {
		if record.Family != "expressions" {
			continue
		}
		found := false
		for _, item := range cases {
			if strings.HasSuffix(item.Label, record.Label) && item.Source == record.Source && item.Want == record.Go {
				found = true
				break
			}
		}
		if !found {
			test.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	plan := &CorpusPlan{Labels: make([]string, len(cases))}
	for i, item := range cases {
		plan.Labels[i] = fmt.Sprintf("case-%06d %s", i, item.Label)
	}
	directory := test.TempDir()
	var outputs [][]byte
	partition := expressionGatePartition(test, cases, count)
	for number, indices := range partition {
		shard := CorpusShard{
			Indices: indices,
			Text:    filepath.Join(directory, fmt.Sprintf("shard-%03d.txt", number)),
			Specs:   filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number)),
		}
		var protocol, expected bytes.Buffer
		subset := make([]PrinterCase, 0, len(indices))
		for _, i := range indices {
			protocol.Write(rows[i])
			protocol.WriteByte('\n')
			expected.Write(answers[i])
			expected.WriteByte('\n')
			subset = append(subset, cases[i])
		}
		encoded, encodeError := json.Marshal(subset)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(shard.Text, protocol.Bytes(), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		if writeError := os.WriteFile(shard.Specs, encoded, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		plan.Shards = append(plan.Shards, shard)
		outputs = append(outputs, expected.Bytes())
	}
	if expressionUnionError := ExpressionUnion(plan, outputs, []byte(want)); expressionUnionError != nil {
		test.Fatal(expressionUnionError)
	}
	gapData, readError := os.ReadFile(filepath.Dir(path) + "/gaps.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var gaps []PrinterCase
	if decodeError := json.Unmarshal(gapData, &gaps); decodeError != nil {
		test.Fatal(decodeError)
	}
	gapAnswers, readError := os.ReadFile(filepath.Dir(path) + "/gap-answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	if len(gaps) != bytes.Count(gapAnswers, []byte("\n")) {
		test.Fatal("gap enumeration/answer count differs")
	}
	whole := &CorpusPlan{
		Labels: append([]string(nil), plan.Labels...),
		Shards: append([]CorpusShard(nil), plan.Shards...),
	}
	gapIDs := make([]int, len(gaps))
	for i, item := range gaps {
		gapIDs[i] = len(cases) + i
		whole.Labels = append(whole.Labels, fmt.Sprintf("case-%06d gap %s", gapIDs[i], item.Label))
	}
	whole.Shards = append(whole.Shards, CorpusShard{Indices: gapIDs})
	wholeOutputs := append(append([][]byte(nil), outputs...), gapAnswers)
	if expressionUnionError := ExpressionUnion(
		whole,
		wholeOutputs,
		append([]byte(want), gapAnswers...),
	); expressionUnionError != nil {
		test.Fatal(expressionUnionError)
	}
	test.Logf(
		"union: %d unique expression case ids in %d units; %d gap ids in one unit; total %d",
		len(cases),
		len(plan.Shards),
		len(gaps),
		len(cases)+len(gaps),
	)
	return plan, outputs
}

// The plan as every unit of one process shares it. Each unit's protocol and
// specifications are held as bytes: the files expressionUnits wrote belong to
// the unit that built the plan and are removed when that unit ends.
type expressionGateUnits struct {
	Plan                       *CorpusPlan
	texts, specs, Answers      [][]byte
	gapText, gapSpecs, GapWant []byte
	builder                    string
	// corpus is the expression corpus product the plan was made from, canonical (tsPrinterOracleCorpus), which the
	// expression shards product is built from.
	corpus string
}

var expressionGateUnitsShared struct {
	once  sync.Once
	units *expressionGateUnits
}

// The first selected unit in a process builds the plan through expressionUnits,
// with every whole-corpus check it makes: each pinned upstream record present,
// every stable key unique and both unions exact. The plan is a pure function of
// the corpus product, so later units in the process read it rather than repeat
// it. TestExpressionsAgainstGoAndPrettierUnion still builds and checks its own.
func ExpressionGateUnitsOnce(test *testing.T) *expressionGateUnits {
	test.Helper()
	// printerCorpusFiles skips without a TypeScript source. Skip here, before the
	// once, so every unit skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" {
		printerCorpusFiles(test)
	}
	shared := &expressionGateUnitsShared
	shared.once.Do(func() {
		cases, want, _ := ExpressionCorpus(test, false)
		plan, answers := ExpressionUnits(test, cases, want, ExpressionsAgainstGoAndPrettierShards-1)
		units := &expressionGateUnits{
			Plan:    &CorpusPlan{Labels: plan.Labels},
			Answers: answers,
			builder: test.Name(),
			corpus:  expressionCorpusProduct(test),
		}
		for _, shard := range plan.Shards {
			text, readError := os.ReadFile(shard.Text)
			if readError != nil {
				test.Fatal(readError)
			}
			specs, readError := os.ReadFile(shard.Specs)
			if readError != nil {
				test.Fatal(readError)
			}
			units.Plan.Shards = append(units.Plan.Shards, CorpusShard{Indices: shard.Indices})
			units.texts, units.specs = append(units.texts, text), append(units.specs, specs)
		}
		directory := filepath.Dir(cases)
		for _, file := range []struct {
			name string
			data *[]byte
		}{{"gaps.txt", &units.gapText}, {"gaps.json", &units.gapSpecs}, {"gap-answers.txt", &units.GapWant}} {
			data, readError := os.ReadFile(filepath.Join(directory, file.name))
			if readError != nil {
				test.Fatal(readError)
			}
			*file.data = data
		}
		shared.units = units
	})
	if shared.units == nil {
		test.Fatal("expression unit preparation failed")
	}
	if shared.units.builder != test.Name() {
		test.Logf("union: checked once in this process by %s", shared.units.builder)
	}
	return shared.units
}

// A unit's own protocol and specifications, written where its children read them.
func (units *expressionGateUnits) Unit(test *testing.T, number int) CorpusShard {
	test.Helper()
	directory := test.TempDir()
	shard := CorpusShard{
		Indices: units.Plan.Shards[number].Indices,
		Text:    filepath.Join(directory, fmt.Sprintf("shard-%03d.txt", number)),
		Specs:   filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number)),
	}
	if writeError := os.WriteFile(shard.Text, units.texts[number], 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(shard.Specs, units.specs[number], 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return shard
}

// The loud gap unit's inputs. The Prettier script knows its gap proof by the name gaps.json.
func (units *expressionGateUnits) Gaps(test *testing.T) string {
	test.Helper()
	directory := test.TempDir()
	for name, data := range map[string][]byte{"gaps.txt": units.gapText, "gaps.json": units.gapSpecs} {
		if writeError := os.WriteFile(filepath.Join(directory, name), data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return directory
}

func ExpressionUnion(plan *CorpusPlan, outputs [][]byte, want []byte) error {
	// Position is part of the case id: generated cases may share a label.
	seen := make(map[int]bool, len(plan.Labels))
	count := 0
	for number, shard := range plan.Shards {
		for _, id := range shard.Indices {
			if id < 0 || id >= len(plan.Labels) || seen[id] {
				return fmt.Errorf("shard-%03d: missing, extra or repeated case id %d", number, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(plan.Labels) {
		return fmt.Errorf("union: %d case ids, want %d", count, len(plan.Labels))
	}
	for id := range plan.Labels {
		if !seen[id] {
			return fmt.Errorf("union: missing case id %d", id)
		}
	}
	merged, mergeCorpusError := MergeCorpus(plan, outputs)
	if mergeCorpusError != nil {
		return mergeCorpusError
	}
	if !bytes.Equal(merged, want) {
		return fmt.Errorf("union differs from unsplit answers")
	}
	return nil
}

func ExpressionCorpus(test *testing.T, processPlan ...bool) (string, string, string) {
	test.Helper()
	root, _ := filepath.Abs(Repository)
	files := printerCorpusFiles(test)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(test, root, files, gaps, "expressions", PrinterGateGoOracle(test)+"/oracle")
	answers, readError := os.ReadFile(directory + "/answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	if keep := os.Getenv("ADAMIC_TS_PRINTER_KEEP"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, name := range []string{
			"coverage.json",
			"cases.json",
			"cases.txt",
			"answers.txt",
			"gaps.txt",
			"gap-answers.txt",
			"gaps.json",
		} {
			data, readError := os.ReadFile(directory + "/" + name)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(keep+"/"+name, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	if len(processPlan) == 0 || processPlan[0] {
		PlanCorpus(test, directory+"/cases.txt")
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

// expressionCorpusProduct is the product ExpressionCorpus copies, its labels canonical.
func expressionCorpusProduct(test *testing.T) string {
	test.Helper()
	root, _ := filepath.Abs(Repository)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	return tsPrinterOracleCorpus(
		test,
		root,
		printerCorpusFiles(test),
		gaps,
		"expressions",
		PrinterGateGoOracle(test)+"/oracle",
	)
}

const ExpressionsAgainstGoAndPrettierShards = 65

// expressionLibraryOracle is a case shard's npm (side 0) or embedded (side 1) Prettier answers, the product
// statementLibraryOracle makes for a statement shard and keyed the same way. The gap unit's proof still runs live.
func ExpressionLibraryOracle(test *testing.T, number, side int, shard CorpusShard, libraryHash, library string) string {
	test.Helper()
	return PrinterGateOnce(test, fmt.Sprintf("expression-library-%d-%d", number, side), func() string {
		return libraryOracle(test, "expressions", number, side, shard.Specs, libraryHash, library, expressionShards)
	})
}

// Inputs describe non-Go read-only shared products; Go builds stay separate.
type PrinterBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

// ExpressionBuild builds inputs under expressionKey's key, once per process: the compiler products' way, which keep
// their keys. An oracle product goes through oracleExpressionBuild.
func ExpressionBuild(test *testing.T, inputs PrinterBuildInputs, build func(dir string) error) string {
	test.Helper()
	return expressionBuildKeyed(test, expressionKey(test, inputs), build)
}

// oracleExpressionBuild is ExpressionBuild for an oracle product: its name says oracle-owner-v1, and only this
// package's binary builds it under a trace (oracleProduct).
func oracleExpressionBuild(test *testing.T, inputs PrinterBuildInputs, build func(dir string) error) string {
	test.Helper()
	key := oracleInputs(expressionKey(test, inputs))
	return expressionBuildKeyed(test, key, oracleBuild(key.Name, build))
}

func expressionBuildKeyed(test *testing.T, key buildcache.Inputs, build func(dir string) error) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	digest, keyError := buildcache.Key(root, key)
	if keyError != nil {
		test.Fatal(keyError)
	}
	return PrinterGateOnce(test, "expression-product-"+digest, func() string {
		return buildcache.Product(test, key, build)
	})
}

// expressionKey is the key expressionBuild builds inputs under: each repository file by its path, and a generated input
// outside the repository by its bytes.
func expressionKey(test *testing.T, inputs PrinterBuildInputs) buildcache.Inputs {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	key := buildcache.Inputs{
		Name:      inputs.Name,
		Flags:     append([]string{}, inputs.Flags...),
		Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	for _, file := range inputs.Files {
		relative, relativePathError := filepath.Rel(root, file)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			data, readError := os.ReadFile(file)
			if readError != nil {
				test.Fatal(readError)
			}
			key.Flags = append(key.Flags, fmt.Sprintf("generated:%s:%x", filepath.Base(file), sha256.Sum256(data)))
		} else {
			key.Files = append(key.Files, filepath.ToSlash(relative))
		}
	}
	key.Flags = append(key.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	return key
}

const Repository = "../../.."

type Run struct {
	Stdout, Stderr []byte
	ExitCode       int
}

// The mutation witnesses are generated families, fixed by their generator's
// seeds and loops. Repository files cannot change their membership. Reuse the
// independent Go oracle while avoiding parsing/formatting unrelated inputs:
// mutantOracleBinary is that oracle with its add hook filtered to change's
// witness families, resolved once by the caller (binary) and run here on an
// empty request.
func mutantOracle(test *testing.T, change Mutation, binary string) (string, string) {
	test.Helper()
	directory := test.TempDir()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	gaps, absolutePathError := filepath.Abs("testdata/notyet.json")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	request, encodeError := json.Marshal(map[string]any{"Files": []string{}, "Directory": directory, "Gaps": gaps})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(directory+"/request.json", request, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	kind := "Expression"
	if change.Entry == "statementsMain.ts" {
		kind = "Statement"
	}
	// The test binary runs in its package's directory, as go test runs it.
	command := toolchain.Command(binary, "-test.count=1", "-test.timeout=0", "-test.run=^TestAdamic"+kind+"Corpus$")
	command.Dir = root + "/cohere/internal/format/javascript"
	command.Env = append(
		os.Environ(),
		"ADAMIC_TS_EXPRESSION_REQUEST="+directory+"/request.json",
		"ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json",
	)
	if output, combinedOutputError := childguard.CombinedOutput(
		command,
		childguard.Options{},
	); combinedOutputError != nil {
		test.Fatalf("Go mutation oracle: %v\n%s", combinedOutputError, output)
	}
	answers, readError := os.ReadFile(directory + "/answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	return directory + "/cases.txt", string(answers)
}

// mutantOracleAnchor is the side files' add hook, which mutantOracleBinary filters to a mutation's witness families.
const mutantOracleAnchor = "add := func(label, source string) { cases = append(cases, portExpressionCase{Label: label, Source: source}) }"

// mutantOracleBinary is change's Go oracle: cohere's internal/format/javascript test binary with the expression side
// file laid over it (and the statement one, for a statement mutation), each add hook filtered to change's witness
// families. It is a buildcache.GoTestMutated product, keyed by the files of the packages go compiles, the overlay's map
// and each planted filter, so it never walks cohere, whose TypeScript submodule a populated tree hashed for every
// lookup (#4ycgm2n). The filter was a scratch copy of the side file, which no product can key.
func mutantOracleBinary(test *testing.T, change Mutation) string {
	test.Helper()
	arguments, filters := mutantOracleRequest(test, change)
	return goTestOracle(
		test,
		"TestProduct_TSPrinterMutantOracle_*",
		"oracle",
		"./internal/format/javascript",
		arguments,
		filters,
	)
}

// mutantOracleInputs is mutantOracleBinary's key, computed without building it.
func mutantOracleInputs(test *testing.T, change Mutation) buildcache.Inputs {
	test.Helper()
	arguments, filters := mutantOracleRequest(test, change)
	inputs, goTestInputsError := buildcache.GoTestInputs(
		"cohere",
		"oracle",
		"./internal/format/javascript",
		arguments,
		filters,
	)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	return inputs
}

// mutantOracleRequest is how mutantOracleBinary builds change's oracle: the overlay argument and each planted filter.
func mutantOracleRequest(test *testing.T, change Mutation) ([]string, []buildcache.Mutation) {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	families, ok := MutantFamilies[change.Name]
	if !ok && change.Name != "hashbang loses its refusal" {
		test.Fatalf("missing witness families for %q", change.Name)
	}
	allowed := "map[string]bool{"
	for _, family := range families {
		allowed += fmt.Sprintf("%q:true,", family)
	}
	allowed += "}"
	replace := map[string]string{}
	var filters []buildcache.Mutation
	for _, name := range []string{"expressions", "statements"} {
		if name == "statements" && change.Entry != "statementsMain.ts" {
			continue
		}
		side := "stage1/cohere/tsprinter/testdata/" + name + "_side_test.go"
		replace[root+"/cohere/internal/format/javascript/adamic_"+name+"_test.go"] = filepath.Join(root, side)
		filters = append(filters, buildcache.Mutation{File: side, Before: mutantOracleAnchor, After: "allowed := " +
			allowed +
			"\nadd := func(label, source string) { if allowed[label] { cases = append(cases, portExpressionCase{Label: label, Source: source}) } }"})
	}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	// The overlay names this checkout's paths: GoTestMutated keys its map and the side files, never its own path.
	path := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return []string{"-overlay=" + path}, filters
}

// A leaf and its product declaration share both the recipe and its key. The
// once belongs to one build product, never to a top-level setup test. Products
// contain immutable outputs; temporary Node sources belong to the caller.
type MutantProductState struct {
	Once      sync.Once
	Directory string
}

var mutantOracleProducts [MutantsShards]MutantProductState

func MutantOracleProduct(test *testing.T, index int) (string, string) {
	test.Helper()
	state := &mutantOracleProducts[index]
	state.Once.Do(func() {
		root, absolutePathError := filepath.Abs(Repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		// The corpus is the oracle's answers: the oracle is a product (mutantOracleBinary, or DocumentOracle for the
		// doc mutation), a deterministic build of what go compiles, so it stands in this key by its own key
		// (mutantOracleKey), and only a build of the corpus resolves the binary. A reader keys the corpus without
		// holding an oracle of about 53 MB (Tests, Oct 12). Then the gaps file, and the code that runs it and captures
		// its output. Not cohere whole: its TypeScript submodule made a populated tree hash 67,000 files per lookup, up
		// to 58 times a cold process (#4ycgm2n).
		state.Directory = oracleExpressionBuild(test, MutantOracleCorpusInputs(
			root,
			index,
			MutantOracleKey(test, root, index),
		), func(directory string) error {
			var cases string
			if Mutations[index].Entry == "docMain.ts" {
				cases, _, _ = DocumentCorpus(test)
			} else {
				cases, _ = mutantOracle(test, Mutations[index], mutantOracleBinary(test, Mutations[index]))
			}
			// Requests and overlays contain private paths and are never published.
			names := []string{filepath.Base(cases), "answers.txt", "cases.json"}
			if Mutations[index].Entry == "docMain.ts" {
				names = []string{"docs.txt", "answers.txt", "docs.json"}
			}
			if Mutations[index].Name == "hashbang loses its refusal" {
				names = append(names, "gaps.txt", "gap-answers.txt")
			}
			for _, name := range names {
				data, readError := os.ReadFile(filepath.Join(filepath.Dir(cases), name))
				if readError != nil {
					return readError
				}
				if writeError := os.WriteFile(filepath.Join(directory, name), data, 0644); writeError != nil {
					return writeError
				}
			}
			return nil
		})
	})
	if state.Directory == "" {
		test.Fatal("oracle product preparation failed")
	}
	name := "cases.txt"
	if Mutations[index].Entry == "docMain.ts" {
		name = "docs.txt"
	}
	answers, readError := os.ReadFile(filepath.Join(state.Directory, "answers.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	return filepath.Join(state.Directory, name), string(answers)
}

// mutantOracleCorpusInputs are mutation index's oracle corpus inputs under root: the oracle by its product's key, the
// gaps file, and the code that runs the oracle and captures its output.
func MutantOracleCorpusInputs(root string, index int, oracle string) PrinterBuildInputs {
	inputs := []string{root + "/internal/childguard"}
	for _, file := range []string{"oracles/oracles.go", "testdata/notyet.json"} {
		inputs = append(inputs, root+"/stage1/cohere/tsprinter/"+file)
	}
	return PrinterBuildInputs{
		Name:  fmt.Sprintf("TS printer mutant %03d Go oracle corpus", index),
		Files: inputs,
		Flags: []string{
			"oracle " + oracle,
			fmt.Sprintf("mutation=%+v", Mutations[index]),
			fmt.Sprintf("families=%q", MutantFamilies[Mutations[index].Name]),
			"oracle-width=80",
		},
		Toolchain: runtime.Version(),
	}
}

// mutantOracleKey is the key of mutation index's oracle product (DocumentOracle for the doc mutation, else
// mutantOracleBinary), computed without building or fetching it.
func MutantOracleKey(test *testing.T, root string, index int) string {
	test.Helper()
	// Only the doc mutation keys the doc oracle, so its package never joins another mutation's reads.
	var inputs buildcache.Inputs
	if Mutations[index].Entry == "docMain.ts" {
		inputs = documentOracleInputs(test)
	} else {
		inputs = mutantOracleInputs(test, Mutations[index])
	}
	key, keyError := buildcache.Key(root, inputs)
	if keyError != nil {
		test.Fatal(keyError)
	}
	return key
}

// Mutation tests need witnesses, while the agreement tests exercise every
// corpus input on unchanged ports. Select independent Go-oracle cases from
// the relevant generated families, keeping runs small even as tracked sources
// grow. Evenly spaced cases retain different operators, widths and contexts.
var MutantFamilies = map[string][]string{
	"program loses its statement separator":  {"program-sequence", "program-boundary"},
	"declaration loses its function keyword": {"program-sequence", "program-boundary"},
	"assignment chain ignores statement boundaries": {
		"statement-composition-regression",
		"variable-statement-edge",
	},
	"variable statement loses its keyword": {
		"variable-statement-composition",
		"variable-statement-edge",
	},
	"await loses its keyword":                             {"await-composition", "await-yield-edge"},
	"yield loses delegation":                              {"yield-composition", "await-yield-edge"},
	"tagged template loses its tag":                       {"tagged-template-composition"},
	"template preview loses optional stopping boundaries": {"template-optional-boundary-composition"},
	"computed key forgets its clean text":                 {"computed-short-key-boundary"},
	"method loses its key":                                {"object-method-composition", "object-accessor-composition"},
	"function loses its keyword":                          {"named-function-composition"},
	"optional chain loses its stopping parentheses":       {"optional-chain-boundary"},
	"return and throw lose their keyword":                 {"named-function-composition", "arrow-composition"},
	"statement expression loses its semicolon": {
		"named-function-composition",
		"statement-composition-regression",
	},
	"body loses its opening brace":            {"named-function-composition", "arrow-composition"},
	"simple statement loses debugger":         {"named-function-composition", "arrow-composition"},
	"arrow loses its arrow token":             {"arrow-composition", "arrow-chain"},
	"template loses its interpolation dollar": {"template-interpolation", "template-width"},
	"member chain loses its method dot":       {"member-call-composition", "member-call-width"},
	"arguments lose their trailing comma":     {"expanded-arguments", "argument-width"},
	"object loses its property colon":         {"object-composition", "object-width"},
	"conditional swaps its separator":         {"conditional-composition", "conditional-nesting"},
	"assignment loses its operator":           {"assignment-composition", "assignment-chain"},
	"sequence loses a comma":                  {"sequence-composition", "sequence-width", "sequence"},
	"expression loses required parentheses":   {"operator-pair", "edge"},
}

type Mutation struct{ Name, File, From, To, Entry string }

var Mutations = []Mutation{
	{
		"protocol width confuses Number with parseFloat",
		"docMain.ts",
		"printWidth: Number(fields[1] ?? '')",
		"printWidth: Number.parseFloat(fields[1] ?? '')",
		"docMain.ts",
	},
	{
		"program loses its statement separator",
		"expressions.ts",
		"if(parts.length > 0) parts.push(this.docs.hardline());\n                    const previous =",
		"if(false) parts.push(this.docs.hardline());\n                    const previous =",
		"statementsMain.ts",
	},
	{
		"declaration loses its function keyword",
		"expressions.ts",
		"`${async}function${generator}${name}`",
		"`${async}${node.kind === 'FunctionDeclaration' ? '' : 'function'}${generator}${name}`",
		"statementsMain.ts",
	},
	{
		"assignment chain ignores statement boundaries",
		"expressions.ts",
		"!['ExpressionStatement', 'VariableDeclarationList', 'VariableStatement']",
		"!['UnrelatedStatement', 'VariableDeclarationList', 'VariableStatement']",
		"main.ts",
	},
	{
		"hashbang loses its refusal",
		"expressions.ts",
		"source.startsWith('#!') || source.startsWith('\\ufeff#!') || ",
		"",
		"main.ts",
	},
	{
		"variable statement loses its keyword",
		"expressions.ts",
		"this.docs.text(kind),",
		"this.docs.text(''),",
		"main.ts",
	},
	{"await loses its keyword", "expressions.ts", "this.docs.text('await ')", "this.docs.text('void ')", "main.ts"},
	{
		"yield loses delegation",
		"expressions.ts",
		"delegated ? 'yield*' : 'yield'",
		"delegated ? 'yield' : 'yield'",
		"main.ts",
	},
	{
		"tagged template loses its tag",
		"expressions.ts",
		"this.print(this.child(id, 0), id, 'tag')",
		"this.docs.text('')",
		"main.ts",
	},
	{
		"template preview loses optional stopping boundaries",
		"expressions.ts",
		"for(const boundary of this.optionalBoundaries) flatPrinter.optionalBoundaries.add(boundary);",
		"for(const boundary of this.optionalBoundaries) flatPrinter.sequenceBoundaries.add(boundary);",
		"main.ts",
	},
	{
		"computed key forgets its clean text",
		"expressions.ts",
		"const keyText = this.docs.textContent(left);",
		"const keyText = this.docs.get(left).kind === 'text' ? this.docs.get(left).text : undefined;",
		"main.ts",
	},
	{"method loses its key", "expressions.ts", "this.propertyKey(index, offset)", "this.docs.text(\"\")", "main.ts"},
	{
		"function loses its keyword",
		"expressions.ts",
		"`${async}function${generator}${name}`",
		"`${async}${generator}${name}`",
		"main.ts",
	},
	{
		"optional chain loses its stopping parentheses",
		"parentheses.ts",
		"if(optionalBoundaries.has(index) && parent >= 0)",
		"if(false && parent >= 0)",
		"main.ts",
	},
	{
		"return and throw lose their keyword",
		"expressions.ts",
		"this.docs.text(node.kind === 'ReturnStatement' ? 'return' : 'throw')",
		"this.docs.text(node.kind === 'ReturnStatement' ? 'throw' : 'throw')",
		"main.ts",
	},
	{
		"statement expression loses its semicolon",
		"expressions.ts",
		"return this.docs.concat([printed, this.docs.text(';')]);",
		"return this.docs.concat([printed, this.docs.text('')]);",
		"main.ts",
	},
	{
		"body loses its opening brace",
		"expressions.ts",
		"this.docs.text('{'),\n                    this.docs.indent(",
		"this.docs.text('['),\n                    this.docs.indent(",
		"main.ts",
	},
	{
		"simple statement loses debugger",
		"expressions.ts",
		"return this.docs.text('debugger;');",
		"return this.docs.text(';');",
		"main.ts",
	},
	{
		"arrow loses its arrow token",
		"expressions.ts",
		"this.docs.text(' =>'),\n                chain ?",
		"this.docs.text(' '),\n                chain ?",
		"main.ts",
	},
	{
		"template loses its interpolation dollar",
		"expressions.ts",
		"this.docs.text('${')",
		"this.docs.text('{')",
		"main.ts",
	},
	{
		"member chain loses its method dot",
		"expressions.ts",
		"this.docs.text(this.ownOptional(index) ? '?.' : '.')",
		"this.docs.text(this.ownOptional(index) ? '?.' : '')",
		"main.ts",
	},
	{
		"arguments lose their trailing comma",
		"expressions.ts",
		"const trailing = this.docs.ifBreak(this.docs.text(','), this.docs.text(''));",
		"const trailing = this.docs.text('');",
		"main.ts",
	},
	{"object loses its property colon", "expressions.ts", "this.docs.text(':')", "this.docs.text('')", "main.ts"},
	{"conditional swaps its separator", "expressions.ts", "this.docs.text('? ')", "this.docs.text(': ')", "main.ts"},
	{
		"assignment loses its operator",
		"expressions.ts",
		"this.docs.text(` ${this.operator(index)}`)",
		"this.docs.text(' ')",
		"main.ts",
	},
	{
		"sequence loses a comma",
		"expressions.ts",
		"parts.push(this.docs.text(','));",
		"parts.push(this.docs.text(''));",
		"main.ts",
	},
	{
		"group ignores width",
		"doc.ts",
		"!document.broken &&\n                        this.fits(flat",
		"!document.broken ||\n                        this.fits(flat",
		"docMain.ts",
	},
	{"fill never packs a pair", "doc.ts", "separatorMode = fitsPair ? 2 : 1;", "separatorMode = 1;", "docMain.ts"},
	{
		"expression loses required parentheses",
		"expressions.ts",
		"return this.parenthesize(id, parent, role)",
		"return parent < -1 && this.parenthesize(id, parent, role)",
		"main.ts",
	},
}

const MutantsShards = 29

func printerGateTSCCorpus(test *testing.T, family string) {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	files, trackedRootFilesError := TrackedRootFiles(test, root, "stage3/drivers/tsc/corpus")
	if trackedRootFilesError != nil {
		test.Fatal(trackedRootFilesError)
	}
	gaps, _ := filepath.Abs("testdata/notyet.json")
	tsPrinterOracleOutputs(test, root, files, gaps, family, PrinterGateGoOracle(test)+"/oracle")
}

// The partition and the npm library's hash are the process's, made once; each
// product writes only its own shard's files.
func StatementDeclaredLibraryOracle(test *testing.T, number, side int) {
	test.Helper()
	shard := StatementGateShardsOnce(test).Shard(test, number)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		test.Skip("set ADAMIC_TS_PRETTIER")
	}
	StatementLibraryOracle(test, number, side, shard, StatementLibraryHashOnce(test), library)
}

func ExpressionDeclaredLibraryOracle(test *testing.T, number, side int) {
	test.Helper()
	units := ExpressionGateUnitsOnce(test)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		test.Skip("set ADAMIC_TS_PRETTIER")
	}
	ExpressionLibraryOracle(
		test,
		number,
		side,
		units.Unit(test, number),
		StatementDirectoryHash(test, library),
		library,
	)
}

type printerGateProductState struct {
	once      sync.Once
	directory string
}

var printerGateProducts sync.Map

// Only process-local initialization is held here. Every persistent product is
// owned, keyed and published by internal/buildcache.
func PrinterGateOnce(test *testing.T, name string, build func() string) string {
	test.Helper()
	value, _ := printerGateProducts.LoadOrStore(name, &printerGateProductState{})
	state := value.(*printerGateProductState)
	state.once.Do(func() { state.directory = build() })
	if state.directory == "" {
		test.Fatalf("product %s preparation failed", name)
	}
	return state.directory
}

// printerGateGoOracle is the directory holding the Go printer oracle, cohere's internal/format/javascript test binary
// with the two side files laid over it: a buildcache.GoTest product, keyed by the files of the packages go compiles,
// the overlay's map and the side files' content, and go's settings. Its key declared all of cohere, the TypeScript
// submodule's 41,000 files included, which cost a populated tree's first lookup 15 to 60 s (#4ycgm2n).
func PrinterGateGoOracle(test *testing.T) string {
	test.Helper()
	return PrinterGateOnce(test, "go-oracle", func() string {
		root, absolutePathError := filepath.Abs(Repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		replace := map[string]string{}
		for _, family := range []string{"expressions", "statements"} {
			replace[root+
				"/cohere/internal/format/javascript/adamic_"+
				family+
				"_test.go"] = root + "/stage1/cohere/tsprinter/testdata/" + family + "_side_test.go"
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		// The overlay names this checkout's paths: GoTest keys its map and the side files' content, never its own path.
		path := filepath.Join(test.TempDir(), "overlay.json")
		if writeError := os.WriteFile(path, overlay, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		return filepath.Dir(goTestOracle(
			test,
			"TestProduct_TSPrinterGoOracle",
			"oracle",
			"./internal/format/javascript",
			[]string{"-overlay=" + path},
			nil,
		))
	})
}

type CorpusShard struct {
	Indices     []int
	Text, Specs string
}

type CorpusPlan struct {
	Labels []string
	Shards []CorpusShard
}

var corpusPlans sync.Map

// Source bytes approximate parser and printer work. Ties use the original case
// number and then the shard number, making the assignment reproducible.
func AssignCorpus(weights []int, count int) [][]int {
	count = min(count, len(weights))
	if count == 0 {
		return nil
	}
	order := make([]int, len(weights))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		if weights[order[i]] != weights[order[j]] {
			return weights[order[i]] > weights[order[j]]
		}
		return order[i] < order[j]
	})
	result, totals := make([][]int, count), make([]int, count)
	for _, index := range order {
		shard := 0
		for i := 1; i < count; i++ {
			if totals[i] < totals[shard] {
				shard = i
			}
		}
		result[shard] = append(result[shard], index)
		totals[shard] += max(weights[index], 1)
	}
	return result
}

func PlanCorpus(test *testing.T, path string) *CorpusPlan {
	test.Helper()
	if cached, ok := corpusPlans.Load(path); ok {
		return cached.(*CorpusPlan)
	}
	base := strings.TrimSuffix(path, filepath.Ext(path))
	textPath, specsPath := base+".txt", base+".json"
	text, readError := os.ReadFile(textPath)
	if readError != nil {
		test.Fatal(readError)
	}
	data, readError := os.ReadFile(specsPath)
	if readError != nil {
		test.Fatal(readError)
	}
	var specs []json.RawMessage
	if decodeError := json.Unmarshal(data, &specs); decodeError != nil {
		test.Fatal(decodeError)
	}
	var rows []string
	if filepath.Base(base) == "docs" {
		for _, part := range strings.Split(string(text), "reset\t")[1:] {
			rows = append(rows, "reset\t"+part)
		}
	} else {
		for _, row := range strings.Split(strings.TrimSuffix(string(text), "\n"), "\n") {
			rows = append(rows, row+"\n")
		}
	}
	if len(rows) != len(specs) || len(rows) == 0 {
		test.Fatalf("%s: %d protocol cases, %d specifications", path, len(rows), len(specs))
	}
	plan := &CorpusPlan{Labels: make([]string, len(rows))}
	weights := make([]int, len(rows))
	for i, raw := range specs {
		var item PrinterCase
		if decodeError := json.Unmarshal(raw, &item); decodeError != nil {
			test.Fatal(decodeError)
		}
		plan.Labels[i] = item.Label
		if item.Label == "" {
			plan.Labels[i] = fmt.Sprintf("%s:case %d", specsPath, i)
		}
		weights[i] = len(rows[i])
	}
	directory := test.TempDir()
	// The test's share of the unit, never the host's cores (testguard.Share).
	for number, indices := range AssignCorpus(weights, min(testguard.Share(), 8)) {
		var batch strings.Builder
		subset := make([]json.RawMessage, 0, len(indices))
		for _, index := range indices {
			batch.WriteString(rows[index])
			subset = append(subset, specs[index])
		}
		shard := CorpusShard{
			Indices: indices,
			Text:    filepath.Join(directory, fmt.Sprintf("shard-%d.txt", number)),
			Specs:   filepath.Join(directory, fmt.Sprintf("shard-%d.json", number)),
		}
		if writeError := os.WriteFile(shard.Text, []byte(batch.String()), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		encoded, encodeError := json.Marshal(subset)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(shard.Specs, encoded, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		plan.Shards = append(plan.Shards, shard)
	}
	corpusPlans.Store(path, plan)
	// Either extension resolves to the same immutable plan, retained by the
	// corpus-owning test until all its parallel mutant subtests finish.
	corpusPlans.Store(textPath, plan)
	corpusPlans.Store(specsPath, plan)
	test.Cleanup(func() { corpusPlans.Delete(textPath); corpusPlans.Delete(specsPath) })
	test.Logf("%s: %d cases in %d largest-first processes", filepath.Base(base), len(rows), len(plan.Shards))
	return plan
}

func MergeCorpus(plan *CorpusPlan, outputs [][]byte) ([]byte, error) {
	if len(outputs) != len(plan.Shards) {
		return nil, fmt.Errorf("corpus: %d shards, want %d", len(outputs), len(plan.Shards))
	}
	merged := make([][]byte, len(plan.Labels))
	for number, shard := range plan.Shards {
		output := outputs[number]
		lines := bytes.Split(bytes.TrimSuffix(output, []byte("\n")), []byte("\n"))
		if len(output) == 0 {
			lines = nil
		}
		if len(lines) != len(shard.Indices) {
			index := shard.Indices[min(len(lines), len(shard.Indices)-1)]
			return nil, fmt.Errorf(
				"shard %d %s: missing or extra case: %d answers, want %d",
				number,
				plan.Labels[index],
				len(lines),
				len(shard.Indices),
			)
		}
		if !bytes.HasSuffix(output, []byte("\n")) {
			return nil, fmt.Errorf(
				"shard %d %s: unterminated answer",
				number,
				plan.Labels[shard.Indices[len(shard.Indices)-1]],
			)
		}
		for i, index := range shard.Indices {
			if index < 0 || index >= len(merged) {
				return nil, fmt.Errorf("shard %d: extra case %d", number, index)
			}
			if merged[index] != nil {
				return nil, fmt.Errorf("shard %d %s: repeated case", number, plan.Labels[index])
			}
			merged[index] = lines[i]
		}
	}
	var result bytes.Buffer
	for index, line := range merged {
		if line == nil {
			return nil, fmt.Errorf("%s: missing case", plan.Labels[index])
		}
		result.Write(line)
		result.WriteByte('\n')
	}
	return result.Bytes(), nil
}

var StatementCaseContexts sync.Map

// The partition as every statement test of one process shares it. Each shard's
// protocol and specifications are held as bytes: the files statementShards wrote
// belong to the test that built the partition and are removed when that test ends.
type StatementGateShards struct {
	Cases, Want, Specs    string
	Shards                []statementShard
	texts, specifications [][]byte
	builder               string
}

var statementGateShardsShared struct {
	once   sync.Once
	shards *StatementGateShards
}

// The first statement test in a process partitions the corpus through
// statementShards, with every whole-corpus check it makes: each pinned upstream
// record present, every stable key unique and the union exact. The partition is
// a pure function of the corpus product, so later tests in the process read it
// rather than repeat it: each shard and each of the 32 library oracle products
// re-partitioned all 58,497 cases and rewrote all 16 shard files, 1.0 to 1.9 s
// apiece. TestStatementsAgainstGoAndPrettierUnion still builds and checks its own.
func StatementGateShardsOnce(test *testing.T) *StatementGateShards {
	test.Helper()
	// printerCorpusFiles skips without a TypeScript source. Skip here, before the
	// once, so every test skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" {
		printerCorpusFiles(test)
	}
	shared := &statementGateShardsShared
	shared.once.Do(func() {
		cases, want, specs := StatementCorpus(test)
		gate := &StatementGateShards{Cases: cases, Want: want, Specs: specs, builder: test.Name()}
		for _, shard := range StatementShards(test, cases, want, specs) {
			text, readError := os.ReadFile(shard.Text)
			if readError != nil {
				test.Fatal(readError)
			}
			specifications, readError := os.ReadFile(shard.Specs)
			if readError != nil {
				test.Fatal(readError)
			}
			gate.Shards = append(
				gate.Shards,
				statementShard{Indices: shard.Indices, Labels: shard.Labels, Want: shard.Want},
			)
			gate.texts, gate.specifications = append(gate.texts, text), append(gate.specifications, specifications)
		}
		shared.shards = gate
	})
	if shared.shards == nil {
		test.Fatal("statement shard preparation failed")
	}
	if shared.shards.builder != test.Name() {
		test.Logf("union: checked once in this process by %s", shared.shards.builder)
	}
	return shared.shards
}

// One shard's protocol and specifications, written where its own test reads them.
func (gate *StatementGateShards) Shard(test *testing.T, number int) statementShard {
	test.Helper()
	directory := test.TempDir()
	shard := gate.Shards[number]
	shard.Text = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.txt", number))
	shard.Specs = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.json", number))
	if writeError := os.WriteFile(shard.Text, gate.texts[number], 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if writeError := os.WriteFile(shard.Specs, gate.specifications[number], 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return shard
}

var statementLibraryHashShared struct {
	once sync.Once
	hash string
}

// Every npm library oracle keys by the bytes of the ADAMIC_TS_PRETTIER install,
// which each shard and each library oracle product hashed whole again.
func StatementLibraryHashOnce(test *testing.T) string {
	test.Helper()
	shared := &statementLibraryHashShared
	shared.once.Do(func() { shared.hash = StatementDirectoryHash(test, os.Getenv("ADAMIC_TS_PRETTIER")) })
	if shared.hash == "" {
		test.Fatal("npm Prettier hash failed")
	}
	return shared.hash
}

func statementProduct(test *testing.T, inputs buildcache.Inputs, build func(dir string) error) string {
	test.Helper()
	return oracleProduct(test, inputs, func(dir string) error {
		start, cpu := time.Now(), StatementCPU()
		if buildError := build(dir); buildError != nil {
			return buildError
		}
		test.Logf("build %s: %.3fs, %.3f CPU s", inputs.Name, time.Since(start).Seconds(), StatementCPU()-cpu)
		return nil
	})
}

func StatementBytesHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func statementFileHash(test *testing.T, path string) string {
	test.Helper()
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	return StatementBytesHash(data)
}

// External npm inputs cannot appear in repository-relative Inputs.Files.
// WalkDir visits names lexically; hash every name and byte, refusing symlinks.
func StatementDirectoryHash(test *testing.T, directory string) string {
	test.Helper()
	hash := sha256.New()
	walkError := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("oracle input symlink: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		name, relativePathError := filepath.Rel(directory, path)
		if relativePathError != nil {
			return relativePathError
		}
		data, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		fmt.Fprintf(hash, "%d:%s:%d:", len(name), name, len(data))
		hash.Write(data)
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Linux accounting includes all reaped descendants (oracle, Node and clang).
func StatementCPU() float64 {
	var self, children syscall.Rusage
	if getrusageError := syscall.Getrusage(syscall.RUSAGE_SELF, &self); getrusageError != nil {
		panic(getrusageError)
	}
	if getrusageError := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children); getrusageError != nil {
		panic(getrusageError)
	}
	return float64(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec) +
		float64(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)/1e6
}

// Indices, rather than labels (which repeat in synthetic cases), are the case IDs.
type statementShard struct {
	Indices           []int
	Labels            []string
	Text, Specs, Want string
}

// Keys use a file's relative path, node mode and ordinal within that file/mode.
// Repository growth cannot change another file's assignment. Generated cases
// use their generator file and mode, with an ordinal local to that mode. A label
// names its file under root or upstream, or as the stored product does (#tqrqx60),
// repo/ or typescript/ and the path inside it, and both spellings key alike.
func StatementCaseKeys(items []PrinterCase, root, upstream string) ([]string, error) {
	keys := make([]string, len(items))
	counts := map[string]int{}
	for i, item := range items {
		base := "stage1/cohere/tsprinter/testdata/statements_side_test.go:" + item.Label
		end := strings.LastIndex(item.Label, ":")
		if end >= 0 {
			start := strings.LastIndex(item.Label[:end], ":")
			if start < 0 {
				return nil, fmt.Errorf("invalid statement label %q", item.Label)
			}
			path, absolutePathError := filepath.Abs(
				outputLabel(item.Label[:start], readingLabelRoots(root, upstream), false),
			)
			if absolutePathError != nil {
				return nil, absolutePathError
			}
			relative, relativePathError := filepath.Rel(root, path)
			if relativePathError != nil {
				return nil, relativePathError
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				relative, relativePathError = filepath.Rel(upstream, path)
				if relativePathError != nil {
					return nil, relativePathError
				}
				if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
					return nil, fmt.Errorf("statement outside corpus roots: %s", path)
				}
				relative = "typescript@050880ce59e30b356b686bd3144efe24f875ebc8/" + filepath.ToSlash(relative)
			}
			base = filepath.ToSlash(relative) + ":" + item.Label[end+1:]
		}
		keys[i] = fmt.Sprintf("%s:%d", base, counts[base])
		counts[base]++
	}
	return keys, nil
}

func StatementPartition(keys []string) ([][]int, error) {
	shards := make([][]int, StatementsAgainstGoAndPrettierShards)
	seen := map[string]bool{}
	for id, key := range keys {
		if seen[key] {
			return nil, fmt.Errorf("repeated stable case key %q", key)
		}
		seen[key] = true
		digest := sha256.Sum256([]byte(key))
		var value uint64
		for _, b := range digest[:8] {
			value = value<<8 | uint64(b)
		}
		number := int(value % uint64(StatementsAgainstGoAndPrettierShards))
		shards[number] = append(shards[number], id)
	}
	if statementUnionError := StatementUnion(len(keys), shards); statementUnionError != nil {
		return nil, statementUnionError
	}
	return shards, nil
}

func StatementUnion(count int, shards [][]int) error {
	if len(shards) != StatementsAgainstGoAndPrettierShards {
		return fmt.Errorf("enumerated %d shards, want %d", len(shards), StatementsAgainstGoAndPrettierShards)
	}
	seen := make([]bool, count)
	total := 0
	for number, ids := range shards {
		if len(ids) == 0 {
			return fmt.Errorf("shard-%03d is empty", number)
		}
		for _, id := range ids {
			if id < 0 || id >= count {
				return fmt.Errorf("shard-%03d: extra case ID %d", number, id)
			}
			if seen[id] {
				return fmt.Errorf("shard-%03d: repeated case ID %d", number, id)
			}
			seen[id] = true
			total++
		}
	}
	for id, found := range seen {
		if !found {
			return fmt.Errorf("missing case ID %d", id)
		}
	}
	if total != count {
		return fmt.Errorf("union count %d, want %d", total, count)
	}
	return nil
}

func StatementShards(test *testing.T, cases, want, specs string) []statementShard {
	test.Helper()
	text, readError := os.ReadFile(cases)
	if readError != nil {
		test.Fatal(readError)
	}
	data, readError := os.ReadFile(specs)
	if readError != nil {
		test.Fatal(readError)
	}
	var items []PrinterCase
	if decodeError := json.Unmarshal(data, &items); decodeError != nil {
		test.Fatal(decodeError)
	}
	rows := strings.Split(strings.TrimSuffix(string(text), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(rows) != len(items) || len(answers) != len(items) {
		test.Fatalf("enumeration: %d inputs, %d specifications, %d answers", len(rows), len(items), len(answers))
	}
	// The shared comparator's full-corpus upstream-presence check belongs here;
	// each leaf then checks the exact pinned outcomes for its own cases.
	for _, record := range UpstreamOutcomes(test) {
		if record.Family != "statements" {
			continue
		}
		found := false
		for _, item := range items {
			if strings.HasSuffix(item.Label, record.Label) && item.Source == record.Source {
				if item.Want != record.Go {
					test.Fatalf("%s: recorded Go outcome changed", record.Label)
				}
				found = true
				break
			}
		}
		if !found {
			test.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	upstream, absolutePathError := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keys, statementCaseKeysError := StatementCaseKeys(items, root, upstream)
	if statementCaseKeysError != nil {
		test.Fatal(statementCaseKeysError)
	}
	partition, statementPartitionError := StatementPartition(keys)
	if statementPartitionError != nil {
		test.Fatal(statementPartitionError)
	}
	dir := test.TempDir()
	result := make([]statementShard, len(partition))
	projected := make([][]int, len(partition))
	for number, ids := range partition {
		shard := statementShard{
			Indices: ids,
			Text:    filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.txt", number)),
			Specs:   filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.json", number)),
		}
		var input, expected strings.Builder
		subset := make([]PrinterCase, 0, len(ids))
		for _, id := range ids {
			input.WriteString(rows[id] + "\n")
			expected.WriteString(answers[id] + "\n")
			subset = append(subset, items[id])
			shard.Labels = append(shard.Labels, fmt.Sprintf("case %d %s", id, items[id].Label))
			projected[number] = append(projected[number], id)
		}
		shard.Want = expected.String()
		encoded, encodeError := json.Marshal(subset)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(shard.Text, []byte(input.String()), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		if writeError := os.WriteFile(shard.Specs, encoded, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		result[number] = shard
	}
	if statementUnionError := StatementUnion(len(items), projected); statementUnionError != nil {
		test.Fatal(statementUnionError)
	}
	return result
}

// statementContext is a case command's context: the shard's deadline, or a context the test cancels at its end. Setup
// has no clock of its own; cancellation still kills the child's whole process group (childguard).
func statementContext(test *testing.T) context.Context {
	test.Helper()
	if value, ok := StatementCaseContexts.Load(test); ok {
		return value.(context.Context)
	}
	cancelContext, cancel := context.WithCancel(context.Background())
	test.Cleanup(cancel)
	return cancelContext
}

// StatementExecute runs name with this process's environment and environment added, in its own process group under
// the case's context, and returns what it printed and how it exited. A child that never started fails the test; one
// that ran and failed is a result. Output is read through childguard.Capture, to its end with no timer deciding when:
// exec's copying under a five-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF five
// seconds after the child exited, and that error came with a ProcessState, so it passed here as exit 0 with the output
// cut where the reads were closed (#n3gk22h's shape). Only an exit error is a result now, told by its ProcessState: an
// error beside a clean exit is never the child's (Capture's reads ended by the context, or the context's own), and an
// exit error is the one error that comes with a failed exit. Reading it there, not with errors.As, keeps os/exec out of
// this package, which the gate's compiler dependency census would read as a launcher. The run had no stall guard to
// lose (Unbounded), and the case's context and the unit's own ceiling bound it.
func StatementExecute(test *testing.T, environment []string, name string, arguments ...string) Run {
	test.Helper()
	ctx := statementContext(test)
	command := toolchain.CommandContextIn(ctx, "", name, arguments...)
	command.Env = append(os.Environ(), environment...)
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	if captureError != nil && (command.ProcessState == nil || command.ProcessState.Success()) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return Run{Stdout: stdout, Stderr: stderr, ExitCode: command.ProcessState.ExitCode()}
}

func statementCorpusBuild(test *testing.T) (string, string, string) {
	test.Helper()
	directory := test.TempDir()
	root, _ := filepath.Abs(Repository)
	files := printerCorpusFiles(test)
	// This count is fixed by the upstream commit, never by the live case total.
	upstream, absolutePathError := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	upstreamFiles, repositoryFiles := 0, 0
	for _, file := range files {
		path, absolutePathError := filepath.Abs(file)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		relative, relativePathError := filepath.Rel(upstream, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			upstreamFiles++
		} else {
			repositoryFiles++
		}
	}
	if upstreamFiles != 77 {
		test.Fatalf("TypeScript 050880ce src/compiler: %d .ts files, want pinned 77", upstreamFiles)
	}
	if repositoryFiles == 0 {
		test.Fatal("repository statement corpus is empty")
	}
	oracleDir := PrinterGateGoOracle(test)
	flags := []string{"Go oracle sha256=" + statementFileHash(test, oracleDir+"/oracle"), "width=80"}
	// An input under ADAMIC_TYPESCRIPT_SOURCE is named by the tree it is in, not where this machine keeps it
	// (#t37sw0f).
	for _, file := range files {
		path, absolutePathError := filepath.Abs(file)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		flags = append(
			flags,
			"input="+tsPrinterOutputLabel(path, root, upstream, true)+" sha256="+statementFileHash(test, file),
		)
	}
	product := statementProduct(test, buildcache.Inputs{
		Name:  "tsprinter-statements-oracle-output-v1",
		Files: []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
		Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}, func(dir string) error {
		// Requests are scratch inputs. Publish only the oracle's four output files.
		request, encodeError := json.Marshal(map[string]any{"Files": files, "Directory": dir})
		if encodeError != nil {
			return encodeError
		}
		if writeError := os.WriteFile(directory+"/request.json", request, 0644); writeError != nil {
			return writeError
		}
		ctx := statementContext(test)
		command := toolchain.CommandContextIn(
			ctx,
			"",
			oracleDir+"/oracle",
			"-test.v",
			"-test.count=1",
			"-test.run=^TestAdamicStatementCorpus$",
			"-test.timeout=0",
		)
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
		// Read through childguard.Capture for StatementExecute's reason, stdout then stderr: a five-second WaitDelay
		// could fail a corpus run that had finished. The stall guard it loses watched a run that prints its one RUN
		// line and then nothing until its summary, so it was never progress; the product unit's ceiling bounds it.
		stdout, stderr, captureError := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		if captureError != nil {
			return fmt.Errorf("Go statement corpus: %w\n%s", captureError, output)
		}
		test.Log(string(output))
		// The oracle labels each case by its file's path here; the product names it repo/ or typescript/ and the path
		// inside, as the expression outputs do, so a runner reading Workshop's product keys the same cases (#tqrqx60).
		return TSPrinterRewriteMetadata(dir, root, upstream, true)
	})
	directory = product
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_KEEP"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, name := range []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"} {
			data, readError := os.ReadFile(directory + "/" + name)
			if readError != nil {
				test.Fatal(readError)
			}
			if writeError := os.WriteFile(keep+"/"+name, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	answers, readError := os.ReadFile(directory + "/answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	PlanCorpus(test, directory+"/cases.txt")
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const StatementsAgainstGoAndPrettierShards = 16

func StatementLibraryOracle(
	test *testing.T,
	number, side int,
	shard statementShard,
	libraryHash, library string,
) string {
	test.Helper()
	return PrinterGateOnce(test, fmt.Sprintf("statement-library-%d-%d", number, side), func() string {
		return libraryOracle(test, "statements", number, side, shard.Specs, libraryHash, library, statementShards)
	})
}

// libraryOracle is family's case shard number's npm (side 0) or embedded (side 1) Prettier answers: node running the
// library's script over the shard's cases. The cases name their files by absolute path, hashed out of Key's reach, so
// the hash reads them as the tree spells them; the npm library is keyed by its bytes (npm bytes sha256), the embedded
// one by Files, neither by where it is, and NODE_PATH is a location (#t37sw0f). The test binary isn't an input: the
// oracle is node running the script over the cases, and hashing the binary keyed a product apart between Workshop's go
// test and a runner's prebuilt binary of the same package (landable-10, key check).
//
// It's a LeanProductOf build (#zb9268t, Node's and Tests' rulings), keyed by its own window, so what its test read
// before it never keys it: two cold builds of one commit keyed these shards apart through what their processes had read
// first. libraryOracleBuild is handed the script, the library and the shard's cases, which it reads from the shards
// product (shards, a product of its own), never from a file this test wrote, which the window could only key by where
// it sits. A process that may not build an oracle finds it (oracleFound) and never builds it.
func libraryOracle(
	test *testing.T,
	family string,
	number, side int,
	specs, libraryHash, library string,
	shards func(*testing.T) string,
) string {
	test.Helper()
	name, script, files, oracleFlags := "npm", "testdata/expressions.mjs", []string{
		"stage1/cohere/tsprinter/testdata/expressions.mjs",
		"stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json",
		"stage1/cohere/tsprinter/testdata/prettier-differences.json",
	}, []string{"npm bytes sha256=" + libraryHash}
	if side == 1 {
		name, script = "embedded", "testdata/embedded.mjs"
		library, _ = filepath.Abs(filepath.Join(Repository, "cohere/internal/format/prettier/bundles"))
		files = []string{
			"stage1/cohere/tsprinter/testdata/embedded.mjs",
			"stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json",
			"cohere/internal/format/prettier/bundles",
		}
		oracleFlags = nil
	}
	script, _ = filepath.Abs(script)
	root, _ := filepath.Abs(Repository)
	upstream, _ := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	data, readError := os.ReadFile(specs)
	if readError != nil {
		test.Fatal(readError)
	}
	cases := StatementBytesHash(
		[]byte(strings.ReplaceAll(strings.ReplaceAll(string(data), upstream+"/", "typescript/"), root+"/", "repo/")),
	)
	inputs := oracleInputs(buildcache.Inputs{
		Name:  fmt.Sprintf("tsprinter-%s-%s-oracle-shard-%03d-v1", family, name, number),
		Files: files,
		Flags: append(
			oracleFlags,
			"cases sha256="+cases,
			"library="+name,
			"script="+script,
			"width=80",
			"NODE_OPTIONS="+os.Getenv("NODE_OPTIONS"),
		),
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
	})
	if directory, found, oracleFoundError := oracleFound(
		inputs,
		oracleProducer,
		buildcache.Tracing(),
	); oracleFoundError != nil {
		test.Fatal(oracleFoundError)
	} else if found {
		return directory
	}
	// The shard's cases as the shards product holds them, the same bytes this process partitioned, or the key's cases
	// sha256 would name cases the build never read.
	held := filepath.Join(shards(test), fmt.Sprintf("shard-%03d.json", number))
	data, readError = os.ReadFile(held)
	if readError != nil {
		test.Fatal(readError)
	}
	if got := StatementBytesHash(data); got != cases {
		test.Fatalf(
			"%s shard %03d: the shards product holds cases sha256 %s, and this process partitioned %s",
			family,
			number,
			got,
			cases,
		)
	}
	start, cpu := time.Now(), StatementCPU()
	directory := buildcache.LeanProductOf(
		test,
		buildcache.Handed{Inputs: inputs, Paths: []string{script, library, held}},
		libraryOracleBuild,
	)
	test.Logf("%s: %.3fs, %.3f CPU s", inputs.Name, time.Since(start).Seconds(), StatementCPU()-cpu)
	return directory
}

// libraryOracleBuild runs node over what it's handed, the script, the library and the shard's cases, and keeps the
// answers it prints. It's a top-level function of what it's handed alone (buildcache.LeanProductOf): no test, no
// package state (TestALibraryOraclesBuildReadsNoStateButItsInitializers). node runs in its own process group, read
// through childguard.Capture for StatementExecute's reason: under a five-second WaitDelay an exec.ErrWaitDelay passed
// the exit check as code 0, and the product stored the answers cut where the reads were closed. Capture kills the group
// once node exits, so nothing node left behind can hold a read open; the product unit's ceiling bounds the run.
func libraryOracleBuild(handed buildcache.Handed, directory string) error {
	backgroundContext := context.Background()
	command := toolchain.CommandContextIn(backgroundContext, "", "node", handed.Paths...)
	command.Env = os.Environ()
	// As in StatementExecute, an error beside a clean exit is never node's answer.
	stdout, stderr, captureError := childguard.Capture(backgroundContext, command)
	if captureError != nil && (command.ProcessState == nil || command.ProcessState.Success()) {
		return fmt.Errorf("running node: %w", captureError)
	}
	if code := command.ProcessState.ExitCode(); code != 0 || len(stderr) != 0 {
		return fmt.Errorf("%s exit %d: %s", handed.Inputs.Name, code, stderr)
	}
	return os.WriteFile(filepath.Join(directory, "answers.txt"), stdout, 0o644)
}

// canonicalRoot and canonicalSource stand for the checkout and the TypeScript source in a shards build: a corpus
// product's labels are already repo/ and typescript/ (TSPrinterRewriteMetadata), and a case's stable key is its path
// inside either, the same whatever they're spelled as.
const canonicalRoot, canonicalSource = "/repo", "/typescript"

// statementShards is the statement shards product, once per process: shard-NNN.json for each of the
// StatementsAgainstGoAndPrettierShards shards, the cases StatementShards gives it, built from the statement corpus
// product.
func statementShards(test *testing.T) string {
	test.Helper()
	return PrinterGateOnce(test, "statement-shards", func() string {
		corpus := filepath.Dir(StatementGateShardsOnce(test).Cases)
		return buildcache.LeanProductOf(test, buildcache.Handed{
			Inputs: shardsInputs("statements", StatementsAgainstGoAndPrettierShards),
			Paths:  []string{corpus},
		}, statementShardsBuild)
	})
}

// expressionShards is the expression shards product, once per process: shard-NNN.json for each case unit, the cases
// ExpressionUnits gives it, built from the expression corpus product.
func expressionShards(test *testing.T) string {
	test.Helper()
	return PrinterGateOnce(test, "expression-shards", func() string {
		corpus := ExpressionGateUnitsOnce(test).corpus
		return buildcache.LeanProductOf(test, buildcache.Handed{
			Inputs: shardsInputs("expressions", ExpressionsAgainstGoAndPrettierShards-1),
			Paths:  []string{corpus},
		}, expressionShardsBuild)
	})
}

// shardsInputs key a family's shards product: the corpus product it's handed keys it through its window, and the code
// that partitions it through the recipe.
func shardsInputs(family string, count int) buildcache.Inputs {
	return oracleInputs(buildcache.Inputs{
		Name:      "tsprinter-" + family + "-shards-v1",
		Files:     []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
		Flags:     []string{fmt.Sprintf("shards=%d", count)},
		Toolchain: []string{runtime.GOOS, runtime.GOARCH},
	})
}

// statementShardsBuild partitions the statement corpus product it's handed as StatementShards does, each shard's cases
// as StatementShards writes them.
func statementShardsBuild(handed buildcache.Handed, directory string) error {
	cases, corpusCasesError := corpusCases(handed.Paths[0])
	if corpusCasesError != nil {
		return corpusCasesError
	}
	keys, statementCaseKeysError := StatementCaseKeys(cases, canonicalRoot, canonicalSource)
	if statementCaseKeysError != nil {
		return statementCaseKeysError
	}
	partition, statementPartitionError := StatementPartition(keys)
	if statementPartitionError != nil {
		return statementPartitionError
	}
	return writeShards(directory, cases, partition)
}

// expressionShardsBuild partitions the expression corpus product it's handed as ExpressionUnits does.
func expressionShardsBuild(handed buildcache.Handed, directory string) error {
	cases, corpusCasesError := corpusCases(handed.Paths[0])
	if corpusCasesError != nil {
		return corpusCasesError
	}
	keys, statementCaseKeysError := StatementCaseKeys(cases, canonicalRoot, canonicalSource)
	if statementCaseKeysError != nil {
		return statementCaseKeysError
	}
	partition, expressionPartitionError := expressionPartition(
		expressionCaseKeys(keys),
		ExpressionsAgainstGoAndPrettierShards-1,
	)
	if expressionPartitionError != nil {
		return expressionPartitionError
	}
	return writeShards(directory, cases, partition)
}

// corpusCases are a corpus product's cases.
func corpusCases(corpus string) ([]PrinterCase, error) {
	data, readError := os.ReadFile(filepath.Join(corpus, "cases.json"))
	if readError != nil {
		return nil, readError
	}
	var cases []PrinterCase
	return cases, json.Unmarshal(data, &cases)
}

// writeShards writes each shard of partition's cases to shard-NNN.json, encoded as the partitioning tests encode them.
func writeShards(directory string, cases []PrinterCase, partition [][]int) error {
	for number, indices := range partition {
		subset := make([]PrinterCase, 0, len(indices))
		for _, index := range indices {
			subset = append(subset, cases[index])
		}
		encoded, encodeError := json.Marshal(subset)
		if encodeError != nil {
			return encodeError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number)),
			encoded,
			0o644,
		); writeError != nil {
			return writeError
		}
	}
	return nil
}

func StatementCorpus(test *testing.T) (string, string, string) {
	test.Helper()
	directory := PrinterGateOnce(
		test,
		"statement-corpus",
		func() string { cases, _, _ := statementCorpusBuild(test); return filepath.Dir(cases) },
	)
	answers, readError := os.ReadFile(directory + "/answers.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const TSCCorpusAgreementShards = 8

type tscAgreementFamily struct {
	Name, Directory                     string
	Cases                               []PrinterCase
	Source, Backend, Sanitized, Release string
}

type TSCAgreementData struct{ Families []tscAgreementFamily }

// tscAgreementRecipe are the files whose code makes the corpus product's bytes: its build, the TSC filter, the oracle
// outputs it asks for and relabels, the oracle it names, the tracked listing of the corpus, and how the oracle runs and
// its output is captured, all in this package since #2ba0c20. printerGateOnce is only a per-process memo, and the
// oracle it builds is keyed by its bytes.
var TSCAgreementRecipe = []string{"stage1/cohere/tsprinter/oracles/oracles.go"}

// tscAgreementInputs are what the corpus product's build reads: the TSC corpus, the gaps file, the Go oracle that
// answers every case, by its bytes as expressionBuild keys a generated input, and the recipe. Keying every tracked file
// of the repository cost the first shard of every process 19 s of walking (#4ycgm2n).
func tscAgreementInputs(test *testing.T) buildcache.Inputs {
	test.Helper()
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	oracle, readError := os.ReadFile(tscOracle(test, root))
	if readError != nil {
		test.Fatal(readError)
	}
	return TSCAgreementKeyInputs(fmt.Sprintf("%x", sha256.Sum256(oracle)))
}

// tscAgreementKeyInputs is tscAgreementInputs for the oracle whose bytes hash to oracle.
func TSCAgreementKeyInputs(oracle string) buildcache.Inputs {
	return buildcache.Inputs{
		Name: "TSC agreement corpus v3",
		Files: append(
			[]string{"stage3/drivers/tsc/corpus", "stage1/cohere/tsprinter/testdata/notyet.json"},
			TSCAgreementRecipe...,
		),
		Flags:     []string{fmt.Sprint(TSCCorpusAgreementShards), "oracle " + oracle},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

// Every independently selected leaf can prepare the corpus; no test ordering is required.
func TSCAgreementPrepare(test *testing.T) string {
	test.Helper()
	return oracleProduct(test, tscAgreementInputs(test), func(directory string) error {
		root, absolutePathError := filepath.Abs(Repository)
		if absolutePathError != nil {
			return absolutePathError
		}
		files, trackedRootFilesError := TrackedRootFiles(test, root, "stage3/drivers/tsc/corpus")
		if trackedRootFilesError != nil {
			return trackedRootFilesError
		}
		oracle := tscOracle(test, root)
		data := TSCAgreementData{}
		for _, statements := range []bool{false, true} {
			family, entry := "expressions", "main.ts"
			if statements {
				family, entry = "statements", "statementsMain.ts"
			}
			cases, original := tscCorpus(test, root, files, family, statements, oracle)
			output := filepath.Join(directory, family)
			if createDirectoryError := os.MkdirAll(output, 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			for _, name := range []string{"coverage.json"} {
				content, readError := os.ReadFile(filepath.Join(original, name))
				if readError != nil {
					return readError
				}
				// A refused file is named by the root the reader supplies (#tqrqx60).
				if writeError := os.WriteFile(
					filepath.Join(output, name),
					[]byte(buildcache.Relative(string(content))),
					0644,
				); writeError != nil {
					return writeError
				}
			}
			// Each case's label names its corpus file by the root the reader supplies, never this checkout (#tqrqx60).
			for i := range cases {
				cases[i].Label = buildcache.Relative(cases[i].Label)
			}
			data.Families = append(
				data.Families,
				tscAgreementFamily{Name: family, Directory: family, Cases: cases, Source: entry},
			)
		}
		encoded, encodeError := json.Marshal(data)
		if encodeError != nil {
			return encodeError
		}
		return os.WriteFile(filepath.Join(directory, "manifest.json"), encoded, 0644)
	})
}

type PrinterCase struct{ Label, Source, Want string }

func tscCorpus(
	test *testing.T,
	root string,
	files []string,
	family string,
	statements bool,
	oracle string,
) ([]PrinterCase, string) {
	test.Helper()
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(test, root, files, gaps, family, oracle)
	data, readError := os.ReadFile(filepath.Join(directory, "cases.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var selected, cases []PrinterCase
	if decodeError := json.Unmarshal(data, &selected); decodeError != nil {
		test.Fatal(decodeError)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var batch strings.Builder
	for _, item := range selected {
		// This focused gate holds repository fragments, not generated cases.
		if strings.HasPrefix(item.Label, root+"/stage3/drivers/tsc/corpus/") {
			cases = append(cases, item)
			batch.WriteString(">" + escape.Replace(item.Source) + "\n")
		}
	}
	path := filepath.Join(directory, "tsc-cases.txt")
	if writeError := os.WriteFile(path, []byte(batch.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	encoded, encodeError := json.Marshal(cases)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(filepath.Join(directory, "tsc-cases.json"), encoded, 0644); writeError != nil {
		test.Fatal(writeError)
	}

	return cases, directory
}

func tscOracle(test *testing.T, root string) string {
	test.Helper()
	return PrinterGateGoOracle(test) + "/oracle"
}

type upstreamOutcome struct {
	Family, Label, Source, Go, Prettier, Embedded, PrettierError, EmbeddedError string
}

func UpstreamOutcomes(test *testing.T) []upstreamOutcome {
	test.Helper()
	data, readError := os.ReadFile("testdata/tsc-upstream-differences.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var pinned struct {
		Version string
		Records []upstreamOutcome
	}
	if decodeError := json.Unmarshal(data, &pinned); decodeError != nil {
		test.Fatal(decodeError)
	}
	if pinned.Version != "3.9.6" {
		test.Fatalf("upstream outcome version: %s", pinned.Version)
	}
	return pinned.Records
}

// oracleProducer is whether this process may settle the oracle products: only this package's test binary, which links
// no compiler, calls EnableProductProduction. An untraced run (a developer's) still builds them where it is; a traced
// tsprinter process finds the oracles binary's settlement instead of settling one under its own compiler-linked recipe.
var oracleProducer bool

// EnableProductProduction grants this process ownership of the oracle products. Only the oracles test binary calls
// it, from TestMain before its tests run.
func EnableProductProduction() {
	oracleProducer = true
}

// oracleInputs is an oracle product's key: its name with ' oracle-owner-v1', as css's are (#dkzphq7), so recordUnion
// never keeps a read set a compiler-linked tsprinter recipe settled under the old name.
func oracleInputs(inputs buildcache.Inputs) buildcache.Inputs {
	inputs.Name += " oracle-owner-v1"
	return inputs
}

// oracleBuild is build, refused under a trace in any process but the oracles binary.
func oracleBuild(name string, build func(directory string) error) func(directory string) error {
	return func(directory string) error {
		if buildcache.Tracing() && !oracleProducer {
			return fmt.Errorf(
				"oracle %s is missing: stage1/cohere/tsprinter/oracles' TestProduct_ units build it first",
				name,
			)
		}
		return build(directory)
	}
}

// oracleFound is the oracle product inputs keys, found, when this process may not build it: under a trace, in a process
// that isn't the oracles binary (producer), it's looked for and never built, so a missing one fails at once, naming its
// owner, as buildcache.BuiltElsewhere says why. Otherwise it isn't found here (false), and the caller builds it. The
// guard is read here, outside the build, so the build reads no package state (#zb9268t, Tests' ruling).
func oracleFound(inputs buildcache.Inputs, producer, tracing bool) (string, bool, error) {
	if producer || !tracing {
		return "", false, nil
	}
	directory, getError := buildcache.Get(inputs, func(string) error { return errorOracleElsewhere })
	if getError != nil {
		return "", false, fmt.Errorf(
			"oracle %s is stage1/cohere/tsprinter/oracles' to build, and it isn't here: its TestProduct_ units build it first. %w",
			inputs.Name,
			getError,
		)
	}
	return directory, true, nil
}

// errorOracleElsewhere is the build a process that may not build an oracle product hands buildcache.Get.
var errorOracleElsewhere = errors.New("built by stage1/cohere/tsprinter/oracles")

// oracleProduct is buildcache.Product for an oracle product, keyed and gated as oracleInputs and oracleBuild say.
func oracleProduct(test *testing.T, inputs buildcache.Inputs, build func(directory string) error) string {
	test.Helper()
	inputs = oracleInputs(inputs)
	return buildcache.Product(test, inputs, oracleBuild(inputs.Name, build))
}

// goTestOracle is cohere's go test -c of pkg, mutated when mutations are given: built by owner, an oracles
// TestProduct_, and found by its key in a traced tsprinter process through buildcache.BuiltElsewhere, which never
// builds. Its name is buildcache's own, so it carries no oracle-owner-v1.
func goTestOracle(
	test *testing.T,
	owner, output, pkg string,
	arguments []string,
	mutations []buildcache.Mutation,
) string {
	test.Helper()
	if buildcache.Tracing() && !oracleProducer {
		inputs, goTestInputsError := buildcache.GoTestInputs("cohere", output, pkg, arguments, mutations)
		if goTestInputsError != nil {
			test.Fatal(goTestInputsError)
		}
		return filepath.Join(buildcache.BuiltElsewhere(test, "the Go oracle "+
			pkg, inputs, "stage1/cohere/tsprinter/oracles "+owner, func(string) error {
			return errors.New("built by stage1/cohere/tsprinter/oracles")
		}), output)
	}
	if len(mutations) > 0 {
		return buildcache.GoTestMutated(test, "cohere", output, pkg, arguments, mutations)
	}
	return buildcache.GoTest(test, "cohere", output, pkg, arguments)
}
