package lower

import (
	"slices"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// walkExpansiveGenerics' answer, its stop, reads only the checker and the modules it's handed: it reads syntax and
// asks l.checker for symbols alone, never for a type, so no type mapper, substitution or kept slot reaches it, and it
// writes nothing on the lowering at all, so nothing a lowering holds before it changes what it gives. The latent
// census makes a fresh lowering for each of its units, and each walked every module of the program again before it
// lowered: 4,064 walks and 55.3 s a census run, where production walks once a lowering (#bmqm0k1). So a census run
// hands its lowerings one expansiveGenericsCache, as it does an inheritanceCache (inheritance_cache.go), and a
// lowering takes the stop the first lowering over its checker and program walked to. The cache is its own file
// because expansive_generics.go is the rule, and this is the census's (#bmqm0k1).

// expansiveGenericsCache is walkExpansiveGenerics' stop for each checker and program a run's lowerings walked.
// Production makes none and walks; the latent census makes one a run (lower.go.txt's latentRun), which statecopy
// shares. A run lowers one unit at a time, so it takes no lock (#bmqm0k1).
type expansiveGenericsCache struct {
	answers map[expansiveGenericsKey]expansiveGenericsAnswer
}

// expansiveGenericsKey is what the stop is a function of, beside the modules the answer holds. The checker is in it
// because the walk asks it for symbols a hit doesn't, which leaves the checker as the walk would only where that
// same checker answered them once already, as inheritanceKey's does. A run asks the program for a checker per file,
// and the program holds those checkers itself, so the key keeps nothing alive the program doesn't (#bmqm0k1).
type expansiveGenericsKey struct {
	checker *checker.Checker
	program *load.Program
}

// expansiveGenericsAnswer is one walk's stop, nil where no generic grows. Every lowering that takes it is handed the
// same value: the walk's only stop is a *Refused carrying its Where, What and Fix strings and nothing of the
// lowering's, and the census's latentRecord, where the stop goes, reads those and writes none (#bmqm0k1).
type expansiveGenericsAnswer struct {
	// modules is what the walk read. A census run hands every unit the same modules; an ask with others walks
	// (#bmqm0k1).
	modules []*ast.SourceFile
	stop    error
}

// expansiveGenericsWalks, when a test sets it, counts the walks walkExpansiveGenerics makes, so a pin holds a run's
// lowerings to one walk per checker and program. It's nil outside those tests (#bmqm0k1).
var expansiveGenericsWalks atomic.Pointer[atomic.Int64]

// expansiveGenericsCompare, set only by tests, walks the program again beside every answer taken from the cache and
// panics where the two differ (expansiveGenericsCompared). Tests that set it aren't parallel, and no parallel test
// runs beside a sequential one (#bmqm0k1).
var expansiveGenericsCompare atomic.Bool

// expansiveGenerics refuses a generic instantiated without end once a lowering: the cached stop where this lowering
// has a cache, else its own walk (walkExpansiveGenerics). It needs no guard, as collectViewedElements'
// viewedElementsFresh is: the walk reads nothing of the lowering's but the checker and the program the key holds, and
// writes nothing of it, so a hit leaves the lowering as the walk would (#bmqm0k1).
func (l *lowering) expansiveGenerics(modules []*ast.SourceFile) error {
	if l.expansiveGenericsHeld == nil {
		return l.walkExpansiveGenerics(modules)
	}
	key := expansiveGenericsKey{checker: l.checker, program: l.program}
	if answer, found := l.expansiveGenericsHeld.answers[key]; found && slices.Equal(answer.modules, modules) {
		if expansiveGenericsCompare.Load() {
			l.expansiveGenericsCompared(answer)
		}
		return answer.stop
	}
	stop := l.walkExpansiveGenerics(modules)
	if l.expansiveGenericsHeld.answers == nil {
		l.expansiveGenericsHeld.answers = map[expansiveGenericsKey]expansiveGenericsAnswer{}
	}
	// Other modules for the same checker and program replace the held answer, as another entry replaces
	// moduleOrder's (#bmqm0k1).
	l.expansiveGenericsHeld.answers[key] = expansiveGenericsAnswer{modules: slices.Clone(modules), stop: stop}
	return stop
}

// expansiveGenericsCompared walks the program on a scratch copy of this lowering beside an answer taken from the
// cache, and panics where the walk's stop isn't the cached one. The walk writes nothing on the lowering today, so the
// copy only keeps this lowering out of reach of a walk that someday does, as inheritanceCompared's keeps this
// lowering's derivedAncestors (#bmqm0k1).
func (l *lowering) expansiveGenericsCompared(answer expansiveGenericsAnswer) {
	scratch := *l
	fresh := scratch.walkExpansiveGenerics(answer.modules)
	if !expansiveGenericsSame(answer.stop, fresh) {
		panic("expansive generics comparison (#bmqm0k1): the cached stop for this checker and program isn't a fresh walk's; cached " + expansiveGenericsText(answer.stop) + ", fresh " + expansiveGenericsText(fresh))
	}
}

// expansiveGenericsSame says two stops are one: both nil, or both a stop with the same words. The words are the
// comparison because a fresh walk builds a new *Refused, never the cached value (#bmqm0k1).
func expansiveGenericsSame(one, other error) bool {
	if one == nil || other == nil {
		return one == nil && other == nil
	}
	return one.Error() == other.Error()
}

// expansiveGenericsText is a stop as the comparison's panic reads it, "(none)" for nil (#bmqm0k1).
func expansiveGenericsText(stop error) string {
	if stop == nil {
		return "(none)"
	}
	return "[" + stop.Error() + "]"
}
