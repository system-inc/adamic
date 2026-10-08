//go:build scoutoracle

// Built by an overlay inside the pinned cohere module. No bridge import.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/linter"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
	"github.com/system-inc/cohere/internal/types/program"
)

func main() {
	config, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	graph, err := program.Build(program.Options{ConfigFileName: config, CurrentDirectory: filepath.Dir(config), SingleThreaded: true})
	if err != nil {
		panic(err)
	}

	paths := []string{os.Args[2]}
	manifest := os.Args[2] == "--manifest"
	if manifest {
		data, e := os.ReadFile(os.Args[3])
		if e != nil {
			panic(e)
		}
		paths = strings.Split(strings.TrimSpace(string(data)), "\n")
	}
	for _, path := range paths {
		if manifest {
			fmt.Printf("file %s\n", path)
		}
		lint(graph, path)
	}
}
func lint(graph *program.Graph, path string) {
	file := graph.Program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
	if file == nil {
		panic("missing source")
	}
	checker, release := graph.CheckerForFile(context.Background(), file)
	defer release()
	diagnostics := linter.LintFile(file, []rule.Rule{rules.NoUnnecessaryBooleanLiteralCompare}, graph.Program, checker)
	fmt.Printf("findings %d\n", len(diagnostics))
	for _, d := range diagnostics {
		if len(d.Fixes) != 1 || len(d.Suggestions) != 0 {
			panic("unexpected pilot repair shape")
		}

		fields := []string{fmt.Sprint(d.Range.Pos()), fmt.Sprint(d.Range.End()), d.RuleName, d.Message.Id, d.Message.Description, fmt.Sprint(d.Fixes[0].Range.Pos()), fmt.Sprint(d.Fixes[0].Range.End()), d.Fixes[0].Text}
		for _, field := range fields {
			fmt.Printf("%d\n%s", len(utf16.Encode([]rune(field))), field)
		}
		fmt.Println()
	}
}
