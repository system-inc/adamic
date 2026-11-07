package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A literal producer known to violate an object-union view is refused at its
// demanded destructured read. Unknown producers retain the runtime boundary.
// Following only const initializers is conservative: a later repair through an
// alias does not certify this construction. Build a member of the target union.
func (lowering *lowering) refuseIncompatibleDestructuredView(pattern, initializer *ast.Node) error {
	if pattern.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	receiver := lowering.checker.GetTypeAtLocation(initializer)
	producer := initializer
	seen := map[*ast.Node]bool{}
	for producer != nil && !seen[producer] {
		seen[producer] = true
		producer = ast.SkipParentheses(producer)
		if producer.Kind == ast.KindAsExpression {
			producer = producer.AsAsExpression().Expression
			continue
		}
		if !ast.IsIdentifier(producer) {
			break
		}
		symbol := lowering.checker.GetSymbolAtLocation(producer)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return nil
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return nil
		}
		producer = declaration.AsVariableDeclaration().Initializer
	}
	if producer == nil || producer.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	actual := lowering.checker.GetTypeAtLocation(producer)
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		if binding.Kind != ast.KindBindingElement || binding.Name() == nil {
			continue
		}
		name := binding.Name()
		if binding.AsBindingElement().PropertyName != nil {
			name = binding.AsBindingElement().PropertyName
		}
		if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral {
			continue
		}
		field := name.Text()
		if !lowering.checkedViewReceiverField(nil, receiver, field) {
			continue
		}
		targetField, sourceField := lowering.checker.GetPropertyOfType(
			receiver,
			field,
		), lowering.checker.GetPropertyOfType(actual, field)
		if targetField == nil || sourceField == nil {
			continue
		}
		target := lowering.checker.GetTypeOfSymbol(targetField)
		if target.Flags()&checker.TypeFlagsUnion == 0 {
			continue
		}
		objects := true
		for _, member := range target.Types() {
			objects = objects && member.Flags()&checker.TypeFlagsObject != 0 && !lowering.callableViewContract(member)
		}
		if objects && !lowering.checker.IsTypeAssignableTo(lowering.checker.GetTypeOfSymbol(sourceField), target) {
			return &Refused{
				Where: lowering.program.Where(binding),
				What:  "a destructured object-union view with an incompatible literal producer",
				Fix: "build a member of " +
					lowering.checker.TypeToString(target) +
					" before destructuring field " +
					field,
			}
		}
	}
	return nil
}
