package javascript

import "testing"

// Source Node's findIndex supplies the index of the witness's sole number member.
// The production selector must return that same zero-based index.
func TestViewMixedUnionSelectsNumber(test *testing.T) {
	test.Parallel()
	const witness = `const snapshot = {kind:'number',value:42}; const members = [{kind:'number'}];`
	runViewNode(test, witness+`console.log(members.findIndex(member => member.kind === snapshot.kind));`, "0\n", "", 0)
	runViewNode(test, viewTestRuntime+
		MixedUnionRuntime()+
		witness+
		`console.log(adamicViewMixedUnionSelect(snapshot,members,undefined,'view.value','number'));`, "0\n", "", 0)
}
