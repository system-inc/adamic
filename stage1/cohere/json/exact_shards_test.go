package json

import "testing"

func TestPortMatchesGoCohereUnion(t *testing.T) {
	t.Parallel()
	allCases, _ := jsonTopCorpus(t)
	cases, shards, dedicated := jsonPortClassOrderPartition(allCases)
	if len(shards) != testPortMatchesGoCohereShards {
		t.Fatal("static port count differs from enumeration")
	}
	if e := jsonPortClassOrderUnion(allCases, cases, shards, dedicated); e != nil {
		t.Fatal(e)
	}
	// Main used 8192 / 4 buckets. Check every actual owner against that
	// historical partition, including empty buckets, so no case can move.
	const mainAgreementBuckets = 2048
	if len(shards) != mainAgreementBuckets {
		t.Fatal("agreement bucket count changed from main")
	}
	for bucket, shard := range shards {
		for _, item := range cases[shard.start:shard.end] {
			if want := jsonCaseShard(item.Name, mainAgreementBuckets); bucket != want {
				t.Fatalf("%s moved from main bucket %d to %d", item.Name, want, bucket)
			}
		}
	}
	t.Logf("exact live union: %d ordinary cases exactly once across %d shards + %d dedicated; %d total", len(cases), len(shards), len(dedicated), len(allCases))
}
