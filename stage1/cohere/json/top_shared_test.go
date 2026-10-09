package json

import (
	"flag"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// Fixed hash buckets have headroom. Keys are repository-relative case paths,
// including generated indices/modes, never enumeration positions.
const testPortMatchesGoCohereShards = 8192
const testUpstreamRepositoryCorpusParityShards = 2048

// ADAMIC_TEST_SHARD=i/n selects ordinals modulo n locally; unset runs every
// generated top-level leaf. Isolated gate leaves use go test -timeout 90s.
// Products and enumeration are shared once per process, before any comparisons.
var jsonTopState struct {
	corpus              sync.Once
	cases               []textCase
	shards              []nativeChunk
	oracle              sync.Once
	oracleDir           string
	oracleCached        bool
	goOnce, clangOnce   sync.Once
	goTools, clangTools []string
	goErr, clangErr     error
	products            sync.Map
	reports             sync.Mutex
	report              map[int]string
}

type jsonTopBuild struct {
	once sync.Once
	dir  string
}

func jsonTopCorpus(t *testing.T) ([]textCase, []nativeChunk) {
	t.Helper()
	jsonTopState.corpus.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		state, err := repositoryJSONAtHEAD(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(state.paths) == 0 {
			t.Fatal("repository JSON corpus is empty")
		}
		jsonTopState.cases = corpusCases(t)
		jsonTopState.shards = jsonHashShards(jsonTopState.cases, testPortMatchesGoCohereShards/4)
	})
	if len(jsonTopState.cases) == 0 {
		t.Fatal("live JSON corpus is empty or preparation failed")
	}
	return jsonTopState.cases, jsonTopState.shards
}
func jsonTopGoToolchain() ([]string, error) {
	jsonTopState.goOnce.Do(func() { jsonTopState.goTools, jsonTopState.goErr = jsonGoToolchain() })
	return jsonTopState.goTools, jsonTopState.goErr
}
func jsonTopClangToolchain() ([]string, error) {
	jsonTopState.clangOnce.Do(func() { jsonTopState.clangTools, jsonTopState.clangErr = jsonClangToolchain() })
	return jsonTopState.clangTools, jsonTopState.clangErr
}
func jsonTopProduct(t *testing.T, key string, in buildcache.Inputs, build func(string) error) string {
	t.Helper()
	v, ready := jsonTopState.products.Load(key)
	if !ready || v.(*jsonTopBuild).dir == "" {
		t.Fatal("shared setup did not prepare product: " + key)
	}
	return v.(*jsonTopBuild).dir
}
func jsonTopOracle(t *testing.T) string {
	t.Helper()
	if !portMatchesShared.ready || jsonTopState.oracleDir == "" {
		t.Fatal("shared setup did not prepare the Go oracle")
	}
	return jsonTopState.oracleDir
}

func jsonTopReport(t *testing.T, i int, value string) {
	if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" {
		if e := os.WriteFile(fmt.Sprintf("%s.%04d", path, i), []byte(value), 0644); e != nil {
			t.Error(e)
		}
		jsonTopState.reports.Lock()
		defer jsonTopState.reports.Unlock()
		if jsonTopState.report == nil {
			jsonTopState.report = map[int]string{}
		}
		jsonTopState.report[i] = value
	}
}

// The local overlay must outlive all parallel leaves. Shared Product directories
// are read-only and never removed by the test process.
// Not parallel: owns process-wide oracle cleanup and report aggregation after m.Run.
func TestMain(m *testing.M) {
	portMatchesSelectSetup()
	code := upstreamParityTestRuns(m)
	if jsonTopState.oracleDir != "" && !jsonTopState.oracleCached {
		os.RemoveAll(jsonTopState.oracleDir)
	}
	if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" && len(jsonTopState.report) > 0 {
		var keys []int
		for k := range jsonTopState.report {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		var s strings.Builder
		for _, k := range keys {
			s.WriteString(jsonTopState.report[k])
		}
		if e := os.WriteFile(path, []byte(s.String()), 0644); e != nil {
			fmt.Fprintln(os.Stderr, e)
			code = 1
		}
	}
	os.Exit(code)
}
func TestPortMatchesGoCohereUnion(t *testing.T) {
	t.Parallel()
	cases, shards := jsonTopCorpus(t)
	if len(shards)*4 != testPortMatchesGoCohereShards {
		t.Fatal("static port count differs from enumeration")
	}
	if e := jsonPortUnion(cases, shards); e != nil {
		t.Fatal(e)
	}
	t.Logf("exact live union: %d cases", len(cases))
}
func TestUpstreamRepositoryCorpusParityUnion(t *testing.T) {
	t.Parallel()
	cases, shards := jsonTopCorpus(t)
	if len(shards) != testUpstreamRepositoryCorpusParityShards {
		t.Fatal("static upstream count differs from enumeration")
	}
	if e := jsonPortUnion(cases, shards); e != nil {
		t.Fatal(e)
	}
	t.Logf("exact live union: %d cases", len(cases))
}

// Go runs serial tests before resuming parallel leaves. Extend only the process
// fallback timeout: setup and each leaf enforce independent 90-second budgets.
func portMatchesSelectSetup() {
	flag.Parse()
	selector := flag.Lookup("test.run").Value.String()
	parts := strings.Split(selector, "/")
	pattern, err := regexp.Compile(parts[0])
	if err != nil {
		return
	} // Let the testing package report invalid selectors.
	selected := false
	for _, family := range []struct {
		name, format string
		count        int
	}{
		{"TestPortMatchesGoCohere", "%03d", testPortMatchesGoCohereSplitShards},
		{"TestPortMatchesGoCohere", "%04d", testPortMatchesGoCohereShards},
		{"TestUpstreamRepositoryCorpusParity", "%04d", testUpstreamRepositoryCorpusParityShards},
	} {
		if family.name == "TestUpstreamRepositoryCorpusParity" && os.Getenv("ADAMIC_JSON_PRETTIER") == "" {
			continue
		}
		for n := 0; n < family.count && !selected; n++ {
			selected = pattern.MatchString(family.name + "_" + fmt.Sprintf(family.format, n))
		}
	}

	if !selected {
		return
	}
	parts[0] = "(?:" + parts[0] + ")|^TestPortMatchesGoCohereSplit_Setup$"
	if err := flag.Set("test.run", strings.Join(parts, "/")); err != nil {
		panic(err)
	}
	timeout, err := time.ParseDuration(flag.Lookup("test.timeout").Value.String())
	if err == nil && timeout > 0 {
		if err := flag.Set("test.timeout", (timeout + 90*time.Second).String()); err != nil {
			panic(err)
		}
	}
}
