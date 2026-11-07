package gitignore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/gitignore/oracles"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

// ADAMIC_TEST_SHARD=i/n runs shard numbers congruent to i modulo n; unset runs all.
// Sixteen comparison buckets leave headroom as the live repository grows. Real-tree
// queries use repository-relative paths and directory mode, never walk positions.
// The remaining eleven units are the fixed port mutants, each checked on its part of
// the live corpus (portMutantCases), preserving their native and Node witnesses and
// Git comparisons.
const testThePortAnswersAsGoCohereAndGitDoShards = 27
const portComparisonShards = 16

// portLargestTree is the oracle's name for the tree holding an ignore file of exactly 100 MiB, asked about
// only when largest says so.
const portLargestTree = "an ignore file of exactly 100 MiB"

// portMutantCases is the part of the live corpus a mutant is run over: the 100 MiB ignore file's tree alone
// for a mutant only it can show (needsLargest), and the whole corpus but that tree for every other. Reading
// 100 MiB natively under the sanitizers was most of each mutant's run, and that tree's one pattern, beside a
// 100 MiB comment, and its two queries at the root show none of the others, so dropping it only drops
// seconds: each mutant must still differ from Go natively and on Node over its part, or its shard fails. git
// has no word on that tree, so a mutant's disagreement with git stands on the same trees either way.
func portMutantCases(asked cases, needsLargest bool) cases {
	part := asked
	part.Trees = nil
	for _, tr := range asked.Trees {
		if (tr.Name == portLargestTree) == needsLargest {
			part.Trees = append(part.Trees, tr)
		}
	}
	if needsLargest {
		part.Patterns, part.Globs = nil, nil
	}
	return part
}

type portShard struct {
	asked cases
	ids   []string
}

func portBucket(key string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum64() % portComparisonShards)
}

func portCaseShards(test *testing.T, asked cases) []portShard {
	test.Helper()
	shards := make([]portShard, portComparisonShards)
	expected := len(asked.Globs)
	for _, tr := range asked.Trees {
		expected += len(tr.Queries)
		if !strings.HasPrefix(tr.Name, "real ") {
			expected++
		}
	}
	for _, p := range asked.Patterns {
		expected += 1 + len(p.Queries)
	}
	live := map[string]bool{}
	add := func(key string, bucket int) {
		if live[key] {
			test.Fatalf("repeated live case %q", key)
		}
		live[key] = true
		shards[bucket].ids = append(shards[bucket].ids, key)
	}
	for _, tr := range asked.Trees {
		if strings.HasPrefix(tr.Name, "real ") {
			if len(tr.Queries) == 0 {
				test.Fatalf("empty repository corpus %q", tr.Name)
			}
			parts := make([]tree, portComparisonShards)
			for i := range parts {
				parts[i] = tr
				parts[i].Queries = nil
				parts[i].Expected = nil
			}
			for i, q := range tr.Queries {
				key := fmt.Sprintf("tree/%s/%s/%t", tr.Name, q.Path, q.IsDirectory)
				b := portBucket(key)
				add(key, b)
				parts[b].Queries = append(parts[b].Queries, q)
				if len(tr.Expected) > 0 {
					parts[b].Expected = append(parts[b].Expected, tr.Expected[i])
				}
			}
			for b, part := range parts {
				if len(part.Queries) > 0 {
					shards[b].asked.Trees = append(shards[b].asked.Trees, part)
				}
			}
		} else {
			key := "tree/" + tr.Name
			b := portBucket(key)
			add(key, b)
			for i, q := range tr.Queries {
				add(fmt.Sprintf("%s/query-%d/%s/%t", key, i, q.Path, q.IsDirectory), b)
			}
			shards[b].asked.Trees = append(shards[b].asked.Trees, tr)
		}
	}
	for _, p := range asked.Patterns {
		key := "patterns/" + p.Name
		b := portBucket(key)
		add(key, b)
		for i, q := range p.Queries {
			add(fmt.Sprintf("%s/query-%d/%s/%t", key, i, q.Path, q.IsDirectory), b)
		}
		shards[b].asked.Patterns = append(shards[b].asked.Patterns, p)
	}
	// Globs are enumerated by the pinned Git corpus and the fixed generated corpus.
	// Their pattern/text/mode keys stay stable even if an earlier corpus adds a case.
	occurrences := map[string]int{}
	for _, g := range asked.Globs {
		base := fmt.Sprintf("glob/%q/%q/%t", g.Pattern, g.Text, g.Path)
		index := occurrences[base]
		occurrences[base]++
		key := fmt.Sprintf("%s/%d", base, index)
		b := portBucket(key)
		add(key, b)
		shards[b].asked.Globs = append(shards[b].asked.Globs, g)
	}
	seen := map[string]bool{}
	total := 0
	for _, s := range shards {
		for _, id := range s.ids {
			if !live[id] || seen[id] {
				test.Fatalf("missing or repeated shard case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != expected || total != len(live) || len(seen) != len(live) {
		test.Fatalf("union: %d cases, live %d", total, len(live))
	}
	for id := range live {
		if !seen[id] {
			test.Fatalf("lost case %q", id)
		}
	}
	test.Logf(
		"union: %d live case ids, each exactly once (%d trees, %d pattern lists, %d globs)",
		total,
		len(asked.Trees),
		len(asked.Patterns),
		len(asked.Globs),
	)
	return shards
}

func portSelected(test *testing.T, index int) bool {
	test.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		test.Fatalf("ADAMIC_TEST_SHARD must be i/n")
	}
	i, e1 := strconv.Atoi(fields[0])
	n, e2 := strconv.Atoi(fields[1])
	if e1 != nil || e2 != nil || n <= 0 || i < 0 || i >= n {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index%n == i
}

// Build products are inputs. Each callback writes only inside its product directory.
func portBuildProduct(test *testing.T, applied *mutant) string {
	test.Helper()
	name := "correct"
	var flags []string
	if applied != nil {
		name = applied.name
		flags = []string{applied.file, applied.from, applied.to}
	}
	flags = append(flags, "sanitize=true", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	files := compilerinputs.Lowering()
	for _, f := range portFiles {
		files = append(files, "stage1/cohere/gitignore/"+f)
	}
	return buildcache.Product(test, buildcache.Inputs{
		Name:      "gitignore-port-" + name,
		Files:     files,
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}, func(dir string) error {
		for _, f := range portFiles {
			data, readError := os.ReadFile(f)
			if readError != nil {
				return readError
			}
			source := string(data)
			if applied != nil && applied.file == f {
				if strings.Count(source, applied.from) != 1 {
					return fmt.Errorf("mutant %q must change exactly one place", name)
				}
				source = strings.Replace(source, applied.from, applied.to, 1)
			}
			if writeError := os.WriteFile(filepath.Join(dir, f), []byte(source), 0644); writeError != nil {
				return writeError
			}
		}
		program := lowered(test, filepath.Join(dir, "main.ts"))
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.mjs"),
			[]byte(javascript.JavaScript(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		source := native.C(program)
		if buildError := native.Build(
			source,
			filepath.Join(dir, "port"),
			native.Options{Sanitize: true},
		); buildError != nil {
			return buildError
		}
		if runtime.GOOS == "darwin" {
			return native.Build(source, filepath.Join(dir, "port-leaks"), native.Options{})
		}
		return nil
	})
}

// portBuildOracleProduct is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's
// internal/gitignore: a buildcache.GoTest product, so go runs as toolchain resolves it in cohere, under its
// environment, and the key is go's own answer there (cohere's go.mod, go.sum and go.work, the overlay's content, go's
// settings and release), not the caller's GOFLAGS and a go version asked elsewhere.
func portBuildOracleProduct(test *testing.T) string { return oracles.Binary(test) }

func portOracle(test *testing.T, binary string, request map[string]any) {
	test.Helper()
	data, encodeError := json.Marshal(request)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	path := filepath.Join(test.TempDir(), "request.json")
	if writeError := os.WriteFile(path, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	result := portExecute(
		test,
		[]string{"ADAMIC_PORT_REQUEST=" + path},
		binary,
		"-test.run=^TestAdamicPortCases$",
		"-test.timeout=0",
	)
	if result.exitCode != 0 {
		test.Fatalf("Go oracle: exit %d: %s %s", result.exitCode, result.stdout, result.stderr)
	}
}

func portAsked(test *testing.T, binary, scratch string) cases {
	test.Helper()
	scratch, resolveSymlinksError := filepath.EvalSymlinks(scratch)
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	seed := int64(generatedSeed)
	if v := os.Getenv("COHERE_GITIGNORE_SEED"); v != "" {
		var parseError error
		seed, parseError = strconv.ParseInt(v, 10, 64)
		if parseError != nil {
			test.Fatal(parseError)
		}
	}
	generated := 40
	if v := os.Getenv("COHERE_GITIGNORE_GENERATED"); v != "" {
		var parseError error
		generated, parseError = strconv.Atoi(v)
		if parseError != nil {
			test.Fatal(parseError)
		}
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// The oracle makes its trees repositories with the git internal/toolchain resolves, under GitEnv, never its PATH's.
	git, pathError := toolchain.Path("git")
	if pathError != nil {
		test.Fatal(pathError)
	}
	output := filepath.Join(scratch, "cases.json")
	portOracle(test, binary, map[string]any{
		"mode":           "generate",
		"scratch":        scratch,
		"seed":           seed,
		"generated":      generated,
		"gitSource":      gitSource(),
		"realTrees":      []string{root, filepath.Join(root, "cohere")},
		"output":         output,
		"largest":        largest(),
		"git":            git,
		"gitEnvironment": toolchain.GitEnv(),
	})
	data, readError := os.ReadFile(output)
	if readError != nil {
		test.Fatal(readError)
	}
	var asked cases
	if decodeError := json.Unmarshal(data, &asked); decodeError != nil {
		test.Fatal(decodeError)
	}
	return asked
}

func portGoAnswers(test *testing.T, binary string, asked cases) string {
	test.Helper()
	dir := test.TempDir()
	data, encodeError := json.Marshal(asked)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	path := filepath.Join(dir, "cases.json")
	if writeError := os.WriteFile(path, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	output := filepath.Join(dir, "answers.txt")
	portOracle(test, binary, map[string]any{"mode": "answer", "cases": path, "output": output})
	data, readError := os.ReadFile(output)
	if readError != nil {
		test.Fatal(readError)
	}
	return string(data)
}

func portNative(test *testing.T, dir, path string) run {
	return portExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, filepath.Join(dir, "port"), path)
}

// The cases are shared only within a run; build products alone use the hash cache.
// Each independently selected leaf prepares immutable inputs once. TestMain
// owns scratch cleanup, including runs that select only setup or one shard.
var portShared struct {
	sync.Mutex
	once            sync.Once
	asked           cases
	shards          []portShard
	oracle, scratch string
	gitAnswers      map[string][]string
	realAnswers     map[string]map[query]string
	totals          agreement
	comparedShards  int

	// Each of portMutantCases's two parts' cases files and Go cohere's answers to them, without the 100 MiB
	// ignore file's tree first and then with it alone, the same for every mutant of the part (and the union,
	// for the first), so made once in a run: the 100 MiB ignore file alone takes seconds to write and answer.
	parts [2]portPart
}

type portPart struct {
	once     sync.Once
	asked    cases
	path     string
	want     string
	answered bool
}

func portHardDeadline(test *testing.T, label string, budget time.Duration) *time.Timer {
	test.Helper()
	return time.AfterFunc(
		budget,
		func() { fmt.Fprintf(os.Stderr, "cooked: %s exceeded 90s hard deadline\n", label); os.Exit(124) },
	)
}

func checkPortAnswerUnion(test *testing.T) {
	rest := preparePortPartAnswers(test, false)
	ownStart := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(ownStart).Seconds()) }()
	deadline := portHardDeadline(test, "union", 90*time.Second)
	defer deadline.Stop()
	if portComparisonShards+len(mutants) != testThePortAnswersAsGoCohereAndGitDoShards {
		test.Fatal("enumerated shard count differs from static planner count")
	}
	seenMutants := map[string]bool{}
	for _, m := range mutants {
		if seenMutants[m.name] {
			test.Fatalf("repeated mutant %q", m.name)
		}
		seenMutants[m.name] = true
	}
	test.Logf("mutant union: %d cases, each in exactly one top-level shard", len(seenMutants))
	asked := portShared.asked
	portCaseShards(test, asked)
	checkPortMutantParts(test, asked, largest())
	// Check the non-vacuity requirement over the live full corpus, independently of
	// which comparison shards the current gate invocation selects. Git's answers are
	// setup's, which asked git check-ignore of every tree that runs it; the 100 MiB
	// ignore file's tree, the one tree outside rest, is not one of them.
	a := againstGit(rest.asked, portShared.gitAnswers, rest.want)
	if a.compared == 0 || a.ignored == 0 || a.reincluded == 0 {
		test.Fatalf("git was asked too little: %+v", a)
	}
	for _, difference := range a.differences {
		test.Error(difference)
	}
}

// checkPortMutantParts holds portMutantCases's two parts to the corpus: every tree in exactly one, the 100 MiB
// ignore file's tree in the second exactly when the corpus holds it (withLargest), and every pattern list and
// glob in the first.
func checkPortMutantParts(test *testing.T, asked cases, withLargest bool) {
	test.Helper()
	rest, alone := portMutantCases(asked, false), portMutantCases(asked, true)
	owners := map[string]int{}
	for _, part := range []cases{rest, alone} {
		for _, tr := range part.Trees {
			owners[tr.Name]++
		}
	}
	for _, tr := range asked.Trees {
		if owners[tr.Name] != 1 {
			test.Fatalf("tree %q is in %d mutant parts", tr.Name, owners[tr.Name])
		}
	}
	if len(rest.Trees)+len(alone.Trees) != len(asked.Trees) ||
		len(rest.Patterns) != len(asked.Patterns) ||
		len(rest.Globs) != len(asked.Globs) {
		test.Fatalf(
			"mutant parts: %d and %d trees of %d, %d pattern lists of %d, %d globs of %d",
			len(rest.Trees),
			len(alone.Trees),
			len(asked.Trees),
			len(rest.Patterns),
			len(asked.Patterns),
			len(rest.Globs),
			len(asked.Globs),
		)
	}
	want := 0
	if withLargest {
		want = 1
	}
	if len(alone.Trees) != want || len(alone.Patterns) != 0 || len(alone.Globs) != 0 {
		test.Fatalf(
			"the 100 MiB part holds %d trees, %d pattern lists and %d globs, want %d tree alone",
			len(alone.Trees),
			len(alone.Patterns),
			len(alone.Globs),
			want,
		)
	}
	test.Logf(
		"mutant parts: %d trees, %d pattern lists and %d globs, and %d tree alone",
		len(rest.Trees),
		len(rest.Patterns),
		len(rest.Globs),
		len(alone.Trees),
	)
}

// Not parallel: owns cleanup of the package-wide shared corpus scratch.
func TestMain(m *testing.M) {
	code := m.Run()
	if portShared.scratch != "" {
		os.RemoveAll(portShared.scratch)
	}
	os.Exit(code)
}

func TestThePortAnswersAsGoCohereAndGitDo_Setup(test *testing.T) {
	test.Parallel()
	preparePortAnswerInputs(test)
}

func preparePortAnswerInputs(test *testing.T) {
	test.Helper()
	portShared.once.Do(func() {
		start := time.Now()
		oracle := portOracleProduct(test)
		scratch, temporaryDirectoryError := os.MkdirTemp("", "gitignore-shards-")
		if temporaryDirectoryError != nil {
			test.Fatal(temporaryDirectoryError)
		}
		portShared.scratch = scratch
		asked := portAsked(test, oracle, scratch)
		shards := portCaseShards(test, asked)
		gitAnswers := map[string][]string{}
		realAnswers := map[string]map[query]string{}
		for _, tr := range asked.Trees {
			if tr.Git != "run" {
				continue
			}
			answers := portGitCheckIgnore(test, tr)
			gitAnswers[tr.Name] = answers
			if strings.HasPrefix(tr.Name, "real ") {
				if len(answers) != len(tr.Queries) {
					test.Fatalf("%s: Git answered %d queries, asked %d", tr.Name, len(answers), len(tr.Queries))
				}
				byQuery := map[query]string{}
				for i, q := range tr.Queries {
					byQuery[q] = answers[i]
				}
				realAnswers[tr.Name] = byQuery
			}
		}
		portShared.asked, portShared.oracle = asked, oracle
		portShared.gitAnswers, portShared.realAnswers = gitAnswers, realAnswers
		portShared.shards = shards
		test.Logf("shared setup ready: %.3fs", time.Since(start).Seconds())
	})
	if portShared.shards == nil {
		test.Fatal("shared preparation failed")
	}
}

// preparePortPartAnswers writes portMutantCases's part's cases file into a directory of its own in the shared
// scratch, which TestMain removes, and asks Go cohere its answers, once in a run.
func preparePortPartAnswers(test *testing.T, needsLargest bool) *portPart {
	test.Helper()
	preparePortAnswerInputs(test)
	part := &portShared.parts[0]
	if needsLargest {
		part = &portShared.parts[1]
	}
	part.once.Do(func() {
		start := time.Now()
		directory, temporaryDirectoryError := os.MkdirTemp(portShared.scratch, "mutant-part-")
		if temporaryDirectoryError != nil {
			test.Fatal(temporaryDirectoryError)
		}
		part.asked = portMutantCases(portShared.asked, needsLargest)
		part.path = casesFile(test, directory, part.asked)
		part.want = portGoAnswers(test, portShared.oracle, part.asked)
		part.answered = true
		test.Logf(
			"shared mutant part answers ready (100 MiB tree alone: %t): %.3fs",
			needsLargest,
			time.Since(start).Seconds(),
		)
	})
	if !part.answered {
		test.Fatal("shared mutant part answers failed")
	}
	return part
}

func runPortAnswerShard(test *testing.T, index int) {
	if !portSelected(test, index) {
		return
	}
	if portComparisonShards+len(mutants) != testThePortAnswersAsGoCohereAndGitDoShards {
		test.Fatal("enumerated shard count differs from static planner count")
	}
	// Standalone selections prepare inputs without a test-side setup deadline.
	// No leaf deadline is active while acquiring or building shared inputs.
	setupStart := time.Now()
	preparePortAnswerInputs(test)
	shards, oracle := portShared.shards, portShared.oracle
	var applied *mutant
	var part *portPart
	if index >= portComparisonShards {
		m := mutants[index-portComparisonShards]
		if m.needsLargest && !largest() {
			test.Skip("ADAMIC_GITIGNORE_LARGEST enables the 100 MiB mutant")
		}
		applied = &m
		part = preparePortPartAnswers(test, m.needsLargest)
	}
	product := portProducts(test, applied)
	test.Logf("shared setup total: %.3fs", time.Since(setupStart).Seconds())
	ownStart := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(ownStart).Seconds()) }()
	deadline := portHardDeadline(test, fmt.Sprintf("shard-%03d", index), 90*time.Second)
	defer deadline.Stop()

	var subset cases
	var path, want string
	if index < portComparisonShards {
		subset = shards[index].asked
		path = casesFile(test, test.TempDir(), subset)
		want = portGoAnswers(test, oracle, subset)
	} else {
		subset, path, want = part.asked, part.path, part.want
	}
	gitAnswers := map[string][]string{}
	for _, tr := range subset.Trees {
		if tr.Git == "run" {
			if strings.HasPrefix(tr.Name, "real ") {
				for _, q := range tr.Queries {
					answer, ok := portShared.realAnswers[tr.Name][q]
					if !ok {
						test.Fatalf("missing shared Git answer for %s %v", tr.Name, q)
					}
					gitAnswers[tr.Name] = append(gitAnswers[tr.Name], answer)
				}
			} else {
				gitAnswers[tr.Name] = portShared.gitAnswers[tr.Name]
			}
		}
	}
	dir := product
	node := portNode(test, filepath.Join(dir, "main.ts"), path)
	// On Linux a comparison shard's one sanitized run is also its leak check: LeakSanitizer acts only at
	// exit, so the run answers as it does with detection off, and a leak is a report on stderr and a
	// nonzero exit, which the native side refuses. The 100 MiB ignore file made running twice cost
	// seconds. A mutant runs with detection off, so it is caught by its answers alone.
	leakChecked := index < portComparisonShards && runtime.GOOS == "linux"
	var nat run
	if leakChecked {
		nat = portExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(dir, "port"), path)
	} else {
		nat = portNative(test, dir, path)
	}
	if index < portComparisonShards {
		backend := portNode(test, filepath.Join(dir, "program.mjs"), path)
		for _, side := range []struct {
			name   string
			result run
		}{{"Node", node}, {"native", nat}, {"JavaScript backend", backend}} {
			if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
				if side.name == "native" && leakChecked {
					test.Fatalf(
						"native, leak detection on: exit %d, stderr %q",
						side.result.exitCode,
						side.result.stderr,
					)
				}
				test.Fatalf("%s: exit %d, stderr %q", side.name, side.result.exitCode, side.result.stderr)
			}
			if diff := firstDifference(string(side.result.stdout), want); diff != "" {
				test.Errorf("%s and Go cohere differ: %s", side.name, diff)
			}
		}
		switch runtime.GOOS {
		case "linux":
			// The native run above was the leak check.
		case "darwin":
			leak := portExecute(test, nil, "leaks", "--atExit", "--", filepath.Join(dir, "port-leaks"), path)
			if leak.exitCode != 0 {
				test.Errorf("leaks: exit %d: %s %s", leak.exitCode, leak.stdout, leak.stderr)
			}
		default:
			test.Fatalf("no leak check for %s", runtime.GOOS)
		}
		a := againstGit(subset, gitAnswers, string(node.stdout))
		for _, diff := range a.differences {
			test.Error(diff)
		}
		portShared.Lock()
		portShared.totals.compared += a.compared
		portShared.totals.ignored += a.ignored
		portShared.totals.reincluded += a.reincluded
		portShared.comparedShards++
		if portShared.comparedShards == portComparisonShards &&
			(portShared.totals.compared == 0 || portShared.totals.ignored == 0 || portShared.totals.reincluded == 0) {
			test.Errorf("git was asked too little: %+v", portShared.totals)
		}
		portShared.Unlock()
		test.Logf("%d live case ids", len(shards[index].ids))
	} else {
		m := mutants[index-portComparisonShards]
		for _, side := range []struct {
			name   string
			result run
		}{{"natively", nat}, {"on Node", node}} {
			if side.result.exitCode != 0 {
				test.Errorf("%s mutant exits %d: %s", side.name, side.result.exitCode, side.result.stderr)
				continue
			}
			diff := firstDifference(string(side.result.stdout), want)
			if diff == "" {
				test.Errorf("%s mutant %q survives", side.name, m.name)
				continue
			}
			test.Logf("%s mutant %q caught: %s", side.name, m.name, diff)
			if side.name == "on Node" && m.seenByGit {
				a := againstGit(subset, gitAnswers, string(side.result.stdout))
				if len(a.differences) == 0 {
					test.Error("mutant agrees with git")
				}
			}
		}
	}
}

// A disagreement planted in one live case is caught by exactly its owning shard.
func TestPortAnswerShardPlantedDisagreement(test *testing.T) {
	test.Parallel()
	deadline := portHardDeadline(test, "TestPortAnswerShardPlantedDisagreement", 90*time.Second)
	defer deadline.Stop()
	asked := cases{Patterns: []patterns{{Name: "planted", Queries: []query{{Path: "one"}}}}}
	shards := portCaseShards(test, asked)
	caught := 0
	owner := -1
	for i, s := range shards {
		got, want := "", ""
		for _, id := range s.ids {
			want += "answer\n"
			if id == "patterns/planted/query-0/one/false" {
				got += "disagreement\n"
			} else {
				got += "answer\n"
			}
		}
		if firstDifference(got, want) != "" {
			caught++
			owner = i
		}
	}
	if caught != 1 || owner != portBucket("patterns/planted") {
		test.Fatalf("planted disagreement caught %d times in shard %d", caught, owner)
	}
	test.Logf("planted disagreement caught only by shard-%03d", owner)
}

func TestThePortAnswersAsGoCohereAndGitDo_000(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 0)
}

func TestThePortAnswersAsGoCohereAndGitDo_001(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 1)
}

func TestThePortAnswersAsGoCohereAndGitDo_002(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 2)
}

func TestThePortAnswersAsGoCohereAndGitDo_003(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 3)
}

func TestThePortAnswersAsGoCohereAndGitDo_004(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 4)
}

func TestThePortAnswersAsGoCohereAndGitDo_005(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 5)
}

func TestThePortAnswersAsGoCohereAndGitDo_006(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 6)
}

func TestThePortAnswersAsGoCohereAndGitDo_007(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 7)
}

func TestThePortAnswersAsGoCohereAndGitDo_008(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 8)
}

func TestThePortAnswersAsGoCohereAndGitDo_009(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 9)
}

func TestThePortAnswersAsGoCohereAndGitDo_010(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 10)
}

func TestThePortAnswersAsGoCohereAndGitDo_011(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 11)
}

func TestThePortAnswersAsGoCohereAndGitDo_012(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 12)
}

func TestThePortAnswersAsGoCohereAndGitDo_013(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 13)
}

func TestThePortAnswersAsGoCohereAndGitDo_014(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 14)
}

func TestThePortAnswersAsGoCohereAndGitDo_015(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 15)
}

func TestThePortAnswersAsGoCohereAndGitDo_016(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 16)
}

func TestThePortAnswersAsGoCohereAndGitDo_017(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 17)
}

func TestThePortAnswersAsGoCohereAndGitDo_018(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 18)
}

func TestThePortAnswersAsGoCohereAndGitDo_019(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 19)
}

func TestThePortAnswersAsGoCohereAndGitDo_020(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 20)
}

func TestThePortAnswersAsGoCohereAndGitDo_021(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 21)
}

func TestThePortAnswersAsGoCohereAndGitDo_022(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 22)
}

func TestThePortAnswersAsGoCohereAndGitDo_023(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 23)
}

func TestThePortAnswersAsGoCohereAndGitDo_024(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 24)
}

func TestThePortAnswersAsGoCohereAndGitDo_025(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 25)
}

func TestThePortAnswersAsGoCohereAndGitDo_026(test *testing.T) {
	test.Parallel()
	runPortAnswerShard(test, 26)
}

// The mutant parts split a corpus with the 100 MiB ignore file's tree and one without it, every case in exactly
// one part, as the union test holds the live corpus's.
func TestPortMutantPartsSplitTheCorpus(test *testing.T) {
	test.Parallel()
	asked := cases{
		Trees:    []tree{{Name: "semantics"}, {Name: portLargestTree}, {Name: "real repository"}},
		Patterns: []patterns{{Name: "house ignore"}},
		Globs:    []glob{{Pattern: "foo", Text: "foo"}},
	}
	checkPortMutantParts(test, asked, true)
	asked.Trees = []tree{asked.Trees[0], asked.Trees[2]}
	checkPortMutantParts(test, asked, false)
}

func TestPortAnswerShardAssignmentSurvivesCorpusGrowth(test *testing.T) {
	test.Parallel()
	deadline := portHardDeadline(test, "TestPortAnswerShardAssignmentSurvivesCorpusGrowth", 90*time.Second)
	defer deadline.Stop()
	before := cases{Trees: []tree{
		{Name: "real repository", Queries: []query{{Path: "a.go"}, {Path: "directory", IsDirectory: true}}},
	}}
	old := portCaseShards(test, before)
	after := before
	after.Trees = append([]tree(nil), before.Trees...)
	after.Trees[0].Queries = append([]query{{Path: "new.go"}}, before.Trees[0].Queries...)
	grown := portCaseShards(test, after)
	owners := map[string]int{}
	for i, s := range grown {
		for _, id := range s.ids {
			owners[id] = i
		}
	}
	for i, s := range old {
		for _, id := range s.ids {
			if owners[id] != i {
				test.Fatalf("adding a repository file moved %q from shard-%03d to shard-%03d", id, i, owners[id])
			}
		}
	}
}

// Shared setup children have no test-side deadline. A leaf deadline covers only
// own work; Go's process timeout covers preparation and execution together.
func portCommand(test *testing.T, name string, args ...string) *exec.Cmd {
	test.Helper()
	command := exec.CommandContext(test.Context(), name, args...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node, leaks) is a tool: the one internal/toolchain resolves, with a tool's environment, not
		// PATH's, as toolchain.Command runs one under a context it has no form for. A program a test built (the Go
		// oracle, the port) runs as named, with this process's environment. Path's error fails the command as it
		// starts.
		if path, pathError := toolchain.Path(name); pathError != nil {
			command.Err = pathError
		} else {
			command = exec.CommandContext(test.Context(), path, args...)
			command.Env = toolchain.Env()
		}
	}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// portExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after
// the child exited, so an oracle, port, Node or leaks run that had finished could fail its shard under load
// (#n3gk22h's shape). Capture has no stall guard, and the runs lose nothing by it: childguard.Run's windows (two
// minutes of silence, thirty before the first byte) are longer than the 90 s Loom gives a test unit, which ends a
// hung run sooner.
func portExecute(test *testing.T, environment []string, name string, args ...string) run {
	test.Helper()
	command := portCommand(test, name, args...)
	if environment != nil {
		// On top of the environment the command already runs with: a tool's is toolchain's, not this process's.
		command.Env = append(command.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(test.Context(), command)
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func portNode(test *testing.T, path string, args ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return portExecute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, args...)...,
	)
}

func portGitCheckIgnore(test *testing.T, tree tree) []string {
	test.Helper()
	scratch := test.TempDir()
	emptyExclude := filepath.Join(scratch, "empty-excludes")
	if writeError := os.WriteFile(emptyExclude, nil, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	var input bytes.Buffer
	for _, query := range tree.Queries {
		input.WriteString(query.Path)
		input.WriteByte(0)
	}
	// GitEnv keeps git's configuration out of reach; these add the scratch home and no optional locks.
	environment := []string{"HOME=" + scratch, "XDG_CONFIG_HOME=" + scratch, "GIT_OPTIONAL_LOCKS=0"}
	arguments := []string{"-c", "core.excludesFile=" + emptyExclude, "-c", "core.ignorecase=false"}
	if !tracked.Git(tree.Disk) {
		// A Loom runner's source has no .git (#cyasrr4), and check-ignore needs a repository even with --no-index. An
		// empty one, with the tree as its work tree and an empty info/exclude, reads only the tree's .gitignore files,
		// as tracked does for its own ignore rules.
		repository := filepath.Join(scratch, "repository")
		initialize := toolchain.GitCommand("init", "-q", "--bare", repository)
		initialize.Env = append(initialize.Env, environment...)
		if output, combinedOutputError := initialize.CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("git init for %s: %v\n%s", tree.Disk, combinedOutputError, output)
		}
		if writeError := os.WriteFile(filepath.Join(repository, "info", "exclude"), nil, 0o644); writeError != nil {
			test.Fatal(writeError)
		}
		arguments = append(
			[]string{"--git-dir=" + repository, "--work-tree=" + tree.Disk, "-c", "core.bare=false"},
			arguments...,
		)
	}
	command := toolchain.GitCommandIn(
		tree.Disk,
		append(arguments, "check-ignore", "--no-index", "--verbose", "--non-matching", "-z", "--stdin")...,
	)
	command.Env = append(command.Env, environment...)
	command.Stdin = &input
	// Read through childguard.Capture for portExecute's reason: a WaitDelay could cut git's answers short under load.
	output, standardError, captureError := childguard.Capture(test.Context(), command)
	// check-ignore exits 1 when nothing is ignored, which is an answer, not a failure.
	var exitError *exec.ExitError
	if captureError != nil && !(errors.As(captureError, &exitError) && exitError.ExitCode() == 1) {
		test.Fatalf("git check-ignore in %s: %v\n%s", tree.Disk, captureError, standardError)
	}
	fields := bytes.Split(bytes.TrimSuffix(output, []byte{0}), []byte{0})
	if len(output) > 0 && len(fields)%4 != 0 {
		test.Fatalf("%s: git printed %d fields, not a multiple of four", tree.Name, len(fields))
	}
	var answers []string
	for index := 0; index+3 < len(fields); index += 4 {
		file, line, pattern, relative := string(
			fields[index],
		), string(fields[index+1]), string(fields[index+2]), string(fields[index+3])
		if file == "" {
			answers = append(answers, "0 ::\t"+relative)
			continue
		}
		ignored := "1"
		if strings.HasPrefix(pattern, "!") {
			ignored = "0"
		}
		answers = append(answers, fmt.Sprintf("%s %s:%s:%s\t%s", ignored, file, line, pattern, relative))
	}
	return answers
}
