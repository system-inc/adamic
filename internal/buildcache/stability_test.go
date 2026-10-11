package buildcache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Two runs of one tree look up the same things relative to where each run's cache and report sit, and must class
// them alike (#n9jc61n, Planner's skip proof on Home, Oct 11: 12 products re-keyed between two runs, each naming its
// own run's directory). The walk down to the cache is no entry, node's search beside the cache's parents is keyed
// relative to the cache, the trace's report is bookkeeping, and the tree's own parents stay above. Mutants: above before
// the walk to a known place (../trees-2 against ../trees-3); no above cache (each run's absolute node_modules); the
// report not bookkeeping (each run's absolute report).
// Not parallel: it sets ADAMIC_TRACE_REPORT.
func TestTwoRunsOfOneTreeClassWhatTheyLookForAlike(t *testing.T) {
	root := "/work/adamic"
	classes := func(run string) []classed {
		t.Helper()
		directory := filepath.Join("/work", run, "e329")
		cache := filepath.Join(directory, "cache")
		t.Setenv("ADAMIC_TRACE_REPORT", filepath.Join(directory, "logs", "product-388.trace.jsonl"))
		facts, err := settleFacts(root, cache, filepath.Join(cache, "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		// Neither run sits in a temporary directory, as Home's trees don't.
		facts.temporary = nil
		var found []classed
		for _, name := range []string{
			filepath.Join("/work", run), directory, cache,
			filepath.Join(directory, "node_modules"), filepath.Join(directory, "package.json"),
			filepath.Join("/work", run, "node_modules"),
			filepath.Join("/work", "node_modules"),
			filepath.Join(directory, "logs", "product-388.trace.jsonl"),
		} {
			found = append(found, facts.classifyPath(name, readStat|readAbsent))
		}
		return found
	}
	first, second := classes("trees-2"), classes("trees-3")
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("two runs of one tree classed their lookups apart:\n%q\n%q", first, second)
	}
	want := []classed{{}, {}, {}, {"above cache", "../node_modules"}, {"above cache", "../package.json"}, {"above cache", "../../node_modules"}, {"above", "../node_modules"}, {}}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("classed %q, want %q", first, want)
	}
}

// A path above the cache is valued by whether and as what it exists beside the cache at hand, as above values one
// beside the tree. Mutant: valued against the tree.
func TestAPathAboveTheCacheIsValuedBesideIt(t *testing.T) {
	t.Parallel()
	root, cache := filepath.Join(t.TempDir(), "adamic"), filepath.Join(t.TempDir(), "trees-2", "e329", "cache")
	for _, directory := range []string{root, cache, filepath.Join(filepath.Dir(root), "beside-the-cache")} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	entry := readEntry{Kind: "above cache", Path: "../beside-the-cache"}
	before, err := entryValue(root, cache, entry, 0, nil)
	if err != nil || before != "absent" {
		t.Fatalf("nothing beside the cache: %q, %v", before, err)
	}
	if err := os.Mkdir(filepath.Join(filepath.Dir(cache), "beside-the-cache"), 0o755); err != nil {
		t.Fatal(err)
	}
	if after, err := entryValue(root, cache, entry, 0, nil); err != nil || after != "directory" {
		t.Fatalf("a directory beside the cache: %q, %v", after, err)
	}
}
