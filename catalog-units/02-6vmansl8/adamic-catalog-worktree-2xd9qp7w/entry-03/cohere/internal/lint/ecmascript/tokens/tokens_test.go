package tokens

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// An import declaration's tokens, with its leading comment and the next statement left out, and the
// three questions asked of a specifier in the middle of its list.
func TestOfAnImportDeclaration(t *testing.T) {
	t.Parallel()

	text := "/* lead */ import { a, /* gap */ b } from 'm';\nconst c = 1;\n"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/file.ts", PathKey: "/file.ts"}, text, core.ScriptKindTS)
	statement := sourceFile.Statements.Nodes[0]
	list := Of(sourceFile, statement)

	var kinds []ast.Kind
	for _, token := range list {
		kinds = append(kinds, token.Kind)
	}
	want := []ast.Kind{ast.KindImportKeyword, ast.KindOpenBraceToken, ast.KindIdentifier, ast.KindCommaToken,
		ast.KindIdentifier, ast.KindCloseBraceToken, ast.KindFromKeyword, ast.KindStringLiteral, ast.KindSemicolonToken}
	if len(kinds) != len(want) {
		t.Fatalf("got %v, want %v", kinds, want)
	}
	for index := range want {
		if kinds[index] != want[index] {
			t.Fatalf("token %d is %v, want %v", index, kinds[index], want[index])
		}
	}

	bStart := len("/* lead */ import { a, /* gap */ ")
	if comma, found := list.Before(bStart); !found || comma.Kind != ast.KindCommaToken || text[comma.Start:comma.End] != "," {
		t.Errorf("before b: %+v, %v, want the comma", comma, found)
	}
	if brace, found := list.After(bStart + 1); !found || brace.Kind != ast.KindCloseBraceToken {
		t.Errorf("after b: %+v, %v, want the closing brace", brace, found)
	}
	if _, found := list.FirstBetween(0, bStart, ast.KindCloseBraceToken); found {
		t.Error("found a closing brace before b, where there is none")
	}
	if brace, found := list.FirstBetween(0, len(text), ast.KindCloseBraceToken); !found || text[brace.Start:brace.End] != "}" {
		t.Errorf("first closing brace: %+v, %v", brace, found)
	}
}
