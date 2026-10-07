// Numeric listener declarations use the pinned parser's enum, not stock TypeScript.
package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	fmt.Printf("%d,%d\n%d,%d\n%d\n%d\n%d\n%d\n%d,%d\n%d\n%d\n", ast.KindClassDeclaration, ast.KindClassExpression, ast.KindReturnStatement, ast.KindArrowFunction, ast.KindCallExpression, ast.KindSourceFile, ast.KindCallExpression, ast.KindSourceFile, ast.KindCallExpression, ast.KindNewExpression, ast.KindSourceFile, ast.KindIdentifier)
}
