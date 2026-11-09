package lower

// Job-only probe for #ny0c3gb (never for main): with ADAMIC_PROBE_RECORDS set, recordMayReach also
// answers with two other source lists and prints every receiver whose verdict differs from today's,
// as Compiler's (a), (b) and (c) and Adamic's ruling ask:
//   (a) today's: each record literal's own type, its contextual type, its variable's declared type;
//   (b) the rule: each fresh literal type replaced by its widened type, which a stored value has;
//   (c) (b) with a literal's own type dropped where a contextual or declared type rides along.
// It prints one histogram line when the lists are built.

import (
	"fmt"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

var recordProbe = os.Getenv("ADAMIC_PROBE_RECORDS") != ""

type recordProbeLists struct {
	known         bool
	widened, kept []*checker.Type
	differs       [3]int
	calls         int
}

var recordProbeState = map[*lowering]*recordProbeLists{}

func (l *lowering) recordProbeLists() *recordProbeLists {
	state := recordProbeState[l]
	if state == nil {
		state = &recordProbeLists{}
		recordProbeState[l] = state
	}
	if state.known {
		return state
	}
	state.known = true
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return state
	}
	fresh, freshEmpty, companions, alone := 0, 0, 0, 0
	widen := func(t *checker.Type) *checker.Type {
		if t != nil && t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0 {
			return l.checker.GetWidenedType(t)
		}
		return t
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression && l.entriesRecordLiteral(node) {
			own := l.checker.GetTypeAtLocation(node)
			freshOwn := own.Flags()&checker.TypeFlagsObject != 0 && own.ObjectFlags()&checker.ObjectFlagsFreshLiteral != 0
			if freshOwn {
				fresh++
				if len(l.checker.GetPropertiesOfType(own)) == 0 && len(l.checker.GetIndexInfosOfType(own)) == 0 {
					freshEmpty++
				}
			}
			ride := []*checker.Type{}
			if contextual := l.slotType(node); contextual != nil {
				ride = append(ride, contextual)
			}
			parent := node.Parent
			for parent != nil && (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindParenthesizedExpression) {
				parent = parent.Parent
			}
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				ride = append(ride, l.checker.GetTypeAtLocation(parent.Name()))
			}
			if freshOwn {
				names := []string{}
				for _, companion := range ride {
					names = append(names, l.checker.TypeToString(companion))
				}
				parentKind := "none"
				if node.Parent != nil {
					parentKind = node.Parent.Kind.String()
				}
				fmt.Fprintf(os.Stderr, "probe-records-literal\t%s\town=%s\tparent=%s\tcompanions=%q\n", l.program.Where(node), l.checker.TypeToString(own), parentKind, names)
			}
			state.widened = append(state.widened, widen(own))
			for _, t := range ride {
				state.widened = append(state.widened, widen(t))
			}
			if len(ride) == 0 {
				alone++
				state.kept = append(state.kept, widen(own))
			} else {
				companions++
			}
			for _, t := range ride {
				state.kept = append(state.kept, widen(t))
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	fmt.Fprintf(os.Stderr, "probe-records-lists\ttoday=%d\twidened=%d\tkept=%d\tfresh=%d\tfreshEmpty=%d\twithCompanion=%d\talone=%d\n", len(l.entriesRecordStorageTypes()), len(state.widened), len(state.kept), fresh, freshEmpty, companions, alone)
	return state
}

// recordProbeVerdicts compares today's verdict with (b)'s and (c)'s for one receiver.
func (l *lowering) recordProbeVerdicts(receiver *checker.Type, today bool) {
	state := l.recordProbeLists()
	state.calls++
	widened := l.typesMayReach(receiver, state.widened, true, 0)
	kept := l.typesMayReach(receiver, state.kept, true, 0)
	if widened == today && kept == widened {
		return
	}
	text := "<nil>"
	id := 0
	if receiver != nil {
		text = l.checker.TypeToString(receiver)
		id = int(receiver.Id())
	}
	if widened != today {
		state.differs[0]++
	}
	if kept != widened {
		state.differs[1]++
	}
	fmt.Fprintf(os.Stderr, "probe-records\treceiver=%d\ta=%t\tb=%t\tc=%t\toverflow=%t\tcalls=%d\tab=%d\tbc=%d\t%s\n", id, today, widened, kept, load.RelationOverflowMarked(l.checker), state.calls, state.differs[0], state.differs[1], text)
}
