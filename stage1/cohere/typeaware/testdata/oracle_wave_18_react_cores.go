// Built as an overlay inside cohere. Calls its unmodified production rule,
// with an independent program loader and walk; imports no bridge code.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
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

var coverageNames = map[string]bool{"react-hooks/set-state-in-effect": true, "react-hooks/set-state-in-render": true, "react-hooks/static-components": true}

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
	if len(args) > 3 && args[2] == "--native-graphs" {
		nativeGraphs(program, roots, paths, args[3])
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
			listeners := subject.Run(rule.Context{SourceFile: file, TypeChecker: typeChecker, Program: rule.ViewProgram(program, file, subject), FileCache: cache, Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; collected = append(collected, d) }}, nil)
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

// Test-only prepared-HIR fixture producer, not a production source adapter.
func quote(s string) string { b, _ := json.Marshal(s); return string(b) }
func nativePlace(file *ast.SourceFile, p hir.Place) string {
	start, end := p.Range.Pos(), p.Range.End()
	if start >= 0 && end <= len(file.Text()) && start < end {
		if trimmed := scanner.SkipTrivia(file.Text(), start); trimmed < end {
			start = trimmed
		}
	}
	return fmt.Sprintf("new HirPlace(%d,%d,%d)", p.Identifier, start, end)
}
func nativePlaces(file *ast.SourceFile, ps []hir.Place) string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = nativePlace(file, p)
	}
	return "[" + strings.Join(out, ",") + "]"
}
func nativeNumbers[T ~uint32](ps []T) string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = fmt.Sprint(p)
	}
	return "[" + strings.Join(out, ",") + "]"
}
func nativePattern(p hir.Pattern) []hir.IdentifierId {
	var out []hir.IdentifierId
	switch p := p.(type) {
	case *hir.PlacePattern:
		out = append(out, p.Place.Identifier)
	case *hir.ObjectPattern:
		for _, v := range p.Properties {
			out = append(out, nativePattern(v.Value)...)
			if v.Default != nil {
				out = append(out, v.Default.Identifier)
			}
		}
		if p.Rest != nil {
			out = append(out, p.Rest.Identifier)
		}
	case *hir.ArrayPattern:
		for _, v := range p.Elements {
			out = append(out, nativePattern(v.Value)...)
			if v.Default != nil {
				out = append(out, v.Default.Identifier)
			}
		}
		if p.Rest != nil {
			out = append(out, p.Rest.Identifier)
		}
	}
	return out
}
func nativeInstruction(file *ast.SourceFile, i *hir.Instruction) string {
	kind := reflect.TypeOf(i.Value).Elem().Name()
	zero := "new HirPlace(-1,-1,-1)"
	p0, p1, p2 := zero, zero, zero
	args, spreads, captures, pattern := "[]", "[]", "[]", "[]"
	nested := -1
	property := ""
	optional := false
	setArgs := func(values []hir.Argument) {
		places := make([]hir.Place, len(values))
		flags := make([]string, len(values))
		for j, v := range values {
			places[j] = v.Place
			flags[j] = fmt.Sprint(v.Spread)
		}
		args = nativePlaces(file, places)
		spreads = "[" + strings.Join(flags, ",") + "]"
	}
	switch v := i.Value.(type) {
	case *hir.LoadLocal:
		p0 = nativePlace(file, v.Place)
	case *hir.LoadContext:
		p0 = nativePlace(file, v.Place)
	case *hir.StoreLocal:
		p0 = nativePlace(file, v.LValue)
		p1 = nativePlace(file, v.Value)
	case *hir.Destructure:
		p0 = nativePlace(file, v.Value)
		pattern = nativeNumbers(nativePattern(v.LValue))
	case *hir.PropertyLoad:
		p0 = nativePlace(file, v.Object)
		property = v.Property
		optional = v.Optional
	case *hir.ComputedLoad:
		p0 = nativePlace(file, v.Object)
		p1 = nativePlace(file, v.Property)
		optional = v.Optional
	case *hir.BinaryExpression:
		p0 = nativePlace(file, v.Left)
		p1 = nativePlace(file, v.Right)
	case *hir.UnaryExpression:
		p0 = nativePlace(file, v.Value)
	case *hir.CallExpression:
		p0 = nativePlace(file, v.Callee)
		setArgs(v.Args)
		optional = v.Optional
	case *hir.MethodCall:
		p0 = nativePlace(file, v.Receiver)
		p1 = nativePlace(file, v.Property)
		setArgs(v.Args)
		optional = v.Optional
	case *hir.NewExpression:
		p0 = nativePlace(file, v.Callee)
		setArgs(v.Args)
	case *hir.FunctionExpression:
		nested = int(v.Function)
		captures = nativePlaces(file, v.Captures)
	case *hir.JsxExpression:
		if v.Tag.Place != nil {
			p0 = nativePlace(file, *v.Tag.Place)
		}
	}
	return fmt.Sprintf("new HirInstruction(%s,%s,%s,%s,%s,%s,%s,%s,%s,%d,%s,%t)", quote(kind), nativePlace(file, i.LValue), p0, p1, p2, args, spreads, captures, pattern, nested, quote(property), optional)
}
func nativeMemo(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := node.Parent.AsCallExpression()
	if call == nil || len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != node {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind == ast.KindIdentifier {
		return callee.Text() == "useMemo"
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		p := callee.AsPropertyAccessExpression()
		o := ast.SkipParentheses(p.Expression)
		return p.QuestionDotToken == nil && p.Name() != nil && p.Name().Text() == "useMemo" && o.Kind == ast.KindIdentifier && o.Text() == "React"
	}
	return false
}
func nativeFunction(file *ast.SourceFile, c *shimchecker.Checker, f *hir.Function) string {
	identities := make([]string, len(f.Identifiers))
	for j, n := range f.Identifiers {
		start, end := -1, -1
		text, alias, symbol := "", "", ""
		if n != nil && n.Node != nil {
			span := rule.TokenRange(file, n.Node)
			start, end = span.Pos(), span.End()
			if start >= 0 && end <= len(file.Text()) && start < end {
				text = file.Text()[start:end]
			}
			valueType := c.GetTypeAtLocation(n.Node)
			if valueType != nil {
				if a := shimchecker.Type_alias(valueType); a != nil && a.Symbol() != nil {
					alias = a.Symbol().Name
				}
				if sym := shimchecker.Type_symbol(valueType); sym != nil {
					symbol = sym.Name
				}
			}
		}
		identities[j] = fmt.Sprintf("new HirIdentity(%d,%d,%s,%s,%s)", start, end, quote(text), quote(alias), quote(symbol))
	}
	blocks := make([]string, len(f.Blocks))
	for j, b := range f.Blocks {
		var instructions, phis []string
		for _, id := range b.Instructions {
			if i := f.Instructions[id]; i != nil {
				instructions = append(instructions, nativeInstruction(file, i))
			}
		}
		for _, p := range b.Phis {
			var operands []hir.IdentifierId
			for _, o := range p.Operands {
				operands = append(operands, o.Identifier)
			}
			phis = append(phis, fmt.Sprintf("new HirPhi(%d,%s)", p.Place.Identifier, nativeNumbers(operands)))
		}
		var tests []hir.Place
		switch term := b.Terminal.(type) {
		case *hir.If:
			tests = append(tests, term.Test)
		case *hir.Branch:
			tests = append(tests, term.Test)
		case *hir.Switch:
			tests = append(tests, term.Test)
			for _, k := range term.Cases {
				if k.Test != nil {
					tests = append(tests, *k.Test)
				}
			}
		}
		_, returning := b.Terminal.(*hir.Return)
		blocks[j] = fmt.Sprintf("new HirBlock(%d,%s,%t,%s,[%s],[%s])", b.Id, nativeNumbers(b.Predecessors), returning, nativePlaces(file, tests), strings.Join(instructions, ","), strings.Join(phis, ","))
	}
	nested := make([]string, len(f.Functions))
	for j, n := range f.Functions {
		nested[j] = nativeFunction(file, c, n)
	}
	return fmt.Sprintf("new HirFunction(%d,[%s],%s,%t,[%s],[%s])", f.Entry, strings.Join(identities, ","), nativePlaces(file, f.Context), nativeMemo(f.Node), strings.Join(blocks, ","), strings.Join(nested, ","))
}
func nativeCreatesJSX(node *ast.Node) bool {
	found := false
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || found {
			return
		}
		switch n.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			found = true
			return
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor, ast.KindClassDeclaration, ast.KindClassExpression:
			return
		}
		n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
	}
	if node != nil {
		walk(node.Body())
	}
	return found
}
func nativeGraphs(program *compiler.Program, roots, paths []string, modules string) {
	fmt.Printf("import { Diagnostic } from %s;\n", quote(filepath.Join(modules, "../../diagnostic.ts")))
	fmt.Printf("import { HirPlace,HirIdentity,HirPhi,HirInstruction,HirBlock,HirFunction } from %s;\n", quote(filepath.Join(modules, "hir.a")))
	for _, n := range []struct{ class, file string }{{"SetStateInEffect", "set_state_in_effect.a"}, {"SetStateInRender", "set_state_in_render.a"}, {"StaticComponents", "static_components.a"}} {
		fmt.Printf("import { %s } from %s;\n", n.class, quote(filepath.Join(modules, n.file)))
	}
	fmt.Print("let count=0;\n")
	for j, path := range paths {
		file := program.GetSourceFile(roots[j])
		if file == nil || len(file.Diagnostics()) != 0 {
			panic("invalid graph fixture")
		}
		c, release := program.GetTypeCheckerForFile(context.Background(), file)
		fmt.Printf("const findings%d:Diagnostic[]=[];\n", j)
		ctx := rule.Context{SourceFile: file, TypeChecker: c, FileCache: rule.NewFileCache()}
		var outer func(*ast.Node)
		outer = func(n *ast.Node) {
			n.ForEachChild(func(n *ast.Node) bool {
				if ast.IsFunctionLike(n) {
					var selectUnit func(*hir.Function, string)
					selectUnit = func(f *hir.Function, kind string) {
						if f == nil {
							return
						}
						eligible := f.Node != nil && utilsreact.IsComponentOrHookLike(f.Node)
						if kind == "StaticComponents" {
							eligible = f.Kind != hir.FunctionKindOther && nativeCreatesJSX(f.Node)
						}
						if !eligible {
							for _, nested := range f.Functions {
								selectUnit(nested, kind)
							}
							return
						}
						lower := hir.ForFunction
						if kind == "SetStateInEffect" {
							lower = hir.ForFunctionWithoutManualMemoization
						}
						f = hir.AsCompilationUnit(ctx, f, lower)
						fmt.Printf("new %s(findings%d).run(%s);\n", kind, j, nativeFunction(file, c, f))
					}
					selectUnit(hir.ForFunction(ctx, n), "SetStateInRender")
					selectUnit(hir.ForFunction(ctx, n), "StaticComponents")
					selectUnit(hir.ForFunctionWithoutManualMemoization(ctx, n), "SetStateInEffect")
					return false
				}
				outer(n)
				return false
			})
		}
		outer(file.AsNode())
		release()
		fmt.Printf("console.log(%s);findings%d.sort((a,b)=>a.written()<b.written()?-1:a.written()>b.written()?1:0);for(const finding of findings%d){console.log(finding.written());}count+=findings%d.length;\n", quote("file\t"+written(path)), j, j, j)
	}
	fmt.Print("console.log(`findings ${count}`);\n")
}
