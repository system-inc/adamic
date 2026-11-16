package lower

import (
	"context"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// placeholderSlotsSource holds a placeholder local, a copy of it, and a placeholder field whose class is seen
// through an interface, so its walk has transfers and a view to follow, and its PlaceholderSources is never empty.
const placeholderSlotsSource = `class Slot {value:string=undefined!;}
interface Shape {value:string}
let first:string=undefined!;
const saved=first;
const source=new Slot();
const viewed:Shape=source;
first='written';
console.log(saved+viewed.value);
`

// placeholderSlotsProgram loads placeholderSlotsSource with its entry's checker.
func placeholderSlotsProgram(t *testing.T) (*load.Program, *checker.Checker, *ast.SourceFile) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.a", placeholderSlotsSource)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	return loaded, check, entry
}

// placeholderSlotsLowering is a lowering as the latent census's fresh makes one, sharing cache.
func placeholderSlotsLowering(program *load.Program, check *checker.Checker, cache *placeholderSlotCache) *lowering {
	return &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1, placeholderSlots: cache}
}

// Lowerings sharing a cache walk a program once for each checker and program, however many there are, which
// is the work #6xt6p1f took off: each of the latent census's units walked every module of the program again.
// A second program, with its own checker, walks once more. Mutant: the cache never consulted, its guard in
// notePlaceholderSlots from `if l.placeholderSlots == nil ||` to `if true || l.placeholderSlots == nil ||`,
// and this counts seven walks.
//
// Not parallel: placeholderSlotWalks counts every walk in the package while it's set, and no parallel test
// runs beside a sequential one.
func TestPlaceholderSlotsAreWalkedOncePerCheckerAndProgram(t *testing.T) {
	var walks atomic.Int64
	placeholderSlotWalks.Store(&walks)
	defer placeholderSlotWalks.Store(nil)
	cache := &placeholderSlotCache{}
	first, firstChecker, _ := placeholderSlotsProgram(t)
	for range 5 {
		lowering := placeholderSlotsLowering(first, firstChecker, cache)
		lowering.notePlaceholderSlots()
		if len(lowering.result.PlaceholderSources) == 0 {
			t.Fatal("the fixture's placeholders gave no PlaceholderSources")
		}
	}
	second, secondChecker, _ := placeholderSlotsProgram(t)
	for range 2 {
		placeholderSlotsLowering(second, secondChecker, cache).notePlaceholderSlots()
	}
	if walked := walks.Load(); walked != 2 {
		t.Fatalf("seven lowerings over two programs, each with its own checker, walked %d times, not once each", walked)
	}
	if len(cache.answers) != 2 {
		t.Fatalf("the cache holds %d answers, not one for each checker and program", len(cache.answers))
	}
}

// A lowering's maps are its own: a write into one lowering's PlaceholderSources or PlaceholderViews shows in
// no later lowering's and not in the cache, whether the lowering filled the cache or took from it. Mutants:
// the cached maps held by reference where they're stored, `sources: maps.Clone(l.result.PlaceholderSources)`
// to `sources: l.result.PlaceholderSources`, so the first's write shows in the cache and the second's maps;
// or where they're taken, `maps.Clone(answer.sources)` to `answer.sources`, so the second's write shows in
// the cache and the third's maps.
func TestPlaceholderSlotsAreClonedInAndOut(t *testing.T) {
	t.Parallel()
	program, check, _ := placeholderSlotsProgram(t)
	cache := &placeholderSlotCache{}
	key := placeholderSlotKey{checker: check, program: program}
	first := placeholderSlotsLowering(program, check, cache)
	first.notePlaceholderSlots()
	if _, found := cache.answers[key]; !found {
		t.Fatal("the first lowering didn't fill the cache")
	}
	first.result.PlaceholderSources["written by the first"] = "first"
	first.result.PlaceholderViews["written by the first"] = "first"
	second := placeholderSlotsLowering(program, check, cache)
	second.notePlaceholderSlots()
	second.result.PlaceholderSources["written by the second"] = "second"
	second.result.PlaceholderViews["written by the second"] = "second"
	third := placeholderSlotsLowering(program, check, cache)
	third.notePlaceholderSlots()
	lacks := func(holder string, sources, views ir.SortedMap[string, string], writes ...string) {
		for _, write := range writes {
			if _, found := sources[write]; found {
				t.Errorf("PlaceholderSources of %s holds what was %s", holder, write)
			}
			if _, found := views[write]; found {
				t.Errorf("PlaceholderViews of %s holds what was %s", holder, write)
			}
		}
	}
	lacks("the cache", cache.answers[key].sources, cache.answers[key].views, "written by the first", "written by the second")
	lacks("the second lowering", second.result.PlaceholderSources, second.result.PlaceholderViews, "written by the first")
	lacks("the third lowering", third.result.PlaceholderSources, third.result.PlaceholderViews, "written by the first", "written by the second")
}

// A lowering under a type mapper, or with a slot kept (keepSlot), can have an answer of its own, so it walks,
// and neither takes the cache's answer nor fills the cache. Nothing calls notePlaceholderSlots under either
// today, so this pins the guard alone. Mutant: the guard dropped, in notePlaceholderSlots
// `if l.placeholderSlots == nil || l.typeMapper != nil || len(l.slots.explicit) != 0 {` to
// `if l.placeholderSlots == nil {`, and both cases take the planted answer and fill the empty cache.
//
// Not parallel: placeholderSlotWalks counts every walk in the package while it's set, and no parallel test
// runs beside a sequential one.
func TestPlaceholderSlotsUnderAMapperOrAKeptSlotLeaveTheCacheAlone(t *testing.T) {
	var walks atomic.Int64
	placeholderSlotWalks.Store(&walks)
	defer placeholderSlotWalks.Store(nil)
	program, check, entry := placeholderSlotsProgram(t)
	key := placeholderSlotKey{checker: check, program: program}
	// let first:string=undefined!, whose type maps to itself and whose initializer is kept in a slot of it:
	// what the mapper or the slot says is no matter here, only that one is in force.
	declaration := entry.Statements.Nodes[2].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]
	held := check.GetTypeAtLocation(declaration.Name())
	guards := []struct {
		name string
		set  func(*lowering)
	}{
		{"a type mapper", func(lowering *lowering) {
			lowering.typeMapper = newTypeMapper([]*checker.Type{held}, []*checker.Type{held})
		}},
		{"a kept slot", func(lowering *lowering) { lowering.keepSlot(declaration.Initializer(), held) }},
	}
	for _, guard := range guards {
		planted := &placeholderSlotCache{answers: map[placeholderSlotKey]placeholderSlotAnswer{key: {sources: ir.SortedMap[string, string]{"planted": "planted"}, views: ir.SortedMap[string, string]{}}}}
		walks.Store(0)
		reading := placeholderSlotsLowering(program, check, planted)
		guard.set(reading)
		reading.notePlaceholderSlots()
		if walked := walks.Load(); walked != 1 {
			t.Errorf("a lowering under %s walked %d times, not once", guard.name, walked)
		}
		if _, found := reading.result.PlaceholderSources["planted"]; found {
			t.Errorf("a lowering under %s took the cache's answer", guard.name)
		}
		empty := &placeholderSlotCache{}
		filling := placeholderSlotsLowering(program, check, empty)
		guard.set(filling)
		filling.notePlaceholderSlots()
		if len(empty.answers) != 0 {
			t.Errorf("a lowering under %s filled the cache", guard.name)
		}
	}
}

// placeholderSlotsShardCount is how many shards the cached-answer comparison divides its programs into. The one
// test ran 12.97 s on a quiet box and 61.6 s on a loaded line box against the 60 s unit budget (#58r9613), so
// four put each shard near 15 s loaded, the count the view walk's comparison took for the same shape (#sq4myx1).
const placeholderSlotsShardCount = 4

// placeholderSlotsPins are the pins' own programs, by the name their shard is chosen by. Their shard writes each
// to a temporary directory under its base name, and pins/ keeps them apart from the corpus's own main.a files.
var placeholderSlotsPins = map[string]string{"pins/main.a": placeholderSlotsSource}

// placeholderSlotsPrograms names every program the cached-answer comparison lowers: the pins', then every .a and
// .ts program of internal/oracle's corpus, the one the view walk's comparison lowers, by its slash-separated path
// under internal/oracle/testdata.
func placeholderSlotsPrograms() ([]string, error) {
	names := slices.Sorted(maps.Keys(placeholderSlotsPins))
	// Reads: internal/oracle/testdata/*
	err := filepath.WalkDir("../oracle/testdata", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// A .d.ts is no program, only declarations another file names.
		if !entry.IsDir() && (filepath.Ext(path) == ".a" || filepath.Ext(path) == ".ts" && !strings.HasSuffix(path, ".d.ts")) {
			names = append(names, strings.TrimPrefix(filepath.ToSlash(path), "../oracle/testdata/"))
		}
		return nil
	})
	return names, err
}

// placeholderSlotsShardPrograms is the programs of names in shard, those nameShard puts in it, as it divides the
// cycle finder's inputs and the view walk's.
func placeholderSlotsShardPrograms(names []string, shard uint64) []string {
	var selected []string
	for _, name := range names {
		if nameShard(name, placeholderSlotsShardCount) == shard {
			selected = append(selected, name)
		}
	}
	return selected
}

// With placeholderSlotCompare on, a lowering that takes the cache's answer walks the program again into a
// scratch result and panics where the two differ. For every program of one shard, this fills a cache from a
// lowering over a checker nothing has asked yet, lowers the program whole through that checker, as the census's
// units ask it between one unit's refusal pass and the next, and then takes the cached answer in a third
// lowering, whose fresh walk runs over the checker the whole lowering left. A program lowered once never takes
// from its cache, so this is how each one hits. TestPlaceholderSlotsCachedAgreeWithAWalk_0 to _3 in
// placeholder_slot_shards_test.go each call it with one shard; each logs its own "N of M programs took a cached
// answer", and TestPlaceholderSlotsShardsCoverEveryProgram holds the shards to every program exactly once.
// Mutant: the answer stored empty, `sources: maps.Clone(l.result.PlaceholderSources)` to
// `sources: ir.SortedMap[string, string]{}`, and every program with a placeholder panics, in every shard.
//
// Not parallel: placeholderSlotCompare changes every lowering with a cache in the package while it's set, and
// no parallel test runs beside a sequential one. Its subtests are parallel with one another, and the switch
// goes off only once they're done.
func runPlaceholderSlotsShard(t *testing.T, shard uint64) {
	t.Helper()
	placeholderSlotCompare.Store(true)
	t.Cleanup(func() { placeholderSlotCompare.Store(false) })
	names, err := placeholderSlotsPrograms()
	if err != nil {
		t.Fatal(err)
	}
	programs := placeholderSlotsShardPrograms(names, shard)
	var compared atomic.Int64
	t.Run("programs", func(t *testing.T) {
		for _, name := range programs {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				path := filepath.Join("../oracle/testdata", filepath.FromSlash(name))
				if source, pinned := placeholderSlotsPins[name]; pinned {
					path = writeAgreementSource(t, filepath.Base(name), source)
				}
				loaded, err := load.Load([]string{path})
				if err != nil {
					return // A program that doesn't load is never lowered.
				}
				defer func() {
					if failure := recover(); failure != nil {
						t.Fatalf("%s: %v", path, failure)
					}
				}()
				entry := loaded.Files()[0]
				check, release := loaded.Checker(context.Background(), entry)
				defer release()
				cache := &placeholderSlotCache{}
				placeholderSlotsLowering(loaded, check, cache).notePlaceholderSlots()
				// Its error is no matter here: whatever it asked of the checker before it stopped is asked.
				_, _ = lowerChecked(loaded, check, entry, nil, false, false)
				hit := placeholderSlotsLowering(loaded, check, cache)
				hit.notePlaceholderSlots()
				if len(hit.result.PlaceholderSources) != 0 {
					compared.Add(1)
				}
			})
		}
	})
	if compared.Load() == 0 {
		t.Fatalf("no program of shard %d took a cached answer with a placeholder in it, so nothing was compared", shard)
	}
	t.Logf("%d of %d programs took a cached answer with placeholders in shard %d of %d, each compared with a fresh walk", compared.Load(), len(programs), shard, placeholderSlotsShardCount)
}
