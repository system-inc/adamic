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
// statement (first_link_units.go), so everything around it lowers on. Nothing is rolled back: what the
// unit's lowering recorded before it stopped stays, and every row after a recovered panic is Suspect.

// firstLinkBody lowers the body of the function at index as lowerBody does. When that stops, it records
// the trap and makes the trap the function's whole body, so the function that met the stop lowers on. Its
// parameters and locals stay as its signature and the stopped lowering left them, and nothing reads them,
// since the trap runs first and never returns. A function inside a unit trapped already lowers as that
// unit's stub, without lowering, and records no second row.
func (l *lowering) firstLinkBody(index int, declaration *ast.Node, pending signed) error {
	if stub, trapped := l.firstLinkTrappedAt(declaration); trapped {
		if l.functionNodes == nil {
			l.functionNodes = map[int]*ast.Node{}
		}
		l.functionNodes[index] = declaration
		l.result.Functions[index].Body = []ir.Statement{stub}
		return nil
	}
	stub, trapped, err := l.firstLinkCatch("Body", declaration, func() error {
		return l.lowerBody(index, declaration, pending.this, pending.defaults, pending.patterns)
	})
	if trapped {
		l.result.Functions[index].Body = []ir.Statement{stub}
	}
	return err
}

// firstLinkStatements lowers a module's top-level statements as statements does, each on its own, so a
// statement that stops lowers as its trap and the statements before it still lower.
func (l *lowering) firstLinkStatements(module *ast.SourceFile) ([]ir.Statement, error) {
	lowered, err := l.nestedDeclarations(module.Statements.Nodes)
	if err != nil {
		return nil, err
	}
	// lowerWithElementKey appends this module's body to Main as it stands, so each statement's IR starts
	// where Main ends now, past what the statements before it lowered to (firstLinkMainAt).
	ahead := len(l.result.Main)
	if l.firstLinkMain == nil {
		l.firstLinkMain = map[*ast.Node]firstLinkMainRange{}
	}
	for _, node := range module.Statements.Nodes {
		statements, err := l.firstLinkStatement(node)
		if err != nil {
			return nil, err
		}
		l.firstLinkMain[node] = firstLinkMainRange{start: ahead + len(lowered), length: len(statements)}
		lowered = append(lowered, statements...)
	}
	return lowered, nil
}

// firstLinkStatement lowers one top-level statement. When that stops, it records the trap and lowers the
// statement as the trap alone, so main's run ends there, as the stop would have ended the lowering. A
// statement trapped already lowers as its stub. A function declaration's own unit is its body, which
// firstLinkBody stubs, so its statement lowers as it always does.
func (l *lowering) firstLinkStatement(node *ast.Node) ([]ir.Statement, error) {
	if stub, trapped := l.firstLinkStubs[node]; trapped && !firstLinkFunctionUnit(node) {
		return []ir.Statement{stub}, nil
	}
	var statements []ir.Statement
	stub, trapped, err := l.firstLinkCatch("Statement", node, func() error {
		var err error
		statements, err = l.statement(node)
		return err
	})
	if trapped {
		return []ir.Statement{stub}, nil
	}
	return statements, err
}

// firstLinkSignature writes the signature of a module function as declareModule does. When that stops, it
// records the trap, makes the trap the function's body, and names the function in signatureTraps, so every
// reference to it stops where it's met (signatureTrap) and declareModule never lowers its body.
func (l *lowering) firstLinkSignature(index int, declaration *ast.Node) error {
	stub, trapped, err := l.firstLinkCatch("Signature", declaration, func() error {
		return l.signature(index, declaration, -1)
	})
	if trapped {
		if l.signatureTraps == nil {
			l.signatureTraps = map[*ast.Symbol]string{}
		}
		// firstLinkCatch has just recorded the signature's row.
		l.signatureTraps[l.symbol(declaration.Name())] = firstLinkMessage(l.traps[len(l.traps)-1])
		l.result.Functions[index].Body = []ir.Statement{stub}
	}
	return err
}

// firstLinkStatics makes a class's statics as declareModule does. When that stops, the class's statement
// is trapped, so main's run ends where the class is declared, and every function inside it lowers as the
// stub. A class trapped already makes none: its row was written when it was trapped, and its code never
// lowers (firstLinkInside).
func (l *lowering) firstLinkStatics(statement *ast.Node) error {
	if _, trapped := l.firstLinkTrappedAt(statement); trapped {
		return nil
	}
	_, _, err := l.firstLinkCatch("Statics", statement, func() error {
		_, err := l.staticInstance(statement)
		return err
	})
	return err
}

// signatureTrap is the NotYet a reference to a module function whose signature trapped stops with, so the
// unit holding the reference traps where it's met; nil for any other symbol.
func (l *lowering) signatureTrap(node *ast.Node, symbol *ast.Symbol) error {
	message, trapped := l.signatureTraps[symbol]
	if !trapped {
		return nil
	}
	return l.notYet(node, "a reference to "+symbol.Name+", whose signature trapped: "+message)
}

// firstLinkCatch runs one step of a unit's lowering. When it stops, by a NotYet, a Refused or a Go panic,
// it records the row with phase, traps unit, and returns its stub. An error that's neither a NotYet nor a
// Refused is no unit's stop, so it isn't trapped, and it's returned as main would return it.
func (l *lowering) firstLinkCatch(phase string, unit *ast.Node, run func() error) (stub ir.Panic, trapped bool, err error) {
	// The lowering puts these back when it returns an error, but not when it panics.
	outer, outerThis, outerIndex, outerClosures := l.function, l.this, l.functionIndex, l.closures
	defer func() {
		recovered := recover()
		if recovered == nil && err == nil {
			return
		}
		row, stopped := l.firstLinkRow(recovered, err, phase, unit, unit)
		if !stopped {
			return
		}
		l.function, l.this, l.functionIndex, l.closures = outer, outerThis, outerIndex, outerClosures
		stub, trapped, err = l.firstLinkRecord(row, unit), true, nil
	}()
	return ir.Panic{}, false, run()
}

// firstLinkRow is the row of a stop, by its error or by the Go panic recovered while it ran: Unit names
// unit, or is empty when the stop has none, and a panic, which carries no site of its own, is at site. An
// error that's neither a NotYet nor a Refused is no row.
func (l *lowering) firstLinkRow(recovered any, err error, phase string, site, unit *ast.Node) (ir.Trap, bool) {
	row := ir.Trap{Phase: phase, Suspect: l.firstLinkPanicked || l.firstLinkRewriteSkipped}
	var notYet *NotYet
	var refused *Refused
	switch {
	case recovered != nil:
		row.Kind, row.Reason = "Panic", fmt.Sprint(recovered)
		if site != nil {
			row.Site = l.program.Where(site)
		}
	case errors.As(err, &notYet):
		row.Kind, row.Reason, row.Site = "NotYet", notYet.What, notYet.Where
	case errors.As(err, &refused):
		row.Kind, row.Reason, row.Fix, row.Site = "Refused", refused.What, refused.Fix, refused.Where
	default:
		return ir.Trap{}, false
	}
	if unit != nil {
		row.Unit = l.firstLinkUnitName(unit)
	}
	return row, true
}

// firstLinkRecord records row and, when the row has a unit, traps it: the unit and everything inside it
// lower as the stub it returns from then on (firstLinkTrappedAt). A row with no unit is a measurement gap,
// and its stub is never used.
func (l *lowering) firstLinkRecord(row ir.Trap, unit *ast.Node) ir.Panic {
	l.traps = append(l.traps, row)
	if row.Kind == "Panic" {
		l.firstLinkPanicked = true
	}
	if unit == nil {
		return ir.Panic{}
	}
	stub := ir.Panic{Message: ir.StringConstant{Index: l.constant(firstLinkMessage(row))}, Trap: true}
	if l.firstLinkTrapped == nil {
		l.firstLinkTrapped = map[*ast.Node]bool{}
		l.firstLinkStubs = map[*ast.Node]ir.Panic{}
	}
	l.firstLinkTrapped[unit] = true
	l.firstLinkStubs[unit] = stub
	return stub
}

// firstLinkMessage is what a row's trap writes after the first-link prefix when a run reaches it.
func firstLinkMessage(row ir.Trap) string {
	message := row.Kind + ": " + row.Reason
	if row.Fix != "" {
		message += " (fix: " + row.Fix + ")"
	}
	return message + " at " + row.Site
}

// firstLinkUnit names a trapped function as a Trap's Unit does: its module's path, then its declaration
// chain joined by dots.
func firstLinkUnit(source ir.SourceIdentity) string {
	if len(source.Declaration) == 0 {
		return source.Module
	}
	return source.Module + " " + strings.Join(source.Declaration, ".")
}
