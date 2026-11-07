package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Enum declarations are validated before registration. Initialization reachability
// is shared with namespaces in namespaceInitialization, which runs once afterward.
func (lowering *lowering) enumInitialization(modules []*ast.SourceFile) error {
	for _, module := range modules {
		for _, statement := range namespaceDeclarations(module.Statements.Nodes) {
			if statement.Kind == ast.KindEnumDeclaration && !ambientDeclaration(statement) {
				if _, enumFieldsError := lowering.enumFields(statement); enumFieldsError != nil {
					return enumFieldsError
				}
			}
		}
	}
	return nil
}
