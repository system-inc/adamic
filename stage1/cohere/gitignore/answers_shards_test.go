package gitignore

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// ADAMIC_TEST_SHARD=i/n runs shard numbers congruent to i modulo n; unset runs all.
// Sixteen comparison buckets leave headroom as the live repository grows. Real-tree
// queries use repository-relative paths and directory mode, never walk positions.
// The remaining eleven units are the fixed port mutants, each checked on the entire
// live corpus, preserving their native and Node witnesses and Git comparisons.
const testThePortAnswersAsGoCohereAndGitDoShards = 27
const portComparisonShards = 16

type portShard struct {
	asked cases
	ids   []string
}

func portBucket(key string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum64() % portComparisonShards)
}

func portCaseShards(t *testing.T, asked cases) []portShard {
	t.Helper()
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
			t.Fatalf("repeated live case %q", key)
		}
		live[key] = true
		shards[bucket].ids = append(shards[bucket].ids, key)
	}
	for _, tr := range asked.Trees {
		if strings.HasPrefix(tr.Name, "real ") {
			if len(tr.Queries) == 0 {
				t.Fatalf("empty repository corpus %q", tr.Name)
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
				t.Fatalf("missing or repeated shard case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if total != expected || total != len(live) || len(seen) != len(live) || total == 0 {
		t.Fatalf("union: %d cases, live %d", total, len(live))
	}
	for id := range live {
		if !seen[id] {
			t.Fatalf("lost case %q", id)
		}
	}
	t.Logf("union: %d live case ids, each exactly once (%d trees, %d pattern lists, %d globs)", total, len(asked.Trees), len(asked.Patterns), len(asked.Globs))
	return shards
}

func portSelected(t *testing.T, index int) bool {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return true
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("ADAMIC_TEST_SHARD must be i/n")
	}
	i, e1 := strconv.Atoi(fields[0])
	n, e2 := strconv.Atoi(fields[1])
	if e1 != nil || e2 != nil || n <= 0 || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return index%n == i
}

// Build products are inputs. Each callback writes only inside its product directory.
func portProducts(t *testing.T, applied *mutant) string {
	t.Helper()
	name := "correct"
	var flags []string
	if applied != nil {
		name = applied.name
		flags = []string{applied.file, applied.from, applied.to}
	}
	flags = append(flags, "sanitize=true", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	files := []string{"internal", "bridge", "go.mod", "go.work", "cohere"}
	for _, f := range portFiles {
		files = append(files, "stage1/cohere/gitignore/"+f)
	}
	return buildcache.Product(t, buildcache.Inputs{Name: "gitignore-port-" + name, Files: files, Flags: flags, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")}}, func(dir string) error {
		for _, f := range portFiles {
			data, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			source := string(data)
			if applied != nil && applied.file == f {
				if strings.Count(source, applied.from) != 1 {
					return fmt.Errorf("mutant %q must change exactly one place", name)
				}
				source = strings.Replace(source, applied.from, applied.to, 1)
			}
			if err := os.WriteFile(filepath.Join(dir, f), []byte(source), 0644); err != nil {
				return err
			}
		}
		program := lowered(t, filepath.Join(dir, "main.ts"))
		if err := os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644); err != nil {
			return err
		}
		source := native.C(program)
		if err := native.Build(source, filepath.Join(dir, "port"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		if runtime.GOOS == "darwin" {
			return native.Build(source, filepath.Join(dir, "port-leaks"), native.Options{})
		}
		return nil
	})
}

func portOracleProduct(t *testing.T) string {
	t.Helper()
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	dir := buildcache.Product(t, buildcache.Inputs{Name: "gitignore-go-oracle", Files: []string{"cohere", "stage1/cohere/gitignore/testdata/cohere_side_test.go", "go.work", "go.mod"}, Flags: []string{"go test -c -overlay", "GOFLAGS=" + os.Getenv("GOFLAGS")}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal", "gitignore", "adamic_port_side_test.go"): side}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err = os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		cmd := bounded(t, "go", "test", "-c", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), "./internal/gitignore")
		cmd.Dir = cohere
		output, err := combinedOutput(cmd)
		if err != nil {
			return fmt.Errorf("%w: %s", err, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}

func portOracle(t *testing.T, binary string, request map[string]any) {
	t.Helper()
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "request.json")
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	result := execute(t, []string{"ADAMIC_PORT_REQUEST=" + path}, binary, "-test.run=^TestAdamicPortCases$", "-test.timeout=75s")
	if result.exitCode != 0 {
		t.Fatalf("Go oracle: exit %d: %s %s", result.exitCode, result.stdout, result.stderr)
	}
}

func portAsked(t *testing.T, binary, scratch string) cases {
	t.Helper()
	scratch, err := filepath.EvalSymlinks(scratch)
	if err != nil {
		t.Fatal(err)
	}
	seed := int64(generatedSeed)
	if v := os.Getenv("COHERE_GITIGNORE_SEED"); v != "" {
		seed, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
	}
	generated := 40
	if v := os.Getenv("COHERE_GITIGNORE_GENERATED"); v != "" {
		generated, err = strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(scratch, "cases.json")
	portOracle(t, binary, map[string]any{"mode": "generate", "scratch": scratch, "seed": seed, "generated": generated, "gitSource": gitSource(), "realTrees": []string{root, filepath.Join(root, "cohere")}, "output": output, "largest": largest()})
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var asked cases
	if err = json.Unmarshal(data, &asked); err != nil {
		t.Fatal(err)
	}
	return asked
}

func portGoAnswers(t *testing.T, binary string, asked cases) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(asked)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "cases.json")
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "answers.txt")
	portOracle(t, binary, map[string]any{"mode": "answer", "cases": path, "output": output})
	data, err = os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func portNative(t *testing.T, dir, path string) run {
	return execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, filepath.Join(dir, "port"), path)
}

// The cases are shared only within a run; build products alone use the hash cache.
// Every selected top-level test retains the scratch tree before calling Parallel,
// so the final test can remove it without racing a reader.
var portShared struct {
	sync.Mutex
	asked           cases
	shards          []portShard
	oracle, scratch string
	refs            int
	totals          agreement
	comparedShards  int
}

func portHardDeadline(t *testing.T, label string, budget time.Duration) *time.Timer {
	t.Helper()
	return time.AfterFunc(budget, func() { fmt.Fprintf(os.Stderr, "cooked: %s exceeded 75s hard deadline\n", label); os.Exit(124) })
}

func checkPortAnswerUnion(t *testing.T) {
	deadline := portHardDeadline(t, "union", 75*time.Second)
	defer deadline.Stop()
	if portComparisonShards+len(mutants) != testThePortAnswersAsGoCohereAndGitDoShards {
		t.Fatal("enumerated shard count differs from static planner count")
	}
	oracle := portOracleProduct(t)
	seenMutants := map[string]bool{}
	for _, m := range mutants {
		if seenMutants[m.name] {
			t.Fatalf("repeated mutant %q", m.name)
		}
		seenMutants[m.name] = true
	}
	t.Logf("mutant union: %d cases, each in exactly one top-level shard", len(seenMutants))
	asked := portAsked(t, oracle, t.TempDir())
	portCaseShards(t, asked)
	// Check the non-vacuity requirement over the live full corpus, independently of
	// which comparison shards the current gate invocation selects.
	gitAnswers := map[string][]string{}
	for _, tr := range asked.Trees {
		if tr.Git == "run" {
			gitAnswers[tr.Name] = gitCheckIgnore(t, tr)
		}
	}
	a := againstGit(asked, gitAnswers, portGoAnswers(t, oracle, asked))
	if a.compared == 0 || a.ignored == 0 || a.reincluded == 0 {
		t.Fatalf("git was asked too little: %+v", a)
	}
	for _, difference := range a.differences {
		t.Error(difference)
	}
}

func runPortAnswerShard(t *testing.T, index int) {
	if !portSelected(t, index) {
		return
	}
	if portComparisonShards+len(mutants) != testThePortAnswersAsGoCohereAndGitDoShards {
		t.Fatal("enumerated shard count differs from static planner count")
	}
	setupStart := time.Now()
	deadline := portHardDeadline(t, fmt.Sprintf("shard-%03d", index), 75*time.Second)
	defer deadline.Stop()
	portShared.Lock()
	if portShared.shards == nil {
		portShared.oracle = portOracleProduct(t)
		scratch, err := os.MkdirTemp("", "gitignore-shards-")
		if err != nil {
			portShared.Unlock()
			t.Fatal(err)
		}
		portShared.scratch = scratch
		portShared.asked = portAsked(t, portShared.oracle, scratch)
		portShared.shards = portCaseShards(t, portShared.asked)
	}
	portShared.refs++
	asked, shards, oracle := portShared.asked, portShared.shards, portShared.oracle
	portShared.Unlock()
	t.Cleanup(func() {
		portShared.Lock()
		defer portShared.Unlock()
		portShared.refs--
		if portShared.refs == 0 {
			os.RemoveAll(portShared.scratch)
			portShared.shards = nil
			portShared.totals = agreement{}
			portShared.comparedShards = 0
		}
	})
	setupElapsed := time.Since(setupStart)
	deadline.Stop()
	t.Parallel()
	deadline = portHardDeadline(t, fmt.Sprintf("shard-%03d", index), 75*time.Second-setupElapsed)
	defer deadline.Stop()
	t.Logf("setup: %.3fs", setupElapsed.Seconds())
	start := time.Now()
	var applied *mutant
	if index >= portComparisonShards {
		m := mutants[index-portComparisonShards]
		if m.needsLargest && !largest() {
			t.Skip("ADAMIC_GITIGNORE_LARGEST enables the 100 MiB mutant")
		}
		applied = &m
	}
	product := portProducts(t, applied)
	t.Logf("build fetch: %.3fs", time.Since(start).Seconds())

	subset := asked
	if index < portComparisonShards {
		subset = shards[index].asked
	}
	path := casesFile(t, subset)
	want := portGoAnswers(t, oracle, subset)
	gitAnswers := map[string][]string{}
	for _, tr := range subset.Trees {
		if tr.Git == "run" {
			gitAnswers[tr.Name] = gitCheckIgnore(t, tr)
		}
	}
	dir := product
	node := onNode(t, filepath.Join(dir, "main.ts"), path)
	nat := portNative(t, dir, path)
	if index < portComparisonShards {
		backend := onNode(t, filepath.Join(dir, "program.mjs"), path)
		for _, side := range []struct {
			name   string
			result run
		}{{"Node", node}, {"native", nat}, {"JavaScript backend", backend}} {
			if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
				t.Fatalf("%s: exit %d, stderr %q", side.name, side.result.exitCode, side.result.stderr)
			}
			if diff := firstDifference(string(side.result.stdout), want); diff != "" {
				t.Errorf("%s and Go cohere differ: %s", side.name, diff)
			}
		}
		var leak run
		switch runtime.GOOS {
		case "linux":
			leak = execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(dir, "port"), path)
		case "darwin":
			leak = execute(t, nil, "leaks", "--atExit", "--", filepath.Join(dir, "port-leaks"), path)
		default:
			t.Fatalf("no leak check for %s", runtime.GOOS)
		}
		if leak.exitCode != 0 {
			t.Errorf("leaks: exit %d: %s %s", leak.exitCode, leak.stdout, leak.stderr)
		}
		a := againstGit(subset, gitAnswers, string(node.stdout))
		for _, diff := range a.differences {
			t.Error(diff)
		}
		portShared.Lock()
		portShared.totals.compared += a.compared
		portShared.totals.ignored += a.ignored
		portShared.totals.reincluded += a.reincluded
		portShared.comparedShards++
		if portShared.comparedShards == portComparisonShards && (portShared.totals.compared == 0 || portShared.totals.ignored == 0 || portShared.totals.reincluded == 0) {
			t.Errorf("git was asked too little: %+v", portShared.totals)
		}
		portShared.Unlock()
		t.Logf("%d live case ids", len(shards[index].ids))
	} else {
		m := mutants[index-portComparisonShards]
		for _, side := range []struct {
			name   string
			result run
		}{{"natively", nat}, {"on Node", node}} {
			if side.result.exitCode != 0 {
				t.Errorf("%s mutant exits %d: %s", side.name, side.result.exitCode, side.result.stderr)
				continue
			}
			diff := firstDifference(string(side.result.stdout), want)
			if diff == "" {
				t.Errorf("%s mutant %q survives", side.name, m.name)
				continue
			}
			t.Logf("%s mutant %q caught: %s", side.name, m.name, diff)
			if side.name == "on Node" && m.seenByGit {
				a := againstGit(subset, gitAnswers, string(side.result.stdout))
				if len(a.differences) == 0 {
					t.Error("mutant agrees with git")
				}
			}
		}
	}
}

// A disagreement planted in one live case is caught by exactly its owning shard.
func TestPortAnswerShardPlantedDisagreement(t *testing.T) {
	deadline := portHardDeadline(t, "TestPortAnswerShardPlantedDisagreement", 75*time.Second)
	defer deadline.Stop()
	asked := cases{Patterns: []patterns{{Name: "planted", Queries: []query{{Path: "one"}}}}}
	shards := portCaseShards(t, asked)
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
		t.Fatalf("planted disagreement caught %d times in shard %d", caught, owner)
	}
	t.Logf("planted disagreement caught only by shard-%03d", owner)
}

func TestThePortAnswersAsGoCohereAndGitDo_000(t *testing.T) { runPortAnswerShard(t, 0) }

func TestThePortAnswersAsGoCohereAndGitDo_001(t *testing.T) { runPortAnswerShard(t, 1) }

func TestThePortAnswersAsGoCohereAndGitDo_002(t *testing.T) { runPortAnswerShard(t, 2) }

func TestThePortAnswersAsGoCohereAndGitDo_003(t *testing.T) { runPortAnswerShard(t, 3) }

func TestThePortAnswersAsGoCohereAndGitDo_004(t *testing.T) { runPortAnswerShard(t, 4) }

func TestThePortAnswersAsGoCohereAndGitDo_005(t *testing.T) { runPortAnswerShard(t, 5) }

func TestThePortAnswersAsGoCohereAndGitDo_006(t *testing.T) { runPortAnswerShard(t, 6) }

func TestThePortAnswersAsGoCohereAndGitDo_007(t *testing.T) { runPortAnswerShard(t, 7) }

func TestThePortAnswersAsGoCohereAndGitDo_008(t *testing.T) { runPortAnswerShard(t, 8) }

func TestThePortAnswersAsGoCohereAndGitDo_009(t *testing.T) { runPortAnswerShard(t, 9) }

func TestThePortAnswersAsGoCohereAndGitDo_010(t *testing.T) { runPortAnswerShard(t, 10) }

func TestThePortAnswersAsGoCohereAndGitDo_011(t *testing.T) { runPortAnswerShard(t, 11) }

func TestThePortAnswersAsGoCohereAndGitDo_012(t *testing.T) { runPortAnswerShard(t, 12) }

func TestThePortAnswersAsGoCohereAndGitDo_013(t *testing.T) { runPortAnswerShard(t, 13) }

func TestThePortAnswersAsGoCohereAndGitDo_014(t *testing.T) { runPortAnswerShard(t, 14) }

func TestThePortAnswersAsGoCohereAndGitDo_015(t *testing.T) { runPortAnswerShard(t, 15) }

func TestThePortAnswersAsGoCohereAndGitDo_016(t *testing.T) { runPortAnswerShard(t, 16) }

func TestThePortAnswersAsGoCohereAndGitDo_017(t *testing.T) { runPortAnswerShard(t, 17) }

func TestThePortAnswersAsGoCohereAndGitDo_018(t *testing.T) { runPortAnswerShard(t, 18) }

func TestThePortAnswersAsGoCohereAndGitDo_019(t *testing.T) { runPortAnswerShard(t, 19) }

func TestThePortAnswersAsGoCohereAndGitDo_020(t *testing.T) { runPortAnswerShard(t, 20) }

func TestThePortAnswersAsGoCohereAndGitDo_021(t *testing.T) { runPortAnswerShard(t, 21) }

func TestThePortAnswersAsGoCohereAndGitDo_022(t *testing.T) { runPortAnswerShard(t, 22) }

func TestThePortAnswersAsGoCohereAndGitDo_023(t *testing.T) { runPortAnswerShard(t, 23) }

func TestThePortAnswersAsGoCohereAndGitDo_024(t *testing.T) { runPortAnswerShard(t, 24) }

func TestThePortAnswersAsGoCohereAndGitDo_025(t *testing.T) { runPortAnswerShard(t, 25) }

func TestThePortAnswersAsGoCohereAndGitDo_026(t *testing.T) { runPortAnswerShard(t, 26) }

func TestPortAnswerShardAssignmentSurvivesCorpusGrowth(t *testing.T) {
	deadline := portHardDeadline(t, "TestPortAnswerShardAssignmentSurvivesCorpusGrowth", 75*time.Second)
	defer deadline.Stop()
	before := cases{Trees: []tree{{Name: "real repository", Queries: []query{{Path: "a.go"}, {Path: "directory", IsDirectory: true}}}}}
	old := portCaseShards(t, before)
	after := before
	after.Trees = append([]tree(nil), before.Trees...)
	after.Trees[0].Queries = append([]query{{Path: "new.go"}}, before.Trees[0].Queries...)
	grown := portCaseShards(t, after)
	owners := map[string]int{}
	for i, s := range grown {
		for _, id := range s.ids {
			owners[id] = i
		}
	}
	for i, s := range old {
		for _, id := range s.ids {
			if owners[id] != i {
				t.Fatalf("adding a repository file moved %q from shard-%03d to shard-%03d", id, i, owners[id])
			}
		}
	}
}
