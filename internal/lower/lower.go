// Package lower turns a checked program into Adamic's IR.
//
// Stage 0 lowers a little and refuses the rest. Every construct it can't lower yet is an error
// naming the construct and where it is, never a skip: a compiler that drops a statement it doesn't
// understand produces a program that looks right and isn't, which is the one thing Adamic must
// never do.
package lower

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// NotYet is a construct stage 0 can't lower yet, with where it is.
type NotYet struct {
	Where string
	What  string
}

func (n *NotYet) Error() string {
	return fmt.Sprintf("%s: stage 0 can't lower %s yet", n.Where, n.What)
}

// Lower lowers a checked program, entry file first.
func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {
	files := program.Files()
	if len(files) != 1 {
		return nil, fmt.Errorf("lower: stage 0 compiles one file, got %d", len(files))
	}
	sourceFile := files[0]
	typeChecker, release := program.Checker(ctx, sourceFile)
	defer release()

	lowering := &lowering{program: program, checker: typeChecker, result: &ir.Program{}}
	// The base name only, so the same program emits the same C on every machine.
	lowering.result.Source = filepath.Base(program.FileName(sourceFile))
	for _, statement := range sourceFile.Statements.Nodes {
		if err := lowering.statement(statement); err != nil {
			return nil, err
		}
	}
	return lowering.result, nil
}

type lowering struct {
	program *load.Program
	checker *checker.Checker
	result  *ir.Program

	// strings indexes result.Strings, so a constant used twice is stored once.
	strings map[string]int
}

func (l *lowering) notYet(node *ast.Node, what string) error {
	return &NotYet{Where: l.program.Where(node), What: what}
}

func (l *lowering) statement(statement *ast.Node) error {
	if statement.Kind != ast.KindExpressionStatement {
		return l.notYet(statement, describe(statement))
	}
	expression := statement.AsExpressionStatement().Expression
	if expression.Kind != ast.KindCallExpression {
		return l.notYet(expression, describe(expression))
	}
	return l.call(expression)
}

// call lowers console.log and console.error with one string constant, the whole of stage 0's
// first slice.
func (l *lowering) call(call *ast.Node) error {
	stream, err := l.consoleStream(call.AsCallExpression().Expression)
	if err != nil {
		return err
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		// The prelude declares one parameter, so the checker has already refused any other count.
		return errors.New("lower: " + l.program.Where(call) + ": console takes one argument, and the checker let another count through")
	}
	argument := arguments[0]
	if argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return l.notYet(argument, "a console argument that isn't a string constant ("+describe(argument)+")")
	}
	l.result.Main = append(l.result.Main, ir.WriteLine{Stream: stream, String: l.constant(argument.Text())})
	return nil
}

// consoleStream is the stream a callee writes to, when it is the prelude's console.log or
// console.error. A local named console is someone else's, and isn't lowered as this one.
func (l *lowering) consoleStream(callee *ast.Node) (ir.Stream, error) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return 0, l.notYet(callee, "a call to "+describe(callee))
	}
	object := callee.AsPropertyAccessExpression().Expression
	symbol := l.checker.GetSymbolAtLocation(object)
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return 0, l.notYet(callee, "a method call")
	}
	switch callee.Name().Text() {
	case "log":
		return ir.Stdout, nil
	case "error":
		return ir.Stderr, nil
	}
	return 0, l.notYet(callee, "console."+callee.Name().Text())
}

func (l *lowering) constant(value string) int {
	if l.strings == nil {
		l.strings = map[string]int{}
	}
	if index, isKnown := l.strings[value]; isKnown {
		return index
	}
	l.strings[value] = len(l.result.Strings)
	l.result.Strings = append(l.result.Strings, value)
	return l.strings[value]
}

// describe names a node's kind for a person: "a VariableStatement", "an ImportDeclaration".
func describe(node *ast.Node) string {
	name := strings.TrimPrefix(node.Kind.String(), "Kind")
	if strings.ContainsAny(name[:1], "AEIOU") {
		return "an " + name
	}
	return "a " + name
}
