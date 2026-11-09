package lower

// Job-only probe for #ny0c3gb (never for main): with ADAMIC_PROBE_RECORDS set, recordMayReach also answers
// as main c1e9379e did (each record literal's own type, contextual type and variable type, and no implicit
// {} source), and prints every receiver where the head's verdict differs from today's, the fourth column
// Adamic's gate asks for. It prints the head's source list once per lowering.

import (
	"fmt"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

var recordProbe = os.Getenv("ADAMIC_PROBE_RECORDS") != ""

type recordProbeState struct {
	today       []*checker.Type
	known       bool
	calls       int
	toNo, toYes int
}

var recordProbeStates = map[*lowering]*recordProbeState{}

// todaySources is main c1e9379e's entriesRecordStorageTypes, verbatim but for its cache.
func (l *lowering) todaySources() *recordProbeState {
	state := recordProbeStates[l]
	if state == nil {
		state = &recordProbeState{}
		recordProbeStates[l] = state
	}
	if state.known {
		return state
	}
	state.known = true
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return state
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression && l.entriesRecordLiteral(node) {
			state.today = append(state.today, l.checker.GetTypeAtLocation(node))
			if contextual := l.slotType(node); contextual != nil {
				state.today = append(state.today, contextual)
			}
			parent := node.Parent
			for parent != nil && (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindParenthesizedExpression) {
				parent = parent.Parent
			}
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				state.today = append(state.today, l.checker.GetTypeAtLocation(parent.Name()))
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	head := []string{}
	fresh := 0
	for _, source := range l.entriesRecordStorageTypes() {
		head = append(head, l.checker.TypeToString(source))
		if source.Flags()&checker.TypeFlagsObject != 0 && source.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
			fresh++
		}
	}
	fmt.Fprintf(os.Stderr, "probe-records-lists\ttoday=%d\thead=%d\tfreshHeadSources=%d\thead sources=%q\n", len(state.today), len(head), fresh, head)
	return state
}

// todayTypesMayReach is main c1e9379e's typesMayReach, verbatim.
func (l *lowering) todayTypesMayReach(receiver *checker.Type, sources []*checker.Type, records bool, depth int) bool {
	if receiver == nil || depth > 8 {
		return true
	}
	if receiver.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range receiver.Types() {
			if member.Flags()&checker.TypeFlagsNullable == 0 && l.todayTypesMayReach(member, sources, records, depth+1) {
				return true
			}
		}
		return false
	}
	if receiver.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range receiver.Types() {
			if l.todayTypesMayReach(member, sources, records, depth+1) {
				return true
			}
		}
	}
	if receiver.Flags()&checker.TypeFlagsInstantiable != 0 {
		if concrete := l.concrete(receiver); concrete != nil && concrete != receiver {
			return l.todayTypesMayReach(concrete, sources, records, depth+1)
		}
		return l.todayTypesMayReach(l.checker.GetBaseConstraintOfType(receiver), sources, records, depth+1)
	}
	if records && l.enumerationRecordStorageType(receiver) {
		return true
	}
	// Both ways: a record seen as a wider type, and a record narrowed to a narrower one (a predicate,
	// a discriminant), arrive alike.
	for _, source := range sources {
		if source == receiver || l.checker.IsTypeAssignableTo(source, receiver) || l.checker.IsTypeAssignableTo(receiver, source) {
			return true
		}
	}
	return false
}

func (l *lowering) recordProbeVerdict(receiver *checker.Type, head bool) {
	state := l.todaySources()
	state.calls++
	today := l.todayTypesMayReach(receiver, state.today, true, 0)
	if today == head {
		return
	}
	if today {
		state.toNo++
	} else {
		state.toYes++
	}
	text, id := "<nil>", 0
	if receiver != nil {
		text, id = l.checker.TypeToString(receiver), int(receiver.Id())
	}
	fmt.Fprintf(os.Stderr, "probe-records\treceiver=%d\ttoday=%t\thead=%t\toverflow=%t\tcalls=%d\ttoNo=%d\ttoYes=%d\t%s\n", id, today, head, load.RelationOverflowMarked(l.checker), state.calls, state.toNo, state.toYes, text)
}
