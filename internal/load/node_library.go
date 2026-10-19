package load

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

const NodeTypesVersion = "25.3.3"
const nodeTypesDirectory = "bundled:///node/node_modules/@types/node"
const nodeTypesRoot = "bundled:///node/node_modules/@types"

// The exact published package travels with the compiler, independent of its
// working directory. Build-time integrity tests hold every embedded byte.
func nodeTypesIndex() (tspath.RootedFilePath, error) {
	return nodeTypesIndexFromFS(embeddedNodeTypes)
}

func nodeTypesIndexFromFS(files fs.FS) (tspath.RootedFilePath, error) {
	data, err := fs.ReadFile(files, "node_types/node_modules/@types/node/package.json")
	if err != nil {
		return "", fmt.Errorf("load: this compiler is missing its bundled Node declarations; reinstall Adamic")
	}
	var pin struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &pin) != nil || pin.Name != "@types/node" || pin.Version != NodeTypesVersion {
		return "", fmt.Errorf("load: this compiler's bundled Node declarations are not @types/node %s; reinstall Adamic", NodeTypesVersion)
	}
	if _, err := fs.ReadFile(files, "node_types/node_modules/@types/node/index.d.ts"); err != nil {
		return "", fmt.Errorf("load: this compiler's bundled Node declarations are incomplete (missing index.d.ts); reinstall Adamic")
	}
	return tspath.RootedDirectoryPathFromNormalized(nodeTypesDirectory).ResolveFile("index.d.ts"), nil
}

func usesNodeModules(program *compiler.Program) bool {
	// A project may request Node globals without importing a node: module.
	// The caller package is hidden in favor of our pinned embedded declarations,
	// so an explicit type-library request needs the same bundled root setup.
	for _, name := range program.Options().Types {
		if name == "node" {
			return true
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindExportDeclaration:
			if specifier := node.ModuleSpecifier(); specifier != nil && strings.HasPrefix(specifier.Text(), "node:") {
				return true
			}
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			if ast.IsIdentifier(call.Expression) && call.Expression.Text() == "require" && call.Arguments != nil && len(call.Arguments.Nodes) == 1 {
				argument := call.Arguments.Nodes[0]
				if argument.Kind == ast.KindStringLiteral && nodeBuiltinModules()[argument.Text()] {
					return true
				}
			}
		case ast.KindImportType:
			argument := node.AsImportTypeNode().Argument
			if argument.Kind == ast.KindLiteralType {
				literal := argument.AsLiteralTypeNode().Literal
				if literal.Kind == ast.KindStringLiteral && nodeBuiltinModules()[literal.Text()] {
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, file := range program.GetSourceFiles() {
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return false
}

// Read module names from the pinned published declarations, so builtin subpaths
// and prefix-only modules follow the bundled Node version rather than a hand list.
var nodeBuiltinModules = sync.OnceValue(func() map[string]bool {
	modules := map[string]bool{}
	declaration := regexp.MustCompile(`(?m)^declare module "([^"]+)"`)
	err := fs.WalkDir(embeddedNodeTypes, "node_types/node_modules/@types/node", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(name, ".d.ts") {
			return nil
		}
		data, err := embeddedNodeTypes.ReadFile(name)
		if err != nil {
			return err
		}
		for _, match := range declaration.FindAllSubmatch(data, -1) {
			modules[string(match[1])] = true
		}
		return nil
	})
	if err != nil {
		panic(err) // The embedded filesystem is fixed at build time.
	}
	return modules
})

// Unit lowerers use declaration identity, not the spelling of an imported name.
func IsNodeLibrary(file *ast.SourceFile) bool {
	return file != nil && strings.HasPrefix(file.FileName().AsString(), nodeTypesDirectory+"/") && strings.HasSuffix(file.FileName().AsString(), ".d.ts")
}

// Match top-level declarations, not console method signatures. The standalone
// prelude has only console; the project overlay adds Console and process together.
func nodePrelude(source string, requiredOverlay ...string) (string, error) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(preludePath), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(preludePath))}, source, core.ScriptKindTS)
	if len(file.Diagnostics()) != 0 {
		return "", fmt.Errorf("load: cannot adapt malformed Node prelude")
	}
	declarations := make(map[string]*ast.Node)
	for _, statement := range file.Statements.Nodes {
		names := []string{}
		switch statement.Kind {
		case ast.KindInterfaceDeclaration:
			if statement.Name().Text() == "Console" {
				names = append(names, "Console")
			}
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
			for _, declaration := range list {
				name := declaration.Name()
				if name.Kind == ast.KindIdentifier && (name.Text() == "console" || name.Text() == "process") {
					if len(list) != 1 {
						return "", fmt.Errorf("load: Node prelude %s must have its own declaration", name.Text())
					}
					names = append(names, name.Text())
				}
			}
		}
		for _, name := range names {
			if declarations[name] != nil {
				return "", fmt.Errorf("load: duplicate %s declaration in Node prelude", name)
			}
			declarations[name] = statement
		}
	}
	required := append([]string{"console"}, requiredOverlay...)
	if declarations["Console"] != nil || declarations["process"] != nil {
		required = append(required, "Console", "process")
	}
	for _, name := range required {
		if declarations[name] == nil {
			return "", fmt.Errorf("load: Node prelude is missing the %s declaration", name)
		}
	}
	replacements := map[string]string{"console": "declare var console: Console;", "Console": "", "process": "declare var process: NodeJS.Process;"}
	// Work backwards so the original parser spans remain valid. Keep leading
	// comments and whitespace: Pos includes trivia before the declaration.
	for index := len(file.Statements.Nodes) - 1; index >= 0; index-- {
		statement := file.Statements.Nodes[index]
		for name, declaration := range declarations {
			if declaration == statement {
				start := scanner.GetTokenPosOfNode(statement, file, false)
				source = source[:start] + replacements[name] + source[statement.End():]
			}
		}
	}
	return source, nil
}
