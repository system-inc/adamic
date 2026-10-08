package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A TypeScript predicate can admit an open structural view. Its narrowed reads
// use the same admission and member checks as an interface cast.
func (l *lowering) predicateViewHatch(node *ast.Node) bool {
	if !l.checkedAssertionSource(node) {
		return false
	}
	annotation := node.AsTypePredicateNode()
	function := node.Parent
	if annotation.Type == nil || annotation.AssertsModifier != nil || !ast.IsFunctionLike(function) || function.Body() == nil {
		return false
	}
	target := l.concrete(l.checker.GetTypeFromTypeNode(annotation.Type))
	of, known := l.representation(target)
	if !l.predicateViewObject(target) || !known || of != ir.Object {
		return false
	}
	for _, parameter := range function.Parameters() {
		if parameter.Name().Kind != ast.KindIdentifier || parameter.Name().Text() != annotation.ParameterName.Text() {
			continue
		}
		source := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(parameter.Name())))
		return l.predicateViewObject(source) && l.checker.IsTypeAssignableTo(target, source) && l.widened(target, source, map[[2]*checker.Type]bool{}) == nil
	}
	return false
}

// Registering the target family also protects narrowed property/element aliases;
// shared views conservatively check member reads across function boundaries.
func (l *lowering) admitPredicateView(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	return l.view(node, value, target)
}

// Compound structural targets reuse the shared intersection/union descriptors.
// Scalars, collections, callable values and nominal classes retain their gates.
func (l *lowering) predicateViewObject(of *checker.Type) bool {
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			if !l.predicateViewObject(member) {
				return false
			}
		}
		return true
	}
	if of.Flags()&checker.TypeFlagsIntersection != 0 {
		return l.structuralViewIntersection(of)
	}
	representation, known := l.representation(of)
	return of.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(of) && !l.callableViewContract(of) && known && representation == ir.Object
}
