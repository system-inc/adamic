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

// Refused is something Adamic 0.1 doesn't allow at all (docs/0.1.md), as opposed to something stage 0
// hasn't learned yet. The difference is a promise, so the message keeps them apart.
type Refused struct {
	Where string
	What  string
	Fix   string
}

func (r *Refused) Error() string {
	return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", r.Where, r.What, r.Fix)
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
	main, err := lowering.statements(sourceFile.Statements.Nodes)
	if err != nil {
		return nil, err
	}
	lowering.result.Main = main
	return lowering.result, nil
}

type lowering struct {
	program *load.Program
	checker *checker.Checker
	result  *ir.Program

	// strings indexes result.Strings, so a constant used twice is stored once.
	strings map[string]int

	// locals maps each variable's symbol to its index in result.Locals.
	locals map[*ast.Symbol]int
}

func (l *lowering) notYet(node *ast.Node, what string) error {
	return &NotYet{Where: l.program.Where(node), What: what}
}

// statements lowers a list of statements.
func (l *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {
	lowered := []ir.Statement{}
	for _, node := range nodes {
		statements, err := l.statement(node)
		if err != nil {
			return nil, err
		}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// statement lowers one statement to none, one or several.
func (l *lowering) statement(node *ast.Node) ([]ir.Statement, error) {
	switch node.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEmptyStatement:
		// Types erase to nothing, and so does an empty statement.
		return nil, nil
	case ast.KindExpressionStatement:
		return l.expressionStatement(node.AsExpressionStatement().Expression)
	case ast.KindVariableStatement:
		return l.variables(node.AsVariableStatement().DeclarationList)
	case ast.KindBlock:
		body, err := l.statements(node.AsBlock().Statements.Nodes)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Block{Body: body}}, nil
	case ast.KindIfStatement:
		return l.ifStatement(node)
	case ast.KindForStatement:
		return l.forStatement(node)
	case ast.KindWhileStatement, ast.KindDoStatement:
		return l.whileStatement(node)
	case ast.KindBreakStatement, ast.KindContinueStatement:
		if node.Label() != nil {
			return nil, l.notYet(node, "a labeled "+strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(node.Kind.String(), "Kind"), "Statement")))
		}
		if node.Kind == ast.KindBreakStatement {
			return []ir.Statement{ir.Break{}}, nil
		}
		return []ir.Statement{ir.Continue{}}, nil
	}
	return nil, l.notYet(node, describe(node))
}

// expressionStatement lowers an expression used as a statement: a console call, an assignment, or
// ++ and --. Any other expression's value would be thrown away, and stage 0 doesn't lower that yet.
func (l *lowering) expressionStatement(expression *ast.Node) ([]ir.Statement, error) {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindCallExpression:
		statement, err := l.console(expression)
		if err != nil {
			return nil, err
		}
		return []ir.Statement{statement}, nil
	case ast.KindBinaryExpression:
		return l.assignment(expression)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return l.increment(expression)
	}
	return nil, l.notYet(expression, describe(expression)+" as a statement")
}

// console lowers console.log and console.error with one string.
func (l *lowering) console(call *ast.Node) (ir.Statement, error) {
	stream, err := l.consoleStream(call.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		// The prelude declares one parameter, so the checker has already refused any other count.
		return nil, errors.New("lower: " + l.program.Where(call) + ": console takes one argument, and the checker let another count through")
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, err
	}
	return ir.WriteLine{Stream: stream, Value: value}, nil
}

// variables lowers const and let declarations, each to a local of its own.
func (l *lowering) variables(list *ast.Node) ([]ir.Statement, error) {
	if list.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, &Refused{Where: l.program.Where(list), What: "var", Fix: "use const or let"}
	}
	statements := []ir.Statement{}
	for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
		name := declaration.Name()
		if !ast.IsIdentifier(name) {
			return nil, l.notYet(name, "a destructuring declaration")
		}
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		var value ir.Expression
		if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil {
			if value, err = l.expression(initializer); err != nil {
				return nil, err
			}
		}
		statements = append(statements, ir.Declare{Local: local, Value: value})
	}
	return statements, nil
}

// declareLocal makes a new local for a declared name, typed by what the checker proved for it.
func (l *lowering) declareLocal(name *ast.Node) (int, error) {
	valueType, err := l.typeOf(name)
	if err != nil {
		return 0, err
	}
	symbol := l.checker.GetSymbolAtLocation(name)
	if symbol == nil {
		return 0, errors.New("lower: " + l.program.Where(name) + ": the checker gave a declaration no symbol")
	}
	if l.locals == nil {
		l.locals = map[*ast.Symbol]int{}
	}
	l.locals[symbol] = len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name.Text(), Type: valueType})
	return l.locals[symbol], nil
}

// local finds the local an identifier refers to.
func (l *lowering) local(identifier *ast.Node) (int, bool) {
	symbol := l.checker.GetSymbolAtLocation(identifier)
	local, isLocal := l.locals[symbol]
	return local, isLocal
}

// assignment lowers =, and the compound assignments, to a local.
func (l *lowering) assignment(node *ast.Node) ([]ir.Statement, error) {
	binary := node.AsBinaryExpression()
	operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]
	if binary.OperatorToken.Kind != ast.KindEqualsToken && !isCompound {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	target := ast.SkipParentheses(binary.Left)
	local, isLocal := l.local(target)
	if !ast.IsIdentifier(target) || !isLocal {
		return nil, l.notYet(target, "assigning to "+describe(target))
	}
	value, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if isCompound {
		current := ir.Read{Local: local, Of: l.result.Locals[local].Type}
		if value, err = l.combine(node, operator, current, value); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{ir.Assign{Local: local, Value: value}}, nil
}

var compoundAssignments = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:             ast.KindPlusToken,
	ast.KindMinusEqualsToken:            ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:         ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:            ast.KindSlashToken,
	ast.KindPercentEqualsToken:          ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken: ast.KindAsteriskAsteriskToken,
}

// increment lowers ++ and -- on a number local, as a statement, where prefix and postfix agree.
func (l *lowering) increment(node *ast.Node) ([]ir.Statement, error) {
	var operator ast.Kind
	var operand *ast.Node
	if node.Kind == ast.KindPrefixUnaryExpression {
		operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	} else {
		operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	}
	if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	operand = ast.SkipParentheses(operand)
	local, isLocal := l.local(operand)
	if !ast.IsIdentifier(operand) || !isLocal {
		return nil, l.notYet(operand, "incrementing "+describe(operand))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	current := ir.Read{Local: local, Of: ir.Number}
	return []ir.Statement{ir.Assign{Local: local, Value: ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}}}, nil
}

func (l *lowering) ifStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsIfStatement()
	condition, err := l.condition(statement.Expression)
	if err != nil {
		return nil, err
	}
	then, err := l.statement(statement.ThenStatement)
	if err != nil {
		return nil, err
	}
	lowered := ir.If{Condition: condition, Then: then}
	if statement.ElseStatement != nil {
		if lowered.Else, err = l.statement(statement.ElseStatement); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{lowered}, nil
}

// forStatement lowers for (initializer; condition; incrementor) to a loop inside a block, so the
// loop's own declarations are scoped to it.
func (l *lowering) forStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForStatement()
	block := ir.Block{}
	if initializer := statement.Initializer; initializer != nil {
		var err error
		var initial []ir.Statement
		if initializer.Kind == ast.KindVariableDeclarationList {
			initial, err = l.variables(initializer)
		} else {
			initial, err = l.expressionStatement(initializer)
		}
		if err != nil {
			return nil, err
		}
		block.Body = append(block.Body, initial...)
	}
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}}
	if statement.Condition != nil {
		condition, err := l.condition(statement.Condition)
		if err != nil {
			return nil, err
		}
		loop.Condition = condition
	}
	if statement.Incrementor != nil {
		update, err := l.expressionStatement(statement.Incrementor)
		if err != nil {
			return nil, err
		}
		loop.Update = update
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	loop.Body = body
	block.Body = append(block.Body, loop)
	return []ir.Statement{block}, nil
}

// whileStatement lowers while and do...while.
func (l *lowering) whileStatement(node *ast.Node) ([]ir.Statement, error) {
	var expression, body *ast.Node
	checkAfter := node.Kind == ast.KindDoStatement
	if checkAfter {
		expression, body = node.AsDoStatement().Expression, node.AsDoStatement().Statement
	} else {
		expression, body = node.AsWhileStatement().Expression, node.AsWhileStatement().Statement
	}
	condition, err := l.condition(expression)
	if err != nil {
		return nil, err
	}
	loweredBody, err := l.statement(body)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Loop{Condition: condition, Body: loweredBody, CheckAfter: checkAfter}}, nil
}

// condition lowers an expression that decides a branch or a loop. 0.1 requires a boolean there
// (docs/0.1.md), so `if (name)` on a string is refused rather than given JavaScript's truthiness.
func (l *lowering) condition(node *ast.Node) (ir.Expression, error) {
	condition, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if condition.Type() != ir.Boolean {
		return nil, &Refused{Where: l.program.Where(node), What: "a " + typeName(condition.Type()) + " as a condition", Fix: "compare it explicitly, like name.length > 0 or count !== 0"}
	}
	return condition, nil
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
