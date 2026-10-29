package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Unsupported contracts are descriptors, never certificates. Their users must
// refuse at a demanded read. Cast admission does not inspect their payloads.
func (l *lowering) viewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	family := l.unsupportedViewFamily(target)
	if of, known := l.viewRepresentation(target); target.Flags()&checker.TypeFlagsUnion != 0 && (!interfaceScalar(target) || !known || of == ir.Union) {
		return viewUnionContractHook(l, node, target, func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) })
	}
	if l.checker.IsArrayType(target) || checker.IsTupleType(target) {
		return l.unionAggregateContract(node, target)
	}
	if l.callableViewContract(target) {
		family = "callable"
	}
	if family == "" {
		id, err := l.strictViewContract(node, target)
		if err == nil {
			return id, nil
		}
		family = "representation conversion"
	}
	// A failing family adapter may have reserved its recursive id already.
	id := l.result.ViewContractTypes[int(target.Id())]
	descriptor := ir.ViewContract{Kind: ir.ViewUnknown, Name: l.checker.TypeToString(target), Unsupported: family}
	if id == 0 {
		id = ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, descriptor)
		l.result.ViewContractTypes[int(target.Id())] = id
	} else {
		l.result.ViewContracts[id-1] = descriptor
	}
	return id, nil
}

func (l *lowering) unsupportedViewFamily(target *checker.Type) string {
	if l.phantomUndefined(target) || target.Flags()&checker.TypeFlagsUndefined != 0 {
		return ""
	}
	if base := l.phantomBase(target); base != nil && interfaceScalar(base) {
		return ""
	}
	flags := target.Flags()
	switch {
	case flags&checker.TypeFlagsAny != 0:
		return "any"
	case flags&checker.TypeFlagsUnknown != 0:
		return "unknown"
	case flags&checker.TypeFlagsNever != 0:
		return "never"
	case flags&checker.TypeFlagsTypeParameter != 0:
		return "generic"
	case flags&checker.TypeFlagsIntersection != 0:
		return "intersection"
	case flags&checker.TypeFlagsNull != 0:
		return ""
	case flags&checker.TypeFlagsUndefined != 0:
		return "nullish"
	case isClassInstance(target):
		return "nominal class"
	case l.isLibraryType(target, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		return "collection"
	case checker.IsTupleType(target):
		return "tuple"
	case flags&checker.TypeFlagsObject != 0 && len(l.checker.GetIndexInfosOfType(target)) != 0 && !l.checker.IsArrayType(target):
		return "dictionary"
	}

	return ""
}

// Demand uses the same allocations, joined arguments/results and projected
// stores as shape certification. Unknown is never an empty proof of safety.
func (l *lowering) checkLazyViewReads() error {
	l.certifyUntaggedCallableProducers()
	l.completeUntaggedRecursiveContracts()
	program := l.result
	if len(program.ViewOrigins) == 0 {
		return unemittableCheckedViewRead(program)
	}
	assignAllocationSites(program)
	graph := newAllocationFlowGraph(program)
	viewed := map[int]bool{}
	unknown := false
	queue := []int{}
	add := func(set ir.AllocationSet) {
		unknown = unknown || set.Unknown
		for _, site := range set.Sites {
			if !viewed[site] {
				viewed[site] = true
				queue = append(queue, site)
			}
		}
	}
	for _, origin := range program.ViewOrigins {
		add(graph.ReachingAllocations(origin))
	}
	index := graph.projectionIndex()
	for len(queue) != 0 {
		site := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if literal, ok := index.records[site]; ok {
			for _, field := range literal.Fields {
				if viewAggregate(field.Value) {
					add(graph.ReachingAllocations(field.Value))
				}
			}
			if literal.Spread != nil {
				add(graph.ReachingAllocations(literal.Spread))
			}
		}
		if literal, ok := index.arrays[site]; ok {
			for _, element := range literal.Elements {
				if viewAggregate(element) {
					add(graph.ReachingAllocations(element))
				}
			}
		}
		for _, values := range index.stores[site] {
			for _, value := range values {
				if viewAggregate(value) {
					add(graph.ReachingAllocations(value))
				}
			}
		}
	}
	// Wider interfaces and instantiated helpers can give the same member a
	// different checker type id. Field-name fallback retains the refusal until
	// the receiver flow proves it cannot receive a viewed allocation.
	unsupportedFields := map[string]string{}
	for _, descriptor := range program.ViewContracts {
		for _, field := range descriptor.Fields {
			if family := program.ViewContracts[field.Contract-1].Unsupported; family != "" {
				unsupportedFields[field.Name] = family
			}
		}
	}
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		var receiver ir.Expression
		var typeID, receiverTypeID int
		var field, where string

		switch read := node.(type) {
		case ir.Property:
			if !program.CheckedField(read.ViewReceiverTypeID, read.Name) {
				return true
			}
			if err := l.callableProducerViewRefusal(graph, read); err != nil {
				refused = err
				return false
			}
			receiver, typeID, field, where = read.Object, read.ViewTypeID, read.Name, read.ViewWhere
			receiverTypeID = read.ViewReceiverTypeID

		default:
			return true
		}
		contract := program.ViewContractTypes[typeID]
		family := ""
		if contract != 0 {
			family = program.ViewContracts[contract-1].Unsupported
		}
		if receiverContract := program.ViewContractTypes[receiverTypeID]; family == "" && receiverContract != 0 {
			family = program.ViewContracts[receiverContract-1].Unsupported
		}
		if family == "" {
			family = unsupportedFields[field]
		}
		if family == "" {
			return true
		}
		reaches := graph.ReachingAllocations(receiver)
		demanded := unknown || reaches.Unknown
		for _, site := range reaches.Sites {
			demanded = demanded || viewed[site]
		}
		if demanded {
			if strings.Contains(family, "views-v3: array element kind") {
				refused = &NotYet{Where: where, What: "checked view union arm awaits views-v3: array element kind"}
				return false
			}
			refused = &Refused{Where: where, What: "checked view read of field " + field + " with unsupported " + family + " contract", Fix: "prove or implement the " + family + " contract before reading this field"}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if refused != nil {
		return refused
	}
	return unemittableCheckedViewRead(program)
}

// A checked field read is emitted either through the union adapter, whose contracts are unions,
// fixed tuples and certified callables, or in place as one of the representations the runtime
// checks a slot against. A read that is neither, such as an interface with a length held as a
// union because a string, an array and an object all satisfy it, has no emission, so it stops at
// the read. A method read is the call's to emit, through the adapter.
//
// An optional field the shape lacks reads undefined, so it goes through the adapter, and only when
// its declared union holds undefined. A read through ?. reads undefined from an absent receiver
// and checks a present one as any other read, through the adapter or in place. Neither carries a
// callable's recorded signature.
func unemittableCheckedViewRead(program *ir.Program) error {
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		read, isProperty := node.(ir.Property)
		if !isProperty || read.View == "" || read.Method || !program.CheckedField(read.ViewReceiverTypeID, read.Name) {
			return true
		}
		contract := emittedViewContract(program, read)
		if contract > 0 && int(contract) <= len(program.ViewContracts) {
			descriptor := program.ViewContracts[contract-1]
			if descriptor.Kind == ir.ViewUnion || descriptor.FixedTuple || descriptor.Kind == ir.ViewCallable && descriptor.ProducerCertified {
				if (read.Optional || read.Absent) && (descriptor.Of == ir.Closure || read.Of == ir.Closure) {
					refused = &NotYet{Where: read.ViewWhere, What: "a checked view read of optional field " + read.Name + " of callable type " + read.ViewType}
					return false
				}
				if read.Absent && !descriptor.Undefined {
					refused = &NotYet{Where: read.ViewWhere, What: "a checked view read of optional field " + read.Name + " whose type " + read.ViewType + " doesn't hold undefined"}
					return false
				}
				return true
			}
		}
		if read.Absent {
			refused = &NotYet{Where: read.ViewWhere, What: "a checked view read of optional field " + read.Name + " of type " + read.ViewType + " outside a union view contract"}
			return false
		}
		switch read.Of {
		case ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map:
			return true
		}
		refused = &NotYet{Where: read.ViewWhere, What: "a checked view read of field " + read.Name + " of type " + read.ViewType + ", which is a " + typeName(read.Of) + " outside a union view contract"}
		return false
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return refused
}

func viewAggregate(value ir.Expression) bool {
	if value == nil {
		return false
	}
	switch value.Type() {
	case ir.Object, ir.Array, ir.Map, ir.Union, ir.Closure:
		return true
	}
	return false
}
