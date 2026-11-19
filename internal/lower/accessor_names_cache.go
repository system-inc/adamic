package lower

import (
	"maps"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// walkAccessorNames' answer, accessorNames, is the modules' syntax alone: it reads each getter's name through
// memberKeyText, asks no checker, reads no field of the lowering, and replaces accessorNames whole, so nothing a
// lowering holds before it changes what it leaves. The latent census makes a fresh lowering for each of its
// units, and each walked every module of the program again before it lowered: 60 s of the tsc entry census's CPU,
// where production walks once a lowering (#hjspb3b). So a census run hands its lowerings one accessorNamesCache,
// as it does an inheritanceCache (inheritance_cache.go), and a lowering takes a copy of the answer the first
// lowering over its program walked. The cache is its own file, not inheritance_cache.go's, because its key is
// another: the program alone, where inheritance's holds the checker it asked (#hjspb3b).

// accessorNamesCache is walkAccessorNames' answer for each program a run's lowerings walked. Production makes
// none and walks; the latent census makes one a run (lower.go.txt's latentRun), which statecopy shares. A run
// lowers one unit at a time, so it takes no lock (#hjspb3b).
type accessorNamesCache struct {
	answers map[accessorNamesKey]accessorNamesAnswer
}

// accessorNamesKey is what the answer is a function of, beside the modules the answer holds: the program, and
// no checker, since the walk asks none, so a run's lowerings take one answer across every file's checker where a
// checker in the key would walk once a file. The program holds its modules itself, so the key keeps nothing
// alive the program doesn't (#hjspb3b).
type accessorNamesKey struct {
	program *load.Program
}

// accessorNamesAnswer is one walk's accessorNames, which no lowering holds: each takes a clone of it (#hjspb3b).
type accessorNamesAnswer struct {
	// modules is what the walk read. A census run hands every unit the same modules; an ask with others walks
	// (#hjspb3b).
	modules       []*ast.SourceFile
	accessorNames map[string]bool
}

// accessorNamesWalks, when a test sets it, counts the walks walkAccessorNames makes, so a pin holds a run's
// lowerings to one walk per program. It's nil outside those tests (#hjspb3b).
var accessorNamesWalks atomic.Pointer[atomic.Int64]

// accessorNamesCompare, set only by tests, walks the program again beside every answer taken from the cache and
// panics where the two differ (accessorNamesCompared). Tests that set it aren't parallel, and no parallel test
// runs beside a sequential one (#hjspb3b).
var accessorNamesCompare atomic.Bool

// noteAccessorNames fills accessorNames once a lowering: a copy of the cached answer where this lowering has a
// cache, else its own walk (walkAccessorNames). It needs no guard, as collectViewedElements'
// viewedElementsFresh is: the walk reads nothing of the lowering's, and writes accessorNames alone, whole, which
// a hit replaces as the walk would (#hjspb3b).
func (l *lowering) noteAccessorNames(modules []*ast.SourceFile) {
	if l.accessorNamesHeld == nil {
		l.walkAccessorNames(modules)
		return
	}
	key := accessorNamesKey{program: l.program}
	if answer, found := l.accessorNamesHeld.answers[key]; found && slices.Equal(answer.modules, modules) {
		if accessorNamesCompare.Load() {
			l.accessorNamesCompared(answer)
		}
		// A clone, since a lowering holds its own map and every later lowering's copy comes from this one
		// (#hjspb3b).
		l.accessorNames = maps.Clone(answer.accessorNames)
		return
	}
	l.walkAccessorNames(modules)
	if l.accessorNamesHeld.answers == nil {
		l.accessorNamesHeld.answers = map[accessorNamesKey]accessorNamesAnswer{}
	}
	// Other modules for the same program replace the held answer, as another entry replaces moduleOrder's
	// (#hjspb3b).
	l.accessorNamesHeld.answers[key] = accessorNamesAnswer{modules: slices.Clone(modules), accessorNames: maps.Clone(l.accessorNames)}
}

// accessorNamesCompared walks the program on a scratch copy of this lowering beside an answer taken from the
// cache, and panics where the walk's accessorNames isn't the cached one. The walk gives the copy a map of its
// own, so this lowering goes on from the cached answer, as it would without the comparison (#hjspb3b).
func (l *lowering) accessorNamesCompared(answer accessorNamesAnswer) {
	scratch := *l
	scratch.walkAccessorNames(answer.modules)
	if !maps.Equal(answer.accessorNames, scratch.accessorNames) {
		panic("accessor names comparison (#hjspb3b): the cached answer for this program isn't a fresh walk's; accessorNames: only cached " + accessorNamesOnly(answer.accessorNames, scratch.accessorNames) + ", only fresh " + accessorNamesOnly(scratch.accessorNames, answer.accessorNames))
	}
}

// accessorNamesOnly is every name one map holds that the other lacks or holds with another value, sorted
// (#hjspb3b).
func accessorNamesOnly(one, other map[string]bool) string {
	names := []string{}
	for name, getter := range one {
		if held, found := other[name]; !found || held != getter {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return "[" + strings.Join(names, " ") + "]"
}
