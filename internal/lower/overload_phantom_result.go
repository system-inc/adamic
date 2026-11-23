package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// overloadPhantomResult judges an overload whose promised result differs from the implementation's only by
// phantom brands (rulings.md:34, with :32's presence refusals), as Adamic ruled on #z1qkw7d. It answers
// whether the overload is served here, and the refusal that stops it when it's a mint .a doesn't admit.
//
// A promised result that's a phantom extension at the top (phantomExtension: an interface extending one
// base and adding only phantom members) is an overload mint: the claim a brand mint's cast makes, made at
// the declaration for every call that resolves to the overload. tsc's getPathComponents promises
// PathPathComponents, an Array<Path> extension with a phantom brand member, over an implementation that
// returns string[] (#z5d0jye step 3). It's held to the cast's rule exactly, brandMintOf, judged as a type
// with no value behind it (a nil operand), since the census reads signatures, not returns. Only a primitive
// brand on the elements comes off first (overloadMintSource), which is that rule's own silent erasure. The
// mint is accepted in .ts and listed in the checked report as the overload's (recordOverloadMint), and
// refused in .a, as a cast's mint is, at each call the overload's verdict stops. An extension brandMintOf
// refuses isn't served here, whatever its base: a member that isn't phantom, a chain of extensions over the
// chain's last base, or a value that doesn't relate to the base.
//
// A primitive phantom brand, at the top or on an array's elements, stays the silent erasure it is for a
// cast (rulings.md:30, overloadPrimitiveBrands).
//
// Any array brand that differs, on the array or on its elements, needs the returned array fresh
// (overloadReturnsFreshArray): otherwise one array is held both as string[] and as Path[], and a plain
// string pushed through the first reads back through the second as a Path no mint made, or an alias's
// push breaks the sortedness a Sorted brand claims; that's the pairing invariance refuses everywhere else
// (Compiler's condition, read by Adamic as any array brand). A readonly promise doesn't lift it, since
// the alias that writes is another one.
func (l *lowering) overloadPhantomResult(implementation, overload *ast.Node, label string, produced, promised *checker.Type) (bool, error) {
	from, to := l.concrete(produced), l.concrete(promised)
	base, _ := l.phantomExtension(to)
	if base == nil {
		return l.overloadPrimitiveBrands(implementation, from, to), nil
	}
	// A narrow overload's every call checks its result through the brand's view (overloadResultCheck), which
	// holds more than a mint claims, so that road keeps it.
	if l.narrowOverload(overload) {
		return false, nil
	}
	mint := l.brandMintOf(nil, l.overloadMintSource(from, base), to)
	if mint == nil {
		return false, nil
	}
	if l.isArrayType(base) && !l.overloadReturnsFreshArray(implementation, map[*ast.Node]bool{}) {
		return false, nil
	}
	if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(overload)), ".a") {
		return false, &Refused{Where: l.program.Where(overload), What: label + " result " + l.checker.TypeToString(promised) + ", a brand mint in .a: an overload result that adds only its phantom member " + strings.Join(mint.members, ", ") + " to implementation result " + l.checker.TypeToString(produced), Fix: "mint the brand in a .ts file, where the overload is accepted and listed in the checked report (rulings.md:29)"}
	}
	l.recordOverloadMint(overload, label, promised, mint)
	return true, nil
}

// overloadMintSource is produced as brandMintOf judges it against base, the promised extension's base: base
// itself when both are arrays of one mutability whose elements differ only by a primitive phantom brand
// (phantomBrandOnly), so only the extension's own members are left to the mint. That's getPathComponents'
// string[] against Array<Path>. Otherwise it's produced. base is asked of the checker directly, so the base
// of a chain of extensions is never an array here, and the chain is judged whole.
func (l *lowering) overloadMintSource(produced, base *checker.Type) *checker.Type {
	if !l.isArrayType(produced) || !l.checker.IsArrayType(base) || l.isLibraryType(produced, "ReadonlyArray") != l.isLibraryType(base, "ReadonlyArray") {
		return produced
	}
	if l.phantomBrandOnly(l.concrete(l.elementTypeOfArray(produced)), l.concrete(l.checker.GetElementTypeOfArrayType(base))) {
		return base
	}
	return produced
}

// phantomBrandOnly is whether from and to differ only by a primitive phantom brand, read both ways, the test
// censusOverload makes of a whole result before any other (rulings.md:30).
func (l *lowering) phantomBrandOnly(from, to *checker.Type) bool {
	return (l.hasPhantom(from) || l.hasPhantom(to)) && l.phantomCast(from, to) && l.phantomAssignable(from, to) && l.phantomAssignable(to, from)
}

// overloadPrimitiveBrands says whether produced and promised, neither a phantom extension the census mints,
// differ only by primitive phantom brands, at the top or as the elements of arrays of the same mutability
// (rulings.md:34 with :32). A primitive brand comes off through phantomBase. The intersection spelling
// `Path[] & { brand }` is never stripped, since :32's refusals don't see it (Language's must). What's left
// on both sides must be the same type, and nothing else may differ.
func (l *lowering) overloadPrimitiveBrands(implementation *ast.Node, produced, promised *checker.Type) bool {
	from, to := l.phantomStripped(produced), l.phantomStripped(promised)
	if from == nil || to == nil {
		return false
	}
	if from.Id() == to.Id() {
		if l.isArrayType(from) {
			// Only an array brand on the implementation's result differs, which the promise drops.
			return l.overloadReturnsFreshArray(implementation, map[*ast.Node]bool{})
		}
		// Only a primitive brand on the value itself differs, if anything does.
		return true
	}
	if !l.isArrayType(from) || !l.isArrayType(to) || l.isLibraryType(from, "ReadonlyArray") != l.isLibraryType(to, "ReadonlyArray") {
		return false
	}
	fromElement, toElement := l.elementTypeOfArray(from), l.elementTypeOfArray(to)
	if fromElement.Id() == toElement.Id() {
		// Two array types with one element and one mutability are one type, held as their base: only an
		// array brand on the implementation's result differs.
		return l.overloadReturnsFreshArray(implementation, map[*ast.Node]bool{})
	}
	strippedFrom, strippedTo := l.phantomStripped(fromElement), l.phantomStripped(toElement)
	if strippedFrom == nil || strippedTo == nil || strippedFrom.Id() != strippedTo.Id() || strippedFrom.Flags()&brandablePrimitiveFlags == 0 {
		return false
	}
	return l.overloadReturnsFreshArray(implementation, map[*ast.Node]bool{})
}

// phantomStripped is t with a primitive brand taken off, or t. nil when t can't be read. A phantom extension
// stays: one at the top of a promise is a mint, judged by overloadPhantomResult, and one anywhere else is a
// difference no erasure covers.
func (l *lowering) phantomStripped(t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	t = l.concrete(t)
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
