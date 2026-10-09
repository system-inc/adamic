package reference

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// IsValueReference says whether an identifier is a value reference rather than something that merely
// spells the same name.
//
// Lifted from `no-restricted-globals`, where it was refined against real trees, when the reference
// tracker below needed the same question asked of every identifier in a file (#jjfa7qb). ESLint's
// scope analysis only ever hands a rule references; walking identifiers directly means meeting
// everything else first, and this is what tells them apart.
//
// Two separate exclusions live here and both are load bearing.
//
// A TYPE POSITION is ESLint's `TYPE_NODES` check. Measured, its five ESTree kinds are three of ours:
// `KindTypeReference` covers a plain annotation, an `implements` clause and an `extends` clause
// alike, `KindTypeQuery` is `typeof X`, and `KindQualifiedName` is `NS.Test`.
//
// A NAME THAT IS NOT A REFERENCE AT ALL: the declaration's own name, a property key, a member's
// property half, an import or export specifier, and a label, none of which is a reference to
// anything. ESLint's clean `foo.bar` restricting `bar` is the case that states the member half, and
// `import foo from 'bar'` restricting `foo` is the one that states the specifier.
//
// Two more arms were written first and left out: one declining a name being introduced, one declining
// a property key. A caller that asks where the symbol it reads is declared already declines every
// shape either could catch, because each resolves to a declaration right there in source. Measured
// under no-restricted-globals rather than argued: neutralising either arm alone survived, neutralising
// one with that caller's check failed 22 lines, and sixteen declaration shapes and seven key shapes
// gave byte-identical output with and without them. The declaration-name test above came back only
// because a declaration file declares nothing in source.
func IsValueReference(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// A declaration's own name introduces it rather than reading it. In a source file a shadow check
	// declines it anyway, since the symbol is declared right there; in a declaration file it sees
	// only a declaration file, so Base's photon_rs_bg.d.ts:41 reported the parameter `top` of a
	// declared function under no-restricted-globals. A shorthand property and a local export
	// specifier are declarations whose name is also a read, so they stay, each resolved to what it
	// reads by ReadSymbol.
	if ast.IsDeclarationName(node) && parent.Kind != ast.KindShorthandPropertyAssignment &&
		parent.Kind != ast.KindExportSpecifier {
		return false
	}

	switch parent.Kind {
	// ESLint's TYPE_NODES.
	case ast.KindTypeReference, ast.KindTypeQuery, ast.KindQualifiedName:
		return false

	// The property half of a member access. `foo.bar` restricting `bar` is one of ESLint's clean
	// cases, and the object half must still be read, which is why this compares rather than
	// declining the whole kind.
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() != node

	// A label is not a value.
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return false

	// An intrinsic JSX tag, `<stop>` or `<my-element>`, is named by HTML rather than read from scope;
	// a capitalized tag is a component and reads its binding. ESLint's parser gives a JSX name its
	// own node type, so this exclusion is free there and has to be written here, where a tag is a
	// plain identifier: the svg `<stop>` reported three times per gradient (#g5b8q7e).
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxClosingElement:
		return !scanner.IsIntrinsicJsxName(node.Text())

	// A JSX attribute name is the component's spelling, never a reference.
	case ast.KindJsxAttribute:
		return false

	// The key half of a destructuring pattern, `{ open: externalOpen }`, names a property of the
	// object being destructured. A shadow check cannot decline it the way it declines an object
	// literal's key: the key resolves to that object type's property, which lives in a declaration
	// file whenever the type does. Collapsible.tsx:43 destructures Radix's props and reported `open`.
	// The default, `{ label = name }`, is a real read.
	case ast.KindBindingElement:
		return parent.AsBindingElement().PropertyName != node

	// An import specifier's imported name, and any name in an export-from specifier, names another
	// module's export rather than a binding in this file. It resolves to that module's declaration,
	// which is a declaration file for any package: Base's
	// `export { print as printGraphQlNode } from 'graphql'` reported once `print` was restricted
	// (#wajqfd1). The local name an import introduces is declared here, so declining it too changes
	// nothing.
	case ast.KindImportSpecifier:
		return false
	case ast.KindExportSpecifier:
		return exportSpecifierReads(parent, node)
	}
	return true
}

// exportSpecifierReads reports whether this name in an export specifier reads a binding of this file.
//
// An export-from specifier reads nothing here. A local one reads the name it exports, which is its
// property name when it renames and its only name when it does not; the name it renames to is the
// exported name, a new spelling rather than a read. ESLint 10 under @typescript-eslint/parser agrees
// on each: `export { print }` and `export { print as printPage }` read `print`, `export { other as
// print }` and every export-from do not.
func exportSpecifierReads(specifier *ast.Node, node *ast.Node) bool {
	if declaration := ast.FindAncestorKind(specifier, ast.KindExportDeclaration); declaration != nil &&
		declaration.AsExportDeclaration().ModuleSpecifier != nil {
		return false
	}
	propertyName := specifier.AsExportSpecifier().PropertyName
	return propertyName == nil || propertyName == node
}

// ReadSymbol returns the symbol a value reference reads.
//
// Two positions answer differently from `GetSymbolAtLocation`. `{ status }` reads `status`, and the
// plain accessor answers with the literal's own property, declared right there in source, so a
// global read looked like a local one: WisdomGateItems.ts:304's bug under no-restricted-globals,
// written as an object instead of a template. A local export specifier is the same shape:
// `export { print }` reads `print`, and the plain accessor answers with the specifier's own export
// symbol. The value symbol and the local target are the bindings read.
func ReadSymbol(typeChecker *checker.Checker, node *ast.Node) *ast.Symbol {
	parent := node.Parent
	switch {
	case parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node:
		return typeChecker.GetShorthandAssignmentValueSymbol(parent)
	case parent != nil && parent.Kind == ast.KindExportSpecifier:
		return typeChecker.GetExportSpecifierLocalTargetSymbol(parent)
	}
	return typeChecker.GetSymbolAtLocation(node)
}
