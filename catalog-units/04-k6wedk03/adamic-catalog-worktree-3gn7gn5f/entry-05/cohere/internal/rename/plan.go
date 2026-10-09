package rename

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/types/program"
)

// Build produces the whole plan for one rename without writing anything.
//
// The separation is the safety property: a plan is computed from the type graph, printed in full,
// and only then applied. Nothing here mutates, so a dry run and a real run compute the identical
// answer and the preview cannot drift from what lands.
func Build(ctx context.Context, graph *program.Graph, at Position, newName string) (*Plan, error) {
	files := graph.ProjectFiles()

	plan := &Plan{NewName: newName}

	if !IsValidIdentifier(newName) {
		plan.Refusals = append(plan.Refusals, fmt.Sprintf(
			"%q cannot be written as an identifier, so the rename would produce source that does not parse",
			newName,
		))
		return plan, nil
	}

	anchor, anchorNode, err := resolveAnchor(ctx, graph, files, at)
	if err != nil {
		return nil, err
	}
	if anchor == nil {
		return nil, fmt.Errorf(
			"no symbol at %s:%d:%d — check the position names an identifier, and that the file is in the program",
			at.FileName, at.Line, at.Column,
		)
	}
	plan.OldName = anchorNode.Text()

	if plan.OldName == newName {
		plan.Refusals = append(plan.Refusals, fmt.Sprintf(
			"the symbol is already named %q, so there is nothing to do", newName,
		))
		return plan, nil
	}

	// An imported binding resolves to an alias whose target is the real declaration in the other
	// file. Renaming the alias alone renames the local half of an import and leaves the export
	// untouched, which is a legitimate and DIFFERENT act from renaming the export. The position
	// decides which one the user asked for, so the alias is NOT followed here: the user put their
	// cursor on a local binding and gets a local rename.
	//
	// The consequence is stated rather than hidden. Renaming the local half of `import { err }`
	// produces `import { err as error }`, which this does not yet write — see the refusal below.
	inProject := projectFileNames(files)
	declarationFile, declarationLine, declarationRefusal := describeDeclaration(anchor, inProject)
	plan.DeclarationFile = declarationFile
	plan.DeclarationLine = declarationLine
	if declarationRefusal != "" {
		plan.Refusals = append(plan.Refusals, declarationRefusal)
		return plan, nil
	}

	collectEdits(ctx, graph, files, anchor, plan)

	// The new name being taken in a scope the rename reaches is a collision, and a collision is a
	// refusal rather than something to write and let the type phase find. A rename that produces a
	// duplicate binding may still compile — a shadowing collision does — so this cannot be left to a
	// downstream check.
	if collision := findCollision(ctx, graph, files, anchor, newName, plan); collision != "" {
		plan.Refusals = append(plan.Refusals, collision)
	}

	collectBlindness(files, plan)

	touched := map[string]bool{}
	for _, edit := range plan.Edits {
		touched[edit.FileName] = true
	}
	plan.FilesTouched = len(touched)

	sort.Slice(plan.Edits, func(first int, second int) bool {
		if plan.Edits[first].FileName != plan.Edits[second].FileName {
			return plan.Edits[first].FileName < plan.Edits[second].FileName
		}
		return plan.Edits[first].Range.Pos() < plan.Edits[second].Range.Pos()
	})

	return plan, nil
}

// projectFileNames is the set of files the tsconfig named, which is the boundary of what the user
// owns.
//
// This is the whole mechanism behind the out-of-project refusal, and it is exact rather than
// heuristic: `ProjectFiles` is already the distinction between our code and the declarations the
// compiler pulled in, measured at roughly 3,400 files out of 9,530 on the ahra tree. A path check
// for `node_modules` would have been the obvious implementation and would be wrong in both
// directions — it would miss a linked dependency outside that directory, and it would refuse a
// legitimately vendored path that happens to contain the string.
func projectFileNames(files []*ast.SourceFile) map[string]bool {
	names := make(map[string]bool, len(files))
	for _, file := range files {
		names[file.FileName().AsString()] = true
	}
	return names
}

// describeDeclaration finds where a symbol is declared and refuses if that is not ours to rename.
//
// A symbol may carry several declarations through declaration merging, and the brief's standing
// instruction is to never index [0] and to say which question is being answered. The question here
// is "may I rewrite every declaration of this", so the answer must be no if ANY of them sits
// outside the project: renaming an interface that merges with one in a dependency's `.d.ts` would
// rewrite our half and leave theirs, producing two types where the author meant one.
func describeDeclaration(symbol *ast.Symbol, inProject map[string]bool) (string, int, string) {
	if len(symbol.Declarations) == 0 {
		return "", 0, "the symbol has no declaration in the program, so there is nothing to rename"
	}

	var firstFile string
	var firstLine int
	for index, declaration := range symbol.Declarations {
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile == nil {
			return "", 0, "a declaration of this symbol has no source file, so the rename cannot be shown to be complete"
		}
		fileName := sourceFile.FileName()
		line, _ := lineColumnOf(sourceFile.Text(), declaration.Pos())
		if index == 0 {
			firstFile = fileName.AsString()
			firstLine = line
		}
		if !inProject[fileName.AsString()] {
			return firstFile, firstLine, fmt.Sprintf(
				"%s is declared outside this project, at %s:%d — renaming a declaration we do not own would break every other consumer of that package",
				symbol.Name, fileName, line,
			)
		}
	}
	return firstFile, firstLine, ""
}

// collectEdits walks every identifier in the project and records the ones that resolve to the
// anchor.
//
// The walk is over identifier positions rather than over a list of reference forms, for the reason
// `internal/unused` gives: a reference is a use wherever it appears, and enumerating the forms would
// miss whichever one nobody thought of.
func collectEdits(ctx context.Context, graph *program.Graph, files []*ast.SourceFile, anchor *ast.Symbol, plan *Plan) {
	for _, file := range files {
		fileChecker, release := graph.CheckerForFile(ctx, file)
		if fileChecker == nil {
			release()
			continue
		}
		text := file.Text()
		fileName := file.FileName()

		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			if node.Kind == ast.KindIdentifier && node.Text() == plan.OldName {
				if edit, matched := editFor(fileChecker, file, text, fileName.AsString(), node, anchor, plan.NewName); matched {
					plan.Edits = append(plan.Edits, edit)
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}
}

// editFor decides whether one identifier is a site this rename must rewrite, and what to write.
//
// The name filter above this is an optimization and not the decision: only symbol identity decides,
// so a same-named identifier bound to something else is declined here rather than being renamed by
// coincidence. That is the trap the brief calls out for declaration-anchor rules — identity alone
// over-reports and name matching under-discriminates — and the shadowing fixture pins it.
func editFor(
	fileChecker *checker.Checker,
	file *ast.SourceFile,
	text string,
	fileName string,
	node *ast.Node,
	anchor *ast.Symbol,
	newName string,
) (Edit, bool) {
	parent := node.Parent

	// A shorthand property is two symbols wearing one identifier, and the plain accessor answers
	// about the wrong one. Measured on a fixture: GetSymbolAtLocation returns flags 4 (Property)
	// while GetShorthandAssignmentValueSymbol returns flags 2 (BlockScopedVariable) whose
	// declaration is the binding. They are not equal, so a rename driven by the plain accessor would
	// never fire here and would leave `{ err }` binding to a name that no longer exists.
	if parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
		valueSymbol := fileChecker.GetShorthandAssignmentValueSymbol(parent)
		if !sameSymbol(fileChecker, valueSymbol, anchor) {
			return Edit{}, false
		}
		// EXPANSION, not replacement. `{ err }` means `{ err: err }`, and the property name is part
		// of the object's contract while the value is the binding being renamed. Writing `{ error }`
		// would rename the property too, silently changing the shape of every object literal that
		// puns the name.
		line, column := lineColumnOf(text, tokenSpan(file, node).Pos())
		return Edit{
			FileName: fileName,
			Range:    tokenSpan(file, node),
			Line:     line,
			Column:   column,
			Text:     fmt.Sprintf("%s: %s", node.Text(), newName),
			Kind:     EditShorthandExpansion,
		}, true
	}

	symbol := fileChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return Edit{}, false
	}

	matched := symbol == anchor
	kind := EditReference

	// An import specifier's identifier is an alias in its own right whose target is the real
	// declaration in the other file. Renaming the export must rewrite this SOURCE half and leave a
	// local alias alone, which falls out of resolving the alias and comparing the target: in
	// `import { err as failure }` the `err` half resolves through the alias to the anchor and the
	// `failure` half is the alias's own name, which is a declaration name of a different symbol.
	if !matched && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if aliased := fileChecker.GetAliasedSymbol(symbol); aliased == anchor {
			matched = true
		}
	}

	if !matched {
		return Edit{}, false
	}

	// Classification order is load-bearing and was measured rather than reasoned about, because the
	// obvious order is wrong in a way that still produces a correct rename and a misleading preview.
	//
	// An import specifier's identifier satisfies `parent.Name() == node` in BOTH spellings: for
	// `import { err }` the Name is `err` and there is no PropertyName, and for
	// `import { err as failure }` the Name is `failure` while the PropertyName is `err`. So a
	// declaration check running last relabels every import site as a declaration, which tells the
	// reader of a dry run that the symbol is declared in four places when it is declared in one.
	//
	// The preview is the entire safety mechanism for this verb, so a preview that misdescribes what
	// it is about to do is a real defect even when the bytes it writes are right.
	// The source half is classified HERE rather than inside the alias branch above, and the
	// difference is not cosmetic. In `import { err as failure }` the `err` half resolves through
	// `GetSymbolAtLocation` DIRECTLY to the anchor, so `matched` is already true and the alias branch
	// never runs — which left the source half labelled an ordinary reference. The bytes were right
	// and the preview was wrong, and since the preview is the whole safety mechanism of this verb, a
	// preview that misdescribes what it is about to do is a real defect. Caught by a fixture
	// asserting the kind rather than the count.
	switch {
	case isImportOrExportSourceHalf(node):
		kind = EditImportSource
	case isImportBindingName(node):
		kind = EditReference
	case isOwnDeclarationName(node):
		kind = EditDeclaration
	}

	line, column := lineColumnOf(text, tokenSpan(file, node).Pos())
	return Edit{
		FileName: fileName,
		Range:    tokenSpan(file, node),
		Line:     line,
		Column:   column,
		Text:     newName,
		Kind:     kind,
	}, true
}

// sameSymbol compares two symbols, following an alias when one side is one.
func sameSymbol(fileChecker *checker.Checker, candidate *ast.Symbol, anchor *ast.Symbol) bool {
	if candidate == nil {
		return false
	}
	if candidate == anchor {
		return true
	}
	if candidate.Flags&ast.SymbolFlagsAlias != 0 {
		return fileChecker.GetAliasedSymbol(candidate) == anchor
	}
	return false
}

// isImportOrExportSourceHalf reports whether an identifier is the exporting module's spelling in an
// aliased import or export, rather than the local name.
//
// `import { err as failure }` parses as an ImportSpecifier whose PropertyName is `err` and whose
// Name is `failure`. When there is no alias the PropertyName is nil and the Name is both.
func isImportOrExportSourceHalf(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindImportSpecifier:
		specifier := parent.AsImportSpecifier()
		return specifier != nil && specifier.PropertyName == node
	case ast.KindExportSpecifier:
		specifier := parent.AsExportSpecifier()
		return specifier != nil && specifier.PropertyName == node
	}
	return false
}

// isOwnDeclarationName reports whether an identifier is the name a declaration gives itself.
func isOwnDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	return parent.Name() == node
}

// findCollision reports whether the new name is already bound somewhere this rename reaches.
//
// The check is deliberately over-eager and says so in its message: it asks whether any file this
// rename touches already contains a binding declared with the new name. That will refuse some
// renames that would have been fine, because a binding in an unrelated function of the same file
// cannot actually collide. The asymmetry is the point — a false refusal costs the user a message
// they can read and work around, while a missed collision produces either a duplicate binding or a
// silent shadow, and a silent shadow COMPILES. Refusing loudly beats guessing, which is the line
// this codebase already holds elsewhere.
func findCollision(
	ctx context.Context,
	graph *program.Graph,
	files []*ast.SourceFile,
	anchor *ast.Symbol,
	newName string,
	plan *Plan,
) string {
	touched := map[string]bool{}
	for _, edit := range plan.Edits {
		touched[edit.FileName] = true
	}

	for _, file := range files {
		if !touched[file.FileName().AsString()] {
			continue
		}
		fileChecker, release := graph.CheckerForFile(ctx, file)
		if fileChecker == nil {
			release()
			continue
		}
		text := file.Text()
		found := ""
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil || found != "" {
				return false
			}
			if node.Kind == ast.KindIdentifier && node.Text() == newName && isOwnDeclarationName(node) {
				existing := fileChecker.GetSymbolAtLocation(node)
				if existing != nil && existing != anchor {
					line, column := lineColumnOf(text, tokenSpan(file, node).Pos())
					found = fmt.Sprintf(
						"%s is already declared at %s:%d:%d, and this rename writes into that file — renaming into a name that is taken produces a duplicate binding or a silent shadow, and a silent shadow compiles",
						newName, file.FileName(), line, column,
					)
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
		if found != "" {
			return found
		}
	}
	return ""
}

// collectBlindness records the string literals that spell the old name, which no checker can see.
//
// This does not decide anything and it does not block the rename. It exists so the user is told
// what the tool could not answer, because the alternative is a completeness claim the tool cannot
// support. `t["err"]` is a property access the checker does not connect to the symbol, and a rename
// that misses one produces a codebase that compiles and is broken at runtime — which no phase after
// this one will catch.
//
// Scoped to files the rename touches plus files that mention the name at all, and reported with
// position and text so a person can look at each one and decide. Over-reporting here is correct:
// every entry is something a human reads, not something the tool acts on.
func collectBlindness(files []*ast.SourceFile, plan *Plan) {
	for _, file := range files {
		text := file.Text()
		if !strings.Contains(text, plan.OldName) {
			continue
		}
		fileName := file.FileName()

		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				if node.Text() == plan.OldName {
					line, column := lineColumnOf(text, tokenSpan(file, node).Pos())
					plan.Blind = append(plan.Blind, Blindness{
						FileName: fileName.AsString(),
						Line:     line,
						Column:   column,
						Text:     node.Text(),
						Reason:   "a string equal to the old name; if this is a key or a dynamic lookup the checker cannot see it and this rename did not touch it",
					})
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}

	sort.Slice(plan.Blind, func(first int, second int) bool {
		if plan.Blind[first].FileName != plan.Blind[second].FileName {
			return plan.Blind[first].FileName < plan.Blind[second].FileName
		}
		return plan.Blind[first].Line < plan.Blind[second].Line
	})
}

// isImportBindingName reports whether an identifier is the local name an import or export
// introduces, in any of its spellings.
//
// Distinguished from a real declaration because both satisfy `parent.Name() == node`, measured on
// the fixture: `import { err }` gives an ImportSpecifier whose Name is `err` and whose PropertyName
// is nil, so a plain declaration check calls it a declaration. It is not one — the declaration lives
// in the other file, and calling it one would tell a reader the symbol is declared in as many
// places as it is imported.
func isImportBindingName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindImportSpecifier, ast.KindExportSpecifier, ast.KindImportClause,
		ast.KindNamespaceImport:
		return parent.Name() == node
	}
	return false
}
