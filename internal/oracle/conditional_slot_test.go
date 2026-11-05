package oracle

import "testing"

// A conditional going straight into a readonly slot judges each branch against that slot, not against
// the branches' union tsc reduces it to (#2vjqseg, tsc's semver.ts:211): it lowers and agrees with Node
// on every backend, and leaks nothing. The control that must stay refused, a conditional with no slot
// whose variable then writes, is internal/lower's TestConditionalBranchesAreJudgedAgainstTheirSlot.
func TestConditionalIntoAReadonlySlotAgreesWithNode(t *testing.T) {
	t.Parallel()
	generatorAgrees(t, "internal/oracle/testdata/conditional_slot/readonly_field.a")
}
