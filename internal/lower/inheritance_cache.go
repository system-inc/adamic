package lower

import (
	"maps"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// walkInheritance's answer, derivedAncestors, reads only the checker and the modules it's handed: it asks
// l.checker's GetTypeAtLocation and GetBaseTypes of each class declaration's name, never through concrete, so no
// type mapper, substitution or kept slot reaches it, and it replaces derivedAncestors whole, so nothing a lowering
// holds before it changes what it leaves. The latent census makes a fresh lowering for each of its units, and each
// walked every module of the program again before it lowered: 42 s of the tsc entry census's CPU, where production
// walks once a lowering (#hjspb3b). So a census run hands its lowerings one inheritanceCache, as it does a
// viewedElementsCache (viewed_elements_cache.go), and a lowering takes a copy of the answer the first lowering over
// its checker and program walked. The cache is its own file because class_inheritance.go is the classes'
// inheritance, and this is the census's (#hjspb3b).

// inheritanceCache is walkInheritance's answer for each checker and program a run's lowerings walked. Production
// makes none and walks; the latent census makes one a run (lower.go.txt's latentRun), which statecopy shares. A
// run lowers one unit at a time, so it takes no lock (#hjspb3b).
type inheritanceCache struct {
	answers map[inheritanceKey]inheritanceAnswer
}

// inheritanceKey is what the answer is a function of, beside the modules the answer holds. The checker is in it
// because the walk asks it questions a hit doesn't, which leaves the checker as the walk would only where that
// same checker answered them once already. A run asks the program for a checker per file, and the program holds
// those checkers itself, so the key keeps nothing alive the program doesn't (#hjspb3b).
type inheritanceKey struct {
	checker *checker.Checker
	program *load.Program
}

// inheritanceAnswer is one walk's derivedAncestors, which no lowering holds: each takes a clone of it (#hjspb3b).
type inheritanceAnswer struct {
	// modules is what the walk read. A census run hands every unit the same modules; an ask with others walks
	// (#hjspb3b).
	modules          []*ast.SourceFile
	derivedAncestors map[*ast.Symbol]bool
}

// inheritanceWalks, when a test sets it, counts the walks walkInheritance makes, so a pin holds a run's lowerings
// to one walk per checker and program. It's nil outside those tests (#hjspb3b).
var inheritanceWalks atomic.Pointer[atomic.Int64]

// inheritanceCompare, set only by tests, walks the program again beside every answer taken from the cache and
// panics where the two differ (inheritanceCompared). Tests that set it aren't parallel, and no parallel test runs
// beside a sequential one (#hjspb3b).
var inheritanceCompare atomic.Bool

// noteInheritance fills derivedAncestors once a lowering: a copy of the cached answer where this lowering has a
// cache, else its own walk (walkInheritance). It needs no guard, as collectViewedElements' viewedElementsFresh
// is: the walk reads nothing of the lowering's but the checker and the program the key holds, and writes
// derivedAncestors alone, whole, which a hit replaces as the walk would (#hjspb3b).
func (l *lowering) noteInheritance(modules []*ast.SourceFile) {
	if l.inheritance == nil {
		l.walkInheritance(modules)
		return
	}
	key := inheritanceKey{checker: l.checker, program: l.program}
	if answer, found := l.inheritance.answers[key]; found && slices.Equal(answer.modules, modules) {
		if inheritanceCompare.Load() {
			l.inheritanceCompared(answer)
		}
		// A clone, since a lowering holds its own map and every later lowering's copy comes from this one
		// (#hjspb3b).
		l.derivedAncestors = maps.Clone(answer.derivedAncestors)
		return
	}
	l.walkInheritance(modules)
	if l.inheritance.answers == nil {
		l.inheritance.answers = map[inheritanceKey]inheritanceAnswer{}
	}
	// Other modules for the same checker and program replace the held answer, as another entry replaces
	// moduleOrder's (#hjspb3b).
	l.inheritance.answers[key] = inheritanceAnswer{modules: slices.Clone(modules), derivedAncestors: maps.Clone(l.derivedAncestors)}
}

// inheritanceCompared walks the program on a scratch copy of this lowering beside an answer taken from the cache,
// and panics where the walk's derivedAncestors isn't the cached one. The walk gives the copy a map of its own, so
// this lowering goes on from the cached answer, as it would without the comparison (#hjspb3b).
func (l *lowering) inheritanceCompared(answer inheritanceAnswer) {
	scratch := *l
	scratch.walkInheritance(answer.modules)
	if !maps.Equal(answer.derivedAncestors, scratch.derivedAncestors) {
		panic("inheritance comparison (#hjspb3b): the cached answer for this checker and program isn't a fresh walk's; derivedAncestors: only cached " + inheritanceOnly(answer.derivedAncestors, scratch.derivedAncestors) + ", only fresh " + inheritanceOnly(scratch.derivedAncestors, answer.derivedAncestors))
	}
}

// inheritanceOnly is the name of every symbol one map holds that the other lacks or holds with another value,
// sorted, so a difference reads as the classes it's about (#hjspb3b).
func inheritanceOnly(one, other map[*ast.Symbol]bool) string {
	names := []string{}
	for symbol, extended := range one {
		if held, found := other[symbol]; found && held == extended {
			continue
		}
		if symbol == nil {
			names = append(names, "(no symbol)")
			continue
		}
		names = append(names, symbol.Name)
	}
	slices.Sort(names)
	return "[" + strings.Join(names, " ") + "]"
}
