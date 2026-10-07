// Built through an overlay inside cohere, using its unmodified syntax-only rules.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/report"
	"github.com/system-inc/cohere/internal/lint/rule"
	adamic "github.com/system-inc/cohere/internal/lint/rules/adamic"
	base "github.com/system-inc/cohere/internal/lint/rules/base"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	nexus "github.com/system-inc/cohere/internal/lint/rules/nexus"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
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
func run(row string, countOnly bool, out *bufio.Writer) int {
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
		if len(d.Fixes) > 0 {
			if len(d.Fixes) != 1 {
				panic("unexpected fix shape")
			}
			repair = "fix"
			replacement = d.Fixes[0].Text
		}
		if len(d.Suggestions) > 0 {
			s := d.Suggestions[0]
			if len(d.Suggestions) != 1 || len(s.Fixes) != 1 {
				panic("unexpected suggestion shape")
			}
			repair = "suggestion"
			replacement = s.Fixes[0].Text
			suggestion = s.Message.Description
		}
		editStart, editEnd := start, end
		if len(d.Fixes) > 0 {
			editStart = d.Fixes[0].Range.Pos()
			editEnd = d.Fixes[0].Range.End()
		}
		if len(d.Suggestions) > 0 {
			editStart = d.Suggestions[0].Fixes[0].Range.Pos()
			editEnd = d.Suggestions[0].Fixes[0].Range.End()
		}
		fmt.Fprintf(out, "range %d %d %s %s\t%s\t%s\t%d %d\n", start, end, d.Message.Id, repair, written(replacement), written(suggestion), editStart, editEnd)
	}
	if fields[6] == "recovery" {
		fmt.Fprintln(out, "recovery findings only")
		return len(diagnostics)
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
	if len(file.Diagnostics()) != 0 && fields[6] != "recovery" {
		panic(fmt.Sprintf("invalid corpus %s: %v; source=%q", path, file.Diagnostics(), source))
	}
	selected := []rule.Rule{rules.NoDebugger, rules.NoEmpty, rules.Eqeqeq, rules.NoVar, rules.NoDuplicateCase, rules.NoContinue, rules.NoWith, rules.NoNew, rules.NoSparseArrays, rules.RequireYield, rules.NoAwaitInLoop, rules.VarsOnTop, rules.NoTemplateCurlyInString, rules.NoDivRegex, rules.NoBitwise, rules.NoLabels, rules.NoSequences, rules.UnicodeBom, rules.NoUnneededTernary, rules.NoWarningComments, rules.NoPlusplus, base.ConsistencyNoConsole, nexus.ConsistencyRequireTypeSuffix, adamic.NoTypePredicate, typescript.MethodSignatureStyle, typescript.NoWrapperObjectTypes, typescript.PreferLiteralEnumMember, nexus.ConsistencyNoEnum, rules.NoNegatedCondition, rules.NoReturnAssign}
	var diagnostics []rule.Diagnostic
	var listeners []rule.Listeners
	for _, subject := range selected {
		if fields[1] != "" && fields[1] != "all" && fields[1] != subject.Name {
			continue
		}
		ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; diagnostics = append(diagnostics, d) }}
		var options any
		if fields[5] != "" {
			switch subject.Name {
			case "@typescript-eslint/method-signature-style":
				var decoded typescript.MethodSignatureStyleOptions
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
				options = decoded
			case "@typescript-eslint/prefer-literal-enum-member":
				var decoded typescript.PreferLiteralEnumMemberOptions
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
				options = decoded
			case "no-return-assign":
				if fields[5][0] != '"' {
					break
				}
				var decoded rules.NoReturnAssignOptions
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
				options = decoded
			}
		}
		if subject.Name == "no-plusplus" && fields[5] != "" {
			var decoded rules.NoPlusplusOptions
			if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
				panic(err)
			}
			options = decoded
		}
		if subject.Name == "eqeqeq" {
			options = rules.EqeqeqOptions{Mode: rules.EqeqeqMode(fields[2]), Null: rules.EqeqeqNullPolicy(fields[3])}
		}
		if subject.Name == "no-bitwise" {
			var decoded rules.NoBitwiseOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
		}
		if subject.Name == "no-labels" {
			var decoded rules.NoLabelsOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
		}
		if subject.Name == "no-sequences" {
			var decoded rules.NoSequencesOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
		}
		if subject.Name == "unicode-bom" {
			var decoded rules.UnicodeBomOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
		}
		if subject.Name == "no-unneeded-ternary" {
			var decoded rules.NoUnneededTernaryOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
		}
		if subject.Name == "no-warning-comments" {
			var decoded rules.NoWarningCommentsOptions
			if fields[5] != "" {
				if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
					panic(err)
				}
			}
			options = decoded
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
