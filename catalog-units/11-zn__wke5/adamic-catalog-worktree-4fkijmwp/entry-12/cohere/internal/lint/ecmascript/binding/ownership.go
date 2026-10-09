// Package binding answers whether an identifier names something this file chose.
//
// Every rule that judges an identifier's spelling has to ask this before it reports, because
// renaming a name somebody else chose changes what the code reaches for rather than what it calls
// something. A census found two implementations in one package, one word apart in their names and
// four kinds apart in their answers, neither able to see the other.
package binding

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// IsForeignName reports whether an identifier names something this file did not choose, so a rule
// judging the spelling has no standing to complain about it.
//
// # The positions, and what each one is
//
//	thing.name          the property half of a read; the name belongs to whatever `thing` is
//	{ name: 1 }         an object-literal key, which shapes a value somebody else consumes
//	import { name }     an imported binding, in all four of its spellings
//	interface { name }  a property signature, which is very often an external wire surface
//	A.name              the right half of a qualified name, which is the other module's spelling
//	name in a type      a type reference, naming a type this file may not own
//	<name />, name=     a JSX tag or attribute
//
// # A property read off `this` is deliberately NOT foreign
//
// The property belongs to the class being read, which this file owns, and its declaration is a
// property declaration that no member check protects. Treating it like any other property read is
// how a class ended up declaring `maximumBackoff` while reading `this.maximumBackoff`, measured in
// the abbreviation rule that carries this exemption today.
//
// The sibling ambiguous-identifier rule exempted it, and that was the one place the two lifted
// implementations genuinely disagreed rather than merely differing in coverage.
//
// **Decided once, by Kirk, rather than left to two files that could not see each other.** The
// reasoning is about ownership rather than about syntax: `document.cookie` is foreign because nobody
// here owns `document`, while `this.e` is a name we chose, on an object we own, in a file we
// control. Renaming it is a rename, not a reach for somebody else's API, so exempting it would mean
// a naming rule cannot see the names we actually control. The `BackoffTask` receipt is the evidence
// rather than the argument: a class declared `maximumBackoff` while reading `this.maximumBackoff`,
// and the rule was blind to it.
//
// The consequence is intended rather than tolerated: `consistency-no-ambiguous-identifier` reports
// `this.e` and its one-letter siblings now, where it did not before.
//
// A later reader who finds this surprising should not helpfully re-add the carve-out. The question
// was asked and answered; re-adding it re-opens a disagreement that cost two rules their agreement
// for as long as neither could see the other.
//
// # JSX has no exemption upstream and needs one here
//
// The ESTree originals never wrote a JSX case, because their parser gives a JSX name its own node
// type and it never reaches an identifier visitor. In typescript-go a JSX tag and a JSX attribute
// are plain `KindIdentifier` nodes, so reproducing only the written-down exemptions cost 3,081 false
// findings on a real tree. That is the general shape of this class of divergence: an exemption the
// original gets free from its AST has to be written down in ours.
func IsForeignName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// The property half of `thing.name`, but not the object half: in `name.other` the object is
		// a binding this file may well own.
		access := parent.AsPropertyAccessExpression()
		if access == nil || access.Name() != node {
			return false
		}
		// See the `this` note above. A property read off `this` is ours.
		return access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword

	case ast.KindPropertyAssignment:
		// The key half of `{ name: 1 }`. A shorthand value is a real reference and is judged.
		assignment := parent.AsPropertyAssignment()
		return assignment != nil && assignment.Name() == node

	case ast.KindBindingElement:
		// The key half of `{ params: parameters }` in a destructuring pattern: the object's spelling,
		// renamed away on the spot, and the same ESTree `Property` key the originals skip. A
		// shorthand `{ params }` has no property name and declares a binding, so it is still judged.
		// Found on api's MappedJoins.test.ts and FileStorageVideoDerivativeService.ts (#dx1vrfm).
		element := parent.AsBindingElement()
		return element != nil && element.PropertyName == node

	case ast.KindImportSpecifier, ast.KindExportSpecifier, ast.KindNamespaceImport,
		ast.KindImportClause:
		// An imported name is external surface this file cannot rename, in all four spellings:
		// `import { name }`, `export { name }`, `import * as name`, and `import name from`.
		return true

	case ast.KindPropertySignature:
		// An interface or type-literal key shapes an external surface often enough that the
		// originals skip the whole family: HTML properties, the cookie specification, wire fields.
		signature := parent.AsPropertySignatureDeclaration()
		return signature != nil && signature.Name() == node

	case ast.KindQualifiedName:
		// The right half of `A.B.C` is the other module's spelling. The left root is still judged,
		// since a locally declared namespace is ours.
		qualified := parent.AsQualifiedName()
		return qualified != nil && qualified.Right == node

	case ast.KindTypeReference:
		// A type this file may not own. The import is already exempt above; this covers the uses.
		reference := parent.AsTypeReferenceNode()
		return reference != nil && reference.TypeName == node

	case ast.KindJsxOpeningElement, ast.KindJsxClosingElement, ast.KindJsxSelfClosingElement,
		ast.KindJsxAttribute:
		// See the JSX note above. An intrinsic element is named by HTML and an attribute is named by
		// whatever component declares it; neither is a binding this file owns.
		return true
	}
	return false
}
