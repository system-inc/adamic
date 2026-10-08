// Built as an overlay inside cohere. Calls its unmodified production rule,
// with an independent program loader and walk; imports no bridge code.
package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	corerules "github.com/system-inc/cohere/internal/lint/rules/core"
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

var coverageNames = map[string]bool{"no-throw-literal": true, "no-useless-backreference": true, "prefer-arrow-callback": true}

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
	fs := bundled.WrapFS(osvfs.FS())
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(configPath)), nil, nil, fs, nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		panic("invalid tsconfig")
	}
	roots := make([]tspath.RootedFilePath, len(paths))
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(configPath), path)
		}
		roots[i] = tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path))
	}
	config.CompilerOptions().AllowNonTsExtensions = core.TSTrue
	for _, path := range config.FileNames() {
		if strings.HasSuffix(path.AsString(), ".d.ts") && !slices.Contains(roots, path) {
			roots = append(roots, path)
		}
	}
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames(roots), Host: host, SingleThreaded: core.TSTrue})
	// Force checker pool initialization into load, just as the native bridge does.
	if len(roots) > 0 {
		_, release := program.GetTypeCheckerForFile(context.Background(), program.GetSourceFile(roots[0]))
		release()
	}
	loadTime := time.Since(started)
	runStarted := time.Now()

	if len(args) > 2 && args[2] == "--valid-sources" {
		for i, path := range paths {
			file := program.GetSourceFile(roots[i])
			if file != nil && len(file.Diagnostics()) == 0 {
				fmt.Println(path)
			}
		}
		return
	}
	countOnly := len(args) > 2 && args[2] == "--count"
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	findings := 0
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
		var subjects []rule.Rule
		for _, subject := range registry.All() {
			if coverageNames[subject.Name] {
				subjects = append(subjects, subject)
			}
		}
		callbacks := map[ast.Kind][]func(*ast.Node){}
		cache := rule.NewFileCache()
		for _, subject := range subjects {
			var settings any
			if subject.Name == "prefer-arrow-callback" {
				settings = corerules.PreferArrowCallbackSettings{AllowNamedFunctions: slices.Contains(args, "--allow-named"), AllowUnboundThis: !slices.Contains(args, "--reject-unbound")}
			}
			listeners := subject.Run(rule.Context{SourceFile: file, TypeChecker: typeChecker, Program: rule.ViewProgram(program, file, subject), FileCache: cache, Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; collected = append(collected, d) }}, settings)
			for kind, callback := range listeners {
				callbacks[kind] = append(callbacks[kind], callback)
			}
		}
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			for _, visit := range callbacks[node.Kind] {
				before := time.Now()
				visit(node)
				ruleTime += time.Since(before)
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(file.AsNode())
		release()
		sort.SliceStable(collected, func(i, j int) bool { return diagnostic(collected[i]) < diagnostic(collected[j]) })
		findings += len(collected)
		if !countOnly {
			fmt.Fprintf(out, "file\t%s\n", written(path))
			for _, d := range collected {
				fmt.Fprintln(out, diagnostic(d))
			}
		}
	}
	fmt.Fprintf(out, "findings %d\n", findings)
	fmt.Fprintf(os.Stderr, "cohere: load_ns=%d rule_ns=%d run_ns=%d\n", loadTime.Nanoseconds(), ruleTime.Nanoseconds(), time.Since(runStarted).Nanoseconds())
}

func diagnostic(d rule.Diagnostic) string {
	line := fmt.Sprintf("%d\t%d\t%s\t%s\t%s\t%d\t%d", d.Range.Pos(), d.Range.End(), d.RuleName, d.Message.Id, written(d.Message.Description), len(d.Fixes), len(d.Suggestions))
	for _, fix := range d.Fixes {
		line += fmt.Sprintf("\t%d\t%d\t%s", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
	}
	for _, suggestion := range d.Suggestions {
		line += fmt.Sprintf("\t%s\t%s\t%d", suggestion.Message.Id, written(suggestion.Message.Description), len(suggestion.Fixes))
		for _, fix := range suggestion.Fixes {
			line += fmt.Sprintf("\t%d\t%d\t%s", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
		}
	}
	return line
}
