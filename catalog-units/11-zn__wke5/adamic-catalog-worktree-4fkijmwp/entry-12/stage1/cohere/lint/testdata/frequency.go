// Run inside cohere through a Go overlay. No Adamic code participates in ranking.
package main

import (
	"bufio"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"os"
	"sort"
	"strings"
)

type result struct {
	name                 string
	compiler, repository int
	failed               string
}

func count(subject rule.Rule, file *ast.SourceFile) (count int, failure string) {
	defer func() {
		if value := recover(); value != nil {
			failure = fmt.Sprint(value)
		}
	}()
	ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(rule.Diagnostic) { count++ }}
	listeners := subject.Run(ctx, nil)
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if visit := listeners[n.Kind]; visit != nil {
			visit(n)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	return
}
func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	registrations := rule.Registered()
	_ = registry.Count()
	registrations = rule.Registered()
	files := []*ast.SourceFile{}
	groups := []string{}
	for _, row := range strings.Split(string(data), "\n") {
		if row == "" {
			continue
		}
		parts := strings.SplitN(row, "\t", 2)
		data, err := os.ReadFile(parts[1])
		if err != nil {
			panic(err)
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(parts[1]), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(parts[1]))}, string(data), core.ScriptKindTS)
		if len(file.Diagnostics()) != 0 {
			fmt.Fprintf(out, "skip-parse\t%s\n", parts[1])
			continue
		}
		files = append(files, file)
		groups = append(groups, parts[0])
	}
	fmt.Fprintf(out, "files\t%d\n", len(files))
	results := []result{}
	for _, registration := range registrations {
		subject := registration.Rule
		if subject.NeedsTypeChecker {
			fmt.Fprintf(out, "skip-types\t%s\n", subject.Name)
			continue
		}
		if registration.RequiresOptions {
			fmt.Fprintf(out, "skip-required-options\t%s\n", subject.Name)
			continue
		}
		r := result{name: subject.Name}
		for i, file := range files {
			n, failure := count(subject, file)
			if failure != "" {
				r.failed = failure
				break
			}
			if groups[i] == "compiler" {
				r.compiler += n
			} else {
				r.repository += n
			}
		}
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.compiler+a.repository != b.compiler+b.repository {
			return a.compiler+a.repository > b.compiler+b.repository
		}
		return a.name < b.name
	})
	fmt.Fprintln(out, "rule\tcompiler\trepository\ttotal\tfailure")
	for _, r := range results {
		fmt.Fprintf(out, "%s\t%d\t%d\t%d\t%s\n", r.name, r.compiler, r.repository, r.compiler+r.repository, strings.ReplaceAll(r.failed, "\n", " "))
	}
}
