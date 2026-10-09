package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestUntaggedViewMemberTags(t *testing.T) {
	t.Parallel()
	contracts := []ir.ViewContract{
		{Kind: ir.ViewUnion, Name: "Left | Right", Of: ir.Object, Members: []ir.ViewContractID{2, 3}},
		{Kind: ir.ViewObject, Name: "Left", Of: ir.Object, Fields: []ir.ViewFieldContract{{Name: "kind", Contract: 4}, {Name: "nested", Contract: 6}}},
		{Kind: ir.ViewObject, Name: "Right", Of: ir.Object, Fields: []ir.ViewFieldContract{{Name: "flavor", Contract: 5}, {Name: "maybe", Contract: 4, Optional: true}}},
		{Kind: ir.ViewScalar, Of: ir.Number, Allowed: []ir.ViewLiteral{{Of: ir.Number, Number: 1}, {Of: ir.Number, Number: 2}}},
		{Kind: ir.ViewScalar, Of: ir.String, Allowed: []ir.ViewLiteral{{Of: ir.String, String: "right"}}},
		{Kind: ir.ViewObject, Of: ir.Object, Fields: []ir.ViewFieldContract{{Name: "label", Contract: 7}}},
		{Kind: ir.ViewScalar, Of: ir.String},
	}
	members, err := UntaggedViewMembers(contracts, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 || members[0].Contract != 2 || members[1].Contract != 3 {
		t.Fatalf("member identity lost: %#v", members)
	}
	if len(members[0].Tags) != 1 || members[0].Tags[0].Field != "kind" || len(members[0].Tags[0].Allowed) != 2 {
		t.Fatalf("own kind alternatives lost: %#v", members[0])
	}
	if len(members[1].Tags) != 1 || members[1].Tags[0].Field != "flavor" {
		t.Fatalf("optional tag used to exclude a member: %#v", members[1])
	}
	// Plans own their literal slice, not the shared registry's writable slice.
	members[0].Tags[0].Allowed[0].Number = 99
	if contracts[3].Allowed[0].Number != 1 {
		t.Fatal("tag plan mutated shared registry")
	}
	if contracts[members[0].Contract-1].Fields[1].Contract != 6 {
		t.Fatal("nested read lost its child contract")
	}
}

func TestUntaggedViewStructuralFallback(t *testing.T) {
	t.Parallel()
	contracts := []ir.ViewContract{
		{Kind: ir.ViewUnion, Name: "Left | Right", Members: []ir.ViewContractID{2, 3}},
		{Kind: ir.ViewObject, Of: ir.Object},
		{Kind: ir.ViewObject, Of: ir.Object},
	}
	members, err := UntaggedViewMembers(contracts, 1)
	if err != nil || len(members) != 2 || len(members[0].Tags) != 0 || len(members[1].Tags) != 0 {
		t.Fatalf("structural fallback lost: %v %#v", err, members)
	}
	for _, kind := range []ir.ViewKind{ir.ViewUnknown, ir.ViewArray, ir.ViewCallable} {
		contracts[2].Kind = kind
		if _, err := UntaggedViewMembers(contracts, 1); err == nil {
			t.Fatalf("member kind %d admitted without its adapter", kind)
		}
	}
	contracts[2].Kind = ir.ViewObject
	contracts[0].Members[1] = 99
	if _, err := UntaggedViewMembers(contracts, 1); err == nil {
		t.Fatal("invalid member id admitted")
	}
}
