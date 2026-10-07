package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"strconv"
	"strings"
)

func main() {
	names := []string{"require-atomic-updates", "require-await", "symbol-description"}
	kinds := [][]ast.Kind{{ast.KindSourceFile}, {ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor}, {ast.KindCallExpression}}
	if len(os.Args) > 1 && os.Args[1] == "--json" {
		result := map[string][]string{}
		for i, name := range names {
			for _, kind := range kinds[i] {
				result[name] = append(result[name], strings.TrimPrefix(kind.String(), "Kind"))
			}
		}
		data, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		return
	}
	for i, name := range names {
		var values []string
		for _, kind := range kinds[i] {
			values = append(values, strconv.Itoa(int(kind)))
		}
		fmt.Printf("%s\t%s\n", name, strings.Join(values, ","))
	}
}
