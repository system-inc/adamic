// Built as an overlay inside cohere. Calls its unmodified production rule,
// with an independent program loader and walk; imports no bridge code.
package main

import (
	"bufio"
	"context"
	"fmt"
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
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
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
	if len(args) < 3 {
		panic("usage: source-gate tsconfig manifest lint-config")
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
	host := compiler.NewCachedFSCompilerHost(filepath.ToSlash(filepath.Dir(configPath)), sourceExtensionFS{bundled.WrapFS(osvfs.FS())}, bundled.LibPath(), nil, nil, nil)
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
		if strings.HasSuffix(roots[i], ".a") {
			roots[i] += ".ts"
		}
	}
	config.CompilerOptions().AllowNonTsExtensions = core.TSTrue
	for _, path := range config.FileNames() {
		if strings.HasSuffix(path, ".d.ts") && !slices.Contains(roots, path) {
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
	lintConfig, err := configuration.Load(args[2])
	if err != nil {
		panic(err)
	}
	countOnly := false
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
			if !lintConfig.Resolve(path).Ignored && lintConfig.Resolve(path).Rules[subject.Name].Severity != configuration.SeverityOff {
				subjects = append(subjects, subject)
			}
		}
		callbacks := map[ast.Kind][]func(*ast.Node){}
		cache := rule.NewFileCache()
		for _, subject := range subjects {
			options, err := registry.Options().Decode(subject.Name, lintConfig.Resolve(path).RawOptionsFor(subject.Name))
			if err != nil {
				panic(err)
			}
			listeners := subject.Run(rule.Context{SourceFile: file, TypeChecker: typeChecker, Program: rule.ViewProgram(program, file, subject), FileCache: cache, Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; collected = append(collected, d) }}, options)
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

// The pinned CLI predates .a discovery and module resolution. Only this source
// gate presents virtual TypeScript names; it writes no .ts files and is not used
// by the rule findings oracle or the checker bridge.
type sourceExtensionFS struct{ vfs.FS }

func (f sourceExtensionFS) FileExists(path string) bool {
	if strings.HasSuffix(path, ".a.ts") {
		path = strings.TrimSuffix(path, ".ts")
	}
	return f.FS.FileExists(path)
}
func (f sourceExtensionFS) ReadFile(path string) (string, bool) {
	mapped := strings.HasSuffix(path, ".a.ts")
	if mapped {
		path = strings.TrimSuffix(path, ".ts")
	}
	text, ok := f.FS.ReadFile(path)
	if mapped {
		text = strings.ReplaceAll(text, ".a';", ".a.ts';")
	}
	return text, ok
}
