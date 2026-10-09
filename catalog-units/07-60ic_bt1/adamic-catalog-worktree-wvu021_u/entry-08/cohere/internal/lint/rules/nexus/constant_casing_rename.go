package nexus

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// unbindableGlobals are names that pass as identifiers and still must not be the target of a
// rename. `arguments` and `eval` are a strict-mode error as a binding; the other three are legal and
// would silently redefine a value every reader assumes is the built-in.
var unbindableGlobals = map[string]bool{
	"arguments": true, "eval": true, "undefined": true, "NaN": true, "Infinity": true,
}

// fileLocalRenameFix renames a constant nothing outside its file can name, at its declaration and
// at every reference, as ONE edit.
//
// One edit rather than one per site, because the engine applies a diagnostic's fixes as independent
// proposals and resolves overlap per proposal. A rename split into pieces can land half-applied when
// another rule's fix overlaps one site, and the half that lands parses, so the parse guard passes a
// file that no longer compiles. A single replacement spanning the first site to the last is all or
// nothing, which is the same trade ESLint's fixer makes when it merges a rule's fixes.
//
// References are resolved through the checker rather than by matching text, so a shadowing binding
// with the same spelling, a property named the same, and an object key are all left alone. Three
// shapes need more than a text swap or refuse outright:
//
//	{ OrderColumns }          shorthand: the key is the property's name and must survive, so it
//	                          becomes { OrderColumns: orderColumns }
//	export { OrderColumns }   the constant leaves the file under this spelling after all, so a rename
//	                          would change somebody else's import; no fix
//	type OrderColumns = ...   a type or namespace merged into the same symbol answers to the old
//	                          name in type positions; no fix
//
// And the new name must be free. Any identifier already spelled that way anywhere in the file
// refuses the fix, which is broader than a scope walk needs to be and is the version that cannot be
// wrong: the rename can then neither capture a reference to something else nor be captured by it.
// A refused fix still reports, so the author renames by hand where the rule cannot prove safety.
func fileLocalRenameFix(ctx rule.Context, name *ast.Node, newName string) (rule.Fix, bool) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil || name == nil || newName == "" {
		return rule.Fix{}, false
	}
	oldName := name.Text()
	// The scanner answers the character question only (it calls `true` a valid identifier), so the
	// keywords come from `reservedWords`, the set this package already keeps for the converter.
	// `internal/rename` pairs the same two checks, and importing it would take this rule package off
	// the leaf (`TestRulePackagesStayLeaves`). No fixture reaches the keyword half, and a mutation
	// confirms it: the rule returns before reporting any name whose lowercase is a keyword
	// (`escapesAReservedWord`). It stays so the fix's safety does not rest on a check in another file.
	if !scanner.IsValidIdentifier(newName) || reservedWords[newName] || unbindableGlobals[newName] {
		return rule.Fix{}, false
	}

	declared := ctx.TypeChecker.GetSymbolAtLocation(name)
	// One declaration, or something merged into the name (a type alias, a namespace) answers to the
	// old spelling where the checker's value lookup cannot see it.
	if declared == nil || len(declared.Declarations) != 1 {
		return rule.Fix{}, false
	}

	type siteEdit struct {
		textRange core.TextRange
		text      string
	}
	var edits []siteEdit
	refused := false

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if refused || node == nil {
			return refused
		}
		if node.Kind == ast.KindIdentifier {
			switch node.Text() {
			case newName:
				refused = true
				return true
			case oldName:
				edit, isReference, refuses := renameSite(ctx, node, name, declared, oldName, newName)
				if refuses {
					refused = true
					return true
				}
				if isReference {
					edits = append(edits, siteEdit{textRange: rule.TokenRange(ctx.SourceFile, node), text: edit})
				}
			}
		}
		node.ForEachChild(visit)
		return refused
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)

	// The declaration's own name is always a site, since the walk visits every identifier and the
	// first comparison in `renameSite` is identity, so the length test cannot fire and no fixture
	// reaches it. It stays as the guard on `edits[0]` below, so a future change to the walk fails as
	// a withheld fix rather than as a panic that loses the file. (A name already in camelCase never
	// gets here either: the rule reports only the other casing, and a same-named target would trip
	// the collision case on the declaration itself.)
	if refused || len(edits) == 0 {
		return rule.Fix{}, false
	}

	sort.Slice(edits, func(first int, second int) bool {
		return edits[first].textRange.Pos() < edits[second].textRange.Pos()
	})

	sourceText := ctx.SourceFile.Text()
	start := edits[0].textRange.Pos()
	end := edits[len(edits)-1].textRange.End()
	var replacement strings.Builder
	cursor := start
	for _, edit := range edits {
		replacement.WriteString(sourceText[cursor:edit.textRange.Pos()])
		replacement.WriteString(edit.text)
		cursor = edit.textRange.End()
	}
	replacement.WriteString(sourceText[cursor:end])

	return rule.ReplaceRange(core.NewTextRange(start, end), replacement.String()), true
}

// renameSite decides one identifier spelled like the constant: whether it is the constant, what it
// becomes, and whether its position forbids the rename altogether.
func renameSite(
	ctx rule.Context,
	identifier *ast.Node,
	declarationName *ast.Node,
	declared *ast.Symbol,
	oldName string,
	newName string,
) (text string, isReference bool, refuses bool) {
	if identifier == declarationName {
		return newName, true, false
	}

	parent := identifier.Parent
	if parent != nil {
		switch parent.Kind {
		case ast.KindShorthandPropertyAssignment:
			if parent.Name() == identifier {
				if ctx.TypeChecker.GetShorthandAssignmentValueSymbol(parent) != declared {
					return "", false, false
				}
				return oldName + ": " + newName, true, false
			}
		case ast.KindExportSpecifier:
			if ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(parent) == declared {
				return "", false, true
			}
			return "", false, false
		}
	}

	if ctx.TypeChecker.GetSymbolAtLocation(identifier) == declared {
		return newName, true, false
	}
	return "", false, false
}
