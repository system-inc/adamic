package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func main() {
	fmt.Println("no_loop_func.a")
	fmt.Println(int(ast.KindFunctionDeclaration))
	fmt.Println(int(ast.KindFunctionExpression))
	fmt.Println(int(ast.KindArrowFunction))
	fmt.Println(int(ast.KindMethodDeclaration))
	fmt.Println(int(ast.KindGetAccessor))
	fmt.Println(int(ast.KindSetAccessor))
	fmt.Println(int(ast.KindConstructor))
	fmt.Println("no_require_imports.a")
	fmt.Println(int(ast.KindCallExpression))
	fmt.Println(int(ast.KindExternalModuleReference))
	fmt.Println("no_import_cycle_load_time_read.a")
	fmt.Println(int(ast.KindSourceFile))
	fmt.Println("wave08-next/no_process_exit_after_output.a")
	fmt.Println(int(ast.KindSourceFile))
	fmt.Println("wave08-next/no_uncleared_race_timeout.a")
	fmt.Println(int(ast.KindCallExpression))
	fmt.Println("wave08-next/require_blocking_standard_streams.a")
	fmt.Println(int(ast.KindSourceFile))
	fmt.Println("wave08-react/globals.a")
	fmt.Println(int(ast.KindIdentifier))
	fmt.Println("wave08-react/immutability.a")
	fmt.Println(int(ast.KindSourceFile))
	fmt.Println("wave08-react/no_deriving_state_in_effects.a")
	fmt.Println(int(ast.KindSourceFile))
}
