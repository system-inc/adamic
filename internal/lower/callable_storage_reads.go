package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

type callableReceiverProof uint8

const (
	callableUnknown callableReceiverProof = iota
	callableIndependent
	callableReceiverNeeded
)

// A copied field is a free callable, not a bound method. Follow known creation
// sites and getter returns; a structural callback signature alone cannot prove
// that the selected value needs no call-site receiver.
func (l *lowering) detachedCallableStorage(node *ast.Node) error {
	if isCallee(node) || l.libraryMember(node) {
		return nil
	}
	if held, known := l.representation(l.checker.GetTypeAtLocation(node)); !known || held != ir.Closure {
		return nil
	}
	symbol := l.memberSymbol(node)
	if symbol == nil {
		return l.notYet(node, "a detached callable without a known storage origin")
	}
	proof := l.callableDeclarationProof(symbol, map[*ast.Node]bool{}, 0)
	if proof == callableUnknown {
		// Signature-only views need the closed program's actual allocation members.
		view := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
		modules, err := l.moduleOrder(l.program.Files()[0])
		if err != nil {
			return err
		}
		found := false
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if candidate.Kind == ast.KindObjectLiteralExpression || candidate.Kind == ast.KindNewExpression {
				shape := l.checker.GetTypeAtLocation(candidate)
				if field := l.checker.GetPropertyOfType(shape, node.Name().Text()); field != nil && l.iterationShapeFits(shape, view) {
					current := l.callableDeclarationProof(field, map[*ast.Node]bool{}, 0)
					if !found {
						proof = current
						found = true
					} else {
						proof = joinCallableProof(proof, current)
					}
				}
			}
			return candidate.ForEachChild(visit)
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
	}
	if proof == callableReceiverNeeded {
		return &Refused{Where: l.program.Where(node), What: "a receiver-dependent callable read without its object (unbound-method)", Fix: "call it in an arrow that keeps its object"}
	}
	if proof != callableIndependent && l.copiedForOptionalCall(node) {
		return l.notYet(node, "a detached callable whose receiver independence is not proven")
	}
	return nil
}

func joinCallableProof(left, right callableReceiverProof) callableReceiverProof {
	if left == callableReceiverNeeded || right == callableReceiverNeeded {
		return callableReceiverNeeded
	}
	if left == callableUnknown || right == callableUnknown {
		return callableUnknown
	}
	return callableIndependent
}

func (l *lowering) callableDeclarationProof(symbol *ast.Symbol, seen map[*ast.Node]bool, depth int) callableReceiverProof {
	if symbol == nil || depth > 16 {
		return callableUnknown
	}
	proof := callableIndependent
	found := false
	for _, root := range l.checker.GetRootSymbols(symbol) {
		for _, declaration := range root.Declarations {
			if seen[declaration] {
				return callableUnknown
			}
			seen[declaration] = true
			current := callableUnknown
			switch declaration.Kind {
			case ast.KindMethodDeclaration, ast.KindMethodSignature:
				current = callableReceiverNeeded
			case ast.KindPropertyAssignment:
				current = l.callableValueProof(declaration.AsPropertyAssignment().Initializer, seen, depth+1)
			case ast.KindPropertyDeclaration:
				current = l.callableValueProof(declaration.AsPropertyDeclaration().Initializer, seen, depth+1)
			case ast.KindVariableDeclaration:
				if declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
					current = l.callableValueProof(declaration.AsVariableDeclaration().Initializer, seen, depth+1)
				}
			case ast.KindFunctionDeclaration:
				if ast.GetThisParameter(declaration) == nil {
					current = callableIndependent
				} else {
					current = callableReceiverNeeded
				}
			case ast.KindGetAccessor:
				current = l.callableReturnProof(declaration.Body(), seen, depth+1)
			}
			delete(seen, declaration)
			proof = joinCallableProof(proof, current)
			found = true
		}
	}
	if !found {
		return callableUnknown
	}
	return proof
}

func (l *lowering) callableValueProof(node *ast.Node, seen map[*ast.Node]bool, depth int) callableReceiverProof {
	if node == nil {
		return callableIndependent
	}
	if depth > 16 {
		return callableUnknown
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindNullKeyword:
		return callableIndependent
	case ast.KindFunctionExpression:
		if dynamicReceiverFunction(node) {
			return callableReceiverNeeded
		}
		return callableIndependent
	case ast.KindIdentifier:
		if node.Text() == "undefined" && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsUndefined != 0 {
			return callableIndependent
		}
		return l.callableDeclarationProof(l.symbol(node), seen, depth+1)
	case ast.KindPropertyAccessExpression:
		return l.callableDeclarationProof(l.memberSymbol(node), seen, depth+1)
	case ast.KindConditionalExpression:
		expression := node.AsConditionalExpression()
		return joinCallableProof(l.callableValueProof(expression.WhenTrue, seen, depth+1), l.callableValueProof(expression.WhenFalse, seen, depth+1))
	}
	return callableUnknown
}

func (l *lowering) callableReturnProof(body *ast.Node, seen map[*ast.Node]bool, depth int) callableReceiverProof {
	if body == nil || depth > 16 {
		return callableUnknown
	}
	proof := callableIndependent
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			proof = joinCallableProof(proof, l.callableValueProof(node.AsReturnStatement().Expression, seen, depth+1))
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return proof
}

// The new origin requirement applies to copies reaching optional invocation.
// Other observations and established ordinary-call paths keep their existing
// checks. Follow local copy aliases, without interpreting truthiness as a call.
func (l *lowering) copiedForOptionalCall(node *ast.Node) bool {
	source, parent := node, node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		source, parent = parent, parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindVariableDeclaration || parent.AsVariableDeclaration().Initializer != source {
		return false
	}
	initial := l.symbol(parent.Name())
	if initial == nil {
		return false
	}
	aliases := map[*ast.Symbol]bool{initial: true}
	file := ast.GetSourceFileOfNode(node)
	changed, found := true, false
	for changed && !found {
		changed = false
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if candidate.Kind == ast.KindIdentifier && aliases[l.symbol(candidate)] {
				use := candidate.Parent
				if use != nil && use.Kind == ast.KindCallExpression && use.AsCallExpression().Expression == candidate && use.AsCallExpression().QuestionDotToken != nil {
					found = true
					return true
				}
				if use != nil && use.Kind == ast.KindVariableDeclaration && use.AsVariableDeclaration().Initializer == candidate {
					symbol := l.symbol(use.Name())
					if symbol != nil && !aliases[symbol] {
						aliases[symbol] = true
						changed = true
					}
				}
			}
			return candidate.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	return found
}
