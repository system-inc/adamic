package lower

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// viewSource is a view's target, the type its value had before it (a cast's operand, a checked
// predicate's argument or a tag-proven predicate's parameter, nil when there's none to name), and
// where its value enters the view in the IR: a cast's CheckedCast, the local a structural cast
// admits, or, for a predicate's view, wherever a read is narrowed to the target.
type viewSource struct {
	source, target *checker.Type
	cast, narrows  bool
	entry          ir.Expression
}

// viewLie is what one view checks that its source doesn't guarantee. Lied are the names read raw
// through any other type: one the source lacks, one the source holds as a checked view itself
// (both lied at the top), and every name inside a field the source doesn't hold as the view does
// (lied inside). Held are the names the source has, each with the source's type for it, honest
// through a type whose field it fits. Honest are the names inside a field the source holds as the
// view does.
type viewLie struct {
	liedTop, liedInside, honest map[string]bool
	held                        map[string]*checker.Type
}

// The sites a viewed value's flow reaches where it enters a view, which no allocation has (those
// are negative): the view's own value, and a value a checked read took out of a viewed one.
const (
	viewedValueSite = 1
	viewedFieldSite = 2
)

// A checked view's fields are checked where they're read, under the view's receiver type
// (CheckedFieldKey). The value the view admitted can be seen as another type after: widened by an
// assignment, an argument, a return, a field or element store, a capture or a conditional's arm.
// A read through that type is a plain read, so it read a lying layout's field as the type it
// claims, and native crashed where Node printed the field. Until such a read checks against the
// view, a read of a field a view checks, through a receiver type that doesn't check it, is NotYet
// when the receiver may hold a viewed value, failing closed when the reach is unknown.
//
// Whether it may is asked of the shared allocation graph with an entry hook: the receiver's flow
// reaches where a value enters a view, and stops there. Asking which allocations it reaches would
// refuse every read made before the view (the tag test, the producer's own type), since those
// reach the same allocation, and a layout lies only once a view claims it. The entries, one for
// each place view registers one:
//
//   - a cast's CheckedCast (interfaceCast, and cast's tag cast), the value itself;
//   - a structural cast (structuralViewCast), which leaves no IR node of its own: every read of
//     the local it casts, a superset of where the admitted value goes;
//   - a predicate's view, a checked call's (predicateMembership) or a tag-proven predicate's
//     (predicateRefusal), no node either: every read of each local the checker narrowed some
//     read of to the target (narrowedEntries), which is where a true result admits the value;
//   - a checked read of an aggregate field, the value inside a viewed value.
//
// A view whose entry names no local (a predicate that narrows a field read, a structural cast of
// something other than a local) falls back to the allocations it may reach, as checkViewCopies
// does. Predicate views give no ViewOrigin, only nil, so that reach is unknown and every read the
// fallback weighs is NotYet. The flow follows what the graph follows: locals (a captured one is
// the same local, in its cell), parameters, returns, conditional arms, and literal fields and
// elements at a constant index. Anything else, a Map value, a for-of element or one at a computed
// index, a callback's parameter, is Unknown, and Unknown is NotYet.
//
// A field the view's source already held is honest through a type it fits, as the cast changed
// nothing the source guaranteed. Views register as their casts lower, so this is decided once
// lowering is done.
func (l *lowering) checkViewWidenedReads() error {
	names := map[string]bool{}
	for key := range l.result.CheckedFields {
		// A record's field is checked under its bare name, and is no view's.
		receiver, name, keyed := strings.Cut(key, ":")
		if _, err := strconv.Atoi(receiver); !keyed || err != nil {
			continue
		}
		names[name] = true
	}
	if len(names) == 0 {
		return nil
	}
	lies := make([]viewLie, len(l.viewSources))
	entryLocals := map[int]bool{}
	var unentered []viewLie
	for i, view := range l.viewSources {
		lies[i] = l.viewLie(view)
		read, isRead := view.entry.(ir.Read)
		switch {
		case view.cast:
		case view.narrows:
			if !l.narrowedEntries(view.target, entryLocals) {
				unentered = append(unentered, lies[i])
			}
		case isRead:
			entryLocals[read.Local] = true
		default:
			unentered = append(unentered, lies[i])
		}
	}
	assignAllocationSites(l.result)
	entered := newAllocationFlowGraph(l.result)
	entered.entry = func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.CheckedCast:
			if value.CheckedFields {
				return viewedValueSite
			}
		case ir.Read:
			if entryLocals[value.Local] {
				return viewedValueSite
			}
		case ir.Property:
			if l.result.CheckedFields[ir.CheckedFieldKey(value.ViewReceiverTypeID, value.Name)] && viewAggregate(value) {
				return viewedFieldSite
			}
		}
		return 0
	}
	var graph *allocationFlowGraph
	var viewed map[int]bool
	unknown := len(l.result.ViewOrigins) == 0
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		read, isProperty := node.(ir.Property)
		// A read with no place in the source is lowering's own, a predicate's tag test, and no read
		// through a type. A method read counts: a callable field is read raw too.
		if !isProperty || read.ViewWhere == "" || !names[read.Name] || l.result.CheckedField(read.ViewReceiverTypeID, read.Name) {
			return true
		}
		field := l.viewTypes[read.ViewTypeID]
		reaches := entered.ReachingAllocations(read.Object)
		inside := reaches.Unknown || slices.Contains(reaches.Sites, viewedFieldSite)
		demanded := inside && !viewVouched(lies, read.Name, field, true, l.checker) || slices.Contains(reaches.Sites, viewedValueSite) && !viewVouched(lies, read.Name, field, false, l.checker)
		if !demanded && len(unentered) != 0 && !viewVouched(unentered, read.Name, field, true, l.checker) {
			if graph == nil {
				var reachUnknown bool
				graph, viewed, reachUnknown = viewedAllocations(l.result, false)
				unknown = unknown || reachUnknown
			}
			allocations := graph.ReachingAllocations(read.Object)
			demanded = unknown || allocations.Unknown
			for _, site := range allocations.Sites {
				demanded = demanded || viewed[site]
			}
		}
		if demanded {
			through := "an unnamed type"
			if receiver := l.viewTypes[read.ViewReceiverTypeID]; receiver != nil {
				through = "type " + l.checker.TypeToString(receiver)
			}
			refused = &NotYet{Where: read.ViewWhere, What: "a read of field " + read.Name + " through " + through + ", which may hold a checked view's value"}
			return false
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	return refused
}

// narrowedEntries adds to locals each local a read of was narrowed to a type holding target, the
// value a predicate's true result admits at its caller. It says false when a field read was
// narrowed to one too, a value with no local to name.
func (l *lowering) narrowedEntries(target *checker.Type, locals map[int]bool) bool {
	for _, narrowed := range l.narrowedFields {
		if l.typeHolds(narrowed, target) {
			return false
		}
	}
	for local, types := range l.narrowedReads {
		for _, narrowed := range types {
			locals[local] = locals[local] || l.typeHolds(narrowed, target)
		}
	}
	return true
}

// typeHolds says whether a type is target, or a union or intersection holding it.
func (l *lowering) typeHolds(held, target *checker.Type) bool {
	held = l.checker.GetNonNullableType(l.concrete(held))
	if held.Id() == target.Id() {
		return true
	}
	if held.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range held.Types() {
			if l.typeHolds(member, target) {
				return true
			}
		}
	}
	return false
}

// narrowedRead keeps the type the checker narrowed a read of a local to, when it isn't the
// local's own (narrowedEntries).
func (l *lowering) narrowedRead(node *ast.Node, local int) {
	symbol := l.symbol(node)
	if symbol == nil {
		return
	}
	narrowed := l.checker.GetTypeAtLocation(node)
	if narrowed.Id() == l.checker.GetTypeOfSymbol(symbol).Id() {
		return
	}
	if l.narrowedReads == nil {
		l.narrowedReads = map[int]map[int]*checker.Type{}
	}
	if l.narrowedReads[local] == nil {
		l.narrowedReads[local] = map[int]*checker.Type{}
	}
	l.narrowedReads[local][int(narrowed.Id())] = narrowed
}

// viewLie is what a view checks that its source doesn't guarantee: its target's own fields, and
// the fields of the contracts inside them.
func (l *lowering) viewLie(view viewSource) viewLie {
	lie := viewLie{liedTop: map[string]bool{}, liedInside: map[string]bool{}, honest: map[string]bool{}, held: map[string]*checker.Type{}}
	var source *checker.Type
	if view.source != nil {
		source = l.checker.GetNonNullableType(l.concrete(view.source))
	}
	contracts := map[string]ir.ViewContractID{}
	if id := l.result.ViewContractTypes[int(view.target.Id())]; id != 0 {
		for _, field := range l.result.ViewContracts[id-1].Fields {
			contracts[field.Name] = field.Contract
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(view.target) {
		var held *ast.Symbol
		if source != nil && !l.result.CheckedField(l.viewReceiverTypeID(source), property.Name) {
			held = l.checker.GetPropertyOfType(source, property.Name)
		}
		if held == nil {
			lie.liedTop[property.Name] = true
		} else {
			lie.held[property.Name] = l.checker.GetTypeOfSymbol(held)
		}
		if held == nil || !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(held), l.checker.GetTypeOfSymbol(property)) {
			l.viewContractNames(contracts[property.Name], lie.liedInside, map[ir.ViewContractID]bool{})
		} else {
			l.viewContractNames(contracts[property.Name], lie.honest, map[ir.ViewContractID]bool{})
		}
	}
	return lie
}

// viewContractNames adds every field name a contract checks, through its fields and members.
func (l *lowering) viewContractNames(id ir.ViewContractID, names map[string]bool, seen map[ir.ViewContractID]bool) {
	if id == 0 || int(id) > len(l.result.ViewContracts) || seen[id] {
		return
	}
	seen[id] = true
	contract := l.result.ViewContracts[id-1]
	for _, field := range contract.Fields {
		names[field.Name] = true
		l.viewContractNames(field.Contract, names, seen)
	}
	for _, member := range contract.Members {
		l.viewContractNames(member, names, seen)
	}
}

// viewVouched says whether every view that checks name had it from its source, as a type that
// fits read, the field's type where it's read: then the read sees what the source already
// guaranteed. Inside is whether the value may be one inside a viewed value, where a name the view
// lies about in a field counts too and a name inside an honest field is vouched for. A read whose
// field type lowering didn't record is never vouched for, and nor is a name no view accounts for.
func viewVouched(lies []viewLie, name string, read *checker.Type, inside bool, typeChecker *checker.Checker) bool {
	vouched := false
	for _, lie := range lies {
		if lie.liedTop[name] || inside && lie.liedInside[name] {
			return false
		}
		if held, ok := lie.held[name]; ok {
			if read == nil || !typeChecker.IsTypeAssignableTo(held, read) {
				return false
			}
			vouched = true
		}
		vouched = vouched || inside && lie.honest[name]
	}
	return vouched
}
