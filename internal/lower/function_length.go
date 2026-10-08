package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// sourceFunctionLength follows ExpectedArgumentCount: optional parameters count,
// defaults and rest end the count, and the erased TypeScript this is excluded.
func sourceFunctionLength(declaration *ast.Node) int {
	count := 0
	for _, node := range declaration.Parameters() {
		parameter := node.AsParameterDeclaration()
		if ast.IsIdentifier(node.Name()) && node.Name().Text() == "this" {
			continue
		}
		if parameter.Initializer != nil || parameter.DotDotDotToken != nil {
			break
		}
		count++
	}
	return count
}
