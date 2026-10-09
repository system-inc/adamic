package module

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// DeclaresTypeNamed reports whether a file declares a type by this name, anywhere in it.
//
// This is the shadow test every rule needs before reporting a global type name: a file that
// declares its own `Function` or `Number` has redefined the name, and reporting it would flag a
// type the author owns.
//
// # Whole file rather than top level, which was a real false positive
//
// oxc's own passing case for `no-unsafe-function-type` is a block-scoped `type Function = () =>
// void;` inside braces, and a scan of top-level statements alone reports it as a violation. That
// shipped in this tree before the corpus was read.
//
// # Deliberately coarser than scope analysis
//
// A declaration anywhere in the file suppresses the rule for the whole file, where oxc suppresses it
// only within the declaring scope. The direction is chosen rather than accidental: a missed finding
// in a file that defines its own `Function` costs nothing, and a false positive on a type the author
// owns costs trust in the rule. Narrowing it correctly needs scope information the checker shim does
// not expose.
//
// # An enum counts, and that was measured rather than assumed
//
// Two of the three implementations this replaces counted interface, type alias and class; the third
// also counted enum, and nobody had stated which was right. An enum declares a value and a TYPE of
// the same name, so it shadows exactly as an interface does.
//
// Established with the checker rather than by reading the specification: in `enum Number { A }
// let value: Number;` the annotation resolves to a declaration in the linted file, while in
// `let value: Number;` alone it resolves into a library declaration file. So the two-kind walks were
// reporting a name the author owns, in both rules that used one.
//
// A variable or function of the same name does not count, because neither shadows the type in a
// type position, so a file containing one is still linted.
func DeclaresTypeNamed(sourceFile *ast.SourceFile, name string) bool {
	if sourceFile == nil || name == "" {
		return false
	}
	found := false
	forEachDeclaredTypeName(sourceFile.AsNode(), func(declared string) bool {
		if declared == name {
			found = true
		}
		return found
	})
	return found
}

// DeclaredTypeNames returns which of the given names a file declares as a type.
//
// One walk answers for every name rather than one walk per name, since the file is the expensive
// part and the map lookup is not. A caller asking about a single name wants `DeclaresTypeNamed`,
// which stops at the first hit.
func DeclaredTypeNames(sourceFile *ast.SourceFile, names map[string]bool) map[string]bool {
	declared := map[string]bool{}
	if sourceFile == nil {
		return declared
	}
	forEachDeclaredTypeName(sourceFile.AsNode(), func(name string) bool {
		if names[name] {
			declared[name] = true
		}
		return false
	})
	return declared
}

// AllDeclaredTypeNames returns every name a file declares as a type.
//
// For a caller that collects rather than tests, such as one refusing to rename onto a name that is
// already taken.
func AllDeclaredTypeNames(sourceFile *ast.SourceFile) map[string]bool {
	declared := map[string]bool{}
	if sourceFile == nil {
		return declared
	}
	forEachDeclaredTypeName(sourceFile.AsNode(), func(name string) bool {
		declared[name] = true
		return false
	})
	return declared
}

// forEachDeclaredTypeName walks a file, calling back with each declared type name. The callback
// returning true stops the walk, which is what lets a single-name caller exit at the first hit.
func forEachDeclaredTypeName(node *ast.Node, callback func(name string) bool) {
	stopped := false

	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil || stopped {
			return
		}

		var name *ast.Node
		switch current.Kind {
		case ast.KindInterfaceDeclaration:
			name = current.AsInterfaceDeclaration().Name()
		case ast.KindTypeAliasDeclaration:
			name = current.AsTypeAliasDeclaration().Name()
		case ast.KindClassDeclaration:
			name = current.AsClassDeclaration().Name()
		case ast.KindEnumDeclaration:
			// See the enum note on DeclaresTypeNamed: an enum declares a type as well as a value,
			// measured with the checker rather than assumed.
			name = current.AsEnumDeclaration().Name()
		}
		if name != nil && name.Kind == ast.KindIdentifier {
			if callback(name.Text()) {
				stopped = true
				return
			}
		}

		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return stopped
		})
	}
	visit(node)
}
