package main

import (
	"context"
	"fmt"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// compileWASI uses the entry module's handleRequest export, as in the runtime's source witness.
// Private or dependency-only names are not handlers. Export declarations remain subject to lowering refusals.
func compileWASI(path string) (*ir.Program, int, int) {
	program, code := check([]string{path})
	if program == nil {
		return nil, -1, code
	}
	name, err := requestFunction(program)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adamic: "+err.Error())
		return nil, -1, 1
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fmt.Fprintln(os.Stderr, relative("adamic: "+err.Error()))
		return nil, -1, 1
	}
	handler := -1
	if name != "" {
		for index, function := range lowered.Functions {
			if function.Name != name || function.Closure {
				continue
			}
			if handler >= 0 {
				fmt.Fprintln(os.Stderr, "adamic: not yet: request handler name is ambiguous across modules; give its declaration a unique name")
				return nil, -1, 1
			}
			handler = index
		}
		if handler < 0 {
			fmt.Fprintln(os.Stderr, "adamic: request handler was not lowered")
			return nil, -1, 1
		}
	}
	return lowered, handler, 0
}

func requestFunction(program *load.Program) (string, error) {
	entry := program.Files()[0]
	typeChecker, release := program.Checker(context.Background(), entry)
	defer release()
	module := typeChecker.GetSymbolAtLocation(entry.AsNode())
	if module == nil {
		return "", nil
	}
	for _, symbol := range typeChecker.GetExportsOfModule(module) {
		if symbol.Name != "handleRequest" {
			continue
		}
		if symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = typeChecker.GetAliasedSymbol(symbol)
		}
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
			return "", fmt.Errorf("not yet: handleRequest must export a named function declaration")
		}
		declaration := symbol.Declarations[0]
		invalid := func() (string, error) {
			return "", fmt.Errorf("%s: request handler must take one required string parameter and return a string", program.Where(declaration))
		}
		if len(declaration.TypeParameters()) != 0 || len(declaration.Parameters()) != 1 {
			return invalid()
		}
		parameter := declaration.Parameters()[0]
		details := parameter.AsParameterDeclaration()
		if !ast.IsIdentifier(parameter.Name()) || details.DotDotDotToken != nil || details.QuestionToken != nil || details.Initializer != nil {
			return invalid()
		}
		if typeChecker.GetTypeAtLocation(parameter).Flags() != checker.TypeFlagsString {
			return invalid()
		}
		signature := typeChecker.GetSignatureFromDeclaration(declaration)
		if typeChecker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsStringLike == 0 {
			return invalid()
		}
		return declaration.Name().Text(), nil
	}
	return "", nil
}
