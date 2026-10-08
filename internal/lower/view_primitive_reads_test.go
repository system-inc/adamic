package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestPrimitiveViewDestructuringAdmission(t *testing.T) {
	for _, probe := range []struct {
		name, field      string
		viewed, admitted bool
	}{
		{"nullable primitive", "string | number | undefined", true, true},
		{"ordinary primitive binding", "string | number | undefined", false, true},
		{"object union remains unsupported", "string | { n: number }", true, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := "interface Base { readonly kind: 'box' | 'other' }; interface Target extends Base { readonly kind: 'box'; readonly value: " + probe.field + " }; function helper(x:Target):void { const {value}=x; console.log(String(value)); }"
			if probe.viewed {
				source += "function cast(x:Base):Target{return x as Target;}"
			}
			program, err := lowerSource(t, source)
			if !probe.admitted {
				if err == nil {
					t.Fatal("unsupported binding was admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, fn := range program.Functions {
				walk(fn.Body, func(node any) bool {
					if field, ok := node.(ir.Property); ok && field.Name == "value" {
						count++
						if !field.Nullish || field.ViewContract == 0 || field.ViewTypeID == 0 || !strings.Contains(field.View, "field value") {
							t.Fatalf("binding lost read contract: %#v", field)
						}
					}
					return true
				})
			}
			if count != 1 {
				t.Fatalf("want one declared binding read, got %d", count)
			}
		})
	}
}

func TestPrimitiveBindingConversionRequiresCompleteDeclaration(t *testing.T) {
	program := &ir.Program{ViewContracts: []ir.ViewContract{
		{Kind: ir.ViewUnion, Of: ir.Union, Members: []ir.ViewContractID{2, 3}},
		{Kind: ir.ViewScalar, Of: ir.String}, {Kind: ir.ViewScalar, Of: ir.Number},
		{Kind: ir.ViewUnknown, Unsupported: "unknown"},
		{Kind: ir.ViewUnion, Of: ir.Union, Members: []ir.ViewContractID{2, 4}},
	}}
	for _, probe := range []struct {
		id    ir.ViewContractID
		label string
		admit bool
	}{
		{1, "value (field value)", true}, {0, "value (field value)", false},
		{5, "value (field value)", false}, {1, "x.value", true},
	} {
		property := ir.Property{Name: "value", Of: ir.Union, Nullish: true, ViewContract: probe.id, View: probe.label}
		if got := primitiveBindingConversion(program, property); got != probe.admit {
			t.Fatalf("conversion boundary %#v: got %v", probe, got)
		}
	}
}
