package lower

import (
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// An immediate call uses its call contract. Parentheses and type-only wrappers
// do not make its callee escape; assignment, return and argument passing do.
func (l *lowering) prepareViewCallableEscape(node *ast.Node, declared *checker.Type, property *ir.Property) {
	outer := node
	for outer.Parent != nil {
		switch outer.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
			outer = outer.Parent
		default:
			goto classified
		}
	}
classified:
	if outer.Parent != nil && outer.Parent.Kind == ast.KindCallExpression && outer.Parent.AsCallExpression().Expression == outer {
		return
	}
	property.ViewEscape = true
	property.ViewEscapeAdamic = filepath.Ext(l.program.FileName(ast.GetSourceFileOfNode(node))) == ".a"
	target := l.checker.GetNonNullableType(declared)
	property.ViewEscapeContract = l.viewCallableInvocationContract(node, target, l.checker.GetSignaturesOfType(target, checker.SignatureKindCall))
}

func checkViewCallableEscape(graph *allocationFlowGraph, read ir.Property, reaches ir.AllocationSet) error {
	if read.ViewEscapeAdamic {
		if graph.viewCallableEscapeProven(read, reaches) {
			return nil
		}
		return &Refused{Where: read.ViewWhere, What: "an unproven escaping callable read of " + read.View, Fix: "prove the producer parameter and result relation before this read, or call the member directly through the view"}
	}
	if read.ViewEscapeContract == 0 || graph.viewCallableUncheckableProducer(read, reaches) {
		return &Refused{Where: read.ViewWhere, What: "an escaping callable read of " + read.View + " with an unsupported value contract", Fix: "prove the callable relation or use a callable with runtime-checkable parameter and result types"}
	}
	return nil
}

// Unknown, missing fields and opaque callable sources are not empty proofs.
// Every possible producer in the same may-flow graph must establish the relation.
func (graph *allocationFlowGraph) viewCallableEscapeProven(read ir.Property, reaches ir.AllocationSet) bool {
	if read.ViewEscapeContract == 0 || reaches.Unknown || len(reaches.Sites) == 0 {
		return false
	}
	program := graph.program
	target := program.ViewContracts[read.ViewEscapeContract-1]
	compatible := func(function int) bool {
		producer := program.Functions[function]
		if producer.RestElement != 0 || len(producer.CallableParameters) != len(target.Parameters) {
			return false
		}
		offset := 0
		if producer.Receiver {
			offset = 1
		}
		if len(producer.Parameters)-offset != len(target.Parameters) {
			return false
		}
		for i, parameter := range target.Parameters {
			// Until adapters exist, a proven escape must also preserve the raw ABI.
			if program.ViewContracts[parameter-1].Of != program.Locals[producer.Parameters[i+offset]].Type {
				return false
			}
			if !viewCallableDomainSubset(program, parameter, producer.CallableParameters[i]) {
				return false
			}
		}
		result := program.ViewContracts[target.Result-1]
		if result.Kind != ir.ViewUndefined && result.Of != producer.Returns {
			return false
		}
		return viewCallableDomainSubset(program, producer.CallableResult, target.Result)
	}
	active := map[int]bool{}
	var inspect func(ir.Expression) bool
	var sources func(int) bool
	sources = func(node int) bool {
		if active[node] || len(graph.sources[node]) == 0 {
			return false
		}
		active[node] = true
		defer delete(active, node)
		for _, value := range graph.sources[node] {
			if !inspect(value) {
				return false
			}
		}
		return true
	}
	inspect = func(value ir.Expression) bool {
		switch value := value.(type) {
		case ir.MakeClosure:
			return compatible(value.Function)
		case ir.Read:
			return sources(value.Local + 1)
		case ir.Conditional:
			return inspect(value.WhenTrue) && inspect(value.WhenNot)
		case ir.Call:
			targets := program.CallTargets(value)
			if len(targets) == 0 {
				return false
			}
			for _, function := range targets {
				if !sources(graph.resultNode(function)) {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
	index := graph.projectionIndex()
	for _, site := range reaches.Sites {
		found := false
		literal, exists := index.records[site]
		if !exists {
			return false
		}
		for _, field := range literal.Fields {
			if field.Name == read.Name {
				found = true
				if !inspect(field.Value) {
					return false
				}
			}
		}
		for _, method := range literal.Methods {
			if method.Name == read.Name {
				found = true
				if !compatible(method.Function) {
					return false
				}
			}
		}
		for _, value := range index.stores[site][read.Name] {
			if !inspect(value) {
				return false
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// This conservative proof uses independently recorded declaration domains,
// never representation masks alone. Unsupported domains remain unproven.
func viewCallableDomainSubset(program *ir.Program, from, to ir.ViewContractID) bool {
	if from == 0 || to == 0 {
		return false
	}
	source, target := program.ViewContracts[from-1], program.ViewContracts[to-1]
	if source.Unsupported != "" || target.Unsupported != "" {
		return false
	}
	if target.Kind == ir.ViewUnknown && target.Of == ir.Union && target.Name == "unknown" {
		return true
	}
	if source.Kind == ir.ViewUnion {
		for _, member := range source.Members {
			if !viewCallableDomainSubset(program, member, to) {
				return false
			}
		}
		return len(source.Members) != 0
	}
	if target.Kind == ir.ViewUnion {
		for _, member := range target.Members {
			if viewCallableDomainSubset(program, from, member) {
				return true
			}
		}
		return false
	}
	if source.Kind != target.Kind || source.Of != target.Of || source.Undefined && !target.Undefined {
		return false
	}
	switch source.Kind {
	case ir.ViewNull, ir.ViewUndefined:
		return true
	case ir.ViewScalar:
		if len(target.Allowed) == 0 {
			return true
		}
		if len(source.Allowed) == 0 {
			return false
		}
		for _, literal := range source.Allowed {
			found := false
			for _, allowed := range target.Allowed {
				found = found || literal == allowed
			}
			if !found {
				return false
			}
		}
		return true
	case ir.ViewObject:
		// Exact fields avoid making a mutable structural variance certificate.
		if len(source.Fields) != len(target.Fields) {
			return false
		}
		for _, field := range target.Fields {
			found := false
			for _, actual := range source.Fields {
				if actual.Name == field.Name && actual.Optional == field.Optional && viewCallableDomainSubset(program, actual.Contract, field.Contract) && viewCallableDomainSubset(program, field.Contract, actual.Contract) {
					found = true
				}
			}
			if !found {
				return false
			}
		}
		return true
	case ir.ViewArray:
		return viewCallableDomainSubset(program, source.Element, target.Element) && viewCallableDomainSubset(program, target.Element, source.Element)
	}
	return false
}
