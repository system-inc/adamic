// Built through an overlay in cohere, with its actual checker and rule.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/report"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
	"github.com/system-inc/cohere/internal/types/program"
)

func written(text string) string {
	var result strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != '\\' {
			result.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&result, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&result, `\u%04x\u%04x`, h, l)
		}
	}
	return result.String()
}

func main() {
	path, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	graph, err := program.Build(program.Options{
		ConfigFileName:   filepath.Join(filepath.Dir(path), "tsconfig.json"),
		CurrentDirectory: filepath.Dir(path), SingleThreaded: true,
	})
	if err != nil {
		panic(err)
	}
	var file *ast.SourceFile
	for _, candidate := range graph.ProjectFiles() {
		if candidate.FileName().AsString() == filepath.ToSlash(path) {
			file = candidate
			break
		}
	}
	if file == nil {
		panic("typed fixture is not in the program")
	}
	if len(file.Diagnostics()) != 0 {
		panic("typed fixture has parse diagnostics")
	}
	checker, release := graph.CheckerForFile(context.Background(), file)
	defer release()
	if checker == nil {
		panic("typed fixture has no checker")
	}
	subject := rule.Bind(nexus.ConsistencyNoAbbreviatedIdentifier,
		rule.Given[nexus.ConsistencyNoAbbreviatedIdentifierOptions]{})
	var diagnostics []rule.Diagnostic
	listeners := subject.Run(rule.Context{
		SourceFile: file, Program: rule.ViewProgram(graph.Program, file, subject.Traits),
		TypeChecker: checker, FileCache: rule.NewFileCache(),
		Report: func(d rule.Diagnostic) {
			d.RuleName = subject.Name
			if d.SourceFile == nil {
				d.SourceFile = file
			}
			diagnostics = append(diagnostics, d)
		},
	})
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if visit := listeners[node.Kind]; visit != nil {
			visit(node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	sort.SliceStable(diagnostics, func(i, j int) bool { return diagnostics[i].Range.Pos() < diagnostics[j].Range.Pos() })
	fmt.Println("case 0")
	for _, d := range diagnostics {
		// This rule offers no repairs. A new repair must fail the pin rather than be omitted from the wire.
		if len(d.Fixes) != 0 || len(d.Suggestions) != 0 {
			panic("typed rule gained a repair")
		}
		var rendered strings.Builder
		report.Write(&rendered, []rule.Diagnostic{d}, report.Coverage{})
		display := rendered.String()
		footer := strings.LastIndex(display, "✗ cohere (")
		if footer < 0 {
			panic("cohere report shape changed")
		}
		fmt.Print(display[:footer])
		start, end := d.Range.Pos(), d.Range.End()
		fmt.Printf("range %d %d %s \t\t\t%d %d\n", start, end, d.Message.Id, start, end)
	}
	fmt.Printf("fixed\t%s\n", written(file.Text()))
}
