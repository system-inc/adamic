package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp/syntax"
	"strconv"
)

func main() {
	file, err := parser.ParseFile(token.NewFileSet(), "cohere/internal/format/doc/string_width_generated.go", nil, 0)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("stage1/cohere/json/widthTables.ts")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "// Generated from cohere/internal/format/doc/string_width_generated.go at 715ba94.\n// RE2 programs retain the upstream emoji pattern's branch order and UTF-16 mapping.")
	for _, decl := range file.Decls {
		g, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range g.Specs {
			v, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			name := v.Names[0].Name
			if name == "wideRanges" {
				fmt.Fprintln(f, "export const wideRanges: readonly (readonly number[])[] = [")
				ast.Inspect(v.Values[0], func(n ast.Node) bool {
					if c, ok := n.(*ast.CompositeLit); ok && c.Type == nil {
						fmt.Fprintf(f, "[%s, %s],\n", c.Elts[0].(*ast.BasicLit).Value, c.Elts[1].(*ast.BasicLit).Value)
						return false
					}
					return true
				})
				fmt.Fprintln(f, "];")
				continue
			}
			if name != "emojiPatternSource" && name != "narrowEmojiPatternSource" {
				continue
			}
			source, err := strconv.Unquote(v.Values[0].(*ast.BasicLit).Value)
			if err != nil {
				panic(err)
			}
			re, err := syntax.Parse(source, syntax.Perl)
			if err != nil {
				panic(err)
			}
			prog, err := syntax.Compile(re.Simplify())
			if err != nil {
				panic(err)
			}
			out := "emoji"
			if name == "narrowEmojiPatternSource" {
				out = "narrow"
			}
			fmt.Fprintf(f, "export const %sStart = %d;\nexport const %s: readonly (readonly number[])[] = [\n", out, prog.Start, out)
			for _, in := range prog.Inst {
				fmt.Fprintf(f, "[%d, %d, %d", in.Op, in.Out, in.Arg)
				for _, r := range in.Rune {
					fmt.Fprintf(f, ", %d", r)
				}
				fmt.Fprintln(f, "],")
			}
			fmt.Fprintln(f, "];")
			fmt.Fprintf(os.Stderr, "%s: %d instructions\n", out, len(prog.Inst))
		}
	}
}
