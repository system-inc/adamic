package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// The whole-program passes after the bodies lower judge, and some rewrite, the IR the bodies made. In a
// first-link lowering, a pass's stop traps the unit it lies in, in the IR already made, and the pass goes
// on past it where it safely can (firstLinkPass). The stops lowering holds rather than returns where it
// meets them are recorded the same way, once the bodies are lowered (firstLinkHeldStops), and a stop that
// names the whole program, never a unit, is a row with no unit (firstLinkProgramStop).

// firstLinkMainRange is where a top-level statement's IR lies in the program's Main, as moduleStatements
// lowered it: length statements from start, counted before forwarderValues are put ahead of Main.
type firstLinkMainRange struct {
	start, length int
}

// firstLinkPass is a whole-program pass's stop err, which any lowering but a first-link one returns as it
// always has. A first-link lowering records it as a row with phase name and goes on. A stop that can't be
// placed in a unit, or at a site this pass stopped at before, is a row of kind "Gap", and the rest of the
// pass is skipped. A stop inside a unit trapped already is an Echo, no wall of its own, and the rest of the
// pass is skipped. Otherwise the stop's unit is trapped in the IR already made (firstLinkStubUnit). A pass
// that rewrites the IR never runs twice, since that could apply its rewrite twice, so its work past the
// stop is skipped and named in Program.TrapRewritesSkipped, and every row after it is Suspect. Any other
// pass runs again by run, and its next stop is taken the same way. A nil run is a stop nothing can run
// again. An error that's neither a NotYet nor a Refused is a row of kind "Error" with no unit, and the
// rest of the pass is skipped. A pass that only judges and gives up so, at a Gap, an Echo or an Error,
// leaves every unit past that stop unjudged, so it's named in Program.TrapJudgmentsAbandoned with the
// stop's site and kind, and every row after it is Suspect (#nxb3xh1). Only a context's cancellation or
// deadline is returned.
//
// A first-link lowering hands every whole-program pass here, stopped or not (err nil), so the checker's
// relation budget is read after each of its runs, the first included, and once more before each run again,
// and a pass that finds it spent is abandoned and never runs again (firstLinkBudgetSpent). The held stops
// come here before the budget's baseline is read, so it isn't read for them.
func (l *lowering) firstLinkPass(name string, rewrites bool, err error, run func() error) error {
	if !l.firstLinkTraps {
		return err
	}
	seen := map[string]bool{}
	for {
		if l.firstLinkBudgetSpent(name, rewrites) {
			return nil
		}
		row, stopped := l.firstLinkRow(nil, err, name, nil, nil)
		if !stopped {
			return err
		}
		if row.Kind == "Error" {
			// It names no site to place, and running the pass again could meet it again.
			l.firstLinkRecord(row, nil)
			l.firstLinkRewriteSkip(name, rewrites)
			l.firstLinkJudgmentAbandon(name, rewrites, row)
			return nil
		}
		unit := l.firstLinkPassUnit(row.Site)
		_, echo := l.firstLinkTrappedAt(unit)
		switch {
		case unit == nil || seen[row.Site]:
			row.Kind, row.Reason = "Gap", row.Kind+": "+row.Reason
			l.firstLinkRecord(row, nil)
			l.firstLinkRewriteSkip(name, rewrites)
			l.firstLinkJudgmentAbandon(name, rewrites, row)
			return nil
		case echo:
			row.Unit, row.Echo = l.firstLinkUnitName(unit), true
			l.firstLinkRecord(row, nil)
			l.firstLinkRewriteSkip(name, rewrites)
			l.firstLinkJudgmentAbandon(name, rewrites, row)
			return nil
		}
		seen[row.Site] = true
		row.Unit = l.firstLinkUnitName(unit)
		l.firstLinkStubUnit(unit, l.firstLinkRecord(row, unit))
		if rewrites || run == nil {
			l.firstLinkRewriteSkip(name, rewrites)
			return nil
		}
		if l.firstLinkBudgetSpent(name, rewrites) {
			return nil
		}
		err = run()
	}
}

// firstLinkBudgetBaseline reads, in a first-link lowering, whether the checker's relation budget is spent
// before the first judging pass runs. Spent already, it's the program's and no pass's: it's recorded once as
// the RelationOverflow row with no unit, as lowerCheckedWithCycleTypes records one, and every judging pass
// after it abandons (firstLinkBudgetSpent).
func (l *lowering) firstLinkBudgetBaseline() {
	if !l.firstLinkTraps {
		return
	}
	l.firstLinkBudgetRead = true
	l.firstLinkOverflow = l.firstLinkOverflowNow()
	if l.firstLinkOverflow != "" {
		l.firstLinkProgramStop("RelationOverflow", l.relationOverflowStop(l.firstLinkOverflow))
	}
}

// firstLinkBudgetSpent reads, past the baseline, whether the checker's relation budget is spent, and when it
// is, abandons the judging pass name with no run again and reports that it did. Past the budget every
// relation query overflows, reports TS2859 and dedupes it against every diagnostic before it, so a run goes
// quadratic and may never return, as findCycles' third run on tsc didn't (J-5cj56ea). It's read as a change
// from the baseline, never as a flag set once: the pass during whose run the budget is first seen spent gets
// the Error row naming it, and every pass after it gets the same row, so each is named in
// TrapJudgmentsAbandoned, or TrapRewritesSkipped for one that rewrites, and the first names the wall.
// RelationOverflow ranges over every relation cache, up to some 16M pairs, so firstLinkPass reads it a
// bounded number of times, once after each run and once before each run again, never in a walk over units,
// and once it's seen spent it isn't read again, since a spent budget stays spent.
//
// A read between runs doesn't bound one run that overflows partway: that run goes on, as long as it takes,
// before anything here reads the budget, and would be its own wall.
func (l *lowering) firstLinkBudgetSpent(name string, rewrites bool) bool {
	if !l.firstLinkBudgetRead {
		return false
	}
	if l.firstLinkOverflow == "" {
		if l.firstLinkOverflow = l.firstLinkOverflowNow(); l.firstLinkOverflow == "" {
			return false
		}
	}
	row, _ := l.firstLinkRow(nil, errors.New("the checker's relation budget is spent: "+l.firstLinkOverflow), name, nil, nil)
	l.firstLinkRecord(row, nil)
	l.firstLinkRewriteSkip(name, rewrites)
	l.firstLinkJudgmentAbandon(name, rewrites, row)
	return true
}

// firstLinkOverflowNow is RelationOverflow's text for the lowering's checker now, or a test's stand-in's.
func (l *lowering) firstLinkOverflowNow() string {
	if l.firstLinkOverflowCheck != nil {
		return l.firstLinkOverflowCheck()
	}
	return l.program.RelationOverflow(l.checker)
}

// firstLinkRewriteSkip notes that a pass that rewrites the IR skipped its work past a stop.
func (l *lowering) firstLinkRewriteSkip(name string, rewrites bool) {
	if !rewrites {
		return
	}
	l.firstLinkRewriteSkipped = true
	l.result.TrapRewritesSkipped = append(l.result.TrapRewritesSkipped, name)
}

// firstLinkJudgmentAbandon notes that a pass that only judges the IR gave up past row, its stop: a Gap, an
// Echo or an Error, by row's own kind. A rewriting pass's skip is firstLinkRewriteSkip's to note.
func (l *lowering) firstLinkJudgmentAbandon(name string, rewrites bool, row ir.Trap) {
	if rewrites {
		return
	}
	why := row.Kind
	if row.Echo {
		why = "Echo"
	}
	l.firstLinkJudgmentAbandoned = true
	l.result.TrapJudgmentsAbandoned = append(l.result.TrapJudgmentsAbandoned, ir.TrapJudgmentAbandoned{Pass: name, Site: row.Site, Why: why})
}

// firstLinkPassUnit places a whole-program pass's stop, by its Where, in the unit it lies in. A unit
// trapped already is placed too, so its stop is an Echo. A top-level statement whose IR can't be found
// in Main can't be trapped there, so its stop is unplaced.
func (l *lowering) firstLinkPassUnit(where string) *ast.Node {
	modules, err := l.moduleOrder(l.entry)
	if err != nil {
		return nil
	}
	for _, module := range modules {
		unit := l.firstLinkPlace(module, where)
		if unit == nil {
			continue
		}
		if _, trapped := l.firstLinkTrappedAt(unit); trapped {
			return unit
		}
		if _, _, placed := l.firstLinkMainAt(unit); !firstLinkFunctionUnit(unit) && !placed {
			return nil
		}
		return unit
	}
	return nil
}

// firstLinkMainAt is where a top-level statement's IR lies in Main now. Main holds every module's body,
// as MainModules counts them, behind forwarderValues once they're put ahead of it. A Main of any other
// length was reshaped by a pass, as NormalizeAsync moves module bodies into its scheduler entry, so no
// statement is placed in it.
func (l *lowering) firstLinkMainAt(unit *ast.Node) (start, length int, placed bool) {
	at, recorded := l.firstLinkMain[unit]
	if !recorded || l.result.AsyncEntry > 0 {
		return 0, 0, false
	}
	bodies := 0
	for _, module := range l.result.MainModules {
		bodies += module.Statements
	}
	shift := len(l.result.Main) - bodies
	if shift != 0 && shift != len(l.forwarderValues) || at.start+shift+at.length > len(l.result.Main) {
		return 0, 0, false
	}
	return at.start + shift, at.length, true
}

// firstLinkStubUnit traps a unit in the IR already made: every function lowered from a declaration inside
// it, the unit itself included, every generic instance among them, has the stub for its body, and a
// top-level statement's own statements in Main become the stub, then empty blocks, so Main keeps its
// length. Then every helper whose owners are all trapped now is stubbed too (firstLinkStubHelpers).
func (l *lowering) firstLinkStubUnit(unit *ast.Node, stub ir.Panic) {
	defer l.firstLinkStubHelpers()
	for index, declaration := range l.functionNodes {
		l.firstLinkStubDeclared(index, declaration, unit, stub)
	}
	if firstLinkFunctionUnit(unit) {
		return
	}
	if start, length, placed := l.firstLinkMainAt(unit); placed && length > 0 {
		l.result.Main[start] = stub
		for index := start + 1; index < start+length; index++ {
			l.result.Main[index] = ir.Block{}
		}
	}
}

// firstLinkStubDeclared stubs the function at index, lowered from declaration, when the declaration lies inside
// unit, unit itself included.
func (l *lowering) firstLinkStubDeclared(index int, declaration, unit *ast.Node, stub ir.Panic) {
	for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor == unit {
			l.result.Functions[index].Body = []ir.Statement{stub}
			return
		}
	}
}

// holdUnlowerable holds a construct found not lowerable where no error can be returned. The first is
// l.unlowerable, which any lowering but a first-link one returns once every module has lowered. A first-link
// lowering holds every one, in the order met, so each traps its own unit (firstLinkHeldStops): holding the
// first alone left every later one untrapped, its unit lowered as it was, once the first's unit trapped.
func (l *lowering) holdUnlowerable(stop error) {
	if l.unlowerable == nil {
		l.unlowerable = stop
	}
	if l.firstLinkTraps {
		l.firstLinkUnlowerables = append(l.firstLinkUnlowerables, stop)
	}
}

// firstLinkHeldStops records, in a first-link lowering, the stops lowering held rather than returned where
// it met them: every unlowerable construct (holdUnlowerable), and every async payload it couldn't lower.
// Each is placed and trapped as a pass's stop is. An unlowerable construct inside a unit trapped already,
// by its body's own trap or an earlier held stop's, is passed over (firstLinkSkipsAt): its code is the stub,
// where its row was an Echo that abandoned the judgment (createScanner's scanNumber on tsc, #nxb3xh1). One
// that's neither a NotYet nor a Refused stays held, so it's returned as it always is. First, every helper whose
// owners the bodies' and statements' traps left all trapped is stubbed with them (firstLinkStubHelpers),
// before any pass judges its code.
func (l *lowering) firstLinkHeldStops() {
	if !l.firstLinkTraps {
		return
	}
	l.firstLinkStubHelpers()
	if l.unlowerable != nil {
		var kept error
		for _, stop := range l.firstLinkUnlowerables {
			if l.firstLinkSkipsAt(firstLinkStopSite(stop)) {
				continue
			}
			if err := l.firstLinkPass("unlowerable", false, stop, nil); err != nil && kept == nil {
				kept = err
			}
		}
		l.unlowerable = kept
	}
	var held []error
	for _, failure := range l.asyncPayloadFailures {
		if err := l.firstLinkPass("asyncPayloadFailures", false, failure, nil); err != nil {
			held = append(held, err)
		}
	}
	l.asyncPayloadFailures = held
}

// firstLinkStopSite is the Where of a NotYet or a Refused, or "" for any other error.
func firstLinkStopSite(stop error) string {
	var notYet *NotYet
	var refused *Refused
	switch {
	case errors.As(stop, &notYet):
		return notYet.Where
	case errors.As(stop, &refused):
		return refused.Where
	}
	return ""
}

// firstLinkProgramStop records, in a first-link lowering, a stop that names the whole program, never a
// unit, as a row with phase and no unit, and reports whether it did, so the caller goes on. Any other
// lowering, or an error that's neither a NotYet nor a Refused, is returned by the caller as it always is.
func (l *lowering) firstLinkProgramStop(phase string, stop error) bool {
	if !l.firstLinkTraps {
		return false
	}
	row, stopped := l.firstLinkRow(nil, stop, phase, nil, nil)
	if !stopped {
		return false
	}
	l.firstLinkRecord(row, nil)
	// lowerWithElementKey hands its rows over as it returns, which may be before this stop.
	l.result.Traps = l.traps
	return true
}

// firstLinkSuspectEvery marks every row recorded so far Suspect, and every row after it (firstLinkSuspect),
// for a stop that may have changed what lowering read anywhere in the program, before its row as well as
// after: a relation the checker gave up on (RelationOverflow).
func (l *lowering) firstLinkSuspectEvery() {
	l.firstLinkSuspect = true
	for index := range l.traps {
		l.traps[index].Suspect = true
	}
	l.result.Traps = l.traps
}
