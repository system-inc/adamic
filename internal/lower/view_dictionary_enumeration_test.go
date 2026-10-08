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
			ViewContracts:          []ir.ViewContract{{Kind: ir.ViewObject, Unsupported: "tuple"}},
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
