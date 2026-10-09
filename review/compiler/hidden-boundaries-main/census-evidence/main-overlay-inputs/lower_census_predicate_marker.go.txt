package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Like the instantiation bridge, use the pinned checker through its own relation
// metadata rather than guessing that a boolean-returning function is a predicate.
//
//go:linkname censusSignaturePredicate github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getTypePredicateOfSignature
func censusSignaturePredicate(*checker.Checker, *checker.Signature) *checker.TypePredicate

func (l *lowering) censusDiscardedMarkerPredicate(source, target *checker.Signature) bool {
	return l.censusNeverRestSignature(target) &&
		l.checker.GetReturnTypeOfSignature(target).Flags()&checker.TypeFlagsVoid != 0 &&
		censusSignaturePredicate(l.checker, source) != nil
}

// A callback contract has no producer body to prove. For this narrowly erased
// use, its predicate cannot be observed: every reference passes the callback
// into a never-rest marker slot. Ordinary calls, escapes and writes keep the
// existing predicate refusal. Actual producer bodies still require their proof.
func (l *lowering) censusPredicateMarkerContract(annotation *ast.Node) bool {
	functionType := annotation.Parent
	if functionType == nil || functionType.Kind != ast.KindFunctionType {
		return false
	}
	parameter := functionType.Parent
	if parameter == nil || parameter.Kind != ast.KindParameter || parameter.Type() != functionType || !ast.IsIdentifier(parameter.Name()) {
		return false
	}
	function := parameter.Parent
	if function == nil || !ast.IsFunctionLike(function) || function.Body() == nil {
		return false
	}
	symbol := l.symbol(parameter.Name())
	if symbol == nil {
		return false
	}
	found, safe := false, true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && !ast.IsPartOfTypeNode(node) && l.symbol(node) == symbol {
			found = true
			argument := node
			for argument.Parent != nil && argument.Parent.Kind == ast.KindParenthesizedExpression {
				argument = argument.Parent
			}
			call := argument.Parent
			accepted := false
			if call != nil && call.Kind == ast.KindCallExpression {
				resolved := l.checker.GetResolvedSignature(call)
				if resolved != nil {
					for index, given := range call.AsCallExpression().Arguments.Nodes {
						if given == argument && index < len(resolved.Parameters()) {
							signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(resolved.Parameters()[index]), checker.SignatureKindCall)
							accepted = len(signatures) == 1 && l.censusNeverRestSignature(signatures[0])
						}
					}
				}
			}
			safe = safe && accepted
		}
		node.ForEachChild(visit)
		return false
	}
	function.Body().ForEachChild(visit)
	return found && safe
}
