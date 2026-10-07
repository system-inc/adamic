package main
import("fmt";"github.com/microsoft/TypeScript/tsc/shim/ast";"github.com/microsoft/TypeScript/tsc/shim/core";"github.com/microsoft/TypeScript/tsc/shim/parser";rules "github.com/system-inc/cohere/internal/lint/rules/core")
func main(){
 texts:=[]string{"async function f(){await x;}","async function f(){for await (const x of y){}}","async function f(){await using x=foo();}","async function f(){const x=0;}","async function f(){using x=foo();}","async function f(){async function g(){await x;}}","async function f(){class A {async g(){await x;}}}","async function f(){for(const x of y){}}"}
 var bodies []*ast.Node
 for _,text:=range texts {file:=parser.ParseSourceFile(ast.SourceFileParseOptions{FileName:"/workspace/test.ts"},text,core.ScriptKindTS);bodies=append(bodies,file.Statements.Nodes[0].Body())}
 fmt.Print(rules.Wave01AwaitKernelCapture(bodies))
}
