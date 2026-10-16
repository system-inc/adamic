package json

import (
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// Fixed hash buckets have headroom. Keys are repository-relative case paths,
// including generated indices/modes, never enumeration positions.
const testPortMatchesGoCohereShards = 2048

// ADAMIC_TEST_SHARD=i/n selects ordinals modulo n locally; unset runs every
// generated top-level leaf. Isolated gate leaves use go test -timeout 90s.
// Products and enumeration are shared once per process, before any comparisons.
// Setup (portMatchesPrepareProducts) writes oracle, the Go oracle's binary, and clangTools once; a leaf reads them
// only after portMatchesPrepare.
var jsonTopState struct {
	corpus     sync.Once
	cases      []textCase
	shards     []nativeChunk
	oracle     string
	clangTools []string
	products   sync.Map
}

type jsonTopBuild struct {
	once sync.Once
	dir  string
}

func jsonTopCorpus(t *testing.T) ([]textCase, []nativeChunk) {
	t.Helper()
	jsonTopState.corpus.Do(func() {
		if tracked, _ := pinnedCorpusProduct(t); len(tracked) == 0 {
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
func jsonTopProduct(t *testing.T, key string, in buildcache.Inputs, build func(string) error) string {
	t.Helper()
	portMatchesPrepare(t)
	v, ready := jsonTopState.products.Load(key)
	if !ready || v.(*jsonTopBuild).dir == "" {
		t.Fatal("shared setup did not prepare product: " + key)
	}
	return v.(*jsonTopBuild).dir
}

// jsonTopOracle is the Go oracle's binary, as shared setup prepared it.
func jsonTopOracle(t *testing.T) string {
	t.Helper()
	portMatchesPrepare(t)
	if jsonTopState.oracle == "" {
		t.Fatal("shared setup did not prepare the Go oracle")
	}
	return jsonTopState.oracle
}
