package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestDictionaryEntryDerivedOriginMutant(t *testing.T) {
	for _, drop := range []bool{false, true} {
		array := ir.Read{Local: 1, Of: ir.Array}
		program := &ir.Program{
			Locals:                 []ir.Local{{Type: ir.Object}, {Type: ir.Array}},
			ViewOrigins:            []ir.Expression{ir.Read{Local: 0, Of: ir.Object}},
			DictionaryEntryOrigins: []ir.Expression{array},
			ViewContractTypes:      map[int]ir.ViewContractID{42: 1},
			ViewContracts:          []ir.ViewContract{{Kind: ir.ViewObject, Of: ir.Object, FixedTuple: true}},
			Main: []ir.Statement{
				ir.Declare{Local: 0, Value: ir.ObjectLiteral{}},
				ir.Declare{Local: 1, Value: ir.ArrayLiteral{Element: ir.Object}},
				ir.ForOf{Iterable: array, ViewRead: ir.ArrayViewRead{ViewTypeID: 42, View: "entries[element]", Element: ir.Object}},
			},
		}
		if drop {
			program.DictionaryEntryOrigins = nil
		}
		err := (&lowering{result: program}).checkLazyViewReads()
		if drop {
			if err != nil {
				t.Fatalf("mutant did not remove the demanded refusal: %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "field [element] with unsupported tuple") {
			t.Fatalf("want named tuple refusal, got %v", err)
		}
	}
}

func TestDictionaryEntryRefusalDoesNotCaptureIndependentTuple(t *testing.T) {
	receiver := ir.Read{Local: 2, Of: ir.Array}
	program := &ir.Program{
		Locals:                 []ir.Local{{Type: ir.Object}, {Type: ir.Array}, {Type: ir.Array}},
		ViewOrigins:            []ir.Expression{ir.Read{Local: 0, Of: ir.Object}},
		DictionaryEntryOrigins: []ir.Expression{ir.Read{Local: 1, Of: ir.Array}},
		ViewContractTypes:      map[int]ir.ViewContractID{42: 1},
		ViewContracts:          []ir.ViewContract{{Kind: ir.ViewObject, Of: ir.Object, FixedTuple: true}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ObjectLiteral{}},
			ir.Declare{Local: 1, Value: ir.ArrayLiteral{Element: ir.Object}},
			ir.Declare{Local: 2, Value: ir.ArrayLiteral{Element: ir.Object}},
			ir.ForOf{Iterable: receiver, ViewRead: ir.ArrayViewRead{ViewTypeID: 42, View: "independent[element]", Element: ir.Object}},
		},
	}
	if err := (&lowering{result: program}).checkLazyViewReads(); err != nil {
		t.Fatalf("independent certified tuple refused: %v", err)
	}
}
