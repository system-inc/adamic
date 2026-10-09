package lint

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"
)

func TestShardsAgree_Union(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("shards_agree_split_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < testShardsAgreeShards; i++ {
		name := fmt.Sprintf("func TestShardsAgree_%03d(", i)
		if bytes.Count(data, []byte(name)) != 1 {
			t.Fatalf("enumeration missing/duplicate %s", name)
		}
	}
	if len(regexp.MustCompile(`(?m)^func TestShardsAgree_[0-9]{3}\(`).FindAll(data, -1)) != testShardsAgreeShards {
		t.Fatal("shard enumeration differs from constant")
	}
	products := testShardsAgreeReady(t)
	rows := products.Rows
	buckets := testShardsAgreePartition(t, rows)
	// Plant one output disagreement and use the same byte oracle as every leaf.
	planted := rows[0]
	caught := []int{}
	for shard, bucket := range buckets {
		for _, row := range bucket {
			want := []byte(row)
			got := append([]byte(nil), want...)
			if row == planted {
				got = append(got, '!')
			}
			if difference(got, want) != "" {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 || caught[0] != testShardsAgreeOwner(t, planted) {
		t.Fatalf("planted disagreement caught by %v", caught)
	}
	t.Logf("union %d cases across %d shards", len(rows), testShardsAgreeShards)
	t.Logf("planted disagreement caught exactly once by TestShardsAgree_%03d", caught[0])
}
