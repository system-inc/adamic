package lower

import (
	"fmt"
	"maps"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// inheritanceSource extends a class that extends another, so its walk notes two extended classes, and a cached
// answer has something to hand out (#hjspb3b).
const inheritanceSource = `class Base { readonly value: number = 1; }
class Middle extends Base {}
class Leaf extends Middle {}
console.log(String(new Leaf().value));
`

// inheritanceLowering is a lowering as the latent census's fresh makes one, sharing cache (#hjspb3b).
func inheritanceLowering(program *load.Program, check *checker.Checker, cache *inheritanceCache) *lowering {
	return &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1, inheritance: cache}
}

// Two lowerings sharing a cache walk a program once, and each holds what a lowering with no cache walks to; three
// lowerings with no cache, as production's, walk three times. That walk was what #hjspb3b took off: each of the
// latent census's units walked every module of the program again. Mutant: the cache never stores, in
// noteInheritance
// `l.inheritance.answers[key] = inheritanceAnswer{modules: slices.Clone(modules), derivedAncestors: maps.Clone(l.derivedAncestors)}`
// to `_ = key`, and the two lowerings sharing a cache count two walks.
//
// Not parallel: inheritanceWalks counts every walk in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestInheritanceIsWalkedOncePerCheckerAndProgram(t *testing.T) {
	var walks atomic.Int64
	inheritanceWalks.Store(&walks)
	defer inheritanceWalks.Store(nil)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", inheritanceSource))
	walked := inheritanceLowering(program, check, nil)
	walked.noteInheritance(modules)
	if names := inheritanceOnly(walked.derivedAncestors, nil); names != "[Base Middle]" {
		t.Fatalf("the fixture's walk noted %s as extended, not [Base Middle]", names)
	}
	walks.Store(0)
	cache := &inheritanceCache{}
	for unit := range 2 {
		shared := inheritanceLowering(program, check, cache)
		shared.noteInheritance(modules)
		if !maps.Equal(shared.derivedAncestors, walked.derivedAncestors) {
			t.Errorf("lowering %d through the cache noted %s as extended, not its own walk's %s", unit+1, inheritanceOnly(shared.derivedAncestors, nil), inheritanceOnly(walked.derivedAncestors, nil))
		}
	}
	if count := walks.Load(); count != 1 {
		t.Errorf("two lowerings sharing a cache walked %d times, not once", count)
	}
	walks.Store(0)
	for range 3 {
		inheritanceLowering(program, check, nil).noteInheritance(modules)
	}
	if count := walks.Load(); count != 3 {
		t.Errorf("three lowerings with no cache walked %d times, not once each", count)
	}
}

// With inheritanceCompare on, a lowering that takes the cache's answer walks again on a scratch copy and panics
// where the two differ: a planted empty answer panics, and the answer a first lowering stored doesn't. Mutant: the
// comparison skipped, in noteInheritance `if inheritanceCompare.Load() {` to `if false {`, and the planted answer
// is taken without a panic.
//
// Not parallel: inheritanceCompare changes every lowering with a cache in the package while it's set, and no
// parallel test runs beside a sequential one.
func TestInheritanceCompareCatchesAWrongAnswer(t *testing.T) {
	inheritanceCompare.Store(true)
	defer inheritanceCompare.Store(false)
	program, check, modules := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", inheritanceSource))
	cache := &inheritanceCache{}
	inheritanceLowering(program, check, cache).noteInheritance(modules)
	// The stored answer is a fresh walk's, so taking it under the comparison doesn't panic (#hjspb3b).
	inheritanceLowering(program, check, cache).noteInheritance(modules)
	key := inheritanceKey{checker: check, program: program}
	answer := cache.answers[key]
	answer.derivedAncestors = map[*ast.Symbol]bool{}
	cache.answers[key] = answer
	defer func() {
		failure := recover()
		if failure == nil || !strings.Contains(fmt.Sprint(failure), "#hjspb3b") {
			t.Fatalf("a planted answer taken under the comparison gave %v, not its panic", failure)
		}
	}()
	inheritanceLowering(program, check, cache).noteInheritance(modules)
}
