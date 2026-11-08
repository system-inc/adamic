package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// graphUnplannedSite is one allocation lowering builds outside expression: where, and what it is.
type graphUnplannedSite struct {
	node *ast.Node
	what string
}

// unplannedAllocation records an allocation lowering built outside expression, which graphAllocation
// never plans, unless its element is cycleInert: no value it holds can lead back to it. An element type
// the caller couldn't name is recorded, failing closed.
func (l *lowering) unplannedAllocation(node *ast.Node, element *checker.Type, what string) {
	if element == nil || !l.cycleInert(element, map[*checker.Type]bool{}) {
		l.graphUnplanned = append(l.graphUnplanned, graphUnplannedSite{node: node, what: what})
	}
}

// cycleInert reports whether no value of a type can hold anything a cycle could return through: a
// primitive (interfaceScalar's string, number or boolean, literals included), or an array or tuple, or
// a union of them, whose elements are all cycle-inert. tsc's string[] values, as executeCommandLine.ts's
// Object.entries reads them, hold only strings. A type met again while walking is taken as inert where
// it was first met, since a type whose every member is inert can't hold anything else.
func (l *lowering) cycleInert(proven *checker.Type, seen map[*checker.Type]bool) bool {
	proven = l.concrete(proven)
	if seen[proven] {
		return true
	}
	seen[proven] = true
	switch {
	case interfaceScalar(proven):
		return true
	case proven.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range proven.Types() {
			if !l.cycleInert(member, seen) {
				return false
			}
		}
		return true
	case l.isArrayType(proven) || checker.IsTupleType(proven):
		for _, element := range l.checker.GetTypeArguments(proven) {
			if !l.cycleInert(element, seen) {
				return false
			}
		}
		return true
	}
	return false
}
