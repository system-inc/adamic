package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type statementSharedProducts struct {
	Cases, Want, Specs, Port, LoweredDir  string
	Binary, Release, Library, LibraryHash string
	Elapsed                               float64
	Shards                                *statementGateShards
}

var statementCaseContexts sync.Map

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
	shards := statementGateShardsOnce(t)
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

// The partition as every statement test of one process shares it. Each shard's
// protocol and specifications are held as bytes: the files statementShards wrote
// belong to the test that built the partition and are removed when that test ends.
type statementGateShards struct {
	cases, want, specs    string
	shards                []statementShard
	texts, specifications [][]byte
	builder               string
}

var statementGateShardsShared struct {
	once   sync.Once
	shards *statementGateShards
}

// The first statement test in a process partitions the corpus through
// statementShards, with every whole-corpus check it makes: each pinned upstream
// record present, every stable key unique and the union exact. The partition is
// a pure function of the corpus product, so later tests in the process read it
// rather than repeat it: each shard and each of the 32 library oracle products
// re-partitioned all 58,497 cases and rewrote all 16 shard files, 1.0 to 1.9 s
// apiece. TestStatementsAgainstGoAndPrettierUnion still builds and checks its own.
func statementGateShardsOnce(t *testing.T) *statementGateShards {
	t.Helper()
	// printerCorpusFiles skips without a TypeScript source. Skip here, before the
	// once, so every test skips alike rather than failing on the builder's skip.
	if os.Getenv("ADAMIC_TYPESCRIPT_SOURCE") == "" {
		printerCorpusFiles(t)
	}
	shared := &statementGateShardsShared
	shared.once.Do(func() {
		cases, want, specs := statementCorpus(t)
		gate := &statementGateShards{cases: cases, want: want, specs: specs, builder: t.Name()}
		for _, shard := range statementShards(t, cases, want, specs) {
			text, err := os.ReadFile(shard.text)
			if err != nil {
				t.Fatal(err)
			}
			specifications, err := os.ReadFile(shard.specs)
			if err != nil {
				t.Fatal(err)
			}
			gate.shards = append(gate.shards, statementShard{indices: shard.indices, labels: shard.labels, want: shard.want})
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
func (gate *statementGateShards) shard(t *testing.T, number int) statementShard {
	t.Helper()
	directory := t.TempDir()
	shard := gate.shards[number]
	shard.text = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.txt", number))
	shard.specs = filepath.Join(directory, fmt.Sprintf("statement-shard-%03d.json", number))
	if err := os.WriteFile(shard.text, gate.texts[number], 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shard.specs, gate.specifications[number], 0644); err != nil {
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
func statementLibraryHashOnce(t *testing.T) string {
	t.Helper()
	shared := &statementLibraryHashShared
	shared.once.Do(func() { shared.hash = statementDirectoryHash(t, os.Getenv("ADAMIC_TS_PRETTIER")) })
	if shared.hash == "" {
		t.Fatal("npm Prettier hash failed")
	}
	return shared.hash
}
