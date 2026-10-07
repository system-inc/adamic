package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The pinned compiler's raw reference flags, without a rule verdict.
func (p *Program) referenceAccess(out *fields, _ *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	if question != "reference-access" {
		return fmt.Errorf("reference-access takes no operands")
	}
	out.yes(ast.IsWriteAccess(node))
	out.yes(ast.IsDeclarationName(node))
	out.yes(ast.IsPartOfTypeNode(node))
	return nil
}
func init() { additionalQuestions["reference-access"] = (*Program).referenceAccess }
