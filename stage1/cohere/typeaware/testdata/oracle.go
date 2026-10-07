// Built as an overlay inside cohere. Calls its unmodified production rule,
// with an independent program loader and walk; imports no bridge code.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	type_checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}
func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		panic("usage: oracle tsconfig manifest [--count]")
	}
	manifest, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	var paths []string
	for _, path := range strings.Split(string(manifest), "\n") {
		if path != "" {
			paths = append(paths, path)
		}
	}
	configPath, err := filepath.Abs(args[0])
	if err != nil {
		panic(err)
	}
	started := time.Now()
	host := compiler.NewCachedFSCompilerHost(filepath.ToSlash(filepath.Dir(configPath)), bundled.WrapFS(osvfs.FS()), bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(filepath.ToSlash(configPath), nil, nil, host, nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		panic("invalid tsconfig")
	}
	roots := make([]string, len(paths))
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(configPath), path)
		}
		roots[i] = filepath.ToSlash(path)
	}
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames(roots), Host: host, SingleThreaded: core.TSTrue})
	loadTime := time.Since(started)
	if len(args) > 2 && args[2] == "--query-cost" {
		count, err := strconv.Atoi(args[3])
		if err != nil || count < 2 {
			panic("invalid query count")
		}
		if len(roots) != 1 {
			panic("query cost requires one probe file")
		}
		file := program.GetSourceFile(roots[0])
		node := file.Statements.Nodes[0].AsExpressionStatement().Expression
		if node.Kind != ast.KindPrefixUnaryExpression || node.Pos() != 0 || node.End() != 2 {
			panic("query cost probe changed")
		}
		typeChecker, release := program.GetTypeCheckerForFile(context.Background(), file)
		defer release()
		units := 0
		var total, first time.Duration
		for index := 0; index < count; index++ {
			before := time.Now()
			actual := type_checking.GetConstrainedTypeAtLocation(typeChecker, node)
			var frame strings.Builder
			for part := range type_checking.UnionTypePartsSeq(actual) {
				name := typeChecker.TypeToString(part)
				fmt.Fprintf(&frame, "%d\n%d\n%s", part.Flags(), len(utf16.Encode([]rune(name))), name)
			}
			units += len(utf16.Encode([]rune(frame.String())))
			elapsed := time.Since(before)
			if index == 0 {
				first = elapsed
			}
			total += elapsed
		}
		fmt.Println(units)
		fmt.Fprintf(os.Stderr, "querycost: load_ns=%d query_ns=%d queries=%d first_query_ns=%d\n", loadTime.Nanoseconds(), total.Nanoseconds(), count, first.Nanoseconds())
		return
	}

	countOnly := len(args) > 2 && args[2] == "--count"
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	findings, queries := 0, 0
	var ruleTime time.Duration
	for i, path := range paths {
		file := program.GetSourceFile(roots[i])
		if file == nil {
			panic("source not loaded")
		}
		if len(file.Diagnostics()) != 0 {
			panic(fmt.Sprintf("parse diagnostics: %s", path))
		}
		typeChecker, release := program.GetTypeCheckerForFile(context.Background(), file)
		var collected []rule.Diagnostic
		subject := rules.NoUnsafeUnaryMinus
		listeners := subject.Run(rule.Context{SourceFile: file, TypeChecker: typeChecker, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) {
			d.RuleName = subject.Name
			collected = append(collected, d)
		}}, nil)
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
				queries++
			}
			if visit := listeners[node.Kind]; visit != nil {
				before := time.Now()
				visit(node)
				ruleTime += time.Since(before)
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(file.AsNode())
		release()
		sort.SliceStable(collected, func(i, j int) bool { return collected[i].Range.Pos() < collected[j].Range.Pos() })
		findings += len(collected)
		if !countOnly {
			fmt.Fprintf(out, "file\t%s\n", written(path))
			for _, d := range collected {
				if len(d.Fixes) != 0 || len(d.Suggestions) != 0 {
					panic("unexpected repair")
				}
				fmt.Fprintf(out, "%d\t%d\t%s\t%s\t%s\t%d\t%d\n", d.Range.Pos(), d.Range.End(), d.RuleName, d.Message.Id, written(d.Message.Description), len(d.Fixes), len(d.Suggestions))
			}
		}
	}
	fmt.Fprintf(out, "findings %d queries %d\n", findings, queries)
	fmt.Fprintf(os.Stderr, "cohere: load_ns=%d rule_ns=%d queries=%d\n", loadTime.Nanoseconds(), ruleTime.Nanoseconds(), queries)
}
