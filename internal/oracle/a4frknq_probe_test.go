package oracle

import (
	"testing"
)

// #a4frknq's probe, on a proof commit only and never merged: how many count rows each of the 16 product shards holds,
// and of what kind, so a shard's seconds can be read against its size.
func TestPerformanceA4frknqShardSizes(t *testing.T) {
	sizes := make([]int, countProductShards)
	for _, each := range countCases() {
		sizes[familyShard(each.key, countProductShards)]++
	}
	t.Logf("A4 shard sizes %v total %d", sizes, len(countCases()))
}
