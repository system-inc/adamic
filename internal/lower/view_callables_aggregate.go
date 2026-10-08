package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

type viewCallablePayloadType struct {
	node   *ast.Node
	proven *checker.Type
}

func (l *lowering) recordViewCallablePayloadType(node *ast.Node, proven *checker.Type) int {
	of, known := l.representation(proven)
	if !known || of != ir.Object && of != ir.Union && of != ir.Array || !l.viewCallableAggregateType(proven) {
		return 0
	}
	if l.viewCallablePayloadTypes == nil {
		l.viewCallablePayloadTypes = map[int]viewCallablePayloadType{}
	}
	id := int(proven.Id())
	l.viewCallablePayloadTypes[id] = viewCallablePayloadType{node, proven}
	return id
}

func (l *lowering) recordViewCallableProducerPayloads(index int, node *ast.Node, signature *checker.Signature) {
	if l.viewCallableProducerPayloads == nil {
		l.viewCallableProducerPayloads = map[int][]viewCallablePayloadType{}
	}
	payloads := []viewCallablePayloadType{}
	for _, parameter := range signature.Parameters() {
		payloads = append(payloads, viewCallablePayloadType{node, l.censusCallableParameterType(parameter)})
	}
	payloads = append(payloads, viewCallablePayloadType{node, l.checker.GetReturnTypeOfSignature(signature)})
	l.viewCallableProducerPayloads[index] = payloads
}

// Registration happens after all bodies: a helper lowered before its caller's
// view must receive the same checks. Preparing descriptors never certifies a
// descendant member or inspects a runtime payload at the cast.
func (l *lowering) prepareViewCallableAggregateSchemas() error {
	program := l.result
	active := false
	for _, contract := range program.ViewContracts {
		for _, field := range contract.Fields {
			if !program.CheckedFields[field.Name] {
				continue
			}
			callable := program.ViewContracts[field.Contract-1]
			if callable.Kind != ir.ViewCallable {
				continue
			}
			children := append(append([]ir.ViewContractID{}, callable.Parameters...), callable.Result)
			for _, child := range children {
				active = active || child != 0 && program.ViewContracts[child-1].PayloadTypeID != 0
			}
		}
	}
	if !active {
		return nil
	}
	roots := []ir.ViewContractID{}
	count := len(program.ViewContracts)
	for i := 0; i < count; i++ {
		payload, exists := l.viewCallablePayloadTypes[program.ViewContracts[i].PayloadTypeID]
		if !exists {
			continue
		}
		id, err := l.viewContract(payload.node, payload.proven)
		if err != nil {
			return err
		}
		program.ViewContracts[i].Payload = id
		roots = append(roots, id)
	}
	// Callback parameters have Unknown may-flow when their invoking edge cannot
	// be resolved. Their declared reads must check too, even at a narrower type.
	for _, payloads := range l.viewCallableProducerPayloads {
		for _, payload := range payloads {
			of, known := l.representation(payload.proven)
			if !known || of != ir.Object && of != ir.Union && of != ir.Array {
				continue
			}
			id, err := l.viewContract(payload.node, payload.proven)
			if err != nil {
				return err
			}
			roots = append(roots, id)
		}
	}
	seen := map[ir.ViewContractID]bool{}
	var visit func(ir.ViewContractID)
	visit = func(id ir.ViewContractID) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		contract := program.ViewContracts[id-1]
		for _, field := range contract.Fields {
			program.CheckedFields[field.Name] = true
			if field.Optional {
				l.optionalViewWriteField(field.Name)
			}
			visit(field.Contract)
		}
		for _, member := range contract.Members {
			visit(member)
		}
		for _, parameter := range contract.Parameters {
			visit(parameter)
		}
		visit(contract.Result)
		visit(contract.Payload)
		visit(contract.Element)
	}
	for _, root := range roots {
		visit(root)
	}
	return nil
}

func viewCallableAggregateResults(program *ir.Program) []ir.CallClosure {
	results := []ir.CallClosure{}
	collect := func(node any) bool {
		if call, ok := node.(ir.CallClosure); ok && viewAggregate(call) {
			results = append(results, call)
		}
		return true
	}
	walk(program.Main, collect)
	for _, function := range program.Functions {
		walk(function.Body, collect)
	}
	return results
}

func (l *lowering) addViewCallableAggregateResults(graph *allocationFlowGraph, calls []ir.CallClosure, viewed map[int]bool, unknown bool, add func(ir.AllocationSet)) {
	for _, call := range calls {
		owner := ir.ClosureOperands(call)[0]
		if property, ok := owner.(ir.Property); ok {
			owner = property.Object
		}
		set := graph.ReachingAllocations(owner)
		demanded := unknown || set.Unknown
		for _, site := range set.Sites {
			demanded = demanded || viewed[site]
		}
		if demanded {
			add(graph.ReachingAllocations(call))
		}
	}
}

func (l *lowering) viewCallableAggregateType(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.viewCallableAggregateType(member) {
				return true
			}
		}
		return false
	}
	of, known := l.representation(proven)
	return known && (of == ir.Object && proven.Flags()&checker.TypeFlagsObject != 0 || of == ir.Array && l.viewArrayElementType(proven) != nil)
}

// A concrete supported read checks its own declared contract. A broader producer
// descriptor with an unknown member of the same name must not poison that check.
func viewCallableConcreteReadContract(program *ir.Program, id ir.ViewContractID) bool {
	if id == 0 {
		return false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return false
	}
	return contract.Kind == ir.ViewScalar && contract.Of != 0 || contract.Kind == ir.ViewCallable && contract.Result != 0
}
