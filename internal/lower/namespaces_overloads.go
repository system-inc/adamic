package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Identical concrete overload contracts (including unused signature type
// parameters) add no narrower promise than the nongeneric implementation body
// already proves. Different, generic and predicate contracts require separate
// verification; never simply discard their signatures.
func (l *lowering) identicalOverloadImplementation(declaration *ast.Node) *ast.Node {
	symbol := l.symbol(declaration.Name())
	if symbol == nil {
		return nil
	}
	var implementation *ast.Node
	for _, candidate := range symbol.Declarations {
		if candidate.Kind != ast.KindFunctionDeclaration || candidate.Parent != declaration.Parent {
			return nil
		}
		if candidate.Body() != nil {
			if implementation != nil {
				return nil
			}
			implementation = candidate
		}
	}
	if implementation == nil {
		return nil
	}
	for _, candidate := range symbol.Declarations {
		if !l.identicalOverloadContract(candidate, implementation) {
			return nil
		}
	}
	return implementation
}
func (l *lowering) identicalOverloadContract(left, right *ast.Node) bool {
	if len(right.TypeParameters()) != 0 {
		return false
	}
	if left.Type() != nil && left.Type().Kind == ast.KindTypePredicate || right.Type() != nil && right.Type().Kind == ast.KindTypePredicate {
		return false
	}
	a, b := left.Parameters(), right.Parameters()
	if len(a) != len(b) {
		return false
	}
	for index, p := range a {
		q := b[index]
		for _, parameter := range []*ast.Node{p, q} {
			d := parameter.AsParameterDeclaration()
			if !ast.IsIdentifier(parameter.Name()) || parameter.Name().Text() == "this" || d.Initializer != nil || d.QuestionToken != nil || d.DotDotDotToken != nil {
				return false
			}
		}
		if !identicalTypes(l.checker, l.checker.GetTypeAtLocation(p.Name()), l.checker.GetTypeAtLocation(q.Name())) {
			return false
		}
	}
	ls, rs := l.checker.GetSignatureFromDeclaration(left), l.checker.GetSignatureFromDeclaration(right)
	return identicalTypes(l.checker, l.checker.GetReturnTypeOfSignature(ls), l.checker.GetReturnTypeOfSignature(rs))
}
