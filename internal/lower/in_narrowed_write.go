package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A write to a property that only an in test gave its object's type stores whatever it's given into the
// object's own slot (#yqbg9vz). On a single object type, `'extra' in view` narrows view to
// Plain & Record<"extra", unknown>, so tsc takes any value for view.extra:
//
//	const real = { name: 'n', extra: 5 };
//	const view: Plain = real;
//	if ('extra' in view) { view.extra = 'text'; }
//	console.log(String(real.extra + 1));
//
// Node prints text1, and natively the string went into real's number slot, which faulted under ASan.
// inProperty admits in as presence because reads stay dynamic; a write through the Record member has no
// slot type to be held to. So it's not yet (Adamic's ruling, Nov 11), not refused: the program is valid,
// Node's answer is right, and a dynamic keyed write could lift it. Every write form counts, as
// GetAssignmentTarget finds them: =, a compound or logical assignment, ++ and --, a destructuring target
// and a for-in or for-of head, through a name or a literal key. A read stays admitted.
//
// The property is judged by where the type gives it, not by the syntax around it: the receiver's
// narrowed type holds it through the library's Record<K, unknown>, the member tsc's narrowTypeByInKeyword
// intersects in, and the type the receiver is declared with doesn't (inNarrowedDeclaredType). A receiver
// declared Plain & Record<"extra", unknown> by its own annotation isn't this road.
func (l *lowering) inNarrowedWriteGap(node *ast.Node) error {
	var receiver *ast.Node
	var names []string
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = node.AsPropertyAccessExpression().Expression
		names = []string{node.AsPropertyAccessExpression().Name().Text()}
	case ast.KindElementAccessExpression:
		receiver = node.AsElementAccessExpression().Expression
		names = l.unionWriteKeys(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression))
	default:
		return nil
	}
	if !ast.IsAssignmentTarget(node) {
		return nil
	}
	narrowed := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	declared := l.inNarrowedDeclaredType(receiver)
	for _, name := range names {
		if name == "" || !l.inNarrowingAdds(narrowed, name) {
			continue // An index key names no property a Record member could add.
		}
		if declared != nil && l.checker.GetPropertyOfType(declared, name) != nil {
			continue
		}
		return l.notYet(node, "a write to '"+name+"', which only an `in` test added to the type")
	}
	return nil
}

func init() {
	refusalRules.register(3450, "inNarrowedWriteGap", (*lowering).inNarrowedWriteGap)
}

// inNarrowingAdds reports whether a part of a receiver's narrowed type, through its unions and
// intersections, is the library's Record<K, unknown> holding name: the member an in test on a single
// object type intersects in for a key no part of the type declares.
func (l *lowering) inNarrowingAdds(narrowed *checker.Type, name string) bool {
	if narrowed.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, part := range narrowed.Types() {
			if l.inNarrowingAdds(part, name) {
				return true
			}
		}
		return false
	}
	alias := narrowed.Alias()
	if alias == nil || alias.Symbol().Name != "Record" || !l.librarySymbol(alias.Symbol()) || len(alias.TypeArguments()) != 2 {
		return false
	}
	return alias.TypeArguments()[1].Flags()&checker.TypeFlagsUnknown != 0 && l.checker.GetPropertyOfType(narrowed, name) != nil
}

// inNarrowedDeclaredType is the type a narrowable reference has before flow narrows it: a variable's or
// a parameter's declared type, or a property's, read through the name that reaches it. Nil for any other
// receiver, `this` among them, and then a write the Record member gives stops, failing closed.
func (l *lowering) inNarrowedDeclaredType(reference *ast.Node) *checker.Type {
	reference = ast.SkipParentheses(reference)
	var symbol *ast.Symbol
	switch reference.Kind {
	case ast.KindIdentifier:
		symbol = l.symbol(reference)
	case ast.KindPropertyAccessExpression:
		symbol = l.checker.GetSymbolAtLocation(reference.AsPropertyAccessExpression().Name())
	default:
		return nil
	}
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsVariable|ast.SymbolFlagsProperty) == 0 {
		return nil
	}
	return l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(symbol))
}
