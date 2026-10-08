package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// Checked literal builtin requires have the same static namespace as imports.
// Dynamic requests, untyped any and observing the module object are refused.
func (l *lowering) nodeRequireBinding(declaration *ast.Node) string {
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return ""
	}
	written := declaration.AsVariableDeclaration()
	if written.Type == nil || written.Type.Kind != ast.KindImportType || written.Initializer == nil || written.Initializer.Kind != ast.KindCallExpression {
		return ""
	}
	annotation := written.Type.AsImportTypeNode()
	if !annotation.IsTypeOf || annotation.Qualifier != nil || annotation.Argument.Kind != ast.KindLiteralType {
		return ""
	}
	literal := annotation.Argument.AsLiteralTypeNode().Literal
	call := written.Initializer.AsCallExpression()
	if !ast.IsIdentifier(call.Expression) || call.Expression.Text() != "require" || len(call.Arguments.Nodes) != 1 || call.Arguments.Nodes[0].Kind != ast.KindStringLiteral {
		return ""
	}
	symbol := l.symbol(call.Expression)
	if symbol == nil {
		return ""
	}
	native := false
	for _, d := range symbol.Declarations {
		native = native || load.IsNodeLibrary(ast.GetSourceFileOfNode(d))
	}
	if !native || literal.Kind != ast.KindStringLiteral {
		return ""
	}
	module := strings.TrimPrefix(call.Arguments.Nodes[0].Text(), "node:")
	if strings.TrimPrefix(literal.Text(), "node:") != module {
		return ""
	}
	switch module {
	case "fs", "path", "os", "process", "perf_hooks", "crypto":
		return module
	}
	return ""
}
