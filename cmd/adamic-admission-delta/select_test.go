package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func TestShardPartitionKeepsDiff(t *testing.T) {
	t.Parallel()
	corpora := []corpus{{Name: "diff", Programs: []program{{Path: "changed.a"}}}, {Name: "fixtures"}, {Name: "fuzz", Programs: []program{}}}
	for i := 0; i < 30; i++ {
		corpora[1].Programs = append(corpora[1].Programs, program{Path: fmt.Sprintf("input-%d.a", i)})
	}
	seen := map[string]int{}
	for i := 1; i <= 3; i++ {
		selected, complete, err := selectCorpora(corpora, "", fmt.Sprintf("%d/3", i))
		if err != nil || complete {
			t.Fatal("shard claims complete coverage", err)
		}
		if len(selected[0].Programs) != 1 || selected[0].Programs[0].Path != "changed.a" {
			t.Fatal("diff was split or lost")
		}
		if len(selected) != 3 || len(selected[2].Programs) != 0 {
			t.Fatal("empty corpus was lost")
		}
		for _, p := range selected[1].Programs {
			seen[p.Path]++
		}
	}
	if len(seen) != 30 {
		t.Fatal("shards missed programs", seen)
	}
	for path, count := range seen {
		if count != 1 {
			t.Fatal("overlapping shards", path, count)
		}
	}
	selected, complete, err := selectCorpora(corpora, "diff", "")
	if err != nil || complete || len(selected) != 1 || len(selected[0].Programs) != 1 {
		t.Fatal("diff-only filter", selected, complete, err)
	}
	for _, shard := range []string{"0/3", "4/3", "1/0", "1", "1/3junk"} {
		if _, _, err := selectCorpora(corpora, "", shard); err == nil {
			t.Fatal("bad shard accepted", shard)
		}
	}
	if _, _, err := selectCorpora(corpora, "missing", ""); err == nil {
		t.Fatal("unknown corpus accepted")
	}
}

func TestShardCommandScope(t *testing.T) {
	t.Parallel()
	dir, sha, _, compiler := commandFixture(t)
	for _, selector := range [][]string{{"--corpus-filter", "diff"}, {"--shard", "1/3"}} {
		args := []string{"--base", sha, "--head", sha, "--base-binary", compiler, "--head-binary", compiler, "--manifest", filepath.Join(dir, "manifest.json"), "--json"}
		o := invokeCommand(t, dir, append(args, selector...)...)
		var r report
		if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
			t.Fatal(err, o.Stderr)
		}
		if o.Exit != 0 || r.Verdict != "partial" || r.CompleteCorpus || r.Admitted != 0 || r.DiffCount != 0 {
			t.Fatalf("partial run looks clean: %+v %s", r, o.Stderr)
		}
		if _, exists := r.Phases["checkout"]; !exists {
			t.Fatal("checkout phase missing")
		}
		if _, exists := r.Phases["provision"]; !exists {
			t.Fatal("provision phase missing")
		}
	}
}
