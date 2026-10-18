package load

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// TS2308 is already a checker rejection. Add the second module at the offending
// star declaration so the diagnostic identifies both sources of the ambiguity.
func starCollision(diagnostic *ast.Diagnostic) string {
	if diagnostic.Code() != 2308 || diagnostic.File() == nil || len(diagnostic.MessageArgs()) != 2 {
		return ""
	}
	for _, statement := range diagnostic.File().Statements.Nodes {
		if statement.Kind != ast.KindExportDeclaration || statement.Pos() > diagnostic.Pos() || statement.End() <= diagnostic.Pos() {
			continue
		}
		declaration := statement.AsExportDeclaration()
		if declaration.ExportClause != nil || declaration.ModuleSpecifier == nil {
			continue
		}
		arguments := diagnostic.MessageArgs()
		return fmt.Sprintf("Adamic 0.1 refuses export * collision for '%s' between '%s' and '%s'; explicitly re-export one binding", arguments[1], strings.Trim(arguments[0], "'\""), declaration.ModuleSpecifier.Text())
	}
	return ""
}
