package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func TestShardPartitionKeepsDiff(test *testing.T) {
	test.Parallel()
	corpora := []corpus{
		{Name: "diff", Programs: []program{{Path: "changed.a"}}},
		{Name: "fixtures"},
		{Name: "fuzz", Programs: []program{}},
	}
	for i := 0; i < 30; i++ {
		corpora[1].Programs = append(corpora[1].Programs, program{Path: fmt.Sprintf("input-%d.a", i)})
	}
	seen := map[string]int{}
	for i := 1; i <= 3; i++ {
		selected, complete, selectCorporaError := selectCorpora(corpora, "", fmt.Sprintf("%d/3", i))
		if selectCorporaError != nil || complete {
			test.Fatal("shard claims complete coverage", selectCorporaError)
		}
		if len(selected[0].Programs) != 1 || selected[0].Programs[0].Path != "changed.a" {
			test.Fatal("diff was split or lost")
		}
		if len(selected) != 3 || len(selected[2].Programs) != 0 {
			test.Fatal("empty corpus was lost")
		}
		for _, p := range selected[1].Programs {
			seen[p.Path]++
		}
	}
	if len(seen) != 30 {
		test.Fatal("shards missed programs", seen)
	}
	for path, count := range seen {
		if count != 1 {
			test.Fatal("overlapping shards", path, count)
		}
	}
	selected, complete, selectCorporaError := selectCorpora(corpora, "diff", "")
	if selectCorporaError != nil || complete || len(selected) != 1 || len(selected[0].Programs) != 1 {
		test.Fatal("diff-only filter", selected, complete, selectCorporaError)
	}
	for _, shard := range []string{"0/3", "4/3", "1/0", "1", "1/3junk"} {
		if _, _, selectCorporaError := selectCorpora(corpora, "", shard); selectCorporaError == nil {
			test.Fatal("bad shard accepted", shard)
		}
	}
	if _, _, selectCorporaError := selectCorpora(corpora, "missing", ""); selectCorporaError == nil {
		test.Fatal("unknown corpus accepted")
	}
}

func TestShardCommandScope(test *testing.T) {
	test.Parallel()
	dir, sha, _, compiler := commandFixture(test)
	for _, selector := range [][]string{{"--corpus-filter", "diff"}, {"--shard", "1/3"}} {
		args := []string{
			"--base",
			sha,
			"--head",
			sha,
			"--base-binary",
			compiler,
			"--head-binary",
			compiler,
			"--manifest",
			filepath.Join(dir, "manifest.json"),
			"--json",
		}
		o := invokeCommand(test, dir, append(args, selector...)...)
		var r report
		if decodeError := json.Unmarshal([]byte(o.Stdout), &r); decodeError != nil {
			test.Fatal(decodeError, o.Stderr)
		}
		if o.Exit != 0 || r.Verdict != "partial" || r.CompleteCorpus || r.Admitted != 0 || r.DiffCount != 0 {
			test.Fatalf("partial run looks clean: %+v %s", r, o.Stderr)
		}
		if _, exists := r.Phases["checkout"]; !exists {
			test.Fatal("checkout phase missing")
		}
		if _, exists := r.Phases["provision"]; !exists {
			test.Fatal("provision phase missing")
		}
	}
}
