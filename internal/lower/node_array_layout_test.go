package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func parserNodeArrayFields() []FieldContract {
	return []FieldContract{
		{Name: "pos", DeclaredType: "number", Of: ir.Number},
		{Name: "end", DeclaredType: "number", Of: ir.Number},
		{Name: "hasTrailingComma", DeclaredType: "boolean", Of: ir.Boolean},
		{Name: "transformFlags", DeclaredType: "TransformFlags", Of: ir.Number},
	}
}

func TestParserNodeArrayLayout(t *testing.T) {
	fields := parserNodeArrayFields()
	layout, err := NodeArrayLayout(ir.Object, fields)
	if err != nil {
		t.Fatal(err)
	}
	for i, name := range []string{"pos", "end", "hasTrailingComma", "transformFlags"} {
		field, err := layout.Read(name)
		if err != nil {
			t.Fatal(err)
		}
		if field.Slot != i || field.PresenceBit != i || field.InitializationBit != i || field.Of != fields[i].Of {
			t.Fatalf("wrong slot for %s: %+v", name, field)
		}
	}
	if _, err := layout.Read("unknown"); err == nil {
		t.Fatal("unknown metadata admitted")
	}
}

func TestParserNodeArrayPresence(t *testing.T) {
	fields := append(parserNodeArrayFields(), FieldContract{Name: "cache", DeclaredType: "string | undefined", Of: ir.String, Optional: true})
	layout, err := NodeArrayLayout(ir.Object, fields)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := layout.Read("cache")
	if err != nil {
		t.Fatal(err)
	}
	if !cache.Optional || cache.Slot != 4 || cache.PresenceBit != 4 || cache.InitializationBit != 4 {
		t.Fatalf("presence %+v", cache)
	}

}

func TestParserNodeArraySchemaFailsClosed(t *testing.T) {
	for _, name := range []string{"pos", "end", "hasTrailingComma", "transformFlags"} {
		fields := parserNodeArrayFields()
		for i := range fields {
			if fields[i].Name == name {
				fields[i].Of = ir.Object
			}
		}
		if _, err := NodeArrayLayout(ir.Object, fields); err == nil {
			t.Fatalf("wrong representation accepted for %s", name)
		}
	}
	if _, err := NodeArrayLayout(ir.Object, append(parserNodeArrayFields(), FieldContract{Name: "newRequired", DeclaredType: "number", Of: ir.Number})); err == nil {
		t.Fatal("source schema drift ignored")
	}
}
