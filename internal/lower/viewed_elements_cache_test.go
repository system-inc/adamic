package lower

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// viewedElementsGrowingSource views a generic that instantiates itself deeper at each step, so its walk stops
// (TestViewedElementsStopOnAGenericThatDeepensItself's first row), and a cached answer has a stop to hand out.
const viewedElementsGrowingSource = `interface Grow<T> { readonly value?: T; readonly next?: Grow<Grow<T>> }
interface N { readonly kind: number; readonly grow?: Grow<number> }
interface Id extends N { readonly kind: 1; readonly text: string }
function isId(n: N): n is Id { return n.kind === 1; }
const id: Id = { kind: 1, text: "a" };
const n: N = id;
if (isId(n)) { console.log(n.text); }
`

// viewedElementsLowering is a lowering as the latent census's fresh makes one, sharing cache.
func viewedElementsLowering(program *load.Program, check *checker.Checker, cache *viewedElementsCache) *lowering {
	return &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1, viewedElements: cache}
}

// viewedElementsProgram loads the program at path with its entry's checker, and orders its modules on a
// disposable lowering, as the census's latentPrepassModules does, so no lowering the test walks on holds an order.
func viewedElementsProgram(t *testing.T, path string) (*load.Program, *checker.Checker, []*ast.SourceFile) {
	t.Helper()
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	modules, err := viewedElementsLowering(loaded, check, nil).moduleOrder(entry)
	if err != nil {
		t.Fatalf("moduleOrder: %v", err)
	}
	return loaded, check, modules
}

// viewedElementsPrograms are the count test's programs: a cyclic one, whose module order proves module reads,
// and one whose walk stops.
var viewedElementsPrograms = []struct {
	name  string
	path  func(t *testing.T) string
	stops bool
}{
	{"a cyclic program", func(t *testing.T) string {
		absolute, err := filepath.Abs("../oracle/testdata/import_cycles/order/main.a")
		if err != nil {
			t.Fatal(err)
		}
		return absolute
	}, false},
	{"a program whose walk stops", func(t *testing.T) string {
		return writeAgreementSource(t, "main.ts", viewedElementsGrowingSource)
	}, true},
}

// Two fresh lowerings sharing a cache walk a program once, and the second is left as its own walk would leave it:
// the same stop, module order, cyclic flag, proven module reads and entries records as a lowering with no cache.
// That walk was what #hjspb3b took off: each of the latent census's units walked every module of the program
// again. Mutant: the cache never stores, in collectViewedElements
// `l.viewedElements.answers[key] = l.viewedElementsTaken(modules, stop)` to `_ = key`, and each program
// counts two walks.
//
// Not parallel: viewedElementsWalks counts every walk in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestViewedElementsAreWalkedOncePerCheckerAndProgram(t *testing.T) {
	var walks atomic.Int64
	viewedElementsWalks.Store(&walks)
	defer viewedElementsWalks.Store(nil)
	for _, row := range viewedElementsPrograms {
		program, check, modules := viewedElementsProgram(t, row.path(t))
		walked := viewedElementsLowering(program, check, nil)
		walkedStop := walked.collectViewedElements(modules)
		if (walkedStop != nil) != row.stops {
			t.Fatalf("%s: its walk stopped at %v", row.name, walkedStop)
		}
		walks.Store(0)
		cache := &viewedElementsCache{}
		for unit := range 2 {
			shared := viewedElementsLowering(program, check, cache)
			stop := shared.collectViewedElements(modules)
			if differences := viewedElementsDifferences(shared.viewedElementsTaken(modules, stop), walked, walkedStop); len(differences) != 0 {
				t.Errorf("%s: lowering %d through the cache isn't its own walk's: %s", row.name, unit+1, strings.Join(differences, "; "))
			}
		}
		if count := walks.Load(); count != 1 {
			t.Errorf("%s: two lowerings sharing a cache walked %d times, not once", row.name, count)
		}
	}
}

// A lowering whose walk could read or write anything a fresh one doesn't hold walks, and neither takes the
// cache's answer nor fills the cache. Nothing in the census makes one, so this pins the guard alone. Mutant:
// the guard dropped, in collectViewedElements `if l.viewedElements == nil || !l.viewedElementsFresh() {` to
// `if l.viewedElements == nil {`, and every row takes the planted stop and fills the empty cache.
//
// Not parallel: viewedElementsWalks counts every walk in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestViewedElementsOfAnUnfreshLoweringLeaveTheCacheAlone(t *testing.T) {
	var walks atomic.Int64
	viewedElementsWalks.Store(&walks)
	defer viewedElementsWalks.Store(nil)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", viewedElementsGrowingSource))
	key := viewedElementsKey{checker: check, program: program}
	held := check.GetNumberType()
	// What each one holds is no matter here, only that it isn't a fresh lowering's zero.
	guards := []struct {
		name string
		set  func(*lowering)
	}{
		{"a type mapper", func(lowering *lowering) {
			lowering.typeMapper = newTypeMapper([]*checker.Type{held}, []*checker.Type{held})
		}},
		{"a substitution", func(lowering *lowering) { lowering.substitution = map[*checker.Type]ir.Type{held: ir.Number} }},
		{"a kept slot", func(lowering *lowering) { lowering.keepSlot(modules[0].Statements.Nodes[0], held) }},
		{"an intersection being described", func(lowering *lowering) { lowering.intersectionViewDepth = 1 }},
		{"an element key", func(lowering *lowering) { lowering.elementKey = &viewElementKey{} }},
		{"a held module order", func(lowering *lowering) {
			if _, err := lowering.moduleOrder(program.Files()[0]); err != nil {
				t.Fatal(err)
			}
		}},
		{"an entries records memo", func(lowering *lowering) { lowering.entriesRecords = map[*ast.Node]bool{} }},
	}
	for _, guard := range guards {
		planted := &viewedElementsCache{answers: map[viewedElementsKey]viewedElementsAnswer{key: {modules: modules, stop: &NotYet{Where: "planted", What: "planted"}}}}
		walks.Store(0)
		reading := viewedElementsLowering(program, check, planted)
		guard.set(reading)
		stop := reading.collectViewedElements(modules)
		if count := walks.Load(); count != 1 {
			t.Errorf("a lowering with %s walked %d times, not once", guard.name, count)
		}
		if stop != nil && strings.Contains(stop.Error(), "planted") {
			t.Errorf("a lowering with %s took the cache's answer", guard.name)
		}
		empty := &viewedElementsCache{}
		filling := viewedElementsLowering(program, check, empty)
		guard.set(filling)
		_ = filling.collectViewedElements(modules)
		if len(empty.answers) != 0 {
			t.Errorf("a lowering with %s filled the cache", guard.name)
		}
	}
}

// With viewedElementsCompare on, a lowering that takes the cache's answer walks again on a scratch copy and
// panics where the two differ: a planted stop panics, and the answer a first lowering stored doesn't. Mutant: the
// comparison skipped, in collectViewedElements `if viewedElementsCompare.Load() {` to `if false {`, and the
// planted stop is taken without a panic.
//
// Not parallel: viewedElementsCompare changes every lowering with a cache in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestViewedElementsCompareCatchesAWrongAnswer(t *testing.T) {
	viewedElementsCompare.Store(true)
	defer viewedElementsCompare.Store(false)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", viewedElementsGrowingSource))
	cache := &viewedElementsCache{}
	_ = viewedElementsLowering(program, check, cache).collectViewedElements(modules)
	// The stored answer is a fresh walk's, so taking it under the comparison doesn't panic.
	_ = viewedElementsLowering(program, check, cache).collectViewedElements(modules)
	key := viewedElementsKey{checker: check, program: program}
	answer := cache.answers[key]
	answer.stop = &NotYet{Where: "planted", What: "planted"}
	cache.answers[key] = answer
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "#hjspb3b") {
			t.Fatalf("a planted stop taken under the comparison gave %v, not its panic", failure)
		}
	}()
	_ = viewedElementsLowering(program, check, cache).collectViewedElements(modules)
}
