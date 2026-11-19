package lower

import (
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// walkViewedElements' answer, its stop and what it leaves on the lowering, reads only the checker, the program
// and the modules it's handed, when the lowering it walks on is fresh (viewedElementsFresh). The latent census
// makes a fresh lowering for each of its units, and each walked every module of the program again before it
// lowered: 284 s of the tsc entry census's 1,413 s of CPU, where production walks once a lowering (#hjspb3b).
// So a census run hands its lowerings one viewedElementsCache, as it does a placeholderSlotCache (placeholder.go,
// #6xt6p1f), and a fresh lowering takes a copy of the answer the first fresh lowering over its checker and
// program walked. The cache is its own file because view_arrays.go is the views' arrays, and this is the
// census's.

// viewedElementsCache is walkViewedElements' answer for each checker and program a run's fresh lowerings walked.
// Production makes none and walks; the latent census makes one a run (lower.go.txt's latentRun), which statecopy
// shares. A run lowers one unit at a time, so it takes no lock.
type viewedElementsCache struct {
	answers map[viewedElementsKey]viewedElementsAnswer
}

// viewedElementsKey is what the answer is a function of, beside the modules the answer holds. A run asks the
// program for a checker per file, and the program holds those checkers itself, so the key keeps nothing alive
// the program doesn't.
type viewedElementsKey struct {
	checker *checker.Checker
	program *load.Program
}

// viewedElementsAnswer is one walk's stop and every lowering field it writes, as they stood right after it, which
// no lowering holds: each takes copies of them. The key the walk filled isn't here, since a fresh lowering has
// none and the walk filled a throwaway one, and neither is objectIntersections, which objectIntersection empties
// of each entry before it returns.
type viewedElementsAnswer struct {
	// modules is what the walk read. A census run hands every unit the same modules; an ask with others walks.
	modules []*ast.SourceFile
	stop    error
	// moduleOrdered, moduleOrders, cyclicModules and provenModuleReads are moduleOrder's, which the walk's record
	// storage questions ask of the program's first file (records_entries.go), and which lowering reads afterwards
	// (cyclicModules in locals.go and class_static.go), so a lowering that takes the answer takes them too.
	moduleOrdered     *moduleOrdering
	moduleOrders      int
	cyclicModules     bool
	provenModuleReads map[*ast.Node]bool
	// entriesRecords is entriesRecordLiteral's memo, which the same questions fill.
	entriesRecords map[*ast.Node]bool
}

// viewedElementsWalks, when a test sets it, counts the walks walkViewedElements makes, so a pin holds a run's
// lowerings to one walk per checker and program. It's nil outside those tests.
var viewedElementsWalks atomic.Pointer[atomic.Int64]

// viewedElementsCompare, set only by tests, walks the program again beside every answer taken from the cache and
// panics where the two differ (viewedElementsCompared). Tests that set it aren't parallel, and no parallel test
// runs beside a sequential one.
var viewedElementsCompare atomic.Bool

// collectViewedElements runs the view key's pre-pass once a lowering: a copy of the cached answer where this
// lowering may take one, else its own walk (walkViewedElements).
func (l *lowering) collectViewedElements(modules []*ast.SourceFile) error {
	if l.viewedElements == nil || !l.viewedElementsFresh() {
		return l.walkViewedElements(modules)
	}
	key := viewedElementsKey{checker: l.checker, program: l.program}
	if answer, found := l.viewedElements.answers[key]; found && slices.Equal(answer.modules, modules) {
		if viewedElementsCompare.Load() {
			l.viewedElementsCompared(answer)
		}
		l.moduleOrdered = viewedElementsOrdering(answer.moduleOrdered)
		l.moduleOrders = answer.moduleOrders
		l.cyclicModules = answer.cyclicModules
		// Clones, since a lowering writes its own maps and every later lowering's copy comes from these.
		l.provenModuleReads = maps.Clone(answer.provenModuleReads)
		l.entriesRecords = maps.Clone(answer.entriesRecords)
		return viewedElementsStop(answer.stop)
	}
	stop := l.walkViewedElements(modules)
	if l.viewedElements.answers == nil {
		l.viewedElements.answers = map[viewedElementsKey]viewedElementsAnswer{}
	}
	// Other modules for the same checker and program replace the held answer, as another entry replaces
	// moduleOrder's.
	l.viewedElements.answers[key] = l.viewedElementsTaken(modules, stop)
	return stop
}

// viewedElementsFresh is whether everything the walk reads or writes on the lowering is at a fresh lowering's
// zero, so its answer is the checker's and the program's alone and a hit's copies replace nothing.
func (l *lowering) viewedElementsFresh() bool {
	// concrete instantiates every type the walk asks about through a type mapper.
	return l.typeMapper == nil &&
		// representationReason answers a generic's type parameters from its substitution.
		len(l.substitution) == 0 &&
		// slotType answers a kept slot before the checker's contextual type.
		len(l.slots.explicit) == 0 &&
		// intersectionViewFamily answers past its depth bound by how many intersections are being described.
		l.intersectionViewDepth == 0 &&
		// The walk fills a lowering's own element key, which the cache doesn't hold.
		l.elementKey == nil &&
		// A held module order answers moduleOrder without its side effects, and a hit would replace it, as it
		// would each of these others the walk writes.
		l.moduleOrdered == nil && l.moduleOrders == 0 && !l.cyclicModules && l.provenModuleReads == nil &&
		// entriesRecordLiteral answers from its memo, and a hit would replace the memo.
		l.entriesRecords == nil &&
		// objectIntersection ends a field that holds an intersection already being proven.
		len(l.objectIntersections) == 0
}

// viewedElementsTaken is this lowering's answer as the walk over modules left it, in copies, so nothing the
// lowering writes later reaches the answer.
func (l *lowering) viewedElementsTaken(modules []*ast.SourceFile, stop error) viewedElementsAnswer {
	return viewedElementsAnswer{
		modules:           slices.Clone(modules),
		stop:              viewedElementsStop(stop),
		moduleOrdered:     viewedElementsOrdering(l.moduleOrdered),
		moduleOrders:      l.moduleOrders,
		cyclicModules:     l.cyclicModules,
		provenModuleReads: maps.Clone(l.provenModuleReads),
		entriesRecords:    maps.Clone(l.entriesRecords),
	}
}

// viewedElementsOrdering is a copy of a held module order, its order its own.
func viewedElementsOrdering(held *moduleOrdering) *moduleOrdering {
	if held == nil {
		return nil
	}
	return &moduleOrdering{entry: held.entry, order: slices.Clone(held.order), err: viewedElementsStop(held.err)}
}

// viewedElementsStop is a copy of a stop, so no two lowerings hold one error a caller may add to. The walk's
// stops are each a NotYet (its own, and dynamicImportGap's through moduleOrder); any other is held as it is.
func viewedElementsStop(stop error) error {
	if notYet, ok := stop.(*NotYet); ok {
		copied := *notYet
		return &copied
	}
	return stop
}

// viewedElementsCompared walks the program on a scratch copy of this lowering beside an answer taken from the
// cache, and panics where the walk's answer isn't the cached one. The guards hold every field the walk writes at
// its zero, so the copy's walk writes none of this lowering's, which goes on from the cached answer, as it would
// without the comparison.
func (l *lowering) viewedElementsCompared(answer viewedElementsAnswer) {
	scratch := *l
	stop := scratch.walkViewedElements(answer.modules)
	if differences := viewedElementsDifferences(answer, &scratch, stop); len(differences) != 0 {
		panic("viewed elements comparison (#hjspb3b): the cached answer for this checker and program isn't a fresh walk's; " + strings.Join(differences, "; "))
	}
}

// viewedElementsDifferences is every way answer isn't what walked holds after its walk stopped at stop.
func viewedElementsDifferences(answer viewedElementsAnswer, walked *lowering, stop error) []string {
	var differences []string
	if cached, fresh := viewedElementsErrorText(answer.stop), viewedElementsErrorText(stop); cached != fresh {
		differences = append(differences, "stop: cached "+strconv.Quote(cached)+", fresh "+strconv.Quote(fresh))
	}
	switch cached, fresh := answer.moduleOrdered, walked.moduleOrdered; {
	case (cached == nil) != (fresh == nil):
		differences = append(differences, "moduleOrdered: cached held "+strconv.FormatBool(cached != nil)+", fresh held "+strconv.FormatBool(fresh != nil))
	case cached != nil && (cached.entry != fresh.entry || !slices.Equal(cached.order, fresh.order) || viewedElementsErrorText(cached.err) != viewedElementsErrorText(fresh.err)):
		differences = append(differences, "moduleOrdered: cached "+strconv.Itoa(len(cached.order))+" modules, error "+strconv.Quote(viewedElementsErrorText(cached.err))+", fresh "+strconv.Itoa(len(fresh.order))+" modules, error "+strconv.Quote(viewedElementsErrorText(fresh.err)))
	}
	if answer.moduleOrders != walked.moduleOrders {
		differences = append(differences, "moduleOrders: cached "+strconv.Itoa(answer.moduleOrders)+", fresh "+strconv.Itoa(walked.moduleOrders))
	}
	if answer.cyclicModules != walked.cyclicModules {
		differences = append(differences, "cyclicModules: cached "+strconv.FormatBool(answer.cyclicModules)+", fresh "+strconv.FormatBool(walked.cyclicModules))
	}
	if !maps.Equal(answer.provenModuleReads, walked.provenModuleReads) {
		differences = append(differences, "provenModuleReads: cached "+strconv.Itoa(len(answer.provenModuleReads))+" reads, fresh "+strconv.Itoa(len(walked.provenModuleReads)))
	}
	if !maps.Equal(answer.entriesRecords, walked.entriesRecords) {
		differences = append(differences, "entriesRecords: cached "+strconv.Itoa(len(answer.entriesRecords))+" literals, fresh "+strconv.Itoa(len(walked.entriesRecords)))
	}
	if len(walked.objectIntersections) != 0 {
		differences = append(differences, "objectIntersections: the fresh walk left "+strconv.Itoa(len(walked.objectIntersections))+", which the cache doesn't carry")
	}
	return differences
}

// viewedElementsErrorText is an error's text, "" for none.
func viewedElementsErrorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
