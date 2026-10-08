package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) closedPredicateRefusal(node *ast.Node) error {
	original := l.proveFlowPredicate(node)
	if original == nil {
		return nil
	}
	if !l.predicateHatchContract(node) {
		return original
	}
	function := node.Parent
	if function.Kind != ast.KindFunctionDeclaration || function.Name() == nil {
		return original
	}
	symbol := l.symbol(function.Name())
	valid := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && n != function.Name() && l.symbol(n) == symbol {
			if n.Parent.Kind != ast.KindCallExpression || n.Parent.AsCallExpression().Expression != n || len(n.Parent.AsCallExpression().Arguments.Nodes) != 1 || ast.SkipParentheses(n.Parent.AsCallExpression().Arguments.Nodes[0]).Kind != ast.KindIdentifier {
				if !l.predicateHatchCallbackUse(n) {
					valid = false
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	if valid {
		return nil
	}
	return original
}

// Only a closed union's exact members can be certified by its discriminant.
// An open structural Node, an optional field refinement, or a callable contract
// cannot be established by checking kind alone.
func (l *lowering) predicateTagContract(source, target *checker.Type) (castProof, bool) {
	if source.Flags()&checker.TypeFlagsUnion == 0 || !l.castUnionWrites(source) {
		return castProof{}, false
	}
	members, targets := castMembers(source), castMembers(target)
	for _, wanted := range targets {
		found := false
		for _, member := range members {
			if checker.Checker_isTypeIdenticalTo(l.checker, member, wanted) {
				found = true
			}
		}
		if !found {
			return castProof{}, false
		}
	}
	proof := castProof{field: "kind"}
	seen := map[string]bool{}
	for _, member := range members {
		property := l.checker.GetPropertyOfType(member, "kind")
		literal := l.fieldLiteral(member, "kind")
		if property == nil || !l.checker.IsReadonlySymbol(property) || literal == nil {
			return castProof{}, false
		}
		key := l.checker.TypeToString(literal)
		if literal.Flags()&checker.TypeFlagsBooleanLiteral == 0 {
			key = fmt.Sprintf("%T:%v", literal.AsLiteralType().Value(), literal.AsLiteralType().Value())
		}
		if seen[key] {
			return castProof{}, false
		}
		seen[key] = true
		for _, wanted := range targets {
			if checker.Checker_isTypeIdenticalTo(l.checker, member, wanted) {
				proof.allowed = append(proof.allowed, literal)
			}
		}
	}
	return proof, len(proof.allowed) > 0
}

func (l *lowering) predicateCheckedRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier || value.Type() != ir.Object || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return value, nil
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return value, nil
	}
	source, target := l.checker.GetTypeOfSymbol(symbol), l.checker.GetTypeAtLocation(node)
	if checker.Checker_isTypeIdenticalTo(l.checker, source, target) {
		return value, nil
	}
	proof, valid := l.predicateTagContract(source, target)
	if !valid {
		return value, nil
	}
	cast := ir.CheckedCast{Value: value, Field: proof.field, Message: "predicate narrowing failed: expected " + l.checker.TypeToString(target)}
	for _, literal := range proof.allowed {
		allowed, of, constant := l.literalConstant(literal)
		if !constant || (cast.FieldType != 0 && cast.FieldType != of) {
			return nil, l.notYet(node, "a predicate discriminant without a literal representation")
		}
		cast.Allowed = append(cast.Allowed, allowed)
		cast.FieldType = of
	}
	return cast, nil
}

// A callback signature is admitted only on a stable parameter of a direct named
// function whose every use is a call with an independently proven named guard.
// Escapes, aliases, overloads and reassigned parameters stay refused.
// A .ts closed union may instead check the consumed narrowing at runtime.
func (l *lowering) predicateCallbackContract(node *ast.Node) (bool, error) {
	signature := node.Parent
	if signature.Parent == nil || signature.Parent.Kind != ast.KindParameter {
		return false, nil
	}
	parameter := signature.Parent
	function := parameter.Parent
	if function.Kind != ast.KindFunctionDeclaration || function.Body() == nil || function.Name() == nil || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return false, nil
	}
	annotation := node.AsTypePredicateNode()
	if annotation.Type == nil || annotation.AssertsModifier != nil {
		return false, nil
	}
	index := -1
	for i, candidate := range function.Parameters() {
		if candidate == parameter {
			index = i
		}
	}
	if index < 0 {
		return false, nil
	}
	failure := (&predicateFlowProof{l: l}).refused(node, "callback arguments do not all have independent body proofs")
	parameterSymbol, functionSymbol := l.symbol(parameter.Name()), l.symbol(function.Name())
	valid := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier && l.symbol(n) == parameterSymbol && ast.IsAssignmentTarget(n) {
			valid = false
		}
		if n.Kind == ast.KindIdentifier && n != function.Name() && l.symbol(n) == functionSymbol {
			if n.Parent.Kind != ast.KindCallExpression || n.Parent.AsCallExpression().Expression != n {
				valid = false
			} else {
				arguments := n.Parent.AsCallExpression().Arguments.Nodes
				if index >= len(arguments) {
					valid = false
				} else {
					argument := ast.SkipParentheses(arguments[index])
					var guard *ast.Node
					if argument.Kind == ast.KindIdentifier {
						if symbol := l.symbol(argument); symbol != nil {
							for _, declaration := range symbol.Declarations {
								if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
									guard = declaration
								}
							}
						}
					}
					if guard == nil || guard.Type() == nil || guard.Type().Kind != ast.KindTypePredicate || guard.Type().AsTypePredicateNode().AssertsModifier != nil || guard.Type().AsTypePredicateNode().Type == nil {
						valid = false
					} else if !checker.Checker_isTypeIdenticalTo(l.checker, l.checker.GetTypeFromTypeNode(annotation.Type), l.checker.GetTypeFromTypeNode(guard.Type().AsTypePredicateNode().Type)) || (l.proveFlowPredicate(guard.Type()) != nil && !(l.predicateHatchContract(node) && l.predicateHatchContract(guard.Type()))) {
						valid = false
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(visit)
	}
	if !valid {
		return true, failure
	}
	return true, nil
}

// Only exact closed members have a complete runtime contract on this branch.
func (l *lowering) predicateHatchContract(node *ast.Node) bool {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return false
	}
	annotation := node.AsTypePredicateNode()
	function := node.Parent
	if annotation.AssertsModifier != nil || annotation.Type == nil || !ast.IsFunctionLike(function) || len(function.Parameters()) != 1 {
		return false
	}
	parameter := function.Parameters()[0]
	if parameter.Name().Kind != ast.KindIdentifier || parameter.Name().Text() != annotation.ParameterName.Text() || parameter.AsParameterDeclaration().Initializer != nil || parameter.AsParameterDeclaration().DotDotDotToken != nil {
		return false
	}
	if function.Kind != ast.KindFunctionType && (function.Kind != ast.KindFunctionDeclaration || function.Name() == nil || function.Body() == nil) {
		return false
	}
	_, valid := l.predicateTagContract(l.checker.GetTypeAtLocation(parameter.Name()), l.checker.GetTypeFromTypeNode(annotation.Type))
	return valid
}

// A guard may escape only to a callback signature that consumes the same closed
// contract in a .ts body, where narrowed reads receive the runtime check.
func (l *lowering) predicateHatchCallbackUse(node *ast.Node) bool {
	if node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := node.Parent.AsCallExpression()
	if call.Expression.Kind != ast.KindIdentifier {
		return false
	}
	symbol := l.symbol(call.Expression)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(declaration)), ".ts") {
			continue
		}
		for index, argument := range call.Arguments.Nodes {
			if argument != node || index >= len(declaration.Parameters()) {
				continue
			}
			parameter := declaration.Parameters()[index]
			signature := parameter.Type()
			if signature != nil && signature.Kind == ast.KindFunctionType && signature.Type() != nil && signature.Type().Kind == ast.KindTypePredicate && l.predicateHatchContract(signature.Type()) {
				return true
			}
		}
	}
	return false
}
