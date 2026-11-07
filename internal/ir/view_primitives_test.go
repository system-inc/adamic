package ir

import "testing"

func TestPrimitiveViewMembers(test *testing.T) {
	test.Parallel()
	program := &Program{ViewContracts: []ViewContract{
		{Kind: ViewScalar, Of: String},
		{Kind: ViewScalar, Of: Boolean, Allowed: []ViewLiteral{{Of: Boolean, Boolean: false}}},
		{Kind: ViewUndefined, Undefined: true},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 2, 3}},
		{Kind: ViewUnknown, Unsupported: "unknown"},
		{Kind: ViewObject, Of: Object},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 5}},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 6}},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{9}},
		{Kind: ViewScalar, Of: MaybeNumber, Undefined: true},
	}}
	members, ok := PrimitiveViewMembers(program, 4)
	if !ok || len(members) != 3 || len(members[1].Allowed) != 1 || members[1].Allowed[0].Boolean {
		test.Fatalf("lost finite false/undefined member: %#v %t", members, ok)
	}
	for _, id := range []ViewContractID{0, 5, 6, 7, 8, 9, 11} {
		if members, ok := PrimitiveViewMembers(program, id); ok || members != nil {
			test.Fatalf("incomplete contract %d enabled primitive dispatch: %#v", id, members)
		}
	}
	members, ok = PrimitiveViewMembers(program, 10)
	if !ok || len(members) != 2 || members[0].Of != Number || members[1].Kind != ViewUndefined {
		test.Fatalf("lost packed optional contract: %#v %t", members, ok)
	}
}
