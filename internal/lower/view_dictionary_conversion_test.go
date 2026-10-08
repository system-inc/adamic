package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestDictionaryConversionRecordOperationOriginMutant(t *testing.T) {
	for _, drop := range []bool{false, true} {
		program := dictionaryConversionMutationProgram(!drop, 1)
		err := dictionaryConversionRecordOperations(program, newAllocationFlowGraph(program))
		if drop {
			if err != nil {
				t.Fatalf("origin mutant did not lose refusal: %v", err)
			}
			t.Log("dropping conversion origin loses pinned record set refusal")
		} else if err == nil || !strings.Contains(err.Error(), "record set through a readonly dictionary view without a producer storage certificate") {
			t.Fatalf("want producer refusal, got %v", err)
		}
	}
}

func dictionaryConversionMutationProgram(mark bool, receiver int) *ir.Program {
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Union}, {Type: ir.Object}, {Type: ir.Record}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.Box{Value: ir.ObjectLiteral{}}},
			ir.Declare{Local: 1, Value: ir.Narrow{Value: ir.Read{Local: 0, Of: ir.Union}, To: ir.Object, Dictionary: mark, DictionaryWhere: "probe.a:2:1"}},
			ir.Declare{Local: 2, Value: ir.RecordLiteral{Element: ir.Array}},
			ir.Evaluate{Value: ir.RecordCall{Method: "set", Arguments: []ir.Expression{ir.Read{Local: receiver, Of: ir.Record}, ir.StringConstant{Index: 0}, ir.ArrayLiteral{Element: ir.String}}, Element: ir.Array, Returns: ir.Array}},
		}, Strings: []string{"other"},
	}
	assignAllocationSites(program)
	return program
}

func TestDictionaryReferenceProjectionKeepsOriginalContract(t *testing.T) {
	l, target, release := mixedUnionLowering(t, `interface Ref {[key:string]:string[]} type Target=string|number|boolean|Ref|number[]|string[]|null|undefined;`)
	defer release()
	node := l.program.Files()[0].Statements.Nodes[0]
	original, err := l.viewContract(node, target)
	if err != nil {
		t.Fatal(err)
	}
	unsupported := l.result.ViewContracts[original-1].Unsupported
	if unsupported == "" {
		t.Fatal("missing original member obligation")
	}
	selected, ok := l.dictionaryReferenceReadContract(node, target)
	if !ok || selected == original {
		t.Fatal("no distinct reference selector")
	}
	if l.result.ViewContractTypes[int(target.Id())] != original || l.result.ViewContracts[original-1].Unsupported != unsupported {
		t.Fatal("original descriptor was weakened")
	}
	read := ir.Property{DictionaryReference: true, DictionaryKey: ir.Undefined{}, ViewContract: selected}
	if !ir.SelectedDictionaryReadCertificate(l.result, read) || ir.PrimitiveDictionaryReadCertificate(l.result, read) {
		t.Fatal("reference selector confused with primitive selector")
	}
}

func TestDictionaryReferenceProjectionCannotTrustUnknown(t *testing.T) {
	for _, source := range []string{`type Target=string|number|boolean|unknown;`, `type Target=string|number|boolean|any;`, `type Target<T>=string|number|boolean|T;`} {
		l, target, release := mixedUnionLowering(t, source)
		node := l.program.Files()[0].Statements.Nodes[0]
		if _, ok := l.dictionaryReferenceReadContract(node, target); ok {
			t.Fatal("opaque arm authorized reference selection")
		}
		release()
	}
}
