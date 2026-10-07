// Built through an overlay inside cohere, using its unmodified syntax-only rules.
package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/report"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
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

// The wave workers' compact protocol is retained where its delimiters cannot collide.
func simpleSuggestion(fixes []rule.Fix) bool {
	if len(fixes) == 0 {
		return false
	}
	for _, fix := range fixes {
		if strings.ContainsAny(fix.Text, "|:") {
			return false
		}
	}
	return true
}
func run(row string, countOnly bool, out *bufio.Writer, graph *program.Graph) int {
	fields := strings.Split(row, "\t")
	for len(fields) < 7 {
		fields = append(fields, "")
	}
	path := fields[0]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	source := string(data)
	if graph == nil {
		for _, item := range registeredRules() {
			if item.subject.NeedsTypeChecker && (fields[1] == "" || fields[1] == "all" || fields[1] == item.subject.Name) {
				fmt.Fprintf(out, "skipped %s no program\n", item.subject.Name)
			}
		}
	}
	diagnostics := collect(path, source, fields, graph)
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
		if len(d.Fixes) > 0 {
			repair = "fix"
			replacement = d.Fixes[0].Text
		}
		if len(d.Suggestions) > 0 {
			s := d.Suggestions[0]
			if len(d.Suggestions) == 1 && len(s.Fixes) == 1 && s.Fixes[0].Range == d.Range {
				repair = "suggestion"
				replacement = s.Fixes[0].Text
				suggestion = s.Message.Description
			} else if len(d.Suggestions) == 1 && simpleSuggestion(s.Fixes) {
				repair = "suggestion-edits:" + s.Message.Id
				var edits []string
				for _, fix := range s.Fixes {
					edits = append(edits, fmt.Sprintf("%d:%d:%s", fix.Range.Pos(), fix.Range.End(), fix.Text))
				}
				replacement = strings.Join(edits, "|")
				suggestion = s.Message.Description
			} else {
				repair, replacement, suggestion = "suggestions", "", ""
			}
		}
		editStart, editEnd := start, end
		if len(d.Fixes) > 0 {
			editStart = d.Fixes[0].Range.Pos()
			editEnd = d.Fixes[0].Range.End()
		}
		if repair == "suggestion" {
			editStart = d.Suggestions[0].Fixes[0].Range.Pos()
			editEnd = d.Suggestions[0].Fixes[0].Range.End()
		}
		fmt.Fprintf(out, "range %d %d %s %s\t%s\t%s\t%d %d\n", start, end, d.Message.Id, repair, written(replacement), written(suggestion), editStart, editEnd)
		// A finding with several automatic edits: cohere's edit engine proposes each one on its own
		// (edit.ProposalsFrom), so the first rides the range line and the rest follow, in order.
		if len(d.Fixes) > 1 {
			for _, fix := range d.Fixes[1:] {
				fmt.Fprintf(out, "fix-edit\t%d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
			}
		}
		if repair == "suggestions" {
			for _, suggestion := range d.Suggestions {
				fmt.Fprintf(out, "suggestion\t%s\t%s\t%d\n", written(suggestion.Message.Id), written(suggestion.Message.Description), len(suggestion.Fixes))
				for _, fix := range suggestion.Fixes {
					fmt.Fprintf(out, "suggestion-edit\t%d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
				}
			}
		}
	}
	if fields[6] == "recovery" {
		fmt.Fprintln(out, "recovery findings only")
		return len(diagnostics)
	}
	firstPass := true
	result, err := edit.FixText(path, source, func(fileName, text string) ([]edit.Proposal, error) {
		if firstPass {
			firstPass = false
			return edit.ProposalsFrom(diagnostics), nil
		}
		return edit.ProposalsFrom(collect(fileName, text, fields, nil)), nil
	}, 10)
	if err != nil {
		panic(fmt.Sprintf("fix failed: %v %+v", err, result))
	}
	for _, rejection := range result.Rejected {
		fmt.Fprintf(out, "rejected %s %d %d %s %s\n", rejection.Proposal.RuleName, rejection.Proposal.Fix.Range.Pos(), rejection.Proposal.Fix.Range.End(), rejection.ConflictsWith, rejection.Reason)
	}
	// A run that exhausts the pass budget is cohere's answer too, not a harness failure: the file is left
	// as found, with a fix-engine rejection, and the rules still proposing are named.
	if !result.Converged {
		fmt.Fprintf(out, "unconverged\t%s\n", strings.Join(result.UnconvergedRules, ","))
	}
	fixed := result.Text
	fmt.Fprintf(out, "fixed\t%s\n", written(fixed))
	return len(diagnostics)
}

// parse is the file as typescript-go parses it, in the script kind its extension names.
func parse(path, source string) *ast.SourceFile {
	kind := core.ScriptKindTS
	switch {
	case strings.HasSuffix(path, ".tsx"):
		kind = core.ScriptKindTSX
	case strings.HasSuffix(path, ".jsx"):
		kind = core.ScriptKindJSX
	case strings.HasSuffix(path, ".js"):
		kind = core.ScriptKindJS
	}
	return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(path), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(path))}, source, kind)
}

func collect(path, source string, fields []string, graph *program.Graph) []rule.Diagnostic {
	file := parse(path, source)
	if graph != nil {
		absolute, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		file = graph.Program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(absolute)))
		if file == nil || file.Text() != source {
			panic("typed row is not the program's source: " + path)
		}
	}
	if len(file.Diagnostics()) != 0 && fields[6] != "recovery" {
		panic(fmt.Sprintf("invalid corpus %s: %v; source=%q", path, file.Diagnostics(), source))
	}
	// The registry's order is the listener order: the five first rules by their pinned order, then the rest
	// by public name. The port dispatches each node in the same order, so ties at one position agree.
	var diagnostics []rule.Diagnostic
	var listeners []rule.Listeners
	checkerContext := rule.Context{}
	if graph != nil {
		checker, release := graph.CheckerForFile(context.Background(), file)
		defer release()
		checkerContext.TypeChecker = checker
	}
	for _, item := range registeredRules() {
		subject := item.subject
		if fields[1] != "" && fields[1] != "all" && fields[1] != subject.Name {
			continue
		}
		if subject.NeedsTypeChecker && graph == nil {
			continue
		}
		ctx := rule.Context{TypeChecker: checkerContext.TypeChecker, SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; diagnostics = append(diagnostics, d) }}
		if graph != nil {
			ctx.Program = rule.ViewProgram(graph.Program, file, subject)
		}
		options := item.options(fields)
		// A row that selects this rule and carries options must reach an adapter that decodes them. An
		// adapter returning nil there would run the rule on its defaults and still agree with any port that
		// reads no options either. An "all" row's options are one bag for every rule, so a rule with none of
		// its own ignores them there.
		if options == nil && fields[1] == subject.Name && fields[5] != "" && fields[5] != "null" {
			panic(fmt.Sprintf("%s: options %s reached an adapter that decodes none", subject.Name, fields[5]))
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
		run(args[0], false, out, nil)
		return
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	var graph *program.Graph
	rows := strings.Split(string(data), "\n")
	if len(rows) > 0 && strings.HasPrefix(rows[0], "program ") {
		config, err := filepath.Abs(strings.TrimPrefix(rows[0], "program "))
		if err != nil {
			panic(err)
		}
		graph, err = program.Build(program.Options{ConfigFileName: config, CurrentDirectory: filepath.Dir(config), SingleThreaded: true})
		if err != nil {
			panic(err)
		}
		rows = rows[1:]
	}
	// --diagnostics answers, per row, whether typescript-go's parse of its file reports a diagnostic: 1 or 0.
	// A test marks such a row "recovery", which compares findings only, rather than asking the oracle to
	// fix a file Go would refuse (a legacy octal escape, say, which no-octal-escape exists to report).
	if len(args) > 2 && args[2] == "--diagnostics" {
		for _, row := range rows {
			if row == "" {
				continue
			}
			path := strings.Split(row, "\t")[0]
			source, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}
			if len(parse(path, string(source)).Diagnostics()) != 0 {
				fmt.Fprintln(out, 1)
			} else {
				fmt.Fprintln(out, 0)
			}
		}
		return
	}
	countOnly := len(args) > 2 && args[2] == "--count"
	count, index := 0, 0
	for _, row := range rows {
		if row == "" {
			continue
		}
		if !countOnly {
			fmt.Fprintf(out, "case %d\n", index)
		}
		count += run(row, countOnly, out, graph)
		index++
	}
	if countOnly {
		fmt.Fprintln(out, count)
	}
}
