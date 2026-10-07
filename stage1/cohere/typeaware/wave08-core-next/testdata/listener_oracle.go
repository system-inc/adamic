package main
import (
 "fmt"
 "strings"
 "strconv"
 "github.com/microsoft/TypeScript/tsc/shim/ast"
)
func main() {
 names:=[]string{"require-atomic-updates","require-await","symbol-description"}
 kinds:=[][]ast.Kind{{ast.KindSourceFile},{ast.KindFunctionDeclaration,ast.KindFunctionExpression,ast.KindArrowFunction,ast.KindMethodDeclaration,ast.KindGetAccessor,ast.KindSetAccessor,ast.KindConstructor},{ast.KindCallExpression}}
 for i,name:=range names {var values []string;for _,kind:=range kinds[i] {values=append(values,strconv.Itoa(int(kind)))};fmt.Printf("%s\t%s\n",name,strings.Join(values,","))}
}
