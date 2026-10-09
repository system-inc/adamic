package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// These descriptors come from declarations, independently of a checked view.
// Unsupported or recursive domains remain unknown, rather than a kind-only proof.
func (l *lowering) viewCallableValueContract(node *ast.Node, target *checker.Type) ir.ViewContractID {
	return l.viewCallableValueDomain(node, target, map[*checker.Type]bool{})
}

func (l *lowering) viewCallableValueDomain(node *ast.Node, target *checker.Type, active map[*checker.Type]bool) ir.ViewContractID {
	if active[target] {
		return 0
	}
	active[target] = true
	defer delete(active, target)
	c := ir.ViewContract{Name: l.checker.TypeToString(target)}
	flags := target.Flags()
	switch {
	case flags&(checker.TypeFlagsUnknown|checker.TypeFlagsNonPrimitive) != 0:
		c.Kind, c.Of = ir.ViewUnknown, ir.Union
	case flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0:
		c.Kind, c.Undefined = ir.ViewUndefined, true
	case flags&checker.TypeFlagsNull != 0:
		c.Kind = ir.ViewNull
	case flags&checker.TypeFlagsUnion != 0:
		c.Kind = ir.ViewUnion
		for _, child := range target.Types() {
			id := l.viewCallableValueDomain(node, child, active)
			if id == 0 {
				return 0
			}
			c.Members = append(c.Members, id)
		}
	default:
		of, known := l.representation(target)
		if !known || l.unsupportedViewFamily(target) != "" || l.callableViewContract(target) {
			return 0
		}
		c.Of = of
		switch of {
		case ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
			c.Kind, c.Allowed, c.Undefined = ir.ViewScalar, l.viewContractLiterals(target), l.includesUndefined(target)
		case ir.Object:
			c.Kind = ir.ViewObject
			for _, field := range l.checker.GetPropertiesOfType(target) {
				id := l.viewCallableValueDomain(node, l.checker.GetTypeOfSymbol(field), active)
				if id == 0 {
					return 0
				}
				c.Fields = append(c.Fields, ir.ViewFieldContract{Name: field.Name, Contract: id, Optional: field.Flags&ast.SymbolFlagsOptional != 0})
			}
		case ir.Array:
			element := l.viewArrayElementType(target)
			if element == nil {
				return 0
			}
			c.Kind, c.Element = ir.ViewArray, l.viewCallableValueDomain(node, element, active)
			if c.Element == 0 {
				return 0
			}
		default:
			return 0
		}
	}
	l.result.ViewContracts = append(l.result.ViewContracts, c)
	return ir.ViewContractID(len(l.result.ViewContracts))
}

func (l *lowering) viewCallableCallContract(node *ast.Node, signatures []*checker.Signature) ir.ViewContractID {
	if len(signatures) == 0 {
		signatures = l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression)), checker.SignatureKindCall)
	}
	return l.viewCallableInvocationContract(node, l.checker.GetTypeAtLocation(node.AsCallExpression().Expression), signatures)
}

func (l *lowering) viewCallableInvocationContract(node *ast.Node, target *checker.Type, signatures []*checker.Signature) ir.ViewContractID {
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 || signatures[0].HasRestParameter() {
		return 0
	}
	s := signatures[0]
	c := ir.ViewContract{Kind: ir.ViewCallable, Name: l.checker.TypeToString(target)}
	for _, p := range s.Parameters() {
		id := l.viewCallableValueContract(node, l.checker.GetTypeOfSymbol(p))
		if id == 0 {
			return 0
		}
		c.Parameters = append(c.Parameters, id)
	}
	c.Result = l.viewCallableValueContract(node, l.checker.GetReturnTypeOfSignature(s))
	if c.Result == 0 {
		return 0
	}
	l.result.ViewContracts = append(l.result.ViewContracts, c)
	return ir.ViewContractID(len(l.result.ViewContracts))
}

// Known producers whose parameter domains have no runtime witness cannot be
// admitted merely because another function has a matching shape or field name.
func (graph *allocationFlowGraph) viewCallableUncheckableProducer(property ir.Property, reaches ir.AllocationSet) bool {
	program := graph.program
	index := graph.projectionIndex()
	seen := map[int]bool{}
	uncheckable := func(function int) bool {
		f := program.Functions[function]
		offset := 0
		if f.Receiver {
			offset = 1
		}
		if f.RestElement != 0 || len(f.Parameters)-offset != len(f.CallableParameters) || f.CallableResultName == "" {
			return true
		}
		for _, id := range f.CallableParameters {
			if id == 0 {
				return true
			}
		}
		return false
	}
	var inspect func(ir.Expression) bool
	inspect = func(value ir.Expression) bool {
		switch value := value.(type) {
		case ir.MakeClosure:
			return uncheckable(value.Function)
		case ir.Read:
			node := value.Local + 1
			if seen[node] {
				return false
			}
			seen[node] = true
			for _, source := range graph.sources[node] {
				if inspect(source) {
					return true
				}
			}
		case ir.Conditional:
			return inspect(value.WhenTrue) || inspect(value.WhenNot)
		case ir.Call:
			for _, function := range program.CallTargets(value) {
				node := graph.resultNode(function)
				if seen[node] {
					continue
				}
				seen[node] = true
				for _, source := range graph.sources[node] {
					if inspect(source) {
						return true
					}
				}
			}
		}
		return false
	}
	for _, site := range reaches.Sites {
		literal := index.records[site]
		for _, field := range literal.Fields {
			if field.Name == property.Name && inspect(field.Value) {
				return true
			}
		}
		for _, method := range literal.Methods {
			if method.Name == property.Name && uncheckable(method.Function) {
				return true
			}
		}
		for _, value := range index.stores[site][property.Name] {
			if inspect(value) {
				return true
			}
		}
	}
	return false
}
