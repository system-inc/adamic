package main
import("fmt";"github.com/microsoft/TypeScript/tsc/shim/ast")
func main(){fmt.Println(int(ast.KindSetAccessor),int(ast.KindConstructor))}
