package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// overloadPhantomResult says whether an overload's promised result differs from the implementation's
// only by phantom brands, below the top level too (rulings.md:34, with :32's presence refusals): tsc's
// getPathComponents promises PathPathComponents, an Array<Path> extension with a phantom brand member,
// over an implementation that returns string[] (#z5d0jye step 3, Adamic's application of :34).
//
// A brand comes off a type only through phantomExtension (an interface extending exactly one base and
// adding only phantom members), the shape arrayBrandIn sees, so every observation of an array brand's
// presence is refused (phantom_array_presence.go). The intersection spelling `Path[] & { brand }` is never
// stripped, since :32's refusals don't see it (Language's must). A primitive brand comes off through
// phantomBase. What's left on both sides must be the same type, at the top or as the elements of arrays
// of the same mutability, and nothing else may differ.
//
// An element brand that differs needs the returned array fresh (overloadReturnsFreshArray): otherwise one
// array is held both as string[] and as Path[], and a plain string pushed through the first reads back
// through the second as a Path no mint made, the pairing invariance refuses everywhere else (Compiler's
// condition). A readonly promise doesn't lift that, since the alias that writes is another one.
func (l *lowering) overloadPhantomResult(implementation *ast.Node, produced, promised *checker.Type) bool {
	from, to := l.phantomStripped(produced), l.phantomStripped(promised)
	if from == nil || to == nil {
		return false
	}
	if from.Id() == to.Id() {
		// Only a brand on the value itself differs, if anything does.
		return true
	}
	if !l.checker.IsArrayType(from) || !l.checker.IsArrayType(to) || l.isLibraryType(from, "ReadonlyArray") != l.isLibraryType(to, "ReadonlyArray") {
		return false
	}
	fromElement, toElement := l.checker.GetElementTypeOfArrayType(from), l.checker.GetElementTypeOfArrayType(to)
	if fromElement.Id() == toElement.Id() {
		// Only the array's own brand differs.
		return true
	}
	strippedFrom, strippedTo := l.phantomStripped(fromElement), l.phantomStripped(toElement)
	if strippedFrom == nil || strippedTo == nil || strippedFrom.Id() != strippedTo.Id() || strippedFrom.Flags()&brandablePrimitiveFlags == 0 {
		return false
	}
	return l.overloadReturnsFreshArray(implementation, map[*ast.Node]bool{})
}

// phantomStripped is t with its phantom brands taken off: every phantomExtension to its base, then a
// primitive brand to its primitive. nil when t can't be read.
func (l *lowering) phantomStripped(t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	t = l.concrete(t)
	for depth := 0; ; depth++ {
		if depth > 16 {
			return nil
		}
		base, _ := l.phantomExtension(t)
		if base == nil {
			break
		}
		t = l.concrete(base)
	}
	if primitive := l.phantomBase(t); primitive != nil {
		return primitive
	}
	return t
}

// overloadReturnsFreshArray says whether every array function returns is one nothing else holds: an
// array literal, a freshValue copy, or the result of a direct call by name to a function declaration
// nothing can replace (directPredicateCallee, as neverCall trusts one) whose own returns all pass. A
// cycle, a nested function's return, or any other return fails closed.
func (l *lowering) overloadReturnsFreshArray(function *ast.Node, visiting map[*ast.Node]bool) bool {
	if function == nil || function.Body() == nil || visiting[function] {
		return false
	}
	visiting[function] = true
	defer delete(visiting, function)
	if function.Body().Kind != ast.KindBlock {
		return l.overloadFreshArrayValue(function.Body(), visiting)
	}
	fresh, returned := true, false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !fresh {
			return true
		}
		if ast.IsFunctionLike(node) || ast.IsClassLike(node) {
			return false // its returns are its own
		}
		if node.Kind == ast.KindReturnStatement {
			returned = true
			expression := node.AsReturnStatement().Expression
			if expression == nil || !l.overloadFreshArrayValue(expression, visiting) {
				fresh = false
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	function.Body().ForEachChild(visit)
	return fresh && returned
}

func (l *lowering) overloadFreshArrayValue(expression *ast.Node, visiting map[*ast.Node]bool) bool {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindCallExpression:
		if l.freshValue(expression) {
			return true
		}
		callee := ast.SkipParentheses(expression.AsCallExpression().Expression)
		symbol := l.symbol(callee)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if !l.directPredicateCallee(callee, declaration) || declaration.Body() == nil {
			return false
		}
		return l.overloadReturnsFreshArray(declaration, visiting)
	case ast.KindNewExpression:
		return l.freshValue(expression)
	}
	return false
}
