package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// signatureProven must come from the implementation certificate, never from a
// runtime function tag or the asserted target type. Reads still check readiness
// and kind; a certificate only authorizes using the declared calling convention.
func (l *lowering) viewCallableCall(node *ast.Node, member string, signatureProven bool) error {
	return checkViewCallableSignature(l.program.Where(node), member, signatureProven)
}

func checkViewCallableSignature(where, member string, signatureProven bool) error {
	if signatureProven {
		return nil
	}
	return &Refused{
		Where: where,
		What:  "a checked view call to member " + member,
		Fix:   "its signature cannot be checked at runtime and no compatible implementation is proven; prove the callable implementation before calling through this view",
	}
}

func init() { viewCallableContractHook = buildViewCallableContract }

func buildViewCallableContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (ir.ViewContractID, error) {
	if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 {
		return 0, l.notYet(node, "a constructor checked-view contract")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewCallable, Name: l.checker.TypeToString(target), Of: ir.Closure})
	l.result.ViewContractTypes[int(target.Id())] = id
	return id, nil
}

// All same-named implementations in this closed program must be compatible.
// This deliberately overapproximates allocation flow. Opaque inputs never count
// as a body certificate, and method bivariance is not used as signature evidence.
func (l *lowering) viewCallableFieldUses(node *ast.Node, target *checker.Type, property *ast.Symbol) error {
	// A supported shape is certified from producer code at each read, not from
	// same-named implementations elsewhere. Unsupported shapes retain the old proof.
	if l.runtimeViewCallableShape(l.checker.GetTypeOfSymbol(property)) {
		return nil
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	wanted := l.checker.GetTypeOfSymbol(property)
	proven, implementations, called := true, 0, false
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if ast.HasSyntacticModifier(part, ast.ModifierFlagsAmbient) {
			proven = false
		}
		if part.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(part.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == property.Name {
				called = true
			}
		}
		// A captured or passed member can be called through an alias. Require the
		// same implementation proof for escaping reads as for direct calls.
		if part.Kind == ast.KindPropertyAccessExpression && part.Name().Text() == property.Name && part.Parent != nil && part.Parent.Kind != ast.KindExpressionStatement {
			called = true
		}
		var value *ast.Node
		if part.Name() != nil && part.Name().Text() == property.Name {
			switch part.Kind {
			case ast.KindPropertyAssignment:
				value = part.AsPropertyAssignment().Initializer
			case ast.KindShorthandPropertyAssignment:
				value = part.Name()
			case ast.KindPropertyDeclaration:
				value = part.AsPropertyDeclaration().Initializer
			case ast.KindMethodDeclaration:
				value = part
			}
		}
		if part.Kind == ast.KindBinaryExpression {
			binary := part.AsBinaryExpression()
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				left := ast.SkipParentheses(binary.Left)
				if left.Kind == ast.KindPropertyAccessExpression && left.Name().Text() == property.Name {
					value = binary.Right
				}
				if left.Kind == ast.KindElementAccessExpression {
					proven = false
				}
			}
		}
		if value != nil && !l.uninitializedInitializer(value) {
			actual := l.checker.GetTypeAtLocation(value)
			if l.callableViewContract(actual) {
				implementations++
				proven = proven && l.viewSignatureCompatible(actual, wanted) && l.viewCallableBody(value, map[*ast.Symbol]bool{})
			}
		}
		part.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if !called {
		return nil
	}
	return l.viewCallableCall(node, l.checker.TypeToString(target)+"."+property.Name, proven && implementations > 0)
}

func (l *lowering) viewCallableBody(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindMethodDeclaration {
		return node.Body() != nil && len(node.TypeParameters()) == 0
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] {
		return false
	}
	seen[symbol] = true
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			return declaration.Body() != nil
		}
		if declaration.Kind == ast.KindVariableDeclaration && declaration.AsVariableDeclaration().Initializer != nil {
			return l.viewCallableBody(declaration.AsVariableDeclaration().Initializer, seen)
		}
	}
	return false
}

func (l *lowering) viewSignatureCompatible(actual, wanted *checker.Type) bool {
	a, w := l.checker.GetSignaturesOfType(actual, checker.SignatureKindCall), l.checker.GetSignaturesOfType(wanted, checker.SignatureKindCall)
	if len(a) != 1 || len(w) != 1 || len(a[0].TypeParameters()) != 0 || len(w[0].TypeParameters()) != 0 || a[0].HasRestParameter() || w[0].HasRestParameter() || len(a[0].Parameters()) != len(w[0].Parameters()) || a[0].MinArgumentCount() != len(a[0].Parameters()) || w[0].MinArgumentCount() != len(w[0].Parameters()) {
		return false
	}
	for index, p := range a[0].Parameters() {
		from, to := l.checker.GetTypeOfSymbol(w[0].Parameters()[index]), l.checker.GetTypeOfSymbol(p)
		f, fk := l.representation(from)
		t, tk := l.representation(to)
		if !fk || !tk || f != t || !l.classAssignable(from, to) || l.widened(from, to, map[[2]*checker.Type]bool{}) != nil {
			return false
		}
	}
	from, to := l.checker.GetReturnTypeOfSignature(a[0]), l.checker.GetReturnTypeOfSignature(w[0])
	if from.Flags()&checker.TypeFlagsVoid != 0 && to.Flags()&checker.TypeFlagsVoid != 0 {
		return true
	}
	f, fk := l.representation(from)
	t, tk := l.representation(to)
	return fk && tk && f == t && l.classAssignable(from, to) && l.widened(from, to, map[[2]*checker.Type]bool{}) == nil
}

// A tag is a nominal certificate only when every construction that can carry
// that tag has the wanted class identity. No extra runtime identity is invented.
func (l *lowering) viewProvenClassCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if !isClassInstance(target) || value.Type() != ir.Object || source.Flags()&checker.TypeFlagsObject == 0 {
		return nil, nil
	}
	for _, property := range l.checker.GetPropertiesOfType(source) {
		literal := l.fieldLiteral(target, property.Name)
		if literal == nil {
			continue
		}
		modules, err := l.moduleOrder(l.program.Files()[0])
		if err != nil {
			return nil, err
		}
		proven, constructions := true, 0
		var visit ast.Visitor
		visit = func(part *ast.Node) bool {
			if ast.HasSyntacticModifier(part, ast.ModifierFlagsAmbient) {
				proven = false
			}
			if part.Kind == ast.KindObjectLiteralExpression || part.Kind == ast.KindNewExpression {
				actual := l.checker.GetTypeAtLocation(part)
				tag := l.checker.GetPropertyOfType(actual, property.Name)
				if tag != nil && l.checker.IsTypeAssignableTo(literal, l.checker.GetTypeOfSymbol(tag)) {
					constructions++
					proven = proven && isClassInstance(actual) && l.classAssignable(actual, target)
				}
			}
			if part.Kind == ast.KindBinaryExpression {
				binary := part.AsBinaryExpression()
				if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					left := ast.SkipParentheses(binary.Left)
					if left.Kind == ast.KindElementAccessExpression || left.Kind == ast.KindPropertyAccessExpression && left.Name().Text() == property.Name {
						proven = false
					}
				}
			}
			part.ForEachChild(visit)
			return false
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
		if !proven || constructions == 0 {
			return nil, l.viewCallableCall(node, l.checker.TypeToString(target)+".<class methods>", false)
		}
		allowed, of, known := l.literalConstant(literal)
		if !known {
			continue
		}
		return ir.CheckedCast{Value: value, Field: property.Name, FieldType: of, Allowed: []ir.Expression{allowed}, CheckedFields: true, Message: "cast failed: this " + l.checker.TypeToString(source) + " is not a " + l.checker.TypeToString(target)}, nil
	}
	return nil, nil
}
