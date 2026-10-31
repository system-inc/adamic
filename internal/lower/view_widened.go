package lower

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// viewSource is a view's target and the type its value had before it: a cast's operand, a checked
// predicate's argument or a tag-proven predicate's parameter. Source is nil when there's none to
// name, and then the view's source guarantees none of its fields.
type viewSource struct {
	source, target *checker.Type
}

// viewLie is what one view checks that its source doesn't guarantee. Lied are the names read raw
// through any other type: one the source lacks, one the source holds as a checked view itself, and
// every name inside a field the source doesn't hold as the view does. Held are the names the
// source has, each with the source's type for it, honest through a type whose field it fits.
// Honest are the names inside a field the source holds as the view does.
type viewLie struct {
	lied, honest map[string]bool
	held         map[string]*checker.Type
}

// A checked view's fields are checked where they're read, under the view's receiver type
// (CheckedFieldKey). The value the view admitted can be seen as another type after: widened by an
// assignment, an argument, a return, a field or element store, a capture or a conditional's arm.
// A read through that type is a plain read, so it read a lying layout's field as the type it
// claims, and native crashed where Node printed the field. Until such a read checks against the
// view, a read of a field a view checks, through a receiver type that doesn't check it, is NotYet
// when the receiver may hold a viewed allocation, failing closed when the reach is unknown, as
// checkViewCopies is for a copy. A field the view's source already held is honest through a type
// it fits, so the source's own reads (the tag test before the cast) still lower. Views register
// as their casts lower, so this is decided once lowering is done.
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
	for i, view := range l.viewSources {
		lies[i] = l.viewLie(view)
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
		if !isProperty || read.ViewWhere == "" || !names[read.Name] || l.result.CheckedField(read.ViewReceiverTypeID, read.Name) || viewVouched(lies, read.Name, l.viewTypes[read.ViewTypeID], l.checker) {
			return true
		}
		if graph == nil {
			var reachUnknown bool
			graph, viewed, reachUnknown = viewedAllocations(l.result)
			unknown = unknown || reachUnknown
		}
		reaches := graph.ReachingAllocations(read.Object)
		demanded := unknown || reaches.Unknown
		for _, site := range reaches.Sites {
			demanded = demanded || viewed[site]
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

// viewLie is what a view checks that its source doesn't guarantee: its target's own fields, and
// the fields of the contracts inside them.
func (l *lowering) viewLie(view viewSource) viewLie {
	lie := viewLie{lied: map[string]bool{}, honest: map[string]bool{}, held: map[string]*checker.Type{}}
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
			lie.lied[property.Name] = true
		} else {
			lie.held[property.Name] = l.checker.GetTypeOfSymbol(held)
		}
		if held == nil || !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(held), l.checker.GetTypeOfSymbol(property)) {
			l.viewContractNames(contracts[property.Name], lie.lied, map[ir.ViewContractID]bool{})
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
// guaranteed. A read whose field type lowering didn't record is never vouched for, and nor is a
// name no recorded view accounts for.
func viewVouched(lies []viewLie, name string, read *checker.Type, typeChecker *checker.Checker) bool {
	vouched := false
	for _, lie := range lies {
		if lie.lied[name] {
			return false
		}
		if held, ok := lie.held[name]; ok {
			if read == nil || !typeChecker.IsTypeAssignableTo(held, read) {
				return false
			}
			vouched = true
		}
		vouched = vouched || lie.honest[name]
	}
	return vouched
}
