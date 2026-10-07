package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"os"
	"path/filepath"
)

func main() {
	configPath := os.Args[1]
	source := os.Args[2]
	host := compiler.NewCachedFSCompilerHost(filepath.ToSlash(filepath.Dir(configPath)), bundled.WrapFS(osvfs.FS()), bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(configPath, nil, nil, host, nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		panic("invalid config")
	}
	config.CompilerOptions().AllowNonTsExtensions = core.TSTrue
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames([]string{source}), Host: host, SingleThreaded: core.TSTrue})
	file := program.GetSourceFile(source)
	if file == nil || len(file.Diagnostics()) > 0 {
		panic("source parse failed")
	}
	checker, release := program.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	ctx := rule.Context{SourceFile: file, TypeChecker: checker}
	var captures []rules.Wave01Expand
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
			switch n.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
				captures = append(captures, rules.Wave01AwaitExpandCapture(ctx, n)...)
			}
		}
		n.ForEachChild(walk)
		return false
	}
	walk(file.AsNode())
	data, err := json.Marshal(captures)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}
