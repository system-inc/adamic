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

// quietHome sets HOME and XDG_CONFIG_HOME to a new temporary home whose Go telemetry is off, for a test that runs go
// there through settleFacts. A fresh home takes the day's upload token, so go forks its telemetry sidecar detached, and
// the sidecar writes into the home after the test returns, failing TempDir's cleanup (chg_gp66hxpv's red on Adam:
// "unlinkat .../002: directory not empty"). With the mode off, go writes nothing there and forks nothing, which the
// cleanup holds: a go with telemetry on writes its counters there before it exits, so anything beside the mode file
// fails the test on any box, not only a loaded one. Mutant: no mode file.
func quietHome(t *testing.T) string {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	configuration, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	telemetry := filepath.Join(configuration, "go", "telemetry")
	if err := os.MkdirAll(telemetry, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(telemetry, "mode"), []byte("off\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Registered after TempDir's, so it runs before the home is removed.
	t.Cleanup(func() {
		entries, err := os.ReadDir(telemetry)
		if err != nil {
			t.Error(err)
			return
		}
		for _, entry := range entries {
			if entry.Name() != "mode" {
				t.Errorf("go wrote its telemetry into the test's home: %s", filepath.Join(telemetry, entry.Name()))
			}
		}
	})
	return home
}

// A directory on the build cache's own chain of parents, listed or opened to list as Go's os.Open does, is keyed as a
// lookup beside the cache's parents is, relative to the cache, never refused under the home, and two homes holding
// different things class it alike (#4gp2ytz: esbuild's resolver lists every parent up to /). A listing off that chain
// under the home, a sibling of the cache's or of one of its parents, and a file on it opened, are still refused under
// the home (Tests' bounds). Mutant: every directory listed under the home keyed relative to the cache.
// Not parallel: it sets HOME, which settleFacts reads.
func TestAListingOfTheCachesParentsIsKeyedBesideTheCache(t *testing.T) {
	classes := func(extra string) []classed {
		t.Helper()
		home := quietHome(t)
		directory := filepath.Join(home, "loom-builder", "trees", "6299aba1")
		cache := filepath.Join(directory, "cache")
		for _, made := range []string{cache, filepath.Join(directory, "logs"), filepath.Join(home, "elsewhere"), filepath.Join(home, extra)} {
			if err := os.MkdirAll(made, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		os.WriteFile(filepath.Join(home, "package.json"), []byte("{}\n"), 0o644)
		facts, err := settleFacts(filepath.Join(home, "loom-trees", "adamic"), cache, filepath.Join(cache, "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		// A home in no temporary directory, as a builder's isn't.
		facts.temporary = nil
		return []classed{
			facts.classifyPath(directory, readListing),
			facts.classifyPath(filepath.Join(home, "loom-builder"), readContent|readListing),
			facts.classifyPath(home, readContent|readListing),
			facts.classifyPath(filepath.Join(directory, "logs"), readListing),
			facts.classifyPath(filepath.Join(home, "elsewhere"), readContent|readListing),
			facts.classifyPath(filepath.Join(home, "package.json"), readContent),
		}
	}
	first, second := classes("first-box-only"), classes("second-box-only")
	if !reflect.DeepEqual(first[:3], second[:3]) {
		t.Fatalf("two homes holding different things classed the cache's parents apart:\n%q\n%q", first, second)
	}
	want := []classed{{"above cache", ".."}, {"above cache", "../../.."}, {"above cache", "../../../.."}}
	if !reflect.DeepEqual(first[:3], want) {
		t.Fatalf("the cache's parents classed %q, want %q", first[:3], want)
	}
	for index, name := range []string{"a sibling of the cache", "a directory of the home off the chain", "a file of the home"} {
		if found := first[3+index]; found.kind != refusal+"home" {
			t.Errorf("%s classed %q, want refused under the home", name, found)
		}
	}
}

// A lookup below a directory of the cache's chain under the home, deeper than its direct children, is keyed beside the
// cache too, so two caches under one home settle to one key (#4gp2ytz: tsc's @types search missed <the cache's
// parent>/node_modules/@types, named by its whole home path, so two checkouts keyed the adapted tree apart, Compiler's
// Golem harness on 1412e316ed). A listing or a content read there is still the home's, refused, and a path that's there
// keeps its own class (chg_gp66hxpv's red). Mutants: the rule dropped, so the home rule names each cache's path; the
// rule firing on a stat too.
// Not parallel: it sets HOME, which settleFacts reads.
func TestALookupBelowTheCachesChainUnderTheHomeIsKeyedBesideTheCache(t *testing.T) {
	home := quietHome(t)
	classes := func(label string) []classed {
		t.Helper()
		cache := filepath.Join(home, "tests-proof", "hc", label, "cache")
		if err := os.MkdirAll(cache, 0o755); err != nil {
			t.Fatal(err)
		}
		facts, err := settleFacts(filepath.Join(home, "tree"), cache, filepath.Join(cache, "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		facts.temporary = nil
		parent := filepath.Dir(cache)
		return []classed{
			facts.classifyPath(filepath.Join(parent, "node_modules", "@types"), readAbsent),
			facts.classifyPath(filepath.Join(filepath.Dir(parent), "node_modules", "@types", "node"), readAbsent),
			facts.classifyPath(filepath.Join(parent, "node_modules", "@types"), readListing),
			facts.classifyPath(filepath.Join(parent, "001", "usr", "bin", "true"), readStat),
		}
	}
	first, second := classes("at28-after"), classes("at29-after-other-path")
	want := []classed{{"above cache", "../node_modules/@types"}, {"above cache", "../../node_modules/@types/node"}}
	if !reflect.DeepEqual(first[:2], want) || !reflect.DeepEqual(second[:2], want) {
		t.Fatalf("lookups below the cache's chain classed %q and %q, want %q at both", first[:2], second[:2], want)
	}
	if first[2].kind != refusal+"home" {
		t.Errorf("a listing below the cache's chain classed %q, want refused under the home", first[2])
	}
	// A path that's there keeps its own class: a test's temporary directory beside its cache under the home holds
	// machine files, and on Adam's runner the rule took a stat of one ahead of its class (chg_gp66hxpv's red). The leaf
	// isn't a compiler's name, since this test compiles nothing (internal/compilercensus).
	if first[3].kind == "above cache" {
		t.Errorf("an existing path below the cache's chain classed %q, want its own class", first[3])
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

// A read of a path under the tools root, through any spelling that resolves to it, classes as the root's read whatever
// home the run had (#zxq3k08, Node's ruling on shape 4): the wasm products' sysroot header probes missed through the
// ~/.adamic-tools link and keyed `home ~/.adamic-tools/...` under the box user's home and `absolute
// /home/ahra/.adamic-tools/...` under a unit's, so every set key moved with the home (the scoped pair on Golem, Oct 25).
// Here the box's root is a directory with a .adamic-tools link beside it, ADAMIC_TOOLS names the directory, and one
// header probe through the root and through the link, under the box user's home and an empty unit home, is the one
// toolset entry all four ways. With ADAMIC_TOOLS unset, the probe through the link under the home is the home's, as
// before. Mutant that fails it: the spellings not mapped onto the root, red on the link rows.
// Not parallel: points the home directory and ADAMIC_TOOLS at each layout's through t.Setenv.
func TestAReadThroughEveryToolsSpellingIsTheRootsRead(t *testing.T) {
	box := quietHome(t)
	unit, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tools := filepath.Join(box, "adamic-tools")
	if err := os.MkdirAll(filepath.Join(tools, "wasi-sdk", "share", "wasi-sysroot", "include"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("adamic-tools", filepath.Join(box, ".adamic-tools")); err != nil {
		t.Fatal(err)
	}
	header := filepath.Join("wasi-sdk", "share", "wasi-sysroot", "include", "stdatomic.h")
	cache := filepath.Join(t.TempDir(), "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_TOOLS", tools)
	for _, home := range []string{box, unit} {
		t.Setenv("HOME", home)
		facts, err := settleFacts(filepath.Join(t.TempDir(), "adamic"), cache, filepath.Join(cache, "trace-1"))
		if err != nil {
			t.Fatal(err)
		}
		for _, spelling := range []string{tools, filepath.Join(box, ".adamic-tools")} {
			if got := facts.classifyPath(filepath.Join(spelling, header), readAbsent); got != (classed{"toolset", "wasi-sdk"}) {
				t.Errorf("under home %s, the header's probe through %s classed %q, want the root's toolset", home, spelling, got)
			}
		}
	}
	t.Setenv("ADAMIC_TOOLS", "")
	t.Setenv("HOME", box)
	facts, err := settleFacts(filepath.Join(t.TempDir(), "adamic"), cache, filepath.Join(cache, "trace-1"))
	if err != nil {
		t.Fatal(err)
	}
	// A home in no temporary directory, as a builder's isn't: this one is under the test's.
	facts.temporary = nil
	if got := facts.classifyPath(filepath.Join(box, ".adamic-tools", header), readAbsent); got.kind != "home" {
		t.Errorf("with ADAMIC_TOOLS unset, the probe through the link classed %q, want the home's as before", got)
	}
}
