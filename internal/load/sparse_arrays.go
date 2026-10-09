package load

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// ContainsSparseArrays conservatively gates operations still implemented for dense
// arrays. This program-wide fact includes imports and aliases; a false negative
// would let a dense-only loop interpret a hole as a value.
func (p *Program) ContainsSparseArrays() bool {
	p.sparseOnce.Do(func() {
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindNewExpression {
				made := node.AsNewExpression()
				callee := ast.SkipParentheses(made.Expression)
				if ast.IsIdentifier(callee) && callee.Text() == "Array" {
					filled := false
					if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "fill" {
						if call := parent.Parent; call != nil && call.Kind == ast.KindCallExpression && call.AsCallExpression().Expression == parent && len(call.AsCallExpression().Arguments.Nodes) == 1 {
							filled = true
						}
					}
					if !filled {
						p.sparseArrays = true
						return true
					}
				}
			}
			return node.ForEachChild(visit)
		}
		for _, file := range p.compiler.GetSourceFiles() {
			if file.IsDeclarationFile || IsLibrary(file) || IsPrelude(file) {
				continue
			}
			if visit(file.AsNode()) {
				break
			}
		}
	})
	return p.sparseArrays
}
