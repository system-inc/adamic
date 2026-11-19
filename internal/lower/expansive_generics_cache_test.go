package lower

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// expansiveGenericsCachePrograms are the count test's programs: one whose generic holds itself plainly, so the walk
// meets an edge and gives no stop, and one whose generic grows, so the walk's stop is a refusal a hit must give in
// the same words (#bmqm0k1).
var expansiveGenericsCachePrograms = []struct {
	name    string
	source  string
	refused bool
}{
	{"no stop", "interface Box<T> { readonly next: Box<T> | undefined; readonly v: T; }\nconst o: Box<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n", false},
	{"refused", "interface Own<T> { readonly next: Own<Own<T>> | undefined; readonly v: T; }\nconst o: Own<number> = { next: undefined, v: 1 };\nconsole.log(String(o.v));\n", true},
}

// expansiveGenericsLowering is a lowering as the latent census's fresh makes one, sharing cache (#bmqm0k1).
func expansiveGenericsLowering(program *load.Program, check *checker.Checker, cache *expansiveGenericsCache) *lowering {
	return &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1, expansiveGenericsHeld: cache}
}

// Two lowerings sharing a cache walk a program once, and each gets the stop a lowering with no cache walks to, the
// refusal's words included; three lowerings with no cache, as production's, walk three times. That walk was what
// #bmqm0k1 took off: each of the latent census's units walked every module of the program again. Mutants
// (expansive-generics-cache.mutants.md): the cache ignored, in expansiveGenerics `if l.expansiveGenericsHeld == nil {`
// to `if true {`, and the two lowerings sharing a cache count two walks; the cache never storing, in
// expansiveGenerics
// `l.expansiveGenericsHeld.answers[key] = expansiveGenericsAnswer{modules: slices.Clone(modules), stop: stop}`
// to `_ = key`, the same.
//
// Not parallel: expansiveGenericsWalks counts every walk in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestExpansiveGenericsIsWalkedOncePerCheckerAndProgram(t *testing.T) {
	var walks atomic.Int64
	expansiveGenericsWalks.Store(&walks)
	defer expansiveGenericsWalks.Store(nil)
	for _, row := range expansiveGenericsCachePrograms {
		program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", row.source))
		walked := expansiveGenericsLowering(program, check, nil).expansiveGenerics(modules)
		var refused *Refused
		if row.refused && (!errors.As(walked, &refused) || !strings.Contains(refused.What, expansiveGenericsWords)) {
			t.Fatalf("%s: the fixture's walk gave %v, not a refusal saying %q", row.name, walked, expansiveGenericsWords)
		}
		if !row.refused && walked != nil {
			t.Fatalf("%s: the fixture's walk gave %v, not no stop", row.name, walked)
		}
		walks.Store(0)
		cache := &expansiveGenericsCache{}
		for unit := range 2 {
			shared := expansiveGenericsLowering(program, check, cache).expansiveGenerics(modules)
			if !expansiveGenericsSame(shared, walked) {
				t.Errorf("%s: lowering %d through the cache gave %s, not its own walk's %s", row.name, unit+1, expansiveGenericsText(shared), expansiveGenericsText(walked))
			}
		}
		if count := walks.Load(); count != 1 {
			t.Errorf("%s: two lowerings sharing a cache walked %d times, not once", row.name, count)
		}
		walks.Store(0)
		for range 3 {
			expansiveGenericsLowering(program, check, nil).expansiveGenerics(modules)
		}
		if count := walks.Load(); count != 3 {
			t.Errorf("%s: three lowerings with no cache walked %d times, not once each", row.name, count)
		}
	}
}

// With expansiveGenericsCompare on, a lowering that takes the cache's stop walks again on a scratch copy and panics
// where the two differ: a planted refusal panics, and the stop a first lowering stored doesn't. Mutant: the
// comparison skipped, in expansiveGenerics `if expansiveGenericsCompare.Load() {` to `if false {`, and the planted
// refusal is taken without a panic.
//
// Not parallel: expansiveGenericsCompare changes every lowering with a cache in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestExpansiveGenericsCompareCatchesAWrongAnswer(t *testing.T) {
	expansiveGenericsCompare.Store(true)
	defer expansiveGenericsCompare.Store(false)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", expansiveGenericsCachePrograms[0].source))
	cache := &expansiveGenericsCache{}
	expansiveGenericsLowering(program, check, cache).expansiveGenerics(modules)
	// The stored stop is a fresh walk's, so taking it under the comparison doesn't panic (#bmqm0k1).
	expansiveGenericsLowering(program, check, cache).expansiveGenerics(modules)
	key := expansiveGenericsKey{checker: check, program: program}
	answer, found := cache.answers[key]
	if !found {
		t.Fatal("the first lowering through the cache stored no stop for its checker and program")
	}
	answer.stop = &Refused{Where: "main.ts:1:1", What: "a planted stop", Fix: "none"}
	cache.answers[key] = answer
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "#bmqm0k1") {
			t.Fatalf("a planted stop taken under the comparison gave %v, not its panic", failure)
		}
	}()
	expansiveGenericsLowering(program, check, cache).expansiveGenerics(modules)
}
