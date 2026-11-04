package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A narrowing to undefined outlives a call too, narrowed.go's mirror. The checker narrows a binding to
// undefined at an assignment or a check, and keeps it across a call that writes the binding again:
//
//	function fill(): void { later = 4; }
//	let later: number | undefined = undefined;
//	fill();
//	console.log(String(later));
//
// Node prints 4. A consumer that asks the checker for the read's type (String here, a template, ??, ===,
// typeof) sees only undefined and folds the read to it, keeping the read for its effects alone, so both
// backends printed undefined (#r0vzaaa). Reading the storage at its declared type means routing every
// such consumer, so until it's routed a read narrowed to a nullish unit (undefined, null, or the two),
// of a binding declared to hold more, that a function other than the one declaring it writes, is not
// yet. A binding only its own function writes can't change behind a call: the checker sees each of those
// writes in flow, and a call back into that function makes a binding of its own.
func (l *lowering) narrowedNullishReadGap(node *ast.Node) error {
	shorthand := node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment
	if node.Kind != ast.KindIdentifier || !l.isExpression(node) || (ast.IsDeclarationName(node) && !shorthand) || (ast.IsAssignmentTarget(node) && !logicalAssignmentTarget(node)) {
		return nil
	}
	symbol := l.bindingNamed(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 || !l.writtenAcrossFunctions(symbol) {
		return nil
	}
	absent, nullish := l.nullishUnit(l.checker.GetTypeAtLocation(node))
	if !nullish {
		return nil
	}
	if _, declaredNullish := l.nullishUnit(l.checker.GetTypeOfSymbol(symbol)); declaredNullish {
		return nil // Declared to hold nothing else, the narrowed value is the only one it can hold.
	}
	return l.notYet(node, "a read narrowed to "+absent+" of a binding a call may write")
}

func init() { refusalRules.register(3350, "narrowedNullishReadGap", (*lowering).narrowedNullishReadGap) }

// bindingNamed is the binding an identifier reads or writes. In { later } and ({ later } = source) the
// name is the literal's property too, and asked for its symbol the checker gives the property.
func (l *lowering) bindingNamed(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent == nil || identifier.Parent.Kind != ast.KindShorthandPropertyAssignment {
		return l.symbol(identifier)
	}
	symbol := l.checker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	if symbol == nil {
		return nil
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	return l.checker.GetExportSymbolOfSymbol(symbol)
}

// nullishUnit reports whether every member of t is undefined or null, and names what it holds: undefined,
// null, or undefined or null.
func (l *lowering) nullishUnit(t *checker.Type) (string, bool) {
	members := []*checker.Type{t}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		members = t.Types()
	}
	undefined, null := false, false
	for _, member := range members {
		switch {
		case member.Flags()&checker.TypeFlagsUndefined != 0:
			undefined = true
		case member.Flags()&checker.TypeFlagsNull != 0:
			null = true
		default:
			return "", false
		}
	}
	switch {
	case undefined && null:
		return "undefined or null", true
	case null:
		return "null", true
	}
	return "undefined", undefined
}

// writtenAcrossFunctions reports whether any assignment to symbol, ??= and ++ included, sits in a
// function, a class field's initializer or a static block other than the one its declaration sits in:
// code a call can run. One walk over every module of the program collects the answer for every symbol
// the first time it's asked; the answer is the program's alone, so nothing invalidates it.
func (l *lowering) writtenAcrossFunctions(symbol *ast.Symbol) bool {
	if l.crossFunctionWrites == nil {
		l.crossFunctionWrites = map[*ast.Symbol]bool{}
		var writes ast.Visitor
		writes = func(node *ast.Node) bool {
			if node.Kind == ast.KindIdentifier && ast.IsAssignmentTarget(node) {
				if written := l.bindingNamed(node); written != nil && written.ValueDeclaration != nil && callBoundary(node) != callBoundary(written.ValueDeclaration) {
					l.crossFunctionWrites[written] = true
				}
			}
			node.ForEachChild(writes)
			return false
		}
		seen := map[*ast.SourceFile]bool{}
		for _, root := range l.program.Files() {
			// Every module the program imports, not only the files it was given: a module's own
			// function can write the binding another module reads (ESMModuleOrder has no effects).
			order, _ := ESMModuleOrder(l.checker, root)
			for _, module := range order {
				if !seen[module] {
					seen[module] = true
					module.AsNode().ForEachChild(writes)
				}
			}
		}
	}
	return l.crossFunctionWrites[symbol]
}

// callBoundary is the nearest code around node that runs when something calls or makes it, rather than
// in the flow around it: a function, a class field's initializer, which runs as its class is made, or a
// static block. Its source file where there is none.
func callBoundary(node *ast.Node) *ast.Node {
	for at := node.Parent; at != nil; at = at.Parent {
		if ast.IsFunctionLike(at) || at.Kind == ast.KindClassStaticBlockDeclaration || at.Kind == ast.KindPropertyDeclaration || at.Kind == ast.KindSourceFile {
			return at
		}
	}
	return nil
}
