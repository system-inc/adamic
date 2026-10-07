package metamorphic

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// wrap moves the module's statements into one function, declared after what stays at the top level
// and called once at the end. What stays: imports, types, classes and enums, generic and overloaded
// functions, and then whatever any of them reaches by a value (a function a class calls, a constant a
// method reads), found to a fixed point, since a class, a generic function expression and a function
// declaration inside a function are what stage 0 doesn't lower yet. Every other function moves in as
// a const holding a function expression of the same name and text, at the top of the body in the
// order they were written: hoisted as the declarations were, so a call that came before its function
// still finds it, and a closure over the module's variables, now locals. A module that exports is
// left alone: what it exports would stop being the module's.
//
// Stage 0 lowers some closures as arrows that it doesn't as function expressions, and some not at
// all, so the functions can move in three ways, tried in that order: as function expressions, as
// arrows, or not at all (they stay at the top level, with every variable they reach).
type functionsAs int

const (
	asFunctionExpressions functionsAs = iota
	asArrows
	atTopLevel
)

func wrap(p *program, functions functionsAs) (string, string) {
	statements := p.file.Statements.Nodes
	for _, statement := range statements {
		switch statement.Kind {
		case ast.KindExportDeclaration, ast.KindExportAssignment:
			return "", "it exports"
		}
		if ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
			return "", "it exports"
		}
	}
	index := map[*ast.Node]int{}
	stays := make([]bool, len(statements))
	for position, statement := range statements {
		index[statement] = position
		switch statement.Kind {
		case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
			ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
			stays[position] = true
		}
		if ast.HasSyntacticModifier(statement, ast.ModifierFlagsAmbient) {
			stays[position] = true
		}
		if statement.Kind == ast.KindFunctionDeclaration && (statement.Body() == nil || statement.TypeParameterList() != nil || functions == atTopLevel) {
			stays[position] = true
		}
	}
	// A function with overloads is several statements of one name: all of them stay.
	for position, statement := range statements {
		if statement.Kind == ast.KindFunctionDeclaration && statement.Body() == nil && statement.Name() != nil {
			for other, sibling := range statements {
				if sibling.Kind == ast.KindFunctionDeclaration && sibling.Name() != nil && sibling.Name().Text() == statement.Name().Text() {
					stays[other] = true
				}
			}
			stays[position] = true
		}
	}
	// What stays reaches by name, to a fixed point.
	for changed := true; changed; {
		changed = false
		for position, statement := range statements {
			if !stays[position] {
				continue
			}
			visit(statement, func(node *ast.Node) bool {
				if node.Kind != ast.KindIdentifier || ast.IsDeclarationName(node) {
					return true
				}
				symbol := p.checker.GetSymbolAtLocation(node)
				if symbol == nil {
					return true
				}
				for _, declaration := range symbol.Declarations {
					if !p.inFile(declaration) {
						continue
					}
					if top := topStatement(declaration); top != nil {
						if reached, found := index[top]; found && !stays[reached] {
							stays[reached] = true
							changed = true
						}
					}
				}
				return true
			})
		}
	}
	var moved int
	for position := range statements {
		if !stays[position] {
			moved++
		}
	}
	if moved == 0 {
		return "", "nothing to wrap"
	}
	name := p.fresh("metamorphicMain")
	var top, hoisted, body strings.Builder
	for position, statement := range statements {
		text := p.text[statement.Pos():statement.End()]
		switch {
		case stays[position]:
			top.WriteString(text)
		case statement.Kind == ast.KindFunctionDeclaration && statement.Name() != nil:
			hoisted.WriteString(p.text[statement.Pos():p.start(statement)])
			hoisted.WriteString("const " + statement.Name().Text() + " = " + p.functionValue(statement, functions) + ";")
		default:
			body.WriteString(text)
		}
	}
	var result strings.Builder
	result.WriteString(top.String())
	result.WriteString("\n\nfunction " + name + "(): void {\n")
	result.WriteString(hoisted.String())
	result.WriteString(body.String())
	result.WriteString("\n}\n\n" + name + "();")
	// Whatever follows the last statement: a comment at the end of the file.
	if len(statements) > 0 {
		result.WriteString(p.text[statements[len(statements)-1].End():])
	} else {
		result.WriteString("\n")
	}
	return result.String(), ""
}

// functionValue is a function declaration's text as a value: the declaration itself, which reads as
// a named function expression, or an arrow with its parameters, return type and body.
func (p *program) functionValue(declaration *ast.Node, functions functionsAs) string {
	if functions == asFunctionExpressions {
		return p.source(declaration)
	}
	parameters := declaration.ParameterList()
	text := "(" + p.text[parameters.Pos():parameters.End()] + ")"
	if returned := declaration.Type(); returned != nil {
		text += ": " + p.source(returned)
	}
	return text + " => " + p.source(declaration.Body())
}
