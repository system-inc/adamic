package ir

import "testing"

func TestTupleViewMembersRejectsUnprovenAlternatives(t *testing.T) {
	p := &Program{ViewContracts: []ViewContract{{Kind: ViewScalar, Of: Number}, {Kind: ViewObject, Of: Object, FixedTuple: true}, {Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 2}}}}
	if ids, ok := TupleViewMembers(p, 3); !ok || len(ids) != 2 {
		t.Fatal("complete tuple plan missing")
	}
	for _, bad := range []ViewContract{{Kind: ViewObject, Of: Object}, {Kind: ViewUnknown}, {Kind: ViewScalar, Of: Number, Unsupported: "any"}, {Kind: ViewCallable, Of: Closure}, {Kind: ViewUnion, Members: []ViewContractID{3}}, {Kind: ViewUnion, Members: []ViewContractID{0}}} {
		p.ViewContracts[1] = bad
		if _, ok := TupleViewMembers(p, 3); ok {
			t.Fatalf("unproven alternative admitted: %#v", bad)
		}
	}
}
