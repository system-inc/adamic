package lower

import (
	"errors"
	"strings"
	"testing"
)

// The premise of the FlowNode adaptation (#5pstmvw). Adamic sees a value through a union as each
// member it's assignable to (invariance.go's widened), and tsc's flow arms aren't discriminated by a
// literal (flags is a number), so a FlowCall is assignable to FlowArrayMutation and FlowCondition too,
// whose writable node could write a wider node into the call's. That's the 39 refusals. With node
// readonly on every arm, no arm the call fits can write its node, so the view relates. FlowLabel's
// antecedent stays writable, since a label writes it as it gathers its antecedents, and a write
// through the FlowLabel itself still lowers. These pins hold that on a few lines before the
// adaptation asks it of tsc's binder and checker.

const flowUnionArms = `interface FlowStart { flags: number; readonly node: string | undefined; readonly antecedent: undefined; }
interface FlowLabel { flags: number; readonly node: undefined; antecedent: FlowNode[] | undefined; }
interface FlowCall { flags: number; readonly node: string; readonly antecedent: FlowNode; }
interface FlowArrayMutation { flags: number; readonly node: string | number; readonly antecedent: FlowNode; }
type FlowNode = FlowStart | FlowLabel | FlowCall | FlowArrayMutation;
`

const flowUnionProgram = `function addAntecedent(label: FlowLabel, antecedent: FlowNode): void {
	if (label.antecedent) label.antecedent.push(antecedent);
	else label.antecedent = [antecedent];
}
const start: FlowStart = { flags: 1, node: undefined, antecedent: undefined };
const label: FlowLabel = { flags: 2, node: undefined, antecedent: undefined };
addAntecedent(label, start);
const call: FlowCall = { flags: 3, node: "f", antecedent: label };
const seen: FlowNode = call;
const holder: { flow: FlowNode | undefined } = { flow: undefined };
holder.flow = call;
addAntecedent(label, call);
console.log(String(seen.flags) + ' ' + String(holder.flow.flags) + ' ' + call.node + ' ' + String(call.antecedent.flags) + ' ' + String(label.antecedent ? label.antecedent.length : 0));
`

func TestFlowUnionReadonlyArmsRelate(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, flowUnionArms+flowUnionProgram); err != nil {
		t.Fatalf("a FlowCall seen as the union, with node readonly on every arm, should relate and lower, got %v", err)
	}
}

// The mutant: with node writable on every arm, a FlowCall seen as the union is seen as
// FlowArrayMutation too, which can write a number where the call's string is read, so the view is
// refused. That the pin above lowers is then the readonly nodes' doing, not a rule that never looks.
func TestFlowUnionWritableArmsRefused(t *testing.T) {
	t.Parallel()
	writable := strings.ReplaceAll(flowUnionArms, "readonly node", "node")
	_, err := lowerSource(t, writable+flowUnionProgram)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "FlowCall seen as") || !strings.Contains(refusal.What, "can write") {
		t.Fatalf("with node writable on every arm, a FlowCall seen as the union should be refused as a view that can write, got %v", err)
	}
	t.Logf("refused as expected: %s at %s", refusal.What, refusal.Where)
}

// FlowUnreachable fits FlowLabel's arm (both nodes undefined, and undefined is in the label's
// antecedent), whose antecedent stays writable, so seeing it as the union may be a view that can write
// FlowNode[] where undefined is read. Logged, not asserted: the adaptation's census says whether tsc
// holds this view anywhere.
func TestFlowUnionUnreachableSeesLabel(t *testing.T) {
	t.Parallel()
	program := strings.Replace(flowUnionArms, "type FlowNode = FlowStart", "interface FlowUnreachable { flags: number; readonly node: undefined; readonly antecedent: undefined; }\ntype FlowNode = FlowUnreachable | FlowStart", 1) + `const unreachable: FlowUnreachable = { flags: 0, node: undefined, antecedent: undefined };
const seen: FlowNode = unreachable;
console.log(String(seen.flags));
`
	if _, err := lowerSource(t, program); err != nil {
		t.Logf("FlowUnreachable seen as the union: %v", err)
		return
	}
	t.Log("FlowUnreachable seen as the union lowers")
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
console.log(String(antecedents(label)));
`
	if _, err := lowerSource(t, program); err != nil {
		t.Logf("the union cast to FlowLabel: %v", err)
		return
	}
	t.Log("the union cast to FlowLabel lowers")
}
