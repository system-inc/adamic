package json

import (
	"fmt"
	"strings"
	"testing"
)

func jsonUpstreamBlockCheck(shard, id, got, want string) error {
	if got != want {
		return fmt.Errorf("%s case %s: upstream difference identity or answer changed; update the checked-in report", shard, id)
	}
	return nil
}

func TestJSONUpstreamShardDisagreement(t *testing.T) {
	t.Parallel()
	cases := make([]textCase, 33)
	for i := range cases {
		cases[i] = textCase{Name: fmt.Sprintf("case-%d.json", i), Text: "{}"}
	}
	shards := jsonHashShards(cases, testUpstreamRepositoryCorpusParityShards)
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	caught := 0
	for ordinal, shard := range shards {
		name := fmt.Sprintf("shard-%04d", ordinal)
		for i := shard.start; i < shard.end; i++ {
			got := ""
			if cases[i].Name == "case-17.json" {
				got = "planted disagreement"
			}
			if err := jsonUpstreamBlockCheck(name, cases[i].Name, got, ""); err != nil {
				caught++
				if ordinal != jsonCaseShard("case-17.json", testUpstreamRepositoryCorpusParityShards) || !strings.Contains(err.Error(), name) {
					t.Fatalf("wrong owner: %v", err)
				}
				t.Logf("caught planted disagreement only in %s: %v", name, err)
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards", caught)
	}
}
