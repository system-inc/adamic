package ir

import "testing"

func TestPrimitiveDictionaryNullableKinds(t *testing.T) {
	program := &Program{ViewContracts: []ViewContract{
		{Kind: ViewScalar, Of: String}, {Kind: ViewScalar, Of: Number},
		{Kind: ViewUnion, Members: []ViewContractID{1, 2}},
		{Kind: ViewNullable, Element: 3, Null: true, Undefined: true},
		{Kind: ViewNullable, Element: 3, Unsupported: "nullable union member selection"},
		{Kind: ViewNullable, Element: 6},
		{Kind: ViewUnknown, Unsupported: "unknown"},
		{Kind: ViewUnion, Members: []ViewContractID{1, 7}},
		{Kind: ViewNull, Null: true},
	}}
	kinds, ok := DictionaryReadKinds(program, 4)
	if !ok {
		t.Fatal("complete primitive nullable descriptor refused")
	}
	for _, wanted := range []string{"string", "number", "null", "undefined"} {
		found := false
		for _, kind := range kinds {
			found = found || kind == wanted
		}
		if !found {
			t.Fatalf("missing %s: %v", wanted, kinds)
		}
	}
	for _, id := range []ViewContractID{0, 5, 6, 7, 8, 10} {
		if _, ok := DictionaryReadKinds(program, id); ok {
			t.Fatalf("unsupported, cyclic or unavailable %d admitted", id)
		}
	}
	if _, ok := DictionaryReadKinds(program, 9); !ok {
		t.Fatal("explicit null descriptor refused")
	}
}

func TestPrimitiveDictionaryReadCertificate(t *testing.T) {
	program := &Program{ViewContracts: []ViewContract{{Kind: ViewScalar, Of: String}, {Kind: ViewObject, Of: Object}}}
	for _, id := range []ViewContractID{0, 2} {
		if PrimitiveDictionaryReadCertificate(program, Property{DictionaryPrimitive: true, DictionaryKey: Undefined{}, ViewContract: id}) {
			t.Fatalf("unavailable/reference %d admitted", id)
		}
	}
	if PrimitiveDictionaryReadCertificate(program, Property{DictionaryKey: Undefined{}, ViewContract: 1}) {
		t.Fatal("ordinary certificate replaced")
	}
	if !PrimitiveDictionaryReadCertificate(program, Property{DictionaryPrimitive: true, DictionaryKey: Undefined{}, ViewContract: 1}) {
		t.Fatal("complete selected primitive refused")
	}
}
