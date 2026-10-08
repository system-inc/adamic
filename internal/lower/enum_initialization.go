package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Enum declarations are validated before registration. Initialization reachability
// is shared with namespaces in namespaceInitialization, which runs once afterward.
func (l *lowering) enumInitialization(modules []*ast.SourceFile) error {
	for _, module := range modules {
		for _, statement := range namespaceDeclarations(module.Statements.Nodes) {
			if statement.Kind == ast.KindEnumDeclaration {
				if _, err := l.enumFields(statement); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
