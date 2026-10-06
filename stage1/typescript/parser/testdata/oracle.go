// Built inside typescript-go through a Go overlay; uses its unmodified parser.
package main

import (
	"bufio"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"os"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

func written(text string) string {
	var out strings.Builder
	// Scanner values can contain CESU-8 lone surrogates. Preserve their code units.
	for i := 0; i < len(text); {
		var r rune
		var size int
		if i+2 < len(text) && text[i] == 0xed && text[i+1] >= 0xa0 && text[i+1] <= 0xbf && text[i+2]&0xc0 == 0x80 {
			r = rune(text[i]&15)<<12 | rune(text[i+1]&63)<<6 | rune(text[i+2]&63)
			size = 3
		} else {
			r, size = utf8.DecodeRuneInString(text[i:])
		}
		i += size
		if r >= 32 && r <= 126 && r != '\\' {
			out.WriteRune(r)
		} else if r <= 0xffff {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			high, low := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, high, low)
		}
	}
	return out.String()
}

func kind(k ast.Kind) string { return strings.TrimPrefix(k.String(), "Kind") }
func walk(out *bufio.Writer, n *ast.Node, depth int, countOnly bool) int {
	text, operator, raw := "", "", ""
	flags, list, trailing, multiline := ast.TokenFlags(0), -1, false, false
	switch n.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier:
		text = n.Text()
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
		text = n.Text()
		flags = n.LiteralLikeData().TokenFlags
	case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail:
		text = n.Text()
		flags = n.TemplateLiteralLikeData().TemplateFlags
		if n.Kind != ast.KindNoSubstitutionTemplateLiteral {
			raw = n.RawText()
		}
	case ast.KindPrefixUnaryExpression:
		operator = kind(n.AsPrefixUnaryExpression().Operator)
	case ast.KindPostfixUnaryExpression:
		operator = kind(n.AsPostfixUnaryExpression().Operator)
	case ast.KindMetaProperty:
		operator = kind(n.AsMetaProperty().KeywordToken)
	case ast.KindArrayLiteralExpression:
		a := n.AsArrayLiteralExpression()
		list = len(a.Elements.Nodes)
		trailing = a.Elements.HasTrailingComma()
		multiline = a.MultiLine
	case ast.KindObjectLiteralExpression:
		a := n.AsObjectLiteralExpression()
		list = len(a.Properties.Nodes)
		trailing = a.Properties.HasTrailingComma()
		multiline = a.MultiLine
	case ast.KindCallExpression, ast.KindNewExpression:
		if a := n.ArgumentList(); a != nil {
			list = len(a.Nodes)
			trailing = a.HasTrailingComma()
		}
	}
	if !countOnly {
		fmt.Fprintf(out, "%d %s %d %d %d %d %d %d %d\t%s\t%s\t%s\n", depth, kind(n.Kind), n.Pos(), n.End(), n.Flags&ast.NodeFlagsOptionalChain, flags, list, core.IfElse(trailing, 1, 0), core.IfElse(multiline, 1, 0), operator, written(text), written(raw))
	}
	count := 1
	n.ForEachChild(func(child *ast.Node) bool { count += walk(out, child, depth+1, countOnly); return false })
	return count
}
func run(out *bufio.Writer, path string, countOnly bool) int {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/source.ts"}, string(text), core.ScriptKindTS)
	if len(f.Diagnostics()) != 0 {
		for _, d := range f.Diagnostics() {
			fmt.Fprintf(os.Stderr, "parser diagnostic %d %d %d\n", d.Code(), d.Pos(), d.Len())
		}
		os.Exit(1)
	}
	count := 0
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Parent != nil && ast.IsExpressionNode(n) {
			if !countOnly {
				fmt.Fprintln(out, "expression")
			}
			count += walk(out, n, 0, countOnly)
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(f.AsNode())
	return count
}
func main() {
	out := bufio.NewWriterSize(os.Stdout, 65536)
	defer out.Flush()
	args := os.Args[1:]
	if args[0] != "--manifest" {
		run(out, args[0], false)
		return
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	countOnly := len(args) > 2 && args[2] == "--count"
	count, index := 0, 0
	for _, path := range strings.Split(string(data), "\n") {
		if path == "" {
			continue
		}
		if !countOnly {
			fmt.Fprintf(out, "case %d\n", index)
		}
		count += run(out, path, countOnly)
		index++
	}
	if countOnly {
		fmt.Fprintln(out, count)
	}
}
