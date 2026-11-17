package lower

import (
	"errors"
	"strings"
	"testing"
)

// The premise of the FlowNode adaptation (#5pstmvw): tsc's FlowNode is a union whose arms declare
// node and antecedent with different types. With both readonly on every arm but FlowLabel, whose
// antecedent a label writes as it gathers its antecedents, a FlowLabel seen as the union is a
// read-only view of antecedent, since a union's property is read only when any arm's is
// (IsReadonlySymbol on the union's own property). So the view relates and reads, and a write
// through the FlowLabel itself still lowers. These pins hold that on six lines before the
// adaptation asks it of tsc's binder and checker.

const flowUnionArms = `interface FlowStart { flags: number; readonly node: string | undefined; readonly antecedent: undefined; }
interface FlowLabel { flags: number; readonly node: undefined; antecedent: FlowNode[] | undefined; }
interface FlowAssignment { flags: number; readonly node: string; readonly antecedent: FlowNode; }
type FlowNode = FlowStart | FlowLabel | FlowAssignment;
`

const flowUnionProgram = `function addAntecedent(label: FlowLabel, antecedent: FlowNode): void {
	(label.antecedent || (label.antecedent = [])).push(antecedent);
}
const start: FlowStart = { flags: 1, node: undefined, antecedent: undefined };
const label: FlowLabel = { flags: 2, node: undefined, antecedent: undefined };
addAntecedent(label, start);
const seen: FlowNode = label;
const holder: { flow: FlowNode | undefined } = { flow: undefined };
holder.flow = label;
const assignment: FlowAssignment = { flags: 3, node: "x", antecedent: label };
console.log(seen.flags, holder.flow.flags, assignment.node, assignment.antecedent.flags, label.antecedent ? label.antecedent.length : 0);
`

func TestFlowUnionReadonlyArmsRelate(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, flowUnionArms+flowUnionProgram); err != nil {
		t.Fatalf("a FlowLabel seen as the union whose other arms' antecedent is readonly should relate and lower, got %v", err)
	}
}

// The mutant: with antecedent writable on every arm, the union's antecedent is a slot that can write
// FlowNode[] | undefined into a FlowAssignment, which reads FlowNode, so seeing a FlowLabel as the
// union is refused. That the pin above lowers is then the readonly arms' doing, not a rule that
// never looks.
func TestFlowUnionWritableArmsRefused(t *testing.T) {
	t.Parallel()
	writable := strings.ReplaceAll(flowUnionArms, "readonly antecedent", "antecedent")
	_, err := lowerSource(t, writable+flowUnionProgram)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "seen as") || !strings.Contains(refusal.What, "can write") {
		t.Fatalf("with antecedent writable on every arm, a FlowLabel seen as the union should be refused as a view that can write, got %v", err)
	}
	t.Logf("refused as expected: %s at %s", refusal.What, refusal.Where)
}

// What a downcast of the union to FlowLabel meets, as tsc's checker does at every
// (flow as FlowLabel): logged, not asserted, so the adaptation's census can be read against it.
func TestFlowUnionDowncastToLabel(t *testing.T) {
	t.Parallel()
	program := flowUnionArms + `function antecedents(flow: FlowNode): number {
	if (flow.flags & 2) {
		const label = flow as FlowLabel;
		return label.antecedent ? label.antecedent.length : 0;
	}
	return 0;
}
const label: FlowLabel = { flags: 2, node: undefined, antecedent: undefined };
console.log(antecedents(label));
`
	if _, err := lowerSource(t, program); err != nil {
		t.Logf("the union cast to FlowLabel: %v", err)
		return
	}
	t.Log("the union cast to FlowLabel lowers")
}
