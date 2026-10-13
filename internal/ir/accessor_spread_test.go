package ir_test

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestObjectSpreadAccessorEffects(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{}, {}, {}, {}},
		Classes:   []ir.Class{{Literal: true, Accessors: []ir.Accessor{{Getter: 0, Setter: -1}}}},
		Functions: []ir.Function{{MayThrow: true}, {Parameters: []int{3}}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ObjectLiteral{}},
			ir.Declare{Local: 1, Value: ir.ObjectLiteral{Class: 1}},
			ir.Assign{Local: 3, Value: ir.ObjectLiteral{}},
		},
	}
	for _, probe := range []struct {
		name   string
		source ir.Expression
		want   bool
	}{
		{"fresh data", ir.ObjectLiteral{}, false},
		{"literal getter", ir.ObjectLiteral{Class: 1}, true},
		{"saved data", ir.Read{Local: 0}, false},
		{"saved getter", ir.Read{Local: 1}, true},
		{"unknown storage", ir.Read{Local: 2}, true},
		{"incoming parameter", ir.Read{Local: 3}, true},
	} {
		if got := program.ObjectSpreadMayThrow(ir.ObjectLiteral{Spread: probe.source}); got != probe.want {
			t.Errorf("%s: throw edge %t, want %t", probe.name, got, probe.want)
		}
	}
}
