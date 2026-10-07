// Built only through an overlay inside cohere. Rule bodies remain unmodified.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	coreRules "github.com/system-inc/cohere/internal/lint/rules/core"
	nextRules "github.com/system-inc/cohere/internal/lint/rules/next"
	tailwindRules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
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
        if !strings.HasPrefix(row.File, "/") { row.File = "/repository/source/" + row.File }
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
				nodes = append(nodes, map[string]any{"kind": strings.TrimPrefix(node.Kind.String(), "Kind"), "pos": offsets[node.Pos()], "end": offsets[node.End()], "text": text, "children": children})
				return len(nodes) - 1
			}
			root := visit(file.AsNode())
			options := string(row.Options)
			if options == "" {
				options = "null"
			}
			projected = append(projected, map[string]any{"file": row.File, "source": row.Source, "rule": row.Rule, "options": options, "ast": nodes, "root": root, "goDiagnostics": len(file.Diagnostics())})
			continue
		}
		var subject rule.Rule
		var options any
		switch row.Rule {
		case "structure/tailwind-no-physical-direction":
			subject = tailwindRules.NoPhysicalDirection
		case "@next/next/google-font-display":
			subject = nextRules.GoogleFontDisplay
		case "@eslint-community/eslint-comments/require-description":
			subject = coreRules.RequireDescription
			if len(row.Options) > 0 && string(row.Options) != "null" {
				options, err = coreRules.DecodeRequireDescriptionOptions(row.Options)
				if err != nil {
					panic(err)
				}
			}
		default:
			panic("unknown rule")
		}
		var findings []rule.Diagnostic
		listeners := subject.Run(rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { findings = append(findings, d) }}, options)
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
			if len(d.Fixes) > 0 || len(d.Suggestions) > 0 {
				panic("unexpected repair")
			}
			fmt.Printf("%d %d %s\t%s\t\t\t\n", d.Range.Pos(), d.Range.End(), d.Message.Id, written(d.Message.Description))
		}
		fmt.Printf("fixed\t%s\n", written(row.Source))
	}
	if astMode {
		if err = json.NewEncoder(os.Stdout).Encode(projected); err != nil {
			panic(err)
		}
	} else if countMode {
		fmt.Println(count)
	}
}
