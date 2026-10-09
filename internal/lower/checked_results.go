package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Narrower overloads check the fields not proved by the implementation result.
// This checks the returned values and retains their original allocation contracts.
func (l *lowering) checkedOverloadFields(node *ast.Node, produced, promised *checker.Type) ([]ir.ContractField, bool) {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return nil, false
	}
	produced, promised = l.concrete(produced), l.concrete(promised)
	if !l.structured(produced) || !l.structured(promised) || l.includesUndefined(produced) {
		return nil, false
	}
	fields := []ir.ContractField{}
	for _, target := range l.checker.GetPropertiesOfType(promised) {
		source := l.checker.GetPropertyOfType(produced, target.Name)
		if source == nil || accessorSymbol(source) || accessorSymbol(target) || source.Flags&ast.SymbolFlagsMethod != 0 || target.Flags&ast.SymbolFlagsMethod != 0 {
			return nil, false
		}
		own, expected := l.checker.GetTypeOfSymbol(source), l.checker.GetTypeOfSymbol(target)
		if l.censusRelated(own, expected) {
			continue
		}
		contract := l.fieldContract(expected)
		kind, known := l.representation(own)
		if contract == nil || !known || kind != contract.Kind || !interfaceScalar(l.checker.GetNonNullableType(expected)) && !contract.Structural && !contract.Reference {
			return nil, false
		}
		fields = append(fields, ir.ContractField{Name: target.Name, Contract: contract})
	}
	return fields, len(fields) != 0
}

func (l *lowering) checkedOverloadResult(call *ast.CallExpression, value ir.Expression, produced, promised *checker.Type) (ir.Expression, bool) {
	fields, known := l.checkedOverloadFields(call.AsNode(), produced, promised)
	if !known {
		return value, false
	}
	if l.result.CheckedWrites == nil {
		l.result.CheckedWrites = map[string]bool{}
	}
	for _, field := range fields {
		l.result.CheckedWrites[field.Name] = true
		l.result.WriteChecks = append(l.result.WriteChecks, ir.WriteCheck{Where: l.program.Where(call.AsNode()), Expression: sourceExpression(call.AsNode()) + "." + field.Name})
	}
	return ir.ContractResult{Value: value, Fields: fields, Expression: sourceExpression(call.AsNode())}, true
}

// One closure ABI serves these overloads: argument counts can vary, but every
// incoming slot and result uses the implementation's representation. Broader
// boxed unions and scalar overload adapters retain their existing diagnostics.
func (l *lowering) checkedOverloadValue(node, overload *ast.Node) bool {
	implementation := l.censusImplementation(overload)
	if implementation == nil || len(implementation.TypeParameters()) != 0 {
		return false
	}
	actual := l.checker.GetSignatureFromDeclaration(implementation)
	actualResult, known := l.representation(l.checker.GetReturnTypeOfSignature(actual))
	if !known || actualResult != ir.Object {
		return false
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
	if len(signatures) < 2 {
		return false
	}
	for _, signature := range signatures {
		result, known := l.representation(l.checker.GetReturnTypeOfSignature(signature))
		if !known || result != actualResult || len(signature.TypeParameters()) != 0 {
			return false
		}
		for index, parameter := range signature.Parameters() {
			if index >= len(actual.Parameters()) {
				return false
			}
			own, known := l.representation(l.censusCallableParameterType(actual.Parameters()[index]))
			view, viewed := l.representation(l.censusCallableParameterType(parameter))
			if !known || !viewed || own != view || censusCallableSlotless(own) {
				return false
			}
		}
	}
	return true
}
