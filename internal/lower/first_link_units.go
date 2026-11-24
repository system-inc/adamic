package lower

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// A first-link lowering's unit is what a trap stubs whole (first_link_traps.go): the innermost function
// with a body that a node lies in, or, outside any function, the module's top-level statement holding it.
// A unit trapped once stays trapped: it and everything inside it lower as its stub from then on.

// firstLinkFunctionUnit reports whether node is a function with a body, which is a unit of its own.
// A declaration without a body, an overload's or an abstract method's, lowers nothing, so it's no unit,
// and a node in it belongs to what holds it.
func firstLinkFunctionUnit(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Body() != nil
	}
	return false
}

// firstLinkUnitOf is the unit node lies in, node itself included: the innermost function with a body, or
// else the top-level statement, the ancestor whose parent is the module. It's nil for a node outside any
// statement, the module itself among them.
func firstLinkUnitOf(node *ast.Node) *ast.Node {
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if firstLinkFunctionUnit(ancestor) || ancestor.Parent != nil && ancestor.Parent.Kind == ast.KindSourceFile {
			return ancestor
		}
	}
	return nil
}

// firstLinkTrappedAt is the stub of the trapped unit node lies in, node itself included, if any is.
func (l *lowering) firstLinkTrappedAt(node *ast.Node) (ir.Panic, bool) {
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if stub, trapped := l.firstLinkStubs[ancestor]; trapped {
			return stub, true
		}
	}
	return ir.Panic{}, false
}

// firstLinkInside stops an expression that lies inside a trapped unit, so the code of a unit the refusal
// pass or an earlier stop trapped never lowers by a road other than its body's, a class's field
// initializers among them, and the unit lowering it traps in its turn. A function that is itself the
// trapped unit is let through, since making its value is outside it, and its body lowers as its stub.
func (l *lowering) firstLinkInside(node *ast.Node) error {
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor == node && firstLinkFunctionUnit(node) || !l.firstLinkTrapped[ancestor] {
			continue
		}
		return l.notYet(node, "code inside "+l.firstLinkUnitName(ancestor)+", a unit already trapped")
	}
	return nil
}

// firstLinkUnitName names a unit as a Trap's Unit does: its module's path, then its declaration chain, or
// "module" for a top-level statement that declares nothing.
func (l *lowering) firstLinkUnitName(unit *ast.Node) string {
	source := l.sourceIdentity(unit)
	if len(source.Declaration) == 0 {
		return source.Module + " module"
	}
	return firstLinkUnit(source)
}

// firstLinkPlace is the unit a stop's site lies in, for a stop that names its site only as a Where,
// file:line:col in load.Program's form: the line counted by ECMAScript's line terminators, the column in
// UTF-16 units, both from 1. It's nil for a site in another file, one the module doesn't hold, or one in
// no unit, a file-level directive's among them.
func (l *lowering) firstLinkPlace(module *ast.SourceFile, where string) *ast.Node {
	file, line, column, parsed := firstLinkSite(where)
	// The module's own Where names its file as every site in it does.
	own, _, _, ownParsed := firstLinkSite(l.program.Where(module.AsNode()))
	if !parsed || !ownParsed || file != own {
		return nil
	}
	lineStarts := scanner.GetECMALineStarts(module)
	if line < 1 || line > len(lineStarts) || column < 1 {
		return nil
	}
	text := module.Text()
	position := int(lineStarts[line-1])
	for units := 1; units < column; {
		if position >= len(text) {
			return nil
		}
		character, size := utf8.DecodeRuneInString(text[position:])
		// A character past the Basic Multilingual Plane is two UTF-16 units, a surrogate pair.
		units++
		if character > 0xFFFF {
			units++
		}
		position += size
	}
	// The innermost node whose first token starts at or before the position and that ends after it. A
	// node's range takes in the trivia before its first token, which a site never points into.
	innermost := module.AsNode()
	for {
		var inner *ast.Node
		innermost.ForEachChild(func(child *ast.Node) bool {
			if scanner.GetTokenPosOfNode(child, module, false) <= position && position < child.End() {
				inner = child
				return true
			}
			return false
		})
		if inner == nil {
			return firstLinkUnitOf(innermost)
		}
		innermost = inner
	}
}

// firstLinkSite reads a Where, file:line:col, into its parts.
func firstLinkSite(where string) (file string, line, column int, parsed bool) {
	columnAt := strings.LastIndexByte(where, ':')
	if columnAt < 0 {
		return "", 0, 0, false
	}
	lineAt := strings.LastIndexByte(where[:columnAt], ':')
	if lineAt < 0 {
		return "", 0, 0, false
	}
	line, lineErr := strconv.Atoi(where[lineAt+1 : columnAt])
	column, columnErr := strconv.Atoi(where[columnAt+1:])
	if lineErr != nil || columnErr != nil {
		return "", 0, 0, false
	}
	return where[:lineAt], line, column, true
}
