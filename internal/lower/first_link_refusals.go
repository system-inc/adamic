package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// firstLinkRefuse is refuse in a first-link lowering. Where refuse returns a module's first refusal, this
// records each as a row with phase "Refusal" and traps the unit it lies in (first_link_units.go), which
// then lowers as its stub, and judges on past it: one run names every refusal in every unit. A trapped
// unit is never judged again, by the registry's walk or by a pre-check, since its code never lowers. A
// directive, a pragma or any other stop no unit holds is a row with no unit, a measurement gap, and the
// judging goes on past it. Its rules and its order are refuse's own: only what happens at a stop differs.
// An error that's neither a NotYet nor a Refused is returned, as refuse returns it.
func (l *lowering) firstLinkRefuse(module *ast.SourceFile) error {
	// As refuse words them, each one, where refuse stops at the first.
	for _, directive := range module.CommentDirectives {
		name := "@ts-ignore"
		if directive.Kind == ast.CommentDirectiveKindExpectError {
			name = "@ts-expect-error"
		}
		line, column := scanner.GetLineAndCharacterOfPosition(module, directive.Loc.Pos())
		l.firstLinkRefusalGap(&Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: name + " suppression directive", Fix: "remove it and fix the type error"})
	}
	for _, pragma := range module.Pragmas {
		if pragma.Name == "ts-nocheck" || pragma.Name == "ts-check" {
			line, column := scanner.GetLineAndCharacterOfPosition(module, pragma.Pos())
			l.firstLinkRefusalGap(&Refused{Where: fmt.Sprintf("%s:%d:%d", l.program.FileName(module), line+1, column+1), What: "@" + pragma.Name + " checking pragma", Fix: "remove it and fix any type errors"})
		}
	}
	// refuse's pre-checks, in its order.
	for _, check := range []func(*ast.SourceFile) error{l.referenceRefusal, l.parallelPreflight, l.mergedFieldRulings, l.functionAnnotations, l.predicateFalseReads, l.firstLinkContracts} {
		if err := l.firstLinkPrecheck(module, check); err != nil {
			return err
		}
	}
	return l.firstLinkRefusalWalk(module)
}

// firstLinkRefusalGap records a refusal stop no unit holds.
func (l *lowering) firstLinkRefusalGap(stop error) {
	if row, stopped := l.firstLinkRow(nil, stop, "Refusal", nil, nil); stopped {
		l.firstLinkRecord(row, nil)
	}
}

// firstLinkPrecheck runs one of refuse's pre-checks until it finds nothing more. A pre-check returns its
// first stop and names it only by its Where, so each stop is placed (firstLinkPlace), its unit trapped, and
// the check run again, whose walk passes over the trapped unit. A stop that can't be placed, or that lies
// in a unit trapped already, which the check doesn't pass over, ends the check with a row with no unit, so
// the loop is bounded by the module's units. A Go panic in the check ends it the same way.
func (l *lowering) firstLinkPrecheck(module *ast.SourceFile, check func(*ast.SourceFile) error) error {
	for {
		var recovered any
		err := func() error {
			defer func() { recovered = recover() }()
			return check(module)
		}()
		if recovered == nil && err == nil {
			return nil
		}
		row, stopped := l.firstLinkRow(recovered, err, "Refusal", module.AsNode(), nil)
		if !stopped {
			return err
		}
		var unit *ast.Node
		if recovered == nil {
			unit = l.firstLinkPlace(module, row.Site)
		}
		if unit == nil {
			l.firstLinkRecord(row, nil)
			return nil
		}
		if _, trapped := l.firstLinkTrappedAt(unit); trapped {
			l.firstLinkRecord(row, nil)
			return nil
		}
		row.Unit = l.firstLinkUnitName(unit)
		l.firstLinkRecord(row, unit)
	}
}

// firstLinkContracts is refuse's walk of the module's calls for their predicate contracts, passing over
// trapped units.
func (l *lowering) firstLinkContracts(module *ast.SourceFile) error {
	var contractError error
	var contracts ast.Visitor
	contracts = func(node *ast.Node) bool {
		if contractError != nil {
			return true
		}
		if l.firstLinkTrapped[node] {
			return false
		}
		if node.Kind == ast.KindCallExpression {
			contractError = l.predicateArguments(node)
		}
		if contractError == nil {
			node.ForEachChild(contracts)
		}
		return contractError != nil
	}
	module.AsNode().ForEachChild(contracts)
	return contractError
}

// firstLinkStop is a stop the registry's walk met at a node, kept to place once the walk is done.
type firstLinkStop struct {
	node *ast.Node
	stop error
}

// firstLinkRefusalWalk is refuse's walk of the registry's rules over the module. A rule's stop at a node
// traps the node's unit, and the walk leaves that unit and goes on after it. A deferred stop, the lowest of
// them (refusalPass's deferred), is kept with its node and placed once the walk is done, so a stronger stop
// in its unit traps the unit first, as refuse prefers the stronger one in the module.
func (l *lowering) firstLinkRefusalWalk(module *ast.SourceFile) error {
	var deferred []firstLinkStop
	// stopped is an error that's no unit's stop, which ends the walk as it ends refuse's. leaving is the unit
	// just trapped, which the walk leaves without visiting more of it.
	var stopped error
	var leaving *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if stopped != nil || leaving != nil {
			return true
		}
		if l.firstLinkTrapped[node] {
			return false
		}
		if branch := l.literalCallableBranch(node); branch != nil {
			visit(branch)
			if leaving == node {
				leaving = nil
			}
			return stopped != nil || leaving != nil
		}
		var stop error
		pruned := false
		// Each rule registers from its own file (refusal_rules.go); the first that refuses a node wins.
		for _, rule := range refusalRules.rules {
			prune, err := rule.check(l, node)
			if err != nil {
				stop = err
				break
			}
			if prune {
				pruned = true
				break
			}
		}
		if l.refusal.deferred != nil {
			deferred = append(deferred, firstLinkStop{node: node, stop: l.refusal.deferred})
			l.refusal.deferred = nil
		}
		if stop != nil {
			unit := firstLinkUnitOf(node)
			row, isStop := l.firstLinkRow(nil, stop, "Refusal", node, unit)
			if !isStop {
				stopped = stop
				return true
			}
			l.firstLinkRecord(row, unit)
			if unit == nil || unit == node {
				return false
			}
			leaving = unit
			return true
		}
		if pruned {
			return false
		}
		node.ForEachChild(visit)
		if leaving == node {
			leaving = nil
		}
		return stopped != nil || leaving != nil
	}
	l.refusal.deferred = nil
	module.AsNode().ForEachChild(func(statement *ast.Node) bool {
		l.firstLinkRefusalStatement(statement, visit)
		if leaving != nil {
			// The unit trapped lies outside the statement the walk was in, so the rest of the statement went
			// unjudged, and the statement is trapped whole.
			if _, trapped := l.firstLinkTrappedAt(statement); !trapped {
				l.firstLinkRecord(ir.Trap{Kind: "Gap", Reason: "the refusal walk left " + l.firstLinkUnitName(leaving) + ", which is outside this statement", Site: l.program.Where(statement), Unit: l.firstLinkUnitName(statement), Phase: "Refusal", Suspect: l.firstLinkPanicked}, statement)
			}
			leaving = nil
		}
		return stopped != nil
	})
	if stopped != nil {
		return stopped
	}
	for _, one := range deferred {
		unit := firstLinkUnitOf(one.node)
		if _, trapped := l.firstLinkTrappedAt(one.node); trapped {
			continue
		}
		if row, isStop := l.firstLinkRow(nil, one.stop, "Refusal", one.node, unit); isStop {
			l.firstLinkRecord(row, unit)
		}
	}
	return nil
}

// firstLinkRefusalStatement walks one top-level statement. A Go panic in a rule traps the statement whole,
// since the rest of it went unjudged.
func (l *lowering) firstLinkRefusalStatement(statement *ast.Node, visit ast.Visitor) {
	defer func() {
		if recovered := recover(); recovered != nil {
			row, _ := l.firstLinkRow(recovered, nil, "Refusal", statement, statement)
			l.firstLinkRecord(row, statement)
		}
	}()
	visit(statement)
}
