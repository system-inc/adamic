package json

import (
	"path/filepath"
	"testing"
	"time"
)

// inventory.json, the corpus's largest case (17.5 MB), shared _0760 with a 3 KB case and was json#20/28's long pole:
// 17.7 s alone on Golem (Go 1.6 s, Node 1.3 s, release 1.9 s, the sanitized run 10.8 s, the JavaScript backend 1.6 s),
// 29.4 s in the unit on Cloud, and 77.6 s in the unit with 8 busy loops halving its CPUs, under the 90 s deadline with
// nothing to spare (#21kptp1). One case doesn't split, so it runs as class order's case does: alone in the unit, before
// go test resumes the parallel shards, with all five sides (jsonPortAgree, the same agreement a top shard checks).

// Not parallel: runs alone in the unit, before the parallel shards resume, so the indivisible case has its CPUs.
func TestPortMatchesGoCohereInventory(t *testing.T) {
	setup := time.Now()
	portMatchesPrepare(t)
	all, _ := jsonTopCorpus(t)
	ordinary, shards, dedicated := jsonPortDedicatedPartition(all)
	if err := jsonPortDedicatedUnion(all, ordinary, shards, dedicated); err != nil {
		t.Fatal(err)
	}
	items := jsonPortDedicatedCase(dedicated, jsonPortInventoryKey)
	if len(items) != 1 {
		t.Fatalf("%s is in the corpus %d times, want 1", jsonPortInventoryKey, len(items))
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("shared setup: %.3fs; exact union: %d ordinary + %d dedicated", time.Since(setup).Seconds(), len(ordinary), len(dedicated))
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	start := time.Now()
	binaries := jsonPortBinaries{
		oracle:    portMatchesShared.oracle,
		entry:     portMatchesShared.entry,
		script:    portMatchesShared.script,
		release:   portMatchesShared.release,
		sanitized: portMatchesShared.sanitized,
		runner:    runner,
	}
	jsonPortAgree(t, items, binaries, "inventory")
	t.Logf("%s: own work %.3fs", jsonPortInventoryKey, time.Since(start).Seconds())
}
