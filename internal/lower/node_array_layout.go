package lower

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// FieldContract describes a resolved field. It cannot represent an unchecked any.
// Layout planning alone never admits an array cast or emits a runtime instruction.
type FieldContract struct {
	Name, DeclaredType string
	Of                 ir.Type
	Optional           bool
}

type NodeArrayField struct {
	FieldContract
	Slot, PresenceBit, InitializationBit int
}

type ArrayLayout struct{ Fields []NodeArrayField }

// NodeArrayLayout assigns fixed metadata slots after the existing array header.
// Byte offsets and allocation emission await the array runtime owner. Optional
// fields carry distinct presence and initialization states; absence is not a value.
func NodeArrayLayout(element ir.Type, extras []FieldContract) (ArrayLayout, error) {
	if element < ir.Number || element > ir.Float64Array {
		return ArrayLayout{}, fmt.Errorf("stage 0 can't lower NodeArray element representation yet")
	}
	required := []struct {
		name string
		of   ir.Type
	}{{"pos", ir.Number}, {"end", ir.Number}, {"hasTrailingComma", ir.Boolean}, {"transformFlags", ir.Number}}
	byName := map[string]FieldContract{}
	for _, field := range extras {
		if _, duplicate := byName[field.Name]; duplicate {
			return ArrayLayout{}, fmt.Errorf("duplicate NodeArray metadata field %s", field.Name)
		}
		if field.Of != ir.Number && field.Of != ir.Boolean && field.Of != ir.String {
			return ArrayLayout{}, fmt.Errorf("stage 0 can't lower NodeArray metadata field %s of type %s yet", field.Name, field.DeclaredType)
		}
		byName[field.Name] = field
	}
	layout := ArrayLayout{}
	add := func(field FieldContract) {
		slot := len(layout.Fields)
		layout.Fields = append(layout.Fields, NodeArrayField{FieldContract: field, Slot: slot, PresenceBit: slot, InitializationBit: slot})
	}
	for _, want := range required {
		field, exists := byName[want.name]
		if !exists || field.Optional || field.Of != want.of {
			return ArrayLayout{}, fmt.Errorf("NodeArray requires %s with its declared required representation", want.name)
		}
		add(field)
		delete(byName, want.name)
	}
	// Preserve declaration order for explicit optional extensions. New required
	// fields fail closed until the pinned source schema has been reviewed.
	for _, field := range extras {
		if _, extra := byName[field.Name]; !extra {
			continue
		}
		if !field.Optional {
			return ArrayLayout{}, fmt.Errorf("stage 0 can't lower NodeArray metadata field %s yet", field.Name)
		}
		add(field)
	}
	return layout, nil
}

func (layout ArrayLayout) Read(field string) (NodeArrayField, error) {
	for _, slot := range layout.Fields {
		if slot.Name == field {
			return slot, nil
		}
	}
	return NodeArrayField{}, fmt.Errorf("stage 0 can't lower NodeArray metadata field %s yet", field)
}
