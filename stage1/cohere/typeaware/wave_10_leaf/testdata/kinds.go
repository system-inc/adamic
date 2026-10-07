package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

func main() {
	for _, k := range []ast.Kind{ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindCallExpression, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor, ast.KindBlock, ast.KindAwaitExpression, ast.KindForOfStatement, ast.KindVariableDeclarationList, ast.KindClassDeclaration, ast.KindClassExpression} {
		fmt.Printf("%s\t%d\n", strings.TrimPrefix(k.String(), "Kind"), k)
	}
	fmt.Printf("ModifierFlagsAsync\t%d\n", ast.ModifierFlagsAsync)
	fmt.Printf("NodeFlagsBlockScoped\t%d\n", ast.NodeFlagsBlockScoped)
	fmt.Printf("NodeFlagsAwaitUsing\t%d\n", ast.NodeFlagsAwaitUsing)
}
