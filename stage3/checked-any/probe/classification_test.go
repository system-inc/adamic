package main

import "testing"

func TestRefusalMustBeInsideDeclaration(t *testing.T) {
	for _, c := range []struct {
		reason string
		inside bool
	}{
		{"entry.ts:12:4: refused", true},
		{"entry.ts:9:4: refused", false},
		{"entry.ts:21:4: refused", false},
		{"capture.d.ts:12:4: refused", false},
		{"no position", false},
	} {
		if got := insideDeclaration(c.reason, "entry.ts", 10, 20); got != c.inside {
			t.Fatalf("%s: inside=%v want %v", c.reason, got, c.inside)
		}
	}
}

func TestUnobservedUseNamesItsOutsideRequirement(t *testing.T) {
	r := map[string]any{"file": "stock.ts", "line": 14, "column": 2, "declaration_name": "Config", "declaration_kind": "InterfaceDeclaration", "binding": "value"}
	_, kind := unobservedUse(r)
	if kind != "type_only_consumer_outside" {
		t.Fatal(kind)
	}
	r["declaration_kind"] = "FunctionDeclaration"
	r["generic_declaration"] = true
	_, kind = unobservedUse(r)
	if kind != "generic_instantiation_outside" {
		t.Fatal(kind)
	}
	r["generic_declaration"] = false
	r["references"] = []any{map[string]any{"line": 30, "column": 2, "within_declaration": false, "kind": "CallExpression"}}
	_, kind = unobservedUse(r)
	if kind != "consumer_outside_declaration" {
		t.Fatal(kind)
	}
}
