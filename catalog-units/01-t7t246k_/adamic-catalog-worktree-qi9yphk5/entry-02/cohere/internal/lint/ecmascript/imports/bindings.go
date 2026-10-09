package imports

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Bindings are the local names a single import declaration introduces.
//
// Three shapes in one statement, and they live in three different places in this AST rather than in
// one flat list. `import Thing, * as All from 'm'` binds a default and a namespace; `import { a, b }`
// binds two named ones; `import 'm'` binds nothing at all.
//
// The ESTree originals see one `specifiers` array holding all three kinds together and filter it,
// which is why a port reading only one field pins half the forms and looks complete. That mistake
// has been made and corrected once in this tree already, in `import-require-module-alias`.
type Bindings struct {
	// Default is the local name of a default import, or nil.
	Default *ast.Node
	// Namespace is the local name of a namespace import, or nil.
	Namespace *ast.Node
	// Named holds each named import's specifier node, which carries both the imported name and the
	// local one when they differ.
	Named []*ast.Node
}

// BindingsOf reads the local names an import declaration introduces.
//
// Three research passes on separate `@next/next` rules each reported this as absent from every
// shelf while two rule packages held near-identical inline copies, which is what turned it from a
// tidy-up into a lift.
//
// A side-effect import (`import 'm'`) answers with an empty Bindings rather than nil, so a caller
// can read the fields without checking the container first. That matters because the shape it
// describes is real: a module imported purely for effect binds nothing, and a rule asking "what did
// this bring into scope" should get an honest empty answer rather than a nil to guard.
func BindingsOf(node *ast.Node) Bindings {
	var bindings Bindings
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return bindings
	}
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return bindings
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil {
		return bindings
	}

	bindings.Default = clause.Name()

	if clause.NamedBindings == nil {
		return bindings
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		bindings.Namespace = clause.NamedBindings
	case ast.KindNamedImports:
		elements := clause.NamedBindings.AsNamedImports().Elements
		if elements == nil {
			return bindings
		}
		bindings.Named = append(bindings.Named, elements.Nodes...)
	}
	return bindings
}

// LocalNameOfDefaultImport returns the local name a file binds a module's default export to.
//
// This is the narrow question three `@next/next` rules ask: `no-script-component-in-head` needs the
// local name of `next/head`'s default so it can match JSX against it, and two siblings need the same
// for `next/script` and `next/document`.
//
// **The module specifier is compared exactly, and that is not what upstream does.** oxc's
// `no-script-component-in-head` never resolves `next/script` at all: it string-compares the child
// element's name against `Script`, so it fires on any local component called `Script` and misses
// `import S from 'next/script'`. That is a defect rather than a judgment, and a port using this
// helper is more correct than its original. Recorded here because a reader comparing the two will
// otherwise assume ours is the one that is wrong.
//
// The second return separates "no such import" from an import whose local name is empty, which the
// parser can produce on malformed input.
func LocalNameOfDefaultImport(node *ast.Node, specifier string) (string, bool) {
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return "", false
	}
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return "", false
	}
	if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return "", false
	}
	if declaration.ModuleSpecifier.Text() != specifier {
		return "", false
	}

	defaultBinding := BindingsOf(node).Default
	if defaultBinding == nil || defaultBinding.Kind != ast.KindIdentifier {
		return "", false
	}
	return defaultBinding.Text(), true
}

// LocalNameOfNamedImport returns the local name a file binds one of a module's named exports to.
//
// This is the aliasing question, and it is a different question from `LocalNameOfDefaultImport`
// rather than a variant of it. `import { Head } from 'next/document'` and
// `import { Head as PageHead } from 'next/document'` bind the same export to two different local
// names, and a rule matching JSX against the imported spelling reports nothing on the second while
// looking like it works on the first.
//
// Three research passes each reported this as missing from every shelf and two of them proposed
// building it from scratch. It is instead three lines over `BindingsOf`, which already returns the
// named specifier nodes precisely because they carry both halves of the alias. Recorded because the
// reports were right that no such function existed and wrong that the data was not already there,
// and the next pass to want an import question should read `Bindings` before concluding it is absent.
//
// `imported` is the name as the source module exports it, never the local alias. Asking for the
// local name would make the function a no-op.
//
// The returned node is the local binding identifier, which is what a caller matches references
// against, and it is nil when the import is absent. A specifier with no alias carries its local name
// in `Name()` and a nil `PropertyName`, so the unaliased form falls out of the same read.
func LocalNameOfNamedImport(node *ast.Node, specifier string, imported string) *ast.Node {
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return nil
	}
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ModuleSpecifier == nil {
		return nil
	}
	if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
		return nil
	}
	if declaration.ModuleSpecifier.Text() != specifier {
		return nil
	}

	// `Named` is populated only from a `KindNamedImports` element list, so every entry is already an
	// import specifier. A kind check here reads as defensive and is unreachable: a mutation removing
	// it changed no fixture, which is how it was found rather than assumed.
	for _, element := range BindingsOf(node).Named {
		if ImportedNameOf(element) == imported {
			return element.Name()
		}
	}
	return nil
}

// ImportedNameOf returns the name a named-import specifier reads from the source module.
//
// The alias lives in `PropertyName` and is nil when there is none, so `{ Head }` and
// `{ Head as PageHead }` both answer `Head` while their local names differ. Reading `Name()` alone
// answers `Head` and `PageHead`, which is the mistake this exists to stop: it silently matches the
// unaliased form and misses the aliased one, and the unaliased form is what every fixture is written
// with.
func ImportedNameOf(specifier *ast.Node) string {
	if specifier == nil || specifier.Kind != ast.KindImportSpecifier {
		return ""
	}
	if propertyName := specifier.PropertyName(); propertyName != nil {
		return propertyName.Text()
	}
	if name := specifier.Name(); name != nil {
		return name.Text()
	}
	return ""
}
