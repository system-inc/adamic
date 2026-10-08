package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Ask the pinned checker to infer exactly the signature it relates to the concrete
// slot. This is a reference to its shim dependency, not a second inference engine.
//
//go:linkname instantiateSignatureInContextOf github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).instantiateSignatureInContextOf
func instantiateSignatureInContextOf(receiver *checker.Checker, signature *checker.Signature, contextual *checker.Signature, inference *checker.InferenceContext, compare checker.TypeComparer) *checker.Signature

func (l *lowering) genericFunctionValue(node *ast.Node, declaration *ast.Node) (ir.Expression, error) {
	resolved := l.contextualGenericSignature(node, l.checker.GetSignatureFromDeclaration(declaration))
	if resolved == nil {
		return nil, l.notYet(node, "a generic function value without a concrete contextual signature")
	}
	index, err := l.instantiateFunctionSignature(node, declaration, resolved)
	if err != nil {
		return nil, err
	}
	return l.functionValue(node, index)
}

// Inspect source rather than already-created forwarders: module function bodies
// are lowered before initializers, and a comparison can live in an earlier body.
func (l *lowering) hasGenericFunctionValues() bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found || ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) {
			symbol := l.symbol(node)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || len(declaration.TypeParameters()) == 0 || node == declaration.Name() {
						continue
					}
					parent, use := node.Parent, node
					for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
						use, parent = parent, parent.Parent
					}
					if parent == nil || parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression != use {
						found = true
						return false
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		file.Node.ForEachChild(visit)
		if found {
			return true
		}
	}
	return false
}

// Object.is and collection searches also observe function identity. Until
// specialized closures carry a shared source identity, refuse those observations.
func (l *lowering) genericFunctionIdentityCall(node *ast.Node) error {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	property := callee.AsPropertyAccessExpression()
	observes := false
	if callee.Name().Text() == "is" && l.isLibraryGlobal(property.Expression, "Object") {
		for _, argument := range call.Arguments.Nodes {
			observes = observes || l.functionIdentityType(l.checker.GetTypeAtLocation(argument))
		}
	} else {
		switch callee.Name().Text() {
		case "includes", "indexOf", "lastIndexOf", "has", "add", "delete", "get", "set":
			receiver := l.concrete(l.checker.GetTypeAtLocation(property.Expression))
			if l.checker.IsArrayType(receiver) || l.isLibraryType(receiver, "Set", "ReadonlySet", "Map", "ReadonlyMap") {
				arguments := l.typeArguments(receiver)
				if len(arguments) > 0 {
					observes = l.functionIdentityType(arguments[0])
				}
			}
		}
	}
	if observes && l.hasGenericFunctionValues() {
		return l.notYet(node, "function identity observation in a program with specialized generic function values")
	}
	return nil
}

func (l *lowering) functionIdentityType(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if l.functionIdentityType(member) {
				return true
			}
		}
		return false
	}
	return len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0
}

// Constructing a collection already hashes and deduplicates its keys, before
// any explicit has/get operation can be guarded.
func (l *lowering) genericFunctionIdentityNew(node *ast.Node) error {
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.isLibraryType(proven, "Set", "Map") {
		return nil
	}
	arguments := l.typeArguments(proven)
	if len(arguments) > 0 && l.functionIdentityType(arguments[0]) && l.hasGenericFunctionValues() {
		return l.notYet(node, "function identity observation in a program with specialized generic function values")
	}
	return nil
}

// Context belongs to the operand itself. The checker propagates it through
// conditional/logical expressions and parentheses; lowering need not infer it
// by climbing the syntax tree or borrowing the other operand's signature.
func (l *lowering) contextualGenericSignature(node *ast.Node, target *checker.Signature) *checker.Signature {
	contextual := l.concrete(l.checker.GetContextualType(node, checker.ContextFlagsNone))
	if contextual == nil || target == nil {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.withoutUndefined(contextual), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 {
		return nil
	}
	return instantiateSignatureInContextOf(l.checker, target, signatures[0], nil, nil)
}

//go:linkname typeFromSignature github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getOrCreateTypeFromSignature
func typeFromSignature(receiver *checker.Checker, signature *checker.Signature) *checker.Type

func (l *lowering) contextualGenericType(node *ast.Node, own *checker.Type) *checker.Type {
	signatures := l.checker.GetSignaturesOfType(own, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) == 0 {
		return own
	}
	if resolved := l.contextualGenericSignature(node, signatures[0]); resolved != nil {
		return typeFromSignature(l.checker, resolved)
	}
	return own
}

// Functions are truthy exactly when their nullable pointer is present. Reuse
// the IR's lazy branches, so neither operand is evaluated twice or eagerly.
func (l *lowering) functionLogical(node *ast.Node, left, right ir.Expression) (ir.Expression, bool) {
	operator := node.AsBinaryExpression().OperatorToken.Kind
	if left.Type() != ir.Closure || (right.Type() != ir.Closure) {
		return nil, false
	}
	if operator == ast.KindBarBarToken {
		return ir.Coalesce{Value: left, Fallback: fit(right, ir.Closure), Of: ir.Closure}, true
	}
	if operator == ast.KindAmpersandAmpersandToken {
		return ir.Conditional{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: left}}, WhenTrue: fit(right, ir.Closure), WhenNot: ir.Undefined{Of: ir.Closure}, Of: ir.Closure}, true
	}
	return nil, false
}
