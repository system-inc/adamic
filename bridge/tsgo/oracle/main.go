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
	if runError := run(); runError != nil {
		fmt.Fprintln(os.Stderr, runError)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: oracle tsconfig manifest [root files...]")
	}
	configurationPath, absolutePathError := filepath.Abs(os.Args[1])
	if absolutePathError != nil {
		return absolutePathError
	}
	manifest, readError := os.ReadFile(os.Args[2])
	if readError != nil {
		return readError
	}
	start := time.Now()
	fs := bundled.WrapFS(osvfs.FS())
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	configuration, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(
		tspath.RootedFilePathFromAbsolute(filepath.ToSlash(configurationPath)),
		nil,
		nil,
		fs,
		nil,
	)
	if configuration == nil || len(diagnostics) > 0 || len(configuration.Errors) > 0 {
		return fmt.Errorf("cannot parse config %s", configurationPath)
	}
	if len(os.Args) > 3 {
		roots := []tspath.RootedFilePath{}
		for _, root := range os.Args[3:] {
			if !filepath.IsAbs(root) {
				root = filepath.Join(filepath.Dir(configurationPath), root)
			}
			roots = append(roots, tspath.RootedFilePathFromAbsolute(filepath.ToSlash(root)))
		}
		configuration = configuration.WithFileNames(roots)
	}
	program := compiler.NewProgram(
		compiler.ProgramOptions{Config: configuration, Host: host, SingleThreaded: core.TSTrue},
	)
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
		position, parseError := strconv.Atoi(fields[1])
		if parseError != nil {
			return parseError
		}
		start = time.Now()
		path, absolutePathError := filepath.Abs(fields[0])
		if absolutePathError != nil {
			return absolutePathError
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
	fmt.Fprintf(
		os.Stderr,
		"tsgo: load_ns=%d query_ns=%d queries=%d\n",
		loadTime.Nanoseconds(),
		queryTime.Nanoseconds(),
		count,
	)
	return nil
}
