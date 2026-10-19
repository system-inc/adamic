package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A literal producer known to violate an object-union view is refused at its
// demanded destructured read. Unknown producers retain the runtime boundary.
// Following only const initializers is conservative: a later repair through an
// alias does not certify this construction. Build a member of the target union.
func (l *lowering) refuseIncompatibleDestructuredView(pattern, initializer *ast.Node) error {
	if pattern.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	receiver := l.checker.GetTypeAtLocation(initializer)
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
		symbol := l.checker.GetSymbolAtLocation(producer)
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
	actual := l.checker.GetTypeAtLocation(producer)
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
		if !l.checkedViewReceiverField(receiver, field) {
			continue
		}
		targetField, sourceField := l.checker.GetPropertyOfType(receiver, field), l.checker.GetPropertyOfType(actual, field)
		if targetField == nil || sourceField == nil {
			continue
		}
		target := l.checker.GetTypeOfSymbol(targetField)
		if target.Flags()&checker.TypeFlagsUnion == 0 {
			continue
		}
		objects := true
		for _, member := range target.Types() {
			objects = objects && member.Flags()&checker.TypeFlagsObject != 0 && !l.callableViewContract(member)
		}
		if objects && !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(sourceField), target) {
			return &Refused{Where: l.program.Where(binding), What: "a destructured object-union view with an incompatible literal producer", Fix: "build a member of " + l.checker.TypeToString(target) + " before destructuring field " + field}
		}
	}
	return nil
}
