package ir

import (
	"slices"
	"testing"
)

func TestPrimitiveUnionSlotWriteDomains(t *testing.T) {
	program := &Program{ViewContracts: []ViewContract{
		{Kind: ViewScalar, Of: Number},
		{Kind: ViewScalar, Of: String},
		{Kind: ViewScalar, Of: Boolean},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 2}},
		{Kind: ViewScalar, Of: Number, Allowed: []ViewLiteral{{Of: Number, Number: 12}}},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{5, 2}},
		{Kind: ViewUnknown, Unsupported: "unknown"},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 7}},
		{Kind: ViewUnion, Of: Union, Members: []ViewContractID{1, 3}},
		{Kind: ViewUnion, Of: Union},
	}}
	for _, test := range []struct {
		from, to ViewContractID
		allow    bool
	}{
		{1, 4, true}, {2, 4, true}, {3, 4, false}, {4, 4, true},
		{5, 6, true}, {1, 6, false}, {6, 4, true}, {4, 6, false},
		{1, 8, false}, {8, 4, false}, {9, 4, false}, {1, 10, false},
	} {
		if got := slices.Contains(ScalarWriteContracts(program, test.from), test.to); got != test.allow {
			t.Fatalf("source contract %d to slot %d: got %v, want %v", test.from, test.to, got, test.allow)
		}
	}
}
