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
func instantiateSignatureInContextOf(
	receiver *checker.Checker,
	signature *checker.Signature,
	contextual *checker.Signature,
	inference *checker.InferenceContext,
	compare checker.TypeComparer,
) *checker.Signature

func (lowering *lowering) genericFunctionValue(node *ast.Node, declaration *ast.Node) (ir.Expression, error) {
	resolved := lowering.contextualGenericSignature(node, lowering.checker.GetSignatureFromDeclaration(declaration))
	if resolved == nil {
		return nil, lowering.notYet(node, "a generic function value without a concrete contextual signature")
	}
	index, instantiateFunctionSignatureError := lowering.instantiateFunctionSignature(node, declaration, resolved)
	if instantiateFunctionSignatureError != nil {
		return nil, instantiateFunctionSignatureError
	}
	return lowering.functionValue(node, index)
}

// Inspect source rather than already-created forwarders: module function bodies
// are lowered before initializers, and a comparison can live in an earlier body.
func (lowering *lowering) hasGenericFunctionValues() bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found || ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) {
			symbol := lowering.symbol(node)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if declaration.Kind != ast.KindFunctionDeclaration ||
						declaration.Body() == nil ||
						len(declaration.TypeParameters()) == 0 ||
						node == declaration.Name() {
						continue
					}
					parent, use := node.Parent, node
					// A qualified generic direct call is not a detached function value.
					if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == node {
						use, parent = parent, parent.Parent
					}
					for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
						use, parent = parent, parent.Parent
					}
					if parent == nil ||
						parent.Kind != ast.KindCallExpression ||
						parent.AsCallExpression().Expression != use {
						found = true
						return false
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range lowering.program.CompilerProgram().GetSourceFiles() {
		file.Node.ForEachChild(visit)
		if found {
			return true
		}
	}
	return false
}

// Object.is and collection searches also observe function identity. Until
// specialized closures carry a shared source identity, refuse those observations.
func (lowering *lowering) genericFunctionIdentityCall(node *ast.Node) error {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	property := callee.AsPropertyAccessExpression()
	observes := false
	if callee.Name().Text() == "is" && lowering.isLibraryGlobal(property.Expression, "Object") {
		for _, argument := range call.Arguments.Nodes {
			observes = observes || lowering.functionIdentityType(lowering.checker.GetTypeAtLocation(argument))
		}
	} else {
		switch callee.Name().Text() {
		case "includes", "indexOf", "lastIndexOf", "has", "add", "delete", "get", "set":
			receiver := lowering.concrete(lowering.checker.GetTypeAtLocation(property.Expression))
			if lowering.isArrayType(receiver) ||
				lowering.isLibraryType(receiver, "Set", "ReadonlySet", "Map", "ReadonlyMap") {
				arguments := lowering.typeArguments(receiver)
				if len(arguments) > 0 {
					observes = lowering.functionIdentityType(arguments[0])
				}
			}
		}
	}
	if observes && lowering.hasGenericFunctionValues() {
		return lowering.notYet(
			node,
			"function identity observation in a program with specialized generic function values",
		)
	}
	return nil
}

func (lowering *lowering) functionIdentityType(proven *checker.Type) bool {
	proven = lowering.concrete(proven)
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if lowering.functionIdentityType(member) {
				return true
			}
		}
		return false
	}
	return len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) > 0
}

// Constructing a collection already hashes and deduplicates its keys, before
// any explicit has/get operation can be guarded.
func (lowering *lowering) genericFunctionIdentityNew(node *ast.Node) error {
	proven := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if !lowering.isLibraryType(proven, "Set", "Map") {
		return nil
	}
	arguments := lowering.typeArguments(proven)
	if len(arguments) > 0 && lowering.functionIdentityType(arguments[0]) && lowering.hasGenericFunctionValues() {
		return lowering.notYet(
			node,
			"function identity observation in a program with specialized generic function values",
		)
	}
	return nil
}

// Context belongs to the operand itself. The checker propagates it through
// conditional/logical expressions and parentheses; lowering need not infer it
// by climbing the syntax tree or borrowing the other operand's signature.
func (lowering *lowering) contextualGenericSignature(node *ast.Node, target *checker.Signature) *checker.Signature {
	contextual := lowering.concrete(lowering.slotType(node))
	if contextual == nil || target == nil {
		return nil
	}
	signatures := lowering.checker.GetSignaturesOfType(lowering.withoutUndefined(contextual), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 {
		return nil
	}
	return instantiateSignatureInContextOf(lowering.checker, target, signatures[0], nil, nil)
}

//go:linkname typeFromSignature github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getOrCreateTypeFromSignature
func typeFromSignature(receiver *checker.Checker, signature *checker.Signature) *checker.Type

func (lowering *lowering) contextualGenericType(node *ast.Node, own *checker.Type) *checker.Type {
	signatures := lowering.checker.GetSignaturesOfType(own, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) == 0 {
		return own
	}
	if resolved := lowering.contextualGenericSignature(node, signatures[0]); resolved != nil {
		return typeFromSignature(lowering.checker, resolved)
	}
	return own
}

// callableContext accepts only the single concrete function destination used by
// generic-value specialization. Object and container views keep main's rules.
func (lowering *lowering) callableContext(node *ast.Node) *checker.Type {
	context := lowering.concrete(lowering.slotType(node))
	if context == nil {
		return nil
	}
	signatures := lowering.checker.GetSignaturesOfType(lowering.withoutUndefined(context), checker.SignatureKindCall)
	if len(signatures) == 1 && len(signatures[0].TypeParameters()) == 0 {
		return context
	}
	return nil
}
