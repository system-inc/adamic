// Built inside cohere through an overlay. Reads unchanged production listeners.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var wanted = map[string]bool{
	"@typescript-eslint/no-for-in-array":         true,
	"@typescript-eslint/prefer-regexp-exec":      true,
	"nexus/correctness-no-collection-misuse":     true,
	"nexus/correctness-no-discarded-outcome":     true,
	"nexus/correctness-no-discarded-pure-result": true,
	"nexus/correctness-no-identical-branches":    true,
	"no-eval":                      true,
	"no-extend-native":             true,
	"no-func-assign":               true,
	"no-new-func":                  true,
	"no-new-native-nonconstructor": true,
	"no-new-wrappers":              true,
	"no-throw-literal":             true,
	"no-useless-backreference":     true,
	"prefer-arrow-callback":        true,
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: oracle tsconfig source")
	}
	configPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	sourcePath, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}
	host := compiler.NewCachedFSCompilerHost(filepath.ToSlash(filepath.Dir(configPath)), bundled.WrapFS(osvfs.FS()), bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(filepath.ToSlash(configPath), nil, nil, host, nil)
	if config == nil || len(diagnostics) != 0 || len(config.Errors) != 0 {
		panic("invalid tsconfig")
	}
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames([]string{filepath.ToSlash(sourcePath)}), Host: host, SingleThreaded: core.TSTrue})
	source := program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(sourcePath)))
	if source == nil {
		panic("source not loaded")
	}
	checker, release := program.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	var lines []string
	for _, subject := range registry.All() {
		if !wanted[subject.Name] {
			continue
		}
		listeners := subject.Run(rule.Context{SourceFile: source, TypeChecker: checker, Program: rule.ViewProgram(program, source, subject), FileCache: rule.NewFileCache(), Report: func(rule.Diagnostic) { panic("registration emitted a diagnostic") }}, nil)
		kinds := make([]int, 0, len(listeners))
		for kind := range listeners {
			kinds = append(kinds, int(kind))
		}
		sort.Ints(kinds)
		values := make([]string, len(kinds))
		for index, kind := range kinds {
			values[index] = strconv.Itoa(kind)
		}
		lines = append(lines, subject.Name+"\t"+strings.Join(values, ","))
	}
	if len(lines) != len(wanted) {
		panic("production rule missing")
	}
	sort.Strings(lines)
	for _, line := range lines {
		fmt.Println(line)
	}
}
