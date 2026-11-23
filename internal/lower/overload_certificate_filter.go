package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Certificate D (#4z7y6pq): an overload whose result is narrower than its implementation's is served when
// the implementation returns its first parameter as the caller passed it, or a fresh array holding only
// that parameter's elements. tsc's core.ts filter is the shape:
//
//	if (array !== undefined) {
//	    const len = array.length;
//	    let i = 0;
//	    while (i < len && f(array[i])) i++;
//	    if (i < len) {
//	        const result = array.slice(0, i);
//	        i++;
//	        while (i < len) {
//	            const item = array[i];
//	            if (f(item)) {
//	                result.push(item);
//	            }
//	            i++;
//	        }
//	        return result;
//	    }
//	}
//	return array;
//
// Its overloads promise T[] where the implementation's result is readonly T[] | undefined. Handed back,
// the argument is what the caller passed, a value of the overload's parameter type; copied, the result is
// a new mutable array of that parameter's elements. Either is the overload's result when its arguments are
// exactly the caller's, which holds only at a direct call the checker resolved to it, so certificate D
// rides certificate A's per-call gate and storage condition (overloadCertificateCall), never a gate of its
// own. An element read is array[i] at a number, the element tsc types it as, read as every other index
// read in the program is (ir.ArrayIndex).

// overloadFilterShape is what certificate D reads from an implementation's body.
type overloadFilterShape struct {
	// array is the index of the parameter every return hands back or copies the elements of: the first.
	array int
	// fresh is whether some return hands back a fresh array rather than the parameter.
	fresh bool
}

// overloadFilterShapeOf is the shape of implementation's body when it's certificate D's. Its first
// parameter is only read: for its length, for an element at a number, for a copy through the library's
// own slice, tested against undefined or for truthiness, or returned bare. Every other parameter is only
// called directly with one element. Each const initialized to an empty array literal or to the
// parameter's slice is fresh, and a fresh local is only the receiver of the library's own push of one
// element, a read of the parameter or a const initialized to one, or returned bare. Every return hands
// back the parameter or one fresh local, and the body's last statement is a return, so it never falls
// off its end into undefined. No closure reads the parameter or a fresh local, since one could write
// either after the return, and nothing reads arguments, which aliases the parameter.
func (l *lowering) overloadFilterShapeOf(implementation *ast.Node) (overloadFilterShape, bool) {
	shape := overloadFilterShape{}
	if !overloadCertificateFunction(implementation) {
		return shape, false
	}
	parameters := implementation.Parameters()
	if len(parameters) == 0 {
		return shape, false
	}
	for _, parameter := range parameters {
		declaration := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || declaration.Initializer != nil || declaration.DotDotDotToken != nil {
			return shape, false // A default runs where the caller passed nothing, so the parameter isn't the caller's argument.
		}
	}
	array := l.symbol(parameters[shape.array].Name())
	statements := implementation.Body().AsBlock().Statements.Nodes
	if len(statements) == 0 || statements[len(statements)-1].Kind != ast.KindReturnStatement || statements[len(statements)-1].AsReturnStatement().Expression == nil {
		return shape, false
	}

	// The body's own consts, outside any closure: those holding one element, and the fresh arrays.
	elements := map[*ast.Symbol]bool{}
	fresh := map[*ast.Symbol]bool{}
	var collect func(node *ast.Node) bool
	collect = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindVariableStatement {
			if declaration := overloadConstDeclaration(node); declaration != nil && declaration.Initializer() != nil {
				initializer := ast.SkipParentheses(declaration.Initializer())
				switch {
				case l.overloadFilterElementAccess(initializer, array):
					elements[l.symbol(declaration.Name())] = true
				case initializer.Kind == ast.KindArrayLiteralExpression && len(initializer.AsArrayLiteralExpression().Elements.Nodes) == 0, l.overloadFilterSlice(initializer, array):
					fresh[l.symbol(declaration.Name())] = true
				}
			}
		}
		return node.ForEachChild(collect)
	}
	implementation.Body().ForEachChild(collect)
	element := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		if ast.IsIdentifier(node) {
			return elements[l.symbol(node)]
		}
		return l.overloadFilterElementAccess(node, array)
	}

	// Each push: the Array prototype's own push of one element onto a fresh local, as a statement.
	receivers := map[*ast.Node]bool{}
	var pushes func(node *ast.Node) bool
	pushes = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		for local := range fresh {
			if receiver, arguments := l.overloadMutatorCall(node, local, "Array", "push", 1); receiver != nil && element(arguments[0]) {
				receivers[receiver] = true
			}
		}
		return node.ForEachChild(pushes)
	}
	implementation.Body().ForEachChild(pushes)

	// Every reference and every return, judged where it's written.
	var returned *ast.Symbol
	confined := true
	var visit func(node *ast.Node, captured bool)
	visit = func(node *ast.Node, captured bool) {
		if !confined {
			return
		}
		if node.Kind == ast.KindReturnStatement && !captured {
			value := node.AsReturnStatement().Expression
			if value == nil || !ast.IsIdentifier(ast.SkipParentheses(value)) {
				confined = false // A bare return is undefined, and any other expression nothing here proved.
				return
			}
			switch symbol := l.symbol(ast.SkipParentheses(value)); {
			case symbol == array:
			case fresh[symbol] && (returned == nil || returned == symbol):
				returned, shape.fresh = symbol, true
			default:
				confined = false // An alias the body didn't build, or a second fresh local (filter-alias).
				return
			}
		}
		if ast.IsIdentifier(node) && !overloadFilterDeclarationName(node) {
			if node.Text() == "arguments" && !overloadFilterPropertyName(node) {
				confined = false // arguments[0] is the parameter under another name.
				return
			}
			symbol := l.valueSymbol(node)
			switch {
			case symbol == array:
				confined = !captured && l.overloadFilterParameterRead(node, array)
			case fresh[symbol]:
				confined = !captured && (receivers[node] || overloadFilterReturned(node))
			case symbol != nil:
				for index, parameter := range parameters {
					if index != shape.array && symbol == l.symbol(parameter.Name()) {
						// The predicate, called with an element and nothing else: never written, passed on or
						// called on a value that isn't one of the parameter's.
						confined = !captured && l.overloadFilterPredicateCall(node, parameters, index, element)
					}
				}
			}
			if !confined {
				return
			}
		}
		inner := captured || ast.IsFunctionLike(node)
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, inner)
			return !confined
		})
	}
	implementation.Body().ForEachChild(func(child *ast.Node) bool {
		visit(child, false)
		return !confined
	})
	return shape, confined
}

// overloadFilterDeclarationName is whether node, an identifier, is the name a variable declaration
// declares rather than a reference.
func overloadFilterDeclarationName(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration && node.Parent.Name() == node
}

// overloadFilterPropertyName is whether node, an identifier, is the name a property access reads.
func overloadFilterPropertyName(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node
}

// overloadFilterReturned is whether node, an identifier, past parentheses, is a return's value.
func overloadFilterReturned(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindReturnStatement
}

// overloadFilterElementAccess is whether node is array[index] on the parameter array names, past
// parentheses, with no ?. and an index tsc types as a number: one of its elements, never a property
// like array["length"].
func (l *lowering) overloadFilterElementAccess(node *ast.Node, array *ast.Symbol) bool {
	if node.Kind != ast.KindElementAccessExpression {
		return false
	}
	access := node.AsElementAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	if access.QuestionDotToken != nil || !ast.IsIdentifier(receiver) || l.valueSymbol(receiver) != array {
		return false
	}
	index := l.concrete(l.checker.GetTypeAtLocation(access.ArgumentExpression))
	return index.Flags()&^checker.TypeFlagsNumberLike == 0
}

// overloadFilterSlice is whether node is the library's own slice called on the parameter array names,
// past parentheses, with no ?. and no type arguments: a new array of its elements.
func (l *lowering) overloadFilterSlice(node *ast.Node, array *ast.Symbol) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || call.TypeArguments != nil || callee.Kind != ast.KindPropertyAccessExpression || callee.AsPropertyAccessExpression().QuestionDotToken != nil {
		return false
	}
	receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) || l.valueSymbol(receiver) != array {
		return false
	}
	return l.overloadIntrinsicMethod(callee, "ReadonlyArray", "slice") || l.overloadIntrinsicMethod(callee, "Array", "slice")
}

// overloadFilterParameterRead is whether reference, an identifier naming the parameter array, only reads
// it where it's written: its length, an element at a number, the receiver of its own slice, a test of it
// against undefined or for truthiness, or a return's value. Anything else, a write through it, an alias, a
// delete or passing it on, is a use no proof here reaches.
func (l *lowering) overloadFilterParameterRead(reference *ast.Node, array *ast.Symbol) bool {
	child, parent := reference, reference.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	if parent == nil {
		return false
	}
	deleted := func(access *ast.Node) bool {
		above := access.Parent
		for above != nil && above.Kind == ast.KindParenthesizedExpression {
			above = above.Parent
		}
		return above != nil && above.Kind == ast.KindDeleteExpression
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		if access.Expression != child || access.QuestionDotToken != nil {
			return false
		}
		switch parent.Name().Text() {
		case "length":
			return !overloadArrayWritten(reference) && !deleted(parent)
		case "slice":
			call := parent.Parent
			for call != nil && call.Kind == ast.KindParenthesizedExpression {
				call = call.Parent
			}
			return call != nil && call.Kind == ast.KindCallExpression && ast.SkipParentheses(call.AsCallExpression().Expression) == parent && l.overloadFilterSlice(call, array)
		}
		return false
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == child && !overloadArrayWritten(reference) && !deleted(parent) && l.overloadFilterElementAccess(parent, array)
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == child
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
			return false
		}
		other := binary.Right
		if binary.Right == child {
			other = binary.Left
		}
		other = ast.SkipParentheses(other)
		return ast.IsIdentifier(other) && l.symbol(other) == l.checker.GetUndefinedSymbol()
	case ast.KindReturnStatement:
		return true
	}
	return false
}

// overloadFilterPredicateCall is whether reference, an identifier naming parameters[index], is the
// callee of a direct call of it with one argument element accepts, past parentheses, with no ?. and no
// type arguments.
func (l *lowering) overloadFilterPredicateCall(reference *ast.Node, parameters []*ast.Node, index int, element func(*ast.Node) bool) bool {
	child, parent := reference, reference.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child, parent = parent, parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == child && l.overloadParameterCall(parameters, parent, element) == index
}

// overloadFilterCertificate is whether certificate D serves overload's result: implementation's body has
// its shape, the overload's first parameter is a value of its result, since the parameter comes back as
// the caller passed it, and, where a fresh array comes back, the overload's result is an array whose
// element that parameter's element relates to. A fresh array is never undefined, so only the parameter's
// own undefined reaches the result. censusOverload asks it with the implementation's binders mapped to
// the overload's. That the predicate takes the element isn't judged here: censusOverload's parameter
// census, which runs first, relates the overload's predicate to the implementation's.
func (l *lowering) overloadFilterCertificate(implementation, overload *ast.Node) bool {
	shape, shaped := l.overloadFilterShapeOf(implementation)
	if !shaped {
		return false
	}
	declared := overload.Parameters()
	if len(declared) > len(implementation.Parameters()) || shape.array >= len(declared) {
		return false
	}
	for _, parameter := range declared {
		if parameter.AsParameterDeclaration().DotDotDotToken != nil {
			return false // A rest parameter takes any count of arguments, so a position says nothing.
		}
	}
	parameter := declared[shape.array]
	if !ast.IsIdentifier(parameter.Name()) {
		return false
	}
	given := l.concrete(l.censusCallableParameterType(l.symbol(parameter.Name())))
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload)))
	if !l.censusRelated(given, promised) {
		return false // The argument handed back isn't a value of the result the overload promises.
	}
	if !shape.fresh {
		return true
	}
	passed, held := l.definedMembers(given), l.definedMembers(promised)
	if len(passed) != 1 || len(held) != 1 || !l.isArrayType(passed[0]) || !l.isArrayType(held[0]) {
		return false
	}
	return l.censusRelated(l.elementTypeOfArray(passed[0]), l.elementTypeOfArray(held[0]))
}
