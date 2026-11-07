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

func TestParserNodeArrayLayout(test *testing.T) {
	test.Parallel()
	fields := parserNodeArrayFields()
	layout, nodeArrayLayoutError := NodeArrayLayout(ir.Object, fields)
	if nodeArrayLayoutError != nil {
		test.Fatal(nodeArrayLayoutError)
	}
	for index, name := range []string{"pos", "end", "hasTrailingComma", "transformFlags"} {
		field, readError := layout.Read(name)
		if readError != nil {
			test.Fatal(readError)
		}
		if field.Slot != index ||
			field.PresenceBit != index ||
			field.InitializationBit != index ||
			field.Of != fields[index].Of {
			test.Fatalf("wrong slot for %s: %+v", name, field)
		}
	}
	if _, readError := layout.Read("unknown"); readError == nil {
		test.Fatal("unknown metadata admitted")
	}
}

func TestParserNodeArrayPresence(test *testing.T) {
	test.Parallel()
	fields := append(
		parserNodeArrayFields(),
		FieldContract{Name: "cache", DeclaredType: "string | undefined", Of: ir.String, Optional: true},
	)
	layout, nodeArrayLayoutError := NodeArrayLayout(ir.Object, fields)
	if nodeArrayLayoutError != nil {
		test.Fatal(nodeArrayLayoutError)
	}
	cache, readError := layout.Read("cache")
	if readError != nil {
		test.Fatal(readError)
	}
	if !cache.Optional || cache.Slot != 4 || cache.PresenceBit != 4 || cache.InitializationBit != 4 {
		test.Fatalf("presence %+v", cache)
	}

}

func TestParserNodeArraySchemaFailsClosed(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"pos", "end", "hasTrailingComma", "transformFlags"} {
		fields := parserNodeArrayFields()
		for index := range fields {
			if fields[index].Name == name {
				fields[index].Of = ir.Object
			}
		}
		if _, nodeArrayLayoutError := NodeArrayLayout(ir.Object, fields); nodeArrayLayoutError == nil {
			test.Fatalf("wrong representation accepted for %s", name)
		}
	}
	if _, nodeArrayLayoutError := NodeArrayLayout(
		ir.Object,
		append(parserNodeArrayFields(), FieldContract{Name: "newRequired", DeclaredType: "number", Of: ir.Number}),
	); nodeArrayLayoutError == nil {
		test.Fatal("source schema drift ignored")
	}
}
