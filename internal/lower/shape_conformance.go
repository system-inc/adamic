package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// certifyAllocationFields records semantic declarations before physical shape
// layout loses distinctions such as number versus boolean. Missing declarations
// stay uncertified. Readiness and mutation effects require separate proofs.
func (l *lowering) certifyAllocationFields(node *ast.Node, literal ir.ObjectLiteral) ir.ObjectLiteral {
	declared := l.checker.GetTypeAtLocation(node)
	if declared == nil {
		return literal
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	for i := range literal.Fields {
		field := &literal.Fields[i]
		if field.Private {
			continue
		}
		properties := l.checker.GetPropertiesOfType(declared)
		// A staged initializer has type never, while its allocated slot retains
		// the contextual declaration (for example ready: boolean). Prefer that
		// declaration where present; extra literal fields keep their own type.
		if ast.IsObjectLiteralExpression(node) && contextual != nil {
			for _, property := range l.checker.GetPropertiesOfType(contextual) {
				if property.Name == field.Name {
					properties = []*ast.Symbol{property}
					break
				}
			}
		}
		for _, property := range properties {
			if property.Name != field.Name {
				continue
			}
			proven := l.checker.GetTypeOfSymbol(property)
			if proven == nil || proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
				break
			}
			field.Certificate = &ir.FieldTypeCertificate{
				CheckerType: int(proven.Id()), DeclaredType: l.checker.TypeToString(proven),
				Optional: property.Flags&ast.SymbolFlagsOptional != 0,
			}
			break
		}
	}
	return literal
}

// eraseProvenViewChecks handles closed, immutable scalar records. Broader shape,
// effect and readiness facts remain checked until their proofs are implemented.
func eraseProvenViewChecks(program *ir.Program) {
	if len(program.CheckedFields) == 0 {
		return
	}
	assignAllocationSites(program)
	flow := newAllocationFlowGraph(program)
	allocations := map[int]ir.ObjectLiteral{}
	mutable := false
	collect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if len(value.GraphTypes) > 0 {
				allocations[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case ir.SetProperty, ir.SetIndex, ir.ObjectCall, ir.CallClosure:
			// Until effects through aliases and opaque callbacks are certified, any
			// such operation blocks this deliberately narrow eraser for the program.
			mutable = true
		}
		return true
	}
	walk(program.Main, collect)
	for _, function := range program.Functions {
		walk(function.Body, collect)
	}
	if mutable {
		return
	}
	fieldProof := func(set ir.AllocationSet, name string, contract ir.ViewContractID) bool {
		if set.Unknown || len(set.Sites) == 0 || contract <= 0 || int(contract) > len(program.ViewContracts) {
			return false
		}
		if program.ViewContracts[contract-1].Kind != ir.ViewScalar {
			return false
		}
		for _, site := range set.Sites {
			literal, exists := allocations[site]
			if !exists || literal.Spread != nil || literal.Class != 0 {
				return false
			}
			found := false
			for _, field := range literal.Fields {
				if field.Name != name {
					continue
				}
				certificate := field.Certificate
				if field.Uninitialized || certificate == nil || certificate.Optional || program.ViewContractTypes[certificate.CheckerType] != contract {
					return false
				}
				found = true
			}
			if !found {
				return false
			}
		}
		return true
	}
	rewrite := func(node any) any {
		switch value := node.(type) {
		case ir.Property:
			if value.View != "" && value.Readiness == "" && !value.Optional && !value.Absent && !value.Method && fieldProof(flow.ReachingAllocations(value.Object), value.Name, value.ViewContract) {
				value.View = ""
				value.ViewType = ""
				value.ViewAllowed = nil
				value.ViewContract = 0
				value.ViewTypeID = 0
			}
			return value
		case ir.CheckedCast:
			contract := value.ViewContract
			if contract <= 0 || int(contract) > len(program.ViewContracts) {
				return value
			}
			target := program.ViewContracts[contract-1]
			if target.Kind != ir.ViewObject || len(target.Fields) == 0 {
				return value
			}
			set := flow.ReachingAllocations(value.Value)
			for _, field := range target.Fields {
				if !fieldProof(set, field.Name, field.Contract) {
					return value
				}
			}
			return value.Value
		}
		return node
	}
	program.Main = rewriteShapeStatements(program.Main, rewrite)
	for i := range program.Functions {
		program.Functions[i].Body = rewriteShapeStatements(program.Functions[i].Body, rewrite)
	}
}

// certifiedCheckedCast retains the complete target contract on the admission
// node, so the eraser cannot mistake a tag proof for a whole-object proof.
func (l *lowering) certifiedCheckedCast(node *ast.Node, value ir.CheckedCast, target *checker.Type) (ir.Expression, error) {
	contract, err := l.viewContract(node, target)
	if err != nil {
		return nil, err
	}
	value.ViewContract = contract
	return value, nil
}
