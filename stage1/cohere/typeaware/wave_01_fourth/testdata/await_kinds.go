package main
import("fmt";"github.com/microsoft/TypeScript/tsc/shim/ast")
func main(){for _,k:=range []ast.Kind{ast.KindAwaitExpression,ast.KindForOfStatement,ast.KindVariableDeclarationList,ast.KindClassDeclaration,ast.KindClassExpression,ast.KindBlock}{fmt.Println(k.String(),int(k))}}
