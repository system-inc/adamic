// Built through an overlay inside cohere, using its unmodified syntax-only rules.
package main

import (
	"bufio"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/report"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	react "github.com/system-inc/cohere/internal/lint/rules/react"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"
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
func run(row string, countOnly bool, out *bufio.Writer) int {
	fields := strings.Split(row, "\t")
	for len(fields) < 5 {
		fields = append(fields, "")
	}
	path := fields[0]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	source := string(data)
	diagnostics := collect(path, source, fields)
	if countOnly {
		return len(diagnostics)
	}
	for _, d := range diagnostics {
		start, end := d.Range.Pos(), d.Range.End()
		var rendered strings.Builder
		report.Write(&rendered, []rule.Diagnostic{d}, report.Coverage{})
		display := rendered.String()
		footer := strings.LastIndex(display, "✗ cohere (")
		if footer < 0 {
			panic("cohere report shape changed")
		}
		fmt.Fprint(out, display[:footer])
		fmt.Fprintf(out, "range %d %d %s\n", start, end, d.Message.Id)
		for _, fix := range d.Fixes {
			fmt.Fprintf(out, "fix %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
		}
		for _, suggestion := range d.Suggestions {
			fmt.Fprintf(out, "suggestion %s\t%s\n", suggestion.Message.Id, written(suggestion.Message.Description))
			for _, fix := range suggestion.Fixes {
				fmt.Fprintf(out, "edit %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
			}
		}
	}
	fixed := source
	if len(edit.ProposalsFrom(diagnostics)) != 0 {
		result, err := edit.FixText(path, source, func(fileName, text string) ([]edit.Proposal, error) {
			return edit.ProposalsFrom(collect(fileName, text, fields)), nil
		}, 10)
		if err != nil || len(result.Rejected) != 0 || !result.Converged {
			panic(fmt.Sprintf("fix failed: %v %+v", err, result))
		}
		fixed = result.Text
	}
	fmt.Fprintf(out, "fixed\t%s\n", written(fixed))
	return len(diagnostics)
}
func collect(path, source string, fields []string) []rule.Diagnostic {
	kind := core.ScriptKindTS
	switch filepath.Ext(path) {
	case ".tsx":
		kind = core.ScriptKindTSX
	case ".jsx":
		kind = core.ScriptKindJSX
	case ".js":
		kind = core.ScriptKindJS
	}
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, source, kind)
	// Fixture capture includes parser recovery cases; the parity harness records those separately.
	selected := []rule.Rule{rules.NoOctalEscape, rules.NoUnexpectedMultiline, rules.NoUnusedPrivateClassMembers, rules.NoUselessConstructor, rules.PreferTemplate, react.ForwardRefUsesRef, react.JsxNoCommentTextnodes, react.NoFindDOMNode, react.NoIsMounted, react.NoRedundantShouldComponentUpdate}
	var diagnostics []rule.Diagnostic
	var listeners []rule.Listeners
	for _, subject := range selected {
		if fields[1] != "" && fields[1] != "all" && fields[1] != subject.Name {
			continue
		}
		ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; diagnostics = append(diagnostics, d) }}
		var options any
		if subject.Name == "eqeqeq" {
			options = rules.EqeqeqOptions{Mode: rules.EqeqeqMode(fields[2]), Null: rules.EqeqeqNullPolicy(fields[3])}
		}
		if subject.Name == "no-empty" {
			options = rules.NoEmptyOptions{AllowEmptyCatch: fields[4] == "true"}
		}
		listeners = append(listeners, subject.Run(ctx, options))
	}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		for _, listener := range listeners {
			if visit := listener[n.Kind]; visit != nil {
				visit(n)
			}
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	sort.SliceStable(diagnostics, func(i, j int) bool { return diagnostics[i].Range.Pos() < diagnostics[j].Range.Pos() })
	return diagnostics
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("missing file")
	}
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	if args[0] == "--jsx-spans" {
		data, err := os.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		kind := core.ScriptKindTS
		switch filepath.Ext(args[1]) {
		case ".tsx":
			kind = core.ScriptKindTSX
		case ".jsx":
			kind = core.ScriptKindJSX
		case ".js":
			kind = core.ScriptKindJS
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: args[1], Path: tspath.Path(args[1])}, string(data), kind)
		jsx := false
		texts := 0
		var walk func(*ast.Node)
		walk = func(n *ast.Node) {
			if n.Kind == ast.KindJsxElement || n.Kind == ast.KindJsxSelfClosingElement || n.Kind == ast.KindJsxFragment {
				jsx = true
			}
			if n.Kind == ast.KindJsxText {
				texts++
				fmt.Fprintf(out, "%d:%d\n", n.Pos(), n.End())
			}
			n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(file.AsNode())
		if jsx && texts == 0 {
			fmt.Fprintln(out, "empty")
		}
		return
	}
	if args[0] != "--manifest" {
		run(args[0], false, out)
		return
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	countOnly := len(args) > 2 && args[2] == "--count"
	count, index := 0, 0
	for _, row := range strings.Split(string(data), "\n") {
		if row == "" {
			continue
		}
		if !countOnly {
			fmt.Fprintf(out, "case %d\n", index)
		}
		count += run(row, countOnly, out)
		index++
	}
	if countOnly {
		fmt.Fprintln(out, count)
	}
}
