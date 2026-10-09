package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Date is nominal at every structural boundary because its scalar internal slot is not an own field.
func (l *lowering) dateViewsMatch(from *checker.Type, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	// Date internal slots cannot be supplied by structural objects or lost through a method interface.
	if l.isLibraryType(from, "Date") != l.isLibraryType(to, "Date") {
		return false
	}
	same := func(inside, viewed *checker.Type) bool {
		return l.dateViewsMatch(inside, viewed, visited)
	}
	fromSignatures := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toSignatures := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	switch {
	case len(fromSignatures) > 0 && len(toSignatures) > 0:
		fromParameters, toParameters := fromSignatures[0].Parameters(), toSignatures[0].Parameters()
		for index := 0; index < len(fromParameters) && index < len(toParameters); index++ {
			if !same(l.checker.GetTypeOfSymbol(fromParameters[index]), l.checker.GetTypeOfSymbol(toParameters[index])) {
				return false
			}
		}
		return same(l.checker.GetReturnTypeOfSignature(fromSignatures[0]), l.checker.GetReturnTypeOfSignature(toSignatures[0]))
	case from.ObjectFlags()&checker.ObjectFlagsReference != 0 && to.ObjectFlags()&checker.ObjectFlagsReference != 0 && (l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet")):
		fromArguments, toArguments := l.typeArguments(from), l.typeArguments(to)
		for index := 0; index < len(fromArguments) && index < len(toArguments); index++ {
			if !same(fromArguments[index], toArguments[index]) {
				return false
			}
		}
	default:
		for _, viewed := range l.checker.GetPropertiesOfType(to) {
			if viewed.Flags&ast.SymbolFlagsMethod != 0 {
				continue
			}
			if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil && !same(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)) {
				return false
			}
		}
	}
	return true
}

// The one union boundary that consumes a Date without erasing its internal slot:
// numeric/string/Date constructor and fs.utimes overloads select by the actual type.
func (l *lowering) dateContextBorrow(node *ast.Node) bool {
	if !l.isLibraryType(l.checker.GetTypeAtLocation(node), "Date") {
		return false
	}
	if l.nodeFSFileReadOnlyArgument(node) {
		return true
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	return parent != nil && parent.Kind == ast.KindNewExpression && l.isLibraryGlobal(parent.AsNewExpression().Expression, "Date")
}
