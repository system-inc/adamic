package estree

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// ParseTypeScript is Prettier's typescript parser, src/language-js/parse/typescript.js: parse, convert,
// postprocess. It returns the Program and the file's comments, ready for comment attachment, and the text
// stripped of those comments, for the printer to share with the postprocess.
//
// Prettier is given a file path, so JSX is decided by the extension, the way typescript-estree decides
// it for a known file type: .tsx parses with JSX and .ts without. A hashbang becomes a line comment,
// upstream's replaceHashbang, before parsing; the caller keeps the original text for printing.
func ParseTypeScript(fileName string, text string, nodes *Arena) (*Node, []*Node, *StrippedText, error) {
	return ParseTypeScriptFrom(fileName, text, nil, nodes)
}

// ParseTypeScriptFrom is ParseTypeScript given a tree someone already parsed, so the file is not parsed a
// second time: the fix engine's guard parses every TypeScript file it is handed, and the formatter parsed
// the same bytes again, 0.18 GB a cold ahra run (#dk2502g).
//
// The tree is taken only where ParseSourceFile would have built the same one: the same bytes, no hashbang
// (which ParseTypeScript rewrites before parsing), the script kind this file's extension picks, and a name
// that is a declaration file exactly when this one is. Those are all the parser reads (parser/parser.go:
// the text, the kind, IsDeclarationFileName of the name, and the module-indicator options, which set a
// property of the file and no node), so a tree that matches them is the tree a fresh parse returns.
// Anything else is parsed here as before, so a nil tree, or the wrong one, costs a parse and never a
// different result.
func ParseTypeScriptFrom(fileName string, text string, parsed *ast.SourceFile, nodes *Arena) (*Node, []*Node, *StrippedText, error) {
	if !treeIsFor(fileName, text, parsed) {
		parsed = ParseSourceFile(fileName, ReplaceHashbang(text))
	}
	return convertTypeScript(parsed, text, nodes)
}

// treeIsFor reports whether parsed is the tree ParseTypeScript would build for text: see ParseTypeScriptFrom.
func treeIsFor(fileName string, text string, parsed *ast.SourceFile) bool {
	return parsed != nil && parsed.Text() == text && parsed.ScriptKind == scriptKindOf(fileName) && !strings.HasPrefix(text, "#!") &&
		tspath.IsDeclarationFileName(parsed.FileName().AsString()) == tspath.IsDeclarationFileName(fileName)
}

// convertTypeScript is the convert and postprocess half of ParseTypeScript, over a tree in hand.
func convertTypeScript(sourceFile *ast.SourceFile, text string, nodes *Arena) (*Node, []*Node, *StrippedText, error) {
	program, comments, err := Convert(sourceFile, nodes)
	if err != nil {
		return nil, nil, nil, err
	}
	program, comments, stripped := Postprocess(program, comments, text)
	return program, comments, stripped, nil
}

// ParseJavaScript stands in for Prettier's babel parser, which it uses for .js, .mjs, .cjs and .jsx.
//
// The file is parsed as TSX rather than as JavaScript. In JavaScript mode typescript-go reparses
// JSDoc into the tree: `@param {string}` becomes a parameter's type, `@typedef` a type alias
// statement, `@type` a cast. Babel's tree has none of that, so a converter reading typescript-go's
// would print types the source never wrote. TSX mode reparses nothing, and accepts JSX the way
// Babel does for these files.
//
// TypeScript's grammar is a superset of JavaScript's with one class of exception: text both read
// differently, chiefly `a < b > (c)`, comparisons to Babel and a generic call to TypeScript. Valid
// JavaScript never contains a TypeScript-only node, so one in the tree is either such an ambiguity
// or a file Babel would reject, and either way the file is refused rather than printed as TypeScript.
//
// The tree is typescript-estree's with Babel's one postprocess difference: see ConvertForBabel.
func ParseJavaScript(fileName string, text string, nodes *Arena) (*Node, []*Node, *StrippedText, error) {
	program, comments, err := ConvertForBabel(ParseSourceFile(fileName, ReplaceHashbang(text)), nodes)
	if err != nil {
		return nil, nil, nil, err
	}
	program, comments, stripped := Postprocess(program, comments, text)
	if found := findTypeScriptNode(program); found != nil {
		return nil, nil, nil, fmt.Errorf("%s: %s at byte %d is TypeScript syntax, which the babel parser does not read",
			fileName, found.Type(), found.Start())
	}
	return program, comments, stripped, nil
}

// findTypeScriptNode is the first node, depth first, whose type is one only TypeScript produces.
func findTypeScriptNode(node *Node) *Node {
	if strings.HasPrefix(node.Type(), "TS") {
		return node
	}
	for _, child := range ChildNodes(node) {
		if found := findTypeScriptNode(child); found != nil {
			return found
		}
	}
	return nil
}

// ParseSourceFile parses one file with typescript-go, choosing the script kind by extension.
func ParseSourceFile(fileName string, text string) *ast.SourceFile {
	rooted := fileName
	if !tspath.IsRootedDiskPath(rooted) {
		rooted = "/" + strings.TrimPrefix(rooted, "/")
	}
	rooted = tspath.NormalizePath(rooted)
	return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePath(rooted), PathKey: tspath.PathKey(rooted)}, text, scriptKindOf(fileName))
}

// scriptKindOf is the script kind ParseSourceFile parses a file as, chosen by extension.
func scriptKindOf(fileName string) core.ScriptKind {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".tsx":
		return core.ScriptKindTSX
	case ".jsx":
		return core.ScriptKindJSX
	case ".js", ".mjs", ".cjs":
		// TSX, not JS: see ParseJavaScript.
		return core.ScriptKindTSX
	}
	return core.ScriptKindTS
}

// ReplaceHashbang is upstream's replaceHashbang: `#!` becomes `//`, the same length.
func ReplaceHashbang(text string) string {
	if strings.HasPrefix(text, "#!") {
		return "//" + text[2:]
	}
	return text
}
