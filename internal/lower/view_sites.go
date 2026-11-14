package lower

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// view judged every property access of the program at every cast into a checked view, walking every node
// of every module each time, so lowering cost casts times program size: tsc's 1,700 such casts over 79
// files spent about 51 s of the census's 183 s of unit time there (#q1hyhzc). An access's answer,
// checkedViewReceiverField, reads only the CheckedFields keys with its name, a view member's "id:name" or
// a record's bare name, and PlainReceivers for its id, which plainReceiver decides at the id's first ask
// and never again. So it changes only when a key of its name is added, and a true answer stays true as
// keys only grow. The walk now runs over an index of the program's accesses built once, and asks again
// only the accesses whose name gained a key since view last judged them (viewWalkIndexed).

// viewSiteIndex is every property access and every named getter of a program's modules, each in the
// preorder view's walk visits them: ForEachChild over each module node, in module order. It's a function
// of the program alone, so one serves every lowering of a latent run (stage3/census/latent's statecopy
// shares it); modules is the order it was built from, and a different one builds it again.
type viewSiteIndex struct {
	modules  []*ast.SourceFile
	accesses []*ast.Node
	getters  []*ast.Node
}

// viewSiteIndexBuilds, when a test sets it, counts the site indexes built, so a pin holds view to one
// build however many casts a program makes. It's nil outside that test.
var viewSiteIndexBuilds atomic.Pointer[atomic.Int64]

// viewWalkCompare, set only by tests, runs viewWalkWhole beside every indexed walk and panics where the
// two disagree (viewWalkCompared). Tests that set it aren't parallel, and no parallel test runs beside a
// sequential one.
var viewWalkCompare atomic.Bool

// sites is the index for modules, built again when it was built from another order.
func (index *viewSiteIndex) sites(modules []*ast.SourceFile) *viewSiteIndex {
	if index.modules != nil && slices.Equal(index.modules, modules) {
		return index
	}
	index.modules, index.accesses, index.getters = modules, nil, nil
	// viewWalkWhole's recursion, without its stop, so the order is its order.
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if part.Kind == ast.KindPropertyAccessExpression {
			index.accesses = append(index.accesses, part)
		}
		if part.Kind == ast.KindGetAccessor && part.Name() != nil {
			index.getters = append(index.getters, part)
		}
		part.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if counter := viewSiteIndexBuilds.Load(); counter != nil {
		counter.Add(1)
	}
	return index
}

// markCheckedField is the one place lowering adds a CheckedFields key, so viewWalkIndexed knows which
// names gained one: checkedFieldWrites counts each new key under the field name it checks. Only
// readiness writes past it, and only once lowering is done, where no view is made again.
// TestCheckedFieldsAreWrittenThroughMarkCheckedField holds every other write to it.
func (l *lowering) markCheckedField(key, name string) {
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	if l.result.CheckedFields[key] {
		return
	}
	l.result.CheckedFields[key] = true
	if l.checkedFieldWrites == nil {
		l.checkedFieldWrites = map[string]int{}
	}
	l.checkedFieldWrites[name]++
}

// viewWalk judges every property access a view's checked fields reach, refusing the first that can't be
// checked, in preorder, and gives the getters named as its fields. A view made under a type mapper walks
// whole: concrete instantiates an access's receiver by the mapper in force, so an access can be asked
// under an id no view without it asks, and its answer isn't the one the index holds.
func (l *lowering) viewWalk(node *ast.Node, modules []*ast.SourceFile, fields map[string]bool) ([]*ast.Node, error) {
	if l.typeMapper != nil {
		return l.viewWalkWhole(modules, fields)
	}
	if viewWalkCompare.Load() {
		return l.viewWalkCompared(node, modules, fields)
	}
	return l.viewWalkIndexed(modules, fields)
}

// viewWalkIndexed is viewWalkWhole over the index. An access found checked is checked still. An access
// whose name gained no key since its last judging can't have changed, so it's skipped, side effects
// included: its ids were asked then, and plainReceiver answers an id once. Any other access is asked as
// the whole walk asks it. A name's count is read at the call's start, so a key one access adds during
// the walk leaves its name to be asked again next call: an access before it in preorder, such as a
// rigid type parameter's whose constraint is the union another access registers, found unchecked here
// as the whole walk finds it, and checked from the next view on.
func (l *lowering) viewWalkIndexed(modules []*ast.SourceFile, fields map[string]bool) ([]*ast.Node, error) {
	if l.viewSites == nil {
		l.viewSites = &viewSiteIndex{}
	}
	index := l.viewSites.sites(modules)
	written := maps.Clone(l.checkedFieldWrites)
	// stale is each name met in this walk, true when its accesses are asked: never judged, or judged at
	// another count.
	stale := map[string]bool{}
	for _, part := range index.accesses {
		name := part.Name().Text()
		known := l.viewSiteTrue[part]
		if !known {
			asked, met := stale[name]
			if !met {
				seen, judged := l.viewSiteSeen[name]
				asked = !judged || seen != written[name]
				stale[name] = asked
			}
			if !asked {
				continue
			}
		}
		var refused error
		// viewWalkWhole's two refusals, unchanged; a write's overrides an accessor's at one access, as there.
		if known || l.checkedViewReceiverField(part, l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression), name) {
			l.markViewSiteTrue(part)
			access := part.AsPropertyAccessExpression()
			if base, _ := l.representation(l.checker.GetTypeAtLocation(access.Expression)); base == ir.Object {
				field := l.checker.GetSymbolAtLocation(part.Name())
				if field != nil && len(l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(field), checker.SignatureKindCall)) == 0 && accessorSymbol(field) {
					refused = l.notYet(part, checkedFieldAliasAccessor)
				}
			}
		}
		if known || l.checkedViewReceiverField(part, l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression), name) {
			l.markViewSiteTrue(part)
			if ast.IsAssignmentTarget(part) {
				if symbol := l.checker.GetSymbolAtLocation(part.Name()); symbol != nil {
					of, _ := l.representation(l.checker.GetTypeOfSymbol(symbol))
					if of == ir.Object {
						refused = l.notYet(part, "writing a checked object field without its source-slot type certificate")
					}
				}
			}
		}
		// A refusal leaves viewSiteSeen as it was: the accesses after it weren't asked.
		if refused != nil {
			return nil, refused
		}
	}
	for name, asked := range stale {
		if asked {
			if l.viewSiteSeen == nil {
				l.viewSiteSeen = map[string]int{}
			}
			l.viewSiteSeen[name] = written[name]
		}
	}
	// A getter named as a contract field is judged by what can reach this view, once lowering is done
	// (refuseReachedContractGetters), never by its name alone (#c7w6j0y).
	var getters []*ast.Node
	for _, getter := range index.getters {
		if fields[getter.Name().Text()] {
			getters = append(getters, getter)
		}
	}
	return getters, nil
}

func (l *lowering) markViewSiteTrue(part *ast.Node) {
	if l.viewSiteTrue == nil {
		l.viewSiteTrue = map[*ast.Node]bool{}
	}
	l.viewSiteTrue[part] = true
}

// viewWalkState is what the whole walk can change and the indexed walk keeps, held apart so the
// comparison mode runs both from one state.
type viewWalkState struct {
	checkedFields      map[string]bool
	plainReceivers     ir.SortedMap[int, bool]
	plainAsked         map[int]bool
	viewTypes          map[int]*checker.Type
	checkedFieldWrites map[string]int
	viewSiteSeen       map[string]int
	viewSiteTrue       map[*ast.Node]bool
}

func (l *lowering) viewWalkState() viewWalkState {
	return viewWalkState{
		checkedFields:      maps.Clone(l.result.CheckedFields),
		plainReceivers:     maps.Clone(l.result.PlainReceivers),
		plainAsked:         maps.Clone(l.plainAsked),
		viewTypes:          maps.Clone(l.viewTypes),
		checkedFieldWrites: maps.Clone(l.checkedFieldWrites),
		viewSiteSeen:       maps.Clone(l.viewSiteSeen),
		viewSiteTrue:       maps.Clone(l.viewSiteTrue),
	}
}

func (l *lowering) restoreViewWalkState(state viewWalkState) {
	l.result.CheckedFields = state.checkedFields
	l.result.PlainReceivers = state.plainReceivers
	l.plainAsked = state.plainAsked
	l.viewTypes = state.viewTypes
	l.checkedFieldWrites = state.checkedFieldWrites
	l.viewSiteSeen = state.viewSiteSeen
	l.viewSiteTrue = state.viewSiteTrue
}

// viewWalkCompared runs the whole walk, puts back the state it began from, runs the indexed walk, and
// panics where the two differ in the refusal, the getters, CheckedFields or PlainReceivers. Lowering goes
// on from the indexed walk's state, so the rest of the program is lowered as production lowers it.
func (l *lowering) viewWalkCompared(node *ast.Node, modules []*ast.SourceFile, fields map[string]bool) ([]*ast.Node, error) {
	before := l.viewWalkState()
	wholeGetters, wholeRefused := l.viewWalkWhole(modules, fields)
	wholeFields, wholePlain := l.result.CheckedFields, l.result.PlainReceivers
	l.restoreViewWalkState(before)
	getters, refused := l.viewWalkIndexed(modules, fields)
	var differences []string
	if fmt.Sprint(wholeRefused) != fmt.Sprint(refused) {
		differences = append(differences, "refusal: whole "+fmt.Sprint(wholeRefused)+", indexed "+fmt.Sprint(refused))
	}
	if !slices.Equal(wholeGetters, getters) {
		differences = append(differences, "getters: whole "+l.viewWalkPlaces(wholeGetters)+", indexed "+l.viewWalkPlaces(getters))
	}
	if !maps.Equal(wholeFields, l.result.CheckedFields) {
		differences = append(differences, "CheckedFields: only whole "+viewWalkOnly(wholeFields, l.result.CheckedFields, func(key string) string { return key })+", only indexed "+viewWalkOnly(l.result.CheckedFields, wholeFields, func(key string) string { return key }))
	}
	if !maps.Equal(wholePlain, l.result.PlainReceivers) {
		differences = append(differences, "PlainReceivers: only whole "+viewWalkOnly[int](wholePlain, l.result.PlainReceivers, strconv.Itoa)+", only indexed "+viewWalkOnly[int](l.result.PlainReceivers, wholePlain, strconv.Itoa))
	}
	if len(differences) != 0 {
		panic("view walk comparison (#q1hyhzc): the indexed walk of the view at " + l.program.Where(node) + " isn't the whole walk's; " + strings.Join(differences, "; "))
	}
	return getters, refused
}

// viewWalkPlaces is where each getter is, for a comparison's message.
func (l *lowering) viewWalkPlaces(getters []*ast.Node) string {
	places := []string{}
	for _, getter := range getters {
		places = append(places, l.program.Where(getter))
	}
	return "[" + strings.Join(places, " ") + "]"
}

// viewWalkOnly is every key of one map that the other lacks or holds with another value, sorted.
func viewWalkOnly[K comparable](one, other map[K]bool, text func(K) string) string {
	keys := []string{}
	for key, value := range one {
		if held, found := other[key]; !found || held != value {
			keys = append(keys, text(key))
		}
	}
	sort.Strings(keys)
	return "[" + strings.Join(keys, " ") + "]"
}
