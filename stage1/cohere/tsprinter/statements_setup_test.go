package tsprinter

import (
	"os"
	"sync"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

type statementSharedProducts struct {
	Cases, Want, Specs, Port, LoweredDir  string
	Binary, Release, Library, LibraryHash string
	Elapsed                               float64
	Shards                                *oracles.StatementGateShards
}

var statementSharedProductsShared struct {
	once     sync.Once
	products *statementSharedProducts
}

// Each selected shard fetches its own inputs. No top-level setup test or
// TestMain subprocess is required, and only case work starts its deadline.
// The first shard in a process prepares them and later shards read what it
// made, where each shard repeated all 1.1 to 2.2 s of it.
func statementReadyShared(t *testing.T) statementSharedProducts {
	t.Helper()
	shards := oracles.StatementGateShardsOnce(t)
	// Skip before the once, so every shard skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TS_PRETTIER") == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to an npm install of prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	shared := &statementSharedProductsShared
	shared.once.Do(func() {
		products := statementPrepareCommon(t, shards)
		shared.products = &products
	})
	if shared.products == nil {
		t.Fatal("statement shared preparation failed")
	}
	return *shared.products
}
