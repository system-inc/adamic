package lint

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// shardsAgreeHeavyWrapper is a heavy row's test as the file spells it: the name it carries, the pass it runs, and the
// arguments it hands testShardsAgreeHeavyPass, which must say the same.
var shardsAgreeHeavyWrapper = regexp.MustCompile(`(?m)^func TestShardsAgree([A-Z][A-Za-z]*)_([0-9A-Z]+)_(Output|Count)\(t \*testing\.T\) \{\n\tt\.Parallel\(\)\n\ttestShardsAgreeHeavyPass\(t, "([A-Za-z]*)", "([0-9A-Z]*)", (true|false)\)\n\}$`)

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
	// Every heavy row has exactly one test per process count and mode, and each runs the pass its name says.
	type pass struct{ name, label, mode string }
	wrappers := map[pass]int{}
	matches := shardsAgreeHeavyWrapper.FindAllSubmatch(data, -1)
	for _, match := range matches {
		mode := "Output"
		if string(match[6]) == "true" {
			mode = "Count"
		}
		if string(match[1]) != string(match[4]) || string(match[2]) != string(match[5]) || string(match[3]) != mode {
			t.Fatalf("heavy wrapper runs another pass than its name: %s", match[0])
		}
		wrappers[pass{string(match[1]), string(match[2]), string(match[3])}]++
	}
	if len(regexp.MustCompile(`(?m)^func TestShardsAgree[A-Z]`).FindAll(data, -1)) != len(matches) {
		t.Fatal("a heavy wrapper is not spelled as TestShardsAgree_Union reads it")
	}
	want := 0
	for _, entry := range testShardsAgreeHeavy {
		for _, each := range testShardsAgreePasses {
			for _, mode := range []string{"Output", "Count"} {
				want++
				if wrappers[pass{entry.Name, each.Label, mode}] != 1 {
					t.Fatalf("heavy row %s needs exactly one TestShardsAgree%s_%s_%s", entry.Key, entry.Name, each.Label, mode)
				}
			}
		}
	}
	if len(wrappers) != want {
		t.Fatalf("%d heavy wrappers, want %d", len(wrappers), want)
	}
	products := testShardsAgreeReady(t)
	rows := products.Rows
	buckets, heavy := testShardsAgreePartition(t, rows)
	// Every row is in exactly one bucket or heavy entry, never both and never neither, and the product holds that split.
	placed := map[string]int{}
	for _, bucket := range buckets {
		for _, row := range bucket {
			placed[row]++
		}
	}
	for _, entry := range heavy {
		for _, row := range entry {
			placed[row]++
		}
	}
	for _, row := range rows {
		if placed[row] != 1 {
			t.Fatalf("row %q placed %d times", row, placed[row])
		}
	}
	// The product holds the same rows in the same places. Recovery only fills a row's seventh field, so each row's path
	// (its first) must match at every index, not just each bucket's length.
	if len(products.Buckets) != len(buckets) || len(products.Heavy) != len(heavy) {
		t.Fatalf("product holds %d buckets and %d heavy entries, partition %d and %d", len(products.Buckets), len(products.Heavy), len(buckets), len(heavy))
	}
	samePlaces := func(label string, partition, product []string) {
		if len(partition) != len(product) {
			t.Fatalf("%s: product holds %d rows, partition %d", label, len(product), len(partition))
		}
		for index := range partition {
			if path, held := strings.SplitN(partition[index], "\t", 2)[0], strings.SplitN(product[index], "\t", 2)[0]; path != held {
				t.Fatalf("%s row %d: product holds %q, partition %q", label, index, held, path)
			}
		}
	}
	for shard := range buckets {
		samePlaces(fmt.Sprintf("bucket %03d", shard), buckets[shard], products.Buckets[shard])
	}
	for index := range heavy {
		samePlaces("heavy "+testShardsAgreeHeavy[index].Name, heavy[index], products.Heavy[index])
	}
	// Plant one disagreement at a time, in one row's answer at one pass, and judge it with the same byte oracle as
	// every leaf: a bucket's test compares its rows at every pass, a heavy test its entry's rows at its own pass. Each
	// plant must be caught by exactly the test that owns that row at that pass.
	type owner struct {
		test   string
		rows   []string
		passes []pass
	}
	var owners []owner
	var everyPass []pass
	for _, each := range testShardsAgreePasses {
		for _, mode := range []string{"Output", "Count"} {
			everyPass = append(everyPass, pass{"", each.Label, mode})
		}
	}
	for shard, bucket := range buckets {
		owners = append(owners, owner{fmt.Sprintf("TestShardsAgree_%03d", shard), bucket, everyPass})
	}
	for index, entry := range testShardsAgreeHeavy {
		for _, each := range everyPass {
			owners = append(owners, owner{fmt.Sprintf("TestShardsAgree%s_%s_%s", entry.Name, each.label, each.mode), heavy[index], []pass{each}})
		}
	}
	plant := func(planted string, at pass, expected string) {
		var caught []string
		for _, candidate := range owners {
			for _, each := range candidate.passes {
				for _, row := range candidate.rows {
					want := []byte(row)
					got := append([]byte(nil), want...)
					if row == planted && each == at {
						got = append(got, '!')
					}
					if difference(got, want) != "" {
						caught = append(caught, candidate.test)
					}
				}
			}
		}
		if len(caught) != 1 || caught[0] != expected {
			t.Fatalf("disagreement planted in %q at %s %s caught by %v, want %s", planted, at.label, at.mode, caught, expected)
		}
	}
	planted := ""
	for _, row := range rows {
		if testShardsAgreeHeavyOwner(row) < 0 {
			planted = row
			break
		}
	}
	for _, each := range everyPass {
		plant(planted, each, fmt.Sprintf("TestShardsAgree_%03d", testShardsAgreeOwner(t, planted)))
	}
	for index, entry := range testShardsAgreeHeavy {
		for _, each := range everyPass {
			plant(heavy[index][0], each, fmt.Sprintf("TestShardsAgree%s_%s_%s", entry.Name, each.label, each.mode))
		}
	}
	t.Logf("union %d cases across %d shards and %d heavy rows", len(rows), testShardsAgreeShards, len(testShardsAgreeHeavy))
	t.Logf("planted disagreements caught exactly once: at every pass by TestShardsAgree_%03d, and by each heavy pass's own test", testShardsAgreeOwner(t, planted))
}
