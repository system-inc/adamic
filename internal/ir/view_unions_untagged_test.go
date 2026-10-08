package ir

import "testing"

func TestViewUnionDiscriminantOverlaps(t *testing.T) {
	contracts := []ViewContract{
		{Kind: ViewObject, Fields: []ViewFieldContract{{Name: "kind", Contract: 3}}},
		{Kind: ViewObject, Fields: []ViewFieldContract{{Name: "kind", Contract: 4}}},
		{Kind: ViewScalar, Of: Number, Allowed: []ViewLiteral{{Of: Number, Number: 1}, {Of: Number, Number: 2}}},
		{Kind: ViewScalar, Of: Number, Allowed: []ViewLiteral{{Of: Number, Number: 2}}},
	}
	root := ViewContract{Kind: ViewUnion, Members: []ViewContractID{1, 2}, Fields: []ViewFieldContract{{Name: "kind", Contract: 3}}}
	if ViewUnionHasDiscriminant(contracts, root) {
		t.Fatal("overlapping tags selected a member")
	}
	contracts[3].Allowed[0].Number = 3
	if !ViewUnionHasDiscriminant(contracts, root) {
		t.Fatal("disjoint tags lost")
	}
	contracts[1].Fields[0].Optional = true
	if ViewUnionHasDiscriminant(contracts, root) {
		t.Fatal("optional absence treated as a certificate")
	}
}
