package main

import "testing"

// The live union is checked without a fixed case total. The top-level shard
// table must retain exactly the declared number of gate-visible tests.
// Not parallel: inventoryEngineSharedBuild and inventoryEngineScratch hold the shared engine build.
func TestInventoryEngineUnion(t *testing.T) {
	if len(inventoryEngineTopLevelTests) != testInventoryEngineShards {
		t.Fatalf("enumerated %d top-level shards, want %d", len(inventoryEngineTopLevelTests), testInventoryEngineShards)
	}
	inventoryEngineShared(t)
}
