// Count syntactic calls and resolved method declarations without lowering.
package main

import (
	"context"
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	paths, err := filepath.Glob("stage1/typescript/parser/*.ts")
	if err != nil {
		panic(err)
	}
	program, err := load.Load(paths)
	if err != nil {
		panic(err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	type counts struct {
		This               int
		ThisCalls          int
		ThisFieldCalls     int
		OtherPropertyCalls int
		DirectCalls        int
		ClassMethods       int
		InterfaceMethods   int
		LibraryMethods     int
		Unresolved         int
	}
	result := map[string]*counts{}
	for _, file := range program.Files() {
		name := filepath.ToSlash(string(file.FileName()))
		if !strings.Contains(name, "stage1/typescript/parser/") || strings.Contains(name, "/testdata/") {
			continue
		}
		row := &counts{}
		result[filepath.Base(name)] = row
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if node.Kind == ast.KindThisKeyword {
				row.This++
			}
			if node.Kind == ast.KindCallExpression {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee.Kind == ast.KindPropertyAccessExpression {
					receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
					if receiver.Kind == ast.KindThisKeyword {
						row.ThisCalls++
					} else if receiver.Kind == ast.KindPropertyAccessExpression && ast.SkipParentheses(receiver.AsPropertyAccessExpression().Expression).Kind == ast.KindThisKeyword {
						row.ThisFieldCalls++
					} else {
						row.OtherPropertyCalls++
					}
					symbol := c.GetSymbolAtLocation(callee)
					if symbol == nil || len(symbol.Declarations) == 0 {
						row.Unresolved++
					} else {
						d := symbol.Declarations[0]
						if load.IsLibrary(ast.GetSourceFileOfNode(d)) {
							row.LibraryMethods++
						} else if d.Kind == ast.KindMethodDeclaration {
							row.ClassMethods++
						} else {
							row.InterfaceMethods++
						}
					}
				} else {
					row.DirectCalls++
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	json.NewEncoder(os.Stdout).Encode(result)
}
