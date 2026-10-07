package main
import("fmt";"github.com/microsoft/TypeScript/tsc/shim/ast")
func main(){for _,k:=range []ast.Kind{ast.KindFunctionDeclaration,ast.KindMethodDeclaration,ast.KindGetAccessor,ast.KindFunctionExpression,ast.KindArrowFunction,ast.KindReturnStatement,ast.KindIdentifier,ast.KindPrivateIdentifier,ast.KindSuperKeyword,ast.KindElementAccessExpression,ast.KindIfStatement,ast.KindCallExpression,ast.KindSourceFile,ast.KindForOfStatement,ast.KindForStatement}{fmt.Printf("%s %d\n",k.String(),k)}}
