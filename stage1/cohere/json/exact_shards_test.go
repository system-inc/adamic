package json

import "testing"

func TestPortMatchesGoCohereUnion(test *testing.T) {
	test.Parallel()
	allCases, _ := jsonTopCorpus(test)
	cases, shards, dedicated := jsonPortDedicatedPartition(allCases)
	if len(shards) != testPortMatchesGoCohereShards {
		test.Fatal("static port count differs from enumeration")
	}
	if jsonPortDedicatedUnionError := jsonPortDedicatedUnion(
		allCases,
		cases,
		shards,
		dedicated,
	); jsonPortDedicatedUnionError != nil {
		test.Fatal(jsonPortDedicatedUnionError)
	}
	// A dedicated file that leaves the pinned corpus would leave its owners with nothing to check.
	if len(dedicated) != len(jsonPortDedicatedKeys) {
		test.Fatalf("the corpus holds %d of the %d dedicated cases", len(dedicated), len(jsonPortDedicatedKeys))
	}
	// Main used 8192 / 4 buckets. Check every actual owner against that
	// historical partition, including empty buckets, so no case can move.
	const mainAgreementBuckets = 2048
	if len(shards) != mainAgreementBuckets {
		test.Fatal("agreement bucket count changed from main")
	}
	for bucket, shard := range shards {
		for _, item := range cases[shard.start:shard.end] {
			if want := jsonCaseShard(item.Name, mainAgreementBuckets); bucket != want {
				test.Fatalf("%s moved from main bucket %d to %d", item.Name, want, bucket)
			}
		}
	}
	test.Logf(
		"exact live union: %d ordinary cases exactly once across %d shards + %d dedicated; %d total",
		len(cases),
		len(shards),
		len(dedicated),
		len(allCases),
	)
}
