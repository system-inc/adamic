// Listener names come from the pinned parser enum.
package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

func kindName(kind ast.Kind) string { return strings.TrimPrefix(kind.String(), "Kind") }

func main() {
	fmt.Printf("%s,%s\n%s,%s\n%s\n%s\n%s\n%s\n%s,%s\n%s\n%s\n", kindName(ast.KindClassDeclaration), kindName(ast.KindClassExpression), kindName(ast.KindReturnStatement), kindName(ast.KindArrowFunction), kindName(ast.KindCallExpression), kindName(ast.KindSourceFile), kindName(ast.KindCallExpression), kindName(ast.KindSourceFile), kindName(ast.KindCallExpression), kindName(ast.KindNewExpression), kindName(ast.KindSourceFile), kindName(ast.KindIdentifier))
}
