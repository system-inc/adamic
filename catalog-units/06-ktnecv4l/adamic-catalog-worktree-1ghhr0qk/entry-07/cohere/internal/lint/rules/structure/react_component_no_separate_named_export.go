package structure

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactComponentNoSeparateNamedExportText is the rule's message, whose wording lives in
// `policy/messages/react-component-no-separate-named-export.json`.
var reactComponentNoSeparateNamedExportText = policy.MessageOf("structure/react-component-no-separate-named-export", "noSeparateNamedExport")

// messageNoSeparateNamedExport is the finding, rendered when it is reported so the text comes from
// the current catalog.
func messageNoSeparateNamedExport() rule.Message {
	return rule.Message{Id: reactComponentNoSeparateNamedExportText.Id, Description: reactComponentNoSeparateNamedExportText.Render(nil)}
}

// ReactComponentNoSeparateNamedExport flags a bare export list naming a component the file declares.
//
//	valid:   export function Button() { ... }
//	valid:   export { Button } from './Button'          a re-export, not this file's declaration
//	invalid: function Button() { ... } export { Button }
//
// A re-export is deliberately exempt and it is the shape this codebase actually uses: an index file
// gathering components from elsewhere has nothing to move the export onto, since the declaration is
// in another file.
//
// Only components. A file exporting a type or a constant from a list at the bottom is a different
// convention question that this rule has no opinion on, and reporting it would make the rule about
// export style in general rather than about where a component announces itself.
//
// # The repair MOVES the export, it does not delete it
//
// The first version of this fixer deleted the statement and nothing else, on the belief that the
// build would break loudly and the fixer's convergence check would catch it. Neither holds: the
// engine checks that the rewritten file PARSES, and an unexported component parses perfectly. Every
// file importing it then fails to resolve, which is the same loss that unexported six interfaces
// across the structure library through `consistent-indexed-object-style`. Measured on
// `function Thing() {...} const helper = 1; export { Thing, helper };`, which it rewrote with neither
// name exported.
//
// So the repair now deletes the list AND puts `export ` on every declaration it named, as ONE edit.
// One edit rather than two because `edit.ProposalsFrom` treats a diagnostic's fixes independently,
// and a half-applied pair whose deletion landed without its insertion is the original defect again.
//
// It is offered only where every name in the list can carry the keyword without changing anything
// else, and otherwise the statement is reported with no repair:
//
//	a name that is not a component declared here    `helper` would lose its export, so declined
//	an aliased specifier `Thing as Other`            moving it changes the exported name
//	a type-only list or specifier                    `export type` is not a declaration modifier
//	a variable statement with several declarators    `export` would export its siblings too
//	a name declared more than once (overloads)       every signature would need the modifier
//	a declaration already carrying a modifier list   that includes `export` or `default`
var ReactComponentNoSeparateNamedExport = rule.Rule{
	Name: "structure/react-component-no-separate-named-export",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		declaredComponents, declarationStatements := componentNamesDeclaredIn(ctx.SourceFile.AsNode())
		if len(declaredComponents) == 0 {
			return nil
		}

		return rule.Listeners{
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()

				// A re-export has somewhere else to be, so there is nothing to move the export
				// onto. This is the shape an index file uses and it must stay silent.
				if declaration.ModuleSpecifier != nil {
					return
				}
				// The kind test cannot fail in practice and is kept for shape. The only clause that
				// is not a named-exports list is a namespace one (`export * as X from '...'`), and
				// every namespace clause carries a module specifier, which the re-export guard
				// above has already returned on. So no fixture can kill a mutant that drops it.
				//
				// Kept rather than reduced to the nil check, because the reason it is unreachable
				// lives four lines away and a reader would have to reconstruct it. Recorded so a
				// future sweep does not go looking for a fixture that cannot exist.
				if declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
					return
				}

				for _, element := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
					name := element.AsExportSpecifier().Name()
					if name == nil || name.Kind != ast.KindIdentifier {
						continue
					}
					if declaredComponents[name.Text()] {
						// Reported once per statement, not per specifier. The repair moves every
						// name in the statement onto its declaration, as one edit.
						fix, repairable := separateNamedExportRepair(ctx, node, declaredComponents,
							declarationStatements)
						if !repairable {
							ctx.ReportNode(node, messageNoSeparateNamedExport())
							return
						}
						ctx.ReportNodeWithFixes(node, messageNoSeparateNamedExport(), fix)
						return
					}
				}
			},
		}
	},
}

// separateNamedExportRepair builds the one edit that deletes the export list and puts `export ` on
// each declaration it named, or reports that no such edit is safe. See the rule's doc comment for
// the declines.
//
// The edit spans from the earliest insertion or deletion to the latest, and its text is that span
// with the insertions made and the statement removed. The statement's own range is `RemoveNode`'s,
// so the blank line it leaves is the same one the earlier fixer left.
func separateNamedExportRepair(ctx rule.Context, exportNode *ast.Node, declaredComponents map[string]bool,
	declarationStatements map[string][]*ast.Node) (rule.Fix, bool) {
	declaration := exportNode.AsExportDeclaration()
	if declaration.IsTypeOnly {
		return rule.Fix{}, false
	}

	type edit struct {
		start, end int
		text       string
	}
	removal := ctx.RemoveNode(exportNode)
	edits := []edit{{start: removal.Range.Pos(), end: removal.Range.End(), text: ""}}

	seen := map[*ast.Node]bool{}
	for _, element := range declaration.ExportClause.AsNamedExports().Elements.Nodes {
		specifier := element.AsExportSpecifier()
		if specifier.IsTypeOnly || specifier.PropertyName != nil {
			return rule.Fix{}, false
		}
		name := specifier.Name()
		if name == nil || name.Kind != ast.KindIdentifier || !declaredComponents[name.Text()] {
			return rule.Fix{}, false
		}
		statements := declarationStatements[name.Text()]
		if len(statements) != 1 {
			return rule.Fix{}, false
		}
		statement := statements[0]
		if seen[statement] {
			return rule.Fix{}, false
		}
		seen[statement] = true
		if modifiers := statement.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				if modifier.Kind == ast.KindExportKeyword || modifier.Kind == ast.KindDefaultKeyword {
					return rule.Fix{}, false
				}
			}
		}
		if statement.Kind == ast.KindVariableStatement {
			list := statement.AsVariableStatement().DeclarationList
			if list == nil || len(list.AsVariableDeclarationList().Declarations.Nodes) != 1 {
				return rule.Fix{}, false
			}
		}
		insertion := ctx.InsertBefore(statement, "export ")
		edits = append(edits, edit{start: insertion.Range.Pos(), end: insertion.Range.End(), text: insertion.Text})
	}

	spanStart, spanEnd := edits[0].start, edits[0].end
	for _, candidate := range edits {
		if candidate.start < spanStart {
			spanStart = candidate.start
		}
		if candidate.end > spanEnd {
			spanEnd = candidate.end
		}
	}

	// Applied back to front within the span, so an earlier edit cannot move a later one's offsets.
	sort.Slice(edits, func(first, second int) bool { return edits[first].start > edits[second].start })
	text := ctx.SourceFile.Text()[spanStart:spanEnd]
	for _, candidate := range edits {
		text = text[:candidate.start-spanStart] + candidate.text + text[candidate.end-spanStart:]
	}
	return rule.ReplaceRange(core.NewTextRange(spanStart, spanEnd), text), true
}

// componentNamesDeclaredIn collects the component-shaped declarations a file makes.
//
// Only top-level declarations, since a component nested inside another function is not part of the
// file's exported surface and could not be named in an export list anyway.
//
// The second return maps EVERY top-level declared name, component or not, to the statements that
// declare it. The repair needs the statement to put `export` on, and needs to see a name declared
// twice, which is how overloads arrive, so it can decline rather than export one signature of many.
func componentNamesDeclaredIn(sourceFile *ast.Node) (map[string]bool, map[string][]*ast.Node) {
	declared := map[string]bool{}
	statements := map[string][]*ast.Node{}

	sourceFile.ForEachChild(func(statement *ast.Node) bool {
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			name := statement.AsFunctionDeclaration().Name()
			if name != nil {
				statements[name.Text()] = append(statements[name.Text()], statement)
			}
			if name != nil && react.IsLikelyComponentName(name.Text()) && HasJsxOrReactHookCalls(statement) {
				declared[name.Text()] = true
			}

		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				return false
			}
			for _, declarationNode := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				declaration := declarationNode.AsVariableDeclaration()
				name := declaration.Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					statements[name.Text()] = append(statements[name.Text()], statement)
				}
				if name == nil || name.Kind != ast.KindIdentifier || !react.IsLikelyComponentName(name.Text()) {
					continue
				}
				// The initializer is what has to look like a component, not the statement: a
				// capitalized constant holding a string is not a component.
				if declaration.Initializer != nil && HasJsxOrReactHookCalls(declaration.Initializer) {
					declared[name.Text()] = true
				}
			}
		}
		return false
	})

	return declared, statements
}
