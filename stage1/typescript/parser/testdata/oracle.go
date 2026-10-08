// Built inside typescript-go through a Go overlay; uses its unmodified parser.
package main

import (
	"bufio"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
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
func semantic(n *ast.Node) string {
	switch n.Kind {
	case ast.KindVariableDeclarationList:
		return fmt.Sprint(n.Flags & ast.NodeFlagsBlockScoped)
	case ast.KindImportClause:
		return kind(n.AsImportClause().PhaseModifier)
	case ast.KindImportSpecifier:
		return fmt.Sprint(core.IfElse(n.AsImportSpecifier().IsTypeOnly, 1, 0))
	case ast.KindExportSpecifier:
		return fmt.Sprint(core.IfElse(n.AsExportSpecifier().IsTypeOnly, 1, 0))
	case ast.KindExportDeclaration:
		return fmt.Sprint(core.IfElse(n.AsExportDeclaration().IsTypeOnly, 1, 0))
	case ast.KindImportEqualsDeclaration:
		return fmt.Sprint(core.IfElse(n.AsImportEqualsDeclaration().IsTypeOnly, 1, 0))
	case ast.KindExportAssignment:
		return fmt.Sprint(core.IfElse(n.AsExportAssignment().IsExportEquals, 1, 0))
	case ast.KindJsxText:
		return fmt.Sprint(core.IfElse(n.AsJsxText().ContainsOnlyTriviaWhiteSpaces, 1, 0))
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		if arguments := n.TypeArgumentList(); arguments != nil {
			return fmt.Sprintf("%d:%d", len(arguments.Nodes), core.IfElse(arguments.HasTrailingComma(), 1, 0))
		}
		return "-1:0"
	}
	return ""
}
func walk(out *bufio.Writer, n *ast.Node, depth int, countOnly bool, source string, whole bool) int {
	text, operator, raw := "", "", ""
	flags, list, trailing, multiline := ast.TokenFlags(0), -1, false, false
	switch n.Kind {
	case ast.KindJsxText:
		text = n.AsJsxText().Text
	case ast.KindJsxAttributes:
		list = len(n.AsJsxAttributes().Properties.Nodes)
	case ast.KindJsxElement, ast.KindJsxFragment:
		list = len(n.Children().Nodes)
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
		} else {
			raw = source[scanner.SkipTrivia(source, n.Pos())+1 : n.End()-1]
		}
	case ast.KindPrefixUnaryExpression:
		operator = kind(n.AsPrefixUnaryExpression().Operator)
	case ast.KindPostfixUnaryExpression:
		operator = kind(n.AsPostfixUnaryExpression().Operator)
	case ast.KindImportType:
		if n.AsImportTypeNode().IsTypeOf {
			operator = "TypeOfKeyword"
		}
	case ast.KindModuleDeclaration:
		operator = kind(n.AsModuleDeclaration().Keyword)
	case ast.KindImportAttributes:
		a := n.AsImportAttributes()
		operator = kind(a.Token)
		list = len(a.Attributes.Nodes)
		trailing = a.Attributes.HasTrailingComma()
		multiline = a.MultiLine
	case ast.KindHeritageClause:
		operator = kind(n.AsHeritageClause().Token)
	case ast.KindTypeOperator:
		operator = kind(n.AsTypeOperatorNode().Operator)
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
		fmt.Fprintf(out, "%d %s %d %d %d %d %d %d %d\t%s\t%s\t%s", depth, kind(n.Kind), n.Pos(), n.End(), n.Flags&ast.NodeFlagsOptionalChain, flags, list, core.IfElse(trailing, 1, 0), core.IfElse(multiline, 1, 0), operator, written(text), written(raw))
		if whole {
			fmt.Fprint(out, "\t", semantic(n))
		}
		fmt.Fprintln(out)
	}
	count := 1
	n.ForEachChild(func(child *ast.Node) bool { count += walk(out, child, depth+1, countOnly, source, whole); return false })
	return count
}

var recovery bool
var docTypes bool
var obsoleteAssertions bool
var jsxRecovery bool

func run(out *bufio.Writer, path string, countOnly bool, whole bool) int {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	script := core.ScriptKindTS
	if strings.HasSuffix(path, ".tsx") {
		script = core.ScriptKindTSX
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".mjs") || strings.HasSuffix(path, ".cjs") {
		script = core.ScriptKindJS
	}
	if strings.HasSuffix(path, ".jsx") {
		script = core.ScriptKindJSX
	}
	f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(path)}, string(text), script)
	obsoleteOnly := obsoleteAssertions && len(f.Diagnostics()) > 0
	for _, d := range f.Diagnostics() {
		if d.Code() != 2880 {
			obsoleteOnly = false
		}
	}
	if obsoleteAssertions && !obsoleteOnly {
		panic("obsolete assertion probe must emit only diagnostic 2880")
	}
	if len(f.Diagnostics()) != 0 && !obsoleteOnly && !recovery && !(jsxRecovery && (script == core.ScriptKindTSX || script == core.ScriptKindJSX)) {
		fmt.Fprintf(os.Stderr, "source: %q\n", text)
		for _, d := range f.Diagnostics() {
			fmt.Fprintf(os.Stderr, "parser diagnostic %s %d %d %d\n", path, d.Code(), d.Pos(), d.Len())
		}
		os.Exit(1)
	}
	if recovery {
		for _, d := range f.Diagnostics() {
			fmt.Fprintf(out, "diagnostic %d %d %d %d\t%s\n", d.Code(), d.Pos(), d.Len(), d.Category(), written(d.String()))
		}
	}
	if docTypes {
		count := 0
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindJSDocTypeExpression {
				if !countOnly {
					fmt.Fprintln(out, "type")
				}
				count += walk(out, n.AsJSDocTypeExpression().Type, 0, countOnly, f.Text(), true)
				return false
			}
			ast.ForEachChildAndJSDoc(n, f, visit)
			return false
		}
		visit(f.AsNode())
		return count
	}
	if whole {
		if !countOnly {
			fmt.Fprintln(out, "file")
		}
		return walk(out, f.AsNode(), 0, countOnly, f.Text(), true)
	}
	count := 0
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Parent != nil && ast.IsExpressionNode(n) {
			if !countOnly {
				fmt.Fprintln(out, "expression")
			}
			count += walk(out, n, 0, countOnly, f.Text(), whole)
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(f.AsNode())
	return count
}

// Use the real parser's literal locations to direct regex/template rescanning.
// A lexical guess about a slash or closing brace could count different tokens.
func tokenSpans(out *bufio.Writer, path string) {
	text, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	source := string(text)
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/source.ts"}, source, core.ScriptKindTS)
	if len(file.Diagnostics()) != 0 {
		panic("token corpus must be well formed")
	}
	literals := map[int]ast.Kind{}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindRegularExpressionLiteral, ast.KindTemplateMiddle, ast.KindTemplateTail:
			literals[scanner.SkipTrivia(source, n.Pos())] = n.Kind
		}
		n.ForEachChild(visit)
		return false
	}
	visit(file.AsNode())
	s := scanner.NewScanner()
	s.SetText(source)
	for {
		k := s.Scan()
		switch literals[s.TokenStart()] {
		case ast.KindRegularExpressionLiteral:
			k = s.ReScanSlashToken(false)
		case ast.KindTemplateMiddle, ast.KindTemplateTail:
			k = s.ReScanTemplateToken(false)
		}
		if k == ast.KindEndOfFile {
			break
		}
		fmt.Fprintf(out, "%d %d\n", s.TokenStart(), s.TokenEnd())
	}
}

func main() {
	out := bufio.NewWriterSize(os.Stdout, 65536)
	defer out.Flush()
	args := os.Args[1:]
	if args[0] == "--jsx-kinds" {
		fmt.Fprintln(out, kind(ast.KindJsxText))
		for k := ast.KindJsxElement; k <= ast.KindJsxNamespacedName; k++ {
			fmt.Fprintln(out, kind(k))
		}
		return
	}
	if args[0] == "--token-spans" {
		tokenSpans(out, args[1])
		return
	}
	if args[0] == "--type-kinds" {
		for k := ast.KindUnknown; k <= ast.KindLastJSDocNode; k++ {
			if ast.IsTypeNodeKind(k) {
				fmt.Fprintln(out, kind(k))
			}
		}
		return
	}
	whole, countOnly := false, false
	for _, arg := range args {
		if arg == "--jsx-recovery" {
			jsxRecovery = true
		}
		if arg == "--allow-obsolete-assert" {
			obsoleteAssertions = true
		}
		if arg == "--doc-types" {
			docTypes = true
		}
		if arg == "--recovery" {
			recovery = true
		}
		if arg == "--whole" {
			whole = true
		}
		if arg == "--count" {
			countOnly = true
		}
	}
	if args[0] != "--manifest" {
		run(out, args[0], false, whole)
		return
	}
	data, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	count, index := 0, 0
	for _, path := range strings.Split(string(data), "\n") {
		if path == "" {
			continue
		}
		if !countOnly {
			fmt.Fprintf(out, "case %d\n", index)
		}
		count += run(out, path, countOnly, whole)
		index++
	}
	if countOnly {
		fmt.Fprintln(out, count)
	}
}
