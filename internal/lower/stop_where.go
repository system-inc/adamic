package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// A judging pass's stop names where it lies through stopWhere, so a first-link lowering places it in a unit and
// traps that unit (firstLinkPass), where a stop naming only the file was a Gap that gave up the rest of the
// pass: checkAccessorSpreads and checkRangeUnionReads did on tsc (#nxb3xh1). stop_where_source_test.go holds
// every judging pass's stops to it.

// stopSite is what a judging pass knows of where its stop lies: the node it met, a where it read, the function
// whose body holds the stop, or the statement of Main that does. function and statement are -1 when unknown.
type stopSite struct {
	node      *ast.Node
	where     string
	function  int
	statement int
}

// stopWhere is where a judging pass's stop lies, inside a unit, or the unit's own declaration when the stop is
// about the unit as a whole: the node's where; else where, when it names a position; else the declaration the
// function's body was lowered from; else the top-level statement whose IR holds that statement of Main, by where
// firstLinkMainAt finds each now. Otherwise it's "", no site: never the file alone and never a module's node,
// which name no unit. An empty where is unplaced, so a first-link lowering records the stop as a Gap and the
// pass's judgment as abandoned, which fails closed. Only a first-link lowering records where each top-level
// statement's IR lies, so a statement of Main is placed in no other.
func (l *lowering) stopWhere(site stopSite) string {
	if site.node != nil && site.node.Kind != ast.KindSourceFile {
		return l.program.Where(site.node)
	}
	if _, _, _, positioned := firstLinkSite(site.where); positioned {
		return site.where
	}
	if declaration := l.functionNodes[site.function]; site.function >= 0 && declaration != nil {
		return l.program.Where(declaration)
	}
	if site.statement >= 0 {
		for statement := range l.firstLinkMain {
			if start, length, placed := l.firstLinkMainAt(statement); placed && start <= site.statement && site.statement < start+length {
				return l.program.Where(statement)
			}
		}
	}
	return ""
}
