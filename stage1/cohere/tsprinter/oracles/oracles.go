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
	"os/exec"
	"path/filepath"
	"runtime"
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
	"github.com/system-inc/adamic/internal/tracked"
)

var repositoryCorpusRoots = []string{
	"bench", "bridge", "cmd", "internal", "stage1", "stage3/drivers/tsc/corpus",
}

// repositoryCorpusCommit pins the repository's half of the printer corpus: every .ts file under repositoryCorpusRoots as
// it was at this commit on main, read from the checkout's objects, never from the working tree. The printers' agreement
// needs a large TypeScript corpus, not a current one, so a .ts edit anywhere in the repository moves no tsprinter
// product; walked live, one edit to lint's main.ts moved all 162 (#wy8gcke, Planner on #rybwgtz). The corpus moves only
// when this pin does. The checkout that builds the corpus product must hold the commit, as json's does.
const RepositoryCorpusCommit = "6f7d5d0ba4a7ccafb6ed69fd531d169f9effe443"

// tsPrinterPinnedCorpusInputs keys the pinned corpus by its commit and roots: a different pin, or another root, is
// another product.
func TSPrinterPinnedCorpusInputs(pin string) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "tsprinter pinned corpus",
		Files: []string{"stage1/cohere/tsprinter/corpus_test.go"},
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
		root, err := filepath.Abs(Repository)
		if err != nil {
			return err
		}
		files, err := buildcache.PinnedFiles(root, inputs, pin, "TypeScript", inRepositoryCorpus)
		if err != nil {
			return err
		}
		var list strings.Builder
		for _, file := range files {
			path := filepath.Join(directory, filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, file.Data, 0o644); err != nil {
				return err
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

func pinnedCorpusFiles(t *testing.T) []string {
	t.Helper()
	pinnedCorpus.once.Do(func() {
		inputs := oracleInputs(TSPrinterPinnedCorpusInputs(RepositoryCorpusCommit))
		directory, err := buildcache.Get(inputs, oracleBuild(inputs.Name, buildTSPrinterPinnedCorpus(inputs, RepositoryCorpusCommit)))
		if err != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", err)
			return
		}
		list, err := os.ReadFile(filepath.Join(directory, corpusListName))
		if err != nil {
			pinnedCorpus.err = fmt.Errorf("tsprinter pinned corpus: %w", err)
			return
		}
		for _, name := range strings.Split(strings.TrimSuffix(string(list), "\n"), "\n") {
			pinnedCorpus.files = append(pinnedCorpus.files, filepath.Join(directory, filepath.FromSlash(name)))
		}
		pinnedCorpus.directory.Store(&directory)
	})
	if pinnedCorpus.err != nil {
		t.Fatal(pinnedCorpus.err)
	}
	return append([]string(nil), pinnedCorpus.files...)
}

// tsPrinterLabelRoots are the directories a case label may name a file under, each with the name a stored product
// spells it by: the checkout and the pinned corpus (the repository at its pin) as repo/, the TypeScript source as
// typescript/. Read back (canonical false), repo/ is the checkout, which comes first.
func tsPrinterLabelRoots(root, external string, canonical bool) [][2]string {
	roots := [][2]string{{root, "repo"}}
	if directory := pinnedCorpus.directory.Load(); canonical && directory != nil {
		roots = append(roots, [2]string{*directory, "repo"})
	}
	return append(roots, [2]string{external, "typescript"})
}

func TrackedRootFiles(t *testing.T, checkout, root string) ([]string, error) {
	t.Helper()
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
	t.Logf("corpus root %s: %d git-tracked .ts files", root, len(files))
	return files, nil
}

func printerCorpusFiles(t *testing.T) []string {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to a TypeScript 6.0.3 source checkout at 050880ce; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	// The pin is a check, never an input: tracked.Head asks git as key-time work, with the home's configuration out of
	// reach (home-6's dry run, Oct 11: this git read ~/.gitconfig inside the build, refusing it and the Go oracle beside it).
	if head, err := tracked.Head(source); err != nil || head != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("TypeScript pin: %q %v", head, err)
	}
	// The repository's half is the pinned corpus product's files, never the working tree's (repositoryCorpusCommit).
	files := pinnedCorpusFiles(t)
	listed, err := TrackedRootFiles(t, source, "src/compiler")
	if err != nil {
		t.Fatal(err)
	}
	return append(files, listed...)
}

func DocumentCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	cohere, _ := filepath.Abs(Repository + "/cohere")
	// The unit only runs the oracle, in the package's directory.
	command := Bounded(t, documentOracle(t), "-test.run=^TestAdamicDocuments$")
	command.Dir = cohere + "/internal/format/doc"
	command.Env = append(os.Environ(), "ADAMIC_TS_DOC_OUTPUT="+directory)
	if output, err := CombinedOutput(command); err != nil {
		t.Fatalf("Go doc corpus %v\n%s", err, output)
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	PlanCorpus(t, directory+"/docs.txt")
	return directory + "/docs.txt", string(answers), directory + "/docs.json"
}

// documentOracle is cohere's doc package with the harness laid over it, a product built ahead (#5qykzj5).
func documentOracle(t *testing.T) string {
	t.Helper()
	return goTestOracle(t, "TestProduct_TSPrinterDocumentOracle", "doc.test", "./internal/format/doc", documentOracleArguments(t), nil)
}

// documentOracleInputs is documentOracle's key, computed without building it.
func documentOracleInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	inputs, err := buildcache.GoTestInputs("cohere", "doc.test", "./internal/format/doc", documentOracleArguments(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

// documentOracleArguments lay the harness over cohere's doc package.
func documentOracleArguments(t *testing.T) []string {
	t.Helper()
	cohere, _ := filepath.Abs(Repository + "/cohere")
	side, _ := filepath.Abs("testdata/doc_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{cohere + "/internal/format/doc/adamic_docs_test.go": side}})
	path := t.TempDir() + "/overlay.json"
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	return []string{"-overlay=" + path}
}

func ExpressionGateKeys(t *testing.T, cases []PrinterCase) []string {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := StatementCaseKeys(cases, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	// Generated expression cases are owned by their actual generator file.
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

func expressionGatePartition(t *testing.T, cases []PrinterCase, count int) [][]int {
	t.Helper()
	keys := ExpressionGateKeys(t, cases)
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			t.Fatalf("repeated stable expression key %q", key)
		}
		seen[key] = true
	}
	return ExpressionGateOwners(keys, count)
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
		var err error
		external, err = filepath.Abs(external)
		if err != nil {
			t.Fatal(err)
		}
	}
	flags := []string{"test-run=" + test, "metadata=canonical-v1"}
	for _, file := range files {
		flags = append(flags, "input-id="+tsPrinterOutputLabel(file, root, external, true))
	}
	inputs := append([]string{oracle, gaps, root + "/stage1/cohere/tsprinter/expression_products_test.go", root + "/stage1/cohere/tsprinter/expressions_test.go"}, files...)
	product := oracleExpressionBuild(t, PrinterBuildInputs{Name: "TS " + family + " oracle outputs", Files: inputs, Flags: flags, Toolchain: runtime.Version()}, func(directory string) error {
		request, err := json.Marshal(map[string]any{"Files": files, "Directory": directory, "Gaps": gaps})
		if err != nil {
			return err
		}
		// Requests name this builder's paths and must never become shared outputs.
		requestPath := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(requestPath, request, 0644); err != nil {
			return err
		}
		command := Bounded(t, oracle, "-test.v", "-test.count=1", "-test.run=^"+test+"$", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), variable+"="+requestPath)
		output, err := CombinedOutput(command)
		if err != nil {
			return fmt.Errorf("Go %s corpus: %w\n%s", family, err, output)
		}
		t.Log(string(output))
		return TSPrinterRewriteMetadata(directory, root, external, true)
	})
	return product
}

func tsPrinterOracleOutputs(t *testing.T, root string, files []string, gaps, family, oracle string) string {
	t.Helper()
	external := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if external != "" {
		var err error
		external, err = filepath.Abs(external)
		if err != nil {
			t.Fatal(err)
		}
	}
	product := PrinterGateOnce(t, "oracle-output-"+family+"-"+StatementBytesHash([]byte(strings.Join(files, "\n"))), func() string { return tsPrinterOracleProduct(t, root, files, gaps, family, oracle) })
	// Shard preparation and caller-requested audits use a private writable copy.
	directory := t.TempDir()
	names := []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"}
	if family == "expressions" {
		names = append(names, "gaps.txt", "gap-answers.txt", "gaps.json")
	}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(product, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := TSPrinterRewriteMetadata(directory, root, external, false); err != nil {
		t.Fatal(err)
	}
	return directory
}

func tsPrinterOutputLabel(label, root, external string, canonical bool) string {
	for _, pair := range tsPrinterLabelRoots(root, external, canonical) {
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
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cases []PrinterCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return err
	}
	for i := range cases {
		cases[i].Label = tsPrinterOutputLabel(cases[i].Label, root, external, canonical)
	}
	data, err = json.Marshal(cases)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	path = filepath.Join(directory, "coverage.json")
	data, err = os.ReadFile(path)
	if err != nil {
		return err
	}
	var report map[string]json.RawMessage
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	if raw, exists := report["file_refusals"]; exists {
		var refusals map[string]string
		if err := json.Unmarshal(raw, &refusals); err != nil {
			return err
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
		report["file_refusals"], err = json.Marshal(rewritten)
		if err != nil {
			return err
		}
	}
	data, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ExpressionUnits(t *testing.T, path, want string, count int) (*CorpusPlan, [][]byte) {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(strings.TrimSuffix(path, ".txt") + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []PrinterCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	rows := bytes.Split(bytes.TrimSuffix(text, []byte("\n")), []byte("\n"))
	answers := bytes.Split(bytes.TrimSuffix([]byte(want), []byte("\n")), []byte("\n"))
	if len(rows) != len(cases) || len(answers) != len(cases) || len(cases) == 0 {
		t.Fatalf("expression enumeration: %d cases, %d rows, %d answers", len(cases), len(rows), len(answers))
	}
	// Every pinned upstream record must occur in the unsplit enumeration. Each
	// unit then checks the exact independently pinned outcome for its own cases.
	for _, record := range UpstreamOutcomes(t) {
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
			t.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	plan := &CorpusPlan{Labels: make([]string, len(cases))}
	for i, item := range cases {
		plan.Labels[i] = fmt.Sprintf("case-%06d %s", i, item.Label)
	}
	directory := t.TempDir()
	var outputs [][]byte
	partition := expressionGatePartition(t, cases, count)
	for number, indices := range partition {
		shard := CorpusShard{Indices: indices, Text: filepath.Join(directory, fmt.Sprintf("shard-%03d.txt", number)), Specs: filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number))}
		var protocol, expected bytes.Buffer
		subset := make([]PrinterCase, 0, len(indices))
		for _, i := range indices {
			protocol.Write(rows[i])
			protocol.WriteByte('\n')
			expected.Write(answers[i])
			expected.WriteByte('\n')
			subset = append(subset, cases[i])
		}
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.Text, protocol.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.Specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		plan.Shards = append(plan.Shards, shard)
		outputs = append(outputs, expected.Bytes())
	}
	if err := ExpressionUnion(plan, outputs, []byte(want)); err != nil {
		t.Fatal(err)
	}
	gapData, err := os.ReadFile(filepath.Dir(path) + "/gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var gaps []PrinterCase
	if err := json.Unmarshal(gapData, &gaps); err != nil {
		t.Fatal(err)
	}
	gapAnswers, err := os.ReadFile(filepath.Dir(path) + "/gap-answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) != bytes.Count(gapAnswers, []byte("\n")) {
		t.Fatal("gap enumeration/answer count differs")
	}
	whole := &CorpusPlan{Labels: append([]string(nil), plan.Labels...), Shards: append([]CorpusShard(nil), plan.Shards...)}
	gapIDs := make([]int, len(gaps))
	for i, item := range gaps {
		gapIDs[i] = len(cases) + i
		whole.Labels = append(whole.Labels, fmt.Sprintf("case-%06d gap %s", gapIDs[i], item.Label))
	}
	whole.Shards = append(whole.Shards, CorpusShard{Indices: gapIDs})
	wholeOutputs := append(append([][]byte(nil), outputs...), gapAnswers)
	if err := ExpressionUnion(whole, wholeOutputs, append([]byte(want), gapAnswers...)); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique expression case ids in %d units; %d gap ids in one unit; total %d", len(cases), len(plan.Shards), len(gaps), len(cases)+len(gaps))
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
func ExpressionGateUnitsOnce(t *testing.T) *expressionGateUnits {
	t.Helper()
	// printerCorpusFiles skips without a TypeScript source. Skip here, before the
	// once, so every unit skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" {
		printerCorpusFiles(t)
	}
	shared := &expressionGateUnitsShared
	shared.once.Do(func() {
		cases, want, _ := ExpressionCorpus(t, false)
		plan, answers := ExpressionUnits(t, cases, want, ExpressionsAgainstGoAndPrettierShards-1)
		units := &expressionGateUnits{Plan: &CorpusPlan{Labels: plan.Labels}, Answers: answers, builder: t.Name()}
		for _, shard := range plan.Shards {
			text, err := os.ReadFile(shard.Text)
			if err != nil {
				t.Fatal(err)
			}
			specs, err := os.ReadFile(shard.Specs)
			if err != nil {
				t.Fatal(err)
			}
			units.Plan.Shards = append(units.Plan.Shards, CorpusShard{Indices: shard.Indices})
			units.texts, units.specs = append(units.texts, text), append(units.specs, specs)
		}
		directory := filepath.Dir(cases)
		for _, file := range []struct {
			name string
			data *[]byte
		}{{"gaps.txt", &units.gapText}, {"gaps.json", &units.gapSpecs}, {"gap-answers.txt", &units.GapWant}} {
			data, err := os.ReadFile(filepath.Join(directory, file.name))
			if err != nil {
				t.Fatal(err)
			}
			*file.data = data
		}
		shared.units = units
	})
	if shared.units == nil {
		t.Fatal("expression unit preparation failed")
	}
	if shared.units.builder != t.Name() {
		t.Logf("union: checked once in this process by %s", shared.units.builder)
	}
	return shared.units
}

// A unit's own protocol and specifications, written where its children read them.
func (units *expressionGateUnits) Unit(t *testing.T, number int) CorpusShard {
	t.Helper()
	directory := t.TempDir()
	shard := CorpusShard{Indices: units.Plan.Shards[number].Indices, Text: filepath.Join(directory, fmt.Sprintf("shard-%03d.txt", number)), Specs: filepath.Join(directory, fmt.Sprintf("shard-%03d.json", number))}
	if err := os.WriteFile(shard.Text, units.texts[number], 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shard.Specs, units.specs[number], 0644); err != nil {
		t.Fatal(err)
	}
	return shard
}

// The loud gap unit's inputs. The Prettier script knows its gap proof by the name gaps.json.
func (units *expressionGateUnits) Gaps(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	for name, data := range map[string][]byte{"gaps.txt": units.gapText, "gaps.json": units.gapSpecs} {
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
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
	merged, err := MergeCorpus(plan, outputs)
	if err != nil {
		return err
	}
	if !bytes.Equal(merged, want) {
		return fmt.Errorf("union differs from unsplit answers")
	}
	return nil
}

func ExpressionCorpus(t *testing.T, processPlan ...bool) (string, string, string) {
	t.Helper()
	root, _ := filepath.Abs(Repository)
	files := printerCorpusFiles(t)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(t, root, files, gaps, "expressions", PrinterGateGoOracle(t)+"/oracle")
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_TS_PRINTER_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"coverage.json", "cases.json", "cases.txt", "answers.txt", "gaps.txt", "gap-answers.txt", "gaps.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(processPlan) == 0 || processPlan[0] {
		PlanCorpus(t, directory+"/cases.txt")
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const ExpressionsAgainstGoAndPrettierShards = 65

// expressionLibraryOracle is a case shard's npm (side 0) or embedded (side 1) Prettier answers, the product
// statementLibraryOracle makes for a statement shard and keyed the same way. The gap unit's proof still runs live.
func ExpressionLibraryOracle(t *testing.T, number, side int, shard CorpusShard, libraryHash, library string) string {
	t.Helper()
	name, script, files, oracleFlags := "npm", "testdata/expressions.mjs", []string{"stage1/cohere/tsprinter/testdata/expressions.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "stage1/cohere/tsprinter/testdata/prettier-differences.json"}, []string{"npm bytes sha256=" + libraryHash}
	if side == 1 {
		name, script = "embedded", "testdata/embedded.mjs"
		library, _ = filepath.Abs(filepath.Join(Repository, "cohere/internal/format/prettier/bundles"))
		files = []string{"stage1/cohere/tsprinter/testdata/embedded.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "cohere/internal/format/prettier/bundles"}
		oracleFlags = nil
	}
	script, _ = filepath.Abs(script)
	return PrinterGateOnce(t, fmt.Sprintf("expression-library-%d-%d", number, side), func() string {
		// The cases name their files by absolute path, so the hash reads them as the tree spells them; the npm library is
		// keyed by its bytes, the embedded one by Files, neither by where it is, and the test binary isn't an input, all as
		// statementLibraryOracle says why (#t37sw0f, landable-10).
		root, _ := filepath.Abs(Repository)
		upstream, _ := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
		specs, err := os.ReadFile(shard.Specs)
		if err != nil {
			t.Fatal(err)
		}
		cases := strings.ReplaceAll(strings.ReplaceAll(string(specs), upstream+"/", "typescript/"), root+"/", "repo/")
		flags := append(oracleFlags, "cases sha256="+StatementBytesHash([]byte(cases)), "library="+name, "script="+script, "width=80", "NODE_OPTIONS="+os.Getenv("NODE_OPTIONS"))
		return statementProduct(t, buildcache.Inputs{
			Name:  fmt.Sprintf("tsprinter-expressions-%s-oracle-shard-%03d-v1", name, number),
			Files: files, Flags: flags,
			Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
		}, func(dir string) error {
			result := StatementExecute(t, nil, "node", script, library, shard.Specs)
			if result.ExitCode != 0 || len(result.Stderr) != 0 {
				return fmt.Errorf("%s oracle exit %d: %s", name, result.ExitCode, result.Stderr)
			}
			return os.WriteFile(filepath.Join(dir, "answers.txt"), result.Stdout, 0644)
		})
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
func ExpressionBuild(t *testing.T, inputs PrinterBuildInputs, build func(dir string) error) string {
	t.Helper()
	return expressionBuildKeyed(t, expressionKey(t, inputs), build)
}

// oracleExpressionBuild is ExpressionBuild for an oracle product: its name says oracle-owner-v1, and only this
// package's binary builds it under a trace (oracleProduct).
func oracleExpressionBuild(t *testing.T, inputs PrinterBuildInputs, build func(dir string) error) string {
	t.Helper()
	key := oracleInputs(expressionKey(t, inputs))
	return expressionBuildKeyed(t, key, oracleBuild(key.Name, build))
}

func expressionBuildKeyed(t *testing.T, key buildcache.Inputs, build func(dir string) error) string {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := buildcache.Key(root, key)
	if err != nil {
		t.Fatal(err)
	}
	return PrinterGateOnce(t, "expression-product-"+digest, func() string {
		return buildcache.Product(t, key, build)
	})
}

// expressionKey is the key expressionBuild builds inputs under: each repository file by its path, and a generated input
// outside the repository by its bytes.
func expressionKey(t *testing.T, inputs PrinterBuildInputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.Inputs{Name: inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	for _, file := range inputs.Files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
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

// bounded prepares a child; execute and combinedOutput run it with progress-based guards.
func Bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	// The child guard owns hang detection, including for Go test children.
	if name == "go" && len(arguments) > 0 && arguments[0] == "test" {
		arguments = append([]string{"test", "-timeout=0"}, arguments[1:]...)
	}
	return exec.Command(name, arguments...)
}

func CombinedOutput(command *exec.Cmd) ([]byte, error) {
	return childguard.CombinedOutput(command, childguard.Options{})
}

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
func mutantOracle(t *testing.T, change Mutation, binary string) (string, string) {
	t.Helper()
	directory := t.TempDir()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	gaps, err := filepath.Abs("testdata/notyet.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(map[string]any{"Files": []string{}, "Directory": directory, "Gaps": gaps})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
		t.Fatal(err)
	}
	kind := "Expression"
	if change.Entry == "statementsMain.ts" {
		kind = "Statement"
	}
	// The test binary runs in its package's directory, as go test runs it.
	command := Bounded(t, binary, "-test.count=1", "-test.timeout=0", "-test.run=^TestAdamic"+kind+"Corpus$")
	command.Dir = root + "/cohere/internal/format/javascript"
	command.Env = append(os.Environ(), "ADAMIC_TS_EXPRESSION_REQUEST="+directory+"/request.json", "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
	if output, err := CombinedOutput(command); err != nil {
		t.Fatalf("Go mutation oracle: %v\n%s", err, output)
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
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
func mutantOracleBinary(t *testing.T, change Mutation) string {
	t.Helper()
	arguments, filters := mutantOracleRequest(t, change)
	return goTestOracle(t, "TestProduct_TSPrinterMutantOracle_*", "oracle", "./internal/format/javascript", arguments, filters)
}

// mutantOracleInputs is mutantOracleBinary's key, computed without building it.
func mutantOracleInputs(t *testing.T, change Mutation) buildcache.Inputs {
	t.Helper()
	arguments, filters := mutantOracleRequest(t, change)
	inputs, err := buildcache.GoTestInputs("cohere", "oracle", "./internal/format/javascript", arguments, filters)
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

// mutantOracleRequest is how mutantOracleBinary builds change's oracle: the overlay argument and each planted filter.
func mutantOracleRequest(t *testing.T, change Mutation) ([]string, []buildcache.Mutation) {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	families, ok := MutantFamilies[change.Name]
	if !ok && change.Name != "hashbang loses its refusal" {
		t.Fatalf("missing witness families for %q", change.Name)
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
		filters = append(filters, buildcache.Mutation{File: side, Before: mutantOracleAnchor, After: "allowed := " + allowed + "\nadd := func(label, source string) { if allowed[label] { cases = append(cases, portExpressionCase{Label: label, Source: source}) } }"})
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	// The overlay names this checkout's paths: GoTestMutated keys its map and the side files, never its own path.
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0o644); err != nil {
		t.Fatal(err)
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

func MutantOracleProduct(t *testing.T, index int) (string, string) {
	t.Helper()
	state := &mutantOracleProducts[index]
	state.Once.Do(func() {
		root, err := filepath.Abs(Repository)
		if err != nil {
			t.Fatal(err)
		}
		// The corpus is the oracle's answers: the oracle is a product (mutantOracleBinary, or documentOracle for the doc
		// mutation), a deterministic build of what go compiles, so it stands in this key by its own key (mutantOracleKey),
		// and only a build of the corpus resolves the binary. A reader keys the corpus without holding an oracle of about
		// 53 MB (Tests, Oct 12). Then the gaps file, and the code that runs it and captures its output. Not cohere whole:
		// its TypeScript submodule made a populated tree hash 67,000 files per lookup, up to 58 times a cold process
		// (#4ycgm2n).
		state.Directory = oracleExpressionBuild(t, MutantOracleCorpusInputs(root, index, MutantOracleKey(t, root, index)), func(directory string) error {
			var cases string
			if Mutations[index].Entry == "docMain.ts" {
				cases, _, _ = DocumentCorpus(t)
			} else {
				cases, _ = mutantOracle(t, Mutations[index], mutantOracleBinary(t, Mutations[index]))
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
				data, err := os.ReadFile(filepath.Join(filepath.Dir(cases), name))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
					return err
				}
			}
			return nil
		})
	})
	if state.Directory == "" {
		t.Fatal("oracle product preparation failed")
	}
	name := "cases.txt"
	if Mutations[index].Entry == "docMain.ts" {
		name = "docs.txt"
	}
	answers, err := os.ReadFile(filepath.Join(state.Directory, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(state.Directory, name), string(answers)
}

// mutantOracleCorpusInputs are mutation index's oracle corpus inputs under root: the oracle by its product's key, the
// gaps file, and the code that runs the oracle and captures its output.
func MutantOracleCorpusInputs(root string, index int, oracle string) PrinterBuildInputs {
	inputs := []string{root + "/internal/childguard"}
	for _, file := range []string{"mutants_test.go", "mutant_oracle_test.go", "mutant_products_test.go", "helpers_test.go", "doc_test.go", "shards_test.go", "testdata/notyet.json", "expressions_test.go"} {
		inputs = append(inputs, root+"/stage1/cohere/tsprinter/"+file)
	}
	return PrinterBuildInputs{
		Name:      fmt.Sprintf("TS printer mutant %03d Go oracle corpus", index),
		Files:     inputs,
		Flags:     []string{"oracle " + oracle, fmt.Sprintf("mutation=%+v", Mutations[index]), fmt.Sprintf("families=%q", MutantFamilies[Mutations[index].Name]), "oracle-width=80"},
		Toolchain: runtime.Version(),
	}
}

// mutantOracleKey is the key of mutation index's oracle product (documentOracle for the doc mutation, else
// mutantOracleBinary), computed without building or fetching it.
func MutantOracleKey(t *testing.T, root string, index int) string {
	t.Helper()
	// Only the doc mutation keys the doc oracle, so its package never joins another mutation's reads.
	var inputs buildcache.Inputs
	if Mutations[index].Entry == "docMain.ts" {
		inputs = documentOracleInputs(t)
	} else {
		inputs = mutantOracleInputs(t, Mutations[index])
	}
	key, err := buildcache.Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// Mutation tests need witnesses, while the agreement tests exercise every
// corpus input on unchanged ports. Select independent Go-oracle cases from
// the relevant generated families, keeping runs small even as tracked sources
// grow. Evenly spaced cases retain different operators, widths and contexts.
var MutantFamilies = map[string][]string{
	"program loses its statement separator":               {"program-sequence", "program-boundary"},
	"declaration loses its function keyword":              {"program-sequence", "program-boundary"},
	"assignment chain ignores statement boundaries":       {"statement-composition-regression", "variable-statement-edge"},
	"variable statement loses its keyword":                {"variable-statement-composition", "variable-statement-edge"},
	"await loses its keyword":                             {"await-composition", "await-yield-edge"},
	"yield loses delegation":                              {"yield-composition", "await-yield-edge"},
	"tagged template loses its tag":                       {"tagged-template-composition"},
	"template preview loses optional stopping boundaries": {"template-optional-boundary-composition"},
	"computed key forgets its clean text":                 {"computed-short-key-boundary"},
	"method loses its key":                                {"object-method-composition", "object-accessor-composition"},
	"function loses its keyword":                          {"named-function-composition"},
	"optional chain loses its stopping parentheses":       {"optional-chain-boundary"},
	"return and throw lose their keyword":                 {"named-function-composition", "arrow-composition"},
	"statement expression loses its semicolon":            {"named-function-composition", "statement-composition-regression"},
	"body loses its opening brace":                        {"named-function-composition", "arrow-composition"},
	"simple statement loses debugger":                     {"named-function-composition", "arrow-composition"},
	"arrow loses its arrow token":                         {"arrow-composition", "arrow-chain"},
	"template loses its interpolation dollar":             {"template-interpolation", "template-width"},
	"member chain loses its method dot":                   {"member-call-composition", "member-call-width"},
	"arguments lose their trailing comma":                 {"expanded-arguments", "argument-width"},
	"object loses its property colon":                     {"object-composition", "object-width"},
	"conditional swaps its separator":                     {"conditional-composition", "conditional-nesting"},
	"assignment loses its operator":                       {"assignment-composition", "assignment-chain"},
	"sequence loses a comma":                              {"sequence-composition", "sequence-width", "sequence"},
	"expression loses required parentheses":               {"operator-pair", "edge"},
}

type Mutation struct{ Name, File, From, To, Entry string }

var Mutations = []Mutation{
	{"protocol width confuses Number with parseFloat", "docMain.ts", "printWidth: Number(fields[1] ?? '')", "printWidth: Number.parseFloat(fields[1] ?? '')", "docMain.ts"},
	{"program loses its statement separator", "expressions.ts", "if(parts.length > 0) parts.push(this.docs.hardline());\n                    const previous =", "if(false) parts.push(this.docs.hardline());\n                    const previous =", "statementsMain.ts"},
	{"declaration loses its function keyword", "expressions.ts", "`${async}function${generator}${name}`", "`${async}${node.kind === 'FunctionDeclaration' ? '' : 'function'}${generator}${name}`", "statementsMain.ts"},
	{"assignment chain ignores statement boundaries", "expressions.ts", "!['ExpressionStatement', 'VariableDeclarationList', 'VariableStatement']", "!['UnrelatedStatement', 'VariableDeclarationList', 'VariableStatement']", "main.ts"},
	{"hashbang loses its refusal", "expressions.ts", "source.startsWith('#!') || source.startsWith('\\ufeff#!') || ", "", "main.ts"},
	{"variable statement loses its keyword", "expressions.ts", "this.docs.text(kind),", "this.docs.text(''),", "main.ts"},
	{"await loses its keyword", "expressions.ts", "this.docs.text('await ')", "this.docs.text('void ')", "main.ts"},
	{"yield loses delegation", "expressions.ts", "delegated ? 'yield*' : 'yield'", "delegated ? 'yield' : 'yield'", "main.ts"},
	{"tagged template loses its tag", "expressions.ts", "this.print(this.child(id, 0), id, 'tag')", "this.docs.text('')", "main.ts"},
	{"template preview loses optional stopping boundaries", "expressions.ts", "for(const boundary of this.optionalBoundaries) flatPrinter.optionalBoundaries.add(boundary);", "for(const boundary of this.optionalBoundaries) flatPrinter.sequenceBoundaries.add(boundary);", "main.ts"},
	{"computed key forgets its clean text", "expressions.ts", "const keyText = this.docs.textContent(left);", "const keyText = this.docs.get(left).kind === 'text' ? this.docs.get(left).text : undefined;", "main.ts"},
	{"method loses its key", "expressions.ts", "this.propertyKey(index, offset)", "this.docs.text(\"\")", "main.ts"},
	{"function loses its keyword", "expressions.ts", "`${async}function${generator}${name}`", "`${async}${generator}${name}`", "main.ts"},
	{"optional chain loses its stopping parentheses", "parentheses.ts", "if(optionalBoundaries.has(index) && parent >= 0)", "if(false && parent >= 0)", "main.ts"},
	{"return and throw lose their keyword", "expressions.ts", "this.docs.text(node.kind === 'ReturnStatement' ? 'return' : 'throw')", "this.docs.text(node.kind === 'ReturnStatement' ? 'throw' : 'throw')", "main.ts"},
	{"statement expression loses its semicolon", "expressions.ts", "return this.docs.concat([printed, this.docs.text(';')]);", "return this.docs.concat([printed, this.docs.text('')]);", "main.ts"},
	{"body loses its opening brace", "expressions.ts", "this.docs.text('{'),\n                    this.docs.indent(", "this.docs.text('['),\n                    this.docs.indent(", "main.ts"},
	{"simple statement loses debugger", "expressions.ts", "return this.docs.text('debugger;');", "return this.docs.text(';');", "main.ts"},
	{"arrow loses its arrow token", "expressions.ts", "this.docs.text(' =>'),\n                chain ?", "this.docs.text(' '),\n                chain ?", "main.ts"},
	{"template loses its interpolation dollar", "expressions.ts", "this.docs.text('${')", "this.docs.text('{')", "main.ts"},
	{"member chain loses its method dot", "expressions.ts", "this.docs.text(this.ownOptional(index) ? '?.' : '.')", "this.docs.text(this.ownOptional(index) ? '?.' : '')", "main.ts"},
	{"arguments lose their trailing comma", "expressions.ts", "const trailing = this.docs.ifBreak(this.docs.text(','), this.docs.text(''));", "const trailing = this.docs.text('');", "main.ts"},
	{"object loses its property colon", "expressions.ts", "this.docs.text(':')", "this.docs.text('')", "main.ts"},
	{"conditional swaps its separator", "expressions.ts", "this.docs.text('? ')", "this.docs.text(': ')", "main.ts"},
	{"assignment loses its operator", "expressions.ts", "this.docs.text(` ${this.operator(index)}`)", "this.docs.text(' ')", "main.ts"},
	{"sequence loses a comma", "expressions.ts", "parts.push(this.docs.text(','));", "parts.push(this.docs.text(''));", "main.ts"},
	{"group ignores width", "doc.ts", "!document.broken &&\n                        this.fits(flat", "!document.broken ||\n                        this.fits(flat", "docMain.ts"},
	{"fill never packs a pair", "doc.ts", "separatorMode = fitsPair ? 2 : 1;", "separatorMode = 1;", "docMain.ts"},
	{"expression loses required parentheses", "expressions.ts", "return this.parenthesize(id, parent, role)", "return parent < -1 && this.parenthesize(id, parent, role)", "main.ts"},
}

const MutantsShards = 29

func printerGateTSCCorpus(t *testing.T, family string) {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	files, err := TrackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
	if err != nil {
		t.Fatal(err)
	}
	gaps, _ := filepath.Abs("testdata/notyet.json")
	tsPrinterOracleOutputs(t, root, files, gaps, family, PrinterGateGoOracle(t)+"/oracle")
}

// The partition and the npm library's hash are the process's, made once; each
// product writes only its own shard's files.
func statementDeclaredLibraryOracle(t *testing.T, number, side int) {
	t.Helper()
	shard := StatementGateShardsOnce(t).Shard(t, number)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER")
	}
	StatementLibraryOracle(t, number, side, shard, StatementLibraryHashOnce(t), library)
}

func expressionDeclaredLibraryOracle(t *testing.T, number, side int) {
	t.Helper()
	units := ExpressionGateUnitsOnce(t)
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER")
	}
	ExpressionLibraryOracle(t, number, side, units.Unit(t, number), StatementDirectoryHash(t, library), library)
}

type printerGateProductState struct {
	once      sync.Once
	directory string
}

var printerGateProducts sync.Map

// Only process-local initialization is held here. Every persistent product is
// owned, keyed and published by internal/buildcache.
func PrinterGateOnce(t *testing.T, name string, build func() string) string {
	t.Helper()
	value, _ := printerGateProducts.LoadOrStore(name, &printerGateProductState{})
	state := value.(*printerGateProductState)
	state.once.Do(func() { state.directory = build() })
	if state.directory == "" {
		t.Fatalf("product %s preparation failed", name)
	}
	return state.directory
}

// printerGateGoOracle is the directory holding the Go printer oracle, cohere's internal/format/javascript test binary
// with the two side files laid over it: a buildcache.GoTest product, keyed by the files of the packages go compiles,
// the overlay's map and the side files' content, and go's settings. Its key declared all of cohere, the TypeScript
// submodule's 41,000 files included, which cost a populated tree's first lookup 15 to 60 s (#4ycgm2n).
func PrinterGateGoOracle(t *testing.T) string {
	t.Helper()
	return PrinterGateOnce(t, "go-oracle", func() string {
		root, err := filepath.Abs(Repository)
		if err != nil {
			t.Fatal(err)
		}
		replace := map[string]string{}
		for _, family := range []string{"expressions", "statements"} {
			replace[root+"/cohere/internal/format/javascript/adamic_"+family+"_test.go"] = root + "/stage1/cohere/tsprinter/testdata/" + family + "_side_test.go"
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}
		// The overlay names this checkout's paths: GoTest keys its map and the side files' content, never its own path.
		path := filepath.Join(t.TempDir(), "overlay.json")
		if err := os.WriteFile(path, overlay, 0o644); err != nil {
			t.Fatal(err)
		}
		return filepath.Dir(goTestOracle(t, "TestProduct_TSPrinterGoOracle", "oracle", "./internal/format/javascript", []string{"-overlay=" + path}, nil))
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

func PlanCorpus(t *testing.T, path string) *CorpusPlan {
	t.Helper()
	if cached, ok := corpusPlans.Load(path); ok {
		return cached.(*CorpusPlan)
	}
	base := strings.TrimSuffix(path, filepath.Ext(path))
	textPath, specsPath := base+".txt", base+".json"
	text, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(specsPath)
	if err != nil {
		t.Fatal(err)
	}
	var specs []json.RawMessage
	if err := json.Unmarshal(data, &specs); err != nil {
		t.Fatal(err)
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
		t.Fatalf("%s: %d protocol cases, %d specifications", path, len(rows), len(specs))
	}
	plan := &CorpusPlan{Labels: make([]string, len(rows))}
	weights := make([]int, len(rows))
	for i, raw := range specs {
		var item PrinterCase
		if err := json.Unmarshal(raw, &item); err != nil {
			t.Fatal(err)
		}
		plan.Labels[i] = item.Label
		if item.Label == "" {
			plan.Labels[i] = fmt.Sprintf("%s:case %d", specsPath, i)
		}
		weights[i] = len(rows[i])
	}
	directory := t.TempDir()
	// The test's share of the unit, never the host's cores (testguard.Share).
	for number, indices := range AssignCorpus(weights, min(testguard.Share(), 8)) {
		var batch strings.Builder
		subset := make([]json.RawMessage, 0, len(indices))
		for _, index := range indices {
			batch.WriteString(rows[index])
			subset = append(subset, specs[index])
		}
		shard := CorpusShard{Indices: indices, Text: filepath.Join(directory, fmt.Sprintf("shard-%d.txt", number)), Specs: filepath.Join(directory, fmt.Sprintf("shard-%d.json", number))}
		if err := os.WriteFile(shard.Text, []byte(batch.String()), 0644); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.Specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		plan.Shards = append(plan.Shards, shard)
	}
	corpusPlans.Store(path, plan)
	// Either extension resolves to the same immutable plan, retained by the
	// corpus-owning test until all its parallel mutant subtests finish.
	corpusPlans.Store(textPath, plan)
	corpusPlans.Store(specsPath, plan)
	t.Cleanup(func() { corpusPlans.Delete(textPath); corpusPlans.Delete(specsPath) })
	t.Logf("%s: %d cases in %d largest-first processes", filepath.Base(base), len(rows), len(plan.Shards))
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
			return nil, fmt.Errorf("shard %d %s: missing or extra case: %d answers, want %d", number, plan.Labels[index], len(lines), len(shard.Indices))
		}
		if !bytes.HasSuffix(output, []byte("\n")) {
			return nil, fmt.Errorf("shard %d %s: unterminated answer", number, plan.Labels[shard.Indices[len(shard.Indices)-1]])
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
func StatementGateShardsOnce(t *testing.T) *StatementGateShards {
	t.Helper()
	// printerCorpusFiles skips without a TypeScript source. Skip here, before the
	// once, so every test skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" {
		printerCorpusFiles(t)
	}
	shared := &statementGateShardsShared
	shared.once.Do(func() {
		cases, want, specs := statementCorpus(t)
		gate := &StatementGateShards{Cases: cases, Want: want, Specs: specs, builder: t.Name()}
		for _, shard := range StatementShards(t, cases, want, specs) {
			text, err := os.ReadFile(shard.Text)
			if err != nil {
				t.Fatal(err)
			}
			specifications, err := os.ReadFile(shard.Specs)
			if err != nil {
				t.Fatal(err)
			}
			gate.Shards = append(gate.Shards, statementShard{Indices: shard.Indices, Labels: shard.Labels, Want: shard.Want})
			gate.texts, gate.specifications = append(gate.texts, text), append(gate.specifications, specifications)
		}
		shared.shards = gate
	})
	if shared.shards == nil {
		t.Fatal("statement shard preparation failed")
	}
	if shared.shards.builder != t.Name() {
		t.Logf("union: checked once in this process by %s", shared.shards.builder)
	}
	return shared.shards
}

// One shard's protocol and specifications, written where its own test reads them.
func (gate *StatementGateShards) Shard(t *testing.T, number int) statementShard {
	t.Helper()
	directory := t.TempDir()
	shard := gate.Shards[number]
	shard.Text = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.txt", number))
	shard.Specs = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.json", number))
	if err := os.WriteFile(shard.Text, gate.texts[number], 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shard.Specs, gate.specifications[number], 0644); err != nil {
		t.Fatal(err)
	}
	return shard
}

var statementLibraryHashShared struct {
	once sync.Once
	hash string
}

// Every npm library oracle keys by the bytes of the ADAMIC_TS_PRETTIER install,
// which each shard and each library oracle product hashed whole again.
func StatementLibraryHashOnce(t *testing.T) string {
	t.Helper()
	shared := &statementLibraryHashShared
	shared.once.Do(func() { shared.hash = StatementDirectoryHash(t, os.Getenv("ADAMIC_TS_PRETTIER")) })
	if shared.hash == "" {
		t.Fatal("npm Prettier hash failed")
	}
	return shared.hash
}

func statementProduct(t *testing.T, inputs buildcache.Inputs, build func(dir string) error) string {
	t.Helper()
	return oracleProduct(t, inputs, func(dir string) error {
		start, cpu := time.Now(), StatementCPU()
		if err := build(dir); err != nil {
			return err
		}
		t.Logf("build %s: %.3fs, %.3f CPU s", inputs.Name, time.Since(start).Seconds(), StatementCPU()-cpu)
		return nil
	})
}

func StatementBytesHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func statementFileHash(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return StatementBytesHash(data)
}

// External npm inputs cannot appear in repository-relative Inputs.Files.
// WalkDir visits names lexically; hash every name and byte, refusing symlinks.
func StatementDirectoryHash(t *testing.T, directory string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("oracle input symlink: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%d:%s:%d:", len(name), name, len(data))
		hash.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Linux accounting includes all reaped descendants (oracle, Node and clang).
func StatementCPU() float64 {
	var self, children syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &self); err != nil {
		panic(err)
	}
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children); err != nil {
		panic(err)
	}
	return float64(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)/1e6
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
			path, err := filepath.Abs(tsPrinterOutputLabel(item.Label[:start], root, upstream, false))
			if err != nil {
				return nil, err
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return nil, err
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				relative, err = filepath.Rel(upstream, path)
				if err != nil {
					return nil, err
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
	if err := StatementUnion(len(keys), shards); err != nil {
		return nil, err
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

func StatementShards(t *testing.T, cases, want, specs string) []statementShard {
	t.Helper()
	text, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(specs)
	if err != nil {
		t.Fatal(err)
	}
	var items []PrinterCase
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSuffix(string(text), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(rows) != len(items) || len(answers) != len(items) {
		t.Fatalf("enumeration: %d inputs, %d specifications, %d answers", len(rows), len(items), len(answers))
	}
	// The shared comparator's full-corpus upstream-presence check belongs here;
	// each leaf then checks the exact pinned outcomes for its own cases.
	for _, record := range UpstreamOutcomes(t) {
		if record.Family != "statements" {
			continue
		}
		found := false
		for _, item := range items {
			if strings.HasSuffix(item.Label, record.Label) && item.Source == record.Source {
				if item.Want != record.Go {
					t.Fatalf("%s: recorded Go outcome changed", record.Label)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("recorded upstream case absent: %s", record.Label)
		}
	}
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := StatementCaseKeys(items, root, upstream)
	if err != nil {
		t.Fatal(err)
	}
	partition, err := StatementPartition(keys)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	result := make([]statementShard, len(partition))
	projected := make([][]int, len(partition))
	for number, ids := range partition {
		shard := statementShard{Indices: ids, Text: filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.txt", number)), Specs: filepath.Join(dir, fmt.Sprintf("statement-shard-%03d.json", number))}
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
		encoded, err := json.Marshal(subset)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.Text, []byte(input.String()), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(shard.Specs, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		result[number] = shard
	}
	if err := StatementUnion(len(items), projected); err != nil {
		t.Fatal(err)
	}
	return result
}

// Case commands use the shard deadline. Setup has no clock of its own;
// cancellation still kills the entire child process group.
func statementCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	if value, ok := StatementCaseContexts.Load(t); ok {
		return statementContextCommand(value.(context.Context), name, arguments...)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return statementContextCommand(ctx, name, arguments...)
}

func statementContextCommand(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		group := command.Process.Pid
		if command.SysProcAttr.Pgid != 0 {
			group = command.SysProcAttr.Pgid
		}
		err := syscall.Kill(-group, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func StatementExecute(t *testing.T, environment []string, name string, arguments ...string) Run {
	t.Helper()
	command := statementCommand(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}

func statementCorpusBuild(t *testing.T) (string, string, string) {
	t.Helper()
	directory := t.TempDir()
	root, _ := filepath.Abs(Repository)
	files := printerCorpusFiles(t)
	// This count is fixed by the upstream commit, never by the live case total.
	upstream, err := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	upstreamFiles, repositoryFiles := 0, 0
	for _, file := range files {
		path, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(upstream, path)
		if err != nil {
			t.Fatal(err)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			upstreamFiles++
		} else {
			repositoryFiles++
		}
	}
	if upstreamFiles != 77 {
		t.Fatalf("TypeScript 050880ce src/compiler: %d .ts files, want pinned 77", upstreamFiles)
	}
	if repositoryFiles == 0 {
		t.Fatal("repository statement corpus is empty")
	}
	oracleDir := PrinterGateGoOracle(t)
	flags := []string{"Go oracle sha256=" + statementFileHash(t, oracleDir+"/oracle"), "width=80"}
	// An input under ADAMIC_TYPESCRIPT_SOURCE is named by the tree it is in, not where this machine keeps it (#t37sw0f).
	for _, file := range files {
		path, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		flags = append(flags, "input="+tsPrinterOutputLabel(path, root, upstream, true)+" sha256="+statementFileHash(t, file))
	}
	product := statementProduct(t, buildcache.Inputs{
		Name:  "tsprinter-statements-oracle-output-v1",
		Files: []string{"stage1/cohere/tsprinter/statements_test.go", "stage1/cohere/tsprinter/statements_shards_test.go"},
		Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}, func(dir string) error {
		// Requests are scratch inputs. Publish only the oracle's four output files.
		request, err := json.Marshal(map[string]any{"Files": files, "Directory": dir})
		if err != nil {
			return err
		}
		if err := os.WriteFile(directory+"/request.json", request, 0644); err != nil {
			return err
		}
		command := statementCommand(t, oracleDir+"/oracle", "-test.v", "-test.count=1", "-test.run=^TestAdamicStatementCorpus$", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_TS_STATEMENT_REQUEST="+directory+"/request.json")
		output, err := CombinedOutput(command)
		if err != nil {
			return fmt.Errorf("Go statement corpus: %w\n%s", err, output)
		}
		t.Log(string(output))
		// The oracle labels each case by its file's path here; the product names it repo/ or typescript/ and the path
		// inside, as the expression outputs do, so a runner reading Workshop's product keys the same cases (#tqrqx60).
		return TSPrinterRewriteMetadata(dir, root, upstream, true)
	})
	directory = product
	if keep := os.Getenv("ADAMIC_TS_STATEMENT_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	PlanCorpus(t, directory+"/cases.txt")
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const StatementsAgainstGoAndPrettierShards = 16

func StatementLibraryOracle(t *testing.T, number, side int, shard statementShard, libraryHash, library string) string {
	t.Helper()
	name, script, files, oracleFlags := "npm", "testdata/expressions.mjs", []string{"stage1/cohere/tsprinter/testdata/expressions.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "stage1/cohere/tsprinter/testdata/prettier-differences.json"}, []string{"npm bytes sha256=" + libraryHash}
	if side == 1 {
		name, script = "embedded", "testdata/embedded.mjs"
		library, _ = filepath.Abs(filepath.Join(Repository, "cohere/internal/format/prettier/bundles"))
		files = []string{"stage1/cohere/tsprinter/testdata/embedded.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "cohere/internal/format/prettier/bundles"}
		oracleFlags = nil
	}
	script, _ = filepath.Abs(script)
	return PrinterGateOnce(t, fmt.Sprintf("statement-library-%d-%d", number, side), func() string {
		// The cases name their files by absolute path, hashed out of Key's reach, so the hash reads them as the tree spells
		// them; the npm library is keyed by its bytes (npm bytes sha256), the embedded one by Files, neither by where it is,
		// and NODE_PATH is a location (#t37sw0f). The test binary isn't an input: the oracle is node running the script over
		// the cases, and hashing the binary keyed a product apart between Workshop's go test and a runner's prebuilt
		// binary of the same package (landable-10, key check).
		root, _ := filepath.Abs(Repository)
		upstream, _ := filepath.Abs(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"))
		specs, err := os.ReadFile(shard.Specs)
		if err != nil {
			t.Fatal(err)
		}
		cases := strings.ReplaceAll(strings.ReplaceAll(string(specs), upstream+"/", "typescript/"), root+"/", "repo/")
		flags := append(oracleFlags, "cases sha256="+StatementBytesHash([]byte(cases)), "library="+name, "script="+script, "width=80", "NODE_OPTIONS="+os.Getenv("NODE_OPTIONS"))
		return statementProduct(t, buildcache.Inputs{
			Name:  fmt.Sprintf("tsprinter-statements-%s-oracle-shard-%03d-v1", name, number),
			Files: files, Flags: flags,
			Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
		}, func(dir string) error {
			result := StatementExecute(t, nil, "node", script, library, shard.Specs)
			if result.ExitCode != 0 || len(result.Stderr) != 0 {
				return fmt.Errorf("%s oracle exit %d: %s", name, result.ExitCode, result.Stderr)
			}
			return os.WriteFile(filepath.Join(dir, "answers.txt"), result.Stdout, 0644)
		})
	})
}

func statementCorpus(t *testing.T) (string, string, string) {
	t.Helper()
	directory := PrinterGateOnce(t, "statement-corpus", func() string { cases, _, _ := statementCorpusBuild(t); return filepath.Dir(cases) })
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
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

// tscAgreementRecipe are the files whose code makes the corpus product's bytes: its build here, the TSC filter, the
// oracle outputs it asks for and relabels, the oracle it names, the tracked listing of the corpus, and how the oracle
// runs and its output is captured (statementContextCommand; bounded and combinedOutput). printerGateOnce, in
// printer_gate_products_test.go, is only a per-process memo, and the oracle it builds is keyed by its bytes.
var TSCAgreementRecipe = []string{
	"stage1/cohere/tsprinter/tsc_corpus_split_test.go",
	"stage1/cohere/tsprinter/tsc_corpus_test.go",
	"stage1/cohere/tsprinter/tsc_units_test.go",
	"stage1/cohere/tsprinter/expression_products_test.go",
	"stage1/cohere/tsprinter/expressions_test.go",
	"stage1/cohere/tsprinter/corpus_test.go",
	"stage1/cohere/tsprinter/statements_shards_test.go",
	"stage1/cohere/tsprinter/helpers_test.go",
}

// tscAgreementInputs are what the corpus product's build reads: the TSC corpus, the gaps file, the Go oracle that
// answers every case, by its bytes as expressionBuild keys a generated input, and the recipe. Keying every tracked file
// of the repository cost the first shard of every process 19 s of walking (#4ycgm2n).
func tscAgreementInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := os.ReadFile(tscOracle(t, root))
	if err != nil {
		t.Fatal(err)
	}
	return TSCAgreementKeyInputs(fmt.Sprintf("%x", sha256.Sum256(oracle)))
}

// tscAgreementKeyInputs is tscAgreementInputs for the oracle whose bytes hash to oracle.
func TSCAgreementKeyInputs(oracle string) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "TSC agreement corpus v3",
		Files:     append([]string{"stage3/drivers/tsc/corpus", "stage1/cohere/tsprinter/testdata/notyet.json"}, TSCAgreementRecipe...),
		Flags:     []string{fmt.Sprint(TSCCorpusAgreementShards), "oracle " + oracle},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

// Every independently selected leaf can prepare the corpus; no test ordering is required.
func TSCAgreementPrepare(t *testing.T) string {
	t.Helper()
	return oracleProduct(t, tscAgreementInputs(t), func(directory string) error {
		root, err := filepath.Abs(Repository)
		if err != nil {
			return err
		}
		files, err := TrackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
		if err != nil {
			return err
		}
		oracle := tscOracle(t, root)
		data := TSCAgreementData{}
		for _, statements := range []bool{false, true} {
			family, entry := "expressions", "main.ts"
			if statements {
				family, entry = "statements", "statementsMain.ts"
			}
			cases, original := tscCorpus(t, root, files, family, statements, oracle)
			output := filepath.Join(directory, family)
			if err := os.MkdirAll(output, 0755); err != nil {
				return err
			}
			for _, name := range []string{"coverage.json"} {
				content, err := os.ReadFile(filepath.Join(original, name))
				if err != nil {
					return err
				}
				// A refused file is named by the root the reader supplies (#tqrqx60).
				if err := os.WriteFile(filepath.Join(output, name), []byte(buildcache.Relative(string(content))), 0644); err != nil {
					return err
				}
			}
			// Each case's label names its corpus file by the root the reader supplies, never this checkout (#tqrqx60).
			for i := range cases {
				cases[i].Label = buildcache.Relative(cases[i].Label)
			}
			data.Families = append(data.Families, tscAgreementFamily{Name: family, Directory: family, Cases: cases, Source: entry})
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "manifest.json"), encoded, 0644)
	})
}

type PrinterCase struct{ Label, Source, Want string }

func tscCorpus(t *testing.T, root string, files []string, family string, statements bool, oracle string) ([]PrinterCase, string) {
	t.Helper()
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(t, root, files, gaps, family, oracle)
	data, err := os.ReadFile(filepath.Join(directory, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var selected, cases []PrinterCase
	if err := json.Unmarshal(data, &selected); err != nil {
		t.Fatal(err)
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
	if err := os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tsc-cases.json"), encoded, 0644); err != nil {
		t.Fatal(err)
	}

	return cases, directory
}

func tscOracle(t *testing.T, root string) string {
	t.Helper()
	return PrinterGateGoOracle(t) + "/oracle"
}

type upstreamOutcome struct {
	Family, Label, Source, Go, Prettier, Embedded, PrettierError, EmbeddedError string
}

func UpstreamOutcomes(t *testing.T) []upstreamOutcome {
	t.Helper()
	data, err := os.ReadFile("testdata/tsc-upstream-differences.json")
	if err != nil {
		t.Fatal(err)
	}
	var pinned struct {
		Version string
		Records []upstreamOutcome
	}
	if err := json.Unmarshal(data, &pinned); err != nil {
		t.Fatal(err)
	}
	if pinned.Version != "3.9.6" {
		t.Fatalf("upstream outcome version: %s", pinned.Version)
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
			return fmt.Errorf("oracle %s is missing: stage1/cohere/tsprinter/oracles' TestProduct_ units build it first", name)
		}
		return build(directory)
	}
}

// oracleProduct is buildcache.Product for an oracle product, keyed and gated as oracleInputs and oracleBuild say.
func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(directory string) error) string {
	t.Helper()
	inputs = oracleInputs(inputs)
	return buildcache.Product(t, inputs, oracleBuild(inputs.Name, build))
}

// goTestOracle is cohere's go test -c of pkg, mutated when mutations are given: built by owner, an oracles
// TestProduct_, and found by its key in a traced tsprinter process through buildcache.BuiltElsewhere, which never
// builds. Its name is buildcache's own, so it carries no oracle-owner-v1.
func goTestOracle(t *testing.T, owner, output, pkg string, arguments []string, mutations []buildcache.Mutation) string {
	t.Helper()
	if buildcache.Tracing() && !oracleProducer {
		inputs, err := buildcache.GoTestInputs("cohere", output, pkg, arguments, mutations)
		if err != nil {
			t.Fatal(err)
		}
		return filepath.Join(buildcache.BuiltElsewhere(t, "the Go oracle "+pkg, inputs, "stage1/cohere/tsprinter/oracles "+owner, func(string) error {
			return errors.New("built by stage1/cohere/tsprinter/oracles")
		}), output)
	}
	if len(mutations) > 0 {
		return buildcache.GoTestMutated(t, "cohere", output, pkg, arguments, mutations)
	}
	return buildcache.GoTest(t, "cohere", output, pkg, arguments)
}
