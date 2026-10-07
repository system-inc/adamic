// Built only through an overlay inside cohere. Rule bodies remain unmodified.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
	typescriptRules "github.com/system-inc/cohere/internal/lint/rules/typescript"
	"os"
	"sort"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func written(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != '\\' {
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
	astMode := os.Args[1] == "--ast"
	countMode := os.Args[1] == "--count"
	path := os.Args[1]
	if astMode || countMode {
		path = os.Args[2]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var rows []struct {
		File, Source, Rule string
		Options            json.RawMessage
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	var projected []map[string]any
	count := 0
	for _, row := range rows {
		if !strings.HasPrefix(row.File, "/") {
			row.File = "/repository/source/" + row.File
		}
		kind := core.ScriptKindTS
		if strings.HasSuffix(row.File, ".tsx") {
			kind = core.ScriptKindTSX
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: row.File, Path: tspath.Path(row.File)}, row.Source, kind)

		if astMode {
			offsets := map[int]int{0: 0}
			unit := 0
			for position, point := range row.Source {
				offsets[position] = unit
				if point > 65535 {
					unit += 2
				} else {
					unit++
				}
				offsets[position+utf8.RuneLen(point)] = unit
			}
			var nodes []map[string]any
			var visit func(*ast.Node) int
			visit = func(node *ast.Node) int {
				children := []int{}
				node.ForEachChild(func(child *ast.Node) bool { children = append(children, visit(child)); return false })
				text := ""
				switch node.Kind {
				case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
					text = node.Text()
				}
				operator := ""
				if node.Kind == ast.KindBinaryExpression {
					operator = strings.TrimPrefix(node.AsBinaryExpression().OperatorToken.Kind.String(), "Kind")
				}
				newFlags := 0
				if node.Kind == ast.KindNewExpression {
					args := node.ArgumentList()
					if args == nil {
						newFlags = 1
					} else if len(args.Nodes) == 0 {
						newFlags = 2
					}
				}
				nodes = append(nodes, map[string]any{"operator": operator, "newFlags": newFlags, "kind": strings.TrimPrefix(node.Kind.String(), "Kind"), "pos": offsets[node.Pos()], "end": offsets[node.End()], "text": text, "children": children})
				return len(nodes) - 1
			}
			root := visit(file.AsNode())
			options := string(row.Options)
			if options == "" {
				options = "null"
			}
			chunks := []string{}
			for start := 0; start < len(row.Source); {
				end := start + 4096
				if end > len(row.Source) {
					end = len(row.Source)
				}
				for end < len(row.Source) && !utf8.RuneStart(row.Source[end]) {
					end++
				}
				chunks = append(chunks, row.Source[start:end])
				start = end
			}
			projected = append(projected, map[string]any{"file": row.File, "source": "", "sourceChunks": chunks, "rule": row.Rule, "options": options, "ast": nodes, "root": root, "goDiagnostics": len(file.Diagnostics())})
			continue
		}
		subject := typescriptRules.ConsistentTypeAssertions
		var options any
		if len(row.Options) > 0 && string(row.Options) != "null" {
			options, err = typescriptRules.DecodeConsistentTypeAssertionsOptions(row.Options)
			if err != nil {
				panic(err)
			}
		}
		var findings []rule.Diagnostic
		listeners := subject.Run(rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; findings = append(findings, d) }}, options)
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if f := listeners[node.Kind]; f != nil {
				f(node)
			}
			node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		walk(file.AsNode())
		sort.SliceStable(findings, func(i, j int) bool { return findings[i].Range.Pos() < findings[j].Range.Pos() })
		count += len(findings)
		if countMode {
			continue
		}
		fmt.Printf("case %s %s\n", written(row.File), row.Rule)
		for _, d := range findings {
			repair := ""
			replacement := ""
			if len(d.Fixes) > 0 {
				repair = "fix"
				replacement = d.Fixes[0].Text
			}
			fmt.Printf("%d %d %s\t%s\t%s\t%s\t\n", d.Range.Pos(), d.Range.End(), d.Message.Id, written(d.Message.Description), repair, written(replacement))
			for _, fix := range d.Fixes {
				fmt.Printf("fix %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
			}
			for _, suggestion := range d.Suggestions {
				fmt.Printf("suggestion %s\t%s\n", suggestion.Message.Id, written(suggestion.Message.Description))
				for _, fix := range suggestion.Fixes {
					fmt.Printf("suggestion-edit %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), written(fix.Text))
				}
				result := row.Source
				fixes := append([]rule.Fix(nil), suggestion.Fixes...)
				sort.SliceStable(fixes, func(i, j int) bool { return fixes[i].Range.Pos() > fixes[j].Range.Pos() })
				for _, fix := range fixes {
					result = result[:fix.Range.Pos()] + fix.Text + result[fix.Range.End():]
				}
				fmt.Printf("suggestion-fixed\t%s\n", written(result))
			}
		}
		result, err := edit.FixText(row.File, row.Source, func(name, text string) ([]edit.Proposal, error) {
			parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: name, Path: tspath.Path(name)}, text, kind)
			var found []rule.Diagnostic
			listeners := subject.Run(rule.Context{SourceFile: parsed, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { d.RuleName = subject.Name; found = append(found, d) }}, options)
			var visit func(*ast.Node)
			visit = func(node *ast.Node) {
				if f := listeners[node.Kind]; f != nil {
					f(node)
				}
				node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
			}
			visit(parsed.AsNode())
			return edit.ProposalsFrom(found), nil
		}, 10)
		if err != nil || !result.Converged {
			panic(fmt.Sprintf("fix failed %v %+v", err, result))
		}
		for _, rejected := range result.Rejected {
			fmt.Printf("rejected %s %d %d %s %s\n", rejected.Proposal.RuleName, rejected.Proposal.Fix.Range.Pos(), rejected.Proposal.Fix.Range.End(), rejected.ConflictsWith, rejected.Reason)
		}
		fmt.Printf("fixed\t%s\n", written(result.Text))
	}
	if astMode {
		if err = json.NewEncoder(os.Stdout).Encode(projected); err != nil {
			panic(err)
		}
	} else if countMode {
		fmt.Println(count)
	}
}
