package lower

import (
	"fmt"
	"maps"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// accessorNamesSource has a getter in a class and one in a literal, so its walk notes two names, and a cached
// answer has something to hand out (#hjspb3b).
const accessorNamesSource = `class Box { readonly held: number = 1; get value(): number { return this.held; } }
const sized = { get size(): number { return 2; } };
console.log(new Box().value + sized.size);
`

// accessorNamesLowering is a lowering as the latent census's fresh makes one, sharing cache (#hjspb3b).
func accessorNamesLowering(program *load.Program, check *checker.Checker, cache *accessorNamesCache) *lowering {
	return &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1, accessorNamesHeld: cache}
}

// Two lowerings sharing a cache walk a program once, the second under no checker at all, since the key is the
// program alone, and each holds what a lowering with no cache walks to; three lowerings with no cache, as
// production's, walk three times. That walk was what #hjspb3b took off: each of the latent census's units walked
// every module of the program again. Mutant: the cache never stores, in noteAccessorNames
// `l.accessorNamesHeld.answers[key] = accessorNamesAnswer{modules: slices.Clone(modules), accessorNames: maps.Clone(l.accessorNames)}`
// to `_ = key`, and the two lowerings sharing a cache count two walks.
//
// Not parallel: accessorNamesWalks counts every walk in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestAccessorNamesAreWalkedOncePerProgram(t *testing.T) {
	var walks atomic.Int64
	accessorNamesWalks.Store(&walks)
	defer accessorNamesWalks.Store(nil)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", accessorNamesSource))
	walked := accessorNamesLowering(program, check, nil)
	walked.noteAccessorNames(modules)
	if names := accessorNamesOnly(walked.accessorNames, nil); names != "[size value]" {
		t.Fatalf("the fixture's walk noted %s as getter names, not [size value]", names)
	}
	walks.Store(0)
	cache := &accessorNamesCache{}
	// The second lowering has no checker, as a lowering of another file has another: the walk asks none, so the
	// answer is the program's whatever checker a lowering holds (#hjspb3b).
	for unit, unitChecker := range []*checker.Checker{check, nil} {
		shared := accessorNamesLowering(program, unitChecker, cache)
		shared.noteAccessorNames(modules)
		if !maps.Equal(shared.accessorNames, walked.accessorNames) {
			t.Errorf("lowering %d through the cache noted %s as getter names, not its own walk's %s", unit+1, accessorNamesOnly(shared.accessorNames, nil), accessorNamesOnly(walked.accessorNames, nil))
		}
	}
	if count := walks.Load(); count != 1 {
		t.Errorf("two lowerings sharing a cache walked %d times, not once", count)
	}
	walks.Store(0)
	for range 3 {
		accessorNamesLowering(program, check, nil).noteAccessorNames(modules)
	}
	if count := walks.Load(); count != 3 {
		t.Errorf("three lowerings with no cache walked %d times, not once each", count)
	}
}

// With accessorNamesCompare on, a lowering that takes the cache's answer walks again on a scratch copy and panics
// where the two differ: a planted empty answer panics, and the answer a first lowering stored doesn't. Mutant: the
// comparison skipped, in noteAccessorNames `if accessorNamesCompare.Load() {` to `if false {`, and the planted
// answer is taken without a panic.
//
// Not parallel: accessorNamesCompare changes every lowering with a cache in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestAccessorNamesCompareCatchesAWrongAnswer(t *testing.T) {
	accessorNamesCompare.Store(true)
	defer accessorNamesCompare.Store(false)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", accessorNamesSource))
	cache := &accessorNamesCache{}
	accessorNamesLowering(program, check, cache).noteAccessorNames(modules)
	// The stored answer is a fresh walk's, so taking it under the comparison doesn't panic (#hjspb3b).
	accessorNamesLowering(program, check, cache).noteAccessorNames(modules)
	key := accessorNamesKey{program: program}
	answer := cache.answers[key]
	answer.accessorNames = map[string]bool{}
	cache.answers[key] = answer
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "#hjspb3b") {
			t.Fatalf("a planted answer taken under the comparison gave %v, not its panic", failure)
		}
	}()
	accessorNamesLowering(program, check, cache).noteAccessorNames(modules)
}
