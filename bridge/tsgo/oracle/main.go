// The oracle calls typescript-go directly. It imports no bridge implementation,
// goes through no C ABI, and has an independent config loader and query loop.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: oracle tsconfig manifest [root files...]")
	}
	configPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(os.Args[2])
	if err != nil {
		return err
	}
	start := time.Now()
	fs := bundled.WrapFS(osvfs.FS())
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(configPath)), nil, nil, fs, nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		return fmt.Errorf("cannot parse config %s", configPath)
	}
	if len(os.Args) > 3 {
		roots := []tspath.RootedFilePath{}
		for _, root := range os.Args[3:] {
			if !filepath.IsAbs(root) {
				root = filepath.Join(filepath.Dir(configPath), root)
			}
			roots = append(roots, tspath.RootedFilePathFromAbsolute(filepath.ToSlash(root)))
		}
		config = config.WithFileNames(roots)
	}
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	loadTime := time.Since(start)
	queryTime := time.Duration(0)
	count := 0
	for _, line := range strings.Split(string(manifest), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 2 {
			return fmt.Errorf("invalid query line")
		}
		position, err := strconv.Atoi(fields[1])
		if err != nil {
			return err
		}
		start = time.Now()
		path, err := filepath.Abs(fields[0])
		if err != nil {
			return err
		}
		file := program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
		if file == nil || position < 0 || position >= len(file.Text()) {
			return fmt.Errorf("invalid query %s", line)
		}
		node := ast.GetNodeAtPosition(file, position, false)
		checker, release := program.GetTypeCheckerForFile(context.Background(), file)
		symbolName := ""
		if symbol := checker.GetSymbolAtLocation(node); symbol != nil {
			symbolName = ast.EscapeAllInternalSymbolNames(ast.SymbolName(symbol))
		}
		typeName := checker.TypeToString(checker.GetTypeAtLocation(node))
		release()
		queryTime += time.Since(start)
		count++
		fmt.Printf("%d %d %d\n%s%s\n", node.Kind, len(symbolName), len(typeName), symbolName, typeName)
	}
	fmt.Fprintf(os.Stderr, "tsgo: load_ns=%d query_ns=%d queries=%d\n", loadTime.Nanoseconds(), queryTime.Nanoseconds(), count)
	return nil
}
