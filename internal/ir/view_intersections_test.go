package ir

import (
	"slices"
	"testing"
)

// boundedProgram builds Tag = A | B, where A and its intersection copy differ
// only by an optional copy of one canonical object and by an incomplete array
// reservation, as complete tsc declarations produce them. Ids are positional:
// 3 Child, 4 and 5 its optional copies, 6 and 7 arrays, 8 A, 9 L & A, 10 B, 12 Tag.
func boundedProgram() *Program {
	number := func(n float64) []ViewLiteral { return []ViewLiteral{{Of: Number, Number: n}} }
	return &Program{ViewContracts: []ViewContract{
		{Kind: ViewScalar, Of: Number, Name: "A.kind", Allowed: number(1)},
		{Kind: ViewScalar, Of: Number, Name: "B.kind", Allowed: number(2)},
		{Kind: ViewObject, Of: Object, Name: "Child", Fields: []ViewFieldContract{{Name: "kind", Contract: 1}}},
		{Kind: ViewObject, Of: Object, Name: "Child | undefined", Undefined: true, ObjectPresent: 3},
		{Kind: ViewObject, Of: Object, Name: "Child | undefined", Undefined: true, ObjectPresent: 3},
		{Kind: ViewArray, Of: Array, Name: "Child[] | undefined", Undefined: true, Element: 3},
		{Kind: ViewArray, Of: Array, Name: "Child[] | undefined", Undefined: true},
		{Kind: ViewObject, Of: Object, Name: "A", Fields: []ViewFieldContract{{Name: "kind", Contract: 1}, {Name: "child", Contract: 4, Optional: true}, {Name: "list", Contract: 6, Optional: true}}},
		{Kind: ViewObject, Of: Object, Name: "L & A", Intersection: true, Fields: []ViewFieldContract{{Name: "kind", Contract: 1}, {Name: "child", Contract: 5, Optional: true}, {Name: "list", Contract: 7, Optional: true}}},
		{Kind: ViewObject, Of: Object, Name: "B", Fields: []ViewFieldContract{{Name: "kind", Contract: 2}}},
		{Kind: ViewScalar, Of: Number, Name: "1 | 2", Allowed: append(number(1), number(2)...)},
		{Kind: ViewUnion, Of: Object, Name: "Tag", Members: []ViewContractID{8, 9, 10}, Fields: []ViewFieldContract{{Name: "kind", Contract: 11}}},
	}}
}

func TestIntersectionUnionArmsCoalesceCopies(t *testing.T) {
	program := boundedProgram()
	tag, arms := IntersectionUnionArms(program, 12)
	if tag != "kind" || !slices.Equal(arms, []ViewContractID{9, 10}) {
		t.Fatalf("arms = %q %v, want kind [9 10]", tag, arms)
	}
	// An array copy with a literal obligation is not an array checked by kind alone.
	program.ViewContracts[6].Allowed = []ViewLiteral{{Of: Number, Number: 1}}
	if tag, _ := IntersectionUnionArms(program, 12); tag != "" {
		t.Fatalf("unequal arms with one tag must not select, got %q", tag)
	}
}

func TestBoundedIntersectionContracts(t *testing.T) {
	program := boundedProgram()
	ids, valid := BoundedIntersectionContracts(program, 12, nil)
	if !valid || !slices.Equal(ids, []ViewContractID{12, 9, 3, 10}) {
		t.Fatalf("contracts = %v %t", ids, valid)
	}
	// A refused descriptor is deferred unless the caller is still deciding it.
	program.ViewContracts[2].Unsupported = "recursive intersection payload"
	if ids, _ := BoundedIntersectionContracts(program, 12, nil); slices.Contains(ids, 3) {
		t.Fatalf("deferred descriptor was walked: %v", ids)
	}
	if ids, _ := BoundedIntersectionContracts(program, 12, func(id ViewContractID) bool { return id == 3 || id == 5 }); !slices.Contains(ids, 3) {
		t.Fatalf("pending descriptor was not walked: %v", ids)
	}
	// A nullable member is outside the bounded walk and refuses the whole read.
	program.ViewContracts[2].Unsupported = ""
	program.ViewContracts[9].Fields = append(program.ViewContracts[9].Fields, ViewFieldContract{Name: "maybe", Contract: 13})
	program.ViewContracts = append(program.ViewContracts, ViewContract{Kind: ViewNullable, Name: "Child | null", Element: 3})
	if _, valid := BoundedIntersectionContracts(program, 12, nil); valid {
		t.Fatal("nullable member must refuse")
	}
}
