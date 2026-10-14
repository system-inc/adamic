package load

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Declaration is one named thing in a program and the type the checker proved for it.
type Declaration struct {
	File   string
	Line   int
	Column int
	Kind   string
	Name   string
	Type   string
}

// declarationKinds are the declarations stage 0 reports, by the word it reports them with.
var declarationKinds = map[ast.Kind]string{
	ast.KindVariableDeclaration:  "variable",
	ast.KindParameter:            "parameter",
	ast.KindBindingElement:       "binding",
	ast.KindFunctionDeclaration:  "function",
	ast.KindClassDeclaration:     "class",
	ast.KindPropertyDeclaration:  "property",
	ast.KindMethodDeclaration:    "method",
	ast.KindInterfaceDeclaration: "interface",
	ast.KindTypeAliasDeclaration: "type",
	ast.KindPropertySignature:    "property",
}

// typeFormat prints types whole: a truncated type is a claim cut off mid-sentence.
const typeFormat = checker.TypeFormatFlagsNoTruncation | checker.TypeFormatFlagsUseAliasDefinedOutsideCurrentScope

// Declarations is every named declaration in the program's own files, in source order, with its
// proven type.
//
// A destructured variable reports each name it binds rather than the pattern, and a type alias
// reports what it stands for rather than its own name.
func (p *Program) Declarations(ctx context.Context) []Declaration {
	declarations := []Declaration{}
	for _, sourceFile := range p.files {
		typeChecker, release := p.compiler.GetTypeCheckerForFile(ctx, sourceFile)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if kind, isReported := declarationKinds[node.Kind]; isReported {
				if name := node.Name(); name != nil && ast.IsIdentifier(name) {
					declarations = append(declarations, p.declaration(typeChecker, sourceFile, node, kind, name))
				}
			}
			node.ForEachChild(visit)
			return false
		}
		sourceFile.AsNode().ForEachChild(visit)
		release()
	}
	return declarations
}

func (p *Program) declaration(typeChecker *checker.Checker, sourceFile *ast.SourceFile, node *ast.Node, kind string, name *ast.Node) Declaration {
	flags := typeFormat
	if node.Kind == ast.KindTypeAliasDeclaration {
		flags |= checker.TypeFormatFlagsInTypeAlias
	}
	line, column := p.lineAndColumn(sourceFile, scanner.GetTokenPosOfNode(name, sourceFile, false))
	return Declaration{
		File:   p.fs.displayName(sourceFile.FileName()),
		Line:   line,
		Column: column,
		Kind:   kind,
		Name:   name.Text(),
		Type:   typeChecker.TypeToStringEx(typeChecker.GetTypeAtLocation(name), nil, flags, nil),
	}
}
