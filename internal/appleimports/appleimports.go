// Package appleimports says whether a program's import graph reaches one of Apple's modules (apple/...), parsed from
// its source. It is the first of the two keys a darwin program is known by (Language's rules 5 and 6 on #an8zcp1); the
// second is the compiler's typed platform stop, platform.MissingExitCode, and neither a diagnostic's words nor a path
// grants it. Two readers hold programs to it: the admission delta at a revision, through git, and internal/acheck in the
// tree it checks. Each says how a file is found and read; the parse and the walk are this package's alone.
package appleimports

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// Graph is a tree of source files, each read once and its import specifiers kept.
type Graph struct {
	exists  func(file string) bool
	read    func(file string) (string, error)
	imports map[string][]string
}

// New is a graph over the files exists says are there, each read by read. Files are slash-separated and relative to
// the tree's root.
func New(exists func(file string) bool, read func(file string) (string, error)) *Graph {
	return &Graph{exists: exists, read: read, imports: map[string][]string{}}
}

// Specifiers is every module specifier file names in an import, an export, an import-equals, an import() call or an
// import type, as written.
func (g *Graph) Specifiers(file string) ([]string, error) {
	if imports, ok := g.imports[file]; ok {
		return imports, nil
	}
	source, err := g.read(file)
	if err != nil {
		return nil, err
	}
	name := "/" + file
	if strings.HasSuffix(name, ".a") {
		name += ".ts"
	}
	kind := core.ScriptKindTS
	if strings.HasSuffix(file, ".tsx") {
		kind = core.ScriptKindTSX
	}
	if strings.HasSuffix(file, ".js") || strings.HasSuffix(file, ".mjs") || strings.HasSuffix(file, ".cjs") {
		kind = core.ScriptKindJS
	}
	if strings.HasSuffix(file, ".jsx") {
		kind = core.ScriptKindJSX
	}
	parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(name), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(name))}, source, kind)
	// The parser recovers through syntax errors. They are not a third key: a
	// typed platform stop leaves the whole Apple row for the darwin witnesses.
	var imports []string
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		var specifier *ast.Node
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindExportDeclaration:
			specifier = node.ModuleSpecifier()
		case ast.KindImportEqualsDeclaration:
			reference := node.AsImportEqualsDeclaration().ModuleReference
			if reference.Kind == ast.KindExternalModuleReference {
				specifier = reference.AsExternalModuleReference().Expression
			}
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			if call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				specifier = call.Arguments.Nodes[0]
			}
		case ast.KindImportType:
			argument := node.AsImportTypeNode().Argument
			if argument.Kind == ast.KindLiteralType {
				specifier = argument.AsLiteralTypeNode().Literal
			}
		}
		if specifier != nil && specifier.Kind == ast.KindStringLiteral {
			imports = append(imports, specifier.Text())
		}
		node.ForEachChild(visit)
		return false
	}
	visit(parsed.AsNode())
	g.imports[file] = imports
	return imports, nil
}

// local is the file a relative specifier in from names, "" when none in the tree is it or it leaves the tree.
func (g *Graph) local(from, specifier string) string {
	file := path.Clean(path.Join(path.Dir(from), specifier))
	if file == ".." || strings.HasPrefix(file, "../") {
		return ""
	}
	// Match the loader's source spellings: X.a is visible as X.a.ts, never
	// guessed from an extensionless X. A real X.a.ts wins over that alias.
	candidates := []string{file}
	switch {
	case strings.HasSuffix(file, ".a"):
		candidates = []string{strings.TrimSuffix(file, ".a") + ".d.a.ts", file + ".ts", file}
	case strings.HasSuffix(file, ".a.ts"):
		candidates = []string{file, strings.TrimSuffix(file, ".ts")}
	case strings.HasSuffix(file, ".jsx"):
		stem := strings.TrimSuffix(file, ".jsx")
		candidates = []string{stem + ".tsx", stem + ".ts", stem + ".d.ts", file, stem + ".js"}
	case strings.HasSuffix(file, ".js"):
		stem := strings.TrimSuffix(file, ".js")
		candidates = []string{stem + ".ts", stem + ".tsx", stem + ".d.ts", file, stem + ".jsx"}
	case strings.HasSuffix(file, ".mjs"):
		stem := strings.TrimSuffix(file, ".mjs")
		candidates = []string{stem + ".mts", stem + ".d.mts", file}
	case strings.HasSuffix(file, ".cjs"):
		stem := strings.TrimSuffix(file, ".cjs")
		candidates = []string{stem + ".cts", stem + ".d.cts", file}
	case path.Ext(file) == "":
		candidates = []string{file + ".ts", file + ".tsx", file + ".d.ts", file + ".js", file + ".jsx", file + "/index.ts", file + "/index.tsx", file + "/index.d.ts", file + "/index.js", file + "/index.jsx"}
	}
	for _, candidate := range candidates {
		if g.exists(candidate) {
			return candidate
		}
	}
	return ""
}

// Reaches says whether file, or a file its relative imports lead to in the tree, imports an apple/ module.
func (g *Graph) Reaches(file string) (bool, error) {
	seen := map[string]bool{}
	var walk func(string) (bool, error)
	walk = func(file string) (bool, error) {
		if seen[file] || !g.exists(file) {
			return false, nil
		}
		seen[file] = true
		imports, err := g.Specifiers(file)
		if err != nil {
			return false, err
		}
		for _, specifier := range imports {
			if strings.HasPrefix(specifier, "apple/") {
				return true, nil
			}
			if strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
				if local := g.local(file, specifier); local != "" {
					found, err := walk(local)
					if found || err != nil {
						return found, err
					}
				}
			}
		}
		return false, nil
	}
	return walk(file)
}
