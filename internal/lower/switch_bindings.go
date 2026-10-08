package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// initializeLocal runs a lexical declaration. Switch bindings already have storage: writing the
// value initializes them only after the initializer has finished, including any calls it makes.
func (l *lowering) initializeLocal(local int, value ir.Expression) []ir.Statement {
	ready := l.result.Locals[local].Ready
	if ready == 0 {
		return []ir.Statement{ir.Declare{Local: local, Value: value}}
	}
	if value == nil {
		// No initializer still ends the dead zone. Storage already holds its zero representation.
		return []ir.Statement{ir.Assign{Local: ready - 1, Value: ir.BooleanConstant{Value: true}}}
	}
	return []ir.Statement{
		ir.Assign{Local: local, Value: value},
		ir.Assign{Local: ready - 1, Value: ir.BooleanConstant{Value: true}},
	}
}

// switchBindings creates the one lexical environment shared by every case. Every variable's
// storage and readiness cell exists before a function can capture it. Functions are hoisted and
// initialized at entry; let and const remain unready until execution reaches their declarations.
func (l *lowering) switchBindings(block *ast.Node) ([]ir.Statement, error) {
	var storage []ir.Statement
	var functions []*ast.Node
	var bind func(*ast.Node) error
	bind = func(name *ast.Node) error {
		if name.Kind == ast.KindArrayBindingPattern || name.Kind == ast.KindObjectBindingPattern {
			for _, element := range name.AsBindingPattern().Elements.Nodes {
				if !skipped(element) {
					if err := bind(element.Name()); err != nil {
						return err
					}
				}
			}
			return nil
		}
		local, err := l.declareLocal(name)
		if err != nil {
			return err
		}
		ready := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name.Text() + "_ready", Type: ir.Boolean, Function: l.functionIndex})
		l.result.Locals[local].Ready = ready + 1
		storage = append(storage, ir.Declare{Local: ready, Value: ir.BooleanConstant{Value: false}}, ir.Declare{Local: local})
		return nil
	}
	for _, clause := range block.AsCaseBlock().Clauses.Nodes {
		for _, statement := range clause.AsCaseOrDefaultClause().Statements.Nodes {
			switch statement.Kind {
			case ast.KindVariableStatement:
				for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					if err := bind(declaration.Name()); err != nil {
						return nil, err
					}
				}
			case ast.KindFunctionDeclaration:
				if err := bind(statement.Name()); err != nil {
					return nil, err
				}
				functions = append(functions, statement)
			}
		}
	}
	for _, declaration := range functions {
		value, err := l.functionExpression(declaration)
		if err != nil {
			return nil, err
		}
		local := l.locals[l.symbol(declaration.Name())]
		storage = append(storage, l.initializeLocal(local, value)...)
	}
	return storage, nil
}
