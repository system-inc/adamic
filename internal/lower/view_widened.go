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
// admits, or, for a predicate's view, wherever a read is narrowed to the target. Node is the view's
// site, where a graph program's view pair the cycle finder can't relate is named (viewPairGuard).
type viewSource struct {
	source, target *checker.Type
	cast, narrows  bool
	// untagged is a view an untagged union's cast made (untaggedUnionCast), whose cast tests nothing, so
	// viewPairGuard pairs every arm of its source with its target's, related or not.
	untagged bool
	entry    ir.Expression
	node     *ast.Node
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
// A read through that type was a plain read, so it read a lying layout's field as the type it
// claims, and native crashed where Node printed the field. So a read of a field a view checks,
// through a receiver type that doesn't check it, is checked where it's made (Program.WidenedReads,
// #zpzm9ta head 1) when the receiver may hold a viewed value, and when the reach is unknown. It's
// checked by name against the read's own declared type, which is all the program assumes of its
// result, and a lie stops there with the checked read's located message. A read whose name a
// getter has stays NotYet, since a checked read can't run one.
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
// fallback weighs is checked. The flow follows what the graph follows: locals (a captured one is
// the same local, in its cell), parameters, returns, conditional arms, and literal fields and
// elements at a constant index. Anything else, a Map value, a for-of element or one at a computed
// index, a callback's parameter, is Unknown, and a read on Unknown reach is checked.
//
// A field the view's source already held is honest through a type it fits, as the cast changed
// nothing the source guaranteed. Views register as their casts lower, so this is decided once
// lowering is done.
//
// Which reads it weighs is viewReadOf's (view_reads.go): a field read of the NotYet class, and a
// whole-value reader (encodeJson), which reads every field of its value and of what that holds raw,
// so it's NotYet when anything in it may come from a view's entry, vouched for or not.
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
	entry := func(value ir.Expression) int {
		switch value := value.(type) {
		case ir.CheckedCast:
			if value.CheckedFields {
				return viewedValueSite
			}
		case ir.Read:
			if entryLocals[value.Local] {
				return viewedValueSite
			}
		case ir.Call:
			// An overload result check's call admits its value into whatever view its membership reads
			// through, and the caller's local holds the result as the call's declared type, never a
			// narrowing, so narrowedEntries finds no read of it. Every such call is an entry, not only
			// one whose check registered a view: a second call site of the same overload finds its view
			// cached and registers none (Language's q06 and v10 on #z5d0jye). A call whose result no view
			// checks adds nothing, since only names some view checks are weighed.
			if l.overloadCheckCalls[value.Function] {
				return viewedValueSite
			}
		case ir.Property:
			// A widened read this pass checked takes a value out of one that may be viewed, as a view member's
			// checked read does, so what it reads is weighed as inside a viewed value.
			if (l.result.CheckedFields[ir.CheckedFieldKey(value.ViewReceiverTypeID, value.Name)] || l.result.WidenedReads[ir.WidenedReadKey(value)]) && viewAggregate(value) {
				return viewedFieldSite
			}
		}
		return 0
	}
	var entered *allocationFlowGraph
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
		if !isProperty || read.ViewWhere == "" || !names[read.Name] || viewReadOf(l.result, read) != viewReadNotYet {
			return true
		}
		field := l.viewTypes[read.ViewTypeID]
		reaches := entered.ReachingAllocations(read.Object)
		inside := reaches.Unknown || slices.Contains(reaches.Sites, viewedFieldSite)
		demanded := inside && !viewVouched(lies, read.Name, field, true, l.checker) || slices.Contains(reaches.Sites, viewedValueSite) && !viewVouched(lies, read.Name, field, false, l.checker)
		if !demanded && len(unentered) != 0 && !viewVouched(unentered, read.Name, field, true, l.checker) {
			if graph == nil {
				var reachUnknown bool
				graph, viewed, reachUnknown = l.viewedAllocations(l.result, false)
				unknown = unknown || reachUnknown
			}
			allocations := graph.ReachingAllocations(read.Object)
			demanded = unknown || allocations.Unknown
			for _, site := range allocations.Sites {
				demanded = demanded || viewed[site]
			}
		}
		if !demanded {
			return true
		}
		if read.Method {
			// A method read keeps its view only where readiness certifies the callable
			// (certifiedUntaggedCallableRead), and is a plain read otherwise, so marking it would let it through
			// unchecked. It stays NotYet until the adapter's road for it is traced.
			refused = &NotYet{Where: read.ViewWhere, What: "a read of method " + read.Name + " through " + l.widenedThrough(read) + ", which may hold a checked view's value"}
			return false
		}
		if l.accessorNames[read.Name] {
			// A checked read looks only at data slots, so where a getter answers the name it reads "found missing" and
			// stops a program Node runs. That divergence is caught here, at compile time, until the checked read can call
			// a getter (Compiler on #zpzm9ta).
			refused = &NotYet{Where: read.ViewWhere, What: "a read of field " + read.Name + " through " + l.widenedThrough(read) + ", which may hold a checked view's value, where a getter named " + read.Name + " may answer it"}
			return false
		}
		// The read is checked where it's made, by name, against its own declared type: the contract viewedField gave
		// it, which readiness keeps once CheckedRead says so. That's enough because the read's result is used at that
		// type, so a stored value that passes the check by name and representation is exactly what the rest of the
		// program assumes, and a lie deeper down stops at the read that meets it, which is weighed here the same way.
		// The mark must be made before readiness strips an unchecked read's view (lower.go runs this pass before
		// readiness), or it changes nothing. A key two reads share, two lowerings of one source read or two reads
		// with no place, checks both, which only adds checks. A generator frame, which has no shape, never reaches a
		// named-field read: checkGeneratorRoads keeps it to its iteration.
		if l.result.WidenedReads == nil {
			l.result.WidenedReads = ir.SortedMap[string, bool]{}
		}
		l.result.WidenedReads[ir.WidenedReadKey(read)] = true
		return true
	}
	// Each mark can make a read of what it reads demanded, and the graph keeps what it reached (its projection
	// frames), so each round asks a fresh one, until a round marks nothing new. A marked read is checked, so a later
	// round passes over it, and the rounds end.
	for {
		entered = l.flowGraph(l.result)
		entered.entry = entry
		marked := len(l.result.WidenedReads)
		walk(l.result.Main, inspect)
		for _, function := range l.result.Functions {
			walk(function.Body, inspect)
		}
		// A read finishAccessors made a dispatcher's call isn't a Property in the program any more, and its
		// dispatcher's fallback reads the slot unchecked, so it's weighed here as it was lowered, its receiver as
		// lowering made it (#3jj9dzg). Its own receiver is all that's walked: a read inside that receiver is in
		// the program as itself, or a dispatched read of its own. A getter answers its name, so where it's demanded
		// it stops NotYet by the getter rule above and is never marked.
		for _, read := range l.dispatchedReads {
			if l.firstLinkSkipsAt(read.ViewWhere) {
				continue
			}
			inspect(read)
		}
		if refused != nil {
			return refused
		}
		if len(l.result.WidenedReads) == marked {
			break
		}
	}
	if len(l.result.WidenedReads) != 0 {
		// checkLazyViewReads judged every checked read when findCycles ran it, before these were marked, so the marked
		// reads are judged now by the same rules: what emission can't express, or an unsupported contract family,
		// stops here with its own message, never at an emitter's compiler-bug panic.
		if err := l.judgeCheckedViewReads(func(read ir.Property) bool { return l.result.WidenedReads[ir.WidenedReadKey(read)] }); err != nil {
			return err
		}
	}
	// A whole-value reader reads every field of the value and of what it holds, through no view and
	// no type the source vouched for, so nothing a view claimed may be anywhere in it.
	for _, whole := range l.wholeObjectCopies {
		if l.firstLinkSkips(whole.node) {
			continue
		}
		if whole.reader == "" {
			continue
		}
		demanded := deepReach(entered, whole.source, func(site int) bool { return site == viewedValueSite || site == viewedFieldSite })
		if !demanded && len(unentered) != 0 {
			if graph == nil {
				var reachUnknown bool
				graph, viewed, reachUnknown = l.viewedAllocations(l.result, false)
				unknown = unknown || reachUnknown
			}
			demanded = unknown || deepReach(graph, whole.source, func(site int) bool { return viewed[site] })
		}
		if demanded {
			if whole.reader == "bind lookup" {
				return l.notYet(whole.node, "bind through a receiver that may hold a checked view")
			}
			return l.notYet(whole.node, "a read of every field of a value of type "+l.checker.TypeToString(whole.copied)+" by "+whole.reader+", which may hold a checked view's value")
		}
	}
	return nil
}

// widenedThrough names the type a widened read is made through, for its message.
func (l *lowering) widenedThrough(read ir.Property) string {
	if receiver := l.viewTypes[read.ViewReceiverTypeID]; receiver != nil {
		return "type " + l.checker.TypeToString(receiver)
	}
	return "an unnamed type"
}

// deepReach says whether a value, or anything its fields, elements and stores hold, may reach a
// site hit names, or an unknown one.
func deepReach(graph *allocationFlowGraph, value ir.Expression, hit func(int) bool) bool {
	index := graph.projectionIndex()
	seen := map[int]bool{}
	queue := []ir.Expression{value}
	for len(queue) != 0 {
		value := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		reaches := graph.ReachingAllocations(value)
		if reaches.Unknown {
			return true
		}
		for _, site := range reaches.Sites {
			if hit(site) {
				return true
			}
			if seen[site] {
				continue
			}
			seen[site] = true
			if literal, ok := index.records[site]; ok {
				for _, field := range literal.Fields {
					if viewAggregate(field.Value) {
						queue = append(queue, field.Value)
					}
				}
				if literal.Spread != nil {
					queue = append(queue, literal.Spread)
				}
			}
			if literal, ok := index.arrays[site]; ok {
				for _, element := range literal.Elements {
					if viewAggregate(element) {
						queue = append(queue, element)
					}
				}
			}
			for _, values := range index.stores[site] {
				for _, stored := range values {
					if viewAggregate(stored) {
						queue = append(queue, stored)
					}
				}
			}
		}
	}
	return false
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

// typeHolds says whether a type is target, or a union or intersection holding it. A member is
// taken as it is: GetNonNullableType of a type parameter is an intersection holding it again.
func (l *lowering) typeHolds(held, target *checker.Type) bool {
	held = l.checker.GetNonNullableType(l.concrete(held))
	return held.Id() == target.Id() || l.membersHold(held, target, map[*checker.Type]bool{})
}

func (l *lowering) membersHold(held, target *checker.Type, seen map[*checker.Type]bool) bool {
	if held.Id() == target.Id() {
		return true
	}
	if seen[held] || held.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) == 0 {
		return false
	}
	seen[held] = true
	for _, member := range held.Types() {
		if l.membersHold(member, target, seen) {
			return true
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
		if source != nil && !l.result.CheckedField(l.viewFieldReceiverTypeID(nil, source), property.Name) {
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
