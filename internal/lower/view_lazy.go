package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Unsupported contracts are descriptors, never certificates. View admission
// rejects them before a partial view can be formed.
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
	if l.checker.IsArrayType(target) {
		family = "array"
	}
	if of, known := l.representation(target); target.Flags()&checker.TypeFlagsUnion != 0 && (!interfaceScalar(target) || !known || of == ir.Union) {
		family = "union"
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
		return nil
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
			if !program.CheckedFields[read.Name] {
				return true
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

func (l *lowering) lazyReadRefusal(node *ast.Node, field, family string) error {
	return &Refused{Where: l.program.Where(node), What: "checked view read of field " + field + " with unsupported " + family + " contract", Fix: "prove or implement the " + family + " contract before reading this field"}
}

// Every member must have a complete checked lowering, even if it is never read.
func (l *lowering) checkViewMembers(node *ast.Node, target *checker.Type) error {
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if err := l.checkViewMembers(node, member); err != nil {
				return err
			}
		}
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		declared := l.checker.GetTypeOfSymbol(property)
		id, err := l.viewContract(node, declared)
		if err != nil || l.unsupportedViewContract(id, map[ir.ViewContractID]bool{}) {
			return &Refused{Where: l.program.Where(node), What: "view type has an unsupported member: " + property.Name, Fix: "use a supported, checked type for " + property.Name + " (currently " + l.checker.TypeToString(declared) + ") before creating this view"}
		}
	}
	return nil
}

func (l *lowering) unsupportedViewContract(id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id == 0 || int(id) > len(l.result.ViewContracts) {
		return true
	}
	if seen[id] {
		return false
	}
	seen[id] = true
	contract := l.result.ViewContracts[id-1]
	if contract.Unsupported != "" || contract.Kind == ir.ViewUnknown {
		return true
	}
	for _, field := range contract.Fields {
		if l.unsupportedViewContract(field.Contract, seen) {
			return true
		}
	}
	for _, member := range contract.Members {
		if l.unsupportedViewContract(member, seen) {
			return true
		}
	}
	for _, element := range contract.Tuple {
		if l.unsupportedViewContract(element, seen) {
			return true
		}
	}
	for _, parameter := range contract.Parameters {
		if l.unsupportedViewContract(parameter, seen) {
			return true
		}
	}
	for _, child := range []ir.ViewContractID{contract.Element, contract.Key, contract.Result, contract.Payload, contract.TupleRest, contract.ObjectPresent} {
		if child != 0 && l.unsupportedViewContract(child, seen) {
			return true
		}
	}
	return false
}
