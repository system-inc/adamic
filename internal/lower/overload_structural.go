package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A structural result need not have a runtime shape predicate when all returns
// are proved under this overload's admitted arguments. The proof uses Adamic's
// relation, including mutable invariance, rather than reversing covariance.
func (l *lowering) overloadStructuralProof(implementation, overload *ast.Node) bool {
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)))
	if promised.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion) == 0 {
		return false
	}
	return l.censusProveOverloadResult(implementation, overload)
}

// Report the first unsupported result component. A wider record is not a
// narrower record merely because the narrower one can be widened back to it.
func (l *lowering) overloadResultPath(produced, promised *checker.Type, path string, visited map[[2]*checker.Type]bool) string {
	produced, promised = l.concrete(produced), l.concrete(promised)
	pair := [2]*checker.Type{produced, promised}
	if visited[pair] {
		return path
	}
	visited[pair] = true
	if produced.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range produced.Types() {
			if !l.censusRelated(member, promised) {
				return l.overloadResultPath(member, promised, path, visited)
			}
		}
	}
	if promised.Flags()&checker.TypeFlagsObject != 0 {
		for _, property := range l.checker.GetPropertiesOfType(promised) {
			own := l.checker.GetPropertyOfType(produced, property.Name)
			next := path + "." + property.Name
			if own == nil || own.Flags&ast.SymbolFlagsOptional != 0 && property.Flags&ast.SymbolFlagsOptional == 0 {
				return next
			}
			source, target := l.checker.GetTypeOfSymbol(own), l.checker.GetTypeOfSymbol(property)
			if !l.checker.IsReadonlySymbol(property) && (l.checker.IsReadonlySymbol(own) || !l.censusRelated(target, source)) {
				return next
			}
			if !l.censusRelated(source, target) {
				return l.overloadResultPath(source, target, next, visited)
			}
		}
	}
	return path
}
