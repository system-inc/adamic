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
	nextrules "github.com/system-inc/cohere/internal/lint/rules/next"
	tsrules "github.com/system-inc/cohere/internal/lint/rules/typescript"
	"os"
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
func extraFixes(fixes []rule.Fix) []rule.Fix {
	if len(fixes) < 2 {
		return nil
	}
	return fixes[1:]
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
		repair, replacement, suggestion := "", "", ""
		editStart, editEnd := start, end
		if len(d.Fixes) > 0 {
			repair = "fix"
			replacement = d.Fixes[0].Text
			editStart, editEnd = d.Fixes[0].Range.Pos(), d.Fixes[0].Range.End()
		}
		if len(d.Suggestions) > 0 {
			first := d.Suggestions[0]
			repair = "suggestion"
			replacement = first.Fixes[0].Text
			suggestion = first.Message.Description
			editStart, editEnd = first.Fixes[0].Range.Pos(), first.Fixes[0].Range.End()
		}
		fmt.Fprintf(out, "range %d %d %s %s\t%s\t%s\t%d %d\n", start, end, d.Message.Id, repair, written(replacement), written(suggestion), editStart, editEnd)
		for _, suggestion := range d.Suggestions {
			fmt.Fprintf(out, "suggestion %s\t%s\n", suggestion.Message.Id, written(suggestion.Message.Description))
			for _, fix := range suggestion.Fixes {
				fmt.Fprintf(out, "suggestion-edit %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
			}
		}
		for _, fix := range extraFixes(d.Fixes) {
			fmt.Fprintf(out, "edit %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
		}

	}
	result, err := edit.FixText(path, source, func(fileName, text string) ([]edit.Proposal, error) {
		return edit.ProposalsFrom(collect(fileName, text, fields)), nil
	}, 10)
	if err != nil || !result.Converged {
		panic(fmt.Sprintf("fix failed: %v %+v", err, result))
	}
	for _, rejection := range result.Rejected {
		fmt.Fprintf(out, "rejected %s %d %d %s %s\n", rejection.Proposal.RuleName, rejection.Proposal.Fix.Range.Pos(), rejection.Proposal.Fix.Range.End(), rejection.ConflictsWith, rejection.Reason)
	}
	fixed := result.Text
	fmt.Fprintf(out, "fixed\t%s\n", written(fixed))
	return len(diagnostics)
}
func collect(path, source string, fields []string) []rule.Diagnostic {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, source, core.ScriptKindTS)
	if len(file.Diagnostics()) != 0 {
		panic(fmt.Sprintf("invalid corpus %s: %v", path, file.Diagnostics()))
	}
	selected := []rule.Rule{rules.NoDebugger, rules.NoEmpty, rules.Eqeqeq, rules.NoVar, rules.NoDuplicateCase, nextrules.NoAssignModuleVariable, tsrules.DefaultParamLast, tsrules.NoConfusingNonNullAssertion, tsrules.NoDuplicateEnumValues, tsrules.NoDynamicDelete, tsrules.NoExtraNonNullAssertion, tsrules.NoMisusedNew, tsrules.NoUnnecessaryParameterPropertyAssignment, tsrules.PreferAsConst, rules.DefaultCaseLast}
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
