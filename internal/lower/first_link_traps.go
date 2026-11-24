package lower

import (
	"errors"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A first-link lowering (Options.FirstLinkTraps) lowers a unit that stops as a trap instead of stopping
// there: its NotYet, its Refused, or a Go panic recovered while it lowered becomes a row of the program's
// Traps and one Panic with Trap set, which ends the program with the first-link prefix and exit 75 when a
// run reaches it. The unit is the innermost function the stop happens in, or a module's top-level
// statement, so everything around it lowers on. Nothing is rolled back: what the unit's lowering recorded
// before it stopped stays, and every row after a recovered panic is Suspect.

// firstLinkBody lowers the body of the function at index as lowerBody does. When that stops, it records
// the trap and makes the trap the function's whole body, so the function that met the stop lowers on. Its
// parameters and locals stay as its signature and the stopped lowering left them, and nothing reads them,
// since the trap runs first and never returns.
func (l *lowering) firstLinkBody(index int, declaration *ast.Node, pending signed) (err error) {
	// lowerBody puts these back when it returns an error, but not when it panics.
	outer, outerThis, outerIndex, outerClosures := l.function, l.this, l.functionIndex, l.closures
	defer func() {
		recovered := recover()
		if recovered == nil && err == nil {
			return
		}
		trap, trapped := l.firstLinkTrap(recovered, err, "Body", declaration, firstLinkUnit(l.result.Functions[index].Source))
		if !trapped {
			return
		}
		l.function, l.this, l.functionIndex, l.closures = outer, outerThis, outerIndex, outerClosures
		l.result.Functions[index].Body = []ir.Statement{trap}
		err = nil
	}()
	return l.lowerBody(index, declaration, pending.this, pending.defaults, pending.patterns)
}

// firstLinkStatements lowers a module's top-level statements as statements does, each on its own, so a
// statement that stops lowers as its trap and the statements before it still lower.
func (l *lowering) firstLinkStatements(module *ast.SourceFile) ([]ir.Statement, error) {
	lowered, err := l.nestedDeclarations(module.Statements.Nodes)
	if err != nil {
		return nil, err
	}
	for _, node := range module.Statements.Nodes {
		statements, err := l.firstLinkStatement(module, node)
		if err != nil {
			return nil, err
		}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// firstLinkStatement lowers one top-level statement. When that stops, it records the trap and lowers the
// statement as the trap alone, so main's run ends there, as the stop would have ended the lowering.
func (l *lowering) firstLinkStatement(module *ast.SourceFile, node *ast.Node) (statements []ir.Statement, err error) {
	// A top-level statement's lowering puts these back when it returns, but not when it panics.
	outer, outerThis, outerIndex, outerClosures := l.function, l.this, l.functionIndex, l.closures
	defer func() {
		recovered := recover()
		if recovered == nil && err == nil {
			return
		}
		trap, trapped := l.firstLinkTrap(recovered, err, "Statement", node, l.modulePath(module)+" module")
		if !trapped {
			return
		}
		l.function, l.this, l.functionIndex, l.closures = outer, outerThis, outerIndex, outerClosures
		statements, err = []ir.Statement{trap}, nil
	}()
	return l.statement(node)
}

// firstLinkTrap records the trap of a unit that stopped, by its error or by the Go panic recovered while
// it lowered, and returns the Panic the unit lowers as. An error that's neither a NotYet nor a Refused is
// no unit's stop, so it isn't trapped, and the caller returns it as main would.
func (l *lowering) firstLinkTrap(recovered any, err error, phase string, node *ast.Node, unit string) (ir.Panic, bool) {
	trap := ir.Trap{Unit: unit, Phase: phase, Suspect: l.firstLinkPanicked}
	var notYet *NotYet
	var refused *Refused
	switch {
	case recovered != nil:
		// A panic carries no site of its own, so its site is the unit's.
		trap.Kind, trap.Reason, trap.Site = "Panic", fmt.Sprint(recovered), l.program.Where(node)
		l.firstLinkPanicked = true
	case errors.As(err, &notYet):
		trap.Kind, trap.Reason, trap.Site = "NotYet", notYet.What, notYet.Where
	case errors.As(err, &refused):
		trap.Kind, trap.Reason, trap.Fix, trap.Site = "Refused", refused.What, refused.Fix, refused.Where
	default:
		return ir.Panic{}, false
	}
	l.traps = append(l.traps, trap)
	message := trap.Kind + ": " + trap.Reason
	if trap.Fix != "" {
		message += " (fix: " + trap.Fix + ")"
	}
	message += " at " + trap.Site
	return ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}, Trap: true}, true
}

// firstLinkUnit names a trapped function as a Trap's Unit does: its module's path, then its declaration
// chain joined by dots.
func firstLinkUnit(source ir.SourceIdentity) string {
	if len(source.Declaration) == 0 {
		return source.Module
	}
	return source.Module + " " + strings.Join(source.Declaration, ".")
}
