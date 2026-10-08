package ir

import "testing"

func TestMixedArrayContractBounds(t *testing.T) {
	program := &Program{ViewContracts: []ViewContract{
		{Kind: ViewScalar, Of: String, Name: "string"},
		{Kind: ViewObject, Of: Object, Fields: []ViewFieldContract{{Name: "version", Contract: 1}}},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 2}},
	}}
	if !MixedArrayContract(program, 3) {
		t.Fatal("represented scalar/object alternatives refused")
	}
	for _, id := range []ViewContractID{0, -1, 4} {
		if MixedArrayContract(program, id) {
			t.Fatal("invalid root admitted")
		}
	}
	for _, id := range []ViewContractID{0, -1, 4} {
		program.ViewContracts[2].Members = []ViewContractID{1, id}
		if MixedArrayContract(program, 3) {
			t.Fatal("invalid member admitted")
		}
	}
	program.ViewContracts[2].Members = []ViewContractID{1, 2}
	program.ViewContracts[1].Unsupported = "recursive intersection payload"
	if MixedArrayContract(program, 3) {
		t.Fatal("unsupported descendant admitted")
	}
	program.ViewContracts[1].Unsupported = ""
	program.ViewContracts[1].FixedTuple = true
	if MixedArrayContract(program, 3) {
		t.Fatal("tuple storage admitted as ordinary object")
	}
	program.ViewContracts[1].FixedTuple = false
	program.ViewContracts[0] = ViewContract{Kind: ViewNull, Of: Object, Null: true}
	if MixedArrayContract(program, 3) {
		t.Fatal("uncertified null/object family admitted")
	}
	program.ViewContracts[0] = ViewContract{Kind: ViewScalar, Of: String}
	program.ViewContracts[1].Fields[0].Contract = 0
	if MixedArrayContract(program, 3) {
		t.Fatal("missing scalar field contract admitted")
	}
}
