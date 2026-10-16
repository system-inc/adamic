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

// go list -test's look for its cached test main at the package's directory joined to the Go cache's absolute path is
// go's, never the tree's, wherever the cache sits (#pqq9z5q): two builders with their caches at two places class it
// alike, as nothing. A missed look at a tree path that only resembles one, or that holds the cache's path as part of a
// name, is still a tree path, and so is anything a look found. Mutant: the joined look classed as a tree path (each
// builder's own exists entry, naming its cache).
// Not parallel: it sets GOCACHE, which settleFacts asks go for.
func TestGoLookingForItsCachedTestMainIsNoTreeRead(t *testing.T) {
	root := filepath.Join(t.TempDir(), "adamic")
	var classes [][]classed
	for _, cache := range []string{"/home/ahra/.cache/go-build", "/home/ahra/loom-sums/s1/gocache"} {
		t.Setenv("GOCACHE", cache)
		facts, err := settleFacts(root, filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		testMain := "/28/28079cc4cf0088b656dc5dcec5353d6596c309cc5a5d1e4eea57b8c285139d88-d"
		classes = append(classes, []classed{
			facts.classifyPath(filepath.Join(root, "cmd/adamic", cache+testMain), readAbsent),
			facts.classifyPath(filepath.Join(root, "bridge/tsgo", cache+testMain), readAbsent),
		})
		if found := facts.classifyPath(filepath.Join(root, "cmd/adamic", cache+"-not", testMain), readAbsent); found.kind != "exists" {
			t.Errorf("a missed look at a tree path beside the cache's: %+v", found)
		}
		if found := facts.classifyPath(filepath.Join(root, "cmd/adamic", cache+testMain), readStat); found.kind == "" {
			t.Errorf("a look that found a tree path holding the cache's: %+v", found)
		}
	}
	if !reflect.DeepEqual(classes[0], classes[1]) || !reflect.DeepEqual(classes[0], []classed{{}, {}}) {
		t.Fatalf("go's look for its cached test main, under two caches: %q and %q, want nothing", classes[0], classes[1])
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
	before, err := treeEntryValue(root, cache, entry)
	if err != nil || before != "absent" {
		t.Fatalf("nothing beside the cache: %q, %v", before, err)
	}
	if err := os.Mkdir(filepath.Join(filepath.Dir(cache), "beside-the-cache"), 0o755); err != nil {
		t.Fatal(err)
	}
	if after, err := treeEntryValue(root, cache, entry); err != nil || after != "directory" {
		t.Fatalf("a directory beside the cache: %q, %v", after, err)
	}
}

// A tree at Workshop's checkout and the same tree at a runner's unit source, deeper, class a lookup that climbs to home
// or past it alike, by where it lands (Builder's dry B2, Oct 12: ../../../../lib was /lib on Workshop and a path under
// ~/loom-serve on a runner, so 29 of 39 product packages moved): the C stack as the machine's system, a path at the top
// as absolute, a file in home as home's. The tree's own parents and what sits just above it stay relative, the same
// at any depth. Mutant: a landed lookup keyed relative to the tree (the two depths class it apart).
// Not parallel: it sets ADAMIC_TRACE_REPORT.
func TestALookupThatClimbsToHomeOrPastItIsKeyedWhereItLands(t *testing.T) {
	// A home no machine has: on a box whose home is /home/ahra, ~/.adamic-tools is a real link to the tools directory,
	// and classifyPath rightly resolves it there, so the test failed on Workshop every run while it passed on a Mac.
	home := "/home/adamic-no-such-home"
	classes := func(root string) []classed {
		t.Helper()
		cache := filepath.Join(home, "loom-builder", "trees", "e329", "cache")
		t.Setenv("ADAMIC_TRACE_REPORT", filepath.Join(home, "loom-builder", "trees", "e329", "logs", "product-1.trace.jsonl"))
		facts, err := settleFacts(root, cache, filepath.Join(cache, "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		facts.temporary, facts.home = nil, home
		var found []classed
		for _, name := range []string{"/lib", "/usr", "/include", "/package.json", filepath.Join(home, ".adamic-tools"),
			filepath.Join(filepath.Dir(root), "package.json"), filepath.Dir(root)} {
			found = append(found, facts.classifyPath(name, readStat|readAbsent))
		}
		return found
	}
	workshop, runner := classes(filepath.Join(home, "loom-trees", "adamic")), classes(filepath.Join(home, "loom-serve", "units", "u1", "source"))
	// A machine directory only looked at is keyed by whether it exists (#vdha67a), so /lib, a directory on Linux and
	// absent on a Mac, is classed by what this machine holds; either way the two depths agree.
	machine := func(name string) classed {
		if isDirectory(name) {
			return classed{"system directory", name}
		}
		return classed{"system", name}
	}
	want := []classed{machine("/lib"), machine("/usr"), {"absolute", "/include"}, {"absolute", "/package.json"}, {"home", "~/.adamic-tools"},
		{"above", "../package.json"}, {"above", ".."}}
	if !reflect.DeepEqual(workshop, want) || !reflect.DeepEqual(runner, want) {
		t.Fatalf("classed\n%q at Workshop's depth and\n%q at a runner's, want\n%q", workshop, runner, want)
	}
}
