package json

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Fixed hash buckets have headroom. Keys are repository-relative case paths,
// including generated indices/modes, never enumeration positions.
const testPortMatchesGoCohereShards = 2048
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
		jsonTopState.shards = jsonHashShards(jsonTopState.cases, testPortMatchesGoCohereShards)
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
	portMatchesPrepare(t)
	v, ready := jsonTopState.products.Load(key)
	if !ready || v.(*jsonTopBuild).dir == "" {
		t.Fatal("shared setup did not prepare product: " + key)
	}
	return v.(*jsonTopBuild).dir
}
func jsonTopOracle(t *testing.T) string {
	t.Helper()
	portMatchesPrepare(t)
	if jsonTopState.oracleDir == "" {
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
	code := m.Run()
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
