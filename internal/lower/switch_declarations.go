package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// switchBindings creates storage for every direct case binding before any body runs.
// Explicit blocks keep their own declarations. Each entry gets fresh storage and flags.
func (l *lowering) switchBindings(block *ast.Node) ([]ir.Statement, error) {
	storage := []ir.Statement{}
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
		storage = append(storage, ir.Declare{Local: ready, Value: ir.BooleanConstant{}}, ir.Declare{Local: local})
		return nil
	}
	for _, clause := range block.AsCaseBlock().Clauses.Nodes {
		for _, statement := range clause.AsCaseOrDefaultClause().Statements.Nodes {
			if statement.Kind != ast.KindVariableStatement {
				continue
			}
			list := statement.AsVariableStatement().DeclarationList
			if list.Flags&ast.NodeFlagsBlockScoped == 0 {
				return nil, &Refused{Where: l.program.Where(list), What: "var", Fix: "use const or let"}
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				if err := bind(declaration.Name()); err != nil {
					return nil, err
				}
			}
		}
	}
	return storage, nil
}

// initializeSwitchBindings writes into preallocated storage, then marks it ready.
// A pattern may lower into nested blocks; each individual binding initializes in order.
func (l *lowering) initializeSwitchBindings(body []ir.Statement) []ir.Statement {
	result := []ir.Statement{}
	for _, statement := range body {
		switch declared := statement.(type) {
		case ir.Declare:
			if ready := l.result.Locals[declared.Local].Ready; ready != 0 {
				if declared.Value != nil {
					result = append(result, ir.Assign{Local: declared.Local, Value: declared.Value})
				}
				result = append(result, ir.Assign{Local: ready - 1, Value: ir.BooleanConstant{Value: true}})
				continue
			}
		case ir.Block:
			declared.Body = l.initializeSwitchBindings(declared.Body)
			statement = declared
		}
		result = append(result, statement)
	}
	return result
}
