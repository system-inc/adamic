package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// This fails during lowering, before the raw double could reach adamic_retain.
func TestPrimitiveAdmittingSlotsUseBoxes(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `const object = {}; const array = [0, object]; let value: {} = 1; value = 'x'; value = object; function kind(argument: {}): string { return typeof argument; } console.log(kind(2));`)
	if err != nil {
		t.Fatal(err)
	}
	foundArray, foundParameter, foundBinding := false, false, false
	for _, local := range program.Locals {
		if local.Name == "argument" {
			foundParameter = true
			if local.Type != ir.Union {
				t.Fatalf("{} parameter is %v, want boxed union", local.Type)
			}
		}
		if local.Name == "value" {
			foundBinding = true
			if local.Type != ir.Union {
				t.Fatalf("{} binding is %v, want boxed union", local.Type)
			}
		}
	}
	walk(program.Main, func(node any) bool {
		if array, ok := node.(ir.ArrayLiteral); ok {
			foundArray = true
			if array.Element != ir.Union {
				t.Fatalf("[0, object] element slot is %v, want boxed union", array.Element)
			}
			for _, element := range array.Elements {
				if element.Type() != ir.Union {
					t.Fatalf("raw %v in boxed array slot", element.Type())
				}
			}
		}
		return true
	})
	if !foundArray || !foundParameter || !foundBinding {
		t.Fatal("missing representation witness")
	}
}
